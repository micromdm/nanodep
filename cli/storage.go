// Package cli contains shared command-line helpers and utilities.
package cli

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/micromdm/nanodep/storage"
	"github.com/micromdm/nanodep/storage/diskv"
	"github.com/micromdm/nanodep/storage/file"
	"github.com/micromdm/nanodep/storage/inmem"
	"github.com/micromdm/nanodep/storage/mysql"
)

// Storage parses a storage name and dsn to determine which and return a storage backend.
func Storage(storageName, dsn string) (storage.AllStorage, error) {
	var store storage.AllStorage
	var err error
	switch storageName {
	case "file":
		if dsn == "" {
			dsn = "db"
		}
		store, err = file.New(dsn)
	case "diskv":
		if dsn == "" {
			dsn = "diskv"
		}
		store = diskv.New(dsn)
	case "inmem":
		store = inmem.New()
	case "mysql":
		store, err = mysql.New(mysql.WithDSN(dsn))
	default:
		return nil, fmt.Errorf("unknown storage: %q", storageName)
	}
	return store, err
}
