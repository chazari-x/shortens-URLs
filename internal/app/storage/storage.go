package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"main/internal/app/config"
)

var S struct {
	file string   // Путь до файла хранилища
	URLs []string // Массив URL'ов. Используется, если file не прописан
	ID   int      // Это ID следующего добавляемого элемента в хранилище
}

type Event struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}
func (p *Producer) WriteEvent(event *Event) error {
	return p.encoder.Encode(&event)
}
func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) (*Consumer, error) {
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

func StartStorage() error {
	S.file = config.Conf.FileStoragePath

	consumer, err := NewConsumer(S.file)
	if err != nil {
		log.Fatal(err)
	}
	defer func(consumer *Consumer) {
		err := consumer.Close()
		if err != nil {
			log.Print(err)
		}
	}(consumer)

	for i := 0; ; i++ {
		readEvent, err := consumer.ReadEvent()
		if readEvent == nil {
			break
		} else if err != nil {
			log.Print(err)
		}

		S.ID++
	}

	return nil
}

func Add(url string) (string, error) {
	var id string

	if S.file != "" {
		id = strconv.FormatInt(int64(S.ID), 36)
		S.ID++

		producer, err := NewProducer(S.file)
		if err != nil {
			log.Fatal(err)
		}
		defer func(producer *Producer) {
			err := producer.Close()
			if err != nil {
				log.Print(err)
			}
		}(producer)

		err = producer.WriteEvent(&Event{
			ID:  id,
			URL: url,
		})
		if err != nil {
			return "", err
		}
	} else {
		id = strconv.FormatInt(int64(len(S.URLs)), 36)
		S.URLs = append(S.URLs, url)
	}

	return id, nil
}

func Get(s string) (string, error) {
	id, err := strconv.ParseInt(s, 36, 64)
	if err != nil {
		return "", err
	}

	if S.file != "" {
		if int(id) >= S.ID {
			return "", fmt.Errorf("the storage is empty or the element is missing")
		}

		consumer, err := NewConsumer(S.file)
		if err != nil {
			log.Fatal(err)
		}
		defer func(consumer *Consumer) {
			err := consumer.Close()
			if err != nil {
				log.Print(err)
			}
		}(consumer)

		for i := 0; i <= S.ID; i++ {
			readEvent, err := consumer.ReadEvent()
			if err != nil {
				log.Print(err)
			}

			if readEvent.ID == s {
				return readEvent.URL, nil
			}
		}
	} else {
		if int(id) >= len(S.URLs) {
			return "", fmt.Errorf("the storage is empty or the element is missing")
		}

		return S.URLs[int(id)], nil
	}

	return "", fmt.Errorf("the storage is empty or the element is missing")
}
