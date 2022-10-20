package mongodb

import (
	"context"
	"time"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

const (
	databaseName = "nanodep"

	tokenPKIStoreName  = "token_pki_store"
	authTokenStoreName = "auth_token_store"
	configStoreName    = "config_store"
	cursorStoreName    = "cursor_store"
	profileStoreName   = "profile_store"
)

var latestSort = bson.M{
	"$natural": -1,
}

type MongoDBStorage struct {
	MongoClient         *mongo.Client
	TokenPKICollection  *mongo.Collection
	AuthTokenCollection *mongo.Collection
	ConfigCollection    *mongo.Collection
	CursorCollection    *mongo.Collection
	ProfileCollection   *mongo.Collection
}

type AuthTokenRecord struct {
	Name              string    `bson:"name"`
	ConsumerKey       string    `bson:"consumer_key,omitempty"`
	ConsumerSecret    string    `bson:"consumer_secret,omitempty"`
	AccessToken       string    `bson:"access_token,omitempty"`
	AccessSecret      string    `bson:"access_secret,omitempty"`
	AccessTokenExpiry time.Time `bson:"access_token_expiry,omitempty"`
}

type TokenPKIRecord struct {
	Name        string `bson:"name"`
	Certificate string `bson:"certificate,omitempty"`
	PrivateKey  string `bson:"key,omitempty"`
}

type ConfigRecord struct {
	Name    string `bson:"name"`
	BaseURL string `bson:"base_url,omitempty"`
}

type CursorRecord struct {
	Name   string `bson:"name"`
	Cursor string `bson:"cursor,omitempty"`
}

type ProfileRecord struct {
	Name        string    `bson:"name"`
	ProfileUUID string    `bson:"profile_uuid,omitempty"`
	Timestamp   time.Time `bson:"timestamp,omitempty"`
}

func New(uri string) (*MongoDBStorage, error) {
	var err error
	storage := &MongoDBStorage{}

	mongoOpts := options.Client().ApplyURI(uri)

	storage.MongoClient, err = mongo.NewClient(mongoOpts)
	if err != nil {
		return nil, err
	}

	err = storage.MongoClient.Connect(context.TODO())
	if err != nil {
		return nil, err
	}

	storage.TokenPKICollection = storage.MongoClient.Database(databaseName).Collection(tokenPKIStoreName)
	_, err = storage.TokenPKICollection.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{
			Keys: bson.M{
				"tokenpki_cert_pem": 1,
			},
		},
		{
			Keys: bson.M{
				"tokenpki_key_pem": 2,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	storage.AuthTokenCollection = storage.MongoClient.Database(databaseName).Collection(authTokenStoreName)
	_, err = storage.AuthTokenCollection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		},
	},
	)
	if err != nil {
		return nil, err
	}

	storage.ConfigCollection = storage.MongoClient.Database(databaseName).Collection(configStoreName)
	_, err = storage.ConfigCollection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		},
	},
	)
	if err != nil {
		return nil, err
	}

	storage.CursorCollection = storage.MongoClient.Database(databaseName).Collection(cursorStoreName)
	_, err = storage.CursorCollection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		},
	},
	)
	if err != nil {
		return nil, err
	}

	storage.ProfileCollection = storage.MongoClient.Database(databaseName).Collection(profileStoreName)
	_, err = storage.ProfileCollection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
		},
	},
	)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

// RetrieveAuthTokens reads the JSON DEP OAuth tokens from mongodb for name DEP name.
//
// In order to seed the database correctly, and pass the ckcheck (http/api/ckcheck.go) an empty consumer key
// should be returned if the ErrNoDocuments type is returned from the FindOne query. This will be true if the
// database is empty
// https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo#ErrNoDocuments
func (s *MongoDBStorage) RetrieveAuthTokens(_ context.Context, name string) (*client.OAuth1Tokens, error) {
	tokens := new(client.OAuth1Tokens)
	resp := new(AuthTokenRecord)

	filter := bson.M{
		"name": name,
	}

	err := s.AuthTokenCollection.FindOne(context.TODO(), filter).Decode(&resp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &client.OAuth1Tokens{
				ConsumerKey: "",
			}, storage.ErrNotFound
		}
		return nil, err
	}

	tokens.ConsumerKey = resp.ConsumerKey
	tokens.ConsumerSecret = resp.ConsumerSecret
	tokens.AccessToken = resp.AccessToken
	tokens.AccessSecret = resp.AccessSecret
	tokens.AccessTokenExpiry = resp.AccessTokenExpiry

	return tokens, nil
}

