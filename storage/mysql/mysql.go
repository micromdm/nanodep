package mysql

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/storage"
	"github.com/micromdm/nanodep/storage/mysql/sqlc"
)

// Schema contains the MySQL schema for the DEP storage.
//
//go:embed schema.sql
var Schema string

// MySQLStorage implements a storage.AllStorage using MySQL.
type MySQLStorage struct {
	db *sql.DB
	q  *sqlc.Queries
}

type config struct {
	driver string
	dsn    string
	db     *sql.DB
}

// Option allows configuring a MySQLStorage.
type Option func(*config)

// WithDSN sets the storage MySQL data source name.
func WithDSN(dsn string) Option {
	return func(c *config) {
		c.dsn = dsn
	}
}

// WithDriver sets a custom MySQL driver for the storage.
//
// Default driver is "mysql".
// Value is ignored if WithDB is used.
func WithDriver(driver string) Option {
	return func(c *config) {
		c.driver = driver
	}
}

// WithDB sets a custom MySQL *sql.DB to the storage.
//
// If set, driver passed via WithDriver is ignored.
func WithDB(db *sql.DB) Option {
	return func(c *config) {
		c.db = db
	}
}

// New creates and returns a new MySQLStorage.
func New(opts ...Option) (*MySQLStorage, error) {
	cfg := &config{driver: "mysql"}
	for _, opt := range opts {
		opt(cfg)
	}
	var err error
	if cfg.db == nil {
		cfg.db, err = sql.Open(cfg.driver, cfg.dsn)
		if err != nil {
			return nil, err
		}
	}
	if err = cfg.db.Ping(); err != nil {
		return nil, err
	}
	return &MySQLStorage{db: cfg.db, q: sqlc.New(cfg.db)}, nil
}

const timestampFormat = "2006-01-02 15:04:05"

// RetrieveAuthTokens reads the DEP OAuth tokens for name DEP name.
func (s *MySQLStorage) RetrieveAuthTokens(ctx context.Context, name string) (*client.OAuth1Tokens, error) {
	var (
		consumerKey       sql.NullString
		consumerSecret    sql.NullString
		accessToken       sql.NullString
		accessSecret      sql.NullString
		accessTokenExpiry sql.NullString
	)
	err := s.db.QueryRowContext(
		ctx, `
SELECT
	consumer_key,
	consumer_secret,
	access_token,
	access_secret,
	access_token_expiry
FROM
    dep_names
WHERE
    name = ?;`,
		name,
	).Scan(
		&consumerKey,
		&consumerSecret,
		&accessToken,
		&accessSecret,
		&accessTokenExpiry,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%v: %w", err, storage.ErrNotFound)
		}
		return nil, err
	}
	if !consumerKey.Valid { // all auth token fields are set together
		return nil, fmt.Errorf("consumer key not valid: %w", storage.ErrNotFound)
	}
	accessTokenExpiryTime, err := time.Parse(timestampFormat, accessTokenExpiry.String)
	if err != nil {
		return nil, err
	}
	return &client.OAuth1Tokens{
		ConsumerKey:       consumerKey.String,
		ConsumerSecret:    consumerSecret.String,
		AccessToken:       accessToken.String,
		AccessSecret:      accessSecret.String,
		AccessTokenExpiry: accessTokenExpiryTime,
	}, nil
}

// StoreAuthTokens saves the DEP OAuth tokens for the DEP name.
func (s *MySQLStorage) StoreAuthTokens(ctx context.Context, name string, tokens *client.OAuth1Tokens) error {
	_, err := s.db.ExecContext(
		ctx, `
INSERT INTO dep_names 
	(name, consumer_key, consumer_secret, access_token, access_secret, access_token_expiry)
VALUES 
	(?, ?, ?, ?, ?, ?) as new
ON DUPLICATE KEY UPDATE 
	consumer_key = new.consumer_key,
	consumer_secret = new.consumer_secret,
	access_token = new.access_token,
	access_secret = new.access_secret,
	access_token_expiry = new.access_token_expiry;`,
		name,
		tokens.ConsumerKey,
		tokens.ConsumerSecret,
		tokens.AccessToken,
		tokens.AccessSecret,
		tokens.AccessTokenExpiry.Format(timestampFormat),
	)
	return err
}

// RetrieveConfig reads the JSON DEP config of a DEP name.
//
// Returns (nil, nil) if the DEP name does not exist, or if the config
// for the DEP name does not exist.
func (s *MySQLStorage) RetrieveConfig(ctx context.Context, name string) (*client.Config, error) {
	baseURL, err := s.q.GetConfigBaseURL(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If the DEP name does not exist, then the config does not exist.
			return nil, nil
		}
		return nil, err
	}
	if !baseURL.Valid {
		// If the config_base_url is NULL, then config does not exist.
		return nil, nil
	}
	return &client.Config{
		BaseURL: baseURL.String,
	}, nil
}

