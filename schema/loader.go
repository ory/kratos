// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	_ "github.com/ory/jsonschema/v3/base64loader"
	_ "github.com/ory/jsonschema/v3/fileloader"
	"github.com/ory/jsonschema/v3/httploader"

	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/httpx"
)

// loadRefURL resolves the URL of a `$ref` inside a schema. It enforces the
// scheme allowlist and then delegates to the jsonschema package's global
// loader table. The global `file` loader remains registered so that
// operator-configured top-level schema URLs (resolved outside the compiler)
// keep working.
func loadRefURL(ctx context.Context, raw string) (io.ReadCloser, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid $ref URL %q: %w", raw, err)
	}
	if u.Scheme != "base64" {
		return nil, fmt.Errorf("$ref scheme %q is not permitted in identity schemas", u.Scheme)
	}
	return jsonschema.LoadURL(ctx, raw)
}

// NewCompiler returns a jsonschema.Compiler. When disallowRefs is true, the
// compiler rejects `file://` URLs (and any other non-allowlisted scheme) in
// `$ref` values, preventing an attacker-supplied schema from reading local
// files on the Kratos host. When disallowRefs is false, the compiler uses
// the jsonschema library's default loader table, which preserves legacy
// behavior for operators who intentionally reference local files.
//
// The flag is controlled by `security.disallow_ref_in_identity_schemas`.
// Ory Network forces it on.
func NewCompiler(disallowRefs bool) *jsonschema.Compiler {
	c := jsonschema.NewCompiler()
	if disallowRefs {
		c.LoadURL = loadRefURL
	}
	return c
}

// NewCompilerWithURL returns a NewCompiler with the top-level schema at
// schemaURL already loaded and registered via the trusted global loader
// (which always supports `file://` for operator-configured top-level URLs,
// regardless of disallowRefs). See NewCompiler for the semantics of
// disallowRefs.
func NewCompilerWithURL(ctx context.Context, schemaURL string, disallowRefs bool) (*jsonschema.Compiler, error) {
	ctx = ensureGuardedHTTPClient(ctx)

	resource, err := jsonschema.LoadURL(ctx, schemaURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = resource.Close() }()

	// Read the body with a hard size cap so a malicious schema URL cannot
	// OOM kratos by returning a multi-GB body.
	data, err := io.ReadAll(io.LimitReader(resource, MaxSchemaBodyBytes+1))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(data) > MaxSchemaBodyBytes {
		return nil, errors.Errorf("identity schema rejected: body exceeds %d bytes", MaxSchemaBodyBytes)
	}

	// Decode once for structural prevalidation. The upstream compiler will
	// decode again from the same bytes — accepting a small CPU duplication
	// in exchange for the security gate.
	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := preValidateSchema(doc); err != nil {
		return nil, errors.WithStack(err)
	}

	c := NewCompiler(disallowRefs)
	if err := c.AddResource(schemaURL, bytes.NewReader(data)); err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// ensureGuardedHTTPClient guarantees that the jsonschema httploader will
// pick up a client whose dialer rejects internal IP ranges. Callers who
// have already attached a client (typically via the request middleware
// x.HTTPLoaderContextMiddleware) are unaffected — the existing client is
// preserved. Callers without a client get a fresh resilient client with
// `ResilientClientDisallowInternalIPs` applied so that fetches from
// background paths cannot be steered at IMDS or in-cluster services.
func ensureGuardedHTTPClient(ctx context.Context) context.Context {
	if ctx.Value(httploader.ContextKey) != nil {
		return ctx
	}
	return context.WithValue(ctx, httploader.ContextKey,
		httpx.NewResilientClient(httpx.ResilientClientDisallowInternalIPs()))
}
