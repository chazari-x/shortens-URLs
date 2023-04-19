package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"main/internal/app/config"
)

//type Storage interface {
//	Add(url, user string) (string, error)
//	Get(str string) (string, error)
//	GetAll(user string) ([]URLs, error)
//	PingDB(ctx context.Context) error
//	BatchAdd(url []string, user string) ([]string, error)
//}

type Storage interface {
	Add(url, user string) (string, error)
	BatchAdd(urls []string, user string) ([]string, error)
	Get(str string) (string, error)
	GetAll(user string) ([]URLs, error)
	PingDB(cc context.Context) error
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

var (
	ErrURLConflict  = errors.New("url conflict")
	ErrStorageIsNil = errors.New("the storage is empty or the element is missing")
)

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

type InMemory struct {
	ServerAddress string
	BaseURL       string
}

type InFile struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

type InDB struct {
	ServerAddress string
	BaseURL       string
	DataBaseDSN   string
	DB            *sql.DB
}

func StartStorage(conf config.Config) (*InMemory, *InFile, *InDB, error) {
	s.ID = -1

	if conf.DataBaseDSN != "" {
		var c = &InDB{
			ServerAddress: conf.ServerAddress,
			BaseURL:       conf.BaseURL,
			DataBaseDSN:   conf.DataBaseDSN,
			DB:            nil,
		}

		db, err := c.startDataBase()
		if err != nil {
			return nil, nil, nil, err
		}
		c.DB = db

		return nil, nil, c, nil
	} else if conf.FileStoragePath != "" {
		var c = &InFile{
			ServerAddress:   conf.ServerAddress,
			BaseURL:         conf.BaseURL,
			FileStoragePath: conf.FileStoragePath,
		}

		err := c.startFileStorage()
		if err != nil {
			return nil, nil, nil, err
		}

		return nil, c, nil, nil
	}

	s.URLs = make(map[int]Event)

	var c = &InMemory{
		ServerAddress: conf.ServerAddress,
		BaseURL:       conf.BaseURL,
	}

	return c, nil, nil, nil
}

func (c *InDB) startDataBase() (*sql.DB, error) {
	db, err := sql.Open("postgres", c.DataBaseDSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS shortURL (
								id SERIAL PRIMARY KEY NOT NULL, 
								url VARCHAR UNIQUE NOT NULL, 
								userID VARCHAR NOT NULL)`)
	if err != nil {
		return nil, err
	}

	err = db.QueryRow("SELECT MAX(id) FROM shortURL").Scan(&s.ID)
	if err != nil {
		if strings.Contains(err.Error(), "converting NULL to int is unsupported") {
			return db, nil
		}
		return nil, err
	}

	s.ID--

	return db, nil
}

func (c *InMemory) PingDB(_ context.Context) error {
	return nil
}

func (c *InFile) PingDB(_ context.Context) error {
	return nil
}

func (c *InDB) PingDB(cc context.Context) error {
	ctx, cancel := context.WithTimeout(cc, time.Second)
	defer cancel()

	if err := c.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (c *InFile) startFileStorage() error {
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

func (c *InMemory) Add(url, user string) (string, error) {
	s.ID++

	id := strconv.FormatInt(int64(s.ID), 36)
	s.URLs[s.ID] = Event{
		ID:   id,
		URL:  url,
		User: user,
	}

	return id, nil
}

func (c *InFile) Add(url, user string) (string, error) {
	s.ID++

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

func (c *InDB) Add(addURL, user string) (string, error) {
	s.ID++

	var id int

	err := c.DB.QueryRow(`INSERT INTO shortURL (url, userID) VALUES ($1, $2)
									ON CONFLICT(url) DO UPDATE SET url = $1 RETURNING id`, addURL, user).Scan(&id)
	if err != nil {
		return "", err
	}

	sID := strconv.FormatInt(int64(id-1), 36)

	if id-1 <= s.ID {
		return sID, ErrURLConflict
	}

	s.ID = id - 1

	return sID, nil
}

func (c *InMemory) BatchAdd(urls []string, user string) ([]string, error) {
	s.ID++

	var ids []string

	for i := 0; i < len(urls); i++ {
		id := strconv.FormatInt(int64(s.ID), 36)
		s.URLs[s.ID] = Event{
			ID:   id,
			URL:  urls[i],
			User: user,
		}

		ids = append(ids, id)

		if i < len(urls)-1 {
			s.ID++
		}
	}

	return ids, nil
}

func (c *InFile) BatchAdd(urls []string, user string) ([]string, error) {
	s.ID++

	var ids []string

	producer, err := newProducer(c.FileStoragePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = producer.Close()
	}()

	for i := 0; i < len(urls); i++ {
		id := strconv.FormatInt(int64(s.ID), 36)
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
			s.ID++
		}
	}

	return ids, nil
}

func (c *InDB) BatchAdd(urls []string, user string) ([]string, error) {
	s.ID++

	var ids []string

	tx, err := c.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	insertStmt, err := c.DB.Prepare(`INSERT INTO shortURL (url, userID) VALUES ($1, $2) 
												ON CONFLICT(url) DO UPDATE SET url = $1 RETURNING id`)
	if err != nil {
		return nil, err
	}

	txStmt := tx.Stmt(insertStmt)

	for _, u := range urls {
		var id int
		err = txStmt.QueryRow(u, user).Scan(&id)
		if err != nil {
			return nil, err
		}

		sID := strconv.FormatInt(int64(id-1), 36)

		ids = append(ids, sID)
	}

	return ids, tx.Commit()
}

func (c *InMemory) Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if int(id) > s.ID {
		return "", ErrStorageIsNil
	}

	return s.URLs[int(id)].URL, nil
}

func (c *InFile) Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if int(id) > s.ID {
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

func (c *InDB) Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if int(id) > s.ID {
		return "", ErrStorageIsNil
	}

	var dbItem shortURL

	err = c.DB.QueryRow("SELECT * FROM shortURL WHERE id = $1", id).Scan(&dbItem.ID, &dbItem.URL, &dbItem.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrStorageIsNil
		}
		return "", err
	}

	return dbItem.URL, nil
}

func (c *InMemory) GetAll(user string) ([]URLs, error) {
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

func (c *InDB) GetAll(user string) ([]URLs, error) {
	var UserURLs []URLs

	rows, err := c.DB.Query("SELECT * FROM shortURL WHERE userID = $1", user)
	if err != nil {

		return nil, err
	}

	for rows.Next() {
		var dbItem shortURL
		err = rows.Scan(&dbItem.ID, &dbItem.URL, &dbItem.UserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrStorageIsNil
			}
			return nil, err
		}

		id := strconv.FormatInt(int64(dbItem.ID-1), 36)

		UserURLs = append(UserURLs, URLs{
			ShortURL:    "http://" + c.ServerAddress + c.BaseURL + id,
			OriginalURL: dbItem.URL,
		})
	}

	if rows.Err() != nil {
		return nil, err
	}

	return UserURLs, nil
}
