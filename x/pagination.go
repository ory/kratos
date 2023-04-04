// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/http"
	"net/url"

	"github.com/ory/x/pagination/migrationpagination"

	"github.com/ory/x/pagination/pagepagination"
)

// swagger:model pagination
type PaginationParams struct {
	// Items per Page
	//
	// This is the number of items per page.
	//
	// required: false
	// in: query
	// default: 250
	// min: 1
	// max: 1000
	PerPage int `json:"per_page"`

	// Pagination Page
	//
	// This value is currently an integer, but it is not sequential. The value is not the page number, but a
	// reference. The next page can be any number and some numbers might return an empty list.
	//
	// For example, page 2 might not follow after page 1. And even if page 3 and 5 exist, but page 4 might not exist.
	//
	// required: false
	// in: query
	// default: 1
	// min: 1
	Page int `json:"page"`
}

// ParsePagination parses limit and page from *http.Request with given limits and defaults.
func ParsePagination(r *http.Request) (page, itemsPerPage int) {
	return migrationpagination.NewDefaultPaginator().ParsePagination(r)
}

func PaginationHeader(w http.ResponseWriter, u *url.URL, total int64, page, itemsPerPage int) {
	pagepagination.PaginationHeader(w, u, total, page, itemsPerPage)
}
