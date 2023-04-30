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

func (c *InMemory) BatchUpdate(ids []string, user string) {
	//res := fanOut(&mod.S.URLs, ids, user)

	for _, r := range ids {
		id, err := strconv.ParseInt(r, 36, 64)
		if err != nil {
			log.Print(err)
		}

		if mod.S.URLs[int(id)].UserID == user {
			mod.S.URLs[int(id)] = mod.Event{
				ID:     int(id),
				URL:    mod.S.URLs[int(id)].URL,
				Del:    true,
				UserID: user,
			}
		}
	}
}

//func fanOut(m *map[int]mod.Event, ids []string, user string) []string {
//	var res []string
//	//stopCh := make(chan struct{})
//	resCh := make(chan string)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//	go func(resCh chan string, res []string, wg *sync.WaitGroup) {
//		for r := range resCh {
//			res = append(res, r)
//		}
//		wg.Done()
//	}(resCh, res, wg)
//
//	startFanOut(m, ids, user, resCh)
//
//	wg.Wait()
//	close(resCh)
//
//	log.Print(res)
//	return res
//}
//
//func startFanOut(m *map[int]mod.Event, ids []string, user string, resCh chan<- string) {
//	wg := &sync.WaitGroup{}
//	wg.Add(len(ids))
//	for _, id := range ids {
//		//go workFanOut(*m, id, user, resCh)
//		go func(m map[int]mod.Event, id string, wg *sync.WaitGroup) {
//			i, err := strconv.ParseInt(id, 36, 64)
//			if err != nil {
//				wg.Done()
//				return
//			}
//
//			log.Print(id, " ", user, " ", int(i) < len(m) && m[int(i)].UserID == user)
//			if m[int(i)].UserID == user {
//				resCh <- id
//			}
//
//			wg.Done()
//		}(*m, id, wg)
//	}
//	wg.Wait()
//}
//
////func workFanOut(m map[int]mod.Event, id, user string, resCh chan<- string) {
////	go func() {
////		log.Print(id, " ", user)
////		i, err := strconv.ParseInt(id, 36, 64)
////		if err != nil {
////			return
////		}
////
////		if m[int(i)].UserID == user {
////			resCh <- id
////		}
////	}()
////}
