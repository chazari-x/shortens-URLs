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
		ID:     mod.S.ID,
		URL:    url,
		UserID: user,
	}

	return id, nil
}

func (c *InMemory) BatchAdd(urls []string, user string) ([]string, error) {
	mod.S.ID++

	var ids []string

	for i := 0; i < len(urls); i++ {
		id := strconv.FormatInt(int64(mod.S.ID), 36)
		mod.S.URLs[mod.S.ID] = mod.Event{
			ID:     mod.S.ID,
			URL:    urls[i],
			UserID: user,
		}

		ids = append(ids, id)

		if i < len(urls)-1 {
			mod.S.ID++
		}
	}

	return ids, nil
}

func (c *InMemory) Get(str string) (string, bool, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", false, err
	}

	if int(id) > mod.S.ID {
		return "", false, mod.ErrStorageIsNil
	}

	return mod.S.URLs[int(id)].URL, false, nil
}

func (c *InMemory) GetAll(user string) ([]mod.URLs, error) {
	var UserURLs []mod.URLs
	for _, i := range mod.S.URLs {
		if i.UserID == user {
			id := strconv.FormatInt(int64(i.ID), 36)
			UserURLs = append(UserURLs, mod.URLs{
				ShortURL:    "http://" + c.ServerAddress + c.BaseURL + id,
				OriginalURL: i.URL,
			})
		}
	}

	return UserURLs, nil
}

func (c *InMemory) BatchUpdate(_ []string, _ string) error {
	return errors.New("db is disabled")
}
