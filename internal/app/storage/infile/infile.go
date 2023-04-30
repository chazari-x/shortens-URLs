package infile

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	mod "main/internal/app/storage/model"
)

type InFile struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

func newProducer(fileName string) (*producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *producer) WriteEvent(event mod.Event) error {
	return p.encoder.Encode(&event)
}

func (p *producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func newConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*mod.Event, error) {
	event := &mod.Event{}
	if err := c.decoder.Decode(&event); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (c *InFile) PingDB(_ context.Context) error {
	return errors.New("db is disabled")
}

func (c *InFile) StartFileStorage() error {
	consumer, err := newConsumer(c.FileStoragePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = consumer.Close()
	}()

	maxID := -1
	for i := 0; ; i++ {
		readEvent, err := consumer.ReadEvent()
		if readEvent == nil {
			break
		} else if err != nil {
			return err
		}

		maxID = readEvent.ID
	}

	if maxID != -1 {
		//id, err := strconv.ParseInt(maxID, 36, 64)
		//if err != nil {
		//	return err
		//}

		mod.S.ID = maxID
	}

	return nil
}

func (c *InFile) Add(url, user string) (string, error) {
	mod.S.ID++

	id := strconv.FormatInt(int64(mod.S.ID), 36)

	producer, err := newProducer(c.FileStoragePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = producer.Close()
	}()

	err = producer.WriteEvent(mod.Event{
		ID:     mod.S.ID,
		URL:    url,
		UserID: user,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *InFile) BatchAdd(urls []string, user string) ([]string, error) {
	mod.S.ID++

	var ids []string

	producer, err := newProducer(c.FileStoragePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = producer.Close()
	}()

	for i := 0; i < len(urls); i++ {
		id := strconv.FormatInt(int64(mod.S.ID), 36)
		err = producer.WriteEvent(mod.Event{
			ID:     mod.S.ID,
			URL:    urls[i],
			UserID: user,
		})
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)

		if i < len(urls)-1 {
			mod.S.ID++
		}
	}

	return ids, nil
}

func (c *InFile) Get(str string) (string, bool, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", false, err
	}

	if int(id) > mod.S.ID {
		return "", false, mod.ErrStorageIsNil
	}

	consumer, err := newConsumer(c.FileStoragePath)
	if err != nil {
		return "", false, err
	}
	defer func() {
		_ = consumer.Close()
	}()

	for i := 0; ; i++ {
		readEvent, err := consumer.ReadEvent()
		if readEvent == nil {
			break
		} else if err != nil {
			return "", false, err
		}

		eid := strconv.FormatInt(int64(readEvent.ID), 36)
		if eid == str {
			return readEvent.URL, false, nil
		}
	}

	return "", false, mod.ErrStorageIsNil
}

func (c *InFile) GetAll(user string) ([]mod.URLs, error) {
	var UserURLs []mod.URLs
	consumer, err := newConsumer(c.FileStoragePath)
	if err != nil {
		return UserURLs, err
	}
	defer func() {
		_ = consumer.Close()
	}()

	for i := 0; ; i++ {
		readEvent, err := consumer.ReadEvent()
		if readEvent == nil {
			break
		} else if err != nil {
			return UserURLs, err
		}

		if readEvent.UserID == user {
			id := strconv.FormatInt(int64(readEvent.ID), 36)
			UserURLs = append(UserURLs, mod.URLs{
				ShortURL:    "http://" + c.ServerAddress + c.BaseURL + id,
				OriginalURL: readEvent.URL,
			})
		}
	}

	return UserURLs, nil
}

func (c *InFile) BatchUpdate(_ []string, _ string) error {
	return errors.New("db is disabled")
}
