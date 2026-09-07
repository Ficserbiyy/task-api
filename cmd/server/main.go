// @title Task API
// @description A simple task API written in Go.
// @host 127.0.0.1:8080
// @BasePath /
package main

import (
	"log"
	"net/http"

	_ "github.com/Ficserbiyy/task-api/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	mux := http.NewServeMux()

	// http://127.0.0.1:8080/swagger/index.html
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server listening on http://127.0.0.1:8080")

	if err := http.ListenAndServe("127.0.0.1:8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
