// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ory/herodot"
	"github.com/ory/kratos/hash"
	"github.com/pkg/errors"
	"slices"
)

func (h *Handler) importCredentials(ctx context.Context, i *Identity, creds *IdentityWithCredentials) error {
	if creds == nil {
		return nil
	}

	// We validate the identity to pre-hydrate identity credential identifiers.
	if err := h.r.IdentityManager().ValidateIdentity(ctx, i, new(ManagerOptions)); err != nil {
		return err
	}

	// This method only supports password and OIDC import at the moment.
	// If other methods are added please ensure that the available AAL is set correctly in the identity.
	//
	// It would actually be good if we would validate the identity post-creation to see if the credentials are working.
	if creds.Password != nil {
		// This method is somewhat hacky, because it does not set the credential's identifier. It relies on the
		// identity validation to set the identifier, which is called after this method.
		//
		// It would be good to make this explicit.
		if err := h.ImportPasswordCredentials(ctx, i, creds.Password); err != nil {
			return err
		}
	}

	if creds.OIDC != nil {
		if err := h.importOIDCCredentials(ctx, i, creds.OIDC); err != nil {
			return err
		}
	}

	if creds.SAML != nil {
		if err := h.importSAMLCredentials(ctx, i, creds.SAML); err != nil {
			return err
		}
	}

	if creds.WebAuthn != nil {
		if err := h.importWebAuthnCredentials(ctx, i, creds.WebAuthn.Config.Credentials, creds.WebAuthn.Config.UserHandle); err != nil {
			return err
		}
	}

	if creds.Passkey != nil {
		if err := h.importPasskeyCredentials(ctx, i, creds.WebAuthn.Config.Credentials, creds.WebAuthn.Config.UserHandle); err != nil {
			return err
		}
	}

	if creds.LookupSecret != nil {
		if err := h.importLookupSecretCredentials(ctx, i, creds.LookupSecret); err != nil {
			return err
		}
	}

	if creds.TOTP != nil {
		if err := h.importTOTPCredentials(ctx, i, creds.TOTP); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) importWebAuthnCredentials(_ context.Context, i *Identity, newCredentials CredentialsWebAuthn, userHandle []byte) error {
	resultCredentials, previousUserHandle, previousIdentifiers, err := mergeWebAuthnCredentials(i, CredentialsTypeWebAuthn, newCredentials)
	if err != nil {
		return err
	}

	if len(userHandle) == 0 {
		userHandle = previousUserHandle
	}

	return i.SetCredentialsWithConfig(
		CredentialsTypeWebAuthn,
		Credentials{
			Identifiers: previousIdentifiers,
		},
		CredentialsWebAuthnConfig{
			Credentials: resultCredentials,
			UserHandle:  userHandle,
		},
	)
}

// mergeWebAuthnCredentials merges existing WebAuthn credentials with new ones.
// It updates credentials with the same ID and appends new ones.
// Returns the merged credentials, previous user handle, previous identifiers, and any error.
func mergeWebAuthnCredentials(
	i *Identity,
	credentialType CredentialsType,
	newCredentials CredentialsWebAuthn,
) (CredentialsWebAuthn, []byte, []string, error) {
	var previousUserHandle []byte
	var previousIdentifiers []string

	resultCredentials := newCredentials

	if c, ok := i.GetCredentials(credentialType); ok {
		var original CredentialsWebAuthnConfig
		if err := json.Unmarshal(c.Config, &original); err != nil {
			return nil, nil, nil, errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReason(err.Error()))
		}

		resultCredentials = make(CredentialsWebAuthn, len(original.Credentials))
		copy(resultCredentials, original.Credentials)

		for _, updated := range newCredentials {
			var found bool
			for k, originalCredential := range original.Credentials {
				if bytes.Equal(originalCredential.ID, updated.ID) {
					resultCredentials[k] = updated
					found = true
					break
				}
			}
			if !found {
				resultCredentials = append(resultCredentials, updated)
			}
		}

		previousUserHandle = original.UserHandle
		previousIdentifiers = c.Identifiers
	}

	return resultCredentials, previousUserHandle, previousIdentifiers, nil
}

