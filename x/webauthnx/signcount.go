// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthnx

import (
	"context"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/logrusx"
)

// SignCountDependencies is the minimal set of registry dependencies required to
// persist an updated WebAuthn signature counter after a login.
type SignCountDependencies interface {
	logrusx.Provider
	identity.PrivilegedPoolProvider
}

// UpdateSignCountAfterLogin writes the mutable authenticator state reported
// during a successful login back onto the stored credential: the signature
// counter, the clone warning, and the backup state.
//
// It implements the relying-party side of W3C WebAuthn signature-counter
// handling (https://www.w3.org/TR/webauthn-2/#sctn-sign-counter). When the
// go-webauthn library detects that the counter did not advance it sets
// CloneWarning, which we persist on the credential but deliberately do NOT act
// on to block the login:
//
//   - Synced passkeys (Apple iCloud, Google) legitimately always report a
//     counter of 0, so a hard "reject if not increasing" rule would lock out
//     real users.
//   - The user already proved possession of the authenticator, so a failure to
//     persist the counter must not turn a successful login into an error.
//
// The clone warning is durably stored on the credential and is inspectable via
// the admin identity API, so there is no separate security log event.
//
// The write goes through UpdateCredentialsConfig, which re-reads the stored
// config under an exclusive row lock and applies UpdateFromLogin to the latest
// committed value, so concurrent logins of the same credential serialize
// instead of clobbering each other's counter. That lock lives in its own
// transaction and is deliberately NOT enrolled in the login request: the
// assertion has already succeeded, so a failed or slow counter write must never
// roll back or block the login. Persistence is therefore best-effort — errors
// are logged, not returned.
func UpdateSignCountAfterLogin(
	ctx context.Context,
	d SignCountDependencies,
	i *identity.Identity,
	credentialsType identity.CredentialsType,
	validated *webauthn.Credential,
) {
	err := d.PrivilegedIdentityPool().UpdateCredentialsConfig(ctx, i.ID, credentialsType,
		identity.UpdateConfig(func(conf *identity.CredentialsWebAuthnConfig) error {
			conf.Credentials.UpdateFromLogin(validated)
			return nil
		}),
	)
	if err != nil {
		d.Logger().WithError(err).WithField("identity_id", i.ID).WithField("network_id", i.NID).
			Warn("Unable to persist updated WebAuthn signature counter. The login succeeded but clone detection state was not persisted.")
	}
}
