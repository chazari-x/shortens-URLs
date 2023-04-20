package storage

import (
	"context"

	_ "github.com/lib/pq"
	"main/internal/app/config"
	. "main/internal/app/storage/indb"
	. "main/internal/app/storage/infile"
	. "main/internal/app/storage/inmemory"
	. "main/internal/app/storage/model"
)

type Storage interface {
	Add(url, user string) (string, error)
	BatchAdd(urls []string, user string) ([]string, error)
	Get(str string) (string, error)
	GetAll(user string) ([]URLs, error)
	PingDB(cc context.Context) error
}

func StartStorage(conf config.Config) (*InMemory, *InFile, *InDB, error) {
	S.ID = -1

	if conf.DataBaseDSN != "" {
		var c = &InDB{
			ServerAddress: conf.ServerAddress,
			BaseURL:       conf.BaseURL,
			DataBaseDSN:   conf.DataBaseDSN,
			DB:            nil,
		}

		db, err := c.StartDataBase()
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

		err := c.StartFileStorage()
		if err != nil {
			return nil, nil, nil, err
		}

		return nil, c, nil, nil
	}

	S.URLs = make(map[int]Event)

	var c = &InMemory{
		ServerAddress: conf.ServerAddress,
		BaseURL:       conf.BaseURL,
	}

	return c, nil, nil, nil
}
