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

	// Validate API key if configured
	apiKey := os.Getenv("RUNNER_API_KEY")
	if apiKey != "" {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)

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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Recoding — Code Runner</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: #050709;
      color: #e2e8f0;
      font-family: 'Inter', system-ui, sans-serif;
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
    }
    .card {
      text-align: center;
      padding: 3rem 2.5rem;
      border: 1px solid rgba(255,255,255,0.08);
      border-radius: 16px;
      background: #0c0f14;
      max-width: 420px;
      width: 90%;
      box-shadow: 0 0 60px rgba(99,102,241,0.08);
    }
    .logo {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 10px;
      margin-bottom: 1.5rem;
    }
    svg { display: block; }
    h1 { font-size: 1.1rem; font-weight: 700; color: #f1f5f9; letter-spacing: -0.02em; }
    .sub { font-size: 13px; color: #64748b; margin-top: 0.3rem; }
    .status {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      margin-top: 1.5rem;
      padding: 0.4rem 0.9rem;
      background: rgba(16,185,129,0.1);
      border: 1px solid rgba(16,185,129,0.25);
      border-radius: 20px;
      font-size: 12px;
      font-weight: 600;
      color: #10b981;
      letter-spacing: 0.04em;
      text-transform: uppercase;
    }
    .dot {
      width: 7px; height: 7px;
      border-radius: 50%;
      background: #10b981;
      box-shadow: 0 0 6px #10b981;
      animation: pulse 1.5s ease-in-out infinite;
    }
    @keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.4} }
    .endpoint {
      margin-top: 1.5rem;
      padding: 0.75rem 1rem;
      background: #111520;
      border-radius: 8px;
      font-family: monospace;
      font-size: 12px;
      color: #818cf8;
      text-align: left;
    }
    .endpoint span { color: #64748b; }
  </style>
</head>
<body>
  <div class="card">
    <div class="logo">
      <!-- Hexagon icon -->
      <svg width="28" height="28" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
        <polygon points="16,2 28,9 28,23 16,30 4,23 4,9"
          stroke="#818cf8" stroke-width="2" stroke-linejoin="round" fill="none"/>
      </svg>
      <!-- Server icon -->
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#475569" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <rect x="2" y="2" width="20" height="8" rx="2"/>
        <rect x="2" y="14" width="20" height="8" rx="2"/>
        <line x1="6" y1="6" x2="6.01" y2="6"/>
        <line x1="6" y1="18" x2="6.01" y2="18"/>
      </svg>
    </div>
    <h1>Recoding Code Runner</h1>
    <p class="sub">Lightweight Go execution service</p>
    <div class="status">
      <span class="dot"></span>
      Online
    </div>
    <div class="endpoint">
      <span>POST</span> /run<br>
      <span>GET &nbsp;</span> /health
    </div>
  </div>
</body>
</html>`)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/run", runHandler)
	http.HandleFunc("/health", healthHandler)

	log.Printf("Go runner listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
