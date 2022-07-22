package api

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/log"
	"github.com/micromdm/nanodep/log/ctxlog"
	"github.com/micromdm/nanodep/tokenpki"
)

type TokenPKIRetriever interface {
	RetrieveTokenPKI(context.Context, string) ([]byte, []byte, error)
}

type TokenPKIStorer interface {
	StoreTokenPKI(context.Context, string, []byte, []byte) error
}

const (
	defaultCN = "depserver"
)

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
		logger := ctxlog.Logger(r.Context(), logger)
		if r.URL.Path == "" {
			logger.Info("msg", "DEP name check", "err", "missing DEP name")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		logger = logger.With("name", r.URL.Path)
		daysArg := r.URL.Query().Get("validity_days")
		if daysArg == "" {
			logger.Info("msg", "validity_days check", "err", "missing validity_days")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		validityDays, err := strconv.ParseInt(daysArg, 10, 64)
		if err != nil {
			logger.Info("msg", "validity_days check", "err", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
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
		err = store.StoreTokenPKI(r.Context(), r.URL.Path, pemCert, PEMRSAPrivateKey(key))
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

// RSAKeyFromPEM decodes a PEM RSA private key.
func RSAKeyFromPEM(key []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("PEM type is not RSA PRIVATE KEY")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// DecryptTokenPKIHandler reads the Apple-provided encrypted token ".p7m" file
// from the request body and decrypts it with the keypair generated from
// GetCertTokenPKIHandler.
//
// Note the whole URL path is used as the DEP name. This necessitates
// stripping the URL prefix before using this handler. Also note we expose Go
// errors to the output as this is meant for "API" users.
func DecryptTokenPKIHandler(store TokenPKIRetriever, tokenStore AuthTokensStorer, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := ctxlog.Logger(r.Context(), logger)
		if r.URL.Path == "" {
			logger.Info("msg", "DEP name check", "err", "missing DEP name")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		logger = logger.With("name", r.URL.Path)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Info("msg", "reading request body", "err", err)
			jsonError(w, err)
			return
		}
		defer r.Body.Close()
		certBytes, keyBytes, err := store.RetrieveTokenPKI(r.Context(), r.URL.Path)
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
		key, err := RSAKeyFromPEM(keyBytes)
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
		if !tokens.Valid() {
			logger.Info("msg", "checking auth token validity", "err", "invalid tokens")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err = tokenStore.StoreAuthTokens(r.Context(), r.URL.Path, tokens)
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
}
