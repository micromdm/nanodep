package api

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"strconv"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/tokenpki"

	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
)

type TokenPKIStagingRetriever interface {
	// RetrieveStagingTokenPKI retrieves and returns the PEM bytes for the staged
	// DEP token exchange certificate and private key for name (DEP name).
	RetrieveStagingTokenPKI(ctx context.Context, name string) (pemCert []byte, pemKey []byte, err error)
}

type TokenPKICurrentRetriever interface {
	// RetrieveCurrentTokenPKI reads and returns the PEM bytes for the
	// previously-upstaged DEP token exchange certificate and private
	// key using name (DEP name).
	RetrieveCurrentTokenPKI(ctx context.Context, name string) (pemCert []byte, pemKey []byte, err error)
}

type TokenPKIUpstager interface {
	// UpstageTokenPKI copies the "staging" PKI certificate and key to the current PKI certificate and key.
	// This allows key operations to use the newly uploaded key.
	// Note the OAuth tokens should also be changed at the same time.
	UpstageTokenPKI(ctx context.Context, name string) error
}

type TokenPKIStorer interface {
	// StoreTokenPKI stores the PEM bytes in pemCert and pemKey for name (DEP name).
	// These will be stored as the "staging" set to later be "upstaged."
	StoreTokenPKI(ctx context.Context, name string, pemCert []byte, pemKey []byte) error
}

type DecryptTokenPKIStorage interface {
	TokenPKIStagingRetriever
	TokenPKIUpstager
}

// PEMRSAPrivateKey returns key as a PEM block.
func PEMRSAPrivateKey(key *rsa.PrivateKey) []byte {
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(block)
}

// GetCertTokenPKIHandler generates a new private key and certificate for
// the token PKI exchange with the ABM/ASM/BE portal. Every call to this
// handler generates a new keypair and stores it. The PEM-encoded certificate
// is returned.
//
// Note the whole URL path is used as the DEP name. This necessitates
// stripping the URL prefix before using this handler. Also note we expose Go
// errors to the output as this is meant for "API" users.
func GetCertTokenPKIHandler(store TokenPKIStorer, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			defaultCN   = "depserver"
			defaultDays = 1
		)
		logger := ctxlog.Logger(r.Context(), logger)
		if r.URL.Path == "" {
			logger.Info("msg", "DEP name check", "err", "missing DEP name")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		logger = logger.With("name", r.URL.Path)
		var validityDays int64
		if daysArg := r.URL.Query().Get("validity_days"); daysArg == "" {
			logger.Debug("msg", "using default validity days", "days", defaultDays)
			validityDays = defaultDays
		} else {
			var err error
			validityDays, err = strconv.ParseInt(daysArg, 10, 64)
			if err != nil {
				logger.Info("msg", "validity_days check", "err", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}
		cn := r.URL.Query().Get("cn")
		if cn == "" {
			logger.Debug("msg", "using default CN", "cn", defaultCN)
			cn = defaultCN
		}
		key, cert, err := tokenpki.SelfSignedRSAKeypair(cn, validityDays)
		if err != nil {
			logger.Info("msg", "generating token keypair", "err", err)
			jsonError(w, err)
			return
		}
		pemCert := tokenpki.PEMCertificate(cert.Raw)
		err = store.StoreTokenPKI(r.Context(), r.URL.Path, pemCert, tokenpki.PEMRSAPrivateKey(key))
		if err != nil {
			logger.Info("msg", "storing token keypair", "err", err)
			jsonError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/x-pem-file")
		w.Header().Set("Content-Disposition", `attachment; filename="`+r.URL.Path+`.pem"`)
		w.Write(pemCert)
	}
}

// DecryptTokenPKIHandler reads the Apple-provided encrypted token ".p7m" file
// from the request body and decrypts it with the keypair generated from
// GetCertTokenPKIHandler.
//
// Note the whole URL path is used as the DEP name. This necessitates
// stripping the URL prefix before using this handler. Also note we expose Go
// errors to the output as this is meant for "API" users.
func DecryptTokenPKIHandler(store DecryptTokenPKIStorage, tokenStore AuthTokensStore, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := ctxlog.Logger(r.Context(), logger)
		if r.URL.Path == "" {
			logger.Info("msg", "DEP name check", "err", "missing DEP name")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		force := r.URL.Query().Get("force") == "1"
		logger = logger.With("name", r.URL.Path, "force", force)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Info("msg", "reading request body", "err", err)
			jsonError(w, err)
			return
		}
		defer r.Body.Close()
		certBytes, keyBytes, err := store.RetrieveStagingTokenPKI(r.Context(), r.URL.Path)
		if err != nil {
			logger.Info("msg", "retrieving token keypair", "err", err)
			jsonError(w, err)
			return
		}
		cert, err := tokenpki.CertificateFromPEM(certBytes)
		if err != nil {
			logger.Info("msg", "decoding retrieved certificate", "err", err)
			jsonError(w, err)
			return
		}
		key, err := tokenpki.RSAKeyFromPEM(keyBytes)
		if err != nil {
			logger.Info("msg", "decoding retrieved private key", "err", err)
			jsonError(w, err)
			return
		}
		tokenJSON, err := tokenpki.DecryptTokenJSON(bodyBytes, cert, key)
		if err != nil {
			logger.Info("msg", "decrypting auth tokens", "err", err)
			jsonError(w, err)
			return
		}
		tokens := new(client.OAuth1Tokens)
		err = json.Unmarshal(tokenJSON, tokens)
		if err != nil {
			logger.Info("msg", "decoding decrypted auth tokens", "err", err)
			jsonError(w, err)
			return
		}
		// decryption and unmarshal of tokens successful, now "upgrade"
		// our staging token PKI to the real thing.
		err = store.UpstageTokenPKI(r.Context(), r.URL.Path)
		if err != nil {
			logger.Info("msg", "upstaging token PKI", "err", err)
			jsonError(w, err)
			return
		}
		storeTokens(r.Context(), logger, r.URL.Path, tokens, tokenStore, w, force)
	}
}
