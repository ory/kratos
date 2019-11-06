package password_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice"
	. "github.com/ory/kratos/selfservice/password"
	"github.com/ory/kratos/x"
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
		viper.Set(configuration.ViperKeySelfServiceRegistrationAfterConfig+"."+string(identity.CredentialsTypePassword), hookConfig(returnTs.URL+"/return-ts"))
		viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")

		var newRegistrationRequest = func(t *testing.T, exp time.Duration) *selfservice.RegistrationRequest {
			rr := &selfservice.RegistrationRequest{
				Request: &selfservice.Request{
					ID: "request-" + uuid.New().String(), IssuedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(exp), RequestURL: ts.URL,
					RequestHeaders: http.Header{}, Methods: map[identity.CredentialsType]*selfservice.DefaultRequestMethod{
						identity.CredentialsTypePassword: {
							Method: identity.CredentialsTypePassword, Config: &RequestMethodConfig{
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
			res, err := client.Post(ts.URL+RegistrationPath+"?request="+rid, "application/x-www-form-urlencoded", strings.NewReader(body))
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

		t.Run("case=should return an error because the request is expired", func(t *testing.T) {
			rr := newRegistrationRequest(t, -time.Minute)
			body, res := makeRequest(t, rr.ID, "", http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID, gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.errors.0.message").String(), "expired", "%s", body)
		})

		t.Run("case=should return an error because the password failed validation", func(t *testing.T) {
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-4"},
				"password":        {"password"},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID, gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			fieldNameSet(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.password.error").String(), "data breaches and must no longer be used.", "%s", body)
		})

		t.Run("case=should return an error because not passing validation", func(t *testing.T) {
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-5"},
				"password":        {uuid.New().String()},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID, gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			fieldNameSet(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.traits\\.foobar.error").String(), "foobar is required", "%s", body)
		})

		t.Run("case=should fail because schema did not specify an identifier", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/missing-identifier.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-6"},
				"password":        {uuid.New().String()},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "error-ts")
			assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "No login identifiers", "%s", body)
		})

		t.Run("case=should fail because schema did not specify an identifier", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/i-do-not-exist.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-7"},
				"password":        {uuid.New().String()},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "error-ts")
			assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "Unable to parse JSON schema", "%s", body)
		})

		t.Run("case=should pass and set up a session", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-8"},
				"password":        {uuid.New().String()},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "return-ts")
			assert.Equal(t, `registration-identifier-8`, gjson.GetBytes(body, "identity.traits.username").String(), "%s", body)
		})

		t.Run("case=should fail to register the same user again", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-8"},
				"password":        {uuid.New().String()},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.errors.0.message").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
		})

		t.Run("case=should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")
			rr := &selfservice.RegistrationRequest{
				Request: &selfservice.Request{
					ID:        "request-9",
					ExpiresAt: time.Now().Add(time.Minute),
					Methods: map[identity.CredentialsType]*selfservice.DefaultRequestMethod{
						identity.CredentialsTypePassword: {
							Method: identity.CredentialsTypePassword,
							Config: &RequestMethodConfig{
								Action: "/action", Errors: []selfservice.FormError{{Message: "some error"}},
								Fields: map[string]selfservice.FormField{
									"traits.foo": {
										Name: "traits.foo", Value: "bar", Type: "text",
										Error: &selfservice.FormError{Message: "bar"},
									},
									"password": {Name: "password"},
								},
							},
						},
					},
				},
			}
			require.NoError(t, reg.RegistrationRequestManager().CreateRegistrationRequest(context.Background(), rr))
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-9"},
				"password":        {uuid.New().String()},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, "request-9", gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			fieldNameSet(t, body, "password", "csrf_token", "traits.username")

			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.traits\\.foo.value"))
			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.traits\\.foo.error"))
			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.error"))
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.traits\\.foobar.error").String(), "foobar is required", "%s", body)
		})

		t.Run("case=should work even if password is just numbers", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-10"},
				"password":        {"93172388957812344432"},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "return-ts")
			assert.Equal(t, `registration-identifier-10`, gjson.GetBytes(body, "identity.traits.username").String(), "%s", body)
		})
	})

	t.Run("method=PopulateSignUpMethod", func(t *testing.T) {
		conf, reg := internal.NewMemoryRegistry(t)
		s := NewStrategy(reg, conf).WithTokenGenerator(func(r *http.Request) string {
			return "nosurf"
		})

		viper.Set(configuration.ViperKeyURLsSelfPublic, urlx.ParseOrPanic("https://foo/"))

		sr := selfservice.NewRegistrationRequest(time.Minute, &http.Request{URL: urlx.ParseOrPanic("/")})
		require.NoError(t, s.PopulateRegistrationMethod(&http.Request{}, sr))
		assert.EqualValues(t, &selfservice.DefaultRequestMethod{
			Method: identity.CredentialsTypePassword,
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
		}, sr.Methods[identity.CredentialsTypePassword])
	})
}
