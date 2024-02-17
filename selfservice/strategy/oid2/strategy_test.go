package oid2_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/oid2"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/x/urlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestStrategy(t *testing.T) {
	ctx := context.Background()
	if testing.Short() {
		t.Skip()
	}

	var (
		conf, reg = internal.NewFastRegistryWithMocks(t)
	)

	returnTS := newReturnTs(t, reg)
	conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTS.URL})
	routerP := x.NewRouterPublic()
	routerA := x.NewRouterAdmin()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, routerP, routerA)
	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := newUI(t, reg)
	viperSetProviderConfig(
		t,
		conf,
		// TODO #3631 start a real server here?
		newOid2Provider("valid", "http://localhost:12345"),
	)

	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnabled, true)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
	conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter,
		identity.CredentialsTypeOID2.String()), []config.SelfServiceHook{{Name: "session"}})

	//loginAction := func(flowID uuid.UUID) string {
	//	return ts.URL + login.RouteSubmitFlow + "?flow=" + flowID.String()
	//}
	//newLoginFlow := func(t *testing.T, requestURL string, exp time.Duration, flowType flow.Type) (req *login.Flow) {
	//	// Use NewLoginFlow to instantiate the request but change the things we need to control a copy of it.
	//	req, _, err := reg.LoginHandler().NewLoginFlow(httptest.NewRecorder(),
	//		&http.Request{URL: urlx.ParseOrPanic(requestURL)}, flowType)
	//	require.NoError(t, err)
	//	req.RequestURL = requestURL
	//	req.ExpiresAt = time.Now().Add(exp)
	//	require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), req))
	//
	//	// sanity check
	//	got, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), req.ID)
	//	require.NoError(t, err)
	//
	//	require.Len(t, got.UI.Nodes, len(req.UI.Nodes), "%+v", got)
	//
	//	return
	//}
	//newBrowserLoginFlow := func(t *testing.T, redirectTo string, exp time.Duration) (req *login.Flow) {
	//	return newLoginFlow(t, redirectTo, exp, flow.TypeBrowser)
	//}

	registerAction := func(flowID uuid.UUID) string {
		return ts.URL + registration.RouteSubmitFlow + "?flow=" + flowID.String()
	}
	newRegistrationFlow := func(t *testing.T, redirectTo string, exp time.Duration, flowType flow.Type) *registration.Flow {
		// Use NewLoginFlow to instantiate the request but change the things we need to control a copy of it.
		req, err := reg.RegistrationHandler().NewRegistrationFlow(httptest.NewRecorder(),
			&http.Request{URL: urlx.ParseOrPanic(redirectTo)}, flowType)
		require.NoError(t, err)
		req.RequestURL = redirectTo
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.RegistrationFlowPersister().UpdateRegistrationFlow(context.Background(), req))

		// sanity check
		got, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), req.ID)
		require.NoError(t, err)
		require.Len(t, got.UI.Nodes, len(req.UI.Nodes), "%+v", req)

		return req
	}
	newBrowserRegistrationFlow := func(t *testing.T, redirectTo string, exp time.Duration) *registration.Flow {
		return newRegistrationFlow(t, redirectTo, exp, flow.TypeBrowser)
	}

	makeRequestWithCookieJar := func(t *testing.T, provider string, action string, fv url.Values, jar *cookiejar.Jar) (*http.Response, []byte) {
		fv.Set("provider", provider)
		res, err := testhelpers.NewClientWithCookieJar(t, jar, false).PostForm(action, fv)
		require.NoError(t, err, action)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)

		require.Equal(t, 200, res.StatusCode, "%s: %s\n\t%s", action, res.Request.URL.String(), body)

		return res, body
	}
	makeRequest := func(t *testing.T, provider string, action string, fv url.Values) (*http.Response, []byte) {
		return makeRequestWithCookieJar(t, provider, action, fv, nil)
	}

	assertFormValues := func(t *testing.T, flowID uuid.UUID, provider string) (action string) {
		var config *container.Container
		if req, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), flowID); err == nil {
			require.EqualValues(t, req.ID, flowID)
			config = req.UI
			require.NotNil(t, config)
		} else if req, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), flowID); err == nil {
			require.EqualValues(t, req.ID, flowID)
			config = req.UI
			require.NotNil(t, config)
		} else {
			require.NoError(t, err)
			return
		}

		assert.Equal(t, "POST", config.Method)

		// TODO #3631 re-enable this once PopulateRegistrationMethod has been implemented
		//var found bool
		//for _, field := range config.Nodes {
		//	if strings.Contains(field.ID(), "provider") && field.GetValue() == provider {
		//		found = true
		//		break
		//	}
		//}
		//require.True(t, found, "%+v", assertx.PrettifyJSONPayload(t, config))

		return config.Action
	}
	assertSystemErrorWithMessage := func(t *testing.T, res *http.Response, body []byte, code int, message string) {
		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

		assert.Equal(t, int64(code), gjson.GetBytes(body, "code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "message").String(), message, "%s", body)
	}
	assertSystemErrorWithReason := func(t *testing.T, res *http.Response, body []byte, code int, reason string) {
		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

		assert.Equal(t, int64(code), gjson.GetBytes(body, "code").Int(), "%s", prettyJSON(t, body))
		assert.Contains(t, gjson.GetBytes(body, "reason").String(), reason, "%s", prettyJSON(t, body))
	}

	t.Run("case=should fail because provider does not exist", func(t *testing.T) {
		for k, v := range []string{
			//loginAction(newBrowserLoginFlow(t, returnTS.URL, time.Minute).ID),
			registerAction(newBrowserRegistrationFlow(t, returnTS.URL, time.Minute).ID),
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				res, body := makeRequest(t, "provider-does-not-exist", v, url.Values{})
				assertSystemErrorWithReason(t, res, body, http.StatusNotFound, "is unknown or has not been configured")
			})
		}
	})

	t.Run("case=should fail because flow does not exist", func(t *testing.T) {
		for k, v := range []string{ /*loginAction(x.NewUUID()), */ registerAction(x.NewUUID())} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				res, body := makeRequest(t, "valid", v, url.Values{})
				assertSystemErrorWithMessage(t, res, body, http.StatusNotFound, "Unable to locate the resource")
			})
		}
	})

	t.Run("case=should fail because the flow is expired", func(t *testing.T) {
		for k, v := range []uuid.UUID{
			//newBrowserLoginFlow(t, returnTS.URL, -time.Minute).ID,
			newBrowserRegistrationFlow(t, returnTS.URL, -time.Minute).ID,
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				action := assertFormValues(t, v, "valid")
				res, body := makeRequest(t, "valid", action, url.Values{})

				assert.NotEqual(t, v, gjson.GetBytes(body, "id"))
				require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "flow expired", "%s", body)
			})
		}
	})
}

