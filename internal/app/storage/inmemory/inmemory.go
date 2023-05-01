package inmemory

import (
	"context"
	"errors"
	"log"
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
		Del:    false,
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
			Del:    false,
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

	if !mod.S.URLs[int(id)].Del {
		return mod.S.URLs[int(id)].URL, false, nil
	}

	return "", true, nil
}

func (c *InMemory) GetAll(user string) ([]mod.URLs, error) {
	var UserURLs []mod.URLs
	for _, i := range mod.S.URLs {
		if i.UserID == user && !i.Del {
			id := strconv.FormatInt(int64(i.ID), 36)
			UserURLs = append(UserURLs, mod.URLs{
				ShortURL:    "http://" + c.ServerAddress + c.BaseURL + id,
				OriginalURL: i.URL,
			})
		}
	}

	return UserURLs, nil
}

const workersCount = 5

func (c *InMemory) BatchUpdate(ids []string, user string) error {
	inputCh := make(chan string, len(ids))

	go func() {
		for _, r := range ids {
			inputCh <- r
		}

		close(inputCh)
	}()

	fanOutChs := fanOut(inputCh, workersCount)
	for _, fanOutCh := range fanOutChs {
		newWorker(fanOutCh, user)
	}

	return nil
}

func fanOut(inputCh chan string, n int) []chan string {
	chs := make([]chan string, 0, n)
	for i := 0; i < n; i++ {
		ch := make(chan string)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan string) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			id, ok := <-inputCh
			if !ok {
				return
			}

			ch := chs[i]
			ch <- id
		}
	}()

	return chs
}

func newWorker(input chan string, user string) {
	go func() {
		defer func() {
			if x := recover(); x != nil {
				newWorker(input, user)
				log.Printf("run time panic: %v", x)
			}
		}()

		for sid := range input {
			id, err := strconv.ParseInt(sid, 36, 64)
			if err != nil {
				log.Print(err)
			}

			ok := mod.S.URLs[int(id)].UserID == user && !mod.S.URLs[int(id)].Del
			log.Printf("delete: %5s, user: %s, id: %s, url: %s", strconv.FormatBool(ok), user, sid, mod.S.URLs[int(id)].URL)
			if ok {
				mod.S.URLs[int(id)] = mod.Event{
					ID:     int(id),
					URL:    mod.S.URLs[int(id)].URL,
					Del:    true,
					UserID: user,
				}
			}
		}
	}()
}
