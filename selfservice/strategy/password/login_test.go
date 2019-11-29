package password_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/x/pointerx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/x"
)

func nlr(exp time.Duration) *login.Request {
	id := x.NewUUID()
	return &login.Request{
		ID:         id,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(exp),
		RequestURL: "remove-this-if-test-fails",
		Methods: map[identity.CredentialsType]*login.RequestMethod{
			identity.CredentialsTypePassword: {
				Method: identity.CredentialsTypePassword,
				Config: &login.RequestMethodConfig{
					RequestMethodConfigurator: &form.HTMLForm{
						Action: "/action",
						Fields: form.Fields{
							"identifier": {
								Name:     "identifier",
								Type:     "text",
								Required: true,
							},
							"password": {
								Name:     "password",
								Type:     "password",
								Required: true,
							},
							form.CSRFTokenName: {
								Name:     form.CSRFTokenName,
								Type:     "hidden",
								Required: true,
								Value:    "anti-rf-token",
							},
							"request": {
								Name:     "request",
								Type:     "hidden",
								Required: true,
								Value:    id.String(),
							},
						},
					},
				},
			},
		},
	}
}

type loginStrategyDependencies interface {
	password.ValidationProvider
	identity.PoolProvider
	password.HashProvider
}

func TestLogin(t *testing.T) {
	var ensureFieldsExist = func(t *testing.T, body []byte) {
		for _, k := range []string{"identifier",
			"password",
			"csrf_token",
		} {
			assert.Equal(t, k, gjson.GetBytes(body, fmt.Sprintf("methods.password.config.fields.%s.name", k)).String())
		}
	}

	type testCase struct {
		d              string
		prep           func(t *testing.T, r loginStrategyDependencies)
		ar             *login.Request
		forceRequestID *string
		payload        string
		assert         func(t *testing.T, tc testCase, r *http.Response)
	}

	for k, tc := range []testCase{

		{
			d:       "should show the error ui because the request is malformed",
			ar:      nlr(0),
			payload: "14=)=!(%)$/ZP()GHIÖ",
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				require.Contains(t, r.Request.URL.Path, "login-ts", "%+v", r.Request)
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, tc.ar.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
				assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "methods.password.config.errors.0.message").String(), `invalid URL escape`)
			},
		},
		{
			d:              "should show the error ui because the request id missing",
			ar:             nlr(time.Minute),
			forceRequestID: pointerx.String(""),
			payload:        url.Values{}.Encode(),
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				assert.Contains(t, r.Request.URL.Path, "error-ts")
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "0.code").Int(), "%s", body)
				assert.Equal(t, "Bad Request", gjson.GetBytes(body, "0.status").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "request query parameter is missing or invalid", "%s", body)
			},
		},
		{
			d:              "should return an error because the request does not exist",
			ar:             nlr(0),
			forceRequestID: pointerx.String(x.NewUUID().String()),
			payload: url.Values{
				"identifier": {"identifier"},
				"password":   {"password"},
			}.Encode(),
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				assert.Contains(t, r.Request.URL.Path, "error-ts")
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "0.code").Int(), "%s", body)
				assert.Equal(t, "Not Found", gjson.GetBytes(body, "0.status").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "Unable to find request", "%s", body)
			},
		},
		{
			d:  "should return an error because the request is expired",
			ar: nlr(-time.Hour),
			payload: url.Values{
				"identifier": {"identifier"},
				"password":   {"password"},
			}.Encode(),
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				assert.Contains(t, r.Request.URL.Path, "error-ts")
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "0.code").Int(), "%s", body)
				assert.Equal(t, "Bad Request", gjson.GetBytes(body, "0.status").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "expired", "%s", body)
			},
		},
		{
			d:  "should return an error because the credentials are invalid (user does not exist)",
			ar: nlr(time.Hour),
			payload: url.Values{
				"identifier": {"identifier"},
				"password":   {"password"},
			}.Encode(),
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				require.Contains(t, r.Request.URL.Path, "login-ts")
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, tc.ar.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
				assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())
				assert.Equal(t, `The provided credentials are invalid. Check for spelling mistakes in your password or username, email address, or phone number.`, gjson.GetBytes(body, "methods.password.config.errors.0.message").String())
			},
		},
		{
			d:  "should return an error because no identifier is set",
			ar: nlr(time.Hour),
			payload: url.Values{
				"password": {"password"},
			}.Encode(),
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				require.Contains(t, r.Request.URL.Path, "login-ts")
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				// Let's ensure that the payload is being propagated properly.
				assert.Equal(t, tc.ar.ID.String(), gjson.GetBytes(body, "id").String())
				assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())
				ensureFieldsExist(t, body)
				assert.Equal(t, "identifier: identifier is required", gjson.GetBytes(body, "methods.password.config.fields.identifier.errors.0.message").String(), "%s", body)

				// The password value should not be returned!
				assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.password.value").String())
			},
		},
		{
			d:  "should return an error because no password is set",
			ar: nlr(time.Hour),
			payload: url.Values{
				"identifier": {"identifier"},
			}.Encode(),
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				require.Contains(t, r.Request.URL.Path, "login-ts")
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				// Let's ensure that the payload is being propagated properly.
				assert.Equal(t, tc.ar.ID.String(), gjson.GetBytes(body, "id").String())
				assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())
				ensureFieldsExist(t, body)
				assert.Equal(t, "password: password is required", gjson.GetBytes(body, "methods.password.config.fields.password.errors.0.message").String(), "%s", body)

				assert.Equal(t, "anti-rf-token", gjson.GetBytes(body, "methods.password.config.fields.csrf_token.value").String())
				assert.Equal(t, "identifier", gjson.GetBytes(body, "methods.password.config.fields.identifier.value").String(), "%s", body)

				// This must not include the password!
				assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.password.value").String())
			},
		},
		{
			d:  "should return an error because the credentials are invalid (password not correct)",
			ar: nlr(time.Hour),
			prep: func(t *testing.T, r loginStrategyDependencies) {
				p, _ := r.PasswordHasher().Generate([]byte("password"))
				_, err := r.IdentityPool().Create(context.Background(), &identity.Identity{
					ID:     x.NewUUID(),
					Traits: json.RawMessage(`{}`),
					Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {
							Type:        identity.CredentialsTypePassword,
							Identifiers: []string{"login-identifier-6"},
							Config:      json.RawMessage(`{"hashed_password":"` + string(p) + `"}`),
						},
					},
				})
				require.NoError(t, err)
			},
			payload: url.Values{
				"identifier": {"login-identifier-6"},
				"password":   {"not-password"},
			}.Encode(),
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				require.Contains(t, r.Request.URL.Path, "login-ts")
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, tc.ar.ID.String(), gjson.GetBytes(body, "id").String())
				assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())
				ensureFieldsExist(t, body)
				assert.Equal(t, schema.NewInvalidCredentialsError().(schema.ResultErrors)[0].Description(), gjson.GetBytes(body, "methods.password.config.errors.0.message").String(), "%s", body)

				// This must not include the password!
				assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.password.value").String())
			},
		},
		{
			d:  "should pass because everything is a-ok",
			ar: nlr(time.Hour),
			prep: func(t *testing.T, r loginStrategyDependencies) {
				p, _ := r.PasswordHasher().Generate([]byte("password"))
				_, err := r.IdentityPool().Create(context.Background(), &identity.Identity{
					ID:     x.NewUUID(),
					Traits: json.RawMessage(`{"subject":"login-identifier-7"}`),
					Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {
							Type:        identity.CredentialsTypePassword,
							Identifiers: []string{"login-identifier-7"},
							Config:      json.RawMessage(`{"hashed_password":"` + string(p) + `"}`),
						},
					},
				})
				require.NoError(t, err)
			},
			payload: url.Values{
				"identifier": {"login-identifier-7"},
				"password":   {"password"},
			}.Encode(),
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				assert.Contains(t, r.Request.URL.Path, "return-ts", "%s", r.Request.URL.String())
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, `login-identifier-7`, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
			},
		},
		{
			d: "should return an error because not passing validation and reset previous errors and values",
			ar: &login.Request{
				ID:        x.NewUUID(),
				ExpiresAt: time.Now().Add(time.Minute),
				Methods: map[identity.CredentialsType]*login.RequestMethod{
					identity.CredentialsTypePassword: {
						Method: identity.CredentialsTypePassword,
						Config: &login.RequestMethodConfig{
							RequestMethodConfigurator: &password.RequestMethod{
								HTMLForm: &form.HTMLForm{
									Action: "/action",
									Errors: []form.Error{{Message: "some error"}},
									Fields: form.Fields{
										"identifier": {
											Value:  "baz",
											Name:   "identifier",
											Errors: []form.Error{{Message: "err"}},
										},
										"password": {
											Value:  "bar",
											Name:   "password",
											Errors: []form.Error{{Message: "err"}},
										},
									},
								},
							},
						},
					},
				},
			},
			payload: url.Values{
				"identifier": {"registration-identifier-9"},
				// "password": {uuid.New().String()},
			}.Encode(),
			assert: func(t *testing.T, tc testCase, r *http.Response) {
				require.Contains(t, r.Request.URL.Path, "login-ts")
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, tc.ar.ID.String(), gjson.GetBytes(body, "id").String())
				assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())
				ensureFieldsExist(t, body)

				assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.identity.value"))
				assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.identity.error"))
				assert.Empty(t, gjson.GetBytes(body, "methods.password.config.error"))
				assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.password.errors.0").String(), "password: password is required", "%s", body)
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			_, reg := internal.NewMemoryRegistry(t)
			s := reg.LoginStrategies().MustStrategy(identity.CredentialsTypePassword).(*password.Strategy)
			s.WithTokenGenerator(func(r *http.Request) string {
				return "anti-rf-token"
			})

			router := x.NewRouterPublic()
			reg.LoginHandler().RegisterPublicRoutes(router)
			s.RegisterLoginRoutes(router)
			ts := httptest.NewServer(router)
			defer ts.Close()

			errTs, uiTs, returnTs := errorx.NewErrorTestServer(t, reg), httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				e, err := reg.LoginRequestPersister().GetLoginRequest(context.Background(), x.ParseUUID(r.URL.Query().Get("request")))
				require.NoError(t, err)
				reg.Writer().Write(w, r, e)
			})), newReturnTs(t, reg)
			defer errTs.Close()
			defer uiTs.Close()
			defer returnTs.Close()

			viper.Set(configuration.ViperKeyURLsError, errTs.URL+"/error-ts")
			viper.Set(configuration.ViperKeyURLsLogin, uiTs.URL+"/login-ts")
			viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
			viper.Set(configuration.ViperKeySelfServiceLoginAfterConfig+"."+string(identity.CredentialsTypePassword), hookConfig(returnTs.URL+"/return-ts"))
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/login.schema.json")

			if tc.prep != nil {
				tc.prep(t, reg)
			}

			tc.ar.RequestURL = ts.URL
			require.NoError(t, reg.LoginRequestPersister().CreateLoginRequest(context.TODO(), tc.ar))

			c := ts.Client()
			c.Jar, _ = cookiejar.New(&cookiejar.Options{})

			requestID := tc.ar.ID.String()
			if tc.forceRequestID != nil {
				requestID = *tc.forceRequestID
			}

			res, err := c.Post(ts.URL+password.LoginPath+"?request="+requestID, "application/x-www-form-urlencoded", strings.NewReader(tc.payload))
			require.NoError(t, err)
			defer res.Body.Close()
			require.EqualValues(t, http.StatusOK, res.StatusCode, "Request: %+v\n\t\tResponse: %s", res.Request, res)

			tc.assert(t, tc, res)
		})
	}
}
