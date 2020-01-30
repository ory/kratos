package password_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/x"
)

func checkFormContent(t *testing.T, body []byte, requiredFields ...string) {
	fieldNameSet(t, body, requiredFields)
	outdatedFieldsDoNotExist(t, body)
	formMethodIsPOST(t, body)
}

// fieldNameSet checks if the fields have the right "name" set.
func fieldNameSet(t *testing.T, body []byte, fields []string) {
	for _, f := range fields {
		assert.Equal(t, f, gjson.GetBytes(body, fmt.Sprintf("methods.password.config.fields.#(name==%s).name", f)).String(), "%s", body)
	}
}

// checks if some keys are not set, this should be used to catch regression issues
func outdatedFieldsDoNotExist(t *testing.T, body []byte) {
	for _, k := range []string{"request"} {
		assert.Equal(t, false, gjson.GetBytes(body, fmt.Sprintf("methods.password.config.fields.#(name==%s)", k)).Exists())
	}
}

func formMethodIsPOST(t *testing.T, body []byte) {
	assert.Equal(t, "POST", gjson.GetBytes(body, "methods.password.config.method").String())
}

func TestRegistration(t *testing.T) {
	t.Run("case=registration", func(t *testing.T) {
		_, reg := internal.NewRegistryDefault(t)
		s := reg.RegistrationStrategies().MustStrategy(identity.CredentialsTypePassword).(*password.Strategy)
		s.WithTokenGenerator(func(r *http.Request) string {
			return "nosurf"
		})

		router := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		reg.RegistrationStrategies().RegisterPublicRoutes(router)
		reg.RegistrationHandler().RegisterPublicRoutes(router)
		reg.RegistrationHandler().RegisterAdminRoutes(admin)
		ts := httptest.NewServer(router)
		defer ts.Close()

		errTs, uiTs, returnTs := newErrTs(t, reg), httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			e, err := reg.RegistrationRequestPersister().GetRegistrationRequest(r.Context(), x.ParseUUID(r.URL.Query().Get("request")))
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

		var newRegistrationRequest = func(t *testing.T, exp time.Duration) *registration.Request {
			rr := &registration.Request{
				ID:       x.NewUUID(),
				IssuedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(exp), RequestURL: ts.URL,
				Methods: map[identity.CredentialsType]*registration.RequestMethod{
					identity.CredentialsTypePassword: {
						Method: identity.CredentialsTypePassword,
						Config: &registration.RequestMethodConfig{
							RequestMethodConfigurator: password.RequestMethod{
								HTMLForm: &form.HTMLForm{
									Method: "POST",
									Action: "/action",
									Fields: form.Fields{
										{Name: "password", Type: "password", Required: true},
										{Name: "csrf_token", Type: "hidden", Required: true, Value: "csrf-token"},
									},
								},
							},
						},
					},
				},
			}
			require.NoError(t, reg.RegistrationRequestPersister().CreateRegistrationRequest(context.Background(), rr))
			return rr
		}

		var makeRequest = func(t *testing.T, rid uuid.UUID, body string, expectedStatusCode int) ([]byte, *http.Response) {
			jar, _ := cookiejar.New(&cookiejar.Options{})
			client := http.Client{Jar: jar}
			res, err := client.Post(ts.URL+password.RegistrationPath+"?request="+rid.String(), "application/x-www-form-urlencoded", strings.NewReader(body))
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
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.errors.0.message").String(), "invalid URL escape", "%s", body)
		})

		t.Run("case=should show the error ui because the request id is missing", func(t *testing.T) {
			_ = newRegistrationRequest(t, time.Minute)
			uuidDesNotExistInStore := x.NewUUID()
			body, res := makeRequest(t, uuidDesNotExistInStore, "", http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "error-ts")
			assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Not Found", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.message").String(), "Unable to locate the resource", "%s", body)
		})

		t.Run("case=should return an error because the request is expired", func(t *testing.T) {
			rr := newRegistrationRequest(t, -time.Minute)
			body, res := makeRequest(t, rr.ID, "", http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
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
			assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			checkFormContent(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==password).errors.0").String(), "data breaches and must no longer be used.", "%s", body)
		})

		t.Run("case=should return an error because not passing validation", func(t *testing.T) {
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-5"},
				"password":        {x.NewUUID().String()},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			checkFormContent(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foobar).errors.0").String(), "foobar is required", "%s", body)
		})

		t.Run("case=should fail because schema did not specify an identifier", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/missing-identifier.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-6"},
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "error-ts")
			assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "No login identifiers", "%s", body)
		})

		t.Run("case=should fail because schema does not exist", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/i-do-not-exist.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-7"},
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "error-ts")
			assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "no such file or directory", "%s", body)
		})

		t.Run("case=should pass and set up a session", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-8"},
				"password":        {x.NewUUID().String()},
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
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.errors.0.message").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
		})

		t.Run("case=should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")

			rr := &registration.Request{
				ID:        x.NewUUID(),
				ExpiresAt: time.Now().Add(time.Minute),
				Methods: map[identity.CredentialsType]*registration.RequestMethod{
					identity.CredentialsTypePassword: {
						Method: identity.CredentialsTypePassword,
						Config: &registration.RequestMethodConfig{
							RequestMethodConfigurator: &password.RequestMethod{
								HTMLForm: &form.HTMLForm{
									Method: "POST",
									Action: "/action",
									Errors: []form.Error{{Message: "some error"}},
									Fields: form.Fields{
										{
											Name: "traits.foo", Value: "bar", Type: "text",
											Errors: []form.Error{{Message: "bar"}},
										},
										{Name: "password"},
									},
								},
							},
						},
					},
				},
			}

			require.NoError(t, reg.RegistrationRequestPersister().CreateRegistrationRequest(context.Background(), rr))
			body, res := makeRequest(t, rr.ID, url.Values{
				"traits.username": {"registration-identifier-9"},
				"password":        {x.NewUUID().String()},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			checkFormContent(t, body, "password", "csrf_token", "traits.username")

			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foo).value"), "%s", body)
			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foo).error"))
			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.error"))
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foobar).errors.0").String(), "foobar is required", "%s", body)
		})

		t.Run("case=should work even if password is just numbers", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://stub/registration.schema.json")
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
		_, reg := internal.NewRegistryDefault(t)
		s := reg.RegistrationStrategies().MustStrategy(identity.CredentialsTypePassword).(*password.Strategy)
		s.WithTokenGenerator(func(r *http.Request) string {
			return "nosurf"
		})

		viper.Set(configuration.ViperKeyURLsSelfPublic, urlx.ParseOrPanic("https://foo/"))
		viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://stub/registration.schema.json")

		sr := registration.NewRequest(time.Minute, "nosurf", &http.Request{URL: urlx.ParseOrPanic("/")})
		require.NoError(t, s.PopulateRegistrationMethod(&http.Request{}, sr))

		expected := &registration.RequestMethod{
			Method: identity.CredentialsTypePassword,
			Config: &registration.RequestMethodConfig{
				RequestMethodConfigurator: &password.RequestMethod{
					HTMLForm: &form.HTMLForm{
						Action: "https://foo" + password.RegistrationPath + "?request=" + sr.ID.String(),
						Method: "POST",
						Fields: form.Fields{
							{
								Name:     "password",
								Type:     "password",
								Required: true,
							},
							{
								Name:     "csrf_token",
								Type:     "hidden",
								Required: true,
								Value:    "nosurf",
							},
							{
								Name: "traits.foobar",
								Type: "text",
							},
							{
								Name: "traits.username",
								Type: "text",
							},
						},
					},
				},
			},
		}
		sort.Sort(expected.Config.RequestMethodConfigurator.(*password.RequestMethod).HTMLForm.Fields)

		actual := sr.Methods[identity.CredentialsTypePassword]
		assert.EqualValues(t, expected.Config.RequestMethodConfigurator.(*password.RequestMethod).HTMLForm, actual.Config.RequestMethodConfigurator.(*password.RequestMethod).HTMLForm)
	})
}
