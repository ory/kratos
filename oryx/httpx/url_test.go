// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package httpx

import (
	"net/http"
	"net/url"
	"testing"
)

func TestIncomingRequestURL(t *testing.T) {
	for _, tc := range []struct {
		name     string
		rawPath  string
		host     string
		headers  map[string]string
		expected string
	}{
		{
			name:     "plain request without forwarded headers defaults scheme by TLS",
			rawPath:  "/self-service/login",
			host:     "example.com",
			expected: "http://example.com/self-service/login",
		},
		{
			name:    "honours X-Forwarded-Host and X-Forwarded-Proto",
			rawPath: "/self-service/login",
			host:    "internal:4433",
			headers: map[string]string{
				"X-Forwarded-Host":  "auth.example.com",
				"X-Forwarded-Proto": "https",
			},
			expected: "https://auth.example.com/self-service/login",
		},
		{
			name:    "re-adds stripped path prefix from X-Forwarded-Prefix",
			rawPath: "/self-service/login",
			host:    "auth.example.com",
			headers: map[string]string{
				"X-Forwarded-Proto":  "https",
				"X-Forwarded-Prefix": "/auth",
			},
			expected: "https://auth.example.com/auth/self-service/login",
		},
		{
			name:    "normalises a prefix with surrounding slashes",
			rawPath: "/self-service/login",
			host:    "auth.example.com",
			headers: map[string]string{
				"X-Forwarded-Proto":  "https",
				"X-Forwarded-Prefix": "auth/",
			},
			expected: "https://auth.example.com/auth/self-service/login",
		},
		{
			name:    "nested prefix",
			rawPath: "/self-service/login",
			host:    "auth.example.com",
			headers: map[string]string{
				"X-Forwarded-Proto":  "https",
				"X-Forwarded-Prefix": "/ory/auth",
			},
			expected: "https://auth.example.com/ory/auth/self-service/login",
		},
		{
			name:    "empty prefix is a no-op",
			rawPath: "/self-service/login",
			host:    "auth.example.com",
			headers: map[string]string{
				"X-Forwarded-Proto":  "https",
				"X-Forwarded-Prefix": "",
			},
			expected: "https://auth.example.com/self-service/login",
		},
		{
			name:    "encoded prefix: Path decoded, URL not double-encoded",
			rawPath: "/self-service/login",
			host:    "auth.example.com",
			headers: map[string]string{
				"X-Forwarded-Proto":  "https",
				"X-Forwarded-Prefix": "/my%20app",
			},
			expected: "https://auth.example.com/my%20app/self-service/login",
		},
		{
			name:    "root prefix is a no-op",
			rawPath: "/self-service/login",
			host:    "auth.example.com",
			headers: map[string]string{
				"X-Forwarded-Proto":  "https",
				"X-Forwarded-Prefix": "/",
			},
			expected: "https://auth.example.com/self-service/login",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &http.Request{
				URL:    &url.URL{Path: tc.rawPath},
				Host:   tc.host,
				Header: http.Header{},
			}
			for k, v := range tc.headers {
				r.Header.Set(k, v)
			}

			got := IncomingRequestURL(r).String()
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
