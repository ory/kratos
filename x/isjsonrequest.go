// Copyright © 2022 Ory Corp

package x

import (
	"net/http"

	"github.com/golang/gddo/httputil"
)

var offers = []string{"text/html", "text/*", "*/*", "application/json"}
var defaultOffer = "text/html"

func IsJSONRequest(r *http.Request) bool {
	return httputil.NegotiateContentType(r, offers, defaultOffer) == "application/json" ||
		r.Header.Get("Content-Type") == "application/json"
}

func IsBrowserRequest(r *http.Request) bool {
	return httputil.NegotiateContentType(r, offers, defaultOffer) == "text/html"
}
