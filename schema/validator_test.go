// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"bytes"
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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

// TestSchemaValidator_BoundedNestingDepth asserts that an identity schema with
// pathological `properties.x.properties.x...` nesting is rejected with a
// bounded error rather than driving the upstream compiler into deep
// recursion.
//
// Background: github.com/ory/jsonschema/v3 compiler.go:218 (compileMap)
// recurses through every level of `properties`, `allOf`, `anyOf`, `oneOf`,
// `items`, etc., with no depth cap. A customer-uploaded schema therefore
// allocates one *Schema node per level. Go stdlib's encoding/json caps
// nesting at 10000 (Go 1.21+), which bounds the absolute worst case but
// still permits ~5000 levels of kratos compile-time recursion before that
// limit kicks in — far above any realistic identity schema (≤10 levels).
//
// Currently this test fails: Validate compiles a 1024-deep schema silently
// and returns nil. After a kratos-side pre-parse gate in
// kratos-oss/schema/ is added, this test must pass: schemas exceeding a
// sensible nesting limit (suggest 32) are rejected before compile runs.
func TestSchemaValidator_BoundedNestingDepth(t *testing.T) {
	t.Parallel()

	// Depth chosen well above any realistic identity schema (5-10 levels in
	// practice) and well above the suggested kratos-side cap (32), but well
	// below Go stdlib's JSON nesting cap (10000) so we exercise the kratos
	// path, not the stdlib safety net.
	const depth = 1024

	var sb strings.Builder
	sb.WriteString(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"traits":`)
	for range depth {
		sb.WriteString(`{"type":"object","properties":{"x":`)
	}
	sb.WriteString(`{"type":"string"}`)
	for range depth {
		sb.WriteString(`}}`)
	}
	sb.WriteString(`}}`)

	schemaURL := "base64://" + base64.StdEncoding.EncodeToString([]byte(sb.String()))

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, httploader.ContextKey, httpx.NewResilientClient())

	// Document is intentionally trivial — we are testing the schema compile
	// path, not document validation. A well-formed empty traits object should
	// be accepted by a sane schema; a deep-but-otherwise-valid schema must be
	// rejected at the kratos boundary, not silently compiled.
	err := NewValidator().Validate(ctx, schemaURL,
		json.RawMessage(`{"traits":{}}`),
		WithDisallowRefs(true))

	require.Error(t, err,
		"identity schema with %d levels of properties.x nesting must be rejected; "+
			"the upstream jsonschema compiler has no depth cap and will recurse until OOM",
		depth)

	// After fix, the rejection should be attributable to the depth limit.
	// herodot wraps the cause behind a surface "Invalid configuration"
	// message — inspect the full chain via %+v.
	full := strings.ToLower(fmt.Sprintf("%+v", err))
	assert.True(t,
		strings.Contains(full, "depth") || strings.Contains(full, "nest"),
		"error should indicate the depth/nesting limit, got: %s", full)
}

// TestSchemaValidator_RejectsSelfReferentialRef asserts that an identity
// schema whose `$ref` resolves back to the same schema (e.g. `{"$ref":"#"}`)
// is rejected at compile time.
//
// Background: github.com/ory/jsonschema/v3 has compile-time cycle detection
// (compiler.go:139-202 caches the *Schema in r.schemas[ref] before recursing,
// so compilation terminates), but the resulting *Schema has its `Ref` field
// pointing at itself. At validate time, schema.go:155-169 unconditionally
// invokes s.Ref.validate(v) — for the self-pointer case this recurses
// indefinitely until the goroutine stack hits Go's 1GB limit and the runtime
// raises `fatal error: stack overflow`. That fatal is **not recoverable**
// via recover(); it terminates the entire process. A customer who can
// upload an identity schema with a self-pointer can therefore crash any
// kratos pod that compiles and validates against it.
//
// The compile path is the safe place to test this. Calling
// NewValidator().Validate would crash the test binary — see the comment in
// the body. The only legitimate fix is a kratos-side pre-parse gate that
// walks the parsed schema tree, resolves each `$ref`, and rejects refs whose
// target is the same schema (or any schema ancestor that produces an
// equivalent self-cycle).
//
// Currently this test fails: the upstream compiler accepts {"$ref":"#"} and
// returns nil error. After adding the kratos-side gate, this test must pass
// with a 4xx-class error before any *Schema with a self-pointer is built.
func TestSchemaValidator_RejectsSelfReferentialRef(t *testing.T) {
	t.Parallel()

	const schemaJSON = `{"$schema":"http://json-schema.org/draft-07/schema#","$ref":"#"}`
	schemaURL := "base64://" + base64.StdEncoding.EncodeToString([]byte(schemaJSON))

	ctx := context.WithValue(t.Context(), httploader.ContextKey, httpx.NewResilientClient())

	// The kratos-side prevalidate gate (schema/prevalidate.go) rejects the
	// schema at NewCompilerWithURL time, before the upstream compiler
	// produces a *Schema with Ref pointing at itself. That is the safest
	// place to fail: every caller of NewCompilerWithURL is protected, and
	// no path can construct the self-pointer that crashes Validate.
	_, err := NewCompilerWithURL(ctx, schemaURL, true)
	require.Error(t, err,
		"self-referential $ref ({\"$ref\":\"#\"}) must be rejected at the loader, "+
			"before producing a *Schema whose Ref field points at itself")
	assert.Contains(t, err.Error(), "self-referential $ref")

	// End-to-end sanity: Validate must also reject and must not crash the
	// process via fatal stack overflow. With the gate in place, Validate's
	// internal call to NewCompilerWithURL fails before reaching the validate
	// path that would otherwise infinite-recurse on s.Ref == s.
	err = NewValidator().Validate(ctx, schemaURL,
		json.RawMessage(`{}`),
		WithDisallowRefs(true))
	require.Error(t, err)
}

// TestSchemaValidator_PreValidate exercises the kratos-side pre-parse gate
// for identity schemas. Each subtest constructs a schema that hits one of
// the documented limits in prevalidate.go and asserts that NewValidator()
// rejects it before the upstream compiler runs.
func TestSchemaValidator_PreValidate(t *testing.T) {
	t.Parallel()

	loadCtx := func(t *testing.T) context.Context {
		t.Helper()
		return context.WithValue(t.Context(), httploader.ContextKey, httpx.NewResilientClient())
	}

	cases := []struct {
		name        string
		schemaJSON  string
		errFragment string
	}{
		{
			name: "deeply nested object",
			schemaJSON: func() string {
				var sb strings.Builder
				sb.WriteString(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"traits":`)
				const depth = MaxSchemaNestingDepth + 4
				for range depth {
					sb.WriteString(`{"type":"object","properties":{"x":`)
				}
				sb.WriteString(`{"type":"string"}`)
				for range depth {
					sb.WriteString(`}}`)
				}
				sb.WriteString(`}}`)
				return sb.String()
			}(),
			errFragment: "nesting depth",
		},
		{
			name: "object with too many properties",
			schemaJSON: func() string {
				var sb strings.Builder
				sb.WriteString(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{`)
				for i := range MaxSchemaObjectKeys + 5 {
					if i > 0 {
						sb.WriteString(",")
					}
					fmt.Fprintf(&sb, `"p%d":{"type":"string"}`, i)
				}
				sb.WriteString(`}}`)
				return sb.String()
			}(),
			errFragment: "object key count",
		},
		{
			name: "anyOf array arity",
			schemaJSON: func() string {
				var sb strings.Builder
				sb.WriteString(`{"$schema":"http://json-schema.org/draft-07/schema#","anyOf":[`)
				for i := range MaxSchemaArrayElements + 5 {
					if i > 0 {
						sb.WriteString(",")
					}
					sb.WriteString(`{"type":"string"}`)
				}
				sb.WriteString(`]}`)
				return sb.String()
			}(),
			errFragment: "array element count",
		},
		{
			name:        "self-referential $ref to root",
			schemaJSON:  `{"$schema":"http://json-schema.org/draft-07/schema#","$ref":"#"}`,
			errFragment: "self-referential $ref",
		},
		{
			name: "self-referential $ref to root, second attempt",
			schemaJSON: `{
  "properties": {
    "traits": {
      "properties": {
        "name": {
          "$ref": "#/properties/traits/properties/name"
        }
      }
    }
  },
  "$schema": "http://json-schema.org/draft-07/schema#"
}`,
			errFragment: "self-referential $ref",
		},
		{
			name: "indirect $ref cycle a → b → a",
			schemaJSON: `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$defs": {
    "a": {"$ref": "#/$defs/b"},
    "b": {"$ref": "#/$defs/a"}
  },
  "$ref": "#/$defs/a"
}`,
			errFragment: "self-referential $ref cycle",
		},
		{
			name:        "invalid regex in pattern",
			schemaJSON:  `{"$schema":"http://json-schema.org/draft-07/schema#","type":"string","pattern":"(*invalid"}`,
			errFragment: "invalid regex in pattern",
		},
		{
			name:        "invalid regex in patternProperties key",
			schemaJSON:  `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","patternProperties":{"(*invalid":{"type":"string"}}}`,
			errFragment: "invalid regex in patternProperties",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			schemaURL := "base64://" + base64.StdEncoding.EncodeToString([]byte(c.schemaJSON))
			err := NewValidator().Validate(loadCtx(t), schemaURL,
				json.RawMessage(`{"traits":{"name":"value"}}`),
				WithDisallowRefs(true))
			require.Error(t, err)
			assert.Contains(t, fmt.Sprintf("%+v", err), c.errFragment,
				"expected error to mention %q, got %v", c.errFragment, err)
		})
	}
}

