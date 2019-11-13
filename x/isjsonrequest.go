package x

import (
	"net/http"

	"github.com/golang/gddo/httputil"
)

func IsJSONRequest(r *http.Request) bool {
	return httputil.NegotiateContentType(r,
		[]string{"application/json", "text/html", "text/*", "*/*"},
		"text/*",
	) == "application/json"
}
