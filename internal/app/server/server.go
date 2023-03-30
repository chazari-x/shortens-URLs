package server

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

	if err := http.ListenAndServe(c.ServerAddress, gzipHandle(r)); err != nil {
		return err
	}

	return nil
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			_, err := io.WriteString(w, err.Error())
			if err != nil {
				log.Print("gzip - write string err:", err)
			}
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				log.Print("defer gzip.Writer err:", err)
			}
		}(gz)

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
