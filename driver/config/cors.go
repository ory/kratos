// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/cors"
)

func allowOrigin(allowedOrigins []string, origin string) bool {
	if len(allowedOrigins) == 0 {
		return true
	}
	for _, o := range allowedOrigins {
		if o == "*" {
			// allow all origins
			return true
		}
		prefix, suffix, found := strings.Cut(o, "*")
		if !found {
			// not a pattern, check for equality
			if o == origin {
				return true
			}
			continue
		}
		// inspired by https://github.com/rs/cors/blob/066574eebbd0f5f1b6cd1154a160cc292ac1835e/utils.go#L15
		if len(origin) >= len(prefix)+len(suffix) && strings.HasPrefix(origin, prefix) && strings.HasSuffix(origin, suffix) {
			return true
		}
	}
	return false
}

func (p *Config) cors(ctx context.Context, keyPrefix string) (cors.Options, bool) {
	opts, enabled := p.GetProvider(ctx).CORS(keyPrefix, cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Cookie"},
		ExposedHeaders:   []string{"Content-Type", "Set-Cookie"},
		AllowCredentials: true,
	})
	opts.AllowOriginRequestFunc = func(r *http.Request, origin string) bool {
		// load the origins from the config on every request to allow hot-reloading
		allowedOrigins := p.GetProvider(r.Context()).Strings(keyPrefix + ".cors.allowed_origins")
		return allowOrigin(allowedOrigins, origin)
	}

	return opts, enabled
}
