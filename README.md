# CertMagic Storage Backend for SQLite3
This is an implementation of the CertMagic Storage backend interface using SQLite3.
## Prequisites
You should have SQLite3 installed.
## Building
`xcaddy build --with github.com/crmejia/certmagic_sqlite3`

## Caddyfile Example
```
{
    storage sqlite "certmagic" {
         
    }
}

:2015 {
	respond "Hello, World"
}
```
