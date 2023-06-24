// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ory/kratos/text"

	"github.com/golang/gddo/httputil"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/stringsx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
)

type secureRedirectOptions struct {
	allowlist       []url.URL
	defaultReturnTo *url.URL
	returnTo        string
	sourceURL       string
}

type SecureRedirectOption func(*secureRedirectOptions)

// SecureRedirectAllowURLs allows the given URLs for redirects.
func SecureRedirectAllowURLs(urls []url.URL) SecureRedirectOption {
	return func(o *secureRedirectOptions) {
		o.allowlist = append(o.allowlist, urls...)
	}
}

// SecureRedirectUseSourceURL uses the given source URL (checks the `?return_to` value)
// instead of r.URL.
func SecureRedirectUseSourceURL(source string) SecureRedirectOption {
	return func(o *secureRedirectOptions) {
		o.sourceURL = source
	}
}

// SecureRedirectReturnTo uses the provided URL to redirect the user to it.
func SecureRedirectReturnTo(returnTo string) SecureRedirectOption {
	return func(o *secureRedirectOptions) {
		o.returnTo = returnTo
	}
}

// SecureRedirectAllowSelfServiceURLs allows the caller to define `?return_to=` values
// which contain the server's URL and `/self-service` path prefix. Useful for redirecting
// to the login endpoint, for example.
func SecureRedirectAllowSelfServiceURLs(publicURL *url.URL) SecureRedirectOption {
	return func(o *secureRedirectOptions) {
		o.allowlist = append(o.allowlist, *urlx.AppendPaths(publicURL, "/self-service"))
	}
}

// SecureRedirectOverrideDefaultReturnTo overrides the defaultReturnTo address specified
// as the second arg.
func SecureRedirectOverrideDefaultReturnTo(defaultReturnTo *url.URL) SecureRedirectOption {
	return func(o *secureRedirectOptions) {
		o.defaultReturnTo = defaultReturnTo
	}
}

// SecureRedirectToIsAllowedHost validates if the redirect_to param is allowed for a given wildcard
func SecureRedirectToIsAllowedHost(returnTo *url.URL, allowed url.URL) bool {
	if allowed.Host != "" && allowed.Host[:1] == "*" {
		return strings.HasSuffix(strings.ToLower(returnTo.Host), strings.ToLower(allowed.Host)[1:])
	}
	return strings.EqualFold(allowed.Host, returnTo.Host)
}

// TakeOverReturnToParameter carries over the return_to parameter to a new URL
// If `from` does not contain the `return_to` query parameter, the first non-empty value from `fallback` is used instead.
func TakeOverReturnToParameter(from string, to string, fallback ...string) (string, error) {
	fromURL, err := url.Parse(from)
	if err != nil {
		return "", err
	}
	returnTo := stringsx.Coalesce(append([]string{fromURL.Query().Get("return_to")}, fallback...)...)
	// Empty return_to parameter, return early
	if returnTo == "" {
		return to, nil
	}
	toURL, err := url.Parse(to)
	if err != nil {
		return "", err
	}
	toQuery := toURL.Query()
	toQuery.Set("return_to", returnTo)
	toURL.RawQuery = toQuery.Encode()
	return toURL.String(), nil
}

// SecureRedirectTo implements a HTTP redirector who mitigates open redirect vulnerabilities by
// working with allow lists.
func SecureRedirectTo(r *http.Request, defaultReturnTo *url.URL, opts ...SecureRedirectOption) (returnTo *url.URL, err error) {
	o := &secureRedirectOptions{defaultReturnTo: defaultReturnTo}
	for _, opt := range opts {
		opt(o)
	}

	if len(o.allowlist) == 0 {
		return o.defaultReturnTo, nil
	}

	source := RequestURL(r)
	if o.sourceURL != "" {
		source, err = url.ParseRequestURI(o.sourceURL)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to parse the original request URL: %s", err))
		}
	}

	rawReturnTo := stringsx.Coalesce(o.returnTo, source.Query().Get("return_to"))
	if rawReturnTo == "" {
		return o.defaultReturnTo, nil
	}

	returnTo, err = url.Parse(rawReturnTo)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReasonf("Unable to parse the return_to query parameter as an URL: %s", err))
	}

	returnTo.Host = stringsx.Coalesce(returnTo.Host, o.defaultReturnTo.Host)
	returnTo.Scheme = stringsx.Coalesce(returnTo.Scheme, o.defaultReturnTo.Scheme)

	for _, allowed := range o.allowlist {
		if strings.EqualFold(allowed.Scheme, returnTo.Scheme) &&
			SecureRedirectToIsAllowedHost(returnTo, allowed) &&
			strings.HasPrefix(
				stringsx.Coalesce(returnTo.Path, "/"),
				stringsx.Coalesce(allowed.Path, "/")) {
			return returnTo, nil
		}
	}

	return nil, errors.WithStack(herodot.ErrBadRequest.
		WithID(text.ErrIDRedirectURLNotAllowed).
		WithReasonf("Requested return_to URL \"%s\" is not allowed.", returnTo).
		WithDebugf("Allowed domains are: %v", o.allowlist))
}

func SecureContentNegotiationRedirection(
	w http.ResponseWriter, r *http.Request, out interface{},
	requestURL string, writer herodot.Writer, c *config.Config,
	opts ...SecureRedirectOption,
) error {
	switch httputil.NegotiateContentType(r, []string{
		"text/html",
		"application/json",
	}, "text/html") {
	case "application/json":
		writer.Write(w, r, out)
	case "text/html":
		fallthrough
	default:
		ret, err := SecureRedirectTo(r, c.SelfServiceBrowserDefaultReturnTo(r.Context()),
			append([]SecureRedirectOption{
				SecureRedirectUseSourceURL(requestURL),
				SecureRedirectAllowURLs(c.SelfServiceBrowserAllowedReturnToDomains(r.Context())),
				SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL(r.Context())),
			}, opts...)...,
		)
		if err != nil {
			return err
		}

		http.Redirect(w, r, ret.String(), http.StatusSeeOther)
	}

	return nil
}

func ContentNegotiationRedirection(
	w http.ResponseWriter, r *http.Request, out interface{}, writer herodot.Writer, returnTo string,
) {
	switch httputil.NegotiateContentType(r, []string{
		"text/html",
		"application/json",
	}, "text/html") {
	case "application/json":
		writer.Write(w, r, out)
	case "text/html":
		fallthrough
	default:
		http.Redirect(w, r, returnTo, http.StatusSeeOther)
	}
}
