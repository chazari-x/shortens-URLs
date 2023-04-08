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
	conf, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("parse config err: %s", err)
	}

	cModel := storage.NewStorageModel(conf.FileStoragePath)
	err = cModel.StartStorage()
	if err != nil {
		return fmt.Errorf("start storage file path err: %s", err)
	}

	c := handlers.NewController(cModel)

	r := chi.NewRouter()
	r.Get("/"+conf.BaseURL+"{id}", c.Get)
	r.Get("/api/user/urls", c.UserURLs)
	r.Post("/", c.Post)
	r.Post("/api/shorten", c.Shorten)

	return http.ListenAndServe(conf.ServerAddress[:len(conf.ServerAddress)-1], c.GzipHandle(r))
}
