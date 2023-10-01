package main

import (
	"net/http"
)

func main() {
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}
