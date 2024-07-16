package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	if err := setupRoutes(); err != nil {
		slog.Error("setting up routes", "error", err)
		os.Exit(1)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
