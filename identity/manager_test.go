// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ory/x/configx"
	"github.com/ory/x/pointerx"
	"github.com/ory/x/sqlcon"

	_ "embed"

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

//go:embed stub/aal.json
var refreshAALStubs []byte

func TestManager(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t, configx.WithValues(map[string]interface{}{
		config.ViperKeyPublicBaseURL:                     "https://www.ory.sh/",
		config.ViperKeyCourierSMTPURL:                    "smtp://foo@bar@dev.null/",
		config.ViperKeySelfServiceRegistrationLoginHints: true,
	}), configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://./stub/manager.schema.json")))
	ctx, extensionSchemaID := testhelpers.WithAddIdentitySchema(ctx, t, conf, "file://./stub/extension.schema.json")

	t.Run("case=should fail to create because validation fails", func(t *testing.T) {
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.Traits = identity.Traits(`{"email":"not an email"}`)
		require.Error(t, reg.IdentityManager().Create(ctx, i))
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
		fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, original.ID)
		require.NoError(t, err)
		identities := []identity.Identity{*original, *fromStore}
		for k := range identities {
			t.Run(fmt.Sprintf("identity=%d", k), checkExtensionFields(&identities[k], expected))
		}
	}

	t.Run("method=CreateIdentities", func(t *testing.T) {
		t.Run("case=should set AAL to 2 if password and TOTP is set", func(t *testing.T) {
			email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(email, "")
			original.Credentials = map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type: identity.CredentialsTypePassword,
					// By explicitly not setting the identifier, we mimic the behavior of the PATCH endpoint.
					// This tests a bug we introduced on the PATCH endpoint where the AAL value would not be correct.
					Identifiers: []string{},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
				},
				identity.CredentialsTypeTOTP: {
					Type: identity.CredentialsTypeTOTP,
					// By explicitly not setting the identifier, we mimic the behavior of the PATCH endpoint.
					// This tests a bug we introduced on the PATCH endpoint where the AAL value would not be correct.
					Identifiers: []string{},
					Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
				},
			}
			require.NoError(t, reg.IdentityManager().CreateIdentities(ctx, []*identity.Identity{original}))
			fromStore, err := reg.PrivilegedIdentityPool().GetIdentity(ctx, original.ID, identity.ExpandNothing)
			require.NoError(t, err)

			got, ok := fromStore.InternalAvailableAAL.ToAAL()
			require.True(t, ok)
			assert.Equal(t, identity.AuthenticatorAssuranceLevel2, got)
		})
	})

	t.Run("method=Create", func(t *testing.T) {
		t.Run("case=should create identity and track extension fields", func(t *testing.T) {
			email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(email, "")
			require.NoError(t, reg.IdentityManager().Create(ctx, original))
			checkExtensionFieldsForIdentities(t, email, original)
			got, ok := original.InternalAvailableAAL.ToAAL()
			require.True(t, ok)
			assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, got)
		})

		t.Run("case=correctly set AAL", func(t *testing.T) {
			t.Run("case=should set AAL to 0 if no credentials are available", func(t *testing.T) {
				email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
				original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				original.Traits = newTraits(email, "")
				require.NoError(t, reg.IdentityManager().Create(ctx, original))
				got, ok := original.InternalAvailableAAL.ToAAL()
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
				require.NoError(t, reg.IdentityManager().Create(ctx, original))
				got, ok := original.InternalAvailableAAL.ToAAL()
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
				require.NoError(t, reg.IdentityManager().Create(ctx, original))
				got, ok := original.InternalAvailableAAL.ToAAL()
				require.True(t, ok)
				assert.Equal(t, identity.AuthenticatorAssuranceLevel2, got)
			})

			t.Run("case=should set AAL to 2 if only TOTP is set", func(t *testing.T) {
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
				require.NoError(t, reg.IdentityManager().Create(ctx, original))
				got, ok := original.InternalAvailableAAL.ToAAL()
				require.True(t, ok)
				assert.Equal(t, identity.AuthenticatorAssuranceLevel2, got)
			})
		})

		t.Run("case=should expose validation errors with option", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"not an email"}`)
			err := reg.IdentityManager().Create(ctx, original, identity.ManagerExposeValidationErrorsForInternalTypeAssertion)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "\"not an email\" is not valid \"email\"")
		})

		t.Run("case=should not expose validation errors without option", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"not an email"}`)
			err := reg.IdentityManager().Create(ctx, original)
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
					require.NoError(t, reg.IdentityManager().Create(ctx, first))

					second := createIdentity(email, "email_creds", creds)
					err := reg.IdentityManager().Create(ctx, second)
					require.Error(t, err)

					verr := new(identity.ErrDuplicateCredentials)
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
					require.NoError(t, reg.IdentityManager().Create(ctx, first))

					second := createIdentity(email, "email_webauthn", nil)
					err := reg.IdentityManager().Create(ctx, second)
					require.Error(t, err)

					verr := new(identity.ErrDuplicateCredentials)
					assert.ErrorAs(t, err, &verr)
					assert.EqualValues(t, []string{identity.CredentialsTypeWebAuthn.String()}, verr.AvailableCredentials())
					assert.Len(t, verr.AvailableOIDCProviders(), 0)
					assert.Equal(t, verr.IdentifierHint(), email)
				})

				t.Run("type=oidc", func(t *testing.T) {
					email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
					creds := map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypeOIDC: {
							Type: identity.CredentialsTypeOIDC,
							// Identifiers in OIDC are not email addresses, but a unique user ID.
							Identifiers: []string{"google:" + uuid.Must(uuid.NewV4()).String()},
							Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider": "google"},{"provider": "github"}]}`),
						},
					}

					first := createIdentity(email, "email_creds", creds)
					require.NoError(t, reg.IdentityManager().Create(ctx, first))

					second := createIdentity(email, "email_creds", creds)
					err := reg.IdentityManager().Create(ctx, second)
					require.Error(t, err)

					verr := new(identity.ErrDuplicateCredentials)
					assert.ErrorAs(t, err, &verr)
					assert.ElementsMatch(t, []string{"oidc"}, verr.AvailableCredentials())
					assert.ElementsMatch(t, []string{"google", "github"}, verr.AvailableOIDCProviders())
					assert.Equal(t, email, verr.IdentifierHint())
				})

				t.Run("type=password+oidc+webauthn", func(t *testing.T) {
					email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
					creds := map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {
							Type:        identity.CredentialsTypePassword,
							Identifiers: []string{email},
							Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
						},
						identity.CredentialsTypeOIDC: {
							Type: identity.CredentialsTypeOIDC,
							// Identifiers in OIDC are not email addresses, but a unique user ID.
							Identifiers: []string{"google:" + uuid.Must(uuid.NewV4()).String()},
							Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider": "google"},{"provider": "github"}]}`),
						},
						identity.CredentialsTypeWebAuthn: {
							Type:        identity.CredentialsTypeWebAuthn,
							Identifiers: []string{email},
							Config:      sqlxx.JSONRawMessage(`{"credentials": [{"is_passwordless":true}]}`),
						},
					}

					first := createIdentity(email, "email_creds", creds)
					require.NoError(t, reg.IdentityManager().Create(ctx, first))

					second := createIdentity(email, "email_creds", creds)
					err := reg.IdentityManager().Create(ctx, second)
					require.Error(t, err)

					verr := new(identity.ErrDuplicateCredentials)
					assert.ErrorAs(t, err, &verr)
					assert.ElementsMatch(t, []string{"password", "oidc", "webauthn"}, verr.AvailableCredentials())
					assert.ElementsMatch(t, []string{"google", "github"}, verr.AvailableOIDCProviders())
					assert.Equal(t, email, verr.IdentifierHint())
				})

				t.Run("type=code", func(t *testing.T) {
					email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
					creds := map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypeCodeAuth: {
							Type:        identity.CredentialsTypeCodeAuth,
							Identifiers: []string{email},
							Config:      sqlxx.JSONRawMessage(`{}`),
						},
					}

					first := createIdentity(email, "email_creds", creds)
					require.NoError(t, reg.IdentityManager().Create(ctx, first))

					second := createIdentity(email, "email_creds", creds)
					err := reg.IdentityManager().Create(ctx, second)
					require.Error(t, err)

					verr := new(identity.ErrDuplicateCredentials)
					assert.ErrorAs(t, err, &verr)
					assert.EqualValues(t, []string{identity.CredentialsTypeCodeAuth.String()}, verr.AvailableCredentials())
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
					require.NoError(t, reg.IdentityManager().Create(ctx, first))

					second := createIdentity(email, field, nil)
					err := reg.IdentityManager().Create(ctx, second)
					require.Error(t, err)

					verr := new(identity.ErrDuplicateCredentials)
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
					require.NoError(t, reg.IdentityManager().Create(ctx, first))

					second := createIdentity(email, field, nil)
					err := reg.IdentityManager().Create(ctx, second)
					require.Error(t, err)

					verr := new(identity.ErrDuplicateCredentials)
					assert.ErrorAs(t, err, &verr)
					assert.EqualValues(t, []string{identity.CredentialsTypeOIDC.String()}, verr.AvailableCredentials())
					assert.EqualValues(t, verr.AvailableOIDCProviders(), []string{"github", "google"})
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
			require.NoError(t, reg.IdentityManager().Create(ctx, original))

			original.Traits = newTraits("bar@ory.sh", "")
			require.NoError(t, reg.IdentityManager().Update(ctx, original, identity.ManagerAllowWriteProtectedTraits))

			checkExtensionFieldsForIdentities(t, "bar@ory.sh", original)
		})

		t.Run("case=should set AAL to 1 if password is set", func(t *testing.T) {
			email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(email, "")
			require.NoError(t, reg.IdentityManager().Create(ctx, original))
			original.Credentials = map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{email},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
				},
			}
			require.NoError(t, reg.IdentityManager().Update(ctx, original, identity.ManagerAllowWriteProtectedTraits))
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel1, original.InternalAvailableAAL.String)
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
			require.NoError(t, reg.IdentityManager().Create(ctx, original))
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel1, original.InternalAvailableAAL.String)
			require.NoError(t, reg.IdentityManager().Update(ctx, original, identity.ManagerAllowWriteProtectedTraits))
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel1, original.InternalAvailableAAL.String, "Updating without changes should not change AAL")
			original.Credentials[identity.CredentialsTypeTOTP] = identity.Credentials{
				Type:        identity.CredentialsTypeTOTP,
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
			}
			require.NoError(t, reg.IdentityManager().Update(ctx, original, identity.ManagerAllowWriteProtectedTraits))
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel2, original.InternalAvailableAAL.String)
		})

		t.Run("case=should set AAL to 2 if only TOTP is set", func(t *testing.T) {
			email := uuid.Must(uuid.NewV4()).String() + "@ory.sh"
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits(email, "")
			require.NoError(t, reg.IdentityManager().Create(ctx, original))
			original.Credentials = map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeTOTP: {
					Type:        identity.CredentialsTypeTOTP,
					Identifiers: []string{email},
					Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
				},
			}
			require.NoError(t, reg.IdentityManager().Update(ctx, original, identity.ManagerAllowWriteProtectedTraits))
			assert.True(t, original.InternalAvailableAAL.Valid)
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel2, original.InternalAvailableAAL.String)
		})

		t.Run("case=should not update protected traits without option", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits("email-update-1@ory.sh", "")
			require.NoError(t, reg.IdentityManager().Create(ctx, original))

			original.Traits = newTraits("email-update-2@ory.sh", "")
			err := reg.IdentityManager().Update(ctx, original)
			require.Error(t, err)
			assert.Equal(t, identity.ErrProtectedFieldModified, errors.Cause(err))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, original.ID)
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
			require.NoError(t, reg.IdentityManager().Create(ctx, original))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, original.ID)
			require.NoError(t, err)
			checkExtensionFields(fromStore, originalEmail)(t)

			newEmail := x.NewUUID().String() + "@ory.sh"
			original.Traits = newTraits(newEmail, "")
			require.NoError(t, reg.IdentityManager().Update(ctx, original, identity.ManagerAllowWriteProtectedTraits))

			fromStore, err = reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, original.ID)
			require.NoError(t, err)
			checkExtensionFields(fromStore, newEmail)(t)

			recoveryAddresses, err := reg.PrivilegedIdentityPool().ListRecoveryAddresses(ctx, 0, 500)
			require.NoError(t, err)

			var foundRecoveryAddress bool
			for _, a := range recoveryAddresses {
				assert.NotEqual(t, a.Value, originalEmail)
				if a.Value == newEmail {
					foundRecoveryAddress = true
				}
			}
			require.True(t, foundRecoveryAddress)

			verifiableAddresses, err := reg.PrivilegedIdentityPool().ListVerifiableAddresses(ctx, 0, 500)
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
			require.NoError(t, reg.IdentityManager().Create(ctx, original))

			require.NoError(t, reg.IdentityManager().UpdateTraits(
				ctx, original.ID, newTraits("email-updatetraits-2@ory.sh", ""),
				identity.ManagerAllowWriteProtectedTraits))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, original.ID)
			require.NoError(t, err)
			// As UpdateTraits takes only the ID as a parameter it cannot update the identity in place.
			// That is why we only check the identity in the store.
			checkExtensionFields(fromStore, "email-updatetraits-2@ory.sh")(t)
		})

		t.Run("case=should update identity and update extension fields", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)
			require.NoError(t, reg.IdentityManager().Create(ctx, original))

			// These should all fail because they modify existing keys
			require.Error(t, reg.IdentityManager().UpdateTraits(ctx, original.ID, identity.Traits(`{"email":"not-baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)))
			require.Error(t, reg.IdentityManager().UpdateTraits(ctx, original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"not-baz@ory.sh","email_recovery":"not-baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)))
			require.Error(t, reg.IdentityManager().UpdateTraits(ctx, original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"not-baz@ory.sh","unprotected": "foo"}`)))

			require.NoError(t, reg.IdentityManager().UpdateTraits(ctx, original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "bar"}`)))
			checkExtensionFieldsForIdentities(t, "baz@ory.sh", original)

			actual, err := reg.IdentityPool().GetIdentity(ctx, original.ID, identity.ExpandNothing)
			require.NoError(t, err)
			assert.JSONEq(t, `{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_recovery":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "bar"}`, string(actual.Traits))
		})

		t.Run("case=should not update protected traits without option", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits("email-updatetraits-1@ory.sh", "")
			require.NoError(t, reg.IdentityManager().Create(ctx, original))

			err := reg.IdentityManager().UpdateTraits(
				ctx, original.ID, newTraits("email-updatetraits-2@ory.sh", ""))
			require.Error(t, err)
			assert.Equal(t, identity.ErrProtectedFieldModified, errors.Cause(err))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, original.ID)
			require.NoError(t, err)
			// As UpdateTraits takes only the ID as a parameter it cannot update the identity in place.
			// That is why we only check the identity in the store.
			checkExtensionFields(fromStore, "email-updatetraits-1@ory.sh")(t)
		})

		t.Run("case=should always update updated_at field", func(t *testing.T) {
			original := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			original.Traits = newTraits("email-updatetraits-3@ory.sh", "")
			require.NoError(t, reg.IdentityManager().Create(ctx, original))

			time.Sleep(time.Millisecond)

			require.NoError(t, reg.IdentityManager().UpdateTraits(
				ctx, original.ID, newTraits("email-updatetraits-4@ory.sh", ""),
				identity.ManagerAllowWriteProtectedTraits))

			updated, err := reg.IdentityPool().GetIdentity(ctx, original.ID, identity.ExpandNothing)
			require.NoError(t, err)
			assert.NotEqual(t, original.UpdatedAt, updated.UpdatedAt, "UpdatedAt field should be updated")
		})
	})

	t.Run("method=RefreshAvailableAAL", func(t *testing.T) {
		var cases []struct {
			Credentials []identity.Credentials `json:"credentials"`
			Description string                 `json:"description"`
			Expected    string                 `json:"expected"`
		}
		require.NoError(t, json.Unmarshal(refreshAALStubs, &cases))

		for k, tc := range cases {
			t.Run("case="+tc.Description, func(t *testing.T) {
				email := x.NewUUID().String() + "@ory.sh"
				id := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				id.Traits = identity.Traits(`{"email":"` + email + `"}`)
				require.NoError(t, reg.IdentityManager().Create(ctx, id))
				assert.EqualValues(t, identity.NoAuthenticatorAssuranceLevel, id.InternalAvailableAAL.String)

				for _, c := range tc.Credentials {
					for k := range c.Identifiers {
						switch c.Identifiers[k] {
						case "{email}":
							c.Identifiers[k] = email
						case "{id}":
							c.Identifiers[k] = id.ID.String()
						}
					}
					id.SetCredentials(c.Type, c)
				}

				// We use the privileged pool here because we don't want to refresh AAL here but in the code below.
				require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(ctx, id))

				expand := identity.ExpandNothing
				if k%2 == 1 { // expand every other test case to test if RefreshAvailableAAL behaves correctly
					expand = identity.ExpandCredentials
				}

				actual, err := reg.IdentityPool().GetIdentity(ctx, id.ID, expand)
				require.NoError(t, err)
				require.NoError(t, reg.IdentityManager().RefreshAvailableAAL(ctx, actual))
				assert.NotEmpty(t, actual.Credentials)
				assert.EqualValues(t, tc.Expected, actual.InternalAvailableAAL.String)
			})
		}
	})

	t.Run("method=ConflictingIdentity", func(t *testing.T) {
		ctx := ctx

		conflicOnIdentifier := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		conflicOnIdentifier.Traits = identity.Traits(`{"email":"conflict-on-identifier@example.com"}`)
		require.NoError(t, reg.IdentityManager().Create(ctx, conflicOnIdentifier))

		conflicOnVerifiableAddress := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		conflicOnVerifiableAddress.Traits = identity.Traits(`{"email":"user-va@example.com", "email_verify":"conflict-on-va@example.com"}`)
		require.NoError(t, reg.IdentityManager().Create(ctx, conflicOnVerifiableAddress))

		conflicOnRecoveryAddress := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		conflicOnRecoveryAddress.Traits = identity.Traits(`{"email":"user-ra@example.com", "email_recovery":"conflict-on-ra@example.com"}`)
		require.NoError(t, reg.IdentityManager().Create(ctx, conflicOnRecoveryAddress))

		t.Run("case=returns not found if no conflict", func(t *testing.T) {
			found, foundConflictAddress, err := reg.IdentityManager().ConflictingIdentity(ctx, &identity.Identity{
				Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypePassword: {Identifiers: []string{"no-conflict@example.com"}},
				},
			})
			assert.ErrorIs(t, err, sqlcon.ErrNoRows)
			assert.Nil(t, found)
			assert.Empty(t, foundConflictAddress)
		})

		t.Run("case=conflict on identifier", func(t *testing.T) {
			found, foundConflictAddress, err := reg.IdentityManager().ConflictingIdentity(ctx, &identity.Identity{
				Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypePassword: {Identifiers: []string{"conflict-on-identifier@example.com"}},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, conflicOnIdentifier.ID, found.ID)
			assert.Equal(t, "conflict-on-identifier@example.com", foundConflictAddress)
		})

		t.Run("case=conflict on verifiable address", func(t *testing.T) {
			found, foundConflictAddress, err := reg.IdentityManager().ConflictingIdentity(ctx, &identity.Identity{
				VerifiableAddresses: []identity.VerifiableAddress{{
					Value: "conflict-on-va@example.com",
					Via:   "email",
				}},
			})
			require.NoError(t, err)
			assert.Equal(t, conflicOnVerifiableAddress.ID, found.ID)
			assert.Equal(t, "conflict-on-va@example.com", foundConflictAddress)
		})

		t.Run("case=conflict on recovery address", func(t *testing.T) {
			found, foundConflictAddress, err := reg.IdentityManager().ConflictingIdentity(ctx, &identity.Identity{
				RecoveryAddresses: []identity.RecoveryAddress{{
					Value: "conflict-on-ra@example.com",
					Via:   "email",
				}},
			})
			require.NoError(t, err)
			assert.Equal(t, conflicOnRecoveryAddress.ID, found.ID)
			assert.Equal(t, "conflict-on-ra@example.com", foundConflictAddress)
		})
	})
}

func TestManagerNoDefaultNamedSchema(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t, configx.WithValues(map[string]interface{}{
		config.ViperKeyDefaultIdentitySchemaID: "user_v0",
		config.ViperKeyIdentitySchemas: config.Schemas{
			{ID: "user_v0", URL: "file://./stub/manager.schema.json"},
		},
		config.ViperKeyPublicBaseURL: "https://www.ory.sh/",
	}))

	t.Run("case=should create identity with default schema", func(t *testing.T) {
		stateChangedAt := sqlxx.NullTime(time.Now().UTC())
		original := &identity.Identity{
			SchemaID:       "",
			Traits:         []byte(identity.Traits(`{"email":"foo@ory.sh"}`)),
			State:          identity.StateActive,
			StateChangedAt: &stateChangedAt,
		}
		require.NoError(t, reg.IdentityManager().Create(ctx, original))
	})
}