// StoreConfig saves the DEP config for name DEP name.
func (s *MySQLStorage) StoreConfig(ctx context.Context, name string, config *client.Config) error {
	_, err := s.db.ExecContext(
		ctx, `
INSERT INTO dep_names
	(name, config_base_url)
VALUES 
	(?, ?) as new
ON DUPLICATE KEY UPDATE
	config_base_url = new.config_base_url;`,
		name,
		config.BaseURL,
	)
	return err
}

// RetrieveAssignerProfile reads the assigner profile UUID and its timestamp for name DEP name.
//
// Returns an empty profile if it does not exist.
func (s *MySQLStorage) RetrieveAssignerProfile(ctx context.Context, name string) (profileUUID string, modTime time.Time, err error) {
	var (
		profileUUID_ sql.NullString
		modTime_     sql.NullString
	)
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT assigner_profile_uuid, assigner_profile_uuid_at FROM dep_names WHERE name = ?;`,
		name,
	).Scan(
		&profileUUID_, &modTime_,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// an 'empty' profile is valid
			return "", time.Time{}, nil
		}
		return "", time.Time{}, err
	}
	if profileUUID_.Valid {
		profileUUID = profileUUID_.String
	}
	if modTime_.Valid {
		modTime, err = time.Parse(timestampFormat, modTime_.String)
		if err != nil {
			return "", time.Time{}, err
		}
	}
	return profileUUID, modTime, nil
}

// StoreAssignerProfile saves the assigner profile UUID for name DEP name.
func (s *MySQLStorage) StoreAssignerProfile(ctx context.Context, name string, profileUUID string) error {
	_, err := s.db.ExecContext(
		ctx, `
INSERT INTO dep_names
	(name, assigner_profile_uuid, assigner_profile_uuid_at)
VALUES
	(?, ?, CURRENT_TIMESTAMP) as new
ON DUPLICATE KEY UPDATE
	assigner_profile_uuid = new.assigner_profile_uuid,
	assigner_profile_uuid_at = new.assigner_profile_uuid_at;`,
		name,
		profileUUID,
	)
	return err
}

// RetrieveCursor reads the reads the DEP fetch and sync cursor for name DEP name.
//
// Returns an empty cursor if the cursor does not exist.
func (s *MySQLStorage) RetrieveCursor(ctx context.Context, name string) (string, error) {
	cursor, err := s.q.GetSyncerCursor(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if !cursor.Valid {
		return "", nil
	}
	return cursor.String, nil
}

// StoreCursor saves the DEP fetch and sync cursor for name DEP name.
func (s *MySQLStorage) StoreCursor(ctx context.Context, name, cursor string) error {
	_, err := s.db.ExecContext(
		ctx, `
INSERT INTO dep_names
	(name, syncer_cursor)
VALUES
	(?, ?) as new
ON DUPLICATE KEY UPDATE
	syncer_cursor = new.syncer_cursor`,
		name,
		cursor,
	)
	return err
}

// StoreTokenPKI stores the PEM bytes in pemCert and pemKey for name DEP name.
func (s *MySQLStorage) StoreTokenPKI(ctx context.Context, name string, pemCert []byte, pemKey []byte) error {
	_, err := s.db.ExecContext(
		ctx, `
INSERT INTO dep_names
	(name, tokenpki_cert_pem, tokenpki_key_pem)
VALUES
	(?, ?, ?) as new
ON DUPLICATE KEY UPDATE
	tokenpki_cert_pem = new.tokenpki_cert_pem,
	tokenpki_key_pem = new.tokenpki_key_pem;`,
		name,
		pemCert,
		pemKey,
	)
	return err
}

// RetrieveTokenPKI reads the PEM bytes for the DEP token exchange certificate
// and private key using name DEP name.
func (s *MySQLStorage) RetrieveTokenPKI(ctx context.Context, name string) (pemCert []byte, pemKey []byte, err error) {
	keypair, err := s.q.GetKeypair(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("%v: %w", err, storage.ErrNotFound)
		}
		return nil, nil, err
	}
	if keypair.TokenpkiCertPem == nil { // tokenpki_cert_pem and tokenpki_key_pem are set together
		return nil, nil, fmt.Errorf("empty certificate: %w", storage.ErrNotFound)
	}
	return keypair.TokenpkiCertPem, keypair.TokenpkiKeyPem, nil
}
