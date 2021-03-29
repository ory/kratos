package x

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const paginationMaxItems = 1000
const paginationDefaultItems = 250

// ParsePagination parses limit and page from *http.Request with given limits and defaults.
func ParsePagination(r *http.Request) (page, itemsPerPage int) {
	if offsetParam := r.URL.Query().Get("page"); offsetParam == "" {
		page = 0
	} else {
		if offset, err := strconv.ParseInt(offsetParam, 10, 0); err != nil {
			page = 0
		} else {
			page = int(offset)
		}
	}

	if limitParam := r.URL.Query().Get("per_page"); limitParam == "" {
		itemsPerPage = paginationDefaultItems
	} else {
		if limit, err := strconv.ParseInt(limitParam, 10, 0); err != nil {
			itemsPerPage = paginationDefaultItems
		} else {
			itemsPerPage = int(limit)
		}
	}

	if itemsPerPage > paginationMaxItems {
		itemsPerPage = paginationMaxItems
	}

	if itemsPerPage < 1 {
		itemsPerPage = 1
	}

	if page < 0 {
		page = 0
	}

	return
}

func header(u *url.URL, rel string, limit, page int64) string {
	q := u.Query()
	q.Set("per_page", fmt.Sprintf("%d", limit))
	q.Set("page", fmt.Sprintf("%d", page/limit))
	u.RawQuery = q.Encode()
	return fmt.Sprintf("<%s>; rel=\"%s\"", u.String(), rel)
}

func PaginationHeader(w http.ResponseWriter, u *url.URL, total int64, page, itemsPerPage int) {
	if itemsPerPage <= 0 {
		itemsPerPage = 1
	}

	itemsPerPage64 := int64(itemsPerPage)
	offset := int64(page) * itemsPerPage64

	// lastOffset will either equal the offset required to contain the remainder,
	// or the limit.
	var lastOffset int64
	if total%itemsPerPage64 == 0 {
		lastOffset = total - itemsPerPage64
	} else {
		lastOffset = (total / itemsPerPage64) * itemsPerPage64
	}

	// Check for last page
	if offset >= lastOffset {
		if total == 0 {
			w.Header().Set("Link", strings.Join([]string{
				header(u, "first", itemsPerPage64, 0),
				header(u, "next", itemsPerPage64, ((offset/itemsPerPage64)+1)*itemsPerPage64),
				header(u, "prev", itemsPerPage64, ((offset/itemsPerPage64)-1)*itemsPerPage64),
			}, ","))
			return
		}

		if total < itemsPerPage64 {
			w.Header().Set("link", header(u, "first", total, 0))
			return
		}

		w.Header().Set("Link", strings.Join([]string{
			header(u, "first", itemsPerPage64, 0),
			header(u, "prev", itemsPerPage64, lastOffset-itemsPerPage64),
		}, ","))
		return
	}

	if offset < itemsPerPage64 {
		w.Header().Set("Link", strings.Join([]string{
			header(u, "next", itemsPerPage64, itemsPerPage64),
			header(u, "last", itemsPerPage64, lastOffset),
		}, ","))
		return
	}

	w.Header().Set("Link", strings.Join([]string{
		header(u, "first", itemsPerPage64, 0),
		header(u, "next", itemsPerPage64, ((offset/itemsPerPage64)+1)*itemsPerPage64),
		header(u, "prev", itemsPerPage64, ((offset/itemsPerPage64)-1)*itemsPerPage64),
		header(u, "last", itemsPerPage64, lastOffset),
	}, ","))
}
