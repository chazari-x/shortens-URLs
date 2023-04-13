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

	sModel, err := storage.StartStorage(conf)
	if err != nil {
		return fmt.Errorf("start storage file path err: %s", err)
	}

	c := handlers.NewController(sModel, conf, sModel.DB)

	r := chi.NewRouter()
	r.Get("/"+conf.BaseURL+"{id}", c.Get)
	r.Get("/api/user/urls", c.UserURLs)
	r.Get("/ping", c.Ping)
	r.Post("/", c.Post)
	r.Post("/api/shorten", c.Shorten)
	r.Post("/api/shorten/batch", c.Batch)

	return http.ListenAndServe(conf.ServerAddress[:len(conf.ServerAddress)-1], c.GzipHandle(r))
}
