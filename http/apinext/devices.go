package apinext

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/micromdm/nanodep/storage"

	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
)

// NewQueryDevicesHandler returns a handler that queries stored devices.
func NewQueryDevicesHandler(store storage.DevicesQuery, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := new(storage.Pagination)

		logger := ctxlog.Logger(r.Context(), logger)

		var err error
		// extract and set pagination limit
		if limitRaw := r.URL.Query().Get("limit"); limitRaw != "" {
			limit, err := strconv.Atoi(limitRaw)
			if err != nil {
				logAndWriteJSONError(logger, w, "converting limit param", err, http.StatusBadRequest)
				return
			}

			p.Limit = &limit
		}

		// extract and set pagination offset
		if offsetRaw := r.URL.Query().Get("offset"); offsetRaw != "" {
			offset, err := strconv.Atoi(offsetRaw)
			if err != nil {
				logAndWriteJSONError(logger, w, "converting offset param", err, http.StatusBadRequest)
				return
			}

			p.Offset = &offset
		}

		// extract and set pagination cursor
		if cursorRaw := r.URL.Query().Get("cursor"); cursorRaw != "" {
			p.Cursor = &cursorRaw
		}

		// assemble the query request
		q := &storage.DevicesQueryRequest{
			Filter: &storage.DevicesQueryFilter{
				DEPNames:      r.URL.Query()["dep_name"],
				SerialNumbers: r.URL.Query()["serial"],
			},
			Pagination: p,
		}

		// perform query
		ret, err := store.QueryDevices(r.Context(), q)
		if err != nil {
			logAndWriteJSONError(logger, w, "querying devices", err, 0)
			return
		}

		// log the success
		logger.Debug("msg", fmt.Sprintf("queried devices: %d", len(ret.Devices)))

		// output the return
		writeJSON(w, ret, http.StatusOK, logger)
	}
}

// NewDeleteDevicesHandler returns a handler that deletes stored devices
// for a single DEP name. The dep_name query parameter is required (and
// singular); serial may repeat. An empty serial list is a no-op.
// Responds 204 with an empty body on success.
func NewDeleteDevicesHandler(store storage.DeviceDeleter, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := ctxlog.Logger(r.Context(), logger)

		depNames := r.URL.Query()["dep_name"]
		if len(depNames) != 1 || depNames[0] == "" {
			logAndWriteJSONError(logger, w, "deleting devices", errors.New("single non-empty dep_name query parameter required"), http.StatusBadRequest)
			return
		}
		depName := depNames[0]
		serials := r.URL.Query()["serial"]

		if err := store.DeleteDevices(r.Context(), depName, serials); err != nil {
			logAndWriteJSONError(logger, w, "deleting devices", err, 0)
			return
		}

		logger.Debug("msg", fmt.Sprintf("deleted devices: dep_name=%s count=%d", depName, len(serials)))

		w.WriteHeader(http.StatusNoContent)
	}
}
