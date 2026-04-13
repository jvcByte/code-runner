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

	cmd := exec.CommandContext(ctx, "go", "run", codePath)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &limitedWriter{w: &stdout, limit: MaxOutputSize}
	cmd.Stderr = &limitedWriter{w: &stderr, limit: MaxOutputSize}

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			slog.Warn("execution failed", "exit_code", exitCode, "stderr_preview", truncate(stderr.String(), 200))
		} else if ctx.Err() == context.DeadlineExceeded {
			exitCode = 124
			stderr.WriteString("\nExecution timed out (10s limit)")
			slog.Warn("execution timed out")
		} else {
			exitCode = 1
			slog.Error("execution error", "error", err.Error())
		}
	}

	return &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
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
