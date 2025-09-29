// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/hook"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func TestAddressVerifier(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.schema.json")
	verifier := hook.NewAddressVerifier(reg)

	t.Run("legacy behavior", func(t *testing.T) {
		// Mock legacy config
		conf.MustSet(context.Background(), config.ViperKeyUseLegacyRequireVerifiedLoginError, true)

		for _, tc := range []struct {
			flow       *login.Flow
			neverError bool
		}{
			{&login.Flow{Active: identity.CredentialsTypePassword}, false},
			{&login.Flow{Active: identity.CredentialsTypeOIDC}, true},
		} {
			t.Run(tc.flow.Active.String()+" flow", func(t *testing.T) {
				for _, uc := range []struct {
					name                string
					verifiableAddresses []identity.VerifiableAddress
					expectedError       error
				}{
					{
						name:                "No Verification Address",
						verifiableAddresses: []identity.VerifiableAddress{},
						expectedError:       herodot.ErrMisconfiguration.WithReason("A misconfiguration prevents login. Expected to find a verification address but this identity does not have one assigned."),
					},
					{
						name: "Single Address Not Verified",
						verifiableAddresses: []identity.VerifiableAddress{
							{ID: uuid.UUID{}, Verified: false},
						},
						expectedError: login.ErrAddressNotVerified,
					},
					{
						name: "Single Address Verified",
						verifiableAddresses: []identity.VerifiableAddress{
							{ID: uuid.UUID{}, Verified: true},
						},
					},
					{
						name: "Multiple Addresses Verified",
						verifiableAddresses: []identity.VerifiableAddress{
							{ID: uuid.UUID{}, Verified: true},
							{ID: uuid.UUID{}, Verified: true},
						},
					},
					{
						name: "Multiple Addresses Not Verified",
						verifiableAddresses: []identity.VerifiableAddress{
							{ID: uuid.UUID{}, Verified: false},
							{ID: uuid.UUID{}, Verified: false},
						},
						expectedError: login.ErrAddressNotVerified,
					},
					{
						name: "One Address Verified And One Not",
						verifiableAddresses: []identity.VerifiableAddress{
							{ID: uuid.UUID{}, Verified: true},
							{ID: uuid.UUID{}, Verified: false},
						},
					},
				} {
					t.Run(uc.name, func(t *testing.T) {
						sessions := &session.Session{
							ID:       x.NewUUID(),
							Identity: &identity.Identity{ID: x.NewUUID(), VerifiableAddresses: uc.verifiableAddresses},
						}
						err := verifier.ExecuteLoginPostHook(nil, httptest.NewRequest("GET", "http://example.com", nil), node.DefaultGroup, tc.flow, sessions)
						if tc.neverError || uc.expectedError == nil {
							assert.NoError(t, err)
						} else {
							assert.ErrorIs(t, err, uc.expectedError)
						}
					})
				}
			})
		}
	})

	t.Run("current behavior", func(t *testing.T) {
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/require_verified.schema.json")
		// Mock non-legacy config
		conf.MustSet(context.Background(), config.ViperKeyUseLegacyRequireVerifiedLoginError, false)

		mockRequest := httptest.NewRequest("GET", "http://example.com", nil)

		// Setup flow with identity addresses needing verification
		loginFlow := &login.Flow{
			Active: identity.CredentialsTypePassword,
			ID:     x.NewUUID(),
		}

		// Must persist the flow first for foreign key constraints
		require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), loginFlow))
		t.Run("json request for unverified address", func(t *testing.T) {
			// Mock JSON request
			mockJSONReq := httptest.NewRequest("GET", "http://example.com", nil)
			mockJSONReq.Header.Set("Content-Type", "application/json")
			mockJSONReq.Header.Set("Accept", "application/json")
			mockResponse := httptest.NewRecorder()

			// Create identity with valid values
			identity := &identity.Identity{
				ID:     x.NewUUID(),
				Traits: identity.Traits(`{"email":"user@example.com"}`),
				VerifiableAddresses: []identity.VerifiableAddress{
					{
						ID:       x.NewUUID(),
						Value:    "user@example.com",
						Verified: false,
						Via:      identity.VerifiableAddressTypeEmail,
					},
				},
			}

			// Persist identity to satisfy foreign key constraints
			require.NoError(t, reg.IdentityManager().Create(context.Background(), identity))

			sessions := &session.Session{
				ID:       x.NewUUID(),
				Identity: identity,
			}

			// Expect verification flow creation and ErrHookAbortFlow
			err := verifier.ExecuteLoginPostHook(mockResponse, mockJSONReq, node.DefaultGroup, loginFlow, sessions)
			assert.ErrorIs(t, err, login.ErrHookAbortFlow)

			// Verify response contains continueWith and ErrAddressNotVerified
			resp := mockResponse.Result()
			body, _ := io.ReadAll(resp.Body)

			var responseErr *herodot.DefaultError

			// Check for required JSON fields
			require.NoError(t, json.Unmarshal([]byte(gjson.GetBytes(body, "error").Raw), &responseErr))
			assert.Contains(t, responseErr.ReasonField, login.ErrAddressNotVerified.Reason(), "%s", string(body))

			// Verify flow has continueWith added
			var continueWith flow.ContinueWithVerificationUI
			require.NoError(t, json.Unmarshal([]byte(gjson.GetBytes(body, "error.details.continue_with.0").Raw), &continueWith))

			// Verify the continueWith fields are properly set
			assert.NotEmpty(t, continueWith.Flow.ID)
			assert.Contains(t, continueWith.Action, "show_verification_ui")
			assert.Contains(t, continueWith.Flow.URL, reg.Config().SelfServiceFlowVerificationUI(context.Background()).String())
			assert.Equal(t, continueWith.Flow.VerifiableAddress, "user@example.com")
		})

		t.Run("browser request for unverified address", func(t *testing.T) {
			// Mock browser request
			mockBrowserReq := httptest.NewRequest("GET", "http://example.com", nil)
			mockBrowserReq.Header.Set("Accept", "text/html")
			mockResponse := httptest.NewRecorder()

			// Create new flow for this test case
			browserFlow := &login.Flow{
				Active: identity.CredentialsTypePassword,
				ID:     x.NewUUID(),
			}
			require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), browserFlow))

			// Create identity with valid values
			identity := &identity.Identity{
				ID:     x.NewUUID(),
				Traits: identity.Traits(`{"email":"user2@example.com"}`),
				VerifiableAddresses: []identity.VerifiableAddress{
					{
						ID:       x.NewUUID(),
						Value:    "user2@example.com",
						Verified: false,
						Via:      identity.VerifiableAddressTypeEmail,
					},
				},
			}
			require.NoError(t, reg.IdentityManager().Create(context.Background(), identity))

			sessions := &session.Session{
				ID:       x.NewUUID(),
				Identity: identity,
			}

			// Expect verification flow creation and redirect
			err := verifier.ExecuteLoginPostHook(mockResponse, mockBrowserReq, node.DefaultGroup, browserFlow, sessions)
			assert.ErrorIs(t, err, login.ErrHookAbortFlow)

			// Verify redirect occurred
			resp := mockResponse.Result()
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
			assert.NotEmpty(t, resp.Header.Get("Location"))
		})

		t.Run("verified address skips verification", func(t *testing.T) {
			// Create new flow for this test case
			verifiedFlow := &login.Flow{
				Active: identity.CredentialsTypePassword,
				ID:     x.NewUUID(),
			}
			require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), verifiedFlow))

			// Create identity with verified address
			identity := &identity.Identity{
				ID:     x.NewUUID(),
				Traits: identity.Traits(`{"email":"verified@example.com"}`),
				VerifiableAddresses: []identity.VerifiableAddress{
					{
						ID:       x.NewUUID(),
						Value:    "verified@example.com",
						Verified: true,
						Via:      identity.VerifiableAddressTypeEmail,
					},
				},
			}
			require.NoError(t, reg.IdentityManager().Create(context.Background(), identity))

			sessions := &session.Session{
				ID:       x.NewUUID(),
				Identity: identity,
			}

			err := verifier.ExecuteLoginPostHook(nil, mockRequest, node.DefaultGroup, verifiedFlow, sessions)
			assert.NoError(t, err)
		})

		t.Run("no verifiable address", func(t *testing.T) {
			// Create new flow for this test case
			noAddressFlow := &login.Flow{
				Active: identity.CredentialsTypePassword,
				ID:     x.NewUUID(),
			}
			require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), noAddressFlow))

			// Create identity with no verifiable addresses
			identity := &identity.Identity{
				ID:                  x.NewUUID(),
				Traits:              identity.Traits(`{}`),
				VerifiableAddresses: []identity.VerifiableAddress{},
			}
			require.NoError(t, reg.IdentityManager().Create(context.Background(), identity))

			sessions := &session.Session{
				ID:       x.NewUUID(),
				Identity: identity,
			}

			err := verifier.ExecuteLoginPostHook(nil, mockRequest, node.DefaultGroup, noAddressFlow, sessions)
			assert.ErrorIs(t, err, herodot.ErrMisconfiguration)
		})
	})
}
