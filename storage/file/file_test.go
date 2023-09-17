package file

import (
	"context"
	"testing"

	"github.com/micromdm/nanodep/storage/test"
)

func TestFileStorage(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	test.TestWithStorages(t, context.Background(), s)
}
