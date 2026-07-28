// Package cli contains shared command-line helpers and utilities.
package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/micromdm/nanodep/storage"
	"github.com/micromdm/nanodep/storage/diskv"
	"github.com/micromdm/nanodep/storage/file"
	"github.com/micromdm/nanodep/storage/inmem"
	"github.com/micromdm/nanodep/storage/mysql"
	"github.com/micromdm/nanodep/storage/pgsql"
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
		store, err = mysqlStorage(dsn, options)
	case "pgsql":
		store, err = pgsql.New(pgsql.WithDSN(dsn))
	default:
		return nil, fmt.Errorf("unknown storage: %q", storageName)
	}
	return store, err
}

// mysqlStorage builds a MySQL storage backend, parsing any storage options.
func mysqlStorage(dsn, options string) (*mysql.MySQLStorage, error) {
	opts := []mysql.Option{mysql.WithDSN(dsn)}
	if options != "" {
		for k, v := range splitOptions(options) {
			switch k {
			case "conn_max_lifetime":
				d, err := time.ParseDuration(v)
				if err != nil {
					return nil, fmt.Errorf("invalid value for conn_max_lifetime option: %w", err)
				}
				opts = append(opts, mysql.WithConnMaxLifetime(d))
			case "conn_max_idle_time":
				d, err := time.ParseDuration(v)
				if err != nil {
					return nil, fmt.Errorf("invalid value for conn_max_idle_time option: %w", err)
				}
				opts = append(opts, mysql.WithConnMaxIdleTime(d))
			default:
				return nil, fmt.Errorf("invalid option: %q", k)
			}
		}
	}
	return mysql.New(opts...)
}

// splitOptions splits a comma-separated list of key=value storage options.
func splitOptions(s string) map[string]string {
	out := make(map[string]string)
	for _, opt := range strings.Split(s, ",") {
		if opt == "" {
			continue
		}
		k, v, _ := strings.Cut(opt, "=")
		out[k] = v
	}
	return out
}
