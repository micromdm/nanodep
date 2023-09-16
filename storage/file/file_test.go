package file

import (
	"testing"

	"github.com/micromdm/nanodep/storage"
	"github.com/micromdm/nanodep/storage/test"
)

func TestFileStorage(t *testing.T) {
	test.Run(t, func(t *testing.T) storage.AllStorage {
		s, err := New(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		return s
	})
}
