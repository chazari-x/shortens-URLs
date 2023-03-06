package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var sURLs []sURL

type sURL struct {
	id  string
	URL string
}

func sGet(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "" {
		for i := 0; i < len(sURLs); i++ {
			if sURLs[i].id == r.URL.Path[1:] {
				w.WriteHeader(http.StatusTemporaryRedirect)
				w.Header().Set("content-type", "text/plain; charset=utf-8")
				_, err := w.Write([]byte(sURLs[i].URL))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				break
			} else if i == len(sURLs)-1 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func sPost(w http.ResponseWriter, r *http.Request) {
	var aURL struct {
		URL string `json:"url"`
	}

	if json.NewDecoder(r.Body).Decode(&aURL) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if aURL.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		sURLs = append(sURLs, struct {
			id  string
			URL string
		}{id: strconv.Itoa(len(sURLs)), URL: aURL.URL})
		w.WriteHeader(http.StatusCreated)

		w.Header().Set("content-type", "text/plain; charset=utf-8")

		_, err := w.Write([]byte("http://localhost:8080/" + strconv.Itoa(len(sURLs)-1)))
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
