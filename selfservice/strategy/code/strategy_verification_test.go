// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/configx"
	keysetpagination "github.com/ory/x/pagination/keysetpagination_v2"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/assertx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func TestVerification(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValues(defaultConfig))

	// Configure an SMS courier channel so verification supports phone numbers.
	conf.MustSet(ctx, config.ViperKeyCourierChannels, []map[string]any{
		{"id": "sms", "type": "http", "request_config": map[string]any{
			"url":    "http://localhost:1234/sms",
			"method": "POST",
			"body":   "base64://ZnVuY3Rpb24oY3R4KSBjdHg=",
		}},
	})

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

		// The email input label should read "Email or phone number" since verification accepts both.
		emailNode := rs.Ui.Nodes[0]
		assert.EqualValues(t, int(text.InfoNodeLabelEmailOrPhone), emailNode.Meta.Label.Id)
		assert.EqualValues(t, "Email or phone number", emailNode.Meta.Label.Text)
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

	t.Run("description=should accept arbitrary addresses without validation error", func(t *testing.T) {
		// The email field now accepts both email addresses and phone numbers.
		// Malformed values are no longer rejected at the schema level — they are
		// treated as unknown addresses (SMS channel for non-@ values, email for @
		// values). The flow still succeeds to prevent address enumeration.
		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.CodeGroup), gjson.Get(actual, "active").String(), "%s", actual)
			msgID := gjson.Get(actual, "ui.messages.0.id").Int()
			assert.True(t,
				msgID == int64(text.InfoSelfServiceVerificationEmailWithCodeSent) ||
					msgID == int64(text.InfoSelfServiceVerificationPhoneWithCodeSent),
				"expected email or phone code-sent message, got %d in %s", msgID, actual)
		}

		for _, addr := range []string{"\\", "asdf", "...", "aiacobelli.sec@gmail.com,alejandro.iacobelli@mercadolibre.com"} {
			values := func(v url.Values) {
				v.Set("email", addr)
			}

			t.Run("type=browser/value="+addr, func(t *testing.T) {
				check(t, expectSuccess(t, nil, false, false, values))
			})

			t.Run("type=spa/value="+addr, func(t *testing.T) {
				check(t, expectSuccess(t, nil, false, true, values))
			})

			t.Run("type=api/value="+addr, func(t *testing.T) {
				check(t, expectSuccess(t, nil, true, false, values))
			})
		}
	})

	t.Run("description=should accept a phone number for verification", func(t *testing.T) {
		phoneNumber := "+12065551234"
		phoneIdentity := &identity.Identity{
			ID:       x.NewUUID(),
			Traits:   identity.Traits(fmt.Sprintf(`{"phone":"%s"}`, phoneNumber)),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {
					Type:        "password",
					Identifiers: []string{phoneNumber},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`),
				},
			},
		}
		require.NoError(t, reg.IdentityManager().Create(context.Background(), phoneIdentity,
			identity.ManagerAllowWriteProtectedTraits))

		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.CodeGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationPhoneWithCodeSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, phoneNumber, "verification code")
			assert.EqualValues(t, "sms", message.Channel)
			code := testhelpers.CourierExpectCodeInMessage(t, message, 1)
			require.NotEmpty(t, code)
		}

		values := func(v url.Values) {
			v.Set("email", phoneNumber)
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

	t.Run("description=should verify a phone number and show phone-specific success message", func(t *testing.T) {
		phone := "+12065559999"
		phoneIdentity := &identity.Identity{
			ID:       x.NewUUID(),
			Traits:   identity.Traits(fmt.Sprintf(`{"phone":"%s"}`, phone)),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {
					Type:        "password",
					Identifiers: []string{phone},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`),
				},
			},
		}
		require.NoError(t, reg.IdentityManager().Create(context.Background(), phoneIdentity,
			identity.ManagerAllowWriteProtectedTraits))

		body := expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", phone)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, phone, "verification code")
		verificationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		body, _ = submitVerificationCode(t, body, testhelpers.NewClientWithCookies(t), verificationCode)

		assert.EqualValues(t, "passed_challenge", gjson.Get(body, "state").String())
		assert.EqualValues(t, text.NewInfoSelfServiceVerificationPhoneSuccessful().Text, gjson.Get(body, "ui.messages.0.text").String())
		assert.EqualValues(t, int(text.InfoSelfServiceVerificationPhoneSuccessful), gjson.Get(body, "ui.messages.0.id").Int())
	})

	t.Run("description=should not send SMS for unknown phone number", func(t *testing.T) {
		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.CodeGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationPhoneWithCodeSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))
		}

		t.Run("type=browser", func(t *testing.T) {
			unknownPhone := "+10005551001"
			check(t, expectSuccess(t, nil, false, false, func(v url.Values) {
				v.Set("email", unknownPhone)
			}))
		})

		t.Run("type=spa", func(t *testing.T) {
			unknownPhone := "+10005551002"
			check(t, expectSuccess(t, nil, false, true, func(v url.Values) {
				v.Set("email", unknownPhone)
			}))
		})

		t.Run("type=api", func(t *testing.T) {
			unknownPhone := "+10005551003"
			check(t, expectSuccess(t, nil, true, false, func(v url.Values) {
				v.Set("email", unknownPhone)
			}))
		})
	})

	t.Run("description=should not send SMS for unknown phone number even with notify_unknown_recipients", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, true)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, false)
		})

		unknownPhone := "+10005559999"

		body := expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", unknownPhone)
		})

		assert.EqualValues(t, string(node.CodeGroup), gjson.Get(body, "active").String(), "%s", body)
		assertx.EqualAsJSON(t, text.NewVerificationPhoneWithCodeSent(), json.RawMessage(gjson.Get(body, "ui.messages.0").Raw))

		// Verify no SMS was enqueued for the unknown phone number.
		messages, _, err := reg.CourierPersister().ListMessages(ctx, courier.ListCourierMessagesParameters{
			Recipient: unknownPhone,
		}, []keysetpagination.Option{})
		require.NoError(t, err)
		assert.Empty(t, messages, "no SMS should be sent to an unknown phone number even with notify_unknown_recipients enabled")
	})

	t.Run("description=should try to verify an email that does not exist", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, true)

		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, false)
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

		testhelpers.AssertMessage(t, body, "The verification code is invalid or has already been used. Please try again.")
	})

	t.Run("description=should not be able to submit email in expired flow", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, 100*time.Millisecond)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetVerificationFlow(t, c, public)

		time.Sleep(101 * time.Millisecond)

		res, err := c.PostForm(rs.Ui.Action, url.Values{"method": {"code"}, "email": {verificationEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Id)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI(ctx).String())
		body := ioutilx.MustReadAll(res.Body)
		assert.Regexpf(t, regexp.MustCompile(`The verification flow expired 0\.0\d minutes ago, please try again\.`), gjson.GetBytes(body, "ui.messages.0.text").Str, "%s", body)
	})

	t.Run("description=should not be able to submit code in expired flow", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, 100*time.Millisecond)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		body := expectSuccess(t, c, false, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
		assert.Contains(t, message.Body, "Verify your account with the following code")

		code := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		time.Sleep(101 * time.Millisecond)

		f, _ := submitVerificationCode(t, body, c, code)

		assert.Regexpf(t, regexp.MustCompile(`The verification flow expired 0\.0\d minutes ago, please try again\.`), gjson.Get(f, "ui.messages.0.text").Str, "%s", body)
	})

	t.Run("description=should verify an email address", func(t *testing.T) {
		var wg sync.WaitGroup
		testhelpers.NewVerifyAfterHookWebHookTarget(ctx, t, conf, func(t *testing.T, msg []byte) {
			defer wg.Done()
			assert.EqualValues(t, true, gjson.GetBytes(msg, "identity.verifiable_addresses.0.verified").Bool(), string(msg))
			assert.EqualValues(t, "completed", gjson.GetBytes(msg, "identity.verifiable_addresses.0.status").String(), string(msg))
		})

		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.CodeGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, verificationEmail, gjson.Get(actual, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationEmailWithCodeSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
			assert.Contains(t, message.Body, "Verify your account with the following code")

			verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			assert.Contains(t, verificationLink, public.URL+verification.RouteSubmitFlow)
			assert.Contains(t, verificationLink, "code=")

			cl := testhelpers.NewClientWithCookies(t)
			res, err := cl.Get(verificationLink)
			require.NoError(t, err)
			defer func() { _ = res.Body.Close() }()

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
			wg.Add(1)
			check(t, expectSuccess(t, nil, false, false, values))
			wg.Wait()
		})

		t.Run("type=spa", func(t *testing.T) {
			wg.Add(1)
			check(t, expectSuccess(t, nil, false, true, values))
			wg.Wait()
		})

		t.Run("type=api", func(t *testing.T) {
			wg.Add(1)
			check(t, expectSuccess(t, nil, true, false, values))
			wg.Wait()
		})
	})

	t.Run("description=should verify an email address when the link is opened in another browser", func(t *testing.T) {
		values := func(v url.Values) {
			v.Set("email", verificationEmail)
		}

		expectSuccess(t, nil, false, false, values)
		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)
		code := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.Get(verificationLink)
		require.NoError(t, err)
		body := string(ioutilx.MustReadAll(res.Body))
		require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
		assert.Contains(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL))[0].Name, nosurfx.CSRFTokenName)

		actualBody, _ := submitVerificationCode(t, body, cl, code)
		assert.EqualValues(t, "passed_challenge", gjson.Get(actualBody, "state").String())
	})

	newValidFlow := func(t *testing.T, fType flow.Type, requestURL string) (*verification.Flow, *code.VerificationCode, string) {
		f, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken, httptest.NewRequest("GET", requestURL, nil), verification.Strategies{code.NewStrategy(reg)}, fType)
		require.NoError(t, err)
		f.State = flow.StateEmailSent
		u, err := url.Parse(f.RequestURL)
		require.NoError(t, err)
		f.OAuth2LoginChallenge = sqlxx.NullString(u.Query().Get("login_challenge"))
		f.IdentityID = uuid.NullUUID{UUID: x.NewUUID(), Valid: true}
		f.SessionID = uuid.NullUUID{UUID: x.NewUUID(), Valid: true}
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
			"csrf_token": {nosurfx.FakeCSRFToken},
		})
		require.NoError(t, err)
		body := ioutilx.MustReadAll(res.Body)

		assert.Equal(t, returnToURL, gjson.GetBytes(body, "ui.nodes.#(attributes.id==continue).attributes.href").String())
	})

	t.Run("case=should respond with replaced error if successful code is submitted again via api", func(t *testing.T) {
		_ = expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
		assert.Contains(t, message.Body, "Verify your account with the following code")

		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		assert.Contains(t, verificationLink, public.URL+verification.RouteSubmitFlow)
		assert.Contains(t, verificationLink, "code=")

		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.Get(verificationLink)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()

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

	resendVerificationCode := func(t *testing.T, client *http.Client, flow string, flowType ClientType, statusCode int) string {
		action := gjson.Get(flow, "ui.action").String()
		assert.NotEmpty(t, action)

		email := gjson.Get(flow, "ui.nodes.#(attributes.name==email).attributes.value").String()

		values := withCSRFToken(t, flowType, flow, url.Values{
			"method": {"code"},
			"email":  {email},
		})

		contentType := "application/json"
		if flowType == RecoveryClientTypeBrowser {
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

		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
		_ = testhelpers.CourierExpectCodeInMessage(t, message, 1)

		c := testhelpers.NewClientWithCookies(t)
		body = resendVerificationCode(t, c, body, RecoveryClientTypeBrowser, http.StatusOK)

		assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())
		assert.Equal(t, verificationEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		message = testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
		verificationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		submitVerificationCode(t, body, c, verificationCode)
	})

	t.Run("case=should not be able to use first code after resending code", func(t *testing.T) {
		body := expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
		firstCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		c := testhelpers.NewClientWithCookies(t)
		body = resendVerificationCode(t, c, body, RecoveryClientTypeBrowser, http.StatusOK)

		assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())
		assert.Equal(t, verificationEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		message = testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
		secondCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		body, res := submitVerificationCode(t, body, c, firstCode)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.AssertMessage(t, body, "The verification code is invalid or has already been used. Please try again.")

		// For good measure, check that the second code still works!

		body, res = submitVerificationCode(t, body, c, secondCode)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.AssertMessage(t, body, "You successfully verified your email address.")
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
		message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
		code := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		body, res := submitVerificationCode(t, body, c, code)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.AssertMessage(t, body, "You successfully verified your email address.")

		body = expectSuccess(t, nil, true, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})
		message = testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Use code")
		code = testhelpers.CourierExpectCodeInMessage(t, message, 1)

		body, res = submitVerificationCode(t, body, c, code)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.AssertMessage(t, body, "You successfully verified your email address.")
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

	t.Run("case=doesn't continue with OAuth2 flow if code is invalid", func(t *testing.T) {
		returnToURL := public.URL + "/after-verification"
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToURL})

		client := testhelpers.NewClientWithCookies(t)
		flow, _, _ := newValidFlow(t, flow.TypeBrowser,
			public.URL+verification.RouteInitBrowserFlow+"?"+url.Values{
				"return_to":       {returnToURL},
				"login_challenge": {"any_valid_challenge"},
			}.Encode())

		body := fmt.Sprintf(
			`{"csrf_token":"%s","code":"%s"}`, flow.CSRFToken, "2475",
		)

		res, err := client.Post(
			public.URL+verification.RouteSubmitFlow+"?"+url.Values{"flow": {flow.ID.String()}}.Encode(),
			"application/json",
			bytes.NewBuffer([]byte(body)),
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		responseBody := gjson.ParseBytes(ioutilx.MustReadAll(res.Body))

		assert.Equal(t, responseBody.Get("state").String(), "sent_email", "%v", responseBody)
		assert.Len(t, responseBody.Get("ui.messages").Array(), 1, "%v", responseBody)
		assert.Equal(t, "The verification code is invalid or has already been used. Please try again.", responseBody.Get("ui.messages.0.text").String(), "%v", responseBody)
	})

	t.Run("description=should apply pending traits change when code is redeemed", func(t *testing.T) {
		// Create an identity with original traits.
		pendingID := &identity.Identity{
			ID:       x.NewUUID(),
			State:    identity.StateActive,
			Traits:   identity.Traits(`{"email":"original-code@ory.sh"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {Type: "password", Identifiers: []string{"original-code@ory.sh"}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)},
			},
		}
		require.NoError(t, reg.IdentityManager().Create(ctx, pendingID, identity.ManagerAllowWriteProtectedTraits))

		// The apply step now requires a live session that created the change.
		sess, err := testhelpers.NewActiveSession(httptest.NewRequest("GET", "/", nil), reg, pendingID,
			time.Now().UTC(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		require.NoError(t, reg.SessionPersister().UpsertSession(ctx, sess))

		// Create a verification flow in StateEmailSent.
		f, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken, httptest.NewRequest("GET", public.URL+verification.RouteInitBrowserFlow, nil), nil, flow.TypeBrowser)
		require.NoError(t, err)
		f.State = flow.StateEmailSent
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

		// Create a settings flow that acts as the origin for the PTC.
		sfApply := &settings.Flow{
			ID:              x.NewUUID(),
			ExpiresAt:       time.Now().UTC().Add(time.Hour),
			IssuedAt:        time.Now().UTC(),
			RequestURL:      public.URL + "/settings",
			IdentityID:      pendingID.ID,
			Identity:        pendingID,
			Type:            flow.TypeBrowser,
			State:           flow.StateShowForm,
			InternalContext: []byte("{}"),
		}
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, sfApply))

		// Create a PendingTraitsChange record linked to this flow.
		sessID := sess.ID
		originFlowIDApply := sfApply.ID
		ptc := &identity.PendingTraitsChange{
			ID:                   x.NewUUID(),
			IdentityID:           pendingID.ID,
			SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
			OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowIDApply, Valid: true},
			NewAddressValue:      "changed-code@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   identity.HashTraits(json.RawMessage(pendingID.Traits)),
			ProposedTraits:       json.RawMessage(`{"email":"changed-code@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		// Create a verification code with a nil address ID (pending-change signal).
		rawCode := code.GenerateCode()
		_, err = reg.VerificationCodePersister().CreateVerificationCode(ctx, &code.CreateVerificationCodeParams{
			RawCode:           rawCode,
			ExpiresIn:         time.Hour,
			VerifiableAddress: ptc, // PendingTraitsChange implements VerifiableAddressLike; ToPersistable() returns nil address
			FlowID:            f.ID,
		})
		require.NoError(t, err)

		// Submit the code via HTTP POST.
		cl := testhelpers.NewClientWithCookies(t)
		action := public.URL + verification.RouteSubmitFlow + "?flow=" + f.ID.String()
		res, err := cl.PostForm(action, url.Values{
			"code":       {rawCode},
			"csrf_token": {nosurfx.FakeCSRFToken},
		})
		require.NoError(t, err)
		body := string(ioutilx.MustReadAll(res.Body))
		require.NoError(t, res.Body.Close())

		require.Equal(t, http.StatusOK, res.StatusCode)
		require.EqualValues(t, "passed_challenge", gjson.Get(body, "state").String(), "%s", body)
		assert.EqualValues(t, text.NewInfoSelfServiceVerificationSuccessful().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)

		// Verify the identity traits were updated.
		updated, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, pendingID.ID)
		require.NoError(t, err)
		assert.EqualValues(t, "changed-code@ory.sh", gjson.GetBytes([]byte(updated.Traits), "email").String())

		// Verify the new address is marked as verified.
		var foundAddress *identity.VerifiableAddress
		for idx := range updated.VerifiableAddresses {
			if updated.VerifiableAddresses[idx].Value == "changed-code@ory.sh" {
				foundAddress = &updated.VerifiableAddresses[idx]
				break
			}
		}
		require.NotNil(t, foundAddress, "new address should exist on identity")
		assert.True(t, foundAddress.Verified)
		assert.EqualValues(t, identity.VerifiableAddressStatusCompleted, foundAddress.Status)
	})

	t.Run("description=fires settings post-persist webhook after pending traits change is applied", func(t *testing.T) {
		// Set up a recording webhook target.
		var receivedBody []byte
		receivedCh := make(chan struct{}, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedBody, _ = io.ReadAll(r.Body)
			select {
			case receivedCh <- struct{}{}:
			default:
			}
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(ts.Close)

		// Register webhook as a settings-profile post-persist hook.
		conf.MustSet(ctx, "selfservice.flows.settings.after.profile.hooks", []map[string]any{
			{
				"hook": "web_hook",
				"config": map[string]any{
					"url":    ts.URL,
					"method": "POST",
					"body":   "base64://" + base64.StdEncoding.EncodeToString([]byte(`function(ctx) ctx`)),
				},
			},
		})
		t.Cleanup(func() {
			conf.MustSet(ctx, "selfservice.flows.settings.after.profile.hooks", []map[string]any{})
		})

		// Create an identity + active session.
		ident := &identity.Identity{
			ID:       x.NewUUID(),
			Traits:   identity.Traits(`{"email":"posthook-original@ory.sh"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			State:    identity.StateActive,
		}
		require.NoError(t, reg.IdentityManager().Create(ctx, ident, identity.ManagerAllowWriteProtectedTraits))
		sess, err := testhelpers.NewActiveSession(httptest.NewRequest("GET", "/", nil), reg, ident, time.Now().UTC(),
			identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		require.NoError(t, reg.SessionPersister().UpsertSession(ctx, sess))

		// Create verification flow + PTC linked to the session.
		f, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken,
			httptest.NewRequest("GET", public.URL+verification.RouteInitBrowserFlow, nil), nil, flow.TypeBrowser)
		require.NoError(t, err)
		f.State = flow.StateEmailSent
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

		// Create a settings flow that acts as the origin for the PTC.
		sfHook := &settings.Flow{
			ID:              x.NewUUID(),
			ExpiresAt:       time.Now().UTC().Add(time.Hour),
			IssuedAt:        time.Now().UTC(),
			RequestURL:      public.URL + "/settings",
			IdentityID:      ident.ID,
			Identity:        ident,
			Type:            flow.TypeBrowser,
			State:           flow.StateShowForm,
			InternalContext: []byte("{}"),
		}
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, sfHook))

		sessID := sess.ID
		originFlowIDHook := sfHook.ID
		ptc := &identity.PendingTraitsChange{
			ID:                   x.NewUUID(),
			IdentityID:           ident.ID,
			SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
			OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowIDHook, Valid: true},
			NewAddressValue:      "posthook-new@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   identity.HashTraits(json.RawMessage(ident.Traits)),
			ProposedTraits:       json.RawMessage(`{"email":"posthook-new@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		rawCode := code.GenerateCode()
		_, err = reg.VerificationCodePersister().CreateVerificationCode(ctx, &code.CreateVerificationCodeParams{
			RawCode:           rawCode,
			ExpiresIn:         time.Hour,
			VerifiableAddress: ptc,
			FlowID:            f.ID,
		})
		require.NoError(t, err)

		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.PostForm(public.URL+verification.RouteSubmitFlow+"?flow="+f.ID.String(),
			url.Values{"code": {rawCode}, "csrf_token": {nosurfx.FakeCSRFToken}})
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusOK, res.StatusCode)

		select {
		case <-receivedCh:
		case <-time.After(5 * time.Second):
			t.Fatal("settings post-persist webhook was not invoked after pending traits change applied")
		}
		assert.Contains(t, string(receivedBody), "posthook-new@ory.sh")
	})

	t.Run("description=post-persist webhook error after pending traits change is surfaced", func(t *testing.T) {
		// Webhook target that always returns 500.
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
		}))
		t.Cleanup(ts.Close)

		conf.MustSet(ctx, "selfservice.flows.settings.after.profile.hooks", []map[string]any{
			{
				"hook": "web_hook",
				"config": map[string]any{
					"url":    ts.URL,
					"method": "POST",
					"body":   "base64://" + base64.StdEncoding.EncodeToString([]byte(`function(ctx) ctx`)),
				},
			},
		})
		t.Cleanup(func() {
			conf.MustSet(ctx, "selfservice.flows.settings.after.profile.hooks", []map[string]any{})
		})

		ident := &identity.Identity{
			ID:       x.NewUUID(),
			Traits:   identity.Traits(`{"email":"posthook-err-original@ory.sh"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			State:    identity.StateActive,
		}
		require.NoError(t, reg.IdentityManager().Create(ctx, ident, identity.ManagerAllowWriteProtectedTraits))
		sess, err := testhelpers.NewActiveSession(httptest.NewRequest("GET", "/", nil), reg, ident, time.Now().UTC(),
			identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		require.NoError(t, reg.SessionPersister().UpsertSession(ctx, sess))

		f, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken,
			httptest.NewRequest("GET", public.URL+verification.RouteInitBrowserFlow, nil), nil, flow.TypeBrowser)
		require.NoError(t, err)
		f.State = flow.StateEmailSent
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

		// Create a settings flow that acts as the origin for the PTC.
		sfErr := &settings.Flow{
			ID:              x.NewUUID(),
			ExpiresAt:       time.Now().UTC().Add(time.Hour),
			IssuedAt:        time.Now().UTC(),
			RequestURL:      public.URL + "/settings",
			IdentityID:      ident.ID,
			Identity:        ident,
			Type:            flow.TypeBrowser,
			State:           flow.StateShowForm,
			InternalContext: []byte("{}"),
		}
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, sfErr))

		sessID := sess.ID
		originFlowIDErr := sfErr.ID
		ptc := &identity.PendingTraitsChange{
			ID:                   x.NewUUID(),
			IdentityID:           ident.ID,
			SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
			OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowIDErr, Valid: true},
			NewAddressValue:      "posthook-err-new@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   identity.HashTraits(json.RawMessage(ident.Traits)),
			ProposedTraits:       json.RawMessage(`{"email":"posthook-err-new@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		rawCode := code.GenerateCode()
		_, err = reg.VerificationCodePersister().CreateVerificationCode(ctx, &code.CreateVerificationCodeParams{
			RawCode:           rawCode,
			ExpiresIn:         time.Hour,
			VerifiableAddress: ptc,
			FlowID:            f.ID,
		})
		require.NoError(t, err)

		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.PostForm(public.URL+verification.RouteSubmitFlow+"?flow="+f.ID.String(),
			url.Values{"code": {rawCode}, "csrf_token": {nosurfx.FakeCSRFToken}})
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		// Apply already committed before the webhook fired: the new traits and
		// the completed PTC status persist even though the hook failed.
		updated, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, ident.ID)
		require.NoError(t, err)
		assert.EqualValues(t, "posthook-err-new@ory.sh", gjson.GetBytes([]byte(updated.Traits), "email").String(),
			"traits are committed before the post-persist hook runs")
	})

	t.Run("description=should reject pending traits change on concurrent modification", func(t *testing.T) {
		// Create an identity with original traits.
		concurrentID := &identity.Identity{
			ID:       x.NewUUID(),
			State:    identity.StateActive,
			Traits:   identity.Traits(`{"email":"original2-code@ory.sh"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {Type: "password", Identifiers: []string{"original2-code@ory.sh"}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)},
			},
		}
		require.NoError(t, reg.IdentityManager().Create(ctx, concurrentID, identity.ManagerAllowWriteProtectedTraits))

		// A live session is required so we reach the traits-hash check rather
		// than bailing at the SessionID-nil guard first.
		sess, err := testhelpers.NewActiveSession(httptest.NewRequest("GET", "/", nil), reg, concurrentID,
			time.Now().UTC(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		require.NoError(t, reg.SessionPersister().UpsertSession(ctx, sess))

		// Compute traits hash from the original traits.
		originalHash := identity.HashTraits(json.RawMessage(concurrentID.Traits))

		// Simulate a concurrent modification: update the identity's traits directly.
		concurrentID.Traits = identity.Traits(`{"email":"concurrent-code@ory.sh"}`)
		require.NoError(t, reg.IdentityManager().Update(ctx, concurrentID, identity.ManagerAllowWriteProtectedTraits))

		// Create a verification flow in StateEmailSent.
		f, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken, httptest.NewRequest("GET", public.URL+verification.RouteInitBrowserFlow, nil), nil, flow.TypeBrowser)
		require.NoError(t, err)
		f.State = flow.StateEmailSent
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

		// Create a settings flow that acts as the origin for the PTC.
		sfConc := &settings.Flow{
			ID:              x.NewUUID(),
			ExpiresAt:       time.Now().UTC().Add(time.Hour),
			IssuedAt:        time.Now().UTC(),
			RequestURL:      public.URL + "/settings",
			IdentityID:      concurrentID.ID,
			Identity:        concurrentID,
			Type:            flow.TypeBrowser,
			State:           flow.StateShowForm,
			InternalContext: []byte("{}"),
		}
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, sfConc))

		// Create a PendingTraitsChange with the OLD traits hash (before concurrent modification).
		sessID := sess.ID
		originFlowIDConc := sfConc.ID
		ptc := &identity.PendingTraitsChange{
			ID:                   x.NewUUID(),
			IdentityID:           concurrentID.ID,
			SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
			OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowIDConc, Valid: true},
			NewAddressValue:      "changed2-code@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   originalHash,
			ProposedTraits:       json.RawMessage(`{"email":"changed2-code@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		// Create a verification code with nil address ID.
		rawCode := code.GenerateCode()
		_, err = reg.VerificationCodePersister().CreateVerificationCode(ctx, &code.CreateVerificationCodeParams{
			RawCode:           rawCode,
			ExpiresIn:         time.Hour,
			VerifiableAddress: ptc,
			FlowID:            f.ID,
		})
		require.NoError(t, err)

		// Submit the code via HTTP POST — should be rejected due to concurrent modification.
		cl := testhelpers.NewClientWithCookies(t)
		action := public.URL + verification.RouteSubmitFlow + "?flow=" + f.ID.String()
		res, err := cl.PostForm(action, url.Values{
			"code":       {rawCode},
			"csrf_token": {nosurfx.FakeCSRFToken},
		})
		require.NoError(t, err)
		body := string(ioutilx.MustReadAll(res.Body))
		require.NoError(t, res.Body.Close())

		// Should show the "code invalid" error (concurrent modification detected).
		assert.Contains(t, body, "The verification code is invalid or has already been used")

		// Identity traits should NOT have been updated.
		unchanged, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, concurrentID.ID)
		require.NoError(t, err)
		assert.EqualValues(t, "concurrent-code@ory.sh", gjson.GetBytes([]byte(unchanged.Traits), "email").String(),
			"traits should remain at the concurrent value, not the proposed value")
	})

	t.Run("description=should reject pending traits change when session is revoked", func(t *testing.T) {
		ident := &identity.Identity{
			ID:       x.NewUUID(),
			State:    identity.StateActive,
			Traits:   identity.Traits(`{"email":"revoke-original@ory.sh"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
		}
		require.NoError(t, reg.IdentityManager().Create(ctx, ident, identity.ManagerAllowWriteProtectedTraits))

		sess, err := testhelpers.NewActiveSession(httptest.NewRequest("GET", "/", nil), reg, ident,
			time.Now().UTC(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		require.NoError(t, reg.SessionPersister().UpsertSession(ctx, sess))

		f, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken,
			httptest.NewRequest("GET", public.URL+verification.RouteInitBrowserFlow, nil), nil, flow.TypeBrowser)
		require.NoError(t, err)
		f.State = flow.StateEmailSent
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

		// Create a settings flow that acts as the origin for the PTC.
		sfRevokeCode := &settings.Flow{
			ID:              x.NewUUID(),
			ExpiresAt:       time.Now().UTC().Add(time.Hour),
			IssuedAt:        time.Now().UTC(),
			RequestURL:      public.URL + "/settings",
			IdentityID:      ident.ID,
			Identity:        ident,
			Type:            flow.TypeBrowser,
			State:           flow.StateShowForm,
			InternalContext: []byte("{}"),
		}
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, sfRevokeCode))

		sessID := sess.ID
		originFlowIDRevokeCode := sfRevokeCode.ID
		ptc := &identity.PendingTraitsChange{
			ID:                   x.NewUUID(),
			IdentityID:           ident.ID,
			SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
			OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowIDRevokeCode, Valid: true},
			NewAddressValue:      "revoke-new@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   identity.HashTraits(json.RawMessage(ident.Traits)),
			ProposedTraits:       json.RawMessage(`{"email":"revoke-new@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		rawCode := code.GenerateCode()
		_, err = reg.VerificationCodePersister().CreateVerificationCode(ctx, &code.CreateVerificationCodeParams{
			RawCode:           rawCode,
			ExpiresIn:         time.Hour,
			VerifiableAddress: ptc,
			FlowID:            f.ID,
		})
		require.NoError(t, err)

		// Revoke the session AFTER the PTC is created — simulates logout between flow start and verification.
		require.NoError(t, reg.SessionPersister().RevokeSessionById(ctx, sess.ID))

		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.PostForm(public.URL+verification.RouteSubmitFlow+"?flow="+f.ID.String(),
			url.Values{"code": {rawCode}, "csrf_token": {nosurfx.FakeCSRFToken}})
		require.NoError(t, err)
		body := string(ioutilx.MustReadAll(res.Body))
		require.NoError(t, res.Body.Close())

		assert.Contains(t, body, "The verification code is invalid or has already been used")

		// Traits should NOT be updated.
		unchanged, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, ident.ID)
		require.NoError(t, err)
		assert.EqualValues(t, "revoke-original@ory.sh", gjson.GetBytes([]byte(unchanged.Traits), "email").String())
	})
}
