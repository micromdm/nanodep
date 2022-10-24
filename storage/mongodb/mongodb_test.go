package mongodb

import (
	"context"
	"os"
	"testing"

	"github.com/micromdm/nanodep/storage"
	"github.com/micromdm/nanodep/storage/storagetest"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoDBStorage(t *testing.T) {

	testDSN := os.Getenv("NANODEP_MONGODB_STORAGE_TEST")
	if testDSN == "" {
		t.Skip("NANODEP_MONGODB_STORAGE_TEST not set")
	}
	initTestDB(t)

	storagetest.Run(t, func(t *testing.T) storage.AllStorage {
		var err error
		dsn := "mongodb://root:root@127.0.0.1:27017"
		storage := &MongoDBStorage{}
		mongoOpts := options.Client().ApplyURI(dsn)

		storage.MongoClient, err = mongo.NewClient(mongoOpts)
		if err != nil {
			t.Fatal(err)
		}
		s, err := New(dsn)
		if err != nil {
			t.Fatal(err)
		}
		return s
	})
}

// initTestDB clears any existing data from the database.
func initTestDB(t *testing.T) error {
	var err error
	ctx := context.TODO()
	dsn := "mongodb://root:root@127.0.0.1:27017"
	storage := &MongoDBStorage{}
	mongoOpts := options.Client().ApplyURI(dsn)

	storage.MongoClient, err = mongo.NewClient(mongoOpts)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.MongoClient.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	storage.TokenPKICollection = storage.MongoClient.Database(databaseName).Collection(tokenPKIStoreName)
	storage.ProfileCollection = storage.MongoClient.Database(databaseName).Collection(profileStoreName)
	storage.CursorCollection = storage.MongoClient.Database(databaseName).Collection(cursorStoreName)
	storage.ConfigCollection = storage.MongoClient.Database(databaseName).Collection(configStoreName)
	storage.AuthTokenCollection = storage.MongoClient.Database(databaseName).Collection(authTokenStoreName)

	storage.AuthTokenCollection.Drop(ctx)
	storage.ConfigCollection.Drop(ctx)
	storage.CursorCollection.Drop(ctx)
	storage.ProfileCollection.Drop(ctx)
	storage.TokenPKICollection.Drop(ctx)

	return nil
}
