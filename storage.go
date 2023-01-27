package certmagicsqlite3

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/caddyserver/certmagic"
	_ "modernc.org/sqlite"
)

type storage struct {
	db *sql.DB
}

func OpenSQLiteStorage(dataSourceName string) (storage, error) {
	if dataSourceName == "" {
		return storage{}, errors.New("data source cannot be empty")
	}

	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return storage{}, err
	}

	for _, stmt := range []string{pragmaWALEnabled, pragma500BusyTimeout, pragmaCaseSenstive} {
		_, err = db.Exec(stmt, nil)
		if err != nil {
			return storage{}, err
		}
	}

	_, err = db.Exec(createTable)
	if err != nil {
		return storage{}, err
	}

	s := storage{
		db: db,
	}
	return s, nil

}

func (s *storage) Store(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if len(value) == 0 {
		return errors.New("value cannot be empty")
	}
	stmt, err := s.db.Prepare(insertKey)
	if err != nil {
		return err
	}

	t := time.Now()
	_, err = stmt.Exec(key, value, t.Format(time.Layout), len(value))
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) Load(ctx context.Context, key string) ([]byte, error) {
	rows, err := s.db.Query(getKey, key)
	if err != nil {
		return nil, err
	}

	value := []byte{}
	for rows.Next() {
		err = rows.Scan(&value)
		if err != nil {
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *storage) Exists(ctx context.Context, key string) bool {
	rows, err := s.db.Query(getKey, key)
	if err != nil {
		return false
	}
	if rows.Next() {
		return true
	}
	return false
}

func (s *storage) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	stmt, err := s.db.Prepare(deleteKey)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(key)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	rows, err := s.db.Query(listKey, prefix)
	if err != nil {
		return []string{}, err
	}

	keyList := []string{}
	for rows.Next() {
		var (
			key string
		)
		err = rows.Scan(&key)
		if err != nil {
			return keyList, err
		}
		keyList = append(keyList, key)
	}

	if err = rows.Err(); err != nil {
		return keyList, err
	}
	return keyList, nil
}

func (s *storage) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
	rows, err := s.db.Query(statKey, key)
	if err != nil {
		return certmagic.KeyInfo{}, err
	}

	keyInfo := certmagic.KeyInfo{}
	for rows.Next() {
		var (
			modifiedString string
			modified       time.Time
			size           int64
		)
		err = rows.Scan(&modifiedString, &size)
		if err != nil {
			return certmagic.KeyInfo{}, err
		}

		modified, err = time.Parse(time.Layout, modifiedString)
		if err != nil {
			return certmagic.KeyInfo{}, err
		}
		keyInfo.Key = key
		keyInfo.Modified = modified
		keyInfo.Size = size
	}

	if err = rows.Err(); err != nil {
		return certmagic.KeyInfo{}, err
	}
	return keyInfo, nil
}

const createTable = `
CREATE TABLE IF NOT EXISTS certmagic(
	key TEXT NOT NULL PRIMARY KEY,
	value BLOB NOT NULL,
	modified TEXT NOT NULL,
	size INTEGER NOT NULL
	);`

const pragmaWALEnabled = `PRAGMA journal_mode = WAL;`
const pragma500BusyTimeout = `PRAGMA busy_timeout = 5000;`
const pragmaCaseSenstive = `PRAGMA case_sensitive_like = true;`

const insertKey = `INSERT OR REPLACE INTO certmagic(key, value, modified, size) VALUES (?, ?, ?, ?);`

const getKey = `SELECT value from certmagic WHERE key = ?`

const listKey = `SELECT key from certmagic WHERE key LIKE ? ||  '%' `

const statKey = `SELECT modified,size from certmagic WHERE key = ?`

const deleteKey = `DELETE FROM certmagic WHERE key=?`
