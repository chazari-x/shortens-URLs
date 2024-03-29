package handlers

import (
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
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

type (
	BatchOriginal struct {
		ID  string `json:"correlation_id"`
		URL string `json:"original_url"`
	}
	BatchShort struct {
		ID  string `json:"correlation_id"`
		URL string `json:"short_url"`
	}
)

type Controller struct {
	sConf   config.Config
	storage storage.Storage
	db      *sql.DB
}

func NewController(c storage.Storage, s config.Config, db *sql.DB) *Controller {
	return &Controller{storage: c, sConf: s, db: db}
}

type Middleware func(http.Handler) http.Handler

func MiddlewaresConveyor(h http.Handler) http.Handler {
	middlewares := []Middleware{gzipMiddleware, cookieMiddleware}
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipMiddleware(next http.Handler) http.Handler {
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

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func makeUserIdentification() (string, error) {
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

var userIdentification = "user_identification"

var identification struct {
	ID string
}

func cookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uid string

		cookie, err := r.Cookie(userIdentification)
		if err != nil {
			if !errors.Is(err, http.ErrNoCookie) {
				log.Print("r.Cookie err: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			uid, err = makeUserIdentification()
			if err != nil {
				log.Print("COOKIE: set user identification err: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     userIdentification,
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

		ctx := context.WithValue(r.Context(), identification, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	url, del, err := c.storage.Get(chi.URLParam(r, "id"))
	if err != nil {
		if strings.Contains(err.Error(), "the storage is empty or the element is missing") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Print("GET: get err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if del {
		w.WriteHeader(http.StatusGone)
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

	uid := fmt.Sprintf("%v", r.Context().Value(identification))

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

	var status = http.StatusCreated

	id, err := c.storage.Add(string(b), uid)

	if err != nil {
		if !strings.Contains(err.Error(), "url conflict") {
			log.Print("POST: add err: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status = http.StatusConflict
	}

	log.Printf("add: %d, user: %s, id: %s, url: %s", status, uid, id, string(b))
	w.WriteHeader(status)

	_, err = w.Write([]byte("http://" + c.sConf.ServerAddress + c.sConf.BaseURL + id))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) Shorten(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := fmt.Sprintf("%v", r.Context().Value(identification))

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

	var status = http.StatusCreated

	id, err := c.storage.Add(url.URL, uid)
	if err != nil {
		if !strings.Contains(err.Error(), "url conflict") {
			log.Printf("add: %s, user: %s, id: %s, url: %s", err, uid, id, url.URL)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status = http.StatusConflict
	}

	log.Printf("add: %d, user: %s, id: %s, url: %s", status, uid, id, url.URL)
	w.WriteHeader(status)

	marshal, err := json.Marshal(short{
		Result: "http://" + c.sConf.ServerAddress + c.sConf.BaseURL + id,
	})
	if err != nil {
		log.Print("SHORTEN: json marshal err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(marshal)
	if err != nil {
		log.Print("SHORTEN: write err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) BatchAdd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := fmt.Sprintf("%v", r.Context().Value(identification))

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("BATCH ADD: read all err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if string(b) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var bOriginal []BatchOriginal

	err = json.Unmarshal(b, &bOriginal)
	if err != nil {
		log.Print("BATCH ADD: json unmarshal err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var urls []string

	for _, i := range bOriginal {
		urls = append(urls, i.URL)
	}

	id, err := c.storage.BatchAdd(urls, uid)
	if err != nil {
		if strings.Contains(err.Error(), "the storage is empty or the element is missing") {
			log.Printf("batchAdd: %d, user: %s, ids: %s, urls: %s", http.StatusBadRequest, uid, id, urls)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("batchAdd: %s, user: %s, ids: %s, urls: %s", err, uid, id, urls)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("batchAdd: true, user: %s, ids: %s, urls: %s", uid, id, urls)

	bShort := make([]BatchShort, len(id))

	for i := 0; i < len(id); i++ {
		bShort[i] = BatchShort{
			ID:  bOriginal[i].ID,
			URL: "http://" + c.sConf.ServerAddress + c.sConf.BaseURL + id[i],
		}
	}

	marshal, err := json.Marshal(bShort)
	if err != nil {
		log.Print("BATCH ADD: json marshal err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(marshal)
	if err != nil {
		log.Print("BATCH ADD: write err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) UserURLs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := fmt.Sprintf("%v", r.Context().Value(identification))

	URLs, err := c.storage.GetAll(uid)
	if err != nil {
		log.Print("UserURLs: GetAll err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(URLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
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

func (c *Controller) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if c.db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err := c.storage.PingDB(r.Context())
	if err != nil {
		log.Print("PING: ping db err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *Controller) BatchUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := fmt.Sprintf("%v", r.Context().Value(identification))

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("BATCH UPDATE: read all err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if string(b) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var ids []string

	err = json.Unmarshal(b, &ids)
	if err != nil {
		log.Print("BATCH UPDATE: json unmarshal err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.storage.BatchUpdate(ids, uid)

	w.WriteHeader(http.StatusAccepted)
}
