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
	MaxCodeSize   = 64 * 1024
	MaxOutputSize = 64 * 1024
	Timeout       = 10 * time.Second
)

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type Executor struct{}

func New() *Executor { return &Executor{} }

func (e *Executor) Run(code, stdin string) (*Result, error) {
	dir, err := os.MkdirTemp("", "runner-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	codePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(codePath, []byte(code), 0644); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	// Compile first to avoid go module cache lock contention under concurrent load.
	binPath := filepath.Join(dir, "runner-bin")
	var compileStderr strings.Builder
	compileCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, codePath)
	compileCmd.Dir = dir
	compileCmd.Stderr = &limitedWriter{w: &compileStderr, limit: MaxOutputSize}
	if err := compileCmd.Run(); err != nil {
		exitCode := 1
		if ctx.Err() == context.DeadlineExceeded {
			exitCode = 124
			compileStderr.WriteString("\nCompilation timed out (10s limit)")
			slog.Warn("compilation timed out")
		} else {
			slog.Warn("execution failed", "exit_code", exitCode, "stderr_preview", truncate(compileStderr.String(), 200))
		}
		return &Result{Stderr: compileStderr.String(), ExitCode: exitCode}, nil
	}

	runCmd := exec.CommandContext(ctx, binPath)
	runCmd.Stdin = strings.NewReader(stdin)
	runCmd.Dir = dir

	var stdout, stderr strings.Builder
	runCmd.Stdout = &limitedWriter{w: &stdout, limit: MaxOutputSize}
	runCmd.Stderr = &limitedWriter{w: &stderr, limit: MaxOutputSize}

	exitCode := 0
	if err := runCmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
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
