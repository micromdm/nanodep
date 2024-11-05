// Package cli contains shared command-line helpers and utilities.
package cli

import (
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/micromdm/nanodep/storage"
	"github.com/micromdm/nanodep/storage/diskv"
	"github.com/micromdm/nanodep/storage/file"
	"github.com/micromdm/nanodep/storage/inmem"
	"github.com/micromdm/nanodep/storage/mysql"
	"github.com/micromdm/nanodep/storage/psql"
)

// Storage parses a storage name and dsn to determine which and return a storage backend.
func Storage(storageName, dsn, options string) (storage.AllStorage, error) {
	var store storage.AllStorage
	var err error
	switch storageName {
	case "filekv":
		if dsn == "" {
			dsn = "dbkv"
		}
		store = diskv.New(dsn)
	case "file":
		if options != "enable_deprecated=1" {
			return nil, errors.New("file backend is deprecated; specify storage options to force enable")
		}
		if dsn == "" {
			dsn = "db"
		}
		store, err = file.New(dsn)
	case "inmem":
		store = inmem.New()
	case "mysql":
		store, err = mysql.New(mysql.WithDSN(dsn))
	case "psql":
		store, err = psql.New(psql.WithDSN(dsn))
	default:
		return nil, fmt.Errorf("unknown storage: %q", storageName)
	}
	return store, err
}
