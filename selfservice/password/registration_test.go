package password_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/x/urlx"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
	"github.com/ory/hive/selfservice"
	. "github.com/ory/hive/selfservice/password"
	"github.com/ory/hive/x"
)

// fieldNameSet checks if the fields have the right "name" set.
func fieldNameSet(t *testing.T, body []byte, fields ...string) {
	for _, f := range fields {
		fieldid := strings.Replace(f, ".", "\\.", -1) // we need to escape this because otherwise json path will interpret this as a nested object (it is not).
		assert.Equal(t, gjson.GetBytes(body, fmt.Sprintf("methods.password.config.fields.%s.name", fieldid)).String(), f, "%s", body)
	}
}

func TestRegistration(t *testing.T) {
	t.Run("case=registration", func(t *testing.T) {

		conf, reg := internal.NewMemoryRegistry(t)
		s := NewStrategy(reg, conf).WithTokenGenerator(func(r *http.Request) string {
			return "nosurf"
		})

		router := x.NewRouterPublic()
		s.SetRoutes(router)

		ts := httptest.NewServer(router)
		defer ts.Close()

		errTs, uiTs, returnTs := newErrTs(t, reg), httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			e, err := reg.RegistrationRequestManager().GetRegistrationRequest(r.Context(), r.URL.Query().Get("request"))
			require.NoError(t, err)
			reg.Writer().Write(w, r, e)
		})), newReturnTs(t, reg)
		defer errTs.Close()
		defer uiTs.Close()
		defer returnTs.Close()

		viper.Set(configuration.ViperKeyURLsError, errTs.URL+"/error-ts")
		viper.Set(configuration.ViperKeyURLsRegistration, uiTs.URL+"/signup-ts")
		viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
		viper.Set(configuration.ViperKeySelfServiceRegistrationAfterConfig+"."+string(CredentialsType), hookConfig(returnTs.URL+"/return-ts"))
		viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")

		var newRegistrationRequest = func(t *testing.T, exp time.Duration) *selfservice.RegistrationRequest {
			rr := &selfservice.RegistrationRequest{
				Request: &selfservice.Request{
					ID: "request-" + uuid.New().String(), IssuedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(exp), RequestURL: ts.URL,
					RequestHeaders: http.Header{}, Methods: map[identity.CredentialsType]*selfservice.DefaultRequestMethod{
						CredentialsType: {
							Method: CredentialsType, Config: &RequestMethodConfig{
								Action: "/action", Fields: selfservice.FormFields{
									"password":   {Name: "password", Type: "password", Required: true},
									"csrf_token": {Name: "csrf_token", Type: "hidden", Required: true, Value: "csrf-token"},
								},
							},
						},
					},
				},
			}
			require.NoError(t, reg.RegistrationRequestManager().CreateRegistrationRequest(context.Background(), rr))
			return rr
		}

		var makeRequest = func(t *testing.T, rid, body string, expectedStatusCode int) ([]byte, *http.Response) {
			jar, _ := cookiejar.New(&cookiejar.Options{})
			client := http.Client{Jar: jar}
			res, err := client.Post(ts.URL+RegistrationPath+"?request=request-"+rid, "application/x-www-form-urlencoded", strings.NewReader(body))
			require.NoError(t, err)
			result, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)
			require.EqualValues(t, expectedStatusCode, res.StatusCode, "Request: %+v\n\t\tResponse: %s", res.Request, res)
			return result, res
		}

		t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, "14=)=!(%)$/ZP()GHIÖ", http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "error-ts")
			assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Bad Request", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "invalid URL escape", "%s", body)
		})

		t.Run("case=should show the error ui because the request id is missing", func(t *testing.T) {
			_ = newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, "does-not-exist", "", http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "error-ts")
			assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Not Found", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "Unable to find request", "%s", body)
		})
	})

	/*
		for k, tc := range []struct {
			d       string
			ar      *selfservice.RegistrationRequest
			payload string
			schema  string
			rid     string
			assert  func(t *testing.T, r *http.Response)
		}{
			{
				d:  "should return an error because the request does not exist",
				ar: newRegistrationRequest("1", time.Minute),
				payload: url.Values{
					"traits.username": {"registration-identifier-1"},
					"password":        {"password"},
				}.Encode(),
				rid: "does-not-exist",
				assert: func(t *testing.T, r *http.Response) {
					assert.Contains(t, r.Request.URL.Path, "error-ts")
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)

					assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "0.code").Int(), "%s", body)
					assert.Equal(t, "Not Found", gjson.GetBytes(body, "0.status").String(), "%s", body)
					assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "request-does-not-exist", "%s", body)
				},
			},
			{
				d:  "should return an error because the request is expired",
				ar: newRegistrationRequest("2", -time.Minute),
				payload: url.Values{
					"traits.username": {"registration-identifier-2"},
					"password":        {"password"},
				}.Encode(),
				rid: "2",
				assert: func(t *testing.T, r *http.Response) {
					assert.Contains(t, r.Request.URL.Path, "/signup-ts")
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)

					assert.Equal(t, "request-2", gjson.GetBytes(body, "id").String(), "%s", body)
					assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
					assert.Contains(t, gjson.GetBytes(body, "methods.password.config.errors.0.message").String(), "expired", "%s", body)
				},
			},
			{
				d:  "should return an error because the password failed validation",
				ar: newRegistrationRequest("4", time.Minute),
				payload: url.Values{
					"traits.username": {"registration-identifier-4"},
					"password":        {"password"},
					"traits.foobar":   {"bar"},
				}.Encode(),
				rid: "4",
				assert: func(t *testing.T, r *http.Response) {
					assert.Contains(t, r.Request.URL.Path, "signup-ts")
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)

					// Let's ensure that the payload is being propagated properly.
					assert.Equal(t, "request-4", gjson.GetBytes(body, "id").String(), "%s", body)
					assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
					fieldNameSet(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
					assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.password.error").String(), "data breaches and must no longer be used.", "%s", body)
				},
			},
			{
				d:   "should return an error because not passing validation",
				ar:  newRegistrationRequest("5", time.Minute),
				rid: "5",
				payload: url.Values{
					"traits.username": {"registration-identifier-5"},
					"password":        {uuid.New().String()},
				}.Encode(),
				assert: func(t *testing.T, r *http.Response) {
					assert.Contains(t, r.Request.URL.Path, "signup-ts")
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)

					// Let's ensure that the payload is being propagated properly.
					assert.Equal(t, "request-5", gjson.GetBytes(body, "id").String(), "%s", body)
					assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
					fieldNameSet(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
					assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.traits\\.foobar.error").String(), "foobar is required", "%s", body)
				},
			},
			{
				d:   "should fail because schema did not specify an identifier",
				ar:  newRegistrationRequest("6", time.Minute),
				rid: "6",
				payload: url.Values{
					"traits.username": {"registration-identifier-6"},
					"password":        {uuid.New().String()},
					"traits.foobar":   {"bar"},
				}.Encode(),
				schema: "file://./stub/missing-identifier.schema.json",
				assert: func(t *testing.T, r *http.Response) {
					assert.Contains(t, r.Request.URL.Path, "error-ts")
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "0.code").Int(), "%s", body)
					assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "0.status").String(), "%s", body)
					assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "No login identifiers", "%s", body)
				},
			},
			{
				d:  "should fail because schema does not exist",
				ar: newRegistrationRequest("7", time.Minute),
				payload: url.Values{
					"traits.username": {"registration-identifier-7"},
					"password":        {uuid.New().String()},
					"traits.foobar":   {"bar"},
				}.Encode(),
				rid:    "7",
				schema: "file://./stub/i-do-not-exist.schema.json",
				assert: func(t *testing.T, r *http.Response) {
					assert.Contains(t, r.Request.URL.Path, "error-ts")
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "0.code").Int(), "%s", body)
					assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "0.status").String(), "%s", body)
					assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "Unable to parse JSON schema", "%s", body)
				},
			},
			{
				d:   "should pass and set up a session",
				ar:  newRegistrationRequest("8", time.Minute),
				rid: "8",
				payload: url.Values{
					"traits.username": {"registration-identifier-8"},
					"password":        {uuid.New().String()},
					"traits.foobar":   {"bar"},
				}.Encode(),
				assert: func(t *testing.T, r *http.Response) {
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Contains(t, r.Request.URL.Path, "return-ts")
					assert.Equal(t, `["registration-identifier-8"]`, gjson.GetBytes(body, "identity.credentials.password.identifiers").String(), "%s", body)
				},
			},
			{
				d:   "should return an error because not passing validation and reset previous errors and values",
				rid: "9",
				ar: &selfservice.RegistrationRequest{
					Request: &selfservice.Request{
						ID:        "request-9",
						ExpiresAt: time.Now().Add(time.Minute),
						Methods: map[identity.CredentialsType]*selfservice.DefaultRequestMethod{
							CredentialsType: {
								Method: CredentialsType,
								Config: &RequestMethodConfig{
									Action: "/action",
									Errors: []selfservice.FormError{{Message: "some error"}},
									Fields: map[string]selfservice.FormField{
										"traits.foo": {
											Name:  "traits.foo",
											Value: "bar",
											Error: &selfservice.FormError{Message: "bar"},
											Type:  "text",
										},
										"password": {
											Name: "password",
										},
									},
								},
							},
						},
					},
				},
				payload: url.Values{
					"traits.username": {"registration-identifier-9"},
					"password":        {uuid.New().String()},
				}.Encode(),
				assert: func(t *testing.T, r *http.Response) {
					assert.Contains(t, r.Request.URL.Path, "signup-ts")
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)

					// Let's ensure that the payload is being propagated properly.
					assert.Equal(t, "request-9", gjson.GetBytes(body, "id").String(), "%s", body)
					assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
					fieldNameSet(t, body, "password", "csrf_token", "traits.username")

					assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.traits\\.foo.value"))
					assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.traits\\.foo.error"))
					assert.Empty(t, gjson.GetBytes(body, "methods.password.config.error"))
					assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.traits\\.foobar.error").String(), "foobar is required", "%s", body)
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
				conf, reg := internal.NewMemoryRegistry(t)
				s := NewStrategy(reg, conf).WithTokenGenerator(func(r *http.Request) string {
					return "nosurf"
				})

				router := x.NewRouterPublic()
				s.SetRoutes(router)

				ts := httptest.NewServer(router)
				defer ts.Close()

				errTs, uiTs, returnTs := newErrTs(t, reg), httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					e, err := reg.RegistrationRequestManager().GetRegistrationRequest(r.Context(), r.URL.Query().Get("request"))
					require.NoError(t, err)
					reg.Writer().Write(w, r, e)
				})), newReturnTs(t, reg)
				defer errTs.Close()
				defer uiTs.Close()
				defer returnTs.Close()

				viper.Set(configuration.ViperKeyURLsError, errTs.URL+"/error-ts")
				viper.Set(configuration.ViperKeyURLsRegistration, uiTs.URL+"/signup-ts")
				viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
				viper.Set(configuration.ViperKeySelfServiceRegistrationAfterConfig+"."+string(CredentialsType), hookConfig(returnTs.URL+"/return-ts"))
				if tc.schema == "" {
					tc.schema = "file://./stub/registration.schema.json"
				}
				viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, tc.schema)

				tc.ar.RequestURL = ts.URL
				require.NoError(t, reg.RegistrationRequestManager().CreateRegistrationRequest(context.Background(), tc.ar))

				c := ts.Client()
				c.Jar, _ = cookiejar.New(&cookiejar.Options{})

				res, err := c.Post(ts.URL+RegistrationPath+"?request=request-"+tc.rid, "application/x-www-form-urlencoded", strings.NewReader(tc.payload))
				require.NoError(t, err)
				defer res.Body.Close()
				require.EqualValues(t, http.StatusOK, res.StatusCode, "Request: %+v\n\t\tResponse: %s", res.Request, res)

				tc.assert(t, res)
			})
		}
	*/

	t.Run("method=PopulateSignUpMethod", func(t *testing.T) {
		conf, reg := internal.NewMemoryRegistry(t)
		s := NewStrategy(reg, conf).WithTokenGenerator(func(r *http.Request) string {
			return "nosurf"
		})
		viper.Set(configuration.ViperKeyURLsSelfPublic, urlx.ParseOrPanic("https://foo/"))

		sr := selfservice.NewRegistrationRequest(time.Minute, &http.Request{URL: urlx.ParseOrPanic("/")})
		require.NoError(t, s.PopulateRegistrationMethod(&http.Request{}, sr))
		assert.EqualValues(t, &selfservice.DefaultRequestMethod{
			Method: CredentialsType,
			Config: &RequestMethodConfig{
				Action: "https://foo" + RegistrationPath + "?request=" + sr.ID,
				Fields: selfservice.FormFields{
					"password": {
						Name:     "password",
						Type:     "password",
						Required: true,
					},
					"csrf_token": {
						Name:     "csrf_token",
						Type:     "hidden",
						Required: true,
						Value:    "nosurf",
					},
				},
			},
		}, sr.Methods[CredentialsType])
	})
}
