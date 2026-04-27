// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link_test

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/x/nosurfx"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/ui/node"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"
)

func TestVerification(t *testing.T) {
	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t)
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
			assert.EqualValues(t, "Enter a valid email address",
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
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, true)

		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, false)
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

		sr, _, err := testhelpers.NewSDKCustomClient(public, c).FrontendAPI.GetVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
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
		assert.Contains(t, message.Body, "Verify your account by opening the following link")

		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		time.Sleep(time.Millisecond * 201)

		// Clear cookies as link might be opened in another browser
		c = testhelpers.NewClientWithCookies(t)
		res, err := c.Get(verificationLink)
		require.NoError(t, err)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI(ctx).String())
		assert.NotContains(t, res.Request.URL.String(), gjson.Get(body, "id").String())

		sr, _, err := testhelpers.NewSDKCustomClient(public, c).FrontendAPI.GetVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, sr.Ui.Messages, 1)
		assert.Contains(t, sr.Ui.Messages[0].Text, "The verification flow expired")
	})

	t.Run("description=should verify an email address", func(t *testing.T) {
		var wg sync.WaitGroup
		testhelpers.NewVerifyAfterHookWebHookTarget(ctx, t, conf, func(t *testing.T, msg []byte) {
			defer wg.Done()
			assert.EqualValues(t, true, gjson.GetBytes(msg, "identity.verifiable_addresses.0.verified").Bool(), string(msg))
			assert.EqualValues(t, "completed", gjson.GetBytes(msg, "identity.verifiable_addresses.0.status").String(), string(msg))
		})
		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.LinkGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, verificationEmail, gjson.Get(actual, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationEmailSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
			assert.Contains(t, message.Body, "Verify your account by opening the following link")

			verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			assert.Contains(t, verificationLink, public.URL+verification.RouteSubmitFlow)
			assert.Contains(t, verificationLink, "token=")

			cl := testhelpers.NewClientWithCookies(t)
			res, err := cl.Get(verificationLink)
			require.NoError(t, err)
			defer func() { _ = res.Body.Close() }()

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
		check := func(t *testing.T, actual string) {
			message := testhelpers.CourierExpectMessage(ctx, t, reg, verificationEmail, "Please verify your email address")
			verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			cl := testhelpers.NewClientWithCookies(t)
			res, err := cl.Get(verificationLink)
			require.NoError(t, err)
			body := string(ioutilx.MustReadAll(res.Body))
			require.NoError(t, res.Body.Close())
			require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
			assert.Contains(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL))[0].Name, nosurfx.CSRFTokenName)

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
		f, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken, httptest.NewRequest("GET", requestURL, nil), nil, fType)
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

	t.Run("description=should apply pending traits change when token is redeemed", func(t *testing.T) {
		t.Parallel()
		// Create an identity with original traits.
		pendingID := &identity.Identity{
			ID:       x.NewUUID(),
			State:    identity.StateActive,
			Traits:   identity.Traits(`{"email":"original@ory.sh"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {Type: "password", Identifiers: []string{"original@ory.sh"}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)},
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
		sf := &settings.Flow{
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
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, sf))

		// Create a PendingTraitsChange record linked to this flow.
		sessID := sess.ID
		originFlowID := sf.ID
		ptc := &identity.PendingTraitsChange{
			ID:                   x.NewUUID(),
			IdentityID:           pendingID.ID,
			SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
			OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowID, Valid: true},
			NewAddressValue:      "changed@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   identity.HashTraits(json.RawMessage(pendingID.Traits)),
			ProposedTraits:       json.RawMessage(`{"email":"changed@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		// Create a verification token with a nil address ID (pending-change signal).
		// The address is only used to construct the token; nil it before persisting
		// so pop does not try to create a non-existent association row.
		token := link.NewSelfServiceVerificationToken(ptc, f, time.Hour)
		token.VerifiableAddress = nil
		require.NoError(t, reg.VerificationTokenPersister().CreateVerificationToken(ctx, token))

		// Redeem the token via HTTP GET.
		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.Get(public.URL + verification.RouteSubmitFlow + "?" + url.Values{"flow": {f.ID.String()}, "token": {token.Token}}.Encode())
		require.NoError(t, err)
		body := string(ioutilx.MustReadAll(res.Body))
		require.NoError(t, res.Body.Close())

		require.Equal(t, http.StatusOK, res.StatusCode)
		require.EqualValues(t, "passed_challenge", gjson.Get(body, "state").String(), "%s", body)
		assert.EqualValues(t, text.NewInfoSelfServiceVerificationSuccessful().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)

		// Verify the identity traits were updated.
		updated, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, pendingID.ID)
		require.NoError(t, err)
		assert.EqualValues(t, "changed@ory.sh", gjson.GetBytes([]byte(updated.Traits), "email").String())

		// Verify the new address is marked as verified.
		var foundAddress *identity.VerifiableAddress
		for idx := range updated.VerifiableAddresses {
			if updated.VerifiableAddresses[idx].Value == "changed@ory.sh" {
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
			Traits:   identity.Traits(`{"email":"posthook-link-original@ory.sh"}`),
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
		sfPosthook := &settings.Flow{
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
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, sfPosthook))

		sessID := sess.ID
		originFlowIDPosthook := sfPosthook.ID
		ptc := &identity.PendingTraitsChange{
			ID:                   x.NewUUID(),
			IdentityID:           ident.ID,
			SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
			OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowIDPosthook, Valid: true},
			NewAddressValue:      "posthook-link-new@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   identity.HashTraits(json.RawMessage(ident.Traits)),
			ProposedTraits:       json.RawMessage(`{"email":"posthook-link-new@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		// Create a verification token with a nil address ID (pending-change signal).
		token := link.NewSelfServiceVerificationToken(ptc, f, time.Hour)
		token.VerifiableAddress = nil
		require.NoError(t, reg.VerificationTokenPersister().CreateVerificationToken(ctx, token))

		// Redeem the token via HTTP GET (the link-based verification path).
		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.Get(public.URL + verification.RouteSubmitFlow + "?" + url.Values{"flow": {f.ID.String()}, "token": {token.Token}}.Encode())
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusOK, res.StatusCode)

		select {
		case <-receivedCh:
		case <-time.After(5 * time.Second):
			t.Fatal("settings post-persist webhook was not invoked after pending traits change applied")
		}
		assert.Contains(t, string(receivedBody), "posthook-link-new@ory.sh")
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
			Traits:   identity.Traits(`{"email":"posthook-link-err-original@ory.sh"}`),
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
			NewAddressValue:      "posthook-link-err-new@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   identity.HashTraits(json.RawMessage(ident.Traits)),
			ProposedTraits:       json.RawMessage(`{"email":"posthook-link-err-new@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		// Create a verification token with a nil address ID (pending-change signal).
		token := link.NewSelfServiceVerificationToken(ptc, f, time.Hour)
		token.VerifiableAddress = nil
		require.NoError(t, reg.VerificationTokenPersister().CreateVerificationToken(ctx, token))

		// Redeem the token via HTTP GET (the link-based verification path).
		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.Get(public.URL + verification.RouteSubmitFlow + "?" + url.Values{"flow": {f.ID.String()}, "token": {token.Token}}.Encode())
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		// Apply already committed before the webhook fired: the new traits and
		// the verified address persist even though the hook failed.
		updated, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, ident.ID)
		require.NoError(t, err)
		assert.EqualValues(t, "posthook-link-err-new@ory.sh", gjson.GetBytes([]byte(updated.Traits), "email").String(),
			"traits are committed before the post-persist hook runs")

		var foundAddress *identity.VerifiableAddress
		for idx := range updated.VerifiableAddresses {
			if updated.VerifiableAddresses[idx].Value == "posthook-link-err-new@ory.sh" {
				foundAddress = &updated.VerifiableAddresses[idx]
				break
			}
		}
		require.NotNil(t, foundAddress, "new address should exist on identity")
		assert.True(t, foundAddress.Verified, "new address should be verified even though the webhook failed")
		assert.EqualValues(t, identity.VerifiableAddressStatusCompleted, foundAddress.Status)
	})

	t.Run("description=should reject pending traits change on concurrent modification", func(t *testing.T) {
		t.Parallel()
		// Create an identity with original traits.
		concurrentID := &identity.Identity{
			ID:       x.NewUUID(),
			State:    identity.StateActive,
			Traits:   identity.Traits(`{"email":"original2@ory.sh"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {Type: "password", Identifiers: []string{"original2@ory.sh"}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)},
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
		concurrentID.Traits = identity.Traits(`{"email":"concurrent@ory.sh"}`)
		require.NoError(t, reg.IdentityManager().Update(ctx, concurrentID, identity.ManagerAllowWriteProtectedTraits))

		// Create a verification flow in StateEmailSent.
		f, err := verification.NewFlow(conf, time.Hour, nosurfx.FakeCSRFToken, httptest.NewRequest("GET", public.URL+verification.RouteInitBrowserFlow, nil), nil, flow.TypeBrowser)
		require.NoError(t, err)
		f.State = flow.StateEmailSent
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

		// Create a settings flow that acts as the origin for the PTC.
		sfConcurrent := &settings.Flow{
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
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, sfConcurrent))

		// Create a PendingTraitsChange with the OLD traits hash (before concurrent modification).
		sessID := sess.ID
		originFlowIDConcurrent := sfConcurrent.ID
		ptc := &identity.PendingTraitsChange{
			ID:                   x.NewUUID(),
			IdentityID:           concurrentID.ID,
			SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
			OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowIDConcurrent, Valid: true},
			NewAddressValue:      "changed2@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   originalHash,
			ProposedTraits:       json.RawMessage(`{"email":"changed2@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		// Create a verification token with a nil address ID.
		token := link.NewSelfServiceVerificationToken(ptc, f, time.Hour)
		token.VerifiableAddress = nil
		require.NoError(t, reg.VerificationTokenPersister().CreateVerificationToken(ctx, token))

		// Redeem via GET (the link-based path).
		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.Get(public.URL + verification.RouteSubmitFlow + "?" + url.Values{"flow": {f.ID.String()}, "token": {token.Token}}.Encode())
		require.NoError(t, err)
		body := string(ioutilx.MustReadAll(res.Body))
		require.NoError(t, res.Body.Close())

		// Should show the "token invalid" error (concurrent modification detected).
		assert.Contains(t, body, "The verification token is invalid or has already been used")

		// Identity traits should NOT have been updated.
		unchanged, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, concurrentID.ID)
		require.NoError(t, err)
		assert.EqualValues(t, "concurrent@ory.sh", gjson.GetBytes([]byte(unchanged.Traits), "email").String(),
			"traits should remain at the concurrent value, not the proposed value")
	})

	t.Run("description=should reject pending traits change when session is revoked", func(t *testing.T) {
		ident := &identity.Identity{
			ID:       x.NewUUID(),
			State:    identity.StateActive,
			Traits:   identity.Traits(`{"email":"revoke-link-original@ory.sh"}`),
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
		sfRevoke := &settings.Flow{
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
		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(ctx, sfRevoke))

		sessID := sess.ID
		originFlowIDRevoke := sfRevoke.ID
		ptc := &identity.PendingTraitsChange{
			ID:                   x.NewUUID(),
			IdentityID:           ident.ID,
			SessionID:            uuid.NullUUID{UUID: sessID, Valid: true},
			OriginSettingsFlowID: uuid.NullUUID{UUID: originFlowIDRevoke, Valid: true},
			NewAddressValue:      "revoke-link-new@ory.sh",
			NewAddressVia:        string(identity.AddressTypeEmail),
			OriginalTraitsHash:   identity.HashTraits(json.RawMessage(ident.Traits)),
			ProposedTraits:       json.RawMessage(`{"email":"revoke-link-new@ory.sh"}`),
			VerificationFlowID:   f.ID,
			Status:               identity.PendingTraitsChangeStatusPending,
		}
		require.NoError(t, reg.PendingTraitsChangePersister().CreatePendingTraitsChange(ctx, ptc))

		token := link.NewSelfServiceVerificationToken(ptc, f, time.Hour)
		token.VerifiableAddress = nil
		require.NoError(t, reg.VerificationTokenPersister().CreateVerificationToken(ctx, token))

		// Revoke the session AFTER the PTC is created — simulates logout between flow start and verification.
		require.NoError(t, reg.SessionPersister().RevokeSessionById(ctx, sess.ID))

		cl := testhelpers.NewClientWithCookies(t)
		res, err := cl.Get(public.URL + verification.RouteSubmitFlow + "?" + url.Values{"flow": {f.ID.String()}, "token": {token.Token}}.Encode())
		require.NoError(t, err)
		body := string(ioutilx.MustReadAll(res.Body))
		require.NoError(t, res.Body.Close())

		assert.Contains(t, body, "The verification token is invalid or has already been used")

		// Traits should NOT be updated.
		unchanged, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, ident.ID)
		require.NoError(t, err)
		assert.EqualValues(t, "revoke-link-original@ory.sh", gjson.GetBytes([]byte(unchanged.Traits), "email").String())
	})

	t.Run("case=doesn't continue with OAuth2 flow if code is invalid", func(t *testing.T) {
		globalReturnTo := public.URL + "/global"
		conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, globalReturnTo)

		client := testhelpers.NewClientWithCookies(t)
		flow, _ := newValidFlow(t, flow.TypeBrowser, public.URL+verification.RouteInitBrowserFlow)

		res, err := client.Get(public.URL + verification.RouteSubmitFlow + "?" + url.Values{
			"flow":  {flow.ID.String()},
			"token": {"invalid token"},
		}.Encode())
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		responseBody := gjson.ParseBytes(ioutilx.MustReadAll(res.Body))

		assert.Equal(t, "choose_method", responseBody.Get("state").String(), "%v", responseBody)
		assert.Len(t, responseBody.Get("ui.messages").Array(), 1, "%v", responseBody)
		assert.Equal(t, "The verification token is invalid or has already been used. Please retry the flow.", responseBody.Get("ui.messages.0.text").String(), "%v", responseBody)
	})
}
