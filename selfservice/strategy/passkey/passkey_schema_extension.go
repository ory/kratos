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
)

type SchemaExtension struct {
	Identifier string
	sync.Mutex
}

func (e *SchemaExtension) Run(_ jsonschema.ValidationContext, s schema.ExtensionConfig, value any) error {
	e.Lock()
	defer e.Unlock()

	if s.Credentials.Passkey.Identifier {
		e.Identifier = strings.ToLower(fmt.Sprintf("%s", value))
	}

	return nil
}

func (e *SchemaExtension) Finish() error { return nil }

// PasskeyIdentifierFromIdentity returns the passkey identifier from the
// identity. It is usually the email address and used to name the passkey in the
// browser.
func (s *Strategy) PasskeyIdentifierFromIdentity(ctx context.Context, id *identity.Identity) string {
	e := new(SchemaExtension)
	_ = s.d.IdentityValidator().ValidateWithRunner(ctx, id, e)

	return e.Identifier
}

func (s *Strategy) PasskeyIdentifierFromTraits(ctx context.Context, traits identity.Traits) string {
	id := identity.NewIdentity("")
	id.Traits = traits

	return s.PasskeyIdentifierFromIdentity(ctx, id)
}
