package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"main/internal/app/handlers"
)

func StartSever() error {
	r := chi.NewRouter()
	r.Get("/{id}", handlers.Get)
	r.Post("/", handlers.Post)
	r.Post("/api/shorten", handlers.Shorten)

	if err := http.ListenAndServe(":8080", r); err != nil {
		return err
	}

	return nil
}
