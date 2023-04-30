package indb

import (
	"context"
	"database/sql"
	"errors"
	"log"
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

var (
	createTable = `CREATE TABLE IF NOT EXISTS shortURL (
						id 		SERIAL 	PRIMARY KEY NOT NULL, 
						url 	VARCHAR UNIQUE 		NOT NULL,
						del 	BOOLEAN 			NOT NULL 	DEFAULT false, 
						userID 	VARCHAR 			NOT NULL)`

	selectMaxID          = `SELECT MAX(id) FROM shortURL`
	selectIDWhereURL     = `SELECT id, del FROM shortURL WHERE url = $1`
	selectAllWhereID     = `SELECT * FROM shortURL WHERE id = $1`
	selectAllWhereUserID = `SELECT * FROM shortURL WHERE userID = $1`

	insertOnConflict = `INSERT INTO shortURL (url, userID) VALUES ($1, $2) ON CONFLICT(url) DO NOTHING RETURNING id`

	updateDelWhereIDAndUserID = `UPDATE shortURL SET del = $3 WHERE id = $1 AND userID = $2`
	updateDelAndUserIDWhereID = `UPDATE shortURL SET del = $2, userID = $3 WHERE id = $1`
)

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

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(selectMaxID).Scan(&mod.S.ID)

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
	var shortURL mod.ShortURL

	err := c.DB.QueryRow(insertOnConflict, addURL, user).Scan(&shortURL.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}

		log.Print(err)
		err = c.DB.QueryRow(selectIDWhereURL, addURL).Scan(&shortURL.ID, &shortURL.Del)
		if err != nil {
			return "", err
		}
	}

	sID := strconv.FormatInt(int64(shortURL.ID-1), 36)

	if shortURL.ID-1 <= mod.S.ID && !shortURL.Del {
		log.Print(sID, " ", mod.ErrURLConflict, " ", addURL)
		return sID, mod.ErrURLConflict
	} else if shortURL.Del {
		_, err = c.DB.Exec(updateDelAndUserIDWhereID, shortURL.ID, false, user)
		if err != nil {
			return "", err
		}
		log.Print(sID, " ", addURL)
		return sID, nil
	}

	log.Print(sID, " ", addURL)
	mod.S.ID = shortURL.ID - 1

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

	insertStmt, err := c.DB.Prepare(insertOnConflict)
	if err != nil {
		return nil, err
	}

	txStmt := tx.Stmt(insertStmt)

	for _, u := range urls {
		var id int
		err = txStmt.QueryRow(u, user).Scan(&id)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}

			err = c.DB.QueryRow(selectIDWhereURL, u).Scan(&id)
			if err != nil {
				return nil, err
			}
		}

		sID := strconv.FormatInt(int64(id-1), 36)

		if id > mod.S.ID {
			mod.S.ID = id - 1
		}

		ids = append(ids, sID)
	}

	return ids, tx.Commit()
}

func (c *InDB) Get(str string) (string, bool, error) {
	id, err := strconv.ParseInt(str, 36, 64)
	if err != nil {
		return "", false, err
	}

	if int(id) > mod.S.ID {
		return "", false, mod.ErrStorageIsNil
	}

	var dbItem mod.ShortURL

	err = c.DB.QueryRow(selectAllWhereID, id+1).Scan(&dbItem.ID, &dbItem.URL, &dbItem.Del, &dbItem.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, mod.ErrStorageIsNil
		}
		return "", false, err
	}

	return dbItem.URL, dbItem.Del, nil
}

func (c *InDB) GetAll(user string) ([]mod.URLs, error) {
	var UserURLs []mod.URLs

	rows, err := c.DB.Query(selectAllWhereUserID, user)
	if err != nil {

		return nil, err
	}

	for rows.Next() {
		var dbItem mod.ShortURL
		err = rows.Scan(&dbItem.ID, &dbItem.URL, &dbItem.Del, &dbItem.UserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, mod.ErrStorageIsNil
			}
			return nil, err
		}

		if !dbItem.Del {
			id := strconv.FormatInt(int64(dbItem.ID-1), 36)

			UserURLs = append(UserURLs, mod.URLs{
				ShortURL:    "http://" + c.ServerAddress + c.BaseURL + id,
				OriginalURL: dbItem.URL,
			})
		}
	}

	if rows.Err() != nil {
		return nil, err
	}

	return UserURLs, nil
}

func (c *InDB) BatchUpdate(ids []string, user string) error {
	tx, err := c.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	updateStmt, err := c.DB.Prepare(updateDelWhereIDAndUserID)
	if err != nil {
		return err
	}

	txStmt := tx.Stmt(updateStmt)

	for _, u := range ids {
		id, err := strconv.ParseInt(u, 36, 64)
		if err != nil {
			return err
		}

		_, err = txStmt.Exec(id+1, user, true)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
