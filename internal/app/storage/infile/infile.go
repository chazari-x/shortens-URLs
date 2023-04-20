package infile

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	. "main/internal/app/storage/model"
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

func (p *producer) WriteEvent(event Event) error {
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

func (c *Consumer) ReadEvent() (*Event, error) {
	event := &Event{}
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

	maxID := "-1"
	for i := 0; ; i++ {
		readEvent, err := consumer.ReadEvent()
		if readEvent == nil {
			break
		} else if err != nil {
			return err
		}

		maxID = readEvent.ID
	}

	if maxID != "-1" {
		id, err := strconv.ParseInt(maxID, 36, 64)
		if err != nil {
			return err
		}

		S.ID = int(id)
	}

	return nil
}

func (c *InFile) Add(url, user string) (string, error) {
	S.ID++

	id := strconv.FormatInt(int64(S.ID), 36)

	producer, err := newProducer(c.FileStoragePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = producer.Close()
	}()

	err = producer.WriteEvent(Event{
		ID:   id,
		URL:  url,
		User: user,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *InFile) BatchAdd(urls []string, user string) ([]string, error) {
	S.ID++

	var ids []string

	producer, err := newProducer(c.FileStoragePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = producer.Close()
	}()

	for i := 0; i < len(urls); i++ {
		id := strconv.FormatInt(int64(S.ID), 36)
		err = producer.WriteEvent(Event{
			ID:   id,
			URL:  urls[i],
			User: user,
		})
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)

		if i < len(urls)-1 {
			S.ID++
		}
	}

	return ids, nil
}

func (c *InFile) Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if int(id) > S.ID {
		return "", ErrStorageIsNil
	}

	consumer, err := newConsumer(c.FileStoragePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = consumer.Close()
	}()

	for i := 0; ; i++ {
		readEvent, err := consumer.ReadEvent()
		if readEvent == nil {
			break
		} else if err != nil {
			return "", err
		}

		if readEvent.ID == str {
			return readEvent.URL, nil
		}
	}

	return "", ErrStorageIsNil
}

func (c *InFile) GetAll(user string) ([]URLs, error) {
	var UserURLs []URLs
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

		if readEvent.User == user {
			UserURLs = append(UserURLs, URLs{
				ShortURL:    "http://" + c.ServerAddress + c.BaseURL + readEvent.ID,
				OriginalURL: readEvent.URL,
			})
		}
	}

	return UserURLs, nil
}
