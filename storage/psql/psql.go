package psql

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/storage"
	"github.com/micromdm/nanodep/storage/psql/sqlc"
)

// PSQL implements storage.AllStorage using PSQL.
type PSQLStorage struct {
	db *sql.DB
	q  *sqlc.Queries
}

type config struct {
	driver string
	dsn    string
	db     *sql.DB
}

// Function callback to configure PSQLStorage
type Option func(*config)

// WithDSN sets the data source name
func WithDSN(dsn string) Option {
	return func(c *config) {
		c.dsn = dsn
	}
}

// WithDriver sets the driver
func WithDriver(driver string) Option {
	return func(c *config) {
		c.driver = driver
	}
}

// WithDB sets the db
func WithDB(db *sql.DB) Option {
	return func(c *config) {
		c.db = db

	}
}

// Create a new PSQLStorage instance
func New(opts ...Option) (*PSQLStorage, error) {
	cfg := &config{driver: "postgres"}
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
	return &PSQLStorage{db: cfg.db, q: sqlc.New(cfg.db)}, nil

}

const timestampFormat = "2006-01-02T15:04:05Z"

// RetrieveAuthTokens reads the DEP OAuth tokens for name (DEP name).
func (s *PSQLStorage) RetrieveAuthTokens(ctx context.Context, name string) (*client.OAuth1Tokens, error) {
	tokenRow, err := s.q.GetAuthTokens(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%v: %w", err, storage.ErrNotFound)
		}
		return nil, err
	}
	if !tokenRow.ConsumerKey.Valid { // all auth token fields are set together
		return nil, fmt.Errorf("consumer key not valid: %w", storage.ErrNotFound)
	}
	fmt.Println(tokenRow.AccessTokenExpiry.String)
	accessTokenExpiryTime, err := time.Parse(timestampFormat, tokenRow.AccessTokenExpiry.String)
	if err != nil {
		return nil, err
	}
	return &client.OAuth1Tokens{
		ConsumerKey:       tokenRow.ConsumerKey.String,
		ConsumerSecret:    tokenRow.ConsumerSecret.String,
		AccessToken:       tokenRow.AccessToken.String,
		AccessSecret:      tokenRow.AccessSecret.String,
		AccessTokenExpiry: accessTokenExpiryTime,
	}, nil
}

// StoreAuthTokens saves the DEP OAuth tokens for the DEP name.
func (s *PSQLStorage) StoreAuthTokens(ctx context.Context, name string, tokens *client.OAuth1Tokens) error {
	return s.q.StoreAuthTokens(ctx, sqlc.StoreAuthTokensParams{
		Name:              name,
		ConsumerKey:       sql.NullString{String: tokens.ConsumerKey, Valid: true},
		ConsumerSecret:    sql.NullString{String: tokens.ConsumerSecret, Valid: true},
		AccessToken:       sql.NullString{String: tokens.AccessToken, Valid: true},
		AccessSecret:      sql.NullString{String: tokens.AccessSecret, Valid: true},
		AccessTokenExpiry: sql.NullString{String: tokens.AccessTokenExpiry.Format(timestampFormat), Valid: true},
	})
}

// RetrieveConfig reads the JSON DEP config of a DEP name.
//
// Returns (nil, nil) if the DEP name does not exist, or if the config
// for the DEP name does not exist.
func (s *PSQLStorage) RetrieveConfig(ctx context.Context, name string) (*client.Config, error) {
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

// StoreConfig saves the DEP config for name (DEP name).
func (s *PSQLStorage) StoreConfig(ctx context.Context, name string, config *client.Config) error {
	return s.q.StoreConfig(ctx, sqlc.StoreConfigParams{
		Name:          name,
		ConfigBaseUrl: sql.NullString{String: config.BaseURL, Valid: true},
	})
}

// RetrieveAssignerProfile reads the assigner profile UUID and its timestamp for name (DEP name).
//
// Returns an empty profile UUID if it does not exist.
func (s *PSQLStorage) RetrieveAssignerProfile(ctx context.Context, name string) (profileUUID string, modTime time.Time, err error) {
	assignerRow, err := s.q.GetAssignerProfile(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// an 'empty' profile UUID is valid, return nil error
			return "", time.Time{}, nil
		}
		return "", time.Time{}, err
	}
	if assignerRow.AssignerProfileUuid.Valid {
		profileUUID = assignerRow.AssignerProfileUuid.String
	}
	if assignerRow.AssignerProfileUuidAt.Valid {
		modTime, err = time.Parse(timestampFormat, assignerRow.AssignerProfileUuidAt.String)
	}
	return
}

// StoreAssignerProfile saves the assigner profile UUID for name (DEP name).
func (s *PSQLStorage) StoreAssignerProfile(ctx context.Context, name string, profileUUID string) error {
	return s.q.StoreAssignerProfile(ctx, sqlc.StoreAssignerProfileParams{
		Name:                name,
		AssignerProfileUuid: sql.NullString{String: profileUUID, Valid: true},
	})
}

// RetrieveCursor reads the reads the DEP fetch and sync cursor for name (DEP name).
//
// Returns an empty cursor if the cursor does not exist.
func (s *PSQLStorage) RetrieveCursor(ctx context.Context, name string) (string, error) {
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

// StoreCursor saves the DEP fetch and sync cursor for name (DEP name).
func (s *PSQLStorage) StoreCursor(ctx context.Context, name, cursor string) error {
	return s.q.StoreCursor(ctx, sqlc.StoreCursorParams{
		Name:         name,
		SyncerCursor: sql.NullString{String: cursor, Valid: true},
	})

}

// StoreTokenPKI stores the staging PEM bytes in pemCert and pemKey for name (DEP name).
func (s *PSQLStorage) StoreTokenPKI(ctx context.Context, name string, pemCert []byte, pemKey []byte) error {
	return s.q.StoreTokenPKI(ctx, sqlc.StoreTokenPKIParams{
		Name:                   name,
		TokenpkiStagingCertPem: pemCert,
		TokenpkiStagingKeyPem:  pemKey,
	})
}

// UpstageTokenPKI copies the staging PKI certificate and private key to the
// current PKI certificate and private key.
func (s *PSQLStorage) UpstageTokenPKI(ctx context.Context, name string) error {
	err := s.q.UpstageKeypair(ctx, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%v: %w", err, storage.ErrNotFound)
	}
	return err
}

// RetrieveStagingTokenPKI returns the PEM bytes for the staged DEP
// token exchange certificate and private key using name (DEP name).
func (s *PSQLStorage) RetrieveStagingTokenPKI(ctx context.Context, name string) ([]byte, []byte, error) {
	keypair, err := s.q.GetStagingKeypair(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("%v: %w", err, storage.ErrNotFound)
		}
		return nil, nil, err
	}
	if keypair.TokenpkiStagingCertPem == nil { // tokenpki_staging_cert_pem and tokenpki_staging_key_pem are set together
		return nil, nil, fmt.Errorf("empty certificate: %w", storage.ErrNotFound)
	}
	return keypair.TokenpkiStagingCertPem, keypair.TokenpkiStagingKeyPem, nil
}

// RetrieveCurrentTokenPKI returns the PEM bytes for the previously-upstaged DEP
// token exchange certificate and private key using name (DEP name).
func (s *PSQLStorage) RetrieveCurrentTokenPKI(ctx context.Context, name string) (pemCert []byte, pemKey []byte, err error) {
	keypair, err := s.q.GetCurrentKeypair(ctx, name)
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
