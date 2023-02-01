package certmagicsqlite3

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Storage{})
	httpcaddyfile.RegisterHandlerDirective("certmagic_sqlite", parseCaddyFile)
}

func (Storage) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.certmagic_sqlite",
		New: func() caddy.Module {
			return new(Storage)
		},
	}
}

func (m *Storage) Provision(ctx caddy.Context) error {
	//	switch m.Output{
	//
	// case "stdout":
	//
	//	m.w = os.Stdout
	//
	// case "stderr":
	//
	//	m.w = os.Stderr
	//
	// default:
	// return fmt.Errorf("an output stream is required, got %s", m.Output)
	//
	//	}
	return nil
}

func (m *Storage) Validate() error {
	// if m.w == nil {
	// 	// return fmt.Errorf("no writer")
	// }
	return nil
}

func (m Storage) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// m.w.Write([]byte(r.RemoteAddr))
	return next.ServeHTTP(w, r)
}

func (m *Storage) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.Args(&m.DataSourceName) {
			return d.ArgErr()
		}
	}
	return nil
}

func parseCaddyFile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Storage
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// interface guards
var (
	_ caddy.Provisioner = (*Storage)(nil)
	_ caddy.Validator   = (*Storage)(nil)
	_ caddyhttp.MiddlewareHandler
	_ caddyfile.Unmarshaler = (*Storage)(nil)
)
