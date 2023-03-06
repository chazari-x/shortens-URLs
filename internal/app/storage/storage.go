package storage

import "strconv"

var sURLs = make(map[int]string)

func Add(url string) (int, error) {
	id := len(sURLs)
	sURLs[id] = url

	return id, nil
}

func Get(sid string) (string, error) {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return "", err
	}
	return sURLs[id], nil
}
