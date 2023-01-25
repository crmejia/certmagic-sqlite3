package certmagicsqlite3

import (
	"context"
	"database/sql"
	"errors"

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

	for _, stmt := range []string{pragmaWALEnabled, pragma500BusyTimeout } {
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

	_, err = stmt.Exec(key, value)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) Load(ctx context.Context, key string) ([]byte, error) {
	rows, err := s.db.Query(getKEY, key)
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

func (s *storage) Exists(ctx context.Context, key string) bool{
	rows, err := s.db.Query(getKEY, key)
	if err != nil {
		return false
	}
	if rows.Next() {
		return true
	}
	return false
}

func (s *storage) Delete(ctx context.Context, key string) error{
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

const createTable = `
CREATE TABLE IF NOT EXISTS certmagic(
	key TEXT NOT NULL PRIMARY KEY,
	value BLOB NOT NULL);`

const pragmaWALEnabled = `PRAGMA journal_mode = WAL;`
const pragma500BusyTimeout = `PRAGMA busy_timeout = 5000;`

const insertKey = `INSERT into certmagic(key, value) VALUES (?, ?);`

const getKEY = `SELECT value from certmagic WHERE key = ?`

const deleteKey = `DELETE FROM certmagic WHERE key=?`
