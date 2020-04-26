package x

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/golang/gddo/httputil"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/stringsx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
)

type secureRedirectOptions struct {
	whitelist       []url.URL
	defaultReturnTo *url.URL
	sourceURL       string
}

type SecureRedirectOption func(*secureRedirectOptions)

// SecureRedirectAllowURLs whitelists the given URLs for redirects.
func SecureRedirectAllowURLs(urls []url.URL) SecureRedirectOption {
	return func(o *secureRedirectOptions) {
		o.whitelist = append(o.whitelist, urls...)
	}
}

// SecureRedirectUseSourceURL uses the given source URL (checks the `?return_to` value)
// instead of r.URL.
func SecureRedirectUseSourceURL(source string) SecureRedirectOption {
	return func(o *secureRedirectOptions) {
		o.sourceURL = source
	}
}

// SecureRedirectAllowSelfServiceURLs allows the caller to define `?return_to=` values
// which contain the server's URL and `/self-service` path prefix. Useful for redirecting
// to the login endpoint, for example.
func SecureRedirectAllowSelfServiceURLs(publicURL *url.URL) SecureRedirectOption {
	return func(o *secureRedirectOptions) {
		o.whitelist = append(o.whitelist, *urlx.AppendPaths(publicURL, "/self-service"))
	}
}

// SecureRedirectOverrideDefaultReturnTo overrides the defaultReturnTo address specified
// as the second arg.
func SecureRedirectOverrideDefaultReturnTo(defaultReturnTo *url.URL) SecureRedirectOption {
	return func(o *secureRedirectOptions) {
		o.defaultReturnTo = defaultReturnTo
	}
}

// SecureRedirectTo implements a HTTP redirector who mitigates open redirect vulnerabilities by
// working with whitelisting.
func SecureRedirectTo(r *http.Request, defaultReturnTo *url.URL, opts ...SecureRedirectOption) (returnTo *url.URL, err error) {
	o := &secureRedirectOptions{defaultReturnTo: defaultReturnTo}
	for _, opt := range opts {
		opt(o)
	}

	if len(o.whitelist) == 0 {
		return o.defaultReturnTo, nil
	}

	source := RequestURL(r)
	if o.sourceURL != "" {
		source, err = url.ParseRequestURI(o.sourceURL)
		if err != nil {
			return nil, herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to parse the original request URL: %s", err)
		}
	}

	if len(source.Query().Get("return_to")) == 0 {
		return o.defaultReturnTo, nil
	} else if returnTo, err = url.ParseRequestURI(source.Query().Get("return_to")); err != nil {
		return nil, herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to parse the return_to query parameter as an URL: %s", err)
	}

	returnTo.Host = stringsx.Coalesce(returnTo.Host, o.defaultReturnTo.Host)
	returnTo.Scheme = stringsx.Coalesce(returnTo.Scheme, o.defaultReturnTo.Scheme)

	var found bool
	for _, allowed := range o.whitelist {
		if strings.EqualFold(allowed.Scheme, returnTo.Scheme) &&
			strings.EqualFold(allowed.Host, returnTo.Host) &&
			strings.HasPrefix(
				stringsx.Coalesce(returnTo.Path, "/"),
				stringsx.Coalesce(allowed.Path, "/")) {
			found = true
		}
	}

	if !found {
		return nil, errors.WithStack(herodot.ErrBadRequest.
			WithReasonf("Requested return_to URL \"%s\" is not whitelisted.", returnTo).
			WithDebugf("Whitelisted domains are: %v", o.whitelist))
	}

	return returnTo, nil
}

func SecureContentNegotiationRedirection(
	w http.ResponseWriter, r *http.Request, out interface{},
	requestURL string, writer herodot.Writer, c configuration.Provider,
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
		ret, err := SecureRedirectTo(r, c.DefaultReturnToURL(),
			append([]SecureRedirectOption{
				SecureRedirectUseSourceURL(requestURL),
				SecureRedirectAllowURLs(c.WhitelistedReturnToDomains()),
				SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL()),
			}, opts...)...,
		)
		if err != nil {
			return err
		}

		http.Redirect(w, r, ret.String(), http.StatusFound)
	}

	return nil
}
