// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"context"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
)

func (s *Strategy) validateCredentials(ctx context.Context, i *identity.Identity) error {
	if err := s.d.IdentityValidator().Validate(ctx, i); err != nil {
		return err
	}

	c := i.GetCredentialsOr(identity.CredentialsTypeWebAuthn, &identity.Credentials{})
	if len(c.Identifiers) == 0 {
		return schema.NewMissingIdentifierError()
	}

	return nil
}
