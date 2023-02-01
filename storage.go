package certmagicsqlite3

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/certmagic"
	_ "modernc.org/sqlite"
)

type Storage struct {
	DataSourceName        string `json:"datasourcename,omitempty"`
	db                    *sql.DB
	lockExpirationTimeOut time.Duration
}

func OpenSQLiteStorage(dataSourceName string) (Storage, error) {
	if dataSourceName == "" {
		return Storage{}, errors.New("data source cannot be empty")
	}

	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return Storage{}, err
	}

	for _, stmt := range []string{pragmaWALEnabled, pragma500BusyTimeout, pragmaCaseSenstive} {
		_, err = db.Exec(stmt, nil)
		if err != nil {
			return Storage{}, err
		}
	}

	_, err = db.Exec(createTable)
	if err != nil {
		return Storage{}, err
	}

	s := Storage{
		db:                    db,
		lockExpirationTimeOut: defaultLockTimeOut,
	}
	return s, nil

}

func (s *Storage) Store(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	storeContext, cancel := context.WithCancel(ctx)
	defer cancel()

	if len(value) == 0 {
		return errors.New("value cannot be empty")
	}
	stmt, err := s.db.Prepare(insertKey)
	if err != nil {
		return err
	}

	t := time.Now()
	_, err = stmt.ExecContext(storeContext, key, value, t.Format(time.Layout), len(value))
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) Load(ctx context.Context, key string) ([]byte, error) {
	loadContext, cancel := context.WithCancel(ctx)
	defer cancel()

	value := []byte{}
	if err := s.db.QueryRowContext(loadContext, getKey, key).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}
	return value, nil
}

func (s *Storage) Exists(ctx context.Context, key string) bool {
	existsContext, cancel := context.WithCancel(ctx)
	defer cancel()

	var value []byte
	if err := s.db.QueryRowContext(existsContext, getKey, key).Scan(&value); err != nil {
		return false
	}
	return true
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	deleteContext, cancel := context.WithCancel(ctx)
	defer cancel()
	if key == "" {
		return errors.New("key cannot be empty")
	}
	stmt, err := s.db.Prepare(deleteKey)
	if err != nil {
		return err
	}

	result, err := stmt.ExecContext(deleteContext, key)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fs.ErrNotExist
	}
	return nil
}

func (s *Storage) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	listContext, cancel := context.WithCancel(ctx)
	defer cancel()

	if recursive {
		return nil, errors.New("recursive not supported")
	}

	rows, err := s.db.QueryContext(listContext, listKey, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keyList := make([]string, 0)
	var key string
	if rows.Next() {
		err = rows.Scan(&key)
		if err != nil {
			return keyList, err
		}
		keyList = append(keyList, key)
		for rows.Next() {
			err = rows.Scan(&key)
			if err != nil {
				return keyList, err
			}
			keyList = append(keyList, key)
		}
	} else {
		return keyList, fs.ErrNotExist
	}

	if err = rows.Err(); err != nil {
		return keyList, err
	}
	return keyList, nil
}

func (s *Storage) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
	statContext, cancel := context.WithCancel(ctx)
	defer cancel()

	var modifiedString string
	var size int64
	if err := s.db.QueryRowContext(statContext, statKey, key).Scan(&modifiedString, &size); err != nil {
		if err == sql.ErrNoRows {
			return certmagic.KeyInfo{}, fs.ErrNotExist
		}
		return certmagic.KeyInfo{}, err
	}
	modified, err := time.Parse(time.Layout, modifiedString)
	if err != nil {
		return certmagic.KeyInfo{}, err
	}
	keyInfo := certmagic.KeyInfo{
		Key:      key,
		Modified: modified,
		Size:     size,
	}
	return keyInfo, nil
}

// Lock unimplemented TODO
func (s *Storage) Lock(ctx context.Context, name string) error {
	return errors.New("unimplemented")
}

// UnLock unimplemented TODO
func (s *Storage) Unlock(ctx context.Context, name string) error {
	return errors.New("unimplemented")
}

func (s *Storage) SetLockTimeOut(timeout time.Duration) {
	s.lockExpirationTimeOut = timeout
}

func (s *Storage) CertMagicStorage() (certmagic.Storage, error) {
	return s, nil
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

const defaultLockTimeOut = 500 * time.Millisecond

// interface guards
var (
	_ certmagic.Storage      = (*Storage)(nil)
	_ caddy.StorageConverter = (*Storage)(nil)
)
