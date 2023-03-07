package storage

import (
	"main/internal/pkg/shortens"
)

var sURLs = make(map[string]string)

func Add(url string) (string, error) {
	id := shortens.Shortens(len(sURLs))
	sURLs[id] = url

	return id, nil
}

func Get(id string) (string, error) {
	return sURLs[id], nil
}