func (h *Handler) importPasskeyCredentials(_ context.Context, i *Identity, newCredentials CredentialsWebAuthn, userHandle []byte) error {
	resultCredentials, previousUserHandle, previousIdentifiers, err := mergeWebAuthnCredentials(i, CredentialsTypePasskey, newCredentials)
	if err != nil {
		return err
	}

	identifiers := previousIdentifiers
	if len(userHandle) == 0 {
		userHandle = previousUserHandle
	} else if !slices.Contains(identifiers, string(userHandle)) {
		identifiers = append(identifiers, string(userHandle))
	}

	return i.SetCredentialsWithConfig(
		CredentialsTypePasskey,
		Credentials{
			Identifiers: identifiers,
		},
		CredentialsWebAuthnConfig{
			Credentials: resultCredentials,
			UserHandle:  userHandle,
		},
	)
}

func (h *Handler) importLookupSecretCredentials(_ context.Context, i *Identity, creds *AdminIdentityImportCredentialsLookupSecret) error {
	var codes []RecoveryCode
	if c, ok := i.GetCredentials(CredentialsTypeLookup); ok {
		var target CredentialsLookupConfig
		if err := json.Unmarshal(c.Config, &target); err != nil {
			return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReason(err.Error()))
		}
		codes = target.RecoveryCodes
	}
	return i.SetCredentialsWithConfig(
		CredentialsTypeLookup,
		Credentials{},
		CredentialsLookupConfig{
			RecoveryCodes: append(codes, creds.Config.Codes...),
		},
	)
}

func (h *Handler) importTOTPCredentials(_ context.Context, i *Identity, creds *AdminIdentityImportCredentialsTOTP) error {
	return i.SetCredentialsWithConfig(CredentialsTypeTOTP, Credentials{}, CredentialsTOTPConfig{TOTPURL: creds.Config.TOTPURL})
}

func (h *Handler) ImportPasswordCredentials(ctx context.Context, i *Identity, creds *AdminIdentityImportCredentialsPassword) (err error) {
	if creds.Config.UsePasswordMigrationHook {
		return i.SetCredentialsWithConfig(CredentialsTypePassword, Credentials{}, CredentialsPassword{UsePasswordMigrationHook: true})
	}

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
			provider := CredentialsOIDCProvider{
				Subject:     p.Subject,
				Provider:    p.Provider,
				UseAutoLink: p.UseAutoLink,
			}
			if p.Organization.Valid {
				provider.Organization = p.Organization.UUID.String()
			}
			providers = append(providers, provider)
		}

		return i.SetCredentialsWithConfig(
			CredentialsTypeOIDC,
			Credentials{Identifiers: ids},
			CredentialsOIDC{Providers: providers},
		)
	}

	if err := json.Unmarshal(c.Config, &target); err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReason(err.Error()))
	}

	for _, p := range creds.Config.Providers {
		c.Identifiers = append(c.Identifiers, OIDCUniqueID(p.Provider, p.Subject))
		provider := CredentialsOIDCProvider{
			Subject:     p.Subject,
			Provider:    p.Provider,
			UseAutoLink: p.UseAutoLink,
		}
		if p.Organization.Valid {
			provider.Organization = p.Organization.UUID.String()
		}
		target.Providers = append(target.Providers, provider)
	}
	return i.SetCredentialsWithConfig(CredentialsTypeOIDC, *c, &target)
}

func (h *Handler) importSAMLCredentials(_ context.Context, i *Identity, creds *AdminIdentityImportCredentialsSAML) error {
	var target CredentialsOIDC
	c, ok := i.GetCredentials(CredentialsTypeSAML)
	if !ok {
		var providers []CredentialsOIDCProvider
		var ids []string
		for _, p := range creds.Config.Providers {
			ids = append(ids, OIDCUniqueID(p.Provider, p.Subject))
			provider := CredentialsOIDCProvider{
				Subject:  p.Subject,
				Provider: p.Provider,
			}
			if p.Organization.Valid {
				provider.Organization = p.Organization.UUID.String()
			}
			providers = append(providers, provider)
		}

		return i.SetCredentialsWithConfig(
			CredentialsTypeSAML,
			Credentials{Identifiers: ids},
			CredentialsOIDC{Providers: providers},
		)
	}

	if err := json.Unmarshal(c.Config, &target); err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithWrap(err).WithReason(err.Error()))
	}

	for _, p := range creds.Config.Providers {
		c.Identifiers = append(c.Identifiers, OIDCUniqueID(p.Provider, p.Subject))
		provider := CredentialsOIDCProvider{
			Subject:  p.Subject,
			Provider: p.Provider,
		}
		if p.Organization.Valid {
			provider.Organization = p.Organization.UUID.String()
		}
		target.Providers = append(target.Providers, provider)
	}
	return i.SetCredentialsWithConfig(CredentialsTypeSAML, *c, &target)
}
