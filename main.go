package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"runner/internal/executor"
	"runner/internal/handlers"
	"runner/internal/middleware"
)

func main() {
	// Structured JSON logging — readable by Render's log viewer
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	apiKey := os.Getenv("RUNNER_API_KEY")

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Index)
	mux.HandleFunc("/health", handlers.Health)
	mux.Handle("/run", middleware.Auth(apiKey, handlers.Run(executor.New())))

	// Wrap entire mux with request logger
	handler := middleware.Logger(mux)

	slog.Info("Go runner starting", "port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
