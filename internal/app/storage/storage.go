package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

var s struct {
	File string   // Путь до файла хранилища
	URLs []string // Массив URL'ов. Используется, если File не прописан
	ID   int      // Это ID следующего добавляемого элемента в хранилище
}

type Event struct {
	ID  string `json:"id"`
	URL string `json:"url"`
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

func (p *producer) WriteEvent(event *Event) error {
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

func StartStorage(FileStoragePath string) error {
	s.File = FileStoragePath

	consumer, err := newConsumer(s.File)
	if err != nil {
		return err
	}
	defer func() {
		err := consumer.Close()
		if err != nil {
			log.Print("consumer close err: ", err)
		}
	}()

	for i := 0; ; i++ {
		readEvent, err := consumer.ReadEvent()
		if readEvent == nil {
			break
		} else if err != nil {
			return err
		}

		s.ID++
	}

	return nil
}

func Add(url string) (string, error) {
	var id string

	if s.File == "" {
		id = strconv.FormatInt(int64(len(s.URLs)), 36)
		s.URLs = append(s.URLs, url)

		return id, nil
	}

	id = strconv.FormatInt(int64(s.ID), 36)
	s.ID++

	producer, err := newProducer(s.File)
	if err != nil {
		return "", err
	}
	defer func() {
		err := producer.Close()
		if err != nil {
			log.Print("producer close err: ", err)
		}
	}()

	err = producer.WriteEvent(&Event{
		ID:  id,
		URL: url,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

func Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if s.File == "" {
		if int(id) >= len(s.URLs) {
			return "", fmt.Errorf("the storage is empty or the element is missing")
		}

		return s.URLs[int(id)], nil
	}

	if int(id) >= s.ID {
		return "", fmt.Errorf("the storage is empty or the element is missing")
	}

	consumer, err := newConsumer(s.File)
	if err != nil {
		return "", err
	}
	defer func() {
		err := consumer.Close()
		if err != nil {
			log.Print("consumer close err: ", err)
		}
	}()

	for i := 0; i <= s.ID; i++ {
		readEvent, err := consumer.ReadEvent()
		if err != nil {
			return "", err
		}

		if readEvent.ID == str {
			return readEvent.URL, nil
		}
	}

	return "", fmt.Errorf("the storage is empty or the element is missing")
}
