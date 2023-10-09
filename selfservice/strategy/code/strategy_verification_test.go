// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/strategy/code"
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
	initViper(t, ctx, conf)

	identityToVerify := &identity.Identity{
		ID:       x.NewUUID(),
		Traits:   identity.Traits(`{"email":"verifyme@ory.sh"}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {
				Type:        "password",
				Identifiers: []string{"recoverme@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`),
			},
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

	submitVerificationCode := func(t *testing.T, body string, c *http.Client, code string) (string, *http.Response) {
		action := gjson.Get(body, "ui.action").String()
		require.NotEmpty(t, action, "%v", string(body))
		csrfToken := extractCsrfToken([]byte(body))

		res, err := c.PostForm(action, url.Values{
			"code":       {code},
			"csrf_token": {csrfToken},
		})
		require.NoError(t, err)

		return string(ioutilx.MustReadAll(res.Body)), res
	}

	t.Run("description=should set all the correct verification payloads after submission", func(t *testing.T) {
		body := expectSuccess(t, nil, false, false, func(v url.Values) {
			v.Set("email", "test@ory.sh")
		})
		testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes").String()), []string{"3.attributes.value"})
	})

	t.Run("description=should set all the correct verification payloads", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetVerificationFlow(t, c, public)

		testhelpers.SnapshotTExcept(t, rs.Ui.Nodes, []string{"2.attributes.value"})
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
		testhelpers.AssertFieldMessage(t, body, "method", `value must be one of "code", "link"`)
	})

	t.Run("description=should require an email to be sent", func(t *testing.T) {
		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.CodeGroup), gjson.Get(actual, "active").String(), "%s", actual)
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
			assert.EqualValues(t, string(node.CodeGroup), gjson.Get(actual, "active").String(), "%s", actual)
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
			assert.EqualValues(t, string(node.CodeGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationEmailWithCodeSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

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

	t.Run("description=clicking link should prefill code", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		f := testhelpers.SubmitVerificationForm(t, false, false, c, public, func(v url.Values) {
			v.Set("email", verificationEmail)
		}, 200, "")
		fID := gjson.Get(f, "id").String()
		res, err := c.Get(public.URL + verification.RouteSubmitFlow + "?flow=" + fID + "&code=12312312")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI(ctx).String()+"?flow=")

		body := ioutilx.MustReadAll(res.Body)

		assert.Equal(t, "12312312", gjson.GetBytes(body, "ui.nodes.#(attributes.name==code).attributes.value").String(), "%v", string(body))
	})

	t.Run("description=should not be able to use an invalid code", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		f := testhelpers.SubmitVerificationForm(t, false, false, c, public, func(v url.Values) {
			v.Set("email", verificationEmail)
		}, 200, "")

		body, res := submitVerificationCode(t, f, c, "12312312")
		assert.Equal(t, http.StatusOK, res.StatusCode)

		testhelpers.AssertMessage(t, []byte(body), "The verification code is invalid or has already been used. Please try again.")
	})

	t.Run("description=should not be able to submit email in expired flow", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Millisecond*10)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetVerificationFlow(t, c, public)

		time.Sleep(time.Millisecond * 11)

		res, err := c.PostForm(rs.Ui.Action, url.Values{"method": {"code"}, "email": {verificationEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Id)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI(ctx).String())
		body := ioutilx.MustReadAll(res.Body)
		testhelpers.AssertMessage(t, body, "The verification flow expired 0.00 minutes ago, please try again.")
	})

	t.Run("description=should not be able to submit code in expired flow", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Millisecond*10)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		body := expectSuccess(t, c, false, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		assert.Contains(t, message.Body, "please verify your account by entering the following code")

		code := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		time.Sleep(time.Millisecond * 11)

		f, _ := submitVerificationCode(t, body, c, code)

		testhelpers.AssertMessage(t, []byte(f), "The verification flow expired 0.00 minutes ago, please try again.")
	})

	t.Run("description=should verify an email address", func(t *testing.T) {
		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.CodeGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, verificationEmail, gjson.Get(actual, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationEmailWithCodeSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
			assert.Contains(t, message.Body, "please verify your account by entering the following code")

			verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			assert.Contains(t, verificationLink, public.URL+verification.RouteSubmitFlow)
			assert.Contains(t, verificationLink, "code=")

			cl := testhelpers.NewClientWithCookies(t)
			res, err := cl.Get(verificationLink)
			require.NoError(t, err)
			defer res.Body.Close()

			f := ioutilx.MustReadAll(res.Body)

			code := gjson.GetBytes(f, "ui.nodes.#(attributes.name==code).attributes.value").String()
			require.NotEmpty(t, code)

			body, res := submitVerificationCode(t, string(f), cl, code)

			assert.Equal(t, http.StatusOK, res.StatusCode)
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
		values := func(v url.Values) {
			v.Set("email", verificationEmail)
		}

		expectSuccess(t, nil, false, false, values)
		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)
		code := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.Get(verificationLink)
		require.NoError(t, err)
		body := string(ioutilx.MustReadAll(res.Body))
		require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
		assert.Contains(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL))[0].Name, x.CSRFTokenName)

		actualBody, _ := submitVerificationCode(t, body, cl, code)
		assert.EqualValues(t, "passed_challenge", gjson.Get(actualBody, "state").String())
	})

	newValidFlow := func(t *testing.T, fType flow.Type, requestURL string) (*verification.Flow, *code.VerificationCode, string) {
		f, err := verification.NewFlow(conf, time.Hour, x.FakeCSRFToken, httptest.NewRequest("GET", requestURL, nil), code.NewStrategy(reg), fType)
		require.NoError(t, err)
		f.State = flow.StateEmailSent
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(context.Background(), f))
		email := identity.NewVerifiableEmailAddress(verificationEmail, identityToVerify.ID)
		identityToVerify.VerifiableAddresses = append(identityToVerify.VerifiableAddresses, *email)
		require.NoError(t, reg.IdentityManager().Update(context.Background(), identityToVerify, identity.ManagerAllowWriteProtectedTraits))

		params := &code.CreateVerificationCodeParams{
			RawCode:           "12312312",
			ExpiresIn:         time.Hour,
			VerifiableAddress: &identityToVerify.VerifiableAddresses[0],
			FlowID:            f.ID,
		}
		verificationCode, err := reg.VerificationCodePersister().CreateVerificationCode(context.Background(), params)
		require.NoError(t, err)
		return f, verificationCode, params.RawCode
	}

	newValidBrowserFlow := func(t *testing.T, requestURL string) (*verification.Flow, *code.VerificationCode, string) {
		return newValidFlow(t, flow.TypeBrowser, requestURL)
	}

	t.Run("case=contains link to return_to", func(t *testing.T) {
		returnToURL := public.URL + "/after-verification"
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToURL})
		client := &http.Client{}

		f, _, rawCode := newValidBrowserFlow(t, public.URL+verification.RouteInitBrowserFlow+"?"+url.Values{"return_to": {returnToURL}}.Encode())

		action := public.URL + verification.RouteSubmitFlow + "?flow=" + f.ID.String()

		res, err := client.PostForm(action, url.Values{
			"code":       {rawCode},
			"csrf_token": {x.FakeCSRFToken},
		})
		require.NoError(t, err)
		body := ioutilx.MustReadAll(res.Body)

		assert.Equal(t, returnToURL, gjson.GetBytes(body, "ui.nodes.#(attributes.id==continue).attributes.href").String())
	})

	t.Run("case=should respond with replaced error if successful code is submitted again via api", func(t *testing.T) {
		_ = expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		assert.Contains(t, message.Body, "please verify your account by entering the following code")

		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		assert.Contains(t, verificationLink, public.URL+verification.RouteSubmitFlow)
		assert.Contains(t, verificationLink, "code=")

		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.Get(verificationLink)
		require.NoError(t, err)
		defer res.Body.Close()

		original := ioutilx.MustReadAll(res.Body)

		code := gjson.GetBytes(original, "ui.nodes.#(attributes.name==code).attributes.value").String()
		require.NotEmpty(t, code)
		action := gjson.GetBytes(original, "ui.action").String()
		require.NotEmpty(t, action)

		c := testhelpers.NewDebugClient(t)
		res, err = c.Post(action, "application/json", strings.NewReader(fmt.Sprintf(`{"code": "%v"}`, code)))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		f1 := ioutilx.MustReadAll(res.Body)

		assert.EqualValues(t, "passed_challenge", gjson.GetBytes(f1, "state").String())

		res, err = c.Post(action, "application/json", strings.NewReader(fmt.Sprintf(`{"code": "%v"}`, code)))
		require.NoError(t, err)
		assert.Equal(t, http.StatusGone, res.StatusCode)

		f2 := ioutilx.MustReadAll(res.Body)
		assert.Equal(t, text.ErrIDSelfServiceFlowReplaced, gjson.GetBytes(f2, "error.id").String())
	})

	resendVerificationCode := func(t *testing.T, client *http.Client, flow string, flowType string, statusCode int) string {
		action := gjson.Get(flow, "ui.action").String()
		assert.NotEmpty(t, action)

		email := gjson.Get(flow, "ui.nodes.#(attributes.name==email).attributes.value").String()

		values := withCSRFToken(t, flowType, flow, url.Values{
			"method": {"code"},
			"email":  {email},
		})

		contentType := "application/json"
		if flowType == RecoveryFlowTypeBrowser {
			contentType = "application/x-www-form-urlencoded"
		}

		res, err := client.Post(action, contentType, bytes.NewBufferString(values))
		require.NoError(t, err)
		assert.Equal(t, statusCode, res.StatusCode)

		return string(ioutilx.MustReadAll(res.Body))
	}

	t.Run("case=should be able to resend code", func(t *testing.T) {
		body := expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		_ = testhelpers.CourierExpectCodeInMessage(t, message, 1)

		c := testhelpers.NewClientWithCookies(t)
		body = resendVerificationCode(t, c, body, RecoveryFlowTypeBrowser, http.StatusOK)

		assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())
		assert.Equal(t, verificationEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		message = testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		verificationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		submitVerificationCode(t, body, c, verificationCode)
	})

	t.Run("case=should not be able to use first code after resending code", func(t *testing.T) {
		body := expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		firstCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		c := testhelpers.NewClientWithCookies(t)
		body = resendVerificationCode(t, c, body, RecoveryFlowTypeBrowser, http.StatusOK)

		assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())
		assert.Equal(t, verificationEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		message = testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		secondCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		body, res := submitVerificationCode(t, body, c, firstCode)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.AssertMessage(t, []byte(body), "The verification code is invalid or has already been used. Please try again.")

		// For good measure, check that the second code still works!

		body, res = submitVerificationCode(t, body, c, secondCode)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.AssertMessage(t, []byte(body), "You successfully verified your email address.")
	})

	t.Run("description=should not be able to use an invalid code more than 5 times", func(t *testing.T) {
		email := strings.ToLower(testhelpers.RandomEmail())
		createIdentityToRecover(t, reg, email)
		c := testhelpers.NewClientWithCookies(t)

		body := expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})
		initialFlowId := gjson.Get(body, "id")

		for submitTry := 0; submitTry < 5; submitTry++ {
			xcbody, _ := submitVerificationCode(t, body, c, "12312312")
			require.Equal(t, initialFlowId.String(), gjson.Get(xcbody, "id").String())

			testhelpers.AssertMessage(t, []byte(xcbody), "The verification code is invalid or has already been used. Please try again.")
		}

		// submit an invalid code for the 6th time
		body, _ = submitVerificationCode(t, body, c, "12312312")

		require.Len(t, gjson.Get(body, "ui.messages").Array(), 1)
		assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "ui.messages.0.text").String())

		// check that a new flow has been created
		assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

		assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==email)").Exists())
	})

	t.Run("description=should be able to verify already verified email address", func(t *testing.T) {
		email := strings.ToLower(testhelpers.RandomEmail())
		createIdentityToRecover(t, reg, email)
		c := testhelpers.NewClientWithCookies(t)

		body := expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})
		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		code := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		body, res := submitVerificationCode(t, body, c, code)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.AssertMessage(t, []byte(body), "You successfully verified your email address.")

		body = expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})
		message = testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
		code = testhelpers.CourierExpectCodeInMessage(t, message, 1)

		body, res = submitVerificationCode(t, body, c, code)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.AssertMessage(t, []byte(body), "You successfully verified your email address.")
	})

	t.Run("case=respects return_to URI parameter", func(t *testing.T) {
		returnToURL := public.URL + "/after-verification"
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToURL})

		for _, fType := range []flow.Type{flow.TypeBrowser, flow.TypeAPI} {
			t.Run(fmt.Sprintf("type=%s", fType), func(t *testing.T) {
				client := testhelpers.NewClientWithCookies(t)
				flow, _, rawCode := newValidFlow(t, fType, public.URL+verification.RouteInitBrowserFlow+"?"+url.Values{"return_to": {returnToURL}}.Encode())

				body := fmt.Sprintf(
					`{"csrf_token":"%s","code":"%s"}`, flow.CSRFToken, rawCode,
				)

				res, err := client.Post(public.URL+verification.RouteSubmitFlow+"?"+url.Values{"flow": {flow.ID.String()}}.Encode(), "application/json", bytes.NewBuffer([]byte(body)))
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
				flow, _, rawCode := newValidFlow(t, fType, public.URL+verification.RouteInitBrowserFlow)

				body := fmt.Sprintf(
					`{"csrf_token":"%s","code":"%s"}`, flow.CSRFToken, rawCode,
				)

				res, err := client.Post(public.URL+verification.RouteSubmitFlow+"?"+url.Values{"flow": {flow.ID.String()}}.Encode(), "application/json", bytes.NewBuffer([]byte(body)))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, res.StatusCode)
				responseBody := gjson.ParseBytes(ioutilx.MustReadAll(res.Body))
				t.Logf("%v", responseBody)

				assert.Equal(t, responseBody.Get("state").String(), "passed_challenge", "%v", responseBody)
				assert.True(t, responseBody.Get("ui.nodes.#(attributes.id==continue)").Exists(), "%v", responseBody)
				assert.Equal(t, globalReturnTo, responseBody.Get("ui.nodes.#(attributes.id==continue).attributes.href").String(), "%v", responseBody)
			})
		}
	})
}
