package executor

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	MaxCodeSize      = 64 * 1024
	MaxOutputSize    = 64 * 1024
	CompileTimeout   = 30 * time.Second
	ExecutionTimeout = 10 * time.Second
)

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type Executor struct {
	sem chan struct{}
}

// New creates an Executor that allows at most maxConcurrent simultaneous compilations.
func New() *Executor {
	return &Executor{sem: make(chan struct{}, 4)}
}

func (e *Executor) Run(code, stdin string) (*Result, error) {
	// Limit concurrent compilations to avoid OOM on memory-constrained hosts (e.g. Render free 512MB).
	e.sem <- struct{}{}
	defer func() { <-e.sem }()

	dir, err := os.MkdirTemp("", "runner-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	codePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(codePath, []byte(code), 0644); err != nil {
		return nil, err
	}

	// Compile with a generous timeout — cold builds can be slow.
	compileCtx, compileCancel := context.WithTimeout(context.Background(), CompileTimeout)
	defer compileCancel()

	binPath := filepath.Join(dir, "runner-bin")
	var compileStderr strings.Builder
	compileCmd := exec.CommandContext(compileCtx, "go", "build", "-o", binPath, codePath)
	compileCmd.Dir = dir
	compileCmd.Env = append(os.Environ(), "GOCACHE=/root/.cache/go-build")
	compileCmd.Stderr = &limitedWriter{w: &compileStderr, limit: MaxOutputSize}
	if err := compileCmd.Run(); err != nil {
		exitCode := 1
		if compileCtx.Err() == context.DeadlineExceeded {
			exitCode = 124
			compileStderr.WriteString("\nCompilation timed out (30s limit)")
			slog.Warn("compilation timed out")
		} else {
			slog.Warn("execution failed", "exit_code", exitCode, "stderr_preview", truncate(compileStderr.String(), 200))
		}
		return &Result{Stderr: compileStderr.String(), ExitCode: exitCode}, nil
	}

	// Run the compiled binary with the execution timeout.
	runCtx, runCancel := context.WithTimeout(context.Background(), ExecutionTimeout)
	defer runCancel()

	runCmd := exec.CommandContext(runCtx, binPath)
	runCmd.Stdin = strings.NewReader(stdin)
	runCmd.Dir = dir

	var stdout, stderr strings.Builder
	runCmd.Stdout = &limitedWriter{w: &stdout, limit: MaxOutputSize}
	runCmd.Stderr = &limitedWriter{w: &stderr, limit: MaxOutputSize}

	exitCode := 0
	if err := runCmd.Run(); err != nil {
		if runCtx.Err() == context.DeadlineExceeded {
			exitCode = 124
			stderr.WriteString("\nExecution timed out (10s limit)")
			slog.Warn("execution timed out")
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			slog.Warn("execution failed", "exit_code", exitCode, "stderr_preview", truncate(stderr.String(), 200))
		} else {
			exitCode = 1
			slog.Error("execution error", "error", err.Error())
		}
	}

	return &Result{
		Stdout:   stdout.String(),
		Stderr:   compileStderr.String() + stderr.String(),
		ExitCode: exitCode,
	}, nil
}

// limitedWriter caps output at a byte limit.
type limitedWriter struct {
	w     *strings.Builder
	limit int
	n     int
}

func (lw *limitedWriter) Write(p []byte) (int, error) {
	remaining := lw.limit - lw.n
	if remaining <= 0 {
		return len(p), nil
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	n, err := lw.w.Write(p)
	lw.n += n
	return n, err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
