// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"cmp"
	"context"
	"net/http"
	"net/url"

	"github.com/golang/gddo/httputil"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
)

type ctxKey struct{}

var baseURLKey ctxKey

func WithBaseURL(ctx context.Context, baseURL *url.URL) context.Context {
	if baseURL == nil {
		return ctx
	}
	baseURL.Scheme = "https" // Force https
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

// FlowBaseURL returns the base URL to be used for a self-service flow. It will
// either take the request URL, or an explicit base URL set in the context.
func FlowBaseURL(ctx context.Context, flow interface{ GetRequestURL() string }) (*url.URL, error) {
	if u := BaseURLFromContext(ctx); u != nil {
		return u, nil
	}
	u, err := url.Parse(flow.GetRequestURL())
	if err != nil {
		return nil, err
	}
	u.Path = "/"

	return u, nil
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

type HTTPClientProvider interface {
	HTTPClient(context.Context, ...httpx.ResilientOptions) *retryablehttp.Client
}
