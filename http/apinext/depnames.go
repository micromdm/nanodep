package apinext

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/micromdm/nanodep/storage"

	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
)

// NewQueryDEPNamesHandler returns a handler that queries DEP names.
func NewQueryDEPNamesHandler(store storage.DEPNamesQuery, logger log.Logger) http.HandlerFunc {
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
		q := &storage.DEPNamesQueryRequest{
			Filter: &storage.DEPNamesQueryFilter{
				DEPNames: r.URL.Query()["dep_name"],
			},
			Pagination: p,
		}

		// perform query
		ret, err := store.QueryDEPNames(r.Context(), q)
		if err != nil {
			logAndWriteJSONError(logger, w, "querying DEP names", err, 0)
			return
		}

		// log the success
		logger.Debug("msg", fmt.Sprintf("queried DEP names: %d", len(ret.DEPNames)))

		// output the return
		writeJSON(w, ret, http.StatusOK, logger)
	}
}
