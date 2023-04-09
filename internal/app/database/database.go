package database

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"main/internal/app/config"
)

type DB struct {
	DB *sql.DB
}

func StartDB(conf config.Config) (DB, error) {
	db, err := sql.Open("postgres", conf.DataBaseDSN)
	if err != nil {
		return DB{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return DB{}, err
	}

	return DB{DB: db}, nil
}

func (db *DB) PingDB(r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
