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
	keyPfxConsumerKey       = "consumer_key."
	keyPfxConsumerSecret    = "consumer_secret."
	keyPfxAccessToken       = "access_token."
	keyPfxAccessSecret      = "access_secret."
	keyPfxAccessTokenExpiry = "access_token_expiry."

	keyPfxConfig = "config."

	keyPfxCursor = "cursor."

	keyPfxAssignerProfile        = "assigner_profile."
	keyPfxAssignerProfileModTime = "assigner_profile_mod_time."

	keyPfxCert        = "cert."
	keyPfxCertStaging = "cert_staging."
	keyPfxKey         = "key."
	keyPfxKeyStaging  = "key_staging."
)

type KV struct {
	b kv.TxnBucketWithCRUD
}

func New(b kv.TxnBucketWithCRUD) *KV {
	return &KV{b: b}
}

// StoreAuthTokens stores the DEP OAuth tokens for name (DEP name).
func (s *KV) StoreAuthTokens(ctx context.Context, name string, tokens *client.OAuth1Tokens) error {
	expiryText, err := tokens.AccessTokenExpiry.MarshalText()
	if err != nil {
		return err
	}
	err = kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		return kv.SetMap(ctx, txn, map[string][]byte{
			keyPfxConsumerKey + name:       []byte(tokens.ConsumerKey),
			keyPfxConsumerSecret + name:    []byte(tokens.ConsumerSecret),
			keyPfxAccessToken + name:       []byte(tokens.AccessToken),
			keyPfxAccessSecret + name:      []byte(tokens.AccessSecret),
			keyPfxAccessTokenExpiry + name: expiryText,
		})
	})
	return err
}

// RetrieveAuthTokens retrieves the OAuth tokens for name (DEP name).
func (s *KV) RetrieveAuthTokens(ctx context.Context, name string) (*client.OAuth1Tokens, error) {
	var tokenMap map[string][]byte
	err := kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		var err error
		tokenMap, err = kv.GetMap(ctx, s.b, []string{
			keyPfxConsumerKey + name,
			keyPfxConsumerSecret + name,
			keyPfxAccessToken + name,
			keyPfxAccessSecret + name,
			keyPfxAccessTokenExpiry + name,
		})
		return err
	})
	if errors.Is(err, kv.ErrKeyNotFound) {
		return nil, fmt.Errorf("%w: %v", storage.ErrNotFound, err)
	} else if err != nil {
		return nil, err
	}
	tokens := &client.OAuth1Tokens{
		ConsumerKey:    string(tokenMap[keyPfxConsumerKey+name]),
		ConsumerSecret: string(tokenMap[keyPfxConsumerSecret+name]),
		AccessToken:    string(tokenMap[keyPfxAccessToken+name]),
		AccessSecret:   string(tokenMap[keyPfxAccessSecret+name]),
	}
	return tokens, tokens.AccessTokenExpiry.UnmarshalText(tokenMap[keyPfxAccessTokenExpiry+name])
}

// StoreConfig stores the config for name (DEP name), overwriting it.
func (s *KV) StoreConfig(ctx context.Context, name string, config *client.Config) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}
	// auto-commit of storage obviates need for txn for single key
	return s.b.Set(ctx, keyPfxConfig+name, configJSON)
}

// RetrieveConfig retrieves config of name (DEP name).
// If the DEP name or config does not exist then a nil config and
// nil error will be returned.
func (s *KV) RetrieveConfig(ctx context.Context, name string) (*client.Config, error) {
	configJSON, err := s.b.Get(ctx, keyPfxConfig+name)
	if errors.Is(err, kv.ErrKeyNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	config := new(client.Config)
	return config, json.Unmarshal(configJSON, config)
}

// StoreAssignerProfile stores the assigner profile UUID for name (DEP name).
func (s *KV) StoreAssignerProfile(ctx context.Context, name string, profileUUID string) error {
	modTimeText, err := time.Now().UTC().MarshalText()
	if err != nil {
		return err
	}
	err = kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		return kv.SetMap(ctx, txn, map[string][]byte{
			keyPfxAssignerProfile + name:        []byte(profileUUID),
			keyPfxAssignerProfileModTime + name: modTimeText,
		})
	})
	return err
}

// RetrieveAssignerProfile retrieves the assigner profile UUID and its
// configured timestamp name (DEP name).
// Returns an empty profile if it does not exist.
func (s *KV) RetrieveAssignerProfile(ctx context.Context, name string) (string, time.Time, error) {
	var profileMap map[string][]byte
	err := kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		var err error
		profileMap, err = kv.GetMap(ctx, s.b, []string{
			keyPfxAssignerProfile + name,
			keyPfxAssignerProfileModTime + name,
		})
		return err
	})
	var modTime time.Time
	if errors.Is(err, kv.ErrKeyNotFound) {
		return "", modTime, nil
	}
	return string(profileMap[keyPfxAssignerProfile+name]),
		modTime,
		modTime.UnmarshalText(profileMap[keyPfxAssignerProfileModTime+name])
}

