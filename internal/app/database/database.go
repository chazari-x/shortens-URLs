package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"main/internal/app/config"
)

type DB struct {
	DB *sql.DB
}

func StartDB(conf config.Config) (*DB, error) {
	db, err := sql.Open("postgres", conf.DataBaseDSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

func (db *DB) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
