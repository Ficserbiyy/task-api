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
	"github.com/Ficserbiyy/task-api/internal/auth"
	"github.com/Ficserbiyy/task-api/internal/handlers"
	"github.com/Ficserbiyy/task-api/internal/services"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
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

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	auth := auth.AuthMiddleware(db)

	// http://127.0.0.1:8080/swagger/index.html
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Post("/auth/register", gormRepository.Register())
	r.Post("/auth/login", gormRepository.Login())
	r.Post("/auth/logout", handlers.Logout)

	r.Group(func(r chi.Router) {
		r.Use(auth)

		r.Get("/me", gormRepository.Me())
		r.Get("/tasks", gormRepository.List())
		r.Post("/tasks", gormRepository.Create())

		r.Get("/tasks/{id}", gormRepository.GetOne())
		r.Patch("/tasks/{id}", gormRepository.Update())
		r.Delete("/tasks/{id}", gormRepository.Delete())
	})

	log.Println("Server listening on http://127.0.0.1:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
