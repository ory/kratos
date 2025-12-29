// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"
)

func TestVerificationWithSessionHook(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)

	conf.MustSet(ctx, config.ViperKeySelfServiceVerificationAfter+".hooks", []map[string]interface{}{
		{"hook": "session"},
	})

	_ = testhelpers.NewVerificationUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)

	createIdentity := func(email string) *identity.Identity {
		id := &identity.Identity{
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{email},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`),
				},
			},
			Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, email)),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			State:    identity.StateActive,
		}
		require.NoError(t, reg.IdentityManager().Create(ctx, id, identity.ManagerAllowWriteProtectedTraits))
		return id
	}

	t.Run("case=browser flow creates session after verification", func(t *testing.T) {
		email := testhelpers.RandomEmail()
		_ = createIdentity(email)

		client := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeVerificationFlowViaBrowser(t, client, false, public)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("email", email)
		values.Set("method", "link")

		body, resp := testhelpers.VerificationMakeRequest(t, false, f, client, testhelpers.EncodeFormAsJSON(t, false, values))
		require.EqualValues(t, http.StatusOK, resp.StatusCode, body)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Please verify")
		require.Contains(t, message.Body, "Verify your account")

		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)
		require.NotEmpty(t, verificationLink)

		resp, err := client.Get(verificationLink)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		session, _, err := testhelpers.NewSDKCustomClient(public, client).FrontendAPI.ToSession(ctx).Execute()
		require.NoError(t, err)
		require.NotNil(t, session)
		traitsJSON, _ := session.Identity.Traits.(map[string]interface{})
		assert.Equal(t, email, traitsJSON["email"])
	})

	t.Run("case=SPA flow creates session after verification", func(t *testing.T) {
		email := testhelpers.RandomEmail()
		_ = createIdentity(email)

		client := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeVerificationFlowViaBrowser(t, client, true, public)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("email", email)
		values.Set("method", "link")

		body, resp := testhelpers.VerificationMakeRequest(t, false, f, client, testhelpers.EncodeFormAsJSON(t, false, values))
		require.EqualValues(t, http.StatusOK, resp.StatusCode, body)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Please verify")
		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		resp, err := client.Get(verificationLink)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		session, _, err := testhelpers.NewSDKCustomClient(public, client).FrontendAPI.ToSession(ctx).Execute()
		require.NoError(t, err)
		require.NotNil(t, session)
		traitsJSON, _ := session.Identity.Traits.(map[string]interface{})
		assert.Equal(t, email, traitsJSON["email"])
	})

	t.Run("case=API flow returns session token after verification", func(t *testing.T) {
		email := testhelpers.RandomEmail()
		_ = createIdentity(email)

		client := &http.Client{}
		f := testhelpers.InitializeVerificationFlowViaAPI(t, client, public)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("email", email)
		values.Set("method", "link")

		body, resp := testhelpers.VerificationMakeRequest(t, true, f, client, testhelpers.EncodeFormAsJSON(t, true, values))
		require.EqualValues(t, http.StatusOK, resp.StatusCode, body)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Please verify")
		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		resp, err := client.Get(verificationLink)
		require.NoError(t, err)
		body = string(ioutilx.MustReadAll(resp.Body))
		require.NoError(t, resp.Body.Close())
		require.EqualValues(t, http.StatusOK, resp.StatusCode, body)

		sessionToken := gjson.Get(body, "session_token").String()
		require.NotEmpty(t, sessionToken, "session_token should be present in API response")

		sessionData := gjson.Get(body, "session")
		require.True(t, sessionData.Exists(), "session should be present in API response")
		assert.Equal(t, email, gjson.Get(body, "session.identity.traits.email").String())

		identityData := gjson.Get(body, "identity")
		require.True(t, identityData.Exists(), "identity should be present")
		assert.Equal(t, email, gjson.Get(body, "identity.traits.email").String())
	})

	t.Run("case=continue_with contains session info for API flow", func(t *testing.T) {
		email := testhelpers.RandomEmail()
		_ = createIdentity(email)

		client := &http.Client{}
		f := testhelpers.InitializeVerificationFlowViaAPI(t, client, public)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("email", email)
		values.Set("method", "link")

		body, resp := testhelpers.VerificationMakeRequest(t, true, f, client, testhelpers.EncodeFormAsJSON(t, true, values))
		require.EqualValues(t, http.StatusOK, resp.StatusCode, body)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Please verify")
		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		resp, err := client.Get(verificationLink)
		require.NoError(t, err)
		body = string(ioutilx.MustReadAll(resp.Body))
		require.NoError(t, resp.Body.Close())
		require.EqualValues(t, http.StatusOK, resp.StatusCode, body)

		continueWith := gjson.Get(body, "continue_with").Array()
		require.NotEmpty(t, continueWith, "continue_with should be present")

		var found bool
		for _, item := range continueWith {
			if item.Get("action").String() == "set_ory_session_token" {
				found = true
				assert.NotEmpty(t, item.Get("ory_session_token").String())
				break
			}
		}
		assert.True(t, found, "continue_with should contain set_ory_session_token action")
	})
}
