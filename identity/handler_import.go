package identity

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/x"
)

func (h *Handler) importCredentials(ctx context.Context, i *Identity, creds *AdminIdentityImportCredentials) error {
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

	if creds.TOTP != nil {
		if err := h.importTOTPCredentials(ctx, i, creds.TOTP); err != nil {
			return err
		}
	}

	if creds.Lookup != nil {
		if err := h.importLookupCredentials(ctx, i, creds.Lookup); err != nil {
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

	if !(hash.IsArgon2idHash(hashed) || hash.IsArgon2iHash(hashed) || hash.IsBcryptHash(hashed) || hash.IsPbkdf2Hash(hashed)) {
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

func (h *Handler) importTOTPCredentials(_ context.Context, i *Identity, creds *AdminIdentityImportCredentialsTOTP) error {
	return i.SetCredentialsWithConfig(CredentialsTypeTOTP, Credentials{Identifiers: []string{i.ID.String()}}, CredentialsTOTP{TOTPURL: creds.Config.TOTPURL})
}

func (h *Handler) importLookupCredentials(_ context.Context, i *Identity, creds *AdminIdentityImportCredentialsLookup) error {
	var target CredentialsLookupSecrets
	c, ok := i.GetCredentials(CredentialsTypeLookup)
	if !ok {
		var lookupSecrets []CredentialsLookupSecret
		for _, s := range creds.Config.LookupSecrets {
			lookupSecrets = append(lookupSecrets, CredentialsLookupSecret(s))
		}

		return i.SetCredentialsWithConfig(
			CredentialsTypeLookup,
			Credentials{Identifiers: []string{i.ID.String()}},
			CredentialsLookupSecrets{LookupSecrets: lookupSecrets},
		)
	}

	if err := json.Unmarshal(c.Config, &target); err != nil {
		return errors.WithStack(x.PseudoPanic.WithWrap(err))
	}

	c.Identifiers = []string{i.ID.String()}
	for _, s := range creds.Config.LookupSecrets {
		target.LookupSecrets = append(target.LookupSecrets, CredentialsLookupSecret(s))
	}

	return i.SetCredentialsWithConfig(CredentialsTypeLookup, *c, &target)
}
