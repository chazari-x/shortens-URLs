package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"main/internal/app/config"
	"main/internal/app/storage"
)

type (
	short struct {
		Result string `json:"result"`
	}

	original struct {
		URL string `json:"url"`
	}
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Print("GZIP: new reader err:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer func() {
				err := gz.Close()
				if err != nil {
					log.Print("GZIP: defer func reader err:", err)
				}
			}()

			r.Body = gz
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			log.Print("GZIP: new writer level err:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer func() {
			err := gz.Close()
			if err != nil {
				log.Print("GZIP: defer writer err:", err)
			}
		}()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	url, err := storage.Get(chi.URLParam(r, "id"))
	if err != nil {
		if strings.Contains(err.Error(), "the storage is empty or the element is missing") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Print("GET: get err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("POST: read all err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if string(b) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := storage.Add(string(b))
	if err != nil {
		log.Print("POST: add err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	c := config.Conf

	if c.BaseURL != "" {
		_, err = w.Write([]byte("http://" + c.ServerAddress + "/" + c.BaseURL + "/" + id))
	} else {
		_, err = w.Write([]byte("http://" + c.ServerAddress + "/" + id))
	}
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func Shorten(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("SHORTEN: read all err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if string(b) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url := original{}

	err = json.Unmarshal(b, &url)
	if err != nil {
		log.Print("SHORTEN: json unmarshal err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := storage.Add(url.URL)
	if err != nil {
		if strings.Contains(err.Error(), "the storage is empty or the element is missing") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Print("SHORTEN: add err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := config.Conf

	var marshal []byte
	if c.BaseURL != "" {
		marshal, err = json.Marshal(short{
			Result: "http://" + c.ServerAddress + "/" + c.BaseURL + "/" + id,
		})
	} else {
		marshal, err = json.Marshal(short{
			Result: "http://" + c.ServerAddress + "/" + id,
		})
	}
	if err != nil {
		log.Print("SHORTEN: json marshal err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(marshal)
	if err != nil {
		log.Print("SHORTEN: write err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
