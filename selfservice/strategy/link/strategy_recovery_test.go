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

	"github.com/ory/kratos/session"

	"github.com/davecgh/go-spew/spew"

	"github.com/gofrs/uuid"

	"github.com/pkg/errors"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/strategy/link"

	"github.com/ory/kratos/ui/node"

	kratos "github.com/ory/kratos-client-go"

	"github.com/ory/kratos/corpx"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/urlx"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/assertx"

	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func init() {
	corpx.RegisterFakes()
}

func TestAdminStrategy(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
	adminSDK := testhelpers.NewSDKClient(adminTS)

	checkLink := func(t *testing.T, l *kratos.SelfServiceRecoveryLink, isBefore time.Time) {
		require.Contains(t, l.RecoveryLink, publicTS.URL+recovery.RouteSubmitFlow)
		rl := urlx.ParseOrPanic(l.RecoveryLink)
		assert.NotEmpty(t, rl.Query().Get("token"))
		assert.NotEmpty(t, rl.Query().Get("flow"))
		require.True(t, (*l.ExpiresAt).Before(isBefore))
	}

	t.Run("no panic on empty body #1384", func(t *testing.T) {
		ctx := context.Background()
		s, err := reg.RecoveryStrategies(ctx).Strategy("link")
		require.NoError(t, err)
		w := httptest.NewRecorder()
		r := &http.Request{URL: new(url.URL)}
		f, err := recovery.NewFlow(reg.Config(ctx), time.Minute, "", r, reg.RecoveryStrategies(ctx), flow.TypeBrowser)
		require.NoError(t, err)
		require.NotPanics(t, func() {
			require.Error(t, s.(*link.Strategy).HandleRecoveryError(w, r, f, nil, errors.New("test")))
		})
	})

	t.Run("description=should not be able to recover an account that does not exist", func(t *testing.T) {
		_, _, err := adminSDK.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(context.Background()).AdminCreateSelfServiceRecoveryLinkBody(kratos.AdminCreateSelfServiceRecoveryLinkBody{
			IdentityId: x.NewUUID().String(),
		}).Execute()
		require.IsType(t, err, new(kratos.GenericOpenAPIError), "%T", err)
		assert.EqualError(t, err.(*kratos.GenericOpenAPIError), "404 Not Found")
	})

	t.Run("description=should create a valid recovery link without email", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{}`)}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		rl, _, err := adminSDK.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(context.Background()).AdminCreateSelfServiceRecoveryLinkBody(kratos.AdminCreateSelfServiceRecoveryLinkBody{
			IdentityId: id.ID.String(),
			ExpiresIn:  pointerx.String("100ms"),
		}).Execute()
		require.NoError(t, err)

		time.Sleep(time.Millisecond * 100)
		checkLink(t, rl, time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()))
		res, err := publicTS.Client().Get(rl.RecoveryLink)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)

		// We end up here because the link is expired.
		assert.Contains(t, res.Request.URL.Path, "/recover", rl.RecoveryLink)
	})

	t.Run("description=should create a valid recovery link and set the expiry time and not be able to recover the account", func(t *testing.T) {
		recoveryEmail := "recover.expired@ory.sh"
		id := identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, recoveryEmail))}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		rl, _, err := adminSDK.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(context.Background()).AdminCreateSelfServiceRecoveryLinkBody(kratos.AdminCreateSelfServiceRecoveryLinkBody{
			IdentityId: id.ID.String(),
			ExpiresIn:  pointerx.String("100ms"),
		}).Execute()
		require.NoError(t, err)

		time.Sleep(time.Millisecond * 100)
		checkLink(t, rl, time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()))
		res, err := publicTS.Client().Get(rl.RecoveryLink)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)

		// We end up here because the link is expired.
		assert.Contains(t, res.Request.URL.Path, "/recover", rl.RecoveryLink)

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		assert.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})

	t.Run("description=should create a valid recovery link and set the expiry time as well and recover the account", func(t *testing.T) {
		recoveryEmail := "recoverme@ory.sh"
		id := identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, recoveryEmail))}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		rl, _, err := adminSDK.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(context.Background()).AdminCreateSelfServiceRecoveryLinkBody(kratos.AdminCreateSelfServiceRecoveryLinkBody{
			IdentityId: id.ID.String(),
		}).Execute()
		require.NoError(t, err)

		checkLink(t, rl, time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()+time.Second))
		res, err := publicTS.Client().Get(rl.RecoveryLink)
		require.NoError(t, err)

		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String())
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.LogJSON(t, rl)

		f, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(res.Request.URL.Query().Get("flow")))
		require.NoError(t, err, "%s", res.Request.URL.String())

		require.Len(t, f.UI.Messages, 1)
		assert.Equal(t, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.", f.UI.Messages[0].Text)

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		assert.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})
}

func TestRecovery(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)

	var createIdentityToRecover = func(email string) {
		var id = &identity.Identity{
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {Type: "password", Identifiers: []string{email}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)}},
			Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, email)),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
		}
		require.NoError(t, reg.IdentityManager().Create(context.Background(), id, identity.ManagerAllowWriteProtectedTraits))

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, email)
		assert.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	}

	var expect = func(t *testing.T, hc *http.Client, isAPI, isSPA bool, values func(url.Values), c int) string {
		if hc == nil {
			hc = testhelpers.NewDebugClient(t)
			if !isAPI {
				hc = testhelpers.NewClientWithCookies(t)
				hc.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
			}
		}

		return testhelpers.SubmitRecoveryForm(t, isAPI, isSPA, hc, public, values, c,
			testhelpers.ExpectURL(isAPI || isSPA, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI().String()))
	}

	var expectValidationError = func(t *testing.T, hc *http.Client, isAPI, isSPA bool, values func(url.Values)) string {
		return expect(t, hc, isAPI, isSPA, values, testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK))
	}

	var expectSuccess = func(t *testing.T, hc *http.Client, isAPI, isSPA bool, values func(url.Values)) string {
		return expect(t, hc, isAPI, isSPA, values, http.StatusOK)
	}

	t.Run("description=should set all the correct recovery payloads after submission", func(t *testing.T) {
		body := expectSuccess(t, nil, false, false, func(v url.Values) {
			v.Set("email", "test@ory.sh")
		})
		testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes").String()), []string{"0.attributes.value"})
	})

	t.Run("description=should set all the correct recovery payloads", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryFlow(t, c, public)

		testhelpers.SnapshotTExcept(t, rs.Ui.Nodes, []string{"0.attributes.value"})
		assert.EqualValues(t, public.URL+recovery.RouteSubmitFlow+"?flow="+rs.Id, rs.Ui.Action)
		assert.Empty(t, rs.Ui.Messages)
	})

	t.Run("description=should require an email to be sent", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, node.RecoveryLinkGroup, gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, "Property email is missing.",
				gjson.Get(actual, "ui.nodes.#(attributes.name==email).messages.0.text").String(),
				"%s", actual)
		}

		var values = func(v url.Values) {
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
		var check = func(t *testing.T, actual string, value string) {
			assert.EqualValues(t, node.RecoveryLinkGroup, gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, fmt.Sprintf("%q is not valid \"email\"", value),
				gjson.Get(actual, "ui.nodes.#(attributes.name==email).messages.0.text").String(),
				"%s", actual)
		}
		for _, email := range []string{"\\", "asdf", "...", "aiacobelli.sec@gmail.com,alejandro.iacobelli@mercadolibre.com"} {
			var values = func(v url.Values) {
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

	t.Run("description=should try to recover an email that does not exist", func(t *testing.T) {
		var email string
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, node.RecoveryLinkGroup, gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewRecoveryEmailSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(t, reg, email, "Account access attempted")
			assert.Contains(t, message.Body, "If this was you, check if you signed up using a different address.")
		}

		var values = func(v url.Values) {
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

	t.Run("description=should not be able to recover an inactive account", func(t *testing.T) {
		var check = func(t *testing.T, recoverySubmissionResponse, recoveryEmail string, isAPI bool) {
			addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
			assert.NoError(t, err)

			recoveryLink := testhelpers.CourierExpectLinkInMessage(t, testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account"), 1)
			cl := testhelpers.NewClientWithCookies(t)

			// Deactivate the identity
			require.NoError(t, reg.Persister().GetConnection(context.Background()).RawQuery("UPDATE identities SET state=? WHERE id = ?", identity.StateInactive, addr.IdentityID).Exec())

			res, err := cl.Get(recoveryLink)
			require.NoError(t, err)

			body := ioutilx.MustReadAll(res.Body)
			if isAPI {
				assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
				assert.Contains(t, res.Request.URL.String(), public.URL+recovery.RouteSubmitFlow)
				assertx.EqualAsJSON(t, session.ErrIdentityDisabled, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
			} else {
				assert.Equal(t, http.StatusOK, res.StatusCode)
				assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String())
				assertx.EqualAsJSON(t, session.ErrIdentityDisabled, json.RawMessage(body), "%s", body)
			}
		}

		t.Run("type=browser", func(t *testing.T) {
			email := "recoverinactive1@ory.sh"
			createIdentityToRecover(email)
			check(t, expectSuccess(t, nil, false, false, func(v url.Values) {
				v.Set("email", email)
			}), email, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			email := "recoverinactive2@ory.sh"
			createIdentityToRecover(email)
			check(t, expectSuccess(t, nil, true, true, func(v url.Values) {
				v.Set("email", email)
			}), email, true)
		})

		t.Run("type=api", func(t *testing.T) {
			email := "recoverinactive3@ory.sh"
			createIdentityToRecover(email)
			check(t, expectSuccess(t, nil, true, false, func(v url.Values) {
				v.Set("email", email)
			}), email, true)
		})
	})

	t.Run("description=should recover an account", func(t *testing.T) {
		var check = func(t *testing.T, recoverySubmissionResponse, recoveryEmail, returnTo string) {
			addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
			assert.NoError(t, err)
			assert.False(t, addr.Verified)
			assert.Nil(t, addr.VerifiedAt)
			assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)

			assert.EqualValues(t, node.RecoveryLinkGroup, gjson.Get(recoverySubmissionResponse, "active").String(), "%s", recoverySubmissionResponse)
			assert.EqualValues(t, recoveryEmail, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", recoverySubmissionResponse)
			require.Len(t, gjson.Get(recoverySubmissionResponse, "ui.messages").Array(), 1, "%s", recoverySubmissionResponse)
			assertx.EqualAsJSON(t, text.NewRecoveryEmailSent(), json.RawMessage(gjson.Get(recoverySubmissionResponse, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account")
			assert.Contains(t, message.Body, "please recover access to your account by clicking the following link")

			recoveryLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			assert.Contains(t, recoveryLink, public.URL+recovery.RouteSubmitFlow)
			assert.Contains(t, recoveryLink, "token=")

			cl := testhelpers.NewClientWithCookies(t)

			res, err := cl.Get(recoveryLink)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String())

			body := ioutilx.MustReadAll(res.Body)
			assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
				gjson.GetBytes(body, "ui.messages.0.text").String())
			assert.Equal(t, returnTo, gjson.GetBytes(body, "return_to").String())

			addr, err = reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
			assert.NoError(t, err)
			assert.True(t, addr.Verified)
			assert.NotEqual(t, sqlxx.NullTime{}, addr.VerifiedAt)
			assert.Equal(t, identity.VerifiableAddressStatusCompleted, addr.Status)

			res, err = cl.Get(public.URL + session.RouteWhoami)
			require.NoError(t, err)
			body = x.MustReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			assert.Equal(t, "link_recovery", gjson.GetBytes(body, "authentication_methods.0.method").String(), "%s", body)
			assert.Equal(t, "aal1", gjson.GetBytes(body, "authenticator_assurance_level").String(), "%s", body)
		}

		t.Run("type=browser", func(t *testing.T) {
			email := "recoverme1@ory.sh"
			createIdentityToRecover(email)
			check(t, expectSuccess(t, nil, false, false, func(v url.Values) {
				v.Set("email", email)
			}), email, "")
		})

		t.Run("type=browser set return_to", func(t *testing.T) {
			email := "recoverme2@ory.sh"
			returnTo := "https://www.ory.sh"
			createIdentityToRecover(email)

			hc := testhelpers.NewClientWithCookies(t)
			hc.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper

			f := testhelpers.InitializeRecoveryFlowViaBrowser(t, hc, false, public, url.Values{"return_to": []string{returnTo}})

			time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

			formPayload := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			formPayload.Set("email", email)

			b, res := testhelpers.RecoveryMakeRequest(t, false, f, hc, testhelpers.EncodeFormAsJSON(t, false, formPayload))
			assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", b)
			expectedURL := testhelpers.ExpectURL(false, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI().String())
			assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

			check(t, b, email, returnTo)
		})

		t.Run("type=spa", func(t *testing.T) {
			email := "recoverme3@ory.sh"
			createIdentityToRecover(email)
			check(t, expectSuccess(t, nil, true, true, func(v url.Values) {
				v.Set("email", email)
			}), email, "")
		})

		t.Run("type=api", func(t *testing.T) {
			email := "recoverme4@ory.sh"
			createIdentityToRecover(email)
			check(t, expectSuccess(t, nil, true, false, func(v url.Values) {
				v.Set("email", email)
			}), email, "")
		})
	})

	t.Run("description=should recover an account and set the csrf cookies", func(t *testing.T) {
		recoveryEmail := "recoverme1@ory.sh"

		var check = func(t *testing.T, actual string) {
			message := testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account")
			recoveryLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			cl := testhelpers.NewClientWithCookies(t)
			cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			res, err := cl.Get(recoveryLink)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())
			assert.Equal(t, http.StatusSeeOther, res.StatusCode)
			require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
			cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
			assert.Contains(t, cookies, x.CSRFTokenName)
			assert.Contains(t, cookies, "ory_kratos_session")

			rl := urlx.ParseOrPanic(recoveryLink)
			actualRes, err := cl.Get(public.URL + recovery.RouteGetFlow + "?id=" + rl.Query().Get("flow"))
			require.NoError(t, err)
			body := x.MustReadAll(actualRes.Body)
			require.NoError(t, actualRes.Body.Close())
			assert.Equal(t, http.StatusOK, actualRes.StatusCode, "%s", body)
		}

		var values = func(v url.Values) {
			v.Set("email", recoveryEmail)
		}

		check(t, expectSuccess(t, nil, false, false, values))
	})

	t.Run("description=should not be able to use an invalid link", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeRecoveryFlowViaBrowser(t, c, false, public, nil)
		res, err := c.Get(f.Ui.Action + "&token=i-do-not-exist")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String()+"?flow=")

		rs, _, err := testhelpers.NewSDKCustomClient(public, c).V0alpha2Api.GetSelfServiceRecoveryFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, rs.Ui.Messages, 1)
		assert.Equal(t, "The recovery token is invalid or has already been used. Please retry the flow.", rs.Ui.Messages[0].Text)
	})

	t.Run("description=should not be able to use an outdated link", func(t *testing.T) {
		recoveryEmail := "recoverme5@ory.sh"
		createIdentityToRecover(recoveryEmail)
		conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryFlow(t, c, public)

		time.Sleep(time.Millisecond * 201)

		res, err := c.PostForm(rs.Ui.Action, url.Values{"email": {recoveryEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Id)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String())

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		assert.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})

	t.Run("description=should not be able to use an outdated flow", func(t *testing.T) {
		recoveryEmail := "recoverme6@ory.sh"
		createIdentityToRecover(recoveryEmail)
		conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		body := expectSuccess(t, c, false, false, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		message := testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account")
		assert.Contains(t, message.Body, "please recover access to your account by clicking the following link")

		recoveryLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		time.Sleep(time.Millisecond * 201)

		res, err := c.Get(recoveryLink)
		require.NoError(t, err)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String())
		assert.NotContains(t, res.Request.URL.String(), gjson.Get(body, "id").String())

		rs, _, err := testhelpers.NewSDKCustomClient(public, c).V0alpha2Api.GetSelfServiceRecoveryFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, rs.Ui.Messages, 1)
		assert.Contains(t, rs.Ui.Messages[0].Text, "The recovery flow expired")

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		assert.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})
}

func TestDisabledEndpoint(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+recovery.StrategyRecoveryLinkName+".enabled", false)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
	adminSDK := testhelpers.NewSDKClient(adminTS)
	_ = testhelpers.NewErrorTestServer(t, reg)

	t.Run("role=admin", func(t *testing.T) {
		t.Run("description=can not create recovery link when link method is disabled", func(t *testing.T) {
			id := identity.Identity{Traits: identity.Traits(`{"email":"recovery-endpoint-disabled@ory.sh"}`)}

			require.NoError(t, reg.IdentityManager().Create(context.Background(),
				&id, identity.ManagerAllowWriteProtectedTraits))

			rl, _, err := adminSDK.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(context.Background()).AdminCreateSelfServiceRecoveryLinkBody(kratos.AdminCreateSelfServiceRecoveryLinkBody{
				IdentityId: id.ID.String(),
			}).Execute()
			assert.Nil(t, rl)
			require.IsType(t, new(kratos.GenericOpenAPIError), err, "%s", err)

			br, _ := err.(*kratos.GenericOpenAPIError)
			assert.Contains(t, string(br.Body()), "This endpoint was disabled by system administrator", "%s", br.Body())
		})
	})

	t.Run("role=public", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)

		t.Run("description=can not recover an account by get request when link method is disabled", func(t *testing.T) {
			f := testhelpers.InitializeRecoveryFlowViaBrowser(t, c, false, publicTS, nil)
			u := publicTS.URL + recovery.RouteSubmitFlow + "?flow=" + f.Id + "&token=endpoint-disabled"
			res, err := c.Get(u)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})

		t.Run("description=can not recover an account by post request when link method is disabled", func(t *testing.T) {
			f := testhelpers.InitializeRecoveryFlowViaBrowser(t, c, false, publicTS, nil)
			u := publicTS.URL + recovery.RouteSubmitFlow + "?flow=" + f.Id
			res, err := c.PostForm(u, url.Values{"email": {"email@ory.sh"}, "method": {"link"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})
	})
}
