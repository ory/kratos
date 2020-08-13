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

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/text"
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
		_, reg := internal.NewFastRegistryWithMocks(t)

		router := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
			"enabled": true})
		ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, admin)
		defer ts.Close()

		errTs, uiTs, returnTs := newErrTs(t, reg), httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			e, err := reg.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("request")))
			require.NoError(t, err)
			reg.Writer().Write(w, r, e)
		})), newReturnTs(t, reg)
		defer errTs.Close()
		defer uiTs.Close()
		defer returnTs.Close()

		viper.Set(configuration.ViperKeySelfServiceErrorUI, errTs.URL+"/error-ts")
		viper.Set(configuration.ViperKeySelfServiceRegistrationUI, uiTs.URL+"/signup-ts")
		viper.Set(configuration.ViperKeyPublicBaseURL, ts.URL)
		viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, returnTs.URL+"/default-return-to")
		viper.Set(configuration.ViperKeySelfServiceRegistrationAfter+"."+configuration.DefaultBrowserReturnURL, returnTs.URL+"/return-ts")
		viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

		var newRegistrationRequest = func(t *testing.T, exp time.Duration) *registration.Flow {
			rr := &registration.Flow{
				ID:       x.NewUUID(),
				Type: flow.TypeBrowser,
				IssuedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(exp), RequestURL: ts.URL,
				Methods: map[identity.CredentialsType]*registration.FlowMethod{
					identity.CredentialsTypePassword: {
						Method: identity.CredentialsTypePassword,
						Config: &registration.FlowMethodConfig{
							FlowMethodConfigurator: password.RequestMethod{
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
			require.NoError(t, reg.RegistrationFlowPersister().CreateRegistrationFlow(context.Background(), rr))
			return rr
		}

		var makeRequestWithCookieJar = func(t *testing.T, rid uuid.UUID, isAPI bool, body string, expectedStatusCode int, jar *cookiejar.Jar) ([]byte, *http.Response) {
			contentType := "application/x-www-form-urlencoded"
			if isAPI {
				contentType = "application/json"
			}
			client := http.Client{Jar: jar}
			res, err := client.Post(ts.URL+password.RegistrationPath+"?request="+rid.String(), contentType, strings.NewReader(body))
			require.NoError(t, err)
			result, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)
			require.EqualValues(t, expectedStatusCode, res.StatusCode, "Request: %+v\n\t\tResponse: %s", res.Request, res)
			return result, res
		}

		var makeRequest = func(t *testing.T, rid uuid.UUID, isAPI bool, body string, expectedStatusCode int) ([]byte, *http.Response) {
			jar, _ := cookiejar.New(&cookiejar.Options{})
			return makeRequestWithCookieJar(t, rid, isAPI, body, expectedStatusCode, jar)
		}

		t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, false, "14=)=!(%)$/ZP()GHIÖ", http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.messages.0.text").String(), "invalid URL escape", "%s", body)
		})

		t.Run("case=should show the error ui because the request id is missing", func(t *testing.T) {
			_ = newRegistrationRequest(t, time.Minute)
			uuidDesNotExistInStore := x.NewUUID()
			body, res := makeRequest(t, uuidDesNotExistInStore, false, "", http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "error-ts")
			assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Not Found", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.message").String(), "Unable to locate the resource", "%s", body)
		})

		t.Run("case=should return an error because the request is expired", func(t *testing.T) {
			rr := newRegistrationRequest(t, -time.Minute)
			body, res := makeRequest(t, rr.ID, false, "", http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.NotEqual(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "messages.0.text").String(), "expired", "%s", body)
		})

		t.Run("case=should return an error because the password failed validation", func(t *testing.T) {
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, false, url.Values{
				"traits.username": {"registration-identifier-4"},
				"password":        {"password"},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			checkFormContent(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==password).messages.0").String(), "data breaches and must no longer be used.", "%s", body)
		})

		t.Run("case=should return an error because not passing validation", func(t *testing.T) {
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, false, url.Values{
				"traits.username": {"registration-identifier-5"},
				"password":        {x.NewUUID().String()},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			checkFormContent(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", body)
		})

		t.Run("case=should fail because schema did not specify an identifier", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/missing-identifier.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, false, url.Values{
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
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/i-do-not-exist.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, false, url.Values{
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
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			viper.Set(
				configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()),
				[]configuration.SelfServiceHook{{Name: "session"}})

			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, false, url.Values{
				"traits.username": {"registration-identifier-8"},
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "return-ts")
			assert.Equal(t, `registration-identifier-8`, gjson.GetBytes(body, "identity.traits.username").String(), "%s", body)
		})

		t.Run("case=should fail to register the same user again", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, false, url.Values{
				"traits.username": {"registration-identifier-8"},
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "signup-ts")
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
		})

		t.Run("case=should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

			rr := &registration.Flow{
				ID:        x.NewUUID(),
				ExpiresAt: time.Now().Add(time.Minute),
				Methods: map[identity.CredentialsType]*registration.FlowMethod{
					identity.CredentialsTypePassword: {
						Method: identity.CredentialsTypePassword,
						Config: &registration.FlowMethodConfig{
							FlowMethodConfigurator: &password.RequestMethod{
								HTMLForm: &form.HTMLForm{
									Method:   "POST",
									Action:   "/action",
									Messages: text.Messages{{Text: "some error"}},
									Fields: form.Fields{
										{
											Name: "traits.foo", Value: "bar", Type: "text",
											Messages: text.Messages{{Text: "bar"}},
										},
										{Name: "password"},
									},
								},
							},
						},
					},
				},
			}

			require.NoError(t, reg.RegistrationFlowPersister().CreateRegistrationFlow(context.Background(), rr))
			body, res := makeRequest(t, rr.ID, false, url.Values{
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
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", body)
		})

		t.Run("case=should work even if password is just numbers", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://stub/registration.schema.json")
			rr := newRegistrationRequest(t, time.Minute)
			body, res := makeRequest(t, rr.ID, false, url.Values{
				"traits.username": {"registration-identifier-10"},
				"password":        {"93172388957812344432"},
				"traits.foobar":   {"bar"},
			}.Encode(), http.StatusOK)
			assert.Contains(t, res.Request.URL.Path, "return-ts")
			assert.Equal(t, `registration-identifier-10`, gjson.GetBytes(body, "identity.traits.username").String(), "%s", body)
		})

		t.Run("case=register and then send same request", func(t *testing.T) {
			jar, _ := cookiejar.New(&cookiejar.Options{})
			formValues := url.Values{
				"traits.username": {"registration-identifier-11"},
				"password":        {"O(lf<ys87LÖ:(h<dsjfl"},
				"traits.foobar":   {"bar"},
			}.Encode()
			rr1 := newRegistrationRequest(t, time.Minute)
			body1, res1 := makeRequestWithCookieJar(t, rr1.ID, false, formValues, http.StatusOK, jar)
			assert.Contains(t, res1.Request.URL.Path, "return-ts")
			rr2 := newRegistrationRequest(t, time.Minute)
			body2, res2 := makeRequestWithCookieJar(t, rr2.ID, false, formValues, http.StatusOK, jar)
			assert.Contains(t, res2.Request.URL.Path, "default-return-to")
			assert.Equal(t, body1, body2)
		})
	})

	t.Run("method=PopulateSignUpMethod", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)

		viper.Set(configuration.ViperKeyPublicBaseURL, urlx.ParseOrPanic("https://foo/"))
		viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://stub/registration.schema.json")
		viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
			"enabled": true})

		sr := registration.NewFlow(time.Minute, "nosurf", &http.Request{URL: urlx.ParseOrPanic("/")}, flow.TypeBrowser)
		require.NoError(t, reg.RegistrationStrategies().MustStrategy(identity.CredentialsTypePassword).(*password.Strategy).PopulateRegistrationMethod(&http.Request{}, sr))

		expected := &registration.FlowMethod{
			Method: identity.CredentialsTypePassword,
			Config: &registration.FlowMethodConfig{
				FlowMethodConfigurator: &password.RequestMethod{
					HTMLForm: &form.HTMLForm{
						Action: "https://foo" + password.RegistrationPath + "?request=" + sr.ID.String(),
						Method: "POST",
						Fields: form.Fields{
							{
								Name:     "csrf_token",
								Type:     "hidden",
								Required: true,
								Value:    x.FakeCSRFToken,
							},
							{
								Name:     "password",
								Type:     "password",
								Required: true,
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

		actual := sr.Methods[identity.CredentialsTypePassword]
		assert.EqualValues(t, expected.Config.FlowMethodConfigurator.(*password.RequestMethod).HTMLForm, actual.Config.FlowMethodConfigurator.(*password.RequestMethod).HTMLForm)
	})
}
