// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"context"
	"encoding/json"
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
	"github.com/ory/x/configx"
	"github.com/ory/x/sqlxx"
)

func TestRegistrationAccountEnumerationMitigation(t *testing.T) {
	ctx := context.Background()
	_, reg := internal.NewFastRegistryWithMocks(t, configx.WithValues(map[string]interface{}{
		config.ViperKeySecurityAccountEnumerationMitigate: true,
		config.ViperKeyDefaultIdentitySchemaID:            "default",
		config.ViperKeyIdentitySchemas:                    config.Schemas{{ID: "default", URL: "file://./stub/email.schema.json"}},
		fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypePassword.String()): true,
		config.ViperKeySelfServiceBrowserDefaultReturnTo:                                                                  "https://www.ory.sh",
		config.ViperKeyURLsAllowedReturnToDomains:                                                                         []string{"https://www.ory.sh"},
	}))

	_ = testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewVerificationUIFlowEchoServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	existingEmail := testhelpers.RandomEmail()
	existingIdentity := &identity.Identity{
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{existingEmail},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
			},
		},
		Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, existingEmail)),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
		State:    identity.StateActive,
	}
	require.NoError(t, reg.IdentityManager().Create(ctx, existingIdentity, identity.ManagerAllowWriteProtectedTraits))
	require.NotNil(t, existingIdentity)

	t.Run("case=duplicate registration does not return error with account enumeration mitigation", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, public, false, false, false)

		body, err := json.Marshal(f)
		require.NoError(t, err)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("traits.email", existingEmail)
		values.Set("password", "some-password-123")
		values.Set("method", "password")

		actual, resp := testhelpers.RegistrationMakeRequest(t, false, false, f, client, testhelpers.EncodeFormAsJSON(t, false, values))

		require.EqualValues(t, http.StatusOK, resp.StatusCode, actual)

		csrfToken := gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		assert.NotEmpty(t, csrfToken)

		assert.Empty(t, gjson.Get(actual, "ui.messages").Array(), "should not contain error messages")

		message := testhelpers.CourierExpectMessage(ctx, t, reg, existingEmail, "Account registration attempt")
		assert.Contains(t, message.Subject, "Account registration attempt")
		assert.Contains(t, message.Body, "Someone tried to create an account with this email address.")
	})

	t.Run("case=duplicate registration sends duplicate email with account enumeration mitigation", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, public, false, false, false)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("traits.email", existingEmail)
		values.Set("password", "some-password-123")
		values.Set("method", "password")

		actual, resp := testhelpers.RegistrationMakeRequest(t, false, false, f, client, testhelpers.EncodeFormAsJSON(t, false, values))

		require.EqualValues(t, http.StatusOK, resp.StatusCode, actual)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, existingEmail, "Account registration attempt")
		assert.Contains(t, message.Subject, "Account registration attempt")
		assert.Contains(t, message.Body, "Someone tried to create an account with this email address.")
	})

	t.Run("case=API flow does not return identity with account enumeration mitigation", func(t *testing.T) {
		client := &http.Client{}
		f := testhelpers.InitializeRegistrationFlowViaAPI(t, client, public)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("traits.email", existingEmail)
		values.Set("password", "some-password-123")
		values.Set("method", "password")

		actual, resp := testhelpers.RegistrationMakeRequest(t, true, false, f, client, testhelpers.EncodeFormAsJSON(t, true, values))

		require.EqualValues(t, http.StatusOK, resp.StatusCode, actual)

		assert.Empty(t, gjson.Get(actual, "identity").String(), "identity should not be returned")
	})

	t.Run("case=session not created during duplicate registration", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, public, false, false, false)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("traits.email", existingEmail)
		values.Set("password", "some-password-123")
		values.Set("method", "password")

		actual, resp := testhelpers.RegistrationMakeRequest(t, false, false, f, client, testhelpers.EncodeFormAsJSON(t, false, values))

		require.EqualValues(t, http.StatusOK, resp.StatusCode, actual)

		session, _, err := testhelpers.NewSDKCustomClient(public, client).FrontendAPI.ToSession(ctx).Execute()
		require.Error(t, err)
		assert.Nil(t, session)
	})
}
