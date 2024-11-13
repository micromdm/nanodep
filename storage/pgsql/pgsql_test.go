package pgsql

import (
	"context"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/micromdm/nanodep/storage/test"
)

func TestPSQLStorage(t *testing.T) {
	testDSN := os.Getenv("NANODEP_PSQL_STORAGE_TEST_DSN")
	if testDSN == "" {
		t.Skip("NANODEP_PSQL_STORAGE_TEST_DSN not set")
	}

	s, err := New(WithDSN(testDSN))
	if err != nil {
		t.Fatal(err)
	}

	test.TestWithStorages(t, context.Background(), s)
}
