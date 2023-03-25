package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"main/internal/app/config"
	"main/internal/app/handlers"
)

func StartSever() error {
	c := config.GetConfig()

	r := chi.NewRouter()

	if c.BaseURL != "" {
		r.Get("/"+c.BaseURL+"/{id}", handlers.Get)
	} else {
		r.Get("/{id}", handlers.Get)
	}
	r.Post("/", handlers.Post)
	r.Post("/api/shorten", handlers.Shorten)

	if err := http.ListenAndServe(c.ServerAddress, r); err != nil {
		return err
	}

	return nil
}
