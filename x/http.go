// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ory/x/httpx"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/golang/gddo/httputil"

	"github.com/ory/herodot"

	"github.com/ory/x/stringsx"
)

func RequestURL(r *http.Request) *url.URL {
	source := *r.URL
	source.Host = stringsx.Coalesce(source.Host, r.Header.Get("X-Forwarded-Host"), r.Host)

	if proto := r.Header.Get("X-Forwarded-Proto"); len(proto) > 0 {
		source.Scheme = proto
	}

	if source.Scheme == "" {
		source.Scheme = "https"
		if r.TLS == nil {
			source.Scheme = "http"
		}
	}

	return &source
}

// SendFlowCompletedAsRedirectOrJSON should be used when a login, registration, ... flow has been completed successfully.
// It will redirect the user to the provided URL if the request accepts HTML, or return a JSON response if the request
func SendFlowCompletedAsRedirectOrJSON(
	w http.ResponseWriter, r *http.Request, writer herodot.Writer, out interface{}, redirectTo string,
) {
	sendFlowAsRedirectOrJSON(w, r, writer, out, redirectTo, http.StatusOK)
}

// SendFlowErrorAsRedirectOrJSON should be used when a login, registration, ... flow has errors (e.g. validation errors
// or missing data) and should be redirected to the provided URL if the request accepts HTML, or return a JSON response
// if the request.
func SendFlowErrorAsRedirectOrJSON(
	w http.ResponseWriter, r *http.Request, writer herodot.Writer, out interface{}, redirectTo string,
) {
	sendFlowAsRedirectOrJSON(w, r, writer, out, redirectTo, http.StatusBadRequest)
}

func sendFlowAsRedirectOrJSON(
	w http.ResponseWriter, r *http.Request, writer herodot.Writer, out interface{}, redirectTo string, jsonResponseCode int,
) {
	switch httputil.NegotiateContentType(r, []string{
		"text/html",
		"application/json",
	}, "text/html") {
	case "application/json":
		if err, ok := out.(error); ok {
			writer.WriteError(w, r, err)
			return
		}

		writer.WriteCode(w, r, jsonResponseCode, out)
	case "text/html":
		fallthrough
	default:
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}
}

func AcceptsJSON(r *http.Request) bool {
	return httputil.NegotiateContentType(r, []string{
		"text/html",
		"application/json",
	}, "text/html") == "application/json"
}

type HTTPClientProvider interface {
	HTTPClient(context.Context, ...httpx.ResilientOptions) *retryablehttp.Client
}
