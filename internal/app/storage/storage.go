package storage

import (
	"fmt"
	"strconv"
)

var storageURLs []string

func Add(url string) (string, error) {
	id := strconv.FormatInt(int64(len(storageURLs)), 36)

	storageURLs = append(storageURLs, url)

	return id, nil
}

func GetOriginal(s string) (string, error) {
	id, err := strconv.ParseInt(s, 36, 64)
	if err != nil {
		return "", err
	}

	if int(id) >= len(storageURLs) {
		return "", fmt.Errorf("the storage is empty or the element is missing")
	}

	return storageURLs[id], nil
}

//func GetShortened(s string) (string, error) {
//	if len(s) == 0 || len(storageURLs) == 0 {
//		return "", fmt.Errorf("the storage is empty or the element is missing")
//	}
//
//	var urlID string
//
//	for i, url := range storageURLs {
//		if url == s {
//			urlID = strconv.FormatInt(int64(i), 36)
//		}
//	}
//
//	return urlID, nil
//}
