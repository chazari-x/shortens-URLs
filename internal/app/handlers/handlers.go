package handlers

import (
	"io"
	"net/http"
	"strconv"

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
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func Post(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)

	if string(b) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		id, err := storage.Add(string(b))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)

		w.Header().Set("content-type", "text/plain; charset=utf-8")

		_, err = w.Write([]byte("http://localhost:8080/" + strconv.Itoa(id)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
