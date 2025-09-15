package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/micromdm/nanodep/cryptoutil"
	"github.com/micromdm/nanodep/godep"
	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
)

type MAIDJWTStorage interface {
	TokenPKICurrentRetriever
	godep.ClientStorage
}

// NewMAIDJWTHandler returns a JWT for DEP Access Management.
// This JWT should be returned for use with an MDM client's CheckIn "GetToken" message.
// Note: this queries the DEP API "live." A cache of some sort may be a future strategy.
func NewMAIDJWTHandler(store MAIDJWTStorage, logger log.Logger, newJTI func() string) http.HandlerFunc {
	if store == nil {
		panic("nil store")
	}
	if logger == nil {
		panic("nil logger")
	}
	if newJTI == nil {
		panic("nil new JTI")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		logger := ctxlog.Logger(r.Context(), logger)
		if r.URL.Path == "" {
			logger.Info("msg", "DEP name check", "err", "missing DEP name")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		name := r.URL.Path
		logger = logger.With("name", name)

		client := godep.NewClient(store, nil)
		detail, err := client.AccountDetail(r.Context(), name)
		if err != nil {
			logger.Info("msg", "getting account detail", "err", err)
			jsonError(w, err)
			return
		}

		json.NewEncoder(os.Stdout).Encode(detail)

		if detail.ServerUuid == nil {
			err = errors.New("nil server UUID")
			logger.Info("msg", "validating account detail", "err", err)
			jsonError(w, err)
			return
		}

		_, keyBytes, err := store.RetrieveCurrentTokenPKI(r.Context(), name)
		if err != nil {
			logger.Info("msg", "retrieving token keypair", "err", err)
			jsonError(w, err)
			return
		}

		key, err := cryptoutil.RSAKeyFromPEM(keyBytes)
		if err != nil {
			logger.Info("msg", "decoding retrieved private key", "err", err)
			jsonError(w, err)
			return
		}

		jwt, err := cryptoutil.NewMAIDJWT(key, *detail.ServerUuid, time.Now(), newJTI())
		if err != nil {
			logger.Info("msg", "creating MAID JWT", "err", err)
			jsonError(w, err)
			return
		}

		w.Header().Set("Content-type", "application/jwt")
		_, err = w.Write([]byte(jwt))
		if err != nil {
			logger.Info("msg", "writing response body", "err", err)
			return
		}
	}
}
