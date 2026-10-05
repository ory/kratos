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
		// The prefix may arrive percent-encoded (e.g. "/my%20app"). url.Path is
		// the decoded path and url.RawPath its encoding, so we must prepend the
		// *decoded* prefix to Path and a *valid encoded* prefix to RawPath, and
		// keep the two consistent — otherwise String() would double-encode the
		// prefix or drop an existing %2F from the request path.
		encodedPrefix := "/" + strings.Trim(prefix, "/")
		if encodedPrefix != "/" {
			// Decode for Path; fall back to the raw value if it is not valid
			// percent-encoding.
			decodedPrefix := encodedPrefix
			if d, err := url.PathUnescape(encodedPrefix); err == nil {
				decodedPrefix = d
			}

			// Determine the request's current encoded path (RawPath is only set
			// when it differs from the default encoding of Path).
			rawPath := source.RawPath
			if rawPath == "" {
				rawPath = source.EscapedPath()
			}

			source.Path = decodedPrefix + source.Path
			combinedRaw := encodedPrefix + rawPath
			// Only set RawPath if it validly encodes the new Path; otherwise
			// leave it empty and let url.String() derive the encoding from Path.
			if p, err := url.PathUnescape(combinedRaw); err == nil && p == source.Path {
				source.RawPath = combinedRaw
			} else {
				source.RawPath = ""
			}
		}
	}

	return &source
}
