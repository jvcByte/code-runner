package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"runner/internal/executor"
)

type runRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Stdin    string `json:"stdin"`
}

type runResponse struct {
	Stdout        string `json:"stdout"`
	Stderr        string `json:"stderr"`
	ExitCode      int    `json:"exit_code"`
	CompileOutput string `json:"compile_output"`
}

func Run(exec *executor.Executor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, executor.MaxCodeSize)

		var req runRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if len(req.Code) > executor.MaxCodeSize {
			http.Error(w, "code too large", http.StatusRequestEntityTooLarge)
			return
		}

		if req.Language != "go" {
			writeJSON(w, runResponse{
				Stderr:   fmt.Sprintf("unsupported language: %s", req.Language),
				ExitCode: 1,
			})
			return
		}

		result, err := exec.Run(req.Code, req.Stdin)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		writeJSON(w, runResponse{
			Stdout:   result.Stdout,
			Stderr:   result.Stderr,
			ExitCode: result.ExitCode,
		})
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
