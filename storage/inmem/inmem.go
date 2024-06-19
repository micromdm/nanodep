// Package inmem implements an in-memory NanoDEP storage backend.
package inmem

import (
	"github.com/micromdm/nanodep/storage/kv"

	"github.com/micromdm/nanolib/storage/kv/kvmap"
)

// InMem is an in-memory storage backend.
type InMem struct {
	*kv.KV
}

// New creates a new storage backend.
func New() *InMem {
	return &InMem{KV: kv.New(kvmap.New())}
}
