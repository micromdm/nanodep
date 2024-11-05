package inmem

import (
	"context"
	"testing"

	"github.com/micromdm/nanodep/storage/test"
)

func TestInMem(t *testing.T) {
	test.TestWithStorages(t, context.Background(), New())
}
