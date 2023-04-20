package inmemory

import (
	"context"
	"errors"
	"strconv"

	. "main/internal/app/storage/model"
)

type InMemory struct {
	ServerAddress string
	BaseURL       string
}

func (c *InMemory) PingDB(_ context.Context) error {
	return errors.New("db is disabled")
}

func (c *InMemory) Add(url, user string) (string, error) {
	S.ID++

	id := strconv.FormatInt(int64(S.ID), 36)
	S.URLs[S.ID] = Event{
		ID:   id,
		URL:  url,
		User: user,
	}

	return id, nil
}

func (c *InMemory) BatchAdd(urls []string, user string) ([]string, error) {
	S.ID++

	var ids []string

	for i := 0; i < len(urls); i++ {
		id := strconv.FormatInt(int64(S.ID), 36)
		S.URLs[S.ID] = Event{
			ID:   id,
			URL:  urls[i],
			User: user,
		}

		ids = append(ids, id)

		if i < len(urls)-1 {
			S.ID++
		}
	}

	return ids, nil
}

func (c *InMemory) Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if int(id) > S.ID {
		return "", ErrStorageIsNil
	}

	return S.URLs[int(id)].URL, nil
}

func (c *InMemory) GetAll(user string) ([]URLs, error) {
	var UserURLs []URLs
	for _, i := range S.URLs {
		if i.User == user {
			UserURLs = append(UserURLs, URLs{
				ShortURL:    "http://" + c.ServerAddress + c.BaseURL + i.ID,
				OriginalURL: i.URL,
			})
		}
	}

	return UserURLs, nil
}
