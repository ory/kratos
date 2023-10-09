// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
