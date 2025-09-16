package api

import (
	"errors"
	"net/http"
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
// Note: this queries the DEP API "live" if a server_uuid query paramter is not provided.
// A cache of some sort may be a future strategy.
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

		serverUUID := r.URL.Query().Get("server_uuid")
		if serverUUID == "" {
			client := godep.NewClient(store)
			detail, err := client.AccountDetail(r.Context(), name)
			if err != nil {
				logger.Info("msg", "getting account detail", "err", err)
				jsonError(w, err)
				return
			}

			if detail.ServerUuid == nil {
				err = errors.New("nil server UUID")
				logger.Info("msg", "validating account detail", "err", err)
				jsonError(w, err)
				return
			}

			serverUUID = *detail.ServerUuid
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

		jti := newJTI()
		jwt, err := cryptoutil.NewMAIDJWT(key, serverUUID, time.Now(), jti)
		if err != nil {
			logger.Info("msg", "creating MAID JWT", "err", err)
			jsonError(w, err)
			return
		}

		w.Header().Set("X-Server-Uuid", serverUUID)
		w.Header().Set("X-Jwt-Jti", jti)
		w.Header().Set("Content-type", "application/jwt")
		_, err = w.Write([]byte(jwt))
		if err != nil {
			logger.Info("msg", "writing response body", "err", err)
			return
		}
	}
}
