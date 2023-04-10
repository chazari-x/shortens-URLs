package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
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

type shortURL struct {
	ID     int
	URL    string
	UserID string
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

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS shortURL (" +
		"id SERIAL PRIMARY KEY NOT NULL, " +
		"url varchar NOT NULL, " +
		"userID varchar NOT NULL)")
	if err != nil {
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
	var err error
	s.ID++

	if c.DataBaseDSN != "" {
		id, err = c.dbAdd(url, user)
	} else if c.FileStoragePath != "" {
		id, err = c.fileAdd(url, user)
	} else {
		id, err = c.memoryAdd(url, user)
	}

	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *Config) memoryAdd(url, user string) (string, error) {
	id := strconv.FormatInt(int64(s.ID), 36)
	s.URLs[s.ID] = Event{
		ID:   id,
		URL:  url,
		User: user,
	}

	return id, nil
}

func (c *Config) fileAdd(url, user string) (string, error) {
	id := strconv.FormatInt(int64(s.ID), 36)

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

func (c *Config) dbAdd(addURL, user string) (string, error) {
	var id int

	err := c.DB.QueryRow("INSERT INTO shortURL (url, userID) VALUES ($1, $2) RETURNING id", addURL, user).Scan(&id)
	if err != nil {
		return "", err
	}

	sID := strconv.FormatInt(int64(id-1), 36)

	return sID, nil
}

func (c *Config) Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if c.DataBaseDSN == "" {
		if int(id) > s.ID {
			return "", fmt.Errorf("the storage is empty or the element is missing")
		}
	}

	var url string
	if c.DataBaseDSN != "" {
		url, err = c.dbGet(int(id) + 1)
	} else if c.FileStoragePath != "" {
		url, err = c.fileGet(str)
	} else {
		url, err = c.memoryGet(int(id))
	}

	if err != nil {
		return "", err
	}

	return url, nil
}

func (c *Config) memoryGet(id int) (string, error) {
	return s.URLs[id].URL, nil
}

func (c *Config) fileGet(str string) (string, error) {
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

func (c *Config) dbGet(id int) (string, error) {
	var dbItem shortURL

	err := c.DB.QueryRow("SELECT * FROM shortURL WHERE id = $1", id).Scan(&dbItem.ID, &dbItem.URL, &dbItem.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return "", fmt.Errorf("the storage is empty or the element is missing")
		}
		return "", err
	}

	return dbItem.URL, nil
}

func (c *Config) GetAll(user string) ([]URLs, error) {
	var userURLs []URLs
	var err error

	if c.DataBaseDSN != "" {
		userURLs, err = c.dbGetAll(user)
	} else if c.FileStoragePath != "" {
		userURLs, err = c.fileGetAll(user)
	} else {
		userURLs, err = c.memoryGetAll(user)
	}

	if err != nil {
		return nil, err
	}

	return userURLs, nil
}

func (c *Config) memoryGetAll(user string) ([]URLs, error) {
	var UserURLs []URLs
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

func (c *Config) fileGetAll(user string) ([]URLs, error) {
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

func (c *Config) dbGetAll(user string) ([]URLs, error) {
	var UserURLs []URLs

	rows, err := c.DB.Query("SELECT * FROM shortURL WHERE userID = $1", user)
	if err != nil {

		return nil, err
	}

	for rows.Next() {
		var dbItem shortURL
		err = rows.Scan(&dbItem.ID, &dbItem.URL, &dbItem.UserID)
		if err != nil {
			if strings.Contains(err.Error(), "no rows in result set") {
				return nil, fmt.Errorf("the storage is empty or the element is missing")
			}
			return nil, err
		}
		UserURLs = append(UserURLs, URLs{
			ShortURL:    "http://" + c.ServerAddress + c.BaseURL + strconv.Itoa(dbItem.ID),
			OriginalURL: dbItem.URL,
		})
	}

	if rows.Err() != nil {
		return nil, err
	}

	return UserURLs, nil
}
