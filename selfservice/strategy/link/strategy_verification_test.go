// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/ui/node"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"
)

func TestVerification(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)

	identityToVerify := &identity.Identity{
		ID:       x.NewUUID(),
		Traits:   identity.Traits(`{"email":"verifyme@ory.sh"}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{"recoverme@ory.sh"}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)},
		},
	}

	verificationEmail := gjson.GetBytes(identityToVerify.Traits, "email").String()

	_ = testhelpers.NewVerificationUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "returned", conf)

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)

	require.NoError(t, reg.IdentityManager().Create(context.Background(), identityToVerify,
		identity.ManagerAllowWriteProtectedTraits))

	expect := func(t *testing.T, hc *http.Client, isAPI, isSPA bool, values func(url.Values), c int) string {
		if hc == nil {
			hc = testhelpers.NewDebugClient(t)
			if !isAPI {
				hc = testhelpers.NewClientWithCookies(t)
				hc.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
			}
		}

		return testhelpers.SubmitVerificationForm(t, isAPI, isSPA, hc, public, values, c,
			testhelpers.ExpectURL(isAPI || isSPA, public.URL+verification.RouteSubmitFlow, conf.SelfServiceFlowVerificationUI(ctx).String()))
	}

	expectValidationError := func(t *testing.T, hc *http.Client, isAPI, isSPA bool, values func(url.Values)) string {
		return expect(t, hc, isAPI, isSPA, values, testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK))
	}

	expectSuccess := func(t *testing.T, hc *http.Client, isAPI, isSPA bool, values func(url.Values)) string {
		return expect(t, hc, isAPI, isSPA, values, http.StatusOK)
	}

	t.Run("description=should set all the correct verification payloads after submission", func(t *testing.T) {
		body := expectSuccess(t, nil, false, false, func(v url.Values) {
			v.Set("email", "test@ory.sh")
		})
		testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes").String()), []string{"0.attributes.value"})
	})

	t.Run("description=should set all the correct verification payloads", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetVerificationFlow(t, c, public)

		testhelpers.SnapshotTExcept(t, rs.Ui.Nodes, []string{"0.attributes.value"})
		assert.EqualValues(t, public.URL+verification.RouteSubmitFlow+"?flow="+rs.Id, rs.Ui.Action)
		assert.Empty(t, rs.Ui.Messages)
	})

	t.Run("description=should not execute submit without correct method set", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetVerificationFlow(t, c, public)

		res, err := c.PostForm(rs.Ui.Action, url.Values{"method": {"not-link"}, "email": {verificationEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI(ctx).String())

		body := ioutilx.MustReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, "Could not find a strategy to verify your account with. Did you fill out the form correctly?", gjson.GetBytes(body, "ui.messages.0.text").String(), "%s", body)
	})

	t.Run("description=should require an email to be sent", func(t *testing.T) {
		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.LinkGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, "Property email is missing.",
				gjson.Get(actual, "ui.nodes.#(attributes.name==email).messages.0.text").String(),
				"%s", actual)
		}

		values := func(v url.Values) {
			v.Del("email")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, nil, false, false, values))
		})

		t.Run("type=spa", func(t *testing.T) {
			check(t, expectValidationError(t, nil, false, true, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, nil, true, false, values))
		})
	})

	t.Run("description=should require a valid email to be sent", func(t *testing.T) {
		check := func(t *testing.T, actual string, value string) {
			assert.EqualValues(t, string(node.LinkGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, fmt.Sprintf("%q is not valid \"email\"", value),
				gjson.Get(actual, "ui.nodes.#(attributes.name==email).messages.0.text").String(),
				"%s", actual)
		}

		for _, email := range []string{"\\", "asdf", "...", "aiacobelli.sec@gmail.com,alejandro.iacobelli@mercadolibre.com"} {
			values := func(v url.Values) {
				v.Set("email", email)
			}

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, nil, false, false, values), email)
			})

			t.Run("type=spa", func(t *testing.T) {
				check(t, expectValidationError(t, nil, false, true, values), email)
			})

			t.Run("type=api", func(t *testing.T) {
				check(t, expectValidationError(t, nil, true, false, values), email)
			})
		}
	})

	t.Run("description=should try to verify an email that does not exist", func(t *testing.T) {
		conf.Set(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, true)

		t.Cleanup(func() {
			conf.Set(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, false)
		})
		var email string
		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.LinkGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationEmailSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Someone tried to verify this email address")
			assert.Contains(t, message.Body, "If this was you, check if you signed up using a different address.")
		}

		values := func(v url.Values) {
			v.Set("email", email)
		}

		t.Run("type=browser", func(t *testing.T) {
			email = x.NewUUID().String() + "@ory.sh"
			check(t, expectSuccess(t, nil, false, false, values))
		})

		t.Run("type=spa", func(t *testing.T) {
			email = x.NewUUID().String() + "@ory.sh"
			check(t, expectSuccess(t, nil, false, true, values))
		})

		t.Run("type=api", func(t *testing.T) {
			email = x.NewUUID().String() + "@ory.sh"
			check(t, expectSuccess(t, nil, true, false, values))
		})
	})

	t.Run("description=should not be able to use an invalid link", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeVerificationFlowViaBrowser(t, c, false, public)
		res, err := c.Get(public.URL + verification.RouteSubmitFlow + "?flow=" + f.Id + "&token=i-do-not-exist")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI(ctx).String()+"?flow=")

		sr, _, err := testhelpers.NewSDKCustomClient(public, c).FrontendApi.GetVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, sr.Ui.Messages, 1)
		assert.Equal(t, "The verification token is invalid or has already been used. Please retry the flow.", sr.Ui.Messages[0].Text)
	})

	t.Run("description=should not be able to request link with an outdated flow", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetVerificationFlow(t, c, public)

		time.Sleep(time.Millisecond * 201)

		res, err := c.PostForm(rs.Ui.Action, url.Values{"method": {"link"}, "email": {verificationEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Id)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI(ctx).String())
	})

	t.Run("description=should not be able to use link with an outdated flow", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		body := expectSuccess(t, c, false, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		assert.Contains(t, message.Body, "Hi, please verify your account by clicking the following link")

		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		time.Sleep(time.Millisecond * 201)

		// Clear cookies as link might be opened in another browser
		c = testhelpers.NewClientWithCookies(t)
		res, err := c.Get(verificationLink)
		require.NoError(t, err)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI(ctx).String())
		assert.NotContains(t, res.Request.URL.String(), gjson.Get(body, "id").String())

		sr, _, err := testhelpers.NewSDKCustomClient(public, c).FrontendApi.GetVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, sr.Ui.Messages, 1)
		assert.Contains(t, sr.Ui.Messages[0].Text, "The verification flow expired")
	})

	t.Run("description=should verify an email address", func(t *testing.T) {
		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.LinkGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, verificationEmail, gjson.Get(actual, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationEmailSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
			assert.Contains(t, message.Body, "please verify your account by clicking the following link")

			verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			assert.Contains(t, verificationLink, public.URL+verification.RouteSubmitFlow)
			assert.Contains(t, verificationLink, "token=")

			cl := testhelpers.NewClientWithCookies(t)
			res, err := cl.Get(verificationLink)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI(ctx).String())
			body := string(ioutilx.MustReadAll(res.Body))
			assert.EqualValues(t, "passed_challenge", gjson.Get(body, "state").String())
			assert.EqualValues(t, text.NewInfoSelfServiceVerificationSuccessful().Text, gjson.Get(body, "ui.messages.0.text").String())

			id, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), identityToVerify.ID)
			require.NoError(t, err)
			require.Len(t, id.VerifiableAddresses, 1)

			address := id.VerifiableAddresses[0]
			assert.EqualValues(t, verificationEmail, address.Value)
			assert.True(t, address.Verified)
			assert.EqualValues(t, identity.VerifiableAddressStatusCompleted, address.Status)
			assert.True(t, time.Time(*address.VerifiedAt).Add(time.Second*5).After(time.Now()))
		}

		values := func(v url.Values) {
			v.Set("email", verificationEmail)
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectSuccess(t, nil, false, false, values))
		})

		t.Run("type=spa", func(t *testing.T) {
			check(t, expectSuccess(t, nil, false, true, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectSuccess(t, nil, true, false, values))
		})
	})

	t.Run("description=should verify an email address when the link is opened in another browser", func(t *testing.T) {
		check := func(t *testing.T, actual string) {
			message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
			verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			cl := testhelpers.NewClientWithCookies(t)
			res, err := cl.Get(verificationLink)
			require.NoError(t, err)
			body := string(ioutilx.MustReadAll(res.Body))
			require.NoError(t, res.Body.Close())
			require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
			assert.Contains(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL))[0].Name, x.CSRFTokenName)

			actualRes, err := cl.Get(public.URL + verification.RouteGetFlow + "?id=" + gjson.Get(body, "id").String())
			require.NoError(t, err)
			actualBody := string(ioutilx.MustReadAll(actualRes.Body))
			require.NoError(t, actualRes.Body.Close())
			assert.Equal(t, http.StatusOK, actualRes.StatusCode)

			assertx.EqualAsJSON(t, body, actualBody)
			assert.EqualValues(t, "passed_challenge", gjson.Get(actualBody, "state").String())
		}

		values := func(v url.Values) {
			v.Set("email", verificationEmail)
		}

		check(t, expectSuccess(t, nil, false, false, values))
	})

	newValidFlow := func(t *testing.T, fType flow.Type, requestURL string) (*verification.Flow, *link.VerificationToken) {
		f, err := verification.NewFlow(conf, time.Hour, x.FakeCSRFToken, httptest.NewRequest("GET", requestURL, nil), nil, fType)
		require.NoError(t, err)
		f.State = flow.StateEmailSent
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(context.Background(), f))
		email := identity.NewVerifiableEmailAddress(verificationEmail, identityToVerify.ID)
		identityToVerify.VerifiableAddresses = append(identityToVerify.VerifiableAddresses, *email)
		require.NoError(t, reg.IdentityManager().Update(context.Background(), identityToVerify, identity.ManagerAllowWriteProtectedTraits))

		token := link.NewSelfServiceVerificationToken(&identityToVerify.VerifiableAddresses[0], f, time.Hour)
		require.NoError(t, reg.VerificationTokenPersister().CreateVerificationToken(context.Background(), token))
		return f, token
	}

	newValidBrowserFlow := func(t *testing.T, requestURL string) (*verification.Flow, *link.VerificationToken) {
		return newValidFlow(t, flow.TypeBrowser, requestURL)
	}

	t.Run("case=respects return_to URI parameter", func(t *testing.T) {
		returnToURL := public.URL + "/after-verification"
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToURL})

		for _, fType := range []flow.Type{flow.TypeBrowser, flow.TypeAPI} {
			t.Run(fmt.Sprintf("type=%s", fType), func(t *testing.T) {
				client := testhelpers.NewClientWithCookies(t)
				flow, token := newValidFlow(t, fType, public.URL+verification.RouteInitBrowserFlow+"?"+url.Values{"return_to": {returnToURL}}.Encode())

				res, err := client.Get(public.URL + verification.RouteSubmitFlow + "?" + url.Values{"flow": {flow.ID.String()}, "token": {token.Token}}.Encode())
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, res.StatusCode)
				responseBody := gjson.ParseBytes(ioutilx.MustReadAll(res.Body))

				assert.Equal(t, responseBody.Get("state").String(), "passed_challenge", "%v", responseBody)
				assert.True(t, responseBody.Get("ui.nodes.#(attributes.id==continue)").Exists(), "%v", responseBody)
				assert.Equal(t, returnToURL, responseBody.Get("ui.nodes.#(attributes.id==continue).attributes.href").String(), "%v", responseBody)
			})
		}
	})

	t.Run("case=contains default return to url", func(t *testing.T) {
		globalReturnTo := public.URL + "/global"
		conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, globalReturnTo)

		for _, fType := range []flow.Type{flow.TypeBrowser, flow.TypeAPI} {
			t.Run(fmt.Sprintf("type=%s", fType), func(t *testing.T) {
				client := testhelpers.NewClientWithCookies(t)
				flow, token := newValidFlow(t, fType, public.URL+verification.RouteInitBrowserFlow)

				res, err := client.Get(public.URL + verification.RouteSubmitFlow + "?" + url.Values{"flow": {flow.ID.String()}, "token": {token.Token}}.Encode())
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, res.StatusCode)
				responseBody := gjson.ParseBytes(ioutilx.MustReadAll(res.Body))

				assert.Equal(t, responseBody.Get("state").String(), "passed_challenge", "%v", responseBody)
				assert.True(t, responseBody.Get("ui.nodes.#(attributes.id==continue)").Exists(), "%v", responseBody)
				assert.Equal(t, globalReturnTo, responseBody.Get("ui.nodes.#(attributes.id==continue).attributes.href").String(), "%v", responseBody)
			})
		}
	})

	t.Run("case=should not be able to use code from different flow", func(t *testing.T) {
		f1, _ := newValidBrowserFlow(t, public.URL+verification.RouteInitBrowserFlow)

		_, t2 := newValidBrowserFlow(t, public.URL+verification.RouteInitBrowserFlow)

		formValues := url.Values{
			"flow":  {f1.ID.String()},
			"token": {t2.Token},
		}
		submitUrl := public.URL + verification.RouteSubmitFlow + "?" + formValues.Encode()

		res, err := public.Client().Get(submitUrl)
		require.NoError(t, err)
		body := ioutilx.MustReadAll(res.Body)

		assert.Equal(t, "The verification token is invalid or has already been used. Please retry the flow.", gjson.GetBytes(body, "ui.messages.0.text").String())
	})
}
