// Package kv implements a NanoDEP storage backend using a key-value store.
package kv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/storage"

	"github.com/micromdm/nanolib/storage/kv"
)

const (
	keySfxConsumerKey       = ".consumer_key"
	keySfxConsumerSecret    = ".consumer_secret"
	keySfxAccessToken       = ".access_token"
	keySfxAccessSecret      = ".access_secret"
	keySfxAccessTokenExpiry = ".access_token_expiry"

	keySfxConfig = ".config"

	keySfxCursor = ".cursor"

	keySfxAssignerProfile        = ".assigner_profile"
	keySfxAssignerProfileModTime = ".assigner_profile_mod_time"

	keySfxCert        = ".cert"
	keySfxCertStaging = ".cert_staging"
	keySfxKey         = ".key"
	keySfxKeyStaging  = ".key_staging"
)

type KV struct {
	b kv.Bucket
}

func New(b kv.Bucket) *KV {
	return &KV{b: b}
}

// StoreAuthTokens saves the DEP OAuth tokens to disk as JSON for name DEP name.
func (s *KV) StoreAuthTokens(ctx context.Context, name string, tokens *client.OAuth1Tokens) error {
	expiryText, err := tokens.AccessTokenExpiry.MarshalText()
	if err != nil {
		return err
	}
	return kv.SetMap(ctx, s.b, map[string][]byte{
		name + keySfxConsumerKey:       []byte(tokens.ConsumerKey),
		name + keySfxConsumerSecret:    []byte(tokens.ConsumerSecret),
		name + keySfxAccessToken:       []byte(tokens.AccessToken),
		name + keySfxAccessSecret:      []byte(tokens.AccessSecret),
		name + keySfxAccessTokenExpiry: expiryText,
	})
}

// RetrieveAuthTokens reads the JSON DEP OAuth tokens from disk for name DEP name.
func (s *KV) RetrieveAuthTokens(ctx context.Context, name string) (*client.OAuth1Tokens, error) {
	tokenMap, err := kv.GetMap(ctx, s.b, []string{
		name + keySfxConsumerKey,
		name + keySfxConsumerSecret,
		name + keySfxAccessToken,
		name + keySfxAccessSecret,
		name + keySfxAccessTokenExpiry,
	})
	if errors.Is(err, kv.ErrKeyNotFound) {
		return nil, fmt.Errorf("%w: %v", storage.ErrNotFound, err)
	} else if err != nil {
		return nil, err
	}
	tokens := &client.OAuth1Tokens{
		ConsumerKey:    string(tokenMap[name+keySfxConsumerKey]),
		ConsumerSecret: string(tokenMap[name+keySfxConsumerSecret]),
		AccessToken:    string(tokenMap[name+keySfxAccessToken]),
		AccessSecret:   string(tokenMap[name+keySfxAccessSecret]),
	}
	return tokens, tokens.AccessTokenExpiry.UnmarshalText(tokenMap[name+keySfxAccessTokenExpiry])
}

// StoreConfig saves the DEP config to disk as JSON for name DEP name.
func (s *KV) StoreConfig(ctx context.Context, name string, config *client.Config) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return s.b.Set(ctx, name+keySfxConfig, configJSON)
}

