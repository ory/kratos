package x

import (
	"net/http"
	"net/url"

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
	// required: false
	// in: query
	// default: 0
	// min: 0
	Page int `json:"page"`
}

const paginationMaxItems = 1000
const paginationDefaultItems = 250

var paginator = &pagepagination.PagePaginator{
	MaxItems:     paginationMaxItems,
	DefaultItems: paginationDefaultItems,
}

// ParsePagination parses limit and page from *http.Request with given limits and defaults.
func ParsePagination(r *http.Request) (page, itemsPerPage int) {
	return paginator.ParsePagination(r)
}

func PaginationHeader(w http.ResponseWriter, u *url.URL, total int64, page, itemsPerPage int) {
	pagepagination.PaginationHeader(w, u, total, page, itemsPerPage)
}
