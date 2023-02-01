package certmagicsqlite3

import (
	"database/sql"
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func init() {
	caddy.RegisterModule(Storage{})
}

func (Storage) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "caddy.storage.sqlite",
		New: func() caddy.Module {
			return new(Storage)
		},
	}
}

func (s *Storage) Provision(ctx caddy.Context) error {
	if s.DataSourceName == "" {
		return errors.New("set data source name")
	}

	db, err := sql.Open("sqlite", s.DataSourceName)
	if err != nil {
		return err
	}

	for _, stmt := range []string{pragmaWALEnabled, pragma500BusyTimeout, pragmaCaseSenstive} {
		_, err = db.Exec(stmt, nil)
		if err != nil {
			return err
		}
	}

	_, err = db.Exec(createTable)
	if err != nil {
		return err
	}

	s.db = db
	s.lockExpirationTimeOut = defaultLockTimeOut
	return nil
}

func (s *Storage) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.Args(&s.DataSourceName) {
			return d.ArgErr()
		}
	}
	return nil
}

// interface guards
var (
	_ caddy.Provisioner     = (*Storage)(nil)
	_ caddyfile.Unmarshaler = (*Storage)(nil)
)
