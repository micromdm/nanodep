// Package apinext implements HTTP handlers for the NanoDEP API.
// It exists to break the circular dependency between the storage and api packages.
package apinext

import (
	"encoding/json"
	"net/http"

	"github.com/micromdm/nanolib/log"
)

// writeJSON encodes v to JSON writing to w using the HTTP status of header.
// An error during encoding is logged to logger if it is not nil.
// If header is 0 then none will be written to w (which defaults to 200).
// Nothing will be encoded nor written to w if v is nil.
func writeJSON(w http.ResponseWriter, v interface{}, header int, logger log.Logger) {
	w.Header().Set("Content-type", "application/json")

	if header > 0 {
		w.WriteHeader(header)
	}

	if v == nil {
		return
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	err := enc.Encode(v)
	if err != nil && logger != nil {
		logger.Info("msg", "encoding json", "err", err)
	}
}

// logAndWriteJSONError logs msg and err to logger as well as writes err to w as JSON.
// If header is 0 it will default to 500.
func logAndWriteJSONError(logger log.Logger, w http.ResponseWriter, msg string, err error, header int) {
	if logger != nil {
		logger.Info("msg", msg, "err", err)
	}

	errStr := "<nil error>"
	if err != nil {
		errStr = err.Error()
	}

	out := &ErrorResponseJson{Error: errStr}

	if header < 1 {
		header = 500
	}

	writeJSON(w, out, header, logger)
}
