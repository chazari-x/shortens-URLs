package main

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var sURLs = make(map[int]string)

func sGet(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "" {
		url, err := Get(r.URL.Path[1:])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if url == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		_, err = w.Write([]byte(url))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func sPost(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)

	if string(b) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		id, err := Add(string(b))
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

func main() {
	r := chi.NewRouter()
	r.Get("/*", sGet)

	r.Post("/", sPost)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Print("listen and serve err: ", err.Error())
	}
}

func Add(url string) (int, error) {
	id := len(sURLs)
	sURLs[id] = url

	return id, nil
}

func Get(sid string) (string, error) {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return "", err
	}
	return sURLs[id], nil
}
