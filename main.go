package main

import (
	"net/http"
	"os"
)

func main() {
	setupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
