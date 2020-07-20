package x

import (
	"net/http"
	"strconv"
)

const paginationMaxItems = 500

// ParsePagination parses limit and page from *http.Request with given limits and defaults.
func ParsePagination(r *http.Request) (page, limit int) {
	if offsetParam := r.URL.Query().Get("page"); offsetParam == "" {
		page = 0
	} else {
		if offset64, err := strconv.ParseInt(offsetParam, 10, 64); err != nil {
			page = 0
		} else {
			page = int(offset64)
		}
	}

	if limitParam := r.URL.Query().Get("limit"); limitParam == "" {
		limit = paginationMaxItems
	} else {
		if limit64, err := strconv.ParseInt(limitParam, 10, 64); err != nil {
			limit = paginationMaxItems
		} else {
			limit = int(limit64)
		}
	}

	if limit > paginationMaxItems {
		limit = paginationMaxItems
	}

	if limit < 0 {
		limit = 0
	}

	if page < 0 {
		page = 0
	}

	return
}
