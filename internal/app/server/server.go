package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"main/internal/app/config"
	"main/internal/app/handlers"
	"main/internal/app/storage"
)

func StartSever() error {
	c, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("parse config err: %s", err)
	}

	if c.FileStoragePath != "" {
		err := storage.StartStorage(c.FileStoragePath)
		if err != nil {
			return fmt.Errorf("start storage file path err: %s", err)
		}
	}

	r := chi.NewRouter()

	if c.BaseURL != "" {
		r.Get("/"+c.BaseURL+"/{id}", handlers.Get)
	} else {
		r.Get("/{id}", handlers.Get)
	}
	r.Post("/", handlers.Post)
	r.Post("/api/shorten", handlers.Shorten)

	if err := http.ListenAndServe(c.ServerAddress, handlers.GzipHandle(r)); err != nil {
		return err
	}

	return nil
}
