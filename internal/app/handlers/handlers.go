package handlers

import (
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"main/internal/app/storage"
)

func Get(w http.ResponseWriter, r *http.Request) {
	url, err := storage.Get(chi.URLParam(r, "id"))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func Post(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if string(b) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := storage.Add(string(b))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte("http://localhost:8080/" + string(id)))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
