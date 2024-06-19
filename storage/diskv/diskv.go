// Package diskv implements a NanoDEP storage backend using the diskv key-value store.
package diskv

import (
	"path/filepath"

	"github.com/micromdm/nanodep/storage/kv"

	"github.com/micromdm/nanolib/storage/kv/kvdiskv"
	"github.com/peterbourgon/diskv/v3"
)

// Diskv is a storage backend that uses diskv.
type Diskv struct {
	*kv.KV
}

func New(path string) *Diskv {
	return &Diskv{KV: kv.New(
		kvdiskv.New(diskv.New(diskv.Options{
			BasePath:     filepath.Join(path, "dep_names"),
			Transform:    kvdiskv.FlatTransform,
			CacheSizeMax: 1024 * 1024,
		})),
	)}
}
