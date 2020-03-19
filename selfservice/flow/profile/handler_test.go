package profile_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/urfave/negroni"

	"github.com/ory/x/pointerx"

	"github.com/ory/x/httpx"

	"github.com/ory/viper"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/client"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
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
		values.Set(pointerx.StringR(f.Name), fmt.Sprintf("%v", f.Value))
	}
	return values
}

func TestUpdateProfile(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
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
	// set this intermediate because kratos needs some valid url for CRUDE operations
	viper.Set(configuration.ViperKeyURLsSelfPublic, "http://example.com")
	viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, "1ns")

	primaryIdentity := &identity.Identity{
		ID: x.NewUUID(),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{"john@doe.com"}, Config: json.RawMessage(`{"hashed_password":"foo"}`)},
		},
		Traits:         identity.Traits(`{"email":"john@doe.com","stringy":"foobar","booly":false,"numby":2.5,"should_long_string":"asdfasdfasdfasdfasfdasdfasdfasdf","should_big_number":2048}`),
		TraitsSchemaID: configuration.DefaultIdentityTraitsSchemaID,
	}

	publicTS, adminTS := func() (*httptest.Server, *httptest.Server) {
		router := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		reg.ProfileManagementHandler().RegisterPublicRoutes(router)
		reg.ProfileManagementHandler().RegisterAdminRoutes(admin)
		route, _ := session.MockSessionCreateHandlerWithIdentity(t, reg, primaryIdentity)
		router.GET("/setSession", route)

		other, _ := session.MockSessionCreateHandlerWithIdentity(t, reg, &identity.Identity{ID: x.NewUUID(), Traits: identity.Traits(`{}`)})
		router.GET("/setSession/other-user", other)
		n := negroni.Classic()
		n.UseHandler(router)
		hh := x.NewTestCSRFHandler(n, reg)
		reg.WithCSRFHandler(hh)
		return httptest.NewServer(hh), httptest.NewServer(admin)
	}()
	defer publicTS.Close()

	viper.Set(configuration.ViperKeyURLsSelfPublic, publicTS.URL)

	primaryUser := func() *http.Client {
		c := session.MockCookieClient(t)
		session.MockHydrateCookieClient(t, c, publicTS.URL+"/setSession")
		return c
	}()

	otherUser := func() *http.Client {
		c := session.MockCookieClient(t)
		session.MockHydrateCookieClient(t, c, publicTS.URL+"/setSession/other-user")
		return c
	}()

	publicClient := client.NewHTTPClientWithConfig(
		nil,
		&client.TransportConfig{Host: urlx.ParseOrPanic(publicTS.URL).Host, BasePath: "/", Schemes: []string{"http"}},
	)

	adminClient := client.NewHTTPClientWithConfig(
		nil,
		&client.TransportConfig{Host: urlx.ParseOrPanic(adminTS.URL).Host, BasePath: "/", Schemes: []string{"http"}},
	)

	makeRequest := func(t *testing.T) *common.GetSelfServiceBrowserProfileManagementRequestOK {
		res, err := primaryUser.Get(publicTS.URL + profile.PublicProfileManagementPath)
		require.NoError(t, err)

		rs, err := publicClient.Common.GetSelfServiceBrowserProfileManagementRequest(
			common.NewGetSelfServiceBrowserProfileManagementRequestParams().WithHTTPClient(primaryUser).
				WithRequest(res.Request.URL.Query().Get("request")),
		)
		require.NoError(t, err)

		return rs
	}

	submitForm := func(t *testing.T, req *common.GetSelfServiceBrowserProfileManagementRequestOK, values url.Values) (string, *common.GetSelfServiceBrowserProfileManagementRequestOK) {
		require.NotNil(t, req.Payload.Methods[profile.FormTraitsID])
		f := req.Payload.Methods[profile.FormTraitsID].Config
		require.NotEmpty(t, f.Action)
		res, err := primaryUser.PostForm(pointerx.StringR(f.Action), values)
		require.NoError(t, err)
		b, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		defer res.Body.Close()
		assert.EqualValues(t, http.StatusNoContent, res.StatusCode, "%s", b)

		assert.Equal(t, ui.URL, res.Request.URL.Scheme+"://"+res.Request.URL.Host)
		assert.Equal(t, "/profile", res.Request.URL.Path, "should end up at the profile URL, used: %s", pointerx.StringR(f.Action))

		rs, err := publicClient.Common.GetSelfServiceBrowserProfileManagementRequest(
			common.NewGetSelfServiceBrowserProfileManagementRequestParams().WithHTTPClient(primaryUser).
				WithRequest(res.Request.URL.Query().Get("request")),
		)
		require.NoError(t, err)
		body, err := json.Marshal(rs.Payload)
		require.NoError(t, err)
		return string(body), rs
	}

	newExpiredRequest := func() *profile.Request {
		return &profile.Request{
			ID:         x.NewUUID(),
			ExpiresAt:  time.Now().Add(-time.Minute),
			IssuedAt:   time.Now().Add(-time.Minute * 2),
			RequestURL: publicTS.URL + login.BrowserLoginPath,
			Identity:   primaryIdentity,
		}
	}

	t.Run("description=call endpoints", func(t *testing.T) {
		pr, ar := x.NewRouterPublic(), x.NewRouterAdmin()
		reg.ProfileManagementHandler().RegisterPublicRoutes(pr)
		reg.ProfileManagementHandler().RegisterAdminRoutes(ar)

		adminTS, publicTS := httptest.NewServer(ar), httptest.NewServer(pr)
		defer adminTS.Close()
		defer publicTS.Close()

		for k, tc := range []*http.Request{
			httpx.MustNewRequest("GET", publicTS.URL+profile.PublicProfileManagementPath, nil, ""),
			httpx.MustNewRequest("GET", publicTS.URL+profile.PublicProfileManagementRequestPath, nil, ""),
			httpx.MustNewRequest("POST", publicTS.URL+profile.PublicProfileManagementUpdatePath, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/x-www-form-urlencoded"),
			httpx.MustNewRequest("POST", publicTS.URL+profile.PublicProfileManagementUpdatePath, strings.NewReader(`{"foo":"bar"}`), "application/json"),
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				res, err := http.DefaultClient.Do(tc)
				require.NoError(t, err)
				assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
			})
		}
	})

	t.Run("daemon=admin", func(t *testing.T) {
		t.Run("description=fetching a non-existent request should return a 404 error", func(t *testing.T) {
			_, err := adminClient.Common.GetSelfServiceBrowserProfileManagementRequest(
				common.NewGetSelfServiceBrowserProfileManagementRequestParams().WithHTTPClient(otherUser).WithRequest("i-do-not-exist"),
			)
			require.Error(t, err)

			require.IsType(t, &common.GetSelfServiceBrowserProfileManagementRequestNotFound{}, err)
			assert.Equal(t, int64(http.StatusNotFound), err.(*common.GetSelfServiceBrowserProfileManagementRequestNotFound).Payload.Error.Code)
		})

		t.Run("description=fetching an expired request returns 410", func(t *testing.T) {
			pr := newExpiredRequest()
			require.NoError(t, reg.ProfileRequestPersister().CreateProfileRequest(context.Background(), pr))

			_, err := adminClient.Common.GetSelfServiceBrowserProfileManagementRequest(
				common.NewGetSelfServiceBrowserProfileManagementRequestParams().WithHTTPClient(primaryUser).WithRequest(pr.ID.String()),
			)
			require.Error(t, err)

			require.IsType(t, &common.GetSelfServiceBrowserProfileManagementRequestGone{}, err, "%+v", err)
			assert.Equal(t, int64(http.StatusGone), err.(*common.GetSelfServiceBrowserProfileManagementRequestGone).Payload.Error.Code)
		})
	})

	t.Run("daemon=public", func(t *testing.T) {
		t.Run("description=fetching a non-existent request should return a 403 error", func(t *testing.T) {
			_, err := publicClient.Common.GetSelfServiceBrowserProfileManagementRequest(
				common.NewGetSelfServiceBrowserProfileManagementRequestParams().WithHTTPClient(otherUser).WithRequest("i-do-not-exist"),
			)
			require.Error(t, err)

			require.IsType(t, &common.GetSelfServiceBrowserProfileManagementRequestForbidden{}, err)
			assert.Equal(t, int64(http.StatusForbidden), err.(*common.GetSelfServiceBrowserProfileManagementRequestForbidden).Payload.Error.Code)
		})

		t.Run("description=fetching an expired request returns 410", func(t *testing.T) {
			pr := newExpiredRequest()
			require.NoError(t, reg.ProfileRequestPersister().CreateProfileRequest(context.Background(), pr))

			_, err := publicClient.Common.GetSelfServiceBrowserProfileManagementRequest(
				common.NewGetSelfServiceBrowserProfileManagementRequestParams().WithHTTPClient(primaryUser).WithRequest(pr.ID.String()),
			)
			require.Error(t, err)

			require.IsType(t, &common.GetSelfServiceBrowserProfileManagementRequestGone{}, err)
			assert.Equal(t, int64(http.StatusGone), err.(*common.GetSelfServiceBrowserProfileManagementRequestGone).Payload.Error.Code)
		})

		t.Run("description=should fail to fetch request if identity changed", func(t *testing.T) {
			res, err := primaryUser.Get(publicTS.URL + profile.PublicProfileManagementPath)
			require.NoError(t, err)

			rid := res.Request.URL.Query().Get("request")
			require.NotEmpty(t, rid)

			_, err = publicClient.Common.GetSelfServiceBrowserProfileManagementRequest(
				common.NewGetSelfServiceBrowserProfileManagementRequestParams().WithHTTPClient(otherUser).WithRequest(rid),
			)
			require.Error(t, err)
			require.IsType(t, &common.GetSelfServiceBrowserProfileManagementRequestForbidden{}, err)
			assert.EqualValues(t, int64(http.StatusForbidden), err.(*common.GetSelfServiceBrowserProfileManagementRequestForbidden).Payload.Error.Code, "should return a 403 error because the identities from the cookies do not match")
		})

		t.Run("description=should fail to post data if CSRF is missing", func(t *testing.T) {
			rs := makeRequest(t)
			require.NotEmpty(t, rs.Payload.Methods[profile.FormTraitsID])
			f := rs.Payload.Methods[profile.FormTraitsID].Config
			res, err := primaryUser.PostForm(pointerx.StringR(f.Action), url.Values{})
			require.NoError(t, err)
			assert.EqualValues(t, 400, res.StatusCode, "should return a 400 error because CSRF token is not set")
		})

		t.Run("description=should redirect to profile management ui and /profiles/requests?request=... should come back with the right information", func(t *testing.T) {
			res, err := primaryUser.Get(publicTS.URL + profile.PublicProfileManagementPath)
			require.NoError(t, err)

			assert.Equal(t, ui.URL, res.Request.URL.Scheme+"://"+res.Request.URL.Host)
			assert.Equal(t, "/profile", res.Request.URL.Path, "should end up at the profile URL")

			rid := res.Request.URL.Query().Get("request")
			require.NotEmpty(t, rid)

			pr, err := publicClient.Common.GetSelfServiceBrowserProfileManagementRequest(
				common.NewGetSelfServiceBrowserProfileManagementRequestParams().WithHTTPClient(primaryUser).WithRequest(rid),
			)
			require.NoError(t, err, "%s", rid)

			assert.Equal(t, rid, string(pr.Payload.ID))
			assert.NotEmpty(t, pr.Payload.Identity)
			assert.Equal(t, primaryIdentity.ID.String(), string(pr.Payload.Identity.ID))
			assert.JSONEq(t, string(primaryIdentity.Traits), x.MustEncodeJSON(t, pr.Payload.Identity.Traits))
			assert.Equal(t, primaryIdentity.TraitsSchemaID, pointerx.StringR(pr.Payload.Identity.TraitsSchemaID))
			assert.Equal(t, publicTS.URL+profile.PublicProfileManagementPath, pointerx.StringR(pr.Payload.RequestURL))

			found := false

			require.NotNil(t, pr.Payload.Methods[profile.FormTraitsID].Config)
			f := pr.Payload.Methods[profile.FormTraitsID].Config

			for i := range f.Fields {
				if pointerx.StringR(f.Fields[i].Name) == form.CSRFTokenName {
					found = true
					require.NotEmpty(t, f.Fields[i])
					f.Fields = append(f.Fields[:i], f.Fields[i+1:]...)
					break
				}
			}
			require.True(t, found)

			assert.EqualValues(t, &models.Form{
				Action: pointerx.String(publicTS.URL + profile.PublicProfileManagementUpdatePath + "?request=" + rid),
				Method: pointerx.String("POST"),
				Fields: models.FormFields{
					&models.FormField{Name: pointerx.String("traits.email"), Type: pointerx.String("text"), Value: "john@doe.com", Disabled: true},
					&models.FormField{Name: pointerx.String("traits.stringy"), Type: pointerx.String("text"), Value: "foobar"},
					&models.FormField{Name: pointerx.String("traits.numby"), Type: pointerx.String("number"), Value: json.Number("2.5")},
					&models.FormField{Name: pointerx.String("traits.booly"), Type: pointerx.String("checkbox"), Value: false},
					&models.FormField{Name: pointerx.String("traits.should_big_number"), Type: pointerx.String("number"), Value: json.Number("2048")},
					&models.FormField{Name: pointerx.String("traits.should_long_string"), Type: pointerx.String("text"), Value: "asdfasdfasdfasdfasfdasdfasdfasdf"},
				},
			}, f)
		})

		t.Run("description=should come back with form errors if some profile data is invalid", func(t *testing.T) {
			rs := makeRequest(t)
			require.NotNil(t, rs.Payload.Methods[profile.FormTraitsID].Config, "%+v", rs.Payload)
			values := fieldsToURLValues(rs.Payload.Methods[profile.FormTraitsID].Config.Fields)
			values.Set("traits.should_long_string", "too-short")
			values.Set("traits.stringy", "bazbar") // it should still override new values!
			actual, _ := submitForm(t, rs, values)

			assert.NotEmpty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==csrf_token).value").String(), "%s", actual)
			assert.Equal(t, "too-short", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).value").String(), "%s", actual)
			assert.Equal(t, "bazbar", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.stringy).value").String(), "%s", actual)
			assert.Equal(t, "2.5", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).value").String(), "%s", actual)
			assert.Equal(t, "length must be >= 25, but got 9", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).errors.0.message").String(), "%s", actual)
		})

		t.Run("description=should update protected field with sudo mode", func(t *testing.T) {
			viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, "1m")
			defer viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, "1ns")

			rs := makeRequest(t)
			require.NotEmpty(t, rs.Payload.Methods[profile.FormTraitsID])
			newEmail := "not-john-doe@mail.com"
			values := fieldsToURLValues(rs.Payload.Methods[profile.FormTraitsID].Config.Fields)
			values.Set("traits.email", newEmail)
			actual, response := submitForm(t, rs, values)
			assert.True(t, pointerx.BoolR(response.Payload.UpdateSuccessful), "%s", actual)

			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).errors").Value(), "%s", actual)

			assert.Equal(t, newEmail, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.email).value").Value(), "%s", actual)

			assert.Equal(t, "foobar", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.stringy).value").String(), "%s", actual) // sanity check if original payload is still here
		})

		t.Run("description=should come back with form errors if trying to update protected field without sudo mode", func(t *testing.T) {
			rs := makeRequest(t)
			require.NotEmpty(t, rs.Payload.Methods[profile.FormTraitsID])
			f := rs.Payload.Methods[profile.FormTraitsID].Config
			values := fieldsToURLValues(f.Fields)
			values.Set("traits.email", "not-john-doe")
			res, err := primaryUser.PostForm(pointerx.StringR(f.Action), values)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Contains(t, res.Request.URL.String(), errTs.URL)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Contains(t, gjson.Get(string(body), "0.reason").String(), "A field was modified that updates one or more credentials-related settings", "%s", body)
		})

		t.Run("description=should retry with invalid payloads multiple times before succeeding", func(t *testing.T) {
			t.Run("flow=fail first update", func(t *testing.T) {
				rs := makeRequest(t)
				require.NotEmpty(t, rs.Payload.Methods[profile.FormTraitsID])
				f := rs.Payload.Methods[profile.FormTraitsID].Config
				values := fieldsToURLValues(f.Fields)
				values.Set("traits.should_big_number", "1")
				actual, response := submitForm(t, rs, values)
				assert.False(t, pointerx.BoolR(response.Payload.UpdateSuccessful), "%s", actual)

				assert.Equal(t, "1", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).value").String(), "%s", actual)
				assert.Equal(t, "must be >= 1200 but found 1", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).errors.0.message").String(), "%s", actual)

				assert.Equal(t, "foobar", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.stringy).value").String(), "%s", actual) // sanity check if original payload is still here
			})

			t.Run("flow=fail second update", func(t *testing.T) {
				rs := makeRequest(t)
				require.NotEmpty(t, rs.Payload.Methods[profile.FormTraitsID])
				f := rs.Payload.Methods[profile.FormTraitsID].Config
				values := fieldsToURLValues(f.Fields)
				values.Del("traits.should_big_number")
				values.Set("traits.should_long_string", "short")
				values.Set("traits.numby", "this-is-not-a-number")
				actual, response := submitForm(t, rs, values)
				assert.False(t, pointerx.BoolR(response.Payload.UpdateSuccessful), "%s", actual)

				assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).errors.0.message").String(), "%s", actual)
				assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).value").String(), "%s", actual)

				assert.Equal(t, "short", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).value").String(), "%s", actual)
				assert.Equal(t, "length must be >= 25, but got 5", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).errors.0.message").String(), "%s", actual)

				assert.Equal(t, "this-is-not-a-number", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).value").String(), "%s", actual)
				assert.Equal(t, "expected number, but got string", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).errors.0.message").String(), "%s", actual)

				assert.Equal(t, "foobar", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.stringy).value").String(), "%s", actual) // sanity check if original payload is still here
			})

			t.Run("flow=succeed with final request", func(t *testing.T) {
				rs := makeRequest(t)
				require.NotEmpty(t, rs.Payload.Methods[profile.FormTraitsID])
				f := rs.Payload.Methods[profile.FormTraitsID].Config
				values := fieldsToURLValues(f.Fields)
				// set email to the one that is in the db as it should not be modified
				values.Set("traits.email", "not-john-doe@mail.com")
				values.Set("traits.numby", "15")
				values.Set("traits.should_big_number", "9001")
				values.Set("traits.should_long_string", "this is such a long string, amazing stuff!")
				actual, response := submitForm(t, rs, values)
				assert.True(t, pointerx.BoolR(response.Payload.UpdateSuccessful), "%s", actual)

				assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).errors").Value(), "%s", actual)
				assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).errors").Value(), "%s", actual)
				assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).errors").Value(), "%s", actual)

				assert.Equal(t, 15.0, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).value").Value(), "%s", actual)
				assert.Equal(t, 9001.0, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).value").Value(), "%s", actual)
				assert.Equal(t, "this is such a long string, amazing stuff!", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).value").Value(), "%s", actual)

				assert.Equal(t, "foobar", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.stringy).value").String(), "%s", actual) // sanity check if original payload is still here
			})

			t.Run("flow=try another update with invalid data", func(t *testing.T) {
				rs := makeRequest(t)
				require.NotEmpty(t, rs.Payload.Methods[profile.FormTraitsID])
				f := rs.Payload.Methods[profile.FormTraitsID].Config
				values := fieldsToURLValues(f.Fields)
				values.Set("traits.should_long_string", "short")
				actual, response := submitForm(t, rs, values)
				assert.False(t, pointerx.BoolR(response.Payload.UpdateSuccessful), "%s", actual)
			})
		})
	})
}
