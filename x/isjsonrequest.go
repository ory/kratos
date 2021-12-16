package x

import (
	"net/http"
)

func IsJSONRequest(r *http.Request) bool {
	return AcceptsJSON(r) ||
		r.Header.Get("Content-Type") == "application/json"
}

func IsBrowserRequest(r *http.Request) bool {
	return AcceptsContentType(r, "text/html") ||
		AcceptsContentType(r, "text/*") ||
		AcceptsContentType(r, "*/*") ||
		!AcceptsJSON(r)
}
