package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/micromdm/nanodep/client"

	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
)

// ErrCKMismatch occurs when an incoming consumer key does not match the
// previous consumer key. It is intended to catch an accidental overwrite
// of an existing DEP name's tokens during a renewal.
// However note that a different DEP username renewing the tokens can
// also trigger it for a legitimate renewal.
var ErrCKMismatch = errors.New("mismatched consumer key")

type AuthTokensStore interface {
	client.AuthTokensRetriever
	AuthTokensStorer
}

type AuthTokensStorer interface {
	StoreAuthTokens(ctx context.Context, name string, tokens *client.OAuth1Tokens) error
}

// RetrieveAuthTokensHandler returns the DEP server OAuth1 tokens for the DEP
// name in the path.
//
// Note the whole URL path is used as the DEP name. This necessitates
// stripping the URL prefix before using this handler. Also note we expose Go
// errors to the output as this is meant for "API" users.
func RetrieveAuthTokensHandler(store client.AuthTokensRetriever, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := ctxlog.Logger(r.Context(), logger)
		if r.URL.Path == "" {
			logger.Info("msg", "DEP name check", "err", "missing DEP name")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		logger = logger.With("name", r.URL.Path)
		tokens, err := store.RetrieveAuthTokens(r.Context(), r.URL.Path)
		if err != nil {
			logger.Info("msg", "retrieving auth tokens", "err", err)
			jsonError(w, err)
			return
		}
		w.Header().Set("Content-type", "application/json")
		err = json.NewEncoder(w).Encode(tokens)
		if err != nil {
			logger.Info("msg", "encoding response body", "err", err)
			return
		}
	}
}

// StoreAuthTokensHandler reads DEP server OAuth1 tokens as a JSON body and
// saves them using store.
//
// Note the whole URL path is used as the DEP name. This necessitates
// stripping the URL prefix before using this handler. Also note we expose Go
// errors to the output as this is meant for "API" users.
func StoreAuthTokensHandler(store AuthTokensStore, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := ctxlog.Logger(r.Context(), logger)
		if r.URL.Path == "" {
			logger.Info("msg", "DEP name check", "err", "missing DEP name")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		force := r.URL.Query().Get("force") == "1"
		logger = logger.With("name", r.URL.Path, "force", force)
		tokens := new(client.OAuth1Tokens)
		err := json.NewDecoder(r.Body).Decode(tokens)
		if err != nil {
			logger.Info("msg", "decoding request body", "err", err)
			jsonError(w, err)
			return
		}
		defer r.Body.Close()
		storeTokens(r.Context(), logger, r.URL.Path, tokens, store, w, force)
	}
}

func storeTokens(ctx context.Context, logger log.Logger, name string, tokens *client.OAuth1Tokens, store AuthTokensStore, w http.ResponseWriter, force bool) {
	if !tokens.Valid() {
		logger.Info("msg", "checking auth token validity", "err", "invalid tokens")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	logger = logger.With("consumer_key", tokens.ConsumerKey)
	if !force {
		prevTokens, err := store.RetrieveAuthTokens(ctx, name)
		if err != nil {
			logger.Debug(
				"msg", "error retrieving auth tokens; proceeding to store",
				"err", err,
			)
		} else if prevTokens != nil && prevTokens.ConsumerKey != tokens.ConsumerKey {
			logger.Info(
				"msg", "checking consumer key (use force to bypass)",
				"err", ErrCKMismatch,
				"prev_consumer_key", prevTokens.ConsumerKey,
			)
			jsonError(w, ErrCKMismatch)
			return
		}
	}
	err := store.StoreAuthTokens(ctx, name, tokens)
	if err != nil {
		logger.Info("msg", "storing auth tokens", "err", err)
		jsonError(w, err)
		return
	}
	logger.Debug("msg", "stored auth tokens")
	w.Header().Set("Content-type", "application/json")
	err = json.NewEncoder(w).Encode(tokens)
	if err != nil {
		logger.Info("msg", "encoding response body", "err", err)
		return
	}
}
