// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package httpx

import (
	"cmp"
	"net/http"
	"net/url"
	"strings"
)

// IncomingRequestURL returns the URL of the incoming HTTP request by looking at the host, TLS, and X-Forwarded-* headers.
func IncomingRequestURL(r *http.Request) *url.URL {
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

	// When deployed behind a reverse proxy that strips a path prefix (e.g.
	// Traefik's StripPrefix, or nginx), the proxy conventionally reports the
	// stripped prefix via the X-Forwarded-Prefix header. Without re-adding it,
	// the reconstructed URL is missing the prefix and any client-facing URL
	// derived from it (e.g. self-service "back" links) points to the wrong,
	// unreachable path. Prepend the forwarded prefix to restore the externally
	// visible path.
	if prefix := strings.TrimSpace(r.Header.Get("X-Forwarded-Prefix")); prefix != "" {
		// Normalise to a single leading slash and no trailing slash so joining
		// with the request path never produces "//" or a missing separator.
		prefix = "/" + strings.Trim(prefix, "/")
		if prefix != "/" {
			source.Path = prefix + source.Path
			if source.RawPath != "" {
				source.RawPath = prefix + source.RawPath
			}
		}
	}

	return &source
}
