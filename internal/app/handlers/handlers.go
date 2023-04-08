package handlers

import (
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

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

type Controller struct {
	storage storage.Storage
}

func NewController(s storage.Storage) *Controller {
	return &Controller{storage: s}
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func setUserIdentification() (string, error) {
	str := time.Now().Format("02012006150405")

	key, err := generateRandom(aes.BlockSize)
	if err != nil {
		return "", err
	}

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return "", err
	}

	id := fmt.Sprintf("%x", aesgcm.Seal(nil, nonce, []byte(str), nil))

	return id, nil
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (c *Controller) GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Print("GZIP: new reader err: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer func() {
				_ = gz.Close()
			}()

			r.Body = gz
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			log.Print("GZIP: new writer level err: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer func() {
			_ = gz.Close()
		}()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	url, err := c.storage.Get(chi.URLParam(r, "id"))
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

func (c *Controller) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	var uid string

	cookie, err := r.Cookie("user_identification")
	if err != nil {
		if !strings.Contains(err.Error(), "named cookie not present") {
			log.Print("POST: r.Cookie err: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		uid, err = setUserIdentification()
		if err != nil {
			log.Print("POST: set user identification err: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "user_identification",
			Value:    uid,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: false,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
	} else {
		uid = cookie.Value
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("POST: read all err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if string(b) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := c.storage.Add(string(b), uid)
	if err != nil {
		log.Print("POST: add err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	conf := config.Conf

	_, err = w.Write([]byte("http://" + conf.ServerAddress + conf.BaseURL + id))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) Shorten(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var uid string

	cookie, err := r.Cookie("user_identification")
	if err != nil {
		if !strings.Contains(err.Error(), "named cookie not present") {
			log.Print("r.Cookie err: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		uid, err = setUserIdentification()
		if err != nil {
			log.Print("POST: set user identification err: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "user_identification",
			Value:    uid,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: false,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
	} else {
		uid = cookie.Value
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("SHORTEN: read all err: ", err)
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
		log.Print("SHORTEN: json unmarshal err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := c.storage.Add(url.URL, uid)
	if err != nil {
		if strings.Contains(err.Error(), "the storage is empty or the element is missing") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Print("SHORTEN: add err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	conf := config.Conf

	marshal, err := json.Marshal(short{
		Result: "http://" + conf.ServerAddress + conf.BaseURL + id,
	})
	if err != nil {
		log.Print("SHORTEN: json marshal err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(marshal)
	if err != nil {
		log.Print("SHORTEN: write err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) UserURLs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("user_identification")
	if err != nil {
		if !strings.Contains(err.Error(), "named cookie not present") {
			log.Print("r.Cookie err: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	conf := config.Conf

	URLs, err := c.storage.GetAll(cookie.Value, conf.ServerAddress, conf.BaseURL)
	if err != nil {
		log.Print("UserURLs: GetAll err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(URLs)
	if err != nil {
		log.Print("UserURLs: json marshal err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Print("UserURLs: write err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