// RetrieveCursor retrieves the cursor for name (DEP name).
// If the DEP name or cursor does not exist an empty cursor and nil error will be returned.
func (s *KV) RetrieveCursor(ctx context.Context, name string) (string, error) {
	cursor, err := s.b.Get(ctx, keyPfxCursor+name)
	if errors.Is(err, kv.ErrKeyNotFound) {
		return "", nil
	}
	return string(cursor), err
}

// StoreCursor stores the cursor for name (DEP name).
func (s *KV) StoreCursor(ctx context.Context, name, cursor string) error {
	// auto-commit of storage obviates need for txn for single key
	return s.b.Set(ctx, keyPfxCursor+name, []byte(cursor))
}

// StoreTokenPKI stores the PEM bytes in pemCert and pemKey for name (DEP name).
func (s *KV) StoreTokenPKI(ctx context.Context, name string, pemCert []byte, pemKey []byte) error {
	return kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		return kv.SetMap(ctx, txn, map[string][]byte{
			keyPfxCertStaging + name: pemCert,
			keyPfxKeyStaging + name:  pemKey,
		})
	})
}

// UpstageTokenPKI copies the staging PKI certificate and key to the current PKI certificate and key.
// Warning: this operation is not atomic.
func (s *KV) UpstageTokenPKI(ctx context.Context, name string) error {
	err := kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		tokenPKIMap, err := kv.GetMap(ctx, txn, []string{
			keyPfxCertStaging + name,
			keyPfxKeyStaging + name,
		})
		if err != nil {
			return err
		}
		return kv.SetMap(ctx, txn, map[string][]byte{
			keyPfxCert + name: tokenPKIMap[keyPfxCertStaging+name],
			keyPfxKey + name:  tokenPKIMap[keyPfxKeyStaging+name],
		})
	})
	return err
}

// RetrieveStagingTokenPKI retrieves and returns the PEM bytes for the staged
// DEP token exchange certificate and private key using name (DEP name).
func (s *KV) RetrieveStagingTokenPKI(ctx context.Context, name string) ([]byte, []byte, error) {
	var tokenPKIMap map[string][]byte
	err := kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		var err error
		tokenPKIMap, err = kv.GetMap(ctx, s.b, []string{
			keyPfxCertStaging + name,
			keyPfxKeyStaging + name,
		})
		return err
	})
	if errors.Is(err, kv.ErrKeyNotFound) {
		return nil, nil, fmt.Errorf("%w: %v", storage.ErrNotFound, err)
	} else if err != nil {
		return nil, nil, err
	}
	return tokenPKIMap[keyPfxCertStaging+name], tokenPKIMap[keyPfxKeyStaging+name], nil
}

// RetrieveCurrentTokenPKI reads and returns the PEM bytes for the
// previously-upstaged DEP token exchange certificate and private key
// using name (DEP name).
func (s *KV) RetrieveCurrentTokenPKI(ctx context.Context, name string) ([]byte, []byte, error) {
	var tokenPKIMap map[string][]byte
	err := kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, txn kv.CRUDBucket) error {
		var err error
		tokenPKIMap, err = kv.GetMap(ctx, s.b, []string{
			keyPfxCert + name,
			keyPfxKey + name,
		})
		return err
	})
	if errors.Is(err, kv.ErrKeyNotFound) {
		return nil, nil, fmt.Errorf("%w: %v", storage.ErrNotFound, err)
	} else if err != nil {
		return nil, nil, err
	}
	return tokenPKIMap[keyPfxCert+name], tokenPKIMap[keyPfxKey+name], nil
}

// QueryDEPNames queries and returns DEP names.
// [ErrOnlyOffset] is returned if cursor pagination is attempted.
// A default limit of 100 results is returned.
// Uses the staged certificate upload as a key for the DEP name.
// This means that a certificate has to have been staged (uploaded)
// for the DEP name to be query-able.
func (s *KV) QueryDEPNames(ctx context.Context, req *storage.DEPNamesQueryRequest) (*storage.DEPNamesQueryResult, error) {
	if err := req.Pagination.ValidErr(); err != nil {
		return nil, fmt.Errorf("pagination invalid: %w", err)
	}
	if req.Pagination != nil && req.Pagination.Cursor != nil {
		// cursor method not supported for this backend
		return nil, storage.ErrOnlyOffset
	}
	// grab the offset and limit from the pagination
	offset, limit := req.Pagination.DefaultOffsetLimit(100)

	var filter []string
	if req != nil && req.Filter != nil {
		filter = req.Filter.DEPNames
	}

	var ret []string
	var found int
	cancel := make(chan struct{})
	for key := range s.b.KeysPrefix(ctx, keyPfxCertStaging, cancel) {
		depName := key[len(keyPfxCertStaging):]

		candidate := true

		if len(filter) > 0 {
			for _, filterName := range filter {
				if filterName == depName {
					goto afterNotFound
				}
			}
			candidate = false
		afterNotFound:
		} // if there is no filter, then all keys are implicitly candidates

		if candidate {
			// only add if past offset and under limit
			if found >= offset && len(ret) < limit {
				ret = append(ret, depName)
			}
			found++

			// stop if hit limit
			if len(ret) >= limit {
				close(cancel)
				break
			}
		}
	}

	return &storage.DEPNamesQueryResult{DEPNames: ret}, nil
}
