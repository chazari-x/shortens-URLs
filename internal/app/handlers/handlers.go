package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"main/internal/app/storage"
)

type short struct {
	Result string `json:"result"`
}

type some struct {
	URL string `json:"url"`
}

func Get(w http.ResponseWriter, r *http.Request) {
	url, err := storage.GetOriginal(chi.URLParam(r, "id"))
	if err != nil {
		if strings.Contains(err.Error(), "the storage is empty or the element is missing") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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

func Shorten(w http.ResponseWriter, r *http.Request) {
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

	url := some{}

	err = json.Unmarshal(b, &url)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := storage.GetShortened(url.URL)
	if err != nil {
		if strings.Contains(err.Error(), "the storage is empty or the element is missing") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	marshal, err := json.Marshal(short{
		Result: "http://localhost:8080/" + string(id),
	})
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(marshal)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
