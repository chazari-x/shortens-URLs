package storage

import (
	"main/internal/pkg/shortens"
)

var storageURLs []string

func Add(url string) (string, error) {
	id := shortens.Short(len(storageURLs))

	storageURLs = append(storageURLs, url)

	return id, nil
}

func Get(id string) (string, error) {
	i, err := shortens.Original(id)
	if err != nil {
		return "", err
	}

	return storageURLs[i], nil
}
