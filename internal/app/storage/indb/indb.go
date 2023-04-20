package indb

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	mod "main/internal/app/storage/model"
)

type InDB struct {
	ServerAddress string
	BaseURL       string
	DataBaseDSN   string
	DB            *sql.DB
}

func (c *InDB) StartDataBase() (*sql.DB, error) {
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

	err = db.QueryRow("SELECT MAX(id) FROM shortURL").Scan(&mod.S.ID)
	if err != nil {
		if strings.Contains(err.Error(), "converting NULL to int is unsupported") {
			return db, nil
		}
		return nil, err
	}

	mod.S.ID--

	return db, nil
}

func (c *InDB) PingDB(cc context.Context) error {
	ctx, cancel := context.WithTimeout(cc, time.Second)
	defer cancel()

	if err := c.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (c *InDB) Add(addURL, user string) (string, error) {
	mod.S.ID++

	var id int

	err := c.DB.QueryRow(`INSERT INTO shortURL (url, userID) VALUES ($1, $2)
									ON CONFLICT(url) DO UPDATE SET url = $1 RETURNING id`, addURL, user).Scan(&id)
	if err != nil {
		return "", err
	}

	sID := strconv.FormatInt(int64(id), 36)

	if id < mod.S.ID {
		return sID, mod.ErrURLConflict
	}

	mod.S.ID = id

	return sID, nil
}

func (c *InDB) BatchAdd(urls []string, user string) ([]string, error) {
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

		sID := strconv.FormatInt(int64(id), 36)

		if id > mod.S.ID {
			mod.S.ID = id
		}

		ids = append(ids, sID)
	}

	return ids, tx.Commit()
}

func (c *InDB) Get(str string) (string, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", err
	}

	if int(id) > mod.S.ID {
		return "", mod.ErrStorageIsNil
	}

	var dbItem mod.ShortURL

	err = c.DB.QueryRow(`SELECT * FROM shortURL WHERE id = $1`, id).Scan(&dbItem.ID, &dbItem.URL, &dbItem.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", mod.ErrStorageIsNil
		}
		return "", err
	}

	return dbItem.URL, nil
}

func (c *InDB) GetAll(user string) ([]mod.URLs, error) {
	var UserURLs []mod.URLs

	rows, err := c.DB.Query(`SELECT * FROM shortURL WHERE userID = $1`, user)
	if err != nil {

		return nil, err
	}

	for rows.Next() {
		var dbItem mod.ShortURL
		err = rows.Scan(&dbItem.ID, &dbItem.URL, &dbItem.UserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, mod.ErrStorageIsNil
			}
			return nil, err
		}

		id := strconv.FormatInt(int64(dbItem.ID-1), 36)

		UserURLs = append(UserURLs, mod.URLs{
			ShortURL:    "http://" + c.ServerAddress + c.BaseURL + id,
			OriginalURL: dbItem.URL,
		})
	}

	if rows.Err() != nil {
		return nil, err
	}

	return UserURLs, nil
}