// RetrieveConfig reads the JSON DEP config of a DEP name.
//
// Returns (nil, nil) if the DEP name does not exist, or if the config
// for the DEP name does not exist.
func (s *KV) RetrieveConfig(ctx context.Context, name string) (*client.Config, error) {
	configJSON, err := s.b.Get(ctx, name+keySfxConfig)
	if errors.Is(err, kv.ErrKeyNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	config := new(client.Config)
	return config, json.Unmarshal(configJSON, config)
}

// StoreAssignerProfile saves the assigner profile UUID to disk for name DEP name.
func (s *KV) StoreAssignerProfile(ctx context.Context, name string, profileUUID string) error {
	modTimeText, err := time.Now().UTC().MarshalText()
	if err != nil {
		return err
	}
	return kv.SetMap(ctx, s.b, map[string][]byte{
		name + keySfxAssignerProfile:        []byte(profileUUID),
		name + keySfxAssignerProfileModTime: modTimeText,
	})
}

// RetrieveAssignerProfile reads the assigner profile UUID and its configured
// timestamp from disk for name DEP name.
//
// Returns an empty profile if it does not exist.
func (s *KV) RetrieveAssignerProfile(ctx context.Context, name string) (string, time.Time, error) {
	profileMap, err := kv.GetMap(ctx, s.b, []string{
		name + keySfxAssignerProfile,
		name + keySfxAssignerProfileModTime,
	})
	var modTime time.Time
	if errors.Is(err, kv.ErrKeyNotFound) {
		return "", modTime, nil
	}
	return string(profileMap[name+keySfxAssignerProfile]),
		modTime,
		modTime.UnmarshalText(profileMap[name+keySfxAssignerProfileModTime])
}

// RetrieveCursor retrieves the cursor from the key-value store for name DEP name.
// If the DEP name or cursor does not exist an empty cursor and nil error is be returned.
func (s *KV) RetrieveCursor(ctx context.Context, name string) (string, error) {
	cursor, err := s.b.Get(ctx, name+keySfxCursor)
	if errors.Is(err, kv.ErrKeyNotFound) {
		return "", nil
	}
	return string(cursor), err
}

// StoreCursor stores the cursor to disk for name DEP name.
func (s *KV) StoreCursor(ctx context.Context, name, cursor string) error {
	return s.b.Set(ctx, name+keySfxCursor, []byte(cursor))
}

// StoreTokenPKI stores the PEM bytes in pemCert and pemKey to disk for name DEP name.
func (s *KV) StoreTokenPKI(ctx context.Context, name string, pemCert []byte, pemKey []byte) error {
	return kv.SetMap(ctx, s.b, map[string][]byte{
		name + keySfxCertStaging: pemCert,
		name + keySfxKeyStaging:  pemKey,
	})
}

// UpstageTokenPKI copies the staging PKI certificate and key to the current PKI certificate and key.
// Warning: this operation is not atomic.
func (s *KV) UpstageTokenPKI(ctx context.Context, name string) error {
	tokenPKIMap, err := kv.GetMap(ctx, s.b, []string{
		name + keySfxCertStaging,
		name + keySfxKeyStaging,
	})
	if err != nil {
		return nil
	}
	return kv.SetMap(ctx, s.b, map[string][]byte{
		name + keySfxCert: tokenPKIMap[name+keySfxCertStaging],
		name + keySfxKey:  tokenPKIMap[name+keySfxKeyStaging],
	})
}

// RetrieveStagingTokenPKI reads and returns the PEM bytes for the staged
// DEP token exchange certificate and private key from disk using name DEP name.
func (s *KV) RetrieveStagingTokenPKI(ctx context.Context, name string) ([]byte, []byte, error) {
	tokenPKIMap, err := kv.GetMap(ctx, s.b, []string{
		name + keySfxCertStaging,
		name + keySfxKeyStaging,
	})
	if errors.Is(err, kv.ErrKeyNotFound) {
		return nil, nil, fmt.Errorf("%w: %v", storage.ErrNotFound, err)
	} else if err != nil {
		return nil, nil, err
	}
	return tokenPKIMap[name+keySfxCertStaging], tokenPKIMap[name+keySfxKeyStaging], nil
}

// RetrieveCurrentTokenPKI reads and returns the PEM bytes for the previously-
// upstaged DEP token exchange certificate and private key from disk using
// name DEP name.
func (s *KV) RetrieveCurrentTokenPKI(ctx context.Context, name string) ([]byte, []byte, error) {
	tokenPKIMap, err := kv.GetMap(ctx, s.b, []string{
		name + keySfxCert,
		name + keySfxKey,
	})
	if errors.Is(err, kv.ErrKeyNotFound) {
		return nil, nil, fmt.Errorf("%w: %v", storage.ErrNotFound, err)
	} else if err != nil {
		return nil, nil, err
	}
	return tokenPKIMap[name+keySfxCert], tokenPKIMap[name+keySfxKey], nil
}
