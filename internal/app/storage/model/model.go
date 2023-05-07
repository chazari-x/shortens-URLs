package model

import "errors"

var S struct {
	URLs map[int]Event // Используется, если File не прописан
	ID   int           // Это ID последнего элемента в хранилище
}

type Event struct {
	ID     int    `json:"id"`
	URL    string `json:"url"`
	Del    bool   `json:"del"`
	UserID string `json:"user_id"`
}

type URLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var (
	ErrURLConflict  = errors.New("url conflict")
	ErrStorageIsNil = errors.New("the storage is empty or the element is missing")
)
