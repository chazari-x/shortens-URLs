package inmemory

import (
	"context"
	"errors"
	"strconv"

	mod "main/internal/app/storage/model"
)

type InMemory struct {
	ServerAddress string
	BaseURL       string
}

func (c *InMemory) PingDB(_ context.Context) error {
	return errors.New("db is disabled")
}

func (c *InMemory) Add(url, user string) (string, error) {
	mod.S.ID++

	id := strconv.FormatInt(int64(mod.S.ID), 36)
	mod.S.URLs[mod.S.ID] = mod.Event{
		ID:   id,
		URL:  url,
		User: user,
	}

	return id, nil
}

func (c *InMemory) BatchAdd(urls []string, user string) ([]string, error) {
	mod.S.ID++

	var ids []string

	for i := 0; i < len(urls); i++ {
		id := strconv.FormatInt(int64(mod.S.ID), 36)
		mod.S.URLs[mod.S.ID] = mod.Event{
			ID:   id,
			URL:  urls[i],
			User: user,
		}

		ids = append(ids, id)

		if i < len(urls)-1 {
			mod.S.ID++
		}
	}

	return ids, nil
}

func (c *InMemory) Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if int(id) > mod.S.ID {
		return "", mod.ErrStorageIsNil
	}

	return mod.S.URLs[int(id)].URL, nil
}

func (c *InMemory) GetAll(user string) ([]mod.URLs, error) {
	var UserURLs []mod.URLs
	for _, i := range mod.S.URLs {
		if i.User == user {
			UserURLs = append(UserURLs, mod.URLs{
				ShortURL:    "http://" + c.ServerAddress + c.BaseURL + i.ID,
				OriginalURL: i.URL,
			})
		}
	}

	return UserURLs, nil
}