package handlers

import (
	"io"
	"net/http"

	"main/internal/app/storage"
)

func Get(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "" {
		url, err := storage.Get(r.URL.Path[1:])
		if err != nil {
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
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func Post(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)

	if string(b) == "" {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		id, err := storage.Add(string(b))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)

		_, err = w.Write([]byte("http://localhost:8080/" + id))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
