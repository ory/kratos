// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"cmp"
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3/httploader"
	"github.com/ory/x/httpx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidator(t *testing.T) {
	router := http.NewServeMux()
	fs := http.StripPrefix("/schema", http.FileServer(http.Dir("stub/validator")))
	router.HandleFunc("/schema/{name}", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	ctx := context.WithValue(ctx, httploader.ContextKey, httpx.NewResilientClient())
	for k, tc := range []struct {
		i   json.RawMessage
		err string
		u   string
	}{
		{
			i: json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
		},
		{
			i:   json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": -1 }`),
			err: "I[#/age] S[#/properties/age/minimum] must be >= 1 but found -1",
		},
		{
			i:   json.RawMessage(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
			err: `I[#] S[#/additionalProperties] additionalProperties "whatever" not allowed`,
		},
		{
			u: ts.URL + "/schema/whatever.schema.json",
			i: json.RawMessage(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
		},
		{
			u:   ts.URL + "/schema/whatever.schema.json",
			i:   json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			err: `I[#] S[#/additionalProperties] additionalProperties "firstName" not allowed`,
		},
		{
			u:   ts.URL,
			i:   json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			err: "Invalid configuration",
		},
		{
			u:   "not-a-url",
			i:   json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			err: "Invalid configuration",
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := NewValidator().Validate(ctx, cmp.Or(tc.u, ts.URL+"/schema/firstName.schema.json"), tc.i)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}
		})
	}
}

// TestSchemaValidator_FileRefExfiltration asserts that identity schemas
// cannot exfiltrate server-side files via `$ref: "file://..."`. The jsonschema
// `file` loader is registered process-wide so operator-configured top-level
// `file://` schemas still work, but per-compiler $ref resolution must reject
// the `file` scheme.
func TestSchemaValidator_FileRefExfiltration(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	secretContents := `"LEAKED_COOKIE_SECRET_VALUE"`

	schemaWithRef := func(refURL string) string {
		return fmt.Sprintf(`{
			"$id": "https://test.example.com/probe.schema.json",
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"properties": {
				"traits": {
					"type": "object",
					"properties": {
						"email": {
							"type": "string",
							"format": "email",
							"ory.sh/kratos": {
								"credentials": {
									"password": {"identifier": true}
								}
							}
						},
						"field": {"$ref": %q}
					},
					"required": ["email"]
				}
			}
		}`, refURL)
	}

	serveSchema := func(t *testing.T, body string) string {
		mux := http.NewServeMux()
		mux.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		})
		ts := httptest.NewServer(mux)
		t.Cleanup(ts.Close)
		return ts.URL + "/schema"
	}

	ctx := context.WithValue(context.Background(), httploader.ContextKey, httpx.NewResilientClient())

	t.Run("case=file ref is rejected before any file access", func(t *testing.T) {
		t.Parallel()
		// Fragment that would pass validation if the file were read, and fail
		// if not. Its contents are a valid JSON-schema fragment.
		schemaFragment := fmt.Sprintf(`{"type": "string", "const": %s}`, secretContents)
		fragmentPath := filepath.Join(dir, "fragment.json")
		require.NoError(t, os.WriteFile(fragmentPath, []byte(schemaFragment), 0o600))

		u := serveSchema(t, schemaWithRef("file://"+fragmentPath))

		t.Run("with disallowRefs=true", func(t *testing.T) {
			err := NewValidator().Validate(ctx, u,
				json.RawMessage(`{"traits": {"email": "a@b.c", "field": "wrong"}}`),
				WithDisallowRefs(true))
			require.Error(t, err)

			// The file scheme must be rejected, and the file contents must not
			// leak into the error.
			var he *herodot.DefaultError
			require.True(t, stderrors.As(err, &he))
			assert.Contains(t, he.Debug(), `"file"`, "rejected scheme should appear in cause")
			assert.NotContains(t, he.Error(), "LEAKED_COOKIE_SECRET_VALUE")
			assert.NotContains(t, he.Debug(), "LEAKED_COOKIE_SECRET_VALUE")
		})

		t.Run("with disallowRefs=false preserves legacy exfiltration", func(t *testing.T) {
			// When the feature flag is off (default for existing deployments),
			// the legacy behavior must remain: the $ref is dereferenced and
			// the file contents flow into validation errors. This documents
			// the intentional opt-in nature of the mitigation.
			err := NewValidator().Validate(ctx, u,
				json.RawMessage(`{"traits": {"email": "a@b.c", "field": "wrong"}}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "LEAKED_COOKIE_SECRET_VALUE",
				"with the flag off, the file contents should still be reachable — this test documents the exploitable default")
		})
	})
}
