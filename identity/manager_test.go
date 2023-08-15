// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ory/x/pointerx"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

func TestManager(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/manager.schema.json")
	extensionSchemaID := testhelpers.UseIdentitySchema(t, conf, "file://./stub/extension.schema.json")
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationLoginHints, true)

	t.Run("case=should fail to create because validation fails", func(t *testing.T) {
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.Traits = identity.Traits(`{"email":"not an email"}`)
		require.Error(t, reg.IdentityManager().Create(context.Background(), i))
	})

	newTraits := func(email string, unprotected string) identity.Traits {
		return identity.Traits(fmt.Sprintf(`{"email":"%[1]s","email_verify":"%[1]s","email_recovery":"%[1]s","email_creds":"%[1]s","unprotected": "%[2]s"}`, email, unprotected))
	}

	checkExtensionFields := func(i *identity.Identity, expected string) func(*testing.T) {
		return func(t *testing.T) {
			require.Len(t, i.VerifiableAddresses, 1)
			assert.EqualValues(t, expected, i.VerifiableAddresses[0].Value)
			assert.EqualValues(t, identity.VerifiableAddressTypeEmail, i.VerifiableAddresses[0].Via)

			require.Len(t, i.RecoveryAddresses, 1)
			assert.EqualValues(t, expected, i.RecoveryAddresses[0].Value)
			assert.EqualValues(t, identity.VerifiableAddressTypeEmail, i.RecoveryAddresses[0].Via)

			require.NotNil(t, i.Credentials[identity.CredentialsTypePassword])
			assert.Equal(t, []string{expected}, i.Credentials[identity.CredentialsTypePassword].Identifiers)
		}
	}

	checkExtensionFieldsForIdentities := func(t *testing.T, expected string, original *identity.Identity) {
		fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
		require.NoError(t, err)
		identities := []identity.Identity{*original, *fromStore}
		for k := range identities {
			t.Run(fmt.Sprintf("identity=%d", k), checkExtensionFields(&identities[k], expected))
		}
	}

	t.Run("method=Create", func(t *testing.T) {
		t.Run("case=should create identity and track extension fields", func(t *testing.T) {
			email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(email, "")
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
			checkExtensionFieldsForIdentities(t, email, original)
			got, ok := original.AvailableAAL.ToAAL()
			require.True(t, ok)
			assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, got)
		})

		t.Run("case=correctly set AAL", func(t *testing.T) {
			t.Run("case=should set AAL to 0 if no credentials are available", func(t *testing.T) {
				email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
				original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				original.Traits = newTraits(email, "")
				require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
				got, ok := original.AvailableAAL.ToAAL()
				require.True(t, ok)
				assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, got)
			})

			t.Run("case=should set AAL to 1 if password is set", func(t *testing.T) {
				email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
				original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				original.Traits = newTraits(email, "")
				original.Credentials = map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypePassword: {
						Type:        identity.CredentialsTypePassword,
						Identifiers: []string{email},
						Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
					},
				}
				require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
				got, ok := original.AvailableAAL.ToAAL()
				require.True(t, ok)
				assert.Equal(t, identity.AuthenticatorAssuranceLevel1, got)
			})

			t.Run("case=should set AAL to 2 if password and TOTP is set", func(t *testing.T) {
				email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
				original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				original.Traits = newTraits(email, "")
				original.Credentials = map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypePassword: {
						Type:        identity.CredentialsTypePassword,
						Identifiers: []string{email},
						Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
					},
					identity.CredentialsTypeTOTP: {
						Type:        identity.CredentialsTypeTOTP,
						Identifiers: []string{email},
						Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
					},
				}
				require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
				got, ok := original.AvailableAAL.ToAAL()
				require.True(t, ok)
				assert.Equal(t, identity.AuthenticatorAssuranceLevel2, got)
			})

			t.Run("case=should set AAL to 0 if only TOTP is set", func(t *testing.T) {
				email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
				original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				original.Traits = newTraits(email, "")
				original.Credentials = map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypeTOTP: {
						Type:        identity.CredentialsTypeTOTP,
						Identifiers: []string{email},
						Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
					},
				}
				require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
				got, ok := original.AvailableAAL.ToAAL()
				require.True(t, ok)
				assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, got)
			})
		})

		t.Run("case=should expose validation errors with option", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"not an email"}`)
			err := reg.IdentityManager().Create(context.Background(), original, identity.ManagerExposeValidationErrorsForInternalTypeAssertion)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "\"not an email\" is not valid \"email\"")
		})

		t.Run("case=should not expose validation errors without option", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"not an email"}`)
			err := reg.IdentityManager().Create(context.Background(), original)
			require.Error(t, err)
			assert.NotContains(t, err.Error(), "\"not an email\" is not valid \"email\"")
		})

		t.Run("case=should correctly hint at the duplicate credential", func(t *testing.T) {
			createIdentity := func(email string, field string, creds map[identity.CredentialsType]identity.Credentials) *identity.Identity {
				i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				i.Traits = identity.Traits(fmt.Sprintf(`{"%s":"%s"}`, field, email))
				i.Credentials = creds
				return i
			}

			t.Run("case=credential identifier duplicate", func(t *testing.T) {
				t.Run("type=password", func(t *testing.T) {
					email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
					creds := map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {
							Type:        identity.CredentialsTypePassword,
							Identifiers: []string{email},
							Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
						},
					}

					first := createIdentity(email, "email_creds", creds)
					require.NoError(t, reg.IdentityManager().Create(context.Background(), first))

					second := createIdentity(email, "email_creds", creds)
					err := reg.IdentityManager().Create(context.Background(), second)
					require.Error(t, err)

					var verr = new(identity.ErrDuplicateCredentials)
					assert.ErrorAs(t, err, &verr)
					assert.EqualValues(t, []string{identity.CredentialsTypePassword.String()}, verr.AvailableCredentials())
					assert.Len(t, verr.AvailableOIDCProviders(), 0)
					assert.Equal(t, verr.IdentifierHint(), email)
				})

				t.Run("type=webauthn", func(t *testing.T) {
					email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
					creds := map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypeWebAuthn: {
							Type:        identity.CredentialsTypeWebAuthn,
							Identifiers: []string{email},
							Config:      sqlxx.JSONRawMessage(`{"credentials": [{"is_passwordless":true}]}`),
						},
					}

					first := createIdentity(email, "email_webauthn", creds)
					require.NoError(t, reg.IdentityManager().Create(context.Background(), first))

					second := createIdentity(email, "email_webauthn", nil)
					err := reg.IdentityManager().Create(context.Background(), second)
					require.Error(t, err)

					var verr = new(identity.ErrDuplicateCredentials)
					assert.ErrorAs(t, err, &verr)
					assert.EqualValues(t, []string{identity.CredentialsTypeWebAuthn.String()}, verr.AvailableCredentials())
					assert.Len(t, verr.AvailableOIDCProviders(), 0)
					assert.Equal(t, verr.IdentifierHint(), email)
				})
			})

			runAddress := func(t *testing.T, field string) {
				t.Run("case=password duplicate", func(t *testing.T) {
					// This test mimics a case where an existing user with email + password exists, and the
					// new user tries to sign up with a verification email (NOT email + password) that matches
					// this existing record. Here, the end result is that we want to show the
					// user: "Sign up with email foo@bar.com and your password instead."
					email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
					creds := map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {
							Type:        identity.CredentialsTypePassword,
							Identifiers: []string{email},
							Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
						},
					}

					first := createIdentity(email, field, creds)
					require.NoError(t, reg.IdentityManager().Create(context.Background(), first))

					second := createIdentity(email, field, nil)
					err := reg.IdentityManager().Create(context.Background(), second)
					require.Error(t, err)

					var verr = new(identity.ErrDuplicateCredentials)
					assert.ErrorAs(t, err, &verr)
					assert.EqualValues(t, []string{identity.CredentialsTypePassword.String()}, verr.AvailableCredentials())
					assert.Len(t, verr.AvailableOIDCProviders(), 0)
					assert.Equal(t, verr.IdentifierHint(), email)
				})

				t.Run("case=OIDC duplicate", func(t *testing.T) {
					// This test mimics a case where user signed up using Social Sign In exists, and the
					// new user tries to sign up with a verification email (NOT email + password) that matches
					// this existing record (for example by using another social sign in provider.
					// Here, the end result is that we want to show "Sign in using google instead".
					email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
					creds := map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypeOIDC: {
							Type: identity.CredentialsTypeOIDC,
							// Identifiers in OIDC are not email addresses, but a unique user ID.
							Identifiers: []string{"google:" + uuid.Must(uuid.NewV4()).String()},
							Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider": "google"},{"provider": "github"}]}`),
						},
					}

					first := createIdentity(email, field, creds)
					require.NoError(t, reg.IdentityManager().Create(context.Background(), first))

					second := createIdentity(email, field, nil)
					err := reg.IdentityManager().Create(context.Background(), second)
					require.Error(t, err)

					var verr = new(identity.ErrDuplicateCredentials)
					assert.ErrorAs(t, err, &verr)
					assert.EqualValues(t, []string{identity.CredentialsTypeOIDC.String()}, verr.AvailableCredentials())
					assert.EqualValues(t, verr.AvailableOIDCProviders(), []string{"google", "github"})
					assert.Equal(t, verr.IdentifierHint(), email)
				})
			}

			t.Run("case=verifiable address", func(t *testing.T) {
				runAddress(t, "email_verify")
			})

			t.Run("case=recovery address", func(t *testing.T) {
				runAddress(t, "email_recovery")
			})
		})
	})

	t.Run("method=Update", func(t *testing.T) {
		t.Run("case=should update identity and update extension fields", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits("baz@ory.sh", "")
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			original.Traits = newTraits("bar@ory.sh", "")
			require.NoError(t, reg.IdentityManager().Update(context.Background(), original, identity.ManagerAllowWriteProtectedTraits))

			checkExtensionFieldsForIdentities(t, "bar@ory.sh", original)
		})

		t.Run("case=should set AAL to 1 if password is set", func(t *testing.T) {
			email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(email, "")
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
			original.Credentials = map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{email},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
				},
			}
			require.NoError(t, reg.IdentityManager().Update(context.Background(), original, identity.ManagerAllowWriteProtectedTraits))
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel1, original.AvailableAAL.String)
		})

		t.Run("case=should set AAL to 2 if password and TOTP is set", func(t *testing.T) {
			email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(email, "")
			original.Credentials = map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{email},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
				},
			}
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel1, original.AvailableAAL.String)
			require.NoError(t, reg.IdentityManager().Update(context.Background(), original, identity.ManagerAllowWriteProtectedTraits))
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel1, original.AvailableAAL.String, "Updating without changes should not change AAL")
			original.Credentials[identity.CredentialsTypeTOTP] = identity.Credentials{
				Type:        identity.CredentialsTypeTOTP,
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
			}
			require.NoError(t, reg.IdentityManager().Update(context.Background(), original, identity.ManagerAllowWriteProtectedTraits))
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel2, original.AvailableAAL.String)
		})

		t.Run("case=should set AAL to 0 if only TOTP is set", func(t *testing.T) {
			email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(email, "")
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
			original.Credentials = map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeTOTP: {
					Type:        identity.CredentialsTypeTOTP,
					Identifiers: []string{email},
					Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
				},
			}
			require.NoError(t, reg.IdentityManager().Update(context.Background(), original, identity.ManagerAllowWriteProtectedTraits))
			assert.True(t, original.AvailableAAL.Valid)
			assert.EqualValues(t, identity.NoAuthenticatorAssuranceLevel, original.AvailableAAL.String)
		})

		t.Run("case=should not update protected traits without option", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits("email-update-1@ory.sh", "")
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			original.Traits = newTraits("email-update-2@ory.sh", "")
			err := reg.IdentityManager().Update(context.Background(), original)
			require.Error(t, err)
			assert.Equal(t, identity.ErrProtectedFieldModified, errors.Cause(err))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
			require.NoError(t, err)
			// As UpdateTraits takes only the ID as a parameter it cannot update the identity in place.
			// That is why we only check the identity in the store.
			checkExtensionFields(fromStore, "email-update-1@ory.sh")(t)
		})

		t.Run("case=should update unprotected traits with multiple credential identifiers", func(t *testing.T) {
			original := identity.NewIdentity(extensionSchemaID)
			original.Traits = identity.Traits(`{"email": "email-update-ewisdfuja@ory.sh", "names": ["username1", "username2"], "age": 30}`)
			require.NoError(t, reg.IdentityManager().Create(ctx, original))
			assert.Len(t, original.Credentials[identity.CredentialsTypePassword].Identifiers, 3)

			original.Traits = identity.Traits(`{"email": "email-update-ewisdfuja@ory.sh", "names": ["username1", "username2"], "age": 31}`)
			require.NoError(t, reg.IdentityManager().Update(ctx, original))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, original.ID)
			require.NoError(t, err)
			assert.JSONEq(t, string(original.Traits), string(fromStore.Traits))
		})

		t.Run("case=should update unprotected traits with verified user", func(t *testing.T) {
			email := x.NewUUID().String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(email, "initial")
			require.NoError(t, reg.IdentityManager().Create(ctx, original))

			// mock successful verification process
			addr := original.VerifiableAddresses[0]
			addr.Verified = true
			addr.VerifiedAt = pointerx.Ptr(sqlxx.NullTime(time.Now().UTC()))
			require.NoError(t, reg.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, &addr))

			// reload to properly set the verified address
			var err error
			original, err = reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, original.ID)
			require.NoError(t, err)

			original.Traits = newTraits(email, "updated")
			require.NoError(t, reg.IdentityManager().Update(ctx, original))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, original.ID)
			require.NoError(t, err)
			assert.JSONEq(t, string(original.Traits), string(fromStore.Traits))
		})

		t.Run("case=changing recovery address removes it from the store", func(t *testing.T) {
			originalEmail := x.NewUUID().String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(originalEmail, "")
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
			require.NoError(t, err)
			checkExtensionFields(fromStore, originalEmail)(t)

			newEmail := x.NewUUID().String() + "@ory.sh"
			original.Traits = newTraits(newEmail, "")
			require.NoError(t, reg.IdentityManager().Update(context.Background(), original, identity.ManagerAllowWriteProtectedTraits))

			fromStore, err = reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
			require.NoError(t, err)
			checkExtensionFields(fromStore, newEmail)(t)

			recoveryAddresses, err := reg.PrivilegedIdentityPool().ListRecoveryAddresses(context.Background(), 0, 500)
			require.NoError(t, err)

			var foundRecoveryAddress bool
			for _, a := range recoveryAddresses {
				assert.NotEqual(t, a.Value, originalEmail)
				if a.Value == newEmail {
					foundRecoveryAddress = true
				}
			}
			require.True(t, foundRecoveryAddress)

			verifiableAddresses, err := reg.PrivilegedIdentityPool().ListVerifiableAddresses(context.Background(), 0, 500)
			require.NoError(t, err)
			var foundVerifiableAddress bool
			for _, a := range verifiableAddresses {
				assert.NotEqual(t, a.Value, originalEmail)
				if a.Value == newEmail {
					foundVerifiableAddress = true
				}
			}
			require.True(t, foundVerifiableAddress)
		})
	})

	t.Run("method=CountActiveFirstFactorCredentials", func(t *testing.T) {
		id := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		count, err := reg.IdentityManager().CountActiveFirstFactorCredentials(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		id.Credentials[identity.CredentialsTypePassword] = identity.Credentials{
			Type:        identity.CredentialsTypePassword,
			Identifiers: []string{"foo"},
			Config:      []byte(`{"hashed_password":"$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`),
		}

		count, err = reg.IdentityManager().CountActiveFirstFactorCredentials(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("method=CountActiveMultiFactorCredentials", func(t *testing.T) {
		id := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		count, err := reg.IdentityManager().CountActiveMultiFactorCredentials(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		id.Credentials[identity.CredentialsTypePassword] = identity.Credentials{
			Type:        identity.CredentialsTypePassword,
			Identifiers: []string{"foo"},
			Config:      []byte(`{"hashed_password":"$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`),
		}

		count, err = reg.IdentityManager().CountActiveMultiFactorCredentials(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		id.Credentials[identity.CredentialsTypeWebAuthn] = identity.Credentials{
			Type:        identity.CredentialsTypeWebAuthn,
			Identifiers: []string{"foo"},
			Config:      []byte(`{"credentials":[{"is_passwordless":false}]}`),
		}

		count, err = reg.IdentityManager().CountActiveMultiFactorCredentials(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("method=UpdateTraits", func(t *testing.T) {
		t.Run("case=should update protected traits with option", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits("email-updatetraits-1@ory.sh", "")
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			require.NoError(t, reg.IdentityManager().UpdateTraits(
				context.Background(), original.ID, newTraits("email-updatetraits-2@ory.sh", ""),
				identity.ManagerAllowWriteProtectedTraits))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
			require.NoError(t, err)
			// As UpdateTraits takes only the ID as a parameter it cannot update the identity in place.
			// That is why we only check the identity in the store.
			checkExtensionFields(fromStore, "email-updatetraits-2@ory.sh")(t)
		})

		t.Run("case=should update identity and update extension fields", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			// These should all fail because they modify existing keys
			require.Error(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"not-baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)))
			require.Error(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"not-baz@ory.sh","email_recovery":"not-baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)))
			require.Error(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"not-baz@ory.sh","unprotected": "foo"}`)))

			require.NoError(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "bar"}`)))
			checkExtensionFieldsForIdentities(t, "baz@ory.sh", original)

			actual, err := reg.IdentityPool().GetIdentity(context.Background(), original.ID, identity.ExpandNothing)
			require.NoError(t, err)
			assert.JSONEq(t, `{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "bar"}`, string(actual.Traits))
		})

		t.Run("case=should not update protected traits without option", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits("email-updatetraits-1@ory.sh", "")
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			err := reg.IdentityManager().UpdateTraits(
				context.Background(), original.ID, newTraits("email-updatetraits-2@ory.sh", ""))
			require.Error(t, err)
			assert.Equal(t, identity.ErrProtectedFieldModified, errors.Cause(err))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
			require.NoError(t, err)
			// As UpdateTraits takes only the ID as a parameter it cannot update the identity in place.
			// That is why we only check the identity in the store.
			checkExtensionFields(fromStore, "email-updatetraits-1@ory.sh")(t)
		})
	})
}

func TestManagerNoDefaultNamedSchema(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyDefaultIdentitySchemaID, "user_v0")
	conf.MustSet(ctx, config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "user_v0", URL: "file://./stub/manager.schema.json"},
	})
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")

	t.Run("case=should create identity with default schema", func(t *testing.T) {
		stateChangedAt := sqlxx.NullTime(time.Now().UTC())
		original := &identity.Identity{
			SchemaID:       "",
			Traits:         []byte(identity.Traits(`{"email":"foo@ory.sh"}`)),
			State:          identity.StateActive,
			StateChangedAt: &stateChangedAt,
		}
		require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
	})
}
