// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"cmp"
	"context"
	"net/http"
	"net/url"

	"github.com/golang/gddo/httputil"

	"github.com/ory/herodot"
)

type ctxKey struct{}

var baseURLKey ctxKey

// WithBaseURL stores the supplied base URL on the context. The URL's scheme
// is preserved verbatim — callers that need an https guarantee must enforce
// it themselves before calling.
func WithBaseURL(ctx context.Context, baseURL *url.URL) context.Context {
	if baseURL == nil {
		return ctx
	}
	return context.WithValue(ctx, baseURLKey, baseURL)
}

func BaseURLFromContext(ctx context.Context) *url.URL {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(baseURLKey); v != nil {
		if u, ok := v.(*url.URL); ok {
			return u
		}
	}
	return nil
}

// BaseURLStringFromContext returns the captured base URL as a string, or an
// empty string if none is set. Used to copy the captured URL onto a flow
// row at creation time (where it lives next to the rest of the flow rather
// than being plumbed via context to every consumer).
func BaseURLStringFromContext(ctx context.Context) string {
	if u := BaseURLFromContext(ctx); u != nil {
		return u.String()
	}
	return ""
}

// CourierBaseURL parses the courier base URL captured on a flow at init time
// (recovery.Flow.CourierBaseURL / verification.Flow.CourierBaseURL) and
// returns it as a *url.URL. When the captured value is empty or unparseable,
// it returns the supplied fallback (typically Config.SelfPublicURL).
func CourierBaseURL(courierBaseURL string, fallback *url.URL) *url.URL {
	if courierBaseURL != "" {
		if u, err := url.Parse(courierBaseURL); err == nil {
			return u
		}
	}
	return fallback
}

func RequestURL(r *http.Request) *url.URL {
	source := *r.URL
	source.Host = cmp.Or(source.Host, r.Header.Get("X-Forwarded-Host"), r.Host)

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

// RequestBaseURL returns the customer-facing base URL of the request.
//
// If a base URL was captured on the request context (e.g. by a proxy-aware
// middleware that validated an Ory-Base-URL-Rewrite / X-Ory-Original-Host
// header), that value wins — it is the URL the end user's browser actually
// used, which may differ from the host this service was reached at. This is
// the value an OIDC/SAML callback must be redirected back to.
//
// Otherwise it falls back to the request's own scheme://host[:port] (honoring
// X-Forwarded-Host / X-Forwarded-Proto via RequestURL), with no path, query,
// or fragment.
func RequestBaseURL(r *http.Request) string {
	if captured := BaseURLStringFromContext(r.Context()); captured != "" {
		return captured
	}
	u := RequestURL(r)
	return (&url.URL{Scheme: u.Scheme, Host: u.Host}).String()
}

// SendFlowCompletedAsRedirectOrJSON should be used when a login, registration, ... flow has been completed successfully.
// It will redirect the user to the provided URL if the request accepts HTML, or return a JSON response if the request is
// an SPA request
func SendFlowCompletedAsRedirectOrJSON(
	w http.ResponseWriter, r *http.Request, writer herodot.Writer, out interface{}, redirectTo string,
) {
	sendFlowAsRedirectOrJSON(w, r, writer, out, redirectTo, http.StatusOK)
}

// SendFlowErrorAsRedirectOrJSON should be used when a login, registration, ... flow has errors (e.g. validation errors
// or missing data) and should be redirected to the provided URL if the request accepts HTML, or return a JSON response
// if the request is an SPA request.
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
