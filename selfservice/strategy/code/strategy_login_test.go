// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"
)

func TestLoginCodeStrategy(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), false)
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.login_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), true)
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh"})

	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	createIdentity := func(t *testing.T, moreIdentifiers ...string) *identity.Identity {
		t.Helper()
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		email := testhelpers.RandomEmail()

		ids := fmt.Sprintf(`"email":"%s"`, email)
		for i, identifier := range moreIdentifiers {
			ids = fmt.Sprintf(`%s,"email_%d":"%s"`, ids, i+1, identifier)
		}

		i.Traits = identity.Traits(fmt.Sprintf(`{%s}`, ids))

		credentials := map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {Identifiers: append([]string{email}, moreIdentifiers...), Type: identity.CredentialsTypePassword, Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\"}")},
			identity.CredentialsTypeOIDC:     {Type: identity.CredentialsTypeOIDC, Identifiers: append([]string{email}, moreIdentifiers...), Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\"}")},
			identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Identifiers: append([]string{email}, moreIdentifiers...), Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\", \"user_handle\": \"rVIFaWRcTTuQLkXFmQWpgA==\"}")},
			identity.CredentialsTypeCodeAuth: {Type: identity.CredentialsTypeCodeAuth, Identifiers: append([]string{email}, moreIdentifiers...), Config: sqlxx.JSONRawMessage("{\"address_type\": \"email\", \"used_at\": \"2023-07-26T16:59:06+02:00\"}")},
		}
		i.Credentials = credentials

		var va []identity.VerifiableAddress
		for _, identifier := range moreIdentifiers {
			va = append(va, identity.VerifiableAddress{Value: identifier, Verified: false, Status: identity.VerifiableAddressStatusCompleted})
		}

		va = append(va, identity.VerifiableAddress{Value: email, Verified: true, Status: identity.VerifiableAddressStatusCompleted})

		i.VerifiableAddresses = va

		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, i))
		return i
	}

	type state struct {
		flowID        string
		csrfToken     string
		identity      *identity.Identity
		client        *http.Client
		loginCode     string
		identityEmail string
	}

	createLoginFlow := func(t *testing.T, moreIdentifiers ...string) *state {
		t.Helper()

		identity := createIdentity(t, moreIdentifiers...)
		client := testhelpers.NewClientWithCookies(t)

		// 1. Initiate flow
		resp, err := client.Get(public.URL + login.RouteInitBrowserFlow)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		flowID := gjson.GetBytes(body, "id").String()
		require.NotEmpty(t, flowID)

		csrfToken := gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		require.NotEmpty(t, csrfToken)

		require.NoError(t, resp.Body.Close())

		loginEmail := gjson.Get(identity.Traits.String(), "email").String()
		require.NotEmpty(t, loginEmail)

		return &state{
			flowID:        flowID,
			csrfToken:     csrfToken,
			identity:      identity,
			identityEmail: loginEmail,
			client:        client,
		}
	}

	type onSubmitAssertion func(t *testing.T, s *state, res *http.Response)

	submitLoginID := func(t *testing.T, s *state, submitAssertion onSubmitAssertion) *state {
		t.Helper()

		payload := strings.NewReader(url.Values{
			"csrf_token": {s.csrfToken},
			"method":     {"code"},
			"identifier": {s.identityEmail},
		}.Encode())

		req, err := http.NewRequestWithContext(ctx, "POST", public.URL+login.RouteSubmitFlow+"?flow="+s.flowID, payload)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := s.client.Do(req)
		require.NoError(t, err)

		if submitAssertion != nil {
			submitAssertion(t, s, resp)
			return s
		}

		require.EqualValues(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		csrfToken := gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		require.NotEmpty(t, csrfToken)

		s.csrfToken = csrfToken

		require.NoError(t, resp.Body.Close())

		return s
	}

	submitLoginCode := func(t *testing.T, s *state, submitAssertion onSubmitAssertion) *state {
		t.Helper()

		req, err := http.NewRequestWithContext(ctx, "POST", public.URL+login.RouteSubmitFlow+"?flow="+s.flowID, strings.NewReader(url.Values{
			"csrf_token": {s.csrfToken},
			"method":     {"code"},
			"code":       {s.loginCode},
			"identifier": {s.identityEmail},
		}.Encode()))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := s.client.Do(req)
		require.NoError(t, err)

		if submitAssertion != nil {
			submitAssertion(t, s, resp)
			return s
		}

		var cookie *http.Cookie
		for _, c := range resp.Cookies() {
			cookie = c
		}
		require.Equal(t, cookie.Name, "ory_kratos_session")
		require.NotEmpty(t, cookie.Value)

		return s
	}

	t.Run("case=should be able to log in with code", func(t *testing.T) {
		// create login flow
		s := createLoginFlow(t)

		// submit email
		s = submitLoginID(t, s, nil)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
		assert.Contains(t, message.Body, "please login to your account by entering the following code")

		loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
		assert.NotEmpty(t, loginCode)

		s.loginCode = loginCode

		// 3. Submit OTP
		submitLoginCode(t, s, nil)
	})

	t.Run("case=should not be able to change submitted id on code submit", func(t *testing.T) {
		// create login flow
		s := createLoginFlow(t)

		// submit email
		s = submitLoginID(t, s, nil)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
		assert.Contains(t, message.Body, "please login to your account by entering the following code")

		loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
		assert.NotEmpty(t, loginCode)

		s.loginCode = loginCode
		s.identityEmail = "not-" + s.identityEmail

		// 3. Submit OTP
		s = submitLoginCode(t, s, func(t *testing.T, s *state, resp *http.Response) {
			require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
			body := ioutilx.MustReadAll(resp.Body)
			assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "account does not exist or has not setup sign in with code")
		})
	})

	t.Run("case=should not be able to proceed to code entry when the account is unknown", func(t *testing.T) {
		s := createLoginFlow(t)

		s.identityEmail = testhelpers.RandomEmail()

		// submit email
		s = submitLoginID(t, s, func(t *testing.T, s *state, resp *http.Response) {
			require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
			body := ioutilx.MustReadAll(resp.Body)
			assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "account does not exist or has not setup sign in with code")
		})
	})

	t.Run("case=should not be able to use valid code after 5 attempts", func(t *testing.T) {
		s := createLoginFlow(t)

		// submit email
		s = submitLoginID(t, s, nil)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
		assert.Contains(t, message.Body, "please login to your account by entering the following code")
		loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
		assert.NotEmpty(t, loginCode)

		for i := 0; i < 5; i++ {

			s.loginCode = "111111"

			// 3. Submit OTP
			s = submitLoginCode(t, s, func(t *testing.T, s *state, resp *http.Response) {
				require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
				body := ioutilx.MustReadAll(resp.Body)
				assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "The login code is invalid or has already been used")
			})
		}

		s.loginCode = loginCode
		// 3. Submit OTP
		s = submitLoginCode(t, s, func(t *testing.T, s *state, resp *http.Response) {
			require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
			body := ioutilx.MustReadAll(resp.Body)
			assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "The request was submitted too often.")
		})
	})

	t.Run("case=code should expire", func(t *testing.T) {
		ctx := context.Background()

		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "10ns")

		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "1h")
		})

		s := createLoginFlow(t)

		// submit email
		s = submitLoginID(t, s, nil)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
		assert.Contains(t, message.Body, "please login to your account by entering the following code")
		loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
		assert.NotEmpty(t, loginCode)

		s.loginCode = loginCode

		submitLoginCode(t, s, func(t *testing.T, s *state, resp *http.Response) {
			require.EqualValues(t, http.StatusGone, resp.StatusCode)
			body := ioutilx.MustReadAll(resp.Body)
			require.Contains(t, gjson.GetBytes(body, "error.reason").String(), "self-service flow expired 0.00 minutes ago")
		})
	})

	t.Run("case=on login with un-verified address, should verify it", func(t *testing.T) {
		s := createLoginFlow(t, testhelpers.RandomEmail())

		loginEmail := gjson.Get(s.identity.Traits.String(), "email_1").String()
		require.NotEmpty(t, loginEmail)

		s.identityEmail = loginEmail

		var va *identity.VerifiableAddress

		for _, v := range s.identity.VerifiableAddresses {
			if v.Value == loginEmail {
				va = &v
				break
			}
		}

		require.NotNil(t, va)
		require.False(t, va.Verified)

		// submit email
		s = submitLoginID(t, s, nil)

		message := testhelpers.CourierExpectMessage(ctx, t, reg, loginEmail, "Login to your account")
		require.Contains(t, message.Body, "please login to your account by entering the following code")

		loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
		require.NotEmpty(t, loginCode)

		s.loginCode = loginCode

		// Submit OTP
		s = submitLoginCode(t, s, nil)

		id, err := reg.PrivilegedIdentityPool().GetIdentity(ctx, s.identity.ID, identity.ExpandEverything)
		require.NoError(t, err)

		va = nil

		for _, v := range id.VerifiableAddresses {
			if v.Value == loginEmail {
				va = &v
				break
			}
		}

		require.NotNil(t, va)
		require.True(t, va.Verified)
	})
}
