package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"main/internal/app/handlers"
)

func StartSever() error {
	r := chi.NewRouter()
	r.Get("/*", handlers.Get)
	r.Post("/", handlers.Post)

	if err := http.ListenAndServe(":8080", r); err != nil {
		return fmt.Errorf("listen and serve err: %s", err.Error())
	}

	return nil
}
