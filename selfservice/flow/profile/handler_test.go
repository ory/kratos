package profile_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/urfave/negroni"

	"github.com/ory/x/httpx"

	"github.com/ory/viper"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/sdk/go/kratos/client"
	"github.com/ory/kratos/sdk/go/kratos/client/public"
	"github.com/ory/kratos/sdk/go/kratos/models"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func fieldsToURLValues(ff models.FormFields) url.Values {
	values := url.Values{}
	for _, f := range ff {
		values.Set(f.Name, fmt.Sprintf("%v", f.Value))
	}
	return values
}

func TestUpdateProfile(t *testing.T) {
	_, reg := internal.NewMemoryRegistry(t)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")

	ui := func() *httptest.Server {
		router := httprouter.New()
		router.GET("/profile", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			w.WriteHeader(http.StatusNoContent)
		})
		router.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			w.WriteHeader(http.StatusUnauthorized)
		})
		router.GET("/error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			w.WriteHeader(http.StatusUnauthorized)
		})
		return httptest.NewServer(router)
	}()
	defer ui.Close()

	errTs := errorx.NewErrorTestServer(t, reg)
	defer errTs.Close()

	viper.Set(configuration.ViperKeyURLsError, errTs.URL)
	viper.Set(configuration.ViperKeyURLsProfile, ui.URL+"/profile")
	viper.Set(configuration.ViperKeyURLsLogin, ui.URL+"/login")

	primaryIdentity := &identity.Identity{
		ID:              uuid.New().String(),
		Credentials:     nil,
		TraitsSchemaURL: "file://./stub/identity.schema.json",
		Traits:          json.RawMessage(`{"stringy":"foobar","booly":false,"numby":2.5}`),
	}

	kratos := func() *httptest.Server {
		router := x.NewRouterPublic()
		reg.ProfileManagementHandler().RegisterPublicRoutes(router)
		route, _ := session.MockSessionCreateHandlerWithIdentity(t, reg, primaryIdentity)
		router.GET("/setSession", route)

		other, _ := session.MockSessionCreateHandlerWithIdentity(t, reg, &identity.Identity{ID: uuid.New().String(), TraitsSchemaURL: "file://./stub/identity.schema.json", Traits: json.RawMessage(`{}`)})
		router.GET("/setSession/other-user", other)
		n := negroni.Classic()
		n.UseHandler(router)
		return httptest.NewServer(nosurf.New(n))
	}()
	defer kratos.Close()

	viper.Set(configuration.ViperKeyURLsSelfPublic, kratos.URL)

	primaryUser := func() *http.Client {
		c := session.MockCookieClient(t)
		session.MockHydrateCookieClient(t, c, kratos.URL+"/setSession")
		return c
	}()

	otherUser := func() *http.Client {
		c := session.MockCookieClient(t)
		session.MockHydrateCookieClient(t, c, kratos.URL+"/setSession/other-user")
		return c
	}()

	kratosClient := client.NewHTTPClientWithConfig(
		nil,
		&client.TransportConfig{Host: urlx.ParseOrPanic(kratos.URL).Host, BasePath: "/", Schemes: []string{"http"}},
	)

	makeRequest := func(t *testing.T) *public.GetProfileManagementRequestOK {
		res, err := primaryUser.Get(kratos.URL + profile.BrowserProfilePath)
		require.NoError(t, err)

		rs, err := kratosClient.Public.GetProfileManagementRequest(
			public.NewGetProfileManagementRequestParams().WithHTTPClient(primaryUser).
				WithRequest(res.Request.URL.Query().Get("request")),
		)
		require.NoError(t, err)

		return rs
	}

	t.Run("description=call endpoints without session set results in an error", func(t *testing.T) {
		kratos := func() *httptest.Server {
			router := x.NewRouterPublic()
			reg.ProfileManagementHandler().RegisterPublicRoutes(router)
			return httptest.NewServer(router)
		}()
		defer kratos.Close()

		for k, tc := range []*http.Request{
			httpx.MustNewRequest("GET", kratos.URL+profile.BrowserProfilePath, nil, ""),
			httpx.MustNewRequest("GET", kratos.URL+profile.BrowserProfileRequestPath, nil, ""),
			httpx.MustNewRequest("POST", kratos.URL+profile.BrowserProfilePath, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/x-www-form-urlencoded"),
			httpx.MustNewRequest("POST", kratos.URL+profile.BrowserProfilePath, strings.NewReader(`{"foo":"bar"}`), "application/json"),
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				res, err := http.DefaultClient.Do(tc)
				require.NoError(t, err)
				assert.EqualValues(t, 401, res.StatusCode)
			})
		}
	})

	t.Run("description=fetching a non-existent request should return a 404 error", func(t *testing.T) {
		_, err := kratosClient.Public.GetProfileManagementRequest(
			public.NewGetProfileManagementRequestParams().WithHTTPClient(otherUser).WithRequest("i-do-not-exist"),
		)
		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, err.(*runtime.APIError).Code)
	})

	t.Run("description=should fail to fetch request if identity changed", func(t *testing.T) {
		res, err := primaryUser.Get(kratos.URL + profile.BrowserProfilePath)
		require.NoError(t, err)

		rid := res.Request.URL.Query().Get("request")
		require.NotEmpty(t, rid)

		_, err = kratosClient.Public.GetProfileManagementRequest(
			public.NewGetProfileManagementRequestParams().WithHTTPClient(otherUser).WithRequest(rid),
		)
		require.Error(t, err)
		assert.EqualValues(t, 403, err.(*runtime.APIError).Code, "should return a 403 error because the identities from the cookies do not match")
	})

	t.Run("description=should fail to post data if CSRF is missing", func(t *testing.T) {
		rs := makeRequest(t)
		f := rs.Payload.Form
		res, err := primaryUser.PostForm(f.Action, url.Values{})
		require.NoError(t, err)
		assert.EqualValues(t, 400, res.StatusCode, "should return a 400 error because CSRF token is not set")
	})

	t.Run("description=should redirect to profile management ui and /profiles/requests?request=... should come back with the right information", func(t *testing.T) {
		res, err := primaryUser.Get(kratos.URL + profile.BrowserProfilePath)
		require.NoError(t, err)

		assert.Equal(t, ui.URL, res.Request.URL.Scheme+"://"+res.Request.URL.Host)
		assert.Equal(t, "/profile", res.Request.URL.Path, "should end up at the profile URL")

		rid := res.Request.URL.Query().Get("request")
		require.NotEmpty(t, rid)

		pr, err := kratosClient.Public.GetProfileManagementRequest(
			public.NewGetProfileManagementRequestParams().WithHTTPClient(primaryUser).WithRequest(rid),
		)
		require.NoError(t, err, "%s", rid)

		assert.Equal(t, rid, pr.Payload.ID)
		assert.NotEmpty(t, pr.Payload.Identity)
		assert.Empty(t, pr.Payload.Identity.Credentials)
		assert.Equal(t, primaryIdentity.ID, *(pr.Payload.Identity.ID))
		assert.JSONEq(t, string(primaryIdentity.Traits), x.MustEncodeJSON(t, pr.Payload.Identity.Traits))
		assert.Equal(t, primaryIdentity.TraitsSchemaURL, pr.Payload.Identity.TraitsSchemaURL)
		assert.Equal(t, kratos.URL+profile.BrowserProfilePath, pr.Payload.RequestURL)

		require.NotEmpty(t, pr.Payload.Form.Fields[form.CSRFTokenName])
		delete(pr.Payload.Form.Fields, form.CSRFTokenName)
		assert.Equal(t, &models.Form{
			Action: kratos.URL + profile.BrowserProfilePath,
			Method: "POST",
			Fields: models.FormFields{
				"request":        models.FormField{Name: "request", Required: true, Type: "hidden", Value: rid},
				"traits.stringy": models.FormField{Name: "traits.stringy", Required: false, Type: "text", Value: "foobar"},
				"traits.numby":   models.FormField{Name: "traits.numby", Required: false, Type: "number", Value: json.Number("2.5")},
				"traits.booly":   models.FormField{Name: "traits.booly", Required: false, Type: "checkbox", Value: false},
			},
		}, pr.Payload.Form)
	})

	t.Run("description=should come back with form errors if some profile data is invalid", func(t *testing.T) {
		rs := makeRequest(t)
		f := rs.Payload.Form
		values := fieldsToURLValues(f.Fields)
		values.Set("traits.should_long_string", "too-short")
		t.Logf("%+v", values)
		res, err := primaryUser.PostForm(f.Action, values)
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusNoContent, res.StatusCode)

		assert.Equal(t, ui.URL, res.Request.URL.Scheme+"://"+res.Request.URL.Host)
		assert.Equal(t, "/profile", res.Request.URL.Path, "should end up at the profile URL")

		rs, err = kratosClient.Public.GetProfileManagementRequest(
			public.NewGetProfileManagementRequestParams().WithHTTPClient(primaryUser).
				WithRequest(res.Request.URL.Query().Get("request")),
		)
		require.NoError(t, err)

		actual, err := json.Marshal(rs.Payload)
		require.NoError(t, err)

		assert.Equal(t, "too-short", gjson.Get(string(actual), "form.fields.traits\\.should_long_string.value").String(), "%s", actual)
		assert.Equal(t, "traits.should_long_string: String length must be greater than or equal to 25", gjson.Get(string(actual), "form.fields.traits\\.should_long_string.errors.0.message").String(), "%s", actual)
	})
}
