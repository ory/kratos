// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/http"
	"net/url"

	"github.com/ory/x/pagination/migrationpagination"
)

// ParsePagination parses limit and page from *http.Request with given limits and defaults.
func ParsePagination(r *http.Request) (page, itemsPerPage int) {
	return migrationpagination.NewDefaultPaginator().ParsePagination(r)
}

func PaginationHeader(w http.ResponseWriter, u *url.URL, total int64, page, itemsPerPage int) {
	migrationpagination.PaginationHeader(w, u, total, page, itemsPerPage)
}
