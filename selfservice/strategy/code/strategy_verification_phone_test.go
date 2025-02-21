// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"encoding/json"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	client "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/ioutilx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestPhoneVerification(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
	initViper(t, ctx, conf)

	var externalVerifyResult string
	var externalVerifyRequestBody string
	initExternalSMSVerifier(t, ctx, conf, "file://./stub/request.config.verification.jsonnet",
		&externalVerifyRequestBody, &externalVerifyResult)

	var identityToVerify = &identity.Identity{
		ID:       x.NewUUID(),
		Traits:   identity.Traits(`{"phone":"+4580010000"}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
	}

	var verificationPhone = gjson.GetBytes(identityToVerify.Traits, "phone").String()

	_ = testhelpers.NewVerificationUIFlowEchoServer(t, reg)

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)

	require.NoError(t, reg.IdentityManager().Create(ctx, identityToVerify,
		identity.ManagerAllowWriteProtectedTraits))

	var expect = func(t *testing.T, hc *http.Client, isAPI, isSPA bool, values func(url.Values), c int) string {
		if hc == nil {
			hc = testhelpers.NewDebugClient(t)
			if !isAPI {
				hc = testhelpers.NewClientWithCookies(t)
				hc.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
			}
		}

		return testhelpers.SubmitVerificationForm(t, isAPI, isSPA, hc, public, values, c,
			testhelpers.ExpectURL(isAPI || isSPA,
				public.URL+verification.RouteSubmitFlow, conf.SelfServiceFlowVerificationUI(ctx).String()))
	}

	var expectSuccess = func(t *testing.T, hc *http.Client, isAPI, isSPA bool,
		values func(url.Values)) string {
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

	t.Run("description=should not be able to verify phone if external verifier is disabled", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyCode)+".external_sms_verify.enabled", false)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyCode)+".external_sms_verify.enabled", true)
		})
		c := testhelpers.NewClientWithCookies(t)
		f := testhelpers.SubmitVerificationForm(t, true, false, c, public, func(v url.Values) {
			v.Set("phone", verificationPhone)
		}, 400, "")

		assert.Contains(t, f, "External SMS verification service is disabled", "%s", f)
	})

	t.Run("description=should not be able to use an invalid code", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		f := testhelpers.SubmitVerificationForm(t, false, false, c, public, func(v url.Values) {
			v.Set("phone", verificationPhone)
		}, 200, "")

		body, res := submitVerificationCode(t, f, c, "1111")
		assert.Equal(t, http.StatusOK, res.StatusCode)

		assert.Contains(t, externalVerifyResult, "code invalid")

		testhelpers.AssertMessage(t, []byte(body), "The verification code is invalid or has already been used. Please try again.")
	})

	t.Run("description=should verify phone with external verify service", func(t *testing.T) {
		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.CodeGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationEmailWithCodeSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			assert.Contains(t, externalVerifyResult, "code has been sent")

			cl := testhelpers.NewClientWithCookies(t)

			body, res := submitVerificationCode(t, actual, cl, "0000")

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, externalVerifyResult, "code valid")
			assert.EqualValues(t, "passed_challenge", gjson.Get(body, "state").String(), "%s", body)
			assert.EqualValues(t, text.NewInfoSelfServicePhoneVerificationSuccessful().Text, gjson.Get(body, "ui.messages.0.text").String())

			id, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), identityToVerify.ID)
			require.NoError(t, err)
			require.Len(t, id.VerifiableAddresses, 1)

			address := id.VerifiableAddresses[0]
			assert.EqualValues(t, verificationPhone, address.Value)
			assert.True(t, address.Verified)
			assert.EqualValues(t, identity.VerifiableAddressStatusCompleted, address.Status)
			assert.True(t, time.Time(*address.VerifiedAt).Add(time.Second*5).After(time.Now()))
		}

		values := func(v url.Values) {
			v.Set("phone", verificationPhone)
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

	t.Run("description=should save transient payload to template data", func(t *testing.T) {
		var doTest = func(t *testing.T, client *http.Client, isAPI bool, f *client.VerificationFlow) {
			externalVerifyRequestBody = ""
			expectSuccess(t, client, isAPI, false,
				func(v url.Values) {
					v.Set("method", "code")
					v.Set("phone", verificationPhone)
					v.Set("transient_payload", `{"branding": "brand-1"}`)
				})
			assert.Equal(t, "code has been sent", externalVerifyResult)
			assert.Contains(t, externalVerifyRequestBody, "brand-1", "%s", externalVerifyRequestBody)
		}

		t.Run("type=browser", func(t *testing.T) {
			c := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeVerificationFlowViaBrowser(t, c, false, public)
			doTest(t, c, false, f)
		})
		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeVerificationFlowViaAPI(t, nil, public)
			doTest(t, nil, true, f)
		})
	})
}
