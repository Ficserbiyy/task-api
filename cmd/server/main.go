// @title Task API
// @version 1.0
// @description A simple task API written in Go.
// @host 127.0.0.1:8080
// @BasePath /
package main

import (
	"log"
	"net/http"

	_ "github.com/Ficserbiyy/task-api/docs"
	"github.com/Ficserbiyy/task-api/internal/handlers"
	"github.com/Ficserbiyy/task-api/internal/services"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	db, err := services.ConnectToDatabase()
	if err != nil {
		log.Fatal(err)
	}

	gormRepository := handlers.TaskRepository{
		DB: db,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/register", gormRepository.Register())
	mux.HandleFunc("POST /auth/login", gormRepository.Login())
	mux.HandleFunc("/auth/logout", handlers.Logout)

	// http://127.0.0.1:8080/swagger/index.html
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server listening on http://127.0.0.1:8080")

	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