// StoreAuthTokens saves the DEP OAuth tokens to mongodb for name DEP name.
func (s *MongoDBStorage) StoreAuthTokens(_ context.Context, name string, tokens *client.OAuth1Tokens) error {
	upsert := true
	filter := bson.M{
		"name": name,
	}
	update := bson.M{
		"$set": &AuthTokenRecord{
			Name:              name,
			ConsumerKey:       tokens.ConsumerKey,
			ConsumerSecret:    tokens.ConsumerSecret,
			AccessToken:       tokens.AccessToken,
			AccessSecret:      tokens.AccessSecret,
			AccessTokenExpiry: tokens.AccessTokenExpiry,
		},
	}

	_, err := s.AuthTokenCollection.UpdateOne(context.TODO(), filter, update, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return err
	}
	return nil
}

// RetrieveConfig reads the JSON DEP config of a DEP name.
//
// Returns (nil, nil) if the DEP name does not exist, or if the config
// for the DEP name does not exist.
func (s *MongoDBStorage) RetrieveConfig(_ context.Context, name string) (*client.Config, error) {
	config := new(client.Config)
	resp := new(ConfigRecord)

	filter := bson.M{
		"name": name,
	}

	err := s.ConfigCollection.FindOne(context.TODO(), filter).Decode(&resp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	config.BaseURL = resp.BaseURL

	return config, nil
}

// StoreConfig saves the DEP config to mongodb for name DEP name.
func (s *MongoDBStorage) StoreConfig(_ context.Context, name string, config *client.Config) error {
	upsert := true
	filter := bson.M{
		"name": name,
	}
	update := bson.M{
		"$set": &ConfigRecord{
			Name:    name,
			BaseURL: config.BaseURL,
		},
	}

	_, err := s.ConfigCollection.UpdateOne(context.TODO(), filter, update, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return err
	}
	return nil
}

// RetrieveAssignerProfile reads the assigner profile UUID and its configured
// timestamp from mongodb for name DEP name.
//
// Returns an empty profile if it does not exist.
func (s *MongoDBStorage) RetrieveAssignerProfile(_ context.Context, name string) (string, time.Time, error) {
	resp := new(ProfileRecord)

	filter := bson.M{
		"name": name,
	}

	err := s.ProfileCollection.FindOne(context.TODO(), filter).Decode(&resp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", time.Time{}, nil
		}
		return "", time.Time{}, err
	}

	return resp.ProfileUUID, resp.Timestamp, nil
}

// StoreAssignerProfile saves the assigner profile UUID to disk for name DEP name.
func (s *MongoDBStorage) StoreAssignerProfile(_ context.Context, name string, profileUUID string) error {
	upsert := true
	filter := bson.M{
		"name": name,
	}
	update := bson.M{
		"$set": &ProfileRecord{
			Name:        name,
			ProfileUUID: profileUUID,
			Timestamp:   time.Now(),
		},
	}

	_, err := s.ProfileCollection.UpdateOne(context.TODO(), filter, update, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return err
	}
	return nil
}

// RetrieveCursor reads the reads the DEP fetch and sync cursor from mongodb
// for name DEP name. We return an empty cursor if the cursor does not exist
// in the database.
func (s *MongoDBStorage) RetrieveCursor(_ context.Context, name string) (string, error) {
	resp := new(CursorRecord)

	filter := bson.M{
		"name": name,
	}

	err := s.CursorCollection.FindOne(context.TODO(), filter).Decode(&resp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}

	return resp.Cursor, nil
}

// StoreCursor saves the DEP fetch and sync cursor to mongodb for name DEP name.
func (s *MongoDBStorage) StoreCursor(_ context.Context, name, cursor string) error {
	upsert := true
	filter := bson.M{
		"name": name,
	}
	update := bson.M{
		"$set": &CursorRecord{
			Name:   name,
			Cursor: cursor,
		},
	}

	_, err := s.CursorCollection.UpdateOne(context.TODO(), filter, update, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return err
	}

	return nil
}

// TokenPKIRecord Store and Retrieve

// StoreTokenPKI stores the PEM bytes in pemCert and pemKey to mongodb for name DEP name.
func (s *MongoDBStorage) StoreTokenPKI(_ context.Context, name string, pemCert []byte, pemKey []byte) error {
	_, err := s.TokenPKICollection.InsertOne(context.TODO(), TokenPKIRecord{
		Certificate: string(pemCert),
		PrivateKey:  string(pemKey),
		Name:        name,
	})
	if err != nil {
		return err
	}
	return nil
}

// RetrieveTokenPKI reads and returns the PEM bytes for the DEP token exchange
// certificate and private key from mongodb using name DEP name.
func (s *MongoDBStorage) RetrieveTokenPKI(_ context.Context, name string) ([]byte, []byte, error) {
	filter := bson.M{
		"name": name,
	}
	res := TokenPKIRecord{}
	err := s.TokenPKICollection.FindOne(context.TODO(), filter, options.FindOne().SetSort(latestSort)).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil, storage.ErrNotFound
		}
		return nil, nil, err
	}

	return []byte(res.Certificate), []byte(res.PrivateKey), nil
}
