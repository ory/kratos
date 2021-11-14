package x

import (
	"net/http"
	"net/url"

	"github.com/ory/x/pagination/pagepagination"
)

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
