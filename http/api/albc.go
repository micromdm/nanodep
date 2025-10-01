package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/micromdm/nanodep/albc"
)

type BypassCodeJSON struct {
	// Raw (hex encoded) form
	Raw string `json:"raw"`
	// Dash-separated "human readable" form
	Code string `json:"code"`
	// PBKDF2 derived hash of bypass code
	Hash string `json:"hash"`
}

// NewBypassCodeHandler returns a utility HTTP handler for working with Apple Activation Lock Bypass Codes.
func NewBypassCodeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var bc albc.BypassCode

		code := r.URL.Query().Get("code")
		raw := r.URL.Query().Get("raw")

		var err error

		if raw != "" && code != "" {
			jsonError(w, errors.New("raw or code but not both"))
			return
		} else if raw == "" && code == "" {
			// no raw or code provided, make a new random code
			bc, err = albc.New()
			if err != nil {
				jsonError(w, err)
				return
			}
		} else if raw != "" {
			// decode and use raw value
			b, err := hex.DecodeString(raw)
			if err != nil {
				jsonError(w, err)
				return
			}
			bc, err = albc.NewFromBytes(b)
			if err != nil {
				jsonError(w, err)
				return
			}
		} else if code != "" {
			// decode the dash-separated "human readable" form
			bc, err = albc.NewFromCode(code)
			if err != nil {
				jsonError(w, err)
				return
			}
		}

		out := &BypassCodeJSON{Raw: hex.EncodeToString(bc[:])}

		out.Code, err = bc.Code()
		if err != nil {
			jsonError(w, err)
			return
		}

		out.Hash, err = bc.Hash()
		if err != nil {
			jsonError(w, err)
			return
		}

		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}
