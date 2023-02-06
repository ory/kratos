// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/x"
)

func (h *Handler) importCredentials(ctx context.Context, i *Identity, creds *IdentityWithCredentials) error {
	if creds == nil {
		return nil
	}

	if creds.Password != nil {
		if err := h.importPasswordCredentials(ctx, i, creds.Password); err != nil {
			return err
		}
	}

	if creds.OIDC != nil {
		if err := h.importOIDCCredentials(ctx, i, creds.OIDC); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) importPasswordCredentials(ctx context.Context, i *Identity, creds *AdminIdentityImportCredentialsPassword) (err error) {
	// In here we deliberately ignore any password policies as the point here is to import passwords, even if they
	// are not matching the policy, as the user needs to able to sign in with their old password.
	hashed := []byte(creds.Config.HashedPassword)
	if len(creds.Config.Password) > 0 {
		// Importing a clear text password
		hashed, err = h.r.Hasher(ctx).Generate(ctx, []byte(creds.Config.Password))
		if err != nil {
			return err
		}

		creds.Config.HashedPassword = string(hashed)
	}

	if !(hash.IsValidHashFormat(hashed)) {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("The imported password does not match any known hash format. For more information see https://www.ory.sh/dr/2"))
	}

	return i.SetCredentialsWithConfig(CredentialsTypePassword, Credentials{}, CredentialsPassword{HashedPassword: string(hashed)})
}

func (h *Handler) importOIDCCredentials(_ context.Context, i *Identity, creds *AdminIdentityImportCredentialsOIDC) error {
	var target CredentialsOIDC
	c, ok := i.GetCredentials(CredentialsTypeOIDC)
	if !ok {
		var providers []CredentialsOIDCProvider
		var ids []string
		for _, p := range creds.Config.Providers {
			ids = append(ids, OIDCUniqueID(p.Provider, p.Subject))
			providers = append(providers, CredentialsOIDCProvider{
				Subject:  p.Subject,
				Provider: p.Provider,
			})
		}

		return i.SetCredentialsWithConfig(
			CredentialsTypeOIDC,
			Credentials{Identifiers: ids},
			CredentialsOIDC{Providers: providers},
		)
	}

	if err := json.Unmarshal(c.Config, &target); err != nil {
		return errors.WithStack(x.PseudoPanic.WithWrap(err))
	}

	for _, p := range creds.Config.Providers {
		c.Identifiers = append(c.Identifiers, OIDCUniqueID(p.Provider, p.Subject))
		target.Providers = append(target.Providers, CredentialsOIDCProvider{
			Subject:  p.Subject,
			Provider: p.Provider,
		})
	}
	return i.SetCredentialsWithConfig(CredentialsTypeOIDC, *c, &target)
}
