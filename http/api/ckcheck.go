package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/micromdm/nanodep/client"
)

var CKMismatch = errors.New("mismatched consumer key")

// CKCheck is a wrapper around token storage to validate the consumer key.
// This attempts to prevent overwriting of incorrect auth tokens.
type CKCheck struct {
	client.AuthTokensRetriever
	AuthTokensStorer
}

type AuthTokensStore interface {
	client.AuthTokensRetriever
	AuthTokensStorer
}

// NewCKCheck creates a new CKCheck.
func NewCKCheck(store AuthTokensStore) *CKCheck {
	return &CKCheck{store, store}
}

// StoreAuthTokens first retrieves the existing auth tokens and checks to make
// sure the consumer key of the provided auth tokens match before then storing
// the provided auth tokens.
func (t *CKCheck) StoreAuthTokens(ctx context.Context, name string, tokens *client.OAuth1Tokens) error {
	prevTokens, err := t.AuthTokensRetriever.RetrieveAuthTokens(ctx, name)
	if err != nil {
		return fmt.Errorf("retrieving auth tokens: %w", err)
	}
	if prevTokens.ConsumerKey != tokens.ConsumerKey && prevTokens.ConsumerKey != "" {
		return CKMismatch
	}
	return t.AuthTokensStorer.StoreAuthTokens(ctx, name, tokens)
}
