// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package totp_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/pquerna/otp"
	stdtotp "github.com/pquerna/otp/totp"

	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/selfservice/strategy/totp"
	"github.com/ory/kratos/ui/node"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/text"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
)

func TestCompleteSettings(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": false})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".profile", map[string]interface{}{"enabled": false})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeTOTP), map[string]interface{}{"enabled": true})
	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewRedirSessionEchoTS(t, reg)
	loginTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)

	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1m")

	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/settings.schema.json")
	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	t.Run("case=device unlinking is available when identity has totp", func(t *testing.T) {
		id, _, _ := createIdentity(t, reg)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)
		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value",
		})
	})

	t.Run("case=device setup is available when identity has no totp yet", func(t *testing.T) {
		id, _, _ := createIdentity(t, reg)
		id.Credentials = nil
		require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(context.Background(), id))

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)
		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value",
			"1.attributes.src",
			"2.attributes.text.context.secret",
			"2.attributes.text.text",
		})
	})

	doAPIFlow := func(t *testing.T, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", "totp")
		v(values)
		payload := testhelpers.EncodeFormAsJSON(t, true, values)
		return testhelpers.SettingsMakeRequest(t, true, false, f, apiClient, payload)
	}

	doBrowserFlow := func(t *testing.T, spa bool, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, spa, publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", "totp")
		v(values)
		return testhelpers.SettingsMakeRequest(t, false, spa, f, browserClient, testhelpers.EncodeFormAsJSON(t, spa, values))
	}

	t.Run("case=should pass without csrf if API flow", func(t *testing.T) {
		id := createIdentityWithoutTOTP(t, reg)

		body, res := doAPIFlow(t, func(v url.Values) {
			v.Del("csrf_token")
			v.Set(node.TOTPCode, "111111")
		}, id)

		assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationTOTPVerifierWrong().Text, gjson.Get(body, totpCodeGJSONQuery+".messages.0.text").String(), "%s", body)
	})

	t.Run("case=should fail if CSRF token is invalid", func(t *testing.T) {
		id := createIdentityWithoutTOTP(t, reg)

		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, func(v url.Values) {
				v.Del("csrf_token")
				v.Set(node.TOTPCode, "111111")
			}, id)

			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "reason").String(), body)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, func(v url.Values) {
				v.Del("csrf_token")
				v.Set(node.TOTPCode, "111111")
			}, id)

			assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "error.reason").String(), body)
		})
	})

	t.Run("type=can not unlink without privileged session", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
		})

		id, _, key := createIdentity(t, reg)
		payload := func(v url.Values) {
			v.Set("totp_unlink", "true")
		}

		checkIdentity := func(t *testing.T) {
			_, cred, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeTOTP, id.ID.String())
			require.NoError(t, err)
			assert.Equal(t, key.URL(), gjson.GetBytes(cred.Config, "totp_url").String())
		}

		t.Run("type=api", func(t *testing.T) {
			actual, res := doAPIFlow(t, payload, id)
			assert.Equal(t, http.StatusForbidden, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "redirect_browser_to").String(), publicTS.URL+"/self-service/login/browser?refresh=true&return_to=")
			assertx.EqualAsJSONExcept(t, settings.NewFlowNeedsReAuth(), json.RawMessage(actual), []string{"redirect_browser_to"})
			checkIdentity(t)
		})

		t.Run("type=spa", func(t *testing.T) {
			actual, res := doBrowserFlow(t, true, payload, id)
			assert.Equal(t, http.StatusForbidden, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "redirect_browser_to").String(), publicTS.URL+"/self-service/login/browser?refresh=true&return_to=")
			assertx.EqualAsJSONExcept(t, settings.NewFlowNeedsReAuth(), json.RawMessage(actual), []string{"redirect_browser_to"})
			checkIdentity(t)
		})

		t.Run("type=browser", func(t *testing.T) {
			actual, res := doBrowserFlow(t, false, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), loginTS.URL+"/login-ts")
			assertx.EqualAsJSON(t, text.NewInfoLoginReAuth().Text, json.RawMessage(gjson.Get(actual, "ui.messages.0.text").Raw), actual)
			checkIdentity(t)
		})
	})

	t.Run("type=can not set up new totp device without privileged session", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
		})

		id := createIdentityWithoutTOTP(t, reg)
		payload := func(v url.Values) {
			v.Set(node.TOTPCode, "111111")
		}

		checkIdentity := func(t *testing.T) {
			_, _, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeTOTP, id.ID.String())
			require.ErrorIs(t, err, sqlcon.ErrNoRows)
		}

		t.Run("type=api", func(t *testing.T) {
			actual, res := doAPIFlow(t, payload, id)
			assert.Equal(t, http.StatusForbidden, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "redirect_browser_to").String(), publicTS.URL+"/self-service/login/browser?refresh=true&return_to=")
			assertx.EqualAsJSONExcept(t, settings.NewFlowNeedsReAuth(), json.RawMessage(actual), []string{"redirect_browser_to"})
			checkIdentity(t)
		})

		t.Run("type=spa", func(t *testing.T) {
			actual, res := doBrowserFlow(t, true, payload, id)
			assert.Equal(t, http.StatusForbidden, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "redirect_browser_to").String(), publicTS.URL+"/self-service/login/browser?refresh=true&return_to=")
			assertx.EqualAsJSONExcept(t, settings.NewFlowNeedsReAuth(), json.RawMessage(actual), []string{"redirect_browser_to"})
			checkIdentity(t)
		})

		t.Run("type=browser", func(t *testing.T) {
			actual, res := doBrowserFlow(t, false, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), loginTS.URL+"/login-ts")
			assertx.EqualAsJSON(t, text.NewInfoLoginReAuth().Text, json.RawMessage(gjson.Get(actual, "ui.messages.0.text").Raw), actual)
			checkIdentity(t)
		})
	})

	t.Run("type=unlink TOTP device", func(t *testing.T) {
		payload := func(v url.Values) {
			v.Set("totp_unlink", "true")
		}

		checkIdentity := func(t *testing.T, id *identity.Identity) {
			_, _, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeTOTP, id.ID.String())
			require.ErrorIs(t, err, sqlcon.ErrNoRows)
		}

		t.Run("type=api", func(t *testing.T) {
			id, _, _ := createIdentity(t, reg)
			actual, res := doAPIFlow(t, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			assert.EqualValues(t, flow.StateSuccess, gjson.Get(actual, "state").String(), actual)
			checkIdentity(t, id)
		})

		t.Run("type=spa", func(t *testing.T) {
			id, _, _ := createIdentity(t, reg)
			actual, res := doBrowserFlow(t, true, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			assert.EqualValues(t, flow.StateSuccess, gjson.Get(actual, "state").String(), actual)
			checkIdentity(t, id)
		})

		t.Run("type=browser", func(t *testing.T) {
			id, _, _ := createIdentity(t, reg)
			actual, res := doBrowserFlow(t, false, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			assert.EqualValues(t, flow.StateSuccess, gjson.Get(actual, "state").String(), actual)
			checkIdentity(t, id)
		})
	})

	t.Run("type=set up TOTP device but code is incorrect", func(t *testing.T) {
		payload := func(v url.Values) {
			v.Set(node.TOTPCode, "111111")
		}

		checkIdentity := func(t *testing.T, id *identity.Identity) {
			_, _, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeTOTP, id.ID.String())
			require.ErrorIs(t, err, sqlcon.ErrNoRows)
		}

		t.Run("type=api", func(t *testing.T) {
			id := createIdentityWithoutTOTP(t, reg)
			actual, res := doAPIFlow(t, payload, id)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			assert.Equal(t, text.NewErrorValidationTOTPVerifierWrong().Text, gjson.Get(actual, totpCodeGJSONQuery+".messages.0.text").String(), actual)
			checkIdentity(t, id)
		})

		t.Run("type=spa", func(t *testing.T) {
			id := createIdentityWithoutTOTP(t, reg)
			actual, res := doBrowserFlow(t, true, payload, id)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			assert.Equal(t, text.NewErrorValidationTOTPVerifierWrong().Text, gjson.Get(actual, totpCodeGJSONQuery+".messages.0.text").String(), actual)
			checkIdentity(t, id)
		})

		t.Run("type=browser", func(t *testing.T) {
			id := createIdentityWithoutTOTP(t, reg)
			actual, res := doBrowserFlow(t, false, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			assert.Equal(t, text.NewErrorValidationTOTPVerifierWrong().Text, gjson.Get(actual, totpCodeGJSONQuery+".messages.0.text").String(), actual)
			checkIdentity(t, id)
		})
	})

	t.Run("type=set up TOTP device", func(t *testing.T) {
		checkIdentity := func(t *testing.T, id *identity.Identity, key string) {
			i, cred, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeTOTP, id.ID.String())
			require.NoError(t, err)
			var c identity.CredentialsTOTPConfig
			require.NoError(t, json.Unmarshal(cred.Config, &c))
			actual, err := otp.NewKeyFromURL(c.TOTPURL)
			require.NoError(t, err)
			assert.Equal(t, key, actual.Secret())
			assert.Contains(t, c.TOTPURL, gjson.GetBytes(i.Traits, "subject").String())
		}

		run := func(t *testing.T, isAPI, isSPA bool, id *identity.Identity, hc *http.Client, f *kratos.SettingsFlow) {
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

			nodes, err := json.Marshal(f.Ui.Nodes)
			require.NoError(t, err)

			key := gjson.GetBytes(nodes, "#(attributes.id==totp_secret_key).attributes.text.context.secret").String()
			require.NotEmpty(t, key, nodes)

			code, err := stdtotp.GenerateCode(key, time.Now())
			require.NoError(t, err)
			values.Set("method", "totp")
			values.Set(node.TOTPCode, code)

			actual, res := testhelpers.SettingsMakeRequest(t, isAPI, isSPA, f, hc, testhelpers.EncodeFormAsJSON(t, isAPI || isSPA, values))
			require.NotEmpty(t, key)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			if isAPI || isSPA {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
				assert.EqualValues(t, flow.StateSuccess, gjson.Get(actual, "state").String(), actual)
			} else {
				assert.Contains(t, res.Request.URL.String(), uiTS.URL)
				assert.EqualValues(t, flow.StateSuccess, gjson.Get(actual, "state").String(), actual)
			}

			actualFlow, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(f.Id))
			require.NoError(t, err)
			assert.Empty(t, gjson.GetBytes(actualFlow.InternalContext, flow.PrefixInternalContextKey(identity.CredentialsTypeTOTP, totp.InternalContextKeyURL)))

			checkIdentity(t, id, key)
			testhelpers.EnsureAAL(t, hc, publicTS, "aal2", string(identity.CredentialsTypeTOTP))
		}

		t.Run("type=api", func(t *testing.T) {
			id := createIdentityWithoutTOTP(t, reg)

			apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
			f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)

			run(t, true, false, id, apiClient, f)
		})

		t.Run("type=spa", func(t *testing.T) {
			id := createIdentityWithoutTOTP(t, reg)

			user := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, user, true, publicTS)

			run(t, false, true, id, user, f)
		})

		t.Run("type=browser", func(t *testing.T) {
			id := createIdentityWithoutTOTP(t, reg)

			user := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, user, false, publicTS)

			run(t, false, false, id, user, f)
		})
	})
}
