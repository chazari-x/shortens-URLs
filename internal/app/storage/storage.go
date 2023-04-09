package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"main/internal/app/config"
)

type Storage interface {
	Add(url, user string) (string, error)
	Get(str string) (string, error)
	GetAll(user string) ([]URLs, error)
	PingDB(r *http.Request) error
}

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DataBaseDSN     string
	DB              *sql.DB
}

var s struct {
	URLs map[int]Event // Используется, если File не прописан
	ID   int           // Это ID последнего элемента в хранилище
}

type Event struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	User string `json:"user"`
}

type URLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

func StartStorage(conf config.Config) (*Config, error) {
	var c = &Config{
		ServerAddress:   conf.ServerAddress,
		BaseURL:         conf.BaseURL,
		FileStoragePath: conf.FileStoragePath,
		DataBaseDSN:     conf.DataBaseDSN,
		DB:              nil,
	}

	if c.DataBaseDSN != "" {
		db, err := c.startDataBase()
		if err != nil {
			return nil, err
		}
		c.DB = db
	} else if c.FileStoragePath != "" {
		err := c.startFileStorage()
		if err != nil {
			return nil, err
		}
	} else {
		s.URLs = make(map[int]Event)
	}

	s.ID = -1
	return c, nil
}

func (c *Config) startDataBase() (*sql.DB, error) {
	db, err := sql.Open("postgres", c.DataBaseDSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func (c *Config) PingDB(r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := c.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Config) startFileStorage() error {
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

		s.ID = int(id)
	}

	return nil
}

func (c *Config) Add(url, user string) (string, error) {
	var id string
	s.ID++

	if c.FileStoragePath == "" {
		id = strconv.FormatInt(int64(s.ID), 36)
		s.URLs[s.ID] = Event{
			ID:   id,
			URL:  url,
			User: user,
		}

		return id, nil
	}

	id = strconv.FormatInt(int64(s.ID), 36)

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

func (c *Config) Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if c.FileStoragePath == "" {
		if int(id) > s.ID {
			return "", fmt.Errorf("the storage is empty or the element is missing")
		}

		return s.URLs[int(id)].URL, nil
	}

	if int(id) > s.ID {
		return "", fmt.Errorf("the storage is empty or the element is missing")
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

	return "", fmt.Errorf("the storage is empty or the element is missing")
}

func (c *Config) GetAll(user string) ([]URLs, error) {
	var UserURLs []URLs
	if c.FileStoragePath == "" {
		for _, i := range s.URLs {
			if i.User == user {
				UserURLs = append(UserURLs, URLs{
					ShortURL:    "http://" + c.ServerAddress + c.BaseURL + i.ID,
					OriginalURL: i.URL,
				})
			}
		}

		return UserURLs, nil
	}

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