// TestSchemaValidator_BodyTooLarge asserts that a schema URL whose response
// body exceeds MaxSchemaBodyBytes is rejected before any parse or compile
// runs. This guards against memory exhaustion when a customer-controlled
// URL serves a multi-megabyte (or larger) body. See F3.
func TestSchemaValidator_BodyTooLarge(t *testing.T) {
	t.Parallel()

	body := bytes.Repeat([]byte("X"), MaxSchemaBodyBytes+1024)
	mux := http.NewServeMux()
	mux.HandleFunc("/big.schema.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	// httptest binds 127.0.0.1 — that is a private IP for the SSRF guard.
	// Use a non-guarded resilient client so we exercise the size cap, not
	// the SSRF guard.
	ctx := context.WithValue(t.Context(), httploader.ContextKey, httpx.NewResilientClient())

	err := NewValidator().Validate(ctx, ts.URL+"/big.schema.json",
		json.RawMessage(`{}`),
		WithDisallowRefs(true))
	require.Error(t, err)
	assert.Contains(t, fmt.Sprintf("%+v", err), "body exceeds")
}

// TestLoadRefURL pins the actual scheme allowlist enforced by loadRefURL,
// resolving the F4 doc-comment drift. The earlier comment claimed
// http/https/base64 were permitted; the implementation has only ever
// accepted base64. Any future relaxation must update this test.
func TestLoadRefURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		raw     string
		wantSub string
	}{
		{"file scheme rejected", "file:///etc/passwd", `scheme "file"`},
		{"http scheme rejected", "http://example.com/schema.json", `scheme "http"`},
		{"https scheme rejected", "https://example.com/schema.json", `scheme "https"`},
		{"data scheme rejected", "data:application/json,{}", `scheme "data"`},
		{"empty scheme rejected", "/relative/path", `scheme ""`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := loadRefURL(t.Context(), c.raw)
			require.Error(t, err)
			assert.Contains(t, err.Error(), c.wantSub)
		})
	}
}

// TestEnsureGuardedHTTPClient_PrivateIPRejected verifies that a context
// without an attached HTTP client gets a default SSRF-guarded one when
// passed to NewCompilerWithURL. A schema URL pointing at httptest's
// loopback listener (127.0.0.1) must be rejected by the dialer before any
// HTTP request hits the test server. See F1.
func TestEnsureGuardedHTTPClient_PrivateIPRejected(t *testing.T) {
	t.Parallel()

	var hit atomic.Bool
	mux := http.NewServeMux()
	mux.HandleFunc("/schema.json", func(w http.ResponseWriter, r *http.Request) {
		hit.Store(true)
		_, _ = w.Write([]byte(`{"type":"object"}`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	// Bare context — no httploader.ContextKey attached. NewCompilerWithURL
	// must attach a guarded default client and refuse the loopback dial.
	_, err := NewCompilerWithURL(t.Context(), ts.URL+"/schema.json", true)
	require.Error(t, err, "fetch via SSRF-guarded default client must reject loopback addresses")
	assert.False(t, hit.Load(),
		"server handler was invoked — SSRF guard did not block the request")
}
