package main

import (
	"encoding/json"
	"fmt"
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

				_, err := fmt.Fprintln(w, sURLs[i].URL)
				if err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError)
					return
				}
				break
			} else if i == len(sURLs)-1 {
				http.Error(w, http.StatusText(http.StatusBadRequest),
					http.StatusBadRequest)
				return
			}
		}
	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}
}

func sPost(w http.ResponseWriter, r *http.Request) {
	var aURL struct {
		AURL string `json:"aurl"`
	}

	if json.NewDecoder(r.Body).Decode(&aURL) != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if aURL.AURL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	} else {
		sURLs = append(sURLs, struct {
			id  string
			URL string
		}{id: strconv.Itoa(len(sURLs)), URL: aURL.AURL})
		w.WriteHeader(http.StatusCreated)

		w.Header().Set("content-type", "text/plain; charset=utf-8")

		_, err := fmt.Fprintln(w, "http://localhost:8080/"+strconv.Itoa(len(sURLs)-1))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	sURLs = append(sURLs, sURL{id: strconv.Itoa(len(sURLs)),
		URL: "https://vk.com/im?peers=c19&sel=390295814&z=photo390295814_457243386%2Fmail320677"})

	r := chi.NewRouter()
	r.Get("/*", sGet)

	r.Post("/", sPost)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Print("listen and serve err: ", err.Error())
	}
}
