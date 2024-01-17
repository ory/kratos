// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/http"
	"net/url"

	"github.com/ory/herodot"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/pagination/migrationpagination"
	"github.com/ory/x/pagination/pagepagination"
)

// The format we need to use in the Page tokens, as it's the only format that is understood by all DBs
const MapPaginationDateFormat = "2006-01-02 15:04:05.99999"

// ParsePagination parses limit and page from *http.Request with given limits and defaults.
func ParsePagination(r *http.Request) (page, itemsPerPage int) {
	return migrationpagination.NewDefaultPaginator().ParsePagination(r)
}

func PaginationHeader(w http.ResponseWriter, u url.URL, total int64, page, itemsPerPage int) {
	migrationpagination.PaginationHeader(w, &u, total, page, itemsPerPage)
}

type Page struct {
	Page, ItemsPerPage int
}

var PagePaginationLimit = 1000

func ParseKeysetOrPagePagination(r *http.Request) ([]keysetpagination.Option, *Page, error) {
	q := r.URL.Query()
	// If we have any new-style pagination parameters, use those and ignore the rest.
	if q.Has("page_token") || q.Has("page_size") {
		keyset, err := keysetpagination.Parse(q, keysetpagination.NewStringPageToken)
		if err != nil {
			return nil, nil, herodot.ErrBadRequest.WithReason(err.Error())
		}
		return keyset, nil, nil
	}
	// allow fallback page pagination with upper limit
	if q.Has("page") {
		paginator := pagepagination.PagePaginator{MaxItems: 500, DefaultItems: 250}
		page, perPage := paginator.ParsePagination(r)
		if page*perPage > PagePaginationLimit {
			return nil, nil, herodot.ErrBadRequest.WithReasonf("Legacy pagination is not supported for enumerating over %d items. Please switch to using page_token and page_size.", PagePaginationLimit)
		}
		return nil, &Page{page, perPage}, nil
	}
	// Allow passing per_page instead of page_size if only the former is set...
	if q.Has("per_page") && !q.Has("page_size") {
		q.Set("page_size", q.Get("per_page"))
		q.Del("per_page")
		r.URL.RawQuery = q.Encode()
	}
	// ... and defaul to keyset pagination
	keyset, err := keysetpagination.Parse(q, keysetpagination.NewStringPageToken)
	if err != nil {
		return nil, nil, herodot.ErrBadRequest.WithReason(err.Error())
	}
	return keyset, nil, nil
}
