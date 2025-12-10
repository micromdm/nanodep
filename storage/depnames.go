package storage

import "context"

// DEPNamesQueryFilter is the filter parameters for querying DEP names.
type DEPNamesQueryFilter struct {
	// DEPNames specifies which DEP names to query for.
	// DEP names in this list which exists are returned.
	DEPNames []string `json:"dep_names"`
}

// DEPNamesQueryRequest is the parameters for querying DEP names.
type DEPNamesQueryRequest struct {
	Filter     *DEPNamesQueryFilter `json:"filter,omitempty"`
	Pagination *Pagination          `json:"pagination,omitempty"`
}

// DEPNamesQueryResult is the resulting paginated of the DEP names query.
type DEPNamesQueryResult struct {
	DEPNames []string `json:"dep_names"`

	PaginationNextCursor
}

type DEPNamesQuery interface {
	// QueryDEPNames queries and returns DEP names.
	QueryDEPNames(ctx context.Context, req *DEPNamesQueryRequest) (*DEPNamesQueryResult, error)
}
