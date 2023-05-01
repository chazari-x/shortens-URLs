package indb

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
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
	var shortURL mod.Event

	err := c.DB.QueryRow(insertOnConflict, addURL, user).Scan(&shortURL.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}

		err = c.DB.QueryRow(selectIDWhereURL, addURL).Scan(&shortURL.ID, &shortURL.Del)
		if err != nil {
			return "", err
		}
	} else {
		mod.S.ID = shortURL.ID - 2
	}

	sID := strconv.FormatInt(int64(shortURL.ID-1), 36)

	if shortURL.ID-1 <= mod.S.ID && !shortURL.Del {
		return sID, mod.ErrURLConflict
	} else if shortURL.Del {
		_, err = c.DB.Exec(updateDelAndUserIDWhereID, shortURL.ID, false, user)
		if err != nil {
			return "", err
		}
		return sID, nil
	}

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

	var dbItem mod.Event

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
		var dbItem mod.Event
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

const workersCount = 5

func (c *InDB) BatchUpdate(ids []string, user string) error {
	tx, err := c.DB.Begin()
	if err != nil {
		log.Print(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	updateStmt, err := c.DB.Prepare(updateDelWhereIDAndUserID)
	if err != nil {
		log.Print(err)
	}

	txStmt := tx.Stmt(updateStmt)

	inputCh := make(chan string, len(ids))

	go func() {
		for _, r := range ids {
			inputCh <- r
		}

		close(inputCh)
	}()

	fanOutChs := fanOut(inputCh, workersCount)
	workerChs := make([]chan *sql.Stmt, 0, workersCount)
	for _, fanOutCh := range fanOutChs {
		workerCh := make(chan *sql.Stmt)
		newWorker(fanOutCh, workerCh, user, txStmt)
		workerChs = append(workerChs, workerCh)
	}

	for w := range fanIn(workerChs...) {
		txStmt = w
	}

	return tx.Commit()
}

func fanOut(inputCh chan string, n int) []chan string {
	chs := make([]chan string, 0, n)
	for i := 0; i < n; i++ {
		ch := make(chan string)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan string) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			id, ok := <-inputCh
			if !ok {
				return
			}

			ch := chs[i]
			ch <- id
		}
	}()

	return chs
}

func fanIn(inputChs ...chan *sql.Stmt) chan *sql.Stmt {
	outCh := make(chan *sql.Stmt)

	go func() {
		wg := &sync.WaitGroup{}

		for _, inputCh := range inputChs {
			wg.Add(1)

			go func(inputCh chan *sql.Stmt) {
				defer wg.Done()
				for item := range inputCh {
					outCh <- item
				}
			}(inputCh)
		}

		wg.Wait()
		close(outCh)
	}()

	return outCh
}

func newWorker(input chan string, out chan *sql.Stmt, user string, txStmt *sql.Stmt) {
	go func() {
		defer func() {
			if x := recover(); x != nil {
				newWorker(input, out, user, txStmt)
				log.Printf("run time panic: %v", x)
			}
		}()

		for sid := range input {
			id, err := strconv.ParseInt(sid, 36, 64)
			if err != nil {
				log.Print(err)
			}

			log.Printf("delete: %s, user: %s, id: %s, url: %s", "try", user, sid, mod.S.URLs[int(id)].URL)

			_, err = txStmt.Exec(id+1, user, true)
			if err != nil {
				log.Print(err)
			}

			out <- txStmt
		}

		close(out)
	}()
}
