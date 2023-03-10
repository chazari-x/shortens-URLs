package storage

import (
	"strconv"
)

var storageURLs []string

func Add(url string) (string, error) {
	id := strconv.FormatInt(int64(len(storageURLs)), 36)

	storageURLs = append(storageURLs, url)

	return id, nil
}

func Get(s string) (string, error) {
	id, err := strconv.ParseInt(s, 36, 64)
	if err != nil {
		return "", err
	}

	return storageURLs[id], nil
}
