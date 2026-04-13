package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RunRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Stdin    string `json:"stdin"`
}

type RunResponse struct {
	Stdout      string `json:"stdout"`
	Stderr      string `json:"stderr"`
	ExitCode    int    `json:"exit_code"`
	CompileOutput string `json:"compile_output"`
}

const (
	maxCodeSize   = 64 * 1024       // 64KB
	maxOutputSize = 64 * 1024       // 64KB
	timeout       = 10 * time.Second
)

func runHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Code) > maxCodeSize {
		http.Error(w, "code too large", http.StatusRequestEntityTooLarge)
		return
	}

	if req.Language != "go" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RunResponse{
			Stderr:   fmt.Sprintf("unsupported language: %s", req.Language),
			ExitCode: 1,
		})
		return
	}

	// Write code to a temp directory
	dir, err := os.MkdirTemp("", "runner-*")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(dir)

	codePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(codePath, []byte(req.Code), 0644); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Run with timeout
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", codePath)
	cmd.Stdin = strings.NewReader(req.Stdin)
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &limitedWriter{w: &stdout, limit: maxOutputSize}
	cmd.Stderr = &limitedWriter{w: &stderr, limit: maxOutputSize}

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			exitCode = 124
			stderr.WriteString("\nExecution timed out (10s limit)")
		} else {
			exitCode = 1
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RunResponse{
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
		ExitCode:    exitCode,
		CompileOutput: "",
	})
}

// limitedWriter caps output at a byte limit
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	http.HandleFunc("/run", runHandler)
	http.HandleFunc("/health", healthHandler)

	log.Printf("Go runner listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
