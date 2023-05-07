package server

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"main/internal/app/config"
	h "main/internal/app/handlers"
	"main/internal/app/storage"
)

func StartSever() error {
	conf, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("parse config err: %s", err)
	}

	memoryModel, fileModel, dbModel, err := storage.StartStorage(conf)
	if err != nil {
		return fmt.Errorf("start storage file path err: %s", err)
	}

	var model storage.Storage
	var db *sql.DB

	if memoryModel != nil {
		model = memoryModel
	} else if fileModel != nil {
		model = fileModel
	} else if dbModel != nil {
		model = dbModel
		db = dbModel.DB
	} else {
		return fmt.Errorf("start storage err")
	}

	c := h.NewController(model, conf, db)

	r := chi.NewRouter()

	r.Get("/"+conf.BaseURL+"{id}", c.Get)
	r.Get("/api/user/urls", c.UserURLs)
	r.Get("/ping", c.Ping)

	r.Post("/", c.Post)
	r.Post("/api/shorten", c.Shorten)
	r.Post("/api/shorten/batch", c.BatchAdd)

	r.Delete("/api/user/urls", c.BatchUpdate)

	return http.ListenAndServe(conf.ServerAddress[:len(conf.ServerAddress)-1], h.MiddlewaresConveyor(r))
}