func newOid2Provider(
	provider string,
	discoveryUrl string,
) oid2.Configuration {
	return oid2.Configuration{
		Provider:     provider,
		DiscoveryUrl: discoveryUrl,
	}
}

func newReturnTs(t *testing.T, reg driver.Registry) *httptest.Server {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app_code" {
			reg.Writer().Write(w, r, "ok")
			return
		}
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		require.NoError(t, err)
		reg.Writer().Write(w, r, sess)
	}))
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL)
	t.Cleanup(ts.Close)
	return ts
}

func newUI(t *testing.T, reg driver.Registry) *httptest.Server {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var e interface{}
		var err error
		if r.URL.Path == "/login" {
			e, err = reg.LoginFlowPersister().GetLoginFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		} else if r.URL.Path == "/registration" {
			e, err = reg.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		} else if r.URL.Path == "/settings" {
			e, err = reg.SettingsFlowPersister().GetSettingsFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		}

		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	t.Cleanup(ts.Close)
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceLoginUI, ts.URL+"/login")
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceRegistrationUI, ts.URL+"/registration")
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceSettingsURL, ts.URL+"/settings")
	return ts
}

func prettyJSON(t *testing.T, body []byte) string {
	var out bytes.Buffer
	require.NoError(t, json.Indent(&out, body, "", "\t"))

	return out.String()
}

func viperSetProviderConfig(t *testing.T, conf *config.Config, providers ...oid2.Configuration) {
	ctx := context.Background()
	baseKey := fmt.Sprintf("%s.%s", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeOID2)
	currentConfig := conf.GetProvider(ctx).Get(baseKey + ".config")
	currentEnabled := conf.GetProvider(ctx).Get(baseKey + ".enabled")

	conf.MustSet(ctx, baseKey+".config", &oid2.ConfigurationCollection{Providers: providers})
	conf.MustSet(ctx, baseKey+".enabled", true)

	t.Cleanup(func() {
		conf.MustSet(ctx, baseKey+".config", currentConfig)
		conf.MustSet(ctx, baseKey+".enabled", currentEnabled)
	})
}
