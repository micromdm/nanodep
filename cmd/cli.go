package cmd

import (
	"fmt"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/http/api"
	"github.com/micromdm/nanodep/storage/file"
	"github.com/micromdm/nanodep/sync"
)

// AllStorage represents all possible required storage used by NanoDEP.
type AllStorage interface {
	client.AuthTokensRetriever
	client.ConfigRetriever
	sync.AssignerProfileRetriever
	sync.CursorStorage
	api.AuthTokensStorer
	api.ConfigStorer
	api.TokenPKIStorer
	api.TokenPKIRetriever
	api.AssignerProfileStorer
}

// ParseStorage parses storage and dsn to determine which and return a storage backend.
func ParseStorage(storage, dsn string) (AllStorage, error) {
	if storage == "" {
		storage = "file"
	}
	var store AllStorage
	var err error
	switch storage {
	case "file":
		if dsn == "" {
			dsn = "db"
		}
		store, err = file.New(dsn)
	default:
		return nil, fmt.Errorf("unknown storage: %q", storage)
	}
	return store, err
}
