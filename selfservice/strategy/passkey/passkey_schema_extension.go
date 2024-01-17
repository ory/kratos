// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/x/stringsx"
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

	return stringsx.Coalesce(e.PasskeyDisplayName, e.WebauthnIdentifier)
}

func (s *Strategy) PasskeyDisplayNameFromTraits(ctx context.Context, traits identity.Traits) string {
	id := identity.NewIdentity("")
	id.Traits = traits

	return s.PasskeyDisplayNameFromIdentity(ctx, id)
}
