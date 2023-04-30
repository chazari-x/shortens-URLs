package storage

import (
	"context"

	_ "github.com/lib/pq"
	"main/internal/app/config"
	d "main/internal/app/storage/indb"
	f "main/internal/app/storage/infile"
	m "main/internal/app/storage/inmemory"
	mod "main/internal/app/storage/model"
)

type Storage interface {
	Add(url, user string) (string, error)
	BatchAdd(urls []string, user string) ([]string, error)
	BatchUpdate(ids []string, user string) error
	Get(str string) (string, bool, error)
	GetAll(user string) ([]mod.URLs, error)
	PingDB(cc context.Context) error
}

func StartStorage(conf config.Config) (*m.InMemory, *f.InFile, *d.InDB, error) {
	mod.S.ID = -1

	if conf.DataBaseDSN != "" {
		var c = &d.InDB{
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
		var c = &f.InFile{
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

	mod.S.URLs = make(map[int]mod.Event)

	var c = &m.InMemory{
		ServerAddress: conf.ServerAddress,
		BaseURL:       conf.BaseURL,
	}

	return c, nil, nil, nil
}
