// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"context"
	"fmt"
	"io"
	"net/url"

	_ "github.com/ory/jsonschema/v3/base64loader"
	_ "github.com/ory/jsonschema/v3/fileloader"
	_ "github.com/ory/jsonschema/v3/httploader"

	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
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
	resource, err := jsonschema.LoadURL(ctx, schemaURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = resource.Close() }()

	c := NewCompiler(disallowRefs)
	if err := c.AddResource(schemaURL, resource); err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}
