// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/x/jsonschemax"
)

type SchemaExtension struct {
	WebauthnIdentifier string
	PasskeyDisplayName string
	sync.Mutex
}

func (e *SchemaExtension) Run(_ jsonschema.ValidationContext, s schema.ExtensionConfig, value any) error {
	e.Lock()
	defer e.Unlock()

	if s.Credentials.WebAuthn.Identifier {
		e.WebauthnIdentifier = strings.ToLower(fmt.Sprintf("%s", value))
	}

	if s.Credentials.Passkey.DisplayName {
		e.PasskeyDisplayName = fmt.Sprintf("%s", value)
	}

	return nil
}

func (e *SchemaExtension) Finish() error { return nil }

// PasskeyDisplayNameFromIdentity returns the passkey display name from the
// identity. It is usually the email address and used to name the passkey in the
// browser.
func (s *Strategy) PasskeyDisplayNameFromIdentity(ctx context.Context, id *identity.Identity) string {
	e := new(SchemaExtension)
	// We can ignore teh error here because proper validation happens once the identity is persisted.
	_ = s.d.IdentityValidator().ValidateWithRunner(ctx, id, e)

	return cmp.Or(e.PasskeyDisplayName, e.WebauthnIdentifier)
}

func (s *Strategy) PasskeyDisplayNameFromTraits(ctx context.Context, traits identity.Traits) string {
	id := identity.NewIdentity("")
	id.Traits = traits

	return s.PasskeyDisplayNameFromIdentity(ctx, id)
}

// PasskeyDisplayNameFromSchema returns every trait path (e.g. ["traits.email", "traits.phone"])
// whose schema flags `passkey.display_name: true` or `webauthn.identifier: true`. The slice
// is sorted alphabetically so server and client agree on precedence. When no
// trait is flagged, it preserves the legacy behavior: fall back to the first
// untitled trait, and return an error only when there is none.
func (s *Strategy) PasskeyDisplayNameFromSchema(ctx context.Context, schemaURL string) ([]string, error) {
	runner, err := schema.NewExtensionRunner(ctx)
	if err != nil {
		return nil, err
	}
	c, err := schema.NewCompilerWithURL(ctx, schemaURL, s.d.Config().SecurityDisallowRefInIdentitySchemas(ctx))
	if err != nil {
		return nil, err
	}
	c.ExtractAnnotations = true
	runner.Register(c)

	paths, err := jsonschemax.ListPaths(ctx, schemaURL, c)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, p := range paths {
		ext, ok := p.CustomProperties[schema.ExtensionName].(*schema.ExtensionConfig)
		if !ok {
			continue
		}
		if ext.Credentials.WebAuthn.Identifier || ext.Credentials.Passkey.DisplayName {
			result = append(result, p.Name)
		}
	}
	if len(result) > 0 {
		slices.Sort(result)
		return result, nil
	}

	for _, p := range paths {
		if strings.HasPrefix(p.Name, "traits.") && p.Title == "" {
			return []string{p.Name}, nil
		}
	}
	return nil, errors.New("no identifier found")
}
