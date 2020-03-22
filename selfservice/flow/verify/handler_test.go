package verify_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/stringslice"
	"github.com/ory/x/urlx"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/client"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/verify"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func TestHandler(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)

	publicTS, adminTS := func() (*httptest.Server, *httptest.Server) {
		public := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		reg.VerificationHandler().RegisterPublicRoutes(public)
		reg.VerificationHandler().RegisterAdminRoutes(admin)
		return httptest.NewServer(x.NewTestCSRFHandler(public, reg)), httptest.NewServer(admin)
	}()
	defer publicTS.Close()
	defer adminTS.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer redirTS.Close()

	verifyTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Query().Get("request")))
	}))
	defer verifyTS.Close()

	errTS := testhelpers.NewErrorTestServer(t, reg)
	defer errTS.Close()

	viper.Set(configuration.ViperKeyURLsSelfPublic, publicTS.URL)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/extension/schema.json")
	viper.Set(configuration.ViperKeyURLsError, errTS.URL)
	viper.Set(configuration.ViperKeyURLsVerification, verifyTS.URL)
	viper.Set(configuration.ViperKeySelfServiceVerifyReturnTo, redirTS.URL)
	viper.Set(configuration.ViperKeyCourierSMTPURL, "smtp://foo:bar@stub/")

	publicClient := client.NewHTTPClientWithConfig(nil,
		&client.TransportConfig{Host: urlx.ParseOrPanic(publicTS.URL).Host, BasePath: "/", Schemes: []string{"http"}})
	adminClient := client.NewHTTPClientWithConfig(nil,
		&client.TransportConfig{Host: urlx.ParseOrPanic(adminTS.URL).Host, BasePath: "/", Schemes: []string{"http"}})

	t.Run("case=fetch unknown request ID", func(t *testing.T) {
		for name, tc := range map[string]struct {
			client    *client.OryKratos
			expectErr interface{}
		}{"public": {client: publicClient, expectErr: new(common.GetSelfServiceVerificationRequestForbidden)},
			"admin": {client: adminClient, expectErr: new(common.GetSelfServiceVerificationRequestNotFound)}} {
			t.Run("api="+name, func(t *testing.T) {
				_, err := tc.client.Common.GetSelfServiceVerificationRequest(common.NewGetSelfServiceVerificationRequestParams().WithRequest("does-not-exist"))
				require.Error(t, err)
				assert.IsType(t, tc.expectErr, err)
			})
		}
	})

	t.Run("case=request verification for unknown via", func(t *testing.T) {
		res, body := x.EasyGet(t,
			&http.Client{Jar: x.EasyCookieJar(t, nil)},
			publicTS.URL+strings.Replace(verify.PublicVerificationInitPath, ":via", "notemail", 1))
		assert.Contains(t, res.Request.URL.String(), errTS.URL)
		assert.EqualValues(t, http.StatusBadRequest, gjson.GetBytes(body, "0.code").Int())
	})

	initURL := publicTS.URL + strings.Replace(verify.PublicVerificationInitPath, ":via", "email", 1)

	t.Run("case=init and validate request payload", func(t *testing.T) {
		hc := &http.Client{Jar: x.EasyCookieJar(t, nil)}

		res, _ := x.EasyGet(t, hc, initURL)
		assert.Contains(t, res.Request.URL.String(), verifyTS.URL)

		rid := res.Request.URL.Query().Get("request")
		require.NotEmpty(t, rid)

		svr, err := adminClient.Common.GetSelfServiceVerificationRequest(common.NewGetSelfServiceVerificationRequestParams().WithRequest(rid))
		require.NoError(t, err)

		assert.Equal(t, "email", string(svr.Payload.Via))
		assert.True(t, time.Time(svr.Payload.ExpiresAt).After(time.Now()))
		assert.Contains(t, svr.Payload.RequestURL, initURL)
		assert.Contains(t, svr.Payload.ID, rid)
		assert.Equal(t, publicTS.URL+strings.Replace(verify.PublicVerificationCompletePath, ":via", "email", 1)+"?request="+rid, *svr.Payload.Form.Action)
		assert.Contains(t, "csrf_token", *svr.Payload.Form.Fields[0].Name)
		assert.Contains(t, "to_verify", *svr.Payload.Form.Fields[1].Name)
		assert.Contains(t, "email", *svr.Payload.Form.Fields[1].Type)
	})

	t.Run("case=fetch request with no csrf from public causes error", func(t *testing.T) {
		_, err := publicClient.Common.GetSelfServiceVerificationRequest(common.NewGetSelfServiceVerificationRequestParams().WithRequest(
			string(x.EasyGetBody(t, &http.Client{Jar: x.EasyCookieJar(t, nil)}, initURL)),
		))
		assert.IsType(t, new(common.GetSelfServiceVerificationRequestForbidden), err)
	})

	genForm := func(t *testing.T, res *common.GetSelfServiceVerificationRequestOK, to string, ignoreFields ...string) (action string, v url.Values) {
		v = make(url.Values)
		action = *res.Payload.Form.Action
		for _, field := range res.Payload.Form.Fields {
			if stringslice.Has(ignoreFields, *field.Name) {
				continue
			}

			if *field.Name == "to_verify" {
				field.Value = to
			}
			v[*field.Name] = []string{fmt.Sprintf("%v", field.Value)}
		}
		return
	}

	var stubIdentity identity.Identity
	require.NoError(t, faker.FakeData(&stubIdentity))
	stubIdentity.Traits = identity.Traits(`{"emails":["exists@ory.sh"]}`)
	address, err := identity.NewVerifiableEmailAddress("exists@ory.sh", stubIdentity.ID, time.Minute)
	require.NoError(t, err)
	stubIdentity.Addresses = append(stubIdentity.Addresses, *address)
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &stubIdentity))

	for name, tc := range map[string]struct {
		to      string
		subject string
	}{
		"untracked": {to: "does-not-exist@ory.sh", subject: "tried to verify"},
		"tracked":   {to: "exists@ory.sh", subject: "Please verify"},
	} {
		t.Run("case=request verify of "+name+" address", func(t *testing.T) {
			hc := &http.Client{Jar: x.EasyCookieJar(t, nil)}
			rid := string(x.EasyGetBody(t, hc, initURL))
			svr, err := publicClient.Common.GetSelfServiceVerificationRequest(common.
				NewGetSelfServiceVerificationRequestParams().WithHTTPClient(hc).
				WithRequest(rid))
			require.NoError(t, err)

			res, err := hc.PostForm(genForm(t, svr, tc.to))
			require.NoError(t, err)

			assert.Equal(t, redirTS.URL, res.Request.URL.String())
			assert.Equal(t, http.StatusNoContent, res.StatusCode)

			m, err := reg.CourierPersister().LatestQueuedMessage(context.Background())
			require.NoError(t, err)
			assert.Contains(t, m.Subject, tc.subject)
			assert.Equal(t, tc.to, m.Recipient)

			t.Run("case=resubmit", func(t *testing.T) {
				res, err := hc.PostForm(genForm(t, svr, tc.to))
				require.NoError(t, err)
				svr, err := adminClient.Common.GetSelfServiceVerificationRequest(common.
					NewGetSelfServiceVerificationRequestParams().WithRequest(string(x.MustReadAll(res.Body))))
				require.NoError(t, err)

				require.Len(t, svr.Payload.Form.Errors, 1)
				assert.Equal(t, "The request was already completed successfully and can not be retried.", svr.Payload.Form.Errors[0].Message)
			})
		})
	}

	t.Run("case=verify address", func(t *testing.T) {
		hc := &http.Client{Jar: x.EasyCookieJar(t, nil)}
		svr, err := publicClient.Common.GetSelfServiceVerificationRequest(common.
			NewGetSelfServiceVerificationRequestParams().WithHTTPClient(hc).
			WithRequest(string(x.EasyGetBody(t, hc, initURL))))
		require.NoError(t, err)

		_, err = hc.PostForm(genForm(t, svr, "exists@ory.sh"))
		require.NoError(t, err)
		m, err := reg.CourierPersister().LatestQueuedMessage(context.Background())
		require.NoError(t, err)

		match := regexp.MustCompile(`<a href="([^"]+)">`).FindStringSubmatch(m.Body)
		require.Len(t, match, 2)

		res, err := hc.Get(match[1])
		require.NoError(t, err)

		assert.Equal(t, redirTS.URL, res.Request.URL.String())
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	})

	t.Run("case=verify unknown code", func(t *testing.T) {
		hc := &http.Client{Jar: x.EasyCookieJar(t, nil)}
		res, _ := x.EasyGet(t, hc,
			publicTS.URL+strings.ReplaceAll(
				strings.ReplaceAll(verify.PublicVerificationConfirmPath, ":code", "unknown-code"),
				":via", "email"))
		assert.Contains(t, res.Request.URL.String(), verifyTS.URL)

		rid := res.Request.URL.Query().Get("request")
		require.NotEmpty(t, rid)

		svr, err := adminClient.Common.GetSelfServiceVerificationRequest(common.NewGetSelfServiceVerificationRequestParams().WithRequest(rid))
		require.NoError(t, err)
		assert.Equal(t, "The verification code has expired or was otherwise invalid. Please request another code.", svr.Payload.Form.Errors[0].Message)
	})

	t.Run("case=complete expired", func(t *testing.T) {
		hc := &http.Client{Jar: x.EasyCookieJar(t, nil)}
		rid := string(x.EasyGetBody(t, hc, initURL))

		vr, err := reg.VerificationPersister().GetVerifyRequest(context.Background(), x.ParseUUID(rid))
		require.NoError(t, err)
		vr.ExpiresAt = time.Now().Add(-time.Minute)
		require.NoError(t, reg.VerificationPersister().UpdateVerifyRequest(context.Background(), vr))

		svr, err := adminClient.Common.GetSelfServiceVerificationRequest(common.
			NewGetSelfServiceVerificationRequestParams().WithRequest(rid))
		res, err := hc.PostForm(genForm(t, svr, "exists@ory.sh"))
		require.NoError(t, err)
		assert.Contains(t, res.Request.URL.String(), verifyTS.URL+"?request=")
		assert.NotContains(t, res.Request.URL.String(), rid)

		svr, err = adminClient.Common.GetSelfServiceVerificationRequest(common.
			NewGetSelfServiceVerificationRequestParams().WithRequest(res.Request.URL.Query().Get("request")))
		require.NoError(t, err)
		require.Len(t, svr.Payload.Form.Errors, 1)
		assert.Equal(t, "The verification request expired 1.00 minutes ago, please try again.", svr.Payload.Form.Errors[0].Message)
	})
}
