package mysql

import (
	"context"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/micromdm/nanodep/storage/test"
)

func TestMySQLStorage(t *testing.T) {
	testDSN := os.Getenv("NANODEP_MYSQL_STORAGE_TEST_DSN")
	if testDSN == "" {
		t.Skip("NANODEP_MYSQL_STORAGE_TEST_DSN not set")
	}

	s, err := New(WithDSN(testDSN))
	if err != nil {
		t.Fatal(err)
	}

	test.TestWithStorages(t, context.Background(), s)
}
