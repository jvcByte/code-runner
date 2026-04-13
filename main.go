package main

import (
	"log"
	"net/http"
	"os"

	"runner/internal/executor"
	"runner/internal/handlers"
	"runner/internal/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	apiKey := os.Getenv("RUNNER_API_KEY")

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Index)
	mux.HandleFunc("/health", handlers.Health)
	mux.Handle("/run", middleware.Auth(apiKey, handlers.Run(executor.New())))

	log.Printf("Go runner listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
