package password_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/x/assertx"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlxx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/pointerx"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func nlr(exp time.Duration, isAPI bool) *login.Flow {
	id := x.NewUUID()
	ft := flow.TypeBrowser
	if isAPI {
		ft = flow.TypeAPI
	}
	return &login.Flow{
		ID:         id,
		Type:       ft,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(exp),
		RequestURL: "remove-this-if-test-fails",
		Methods: map[identity.CredentialsType]*login.FlowMethod{
			identity.CredentialsTypePassword: {
				Method: identity.CredentialsTypePassword,
				Config: &login.FlowMethodConfig{
					FlowMethodConfigurator: &form.HTMLForm{
						Method: "POST",
						Action: "/action",
						Fields: form.Fields{
							{
								Name:     "identifier",
								Type:     "text",
								Required: true,
							},
							{
								Name:     "password",
								Type:     "password",
								Required: true,
							},
							{
								Name:     form.CSRFTokenName,
								Type:     "hidden",
								Required: true,
								Value:    x.FakeCSRFToken,
							},
						},
					},
				},
			},
		},
	}
}

func TestCompleteLogin(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)

	viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true})
	ts, _ := testhelpers.NewKratosServer(t, reg)

	errTs := testhelpers.NewErrorTestServer(t, reg)
	uiTs := testhelpers.NewLoginUIFlowEchoServer(t, reg)
	newReturnTs(t, reg)

	// Overwrite these two:
	viper.Set(configuration.ViperKeySelfServiceErrorUI, errTs.URL+"/error-ts")
	viper.Set(configuration.ViperKeySelfServiceLoginUI, uiTs.URL+"/login-ts")

	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")
	viper.Set(configuration.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	makeRequestRaw := func(t *testing.T, isAPI bool, payload string, requestID string, c *http.Client, esc int) (*http.Response, []byte) {
		req, err := http.NewRequest("POST", ts.URL+password.RouteLogin+"?flow="+requestID, strings.NewReader(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "text/html")
		if isAPI {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
		}

		res, err := c.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, esc, res.StatusCode, "Flow: %+v\n\t\tResponse: %s", res.Request, res)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	initBrowserFlow := func(t *testing.T, payload string, jar *cookiejar.Jar, force bool, esc int) (*http.Response, []byte) {
		c := &http.Client{Jar: jar}
		if jar == nil {
			c.Jar, _ = cookiejar.New(&cookiejar.Options{})
		}

		u := ts.URL + login.RouteInitBrowserFlow
		if force {
			u = u + "?refresh=true"
		}

		res, err := c.Get(u)
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, res.StatusCode, "Flow: %+v\n\t\tResponse: %s", res.Request, res)
		assert.NotEmpty(t, res.Request.URL.Query().Get("flow"))

		return makeRequestRaw(t, false, payload, res.Request.URL.Query().Get("flow"), c, esc)
	}

	initAPIFlow := func(t *testing.T, payload string, force bool, esc int) (*http.Response, []byte) {
		c := &http.Client{}
		u := ts.URL + login.RouteInitAPIFlow
		if force {
			u = u + "?refresh=true"
		}

		res, err := c.Get(u)
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)

		id := gjson.GetBytes(body, "id").String()
		require.NotEmpty(t, id)

		return makeRequestRaw(t, true, payload, id, c, esc)
	}

	fakeRequest := func(t *testing.T, lr *login.Flow, isAPI bool, payload string, forceRequestID *string, jar *cookiejar.Jar, esc int) (*http.Response, []byte) {
		lr.RequestURL = ts.URL
		require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.TODO(), lr))

		requestID := lr.ID.String()
		if forceRequestID != nil {
			requestID = *forceRequestID
		}

		c := &http.Client{Jar: jar}
		if jar == nil {
			c.Jar, _ = cookiejar.New(&cookiejar.Options{})
		}

		return makeRequestRaw(t, isAPI, payload, requestID, c, esc)
	}

	ensureFieldsExist := func(t *testing.T, body []byte) {
		checkFormContent(t, body, "identifier",
			"password",
			"csrf_token")
	}

	createIdentity := func(identifier, password string) {
		p, _ := reg.Hasher().Generate([]byte(password))
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &identity.Identity{
			ID:     x.NewUUID(),
			Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, identifier)),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{identifier},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
				},
			},
		}))
	}

	t.Run("should show the error ui because the request is malformed", func(t *testing.T) {
		run := func(t *testing.T, isAPI bool) (string, *http.Response) {
			lr := nlr(0, isAPI)
			res, body := fakeRequest(t, lr, isAPI, "14=)=!(%)$/ZP()GHIÖ", nil, nil, expectStatusCode(isAPI, http.StatusBadRequest))

			assert.Equal(t, lr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
			return gjson.GetBytes(body, "methods.password.config.messages.0.text").String(), res
		}

		t.Run("type=browser", func(t *testing.T) {
			body, res := run(t, false)
			require.Contains(t, res.Request.URL.Path, "login-ts", res.Request.URL)
			assert.Contains(t, body, `invalid URL escape`)
		})

		t.Run("type=api", func(t *testing.T) {
			body, res := run(t, true)
			require.Contains(t, res.Request.URL.Path, password.RouteLogin, res.Request.URL)
			assert.Contains(t, body, `cannot unmarshal number`)
		})
	})

	t.Run("should show the error ui because the request id missing", func(t *testing.T) {
		run := func(t *testing.T, isAPI bool) (*http.Response, []byte) {
			lr := nlr(time.Minute, isAPI)
			return fakeRequest(t, lr, isAPI, url.Values{}.Encode(), pointerx.String(""), nil, expectStatusCode(isAPI, http.StatusBadRequest))
		}

		t.Run("type=browser", func(t *testing.T) {
			res, body := run(t, false)
			require.Contains(t, res.Request.URL.Path, "error-ts")
			assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Bad Flow", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "request query parameter is missing or invalid", "%s", body)
		})

		t.Run("type=api", func(t *testing.T) {
			res, body := run(t, true)
			require.Contains(t, res.Request.URL.Path, password.RouteLogin, res.Request.URL)
			assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "error.code").Int(), "%s", body)
			assert.Equal(t, "Bad Flow", gjson.GetBytes(body, "error.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), "request query parameter is missing or invalid", "%s", body)
		})
	})

	t.Run("should return an error because the request does not exist", func(t *testing.T) {
		run := func(t *testing.T, isAPI bool, payload string) (*http.Response, []byte) {
			lr := nlr(0, isAPI)
			return fakeRequest(t, lr, isAPI, payload, pointerx.String(x.NewUUID().String()), nil, expectStatusCode(isAPI, http.StatusNotFound))
		}

		t.Run("type=browser", func(t *testing.T) {
			res, body := run(t, false, url.Values{"identifier": {"identifier"},
				"password": {"password"}}.Encode())

			require.Contains(t, res.Request.URL.Path, "error-ts", res.Request.URL)
			assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "0.code").Int(), "%s", body)
			assert.Equal(t, "Not Found", gjson.GetBytes(body, "0.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "0.message").String(), "Unable to locate the resource", "%s", body)
		})

		t.Run("type=api", func(t *testing.T) {
			res, body := run(t, true, x.MustEncodeJSON(t, &password.LoginFormPayload{Identifier: "identifier",
				Password: "password"}))

			require.Contains(t, res.Request.URL.Path, password.RouteLogin, res.Request.URL)
			assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "error.code").Int(), "%s", body)
			assert.Equal(t, "Not Found", gjson.GetBytes(body, "error.status").String(), "%s", body)
			assert.Contains(t, gjson.GetBytes(body, "error.message").String(), "Unable to locate the resource", "%s", body)
		})
	})

	t.Run("should redirect to login init because the request is expired", func(t *testing.T) {
		run := func(t *testing.T, isAPI bool, payload string) (*login.Flow, *http.Response, []byte) {
			lr := nlr(-time.Hour, isAPI)
			res, body := fakeRequest(t, lr, isAPI, payload, nil, nil, http.StatusOK)
			return lr, res, body
		}

		t.Run("type=browser", func(t *testing.T) {
			lr, res, body := run(t, false, url.Values{"identifier": {"identifier"},
				"password": {"password"}}.Encode())
			require.Contains(t, res.Request.URL.Path, "login-ts")
			assert.NotEqual(t, lr.ID, gjson.GetBytes(body, "id"))
			assert.Contains(t, gjson.GetBytes(body, "messages.0").String(), "expired", "%s", body)
		})

		t.Run("type=api", func(t *testing.T) {
			lr, res, body := run(t, true, x.MustEncodeJSON(t, &password.LoginFormPayload{Identifier: "identifier",
				Password: "password"}))
			require.Contains(t, res.Request.URL.Path, login.RouteGetFlow, res.Request.URL)
			assert.NotEqual(t, lr.ID, gjson.GetBytes(body, "id"))
			assert.Contains(t, gjson.GetBytes(body, "messages.0").String(), "expired", "%s", body)
		})
	})

	t.Run("should return an error because the credentials are invalid (user does not exist)", func(t *testing.T) {
		run := func(t *testing.T, isAPI bool, payload string) *http.Response {
			lr := nlr(time.Hour, isAPI)
			res, body := fakeRequest(t, lr, isAPI, payload, nil, nil, expectStatusCode(isAPI, http.StatusBadRequest))
			assert.Equal(t, lr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())
			assert.Equal(t, text.NewErrorValidationInvalidCredentials().Text, gjson.GetBytes(body, "methods.password.config.messages.0.text").String())
			return res
		}

		t.Run("type=browser", func(t *testing.T) {
			require.Contains(t, run(t, false, url.Values{
				"identifier": {"identifier"}, "password": {"password"}}.Encode()).Request.URL.Path, "login-ts")
		})

		t.Run("type=api", func(t *testing.T) {
			require.Contains(t, run(t, true, x.MustEncodeJSON(t, &password.LoginFormPayload{
				Identifier: "identifier", Password: "password"})).Request.URL.Path, password.RouteLogin)
		})
	})

	t.Run("should return an error because no identifier is set", func(t *testing.T) {
		run := func(t *testing.T, isAPI bool, payload string) *http.Response {
			lr := nlr(time.Hour, isAPI)
			res, body := fakeRequest(t, lr, isAPI, payload, nil, nil, expectStatusCode(isAPI, http.StatusBadRequest))

			// Let's ensure that the payload is being propagated properly.
			assert.Equal(t, lr.ID.String(), gjson.GetBytes(body, "id").String())
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())
			ensureFieldsExist(t, body)
			assert.Equal(t, "Property identifier is missing.", gjson.GetBytes(body, "methods.password.config.fields.#(name==identifier).messages.0.text").String(), "%s", body)

			// The password value should not be returned!
			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==password).value").String())
			return res
		}

		t.Run("type=browser", func(t *testing.T) {
			require.Contains(t, run(t, false,
				url.Values{"password": {"password"}}.Encode()).Request.URL.Path, "login-ts")
		})

		t.Run("type=api", func(t *testing.T) {
			require.Contains(t, run(t, true,
				x.MustEncodeJSON(t, &password.LoginFormPayload{Password: "password"})).Request.URL.Path,
				password.RouteLogin)
		})
	})

	t.Run("should return an error because no password is set", func(t *testing.T) {
		run := func(t *testing.T, isAPI bool, payload string) *http.Response {
			lr := nlr(time.Hour, isAPI)
			res, body := fakeRequest(t, lr, isAPI, payload, nil, nil, expectStatusCode(isAPI, http.StatusBadRequest))

			// Let's ensure that the payload is being propagated properly.
			assert.Equal(t, lr.ID.String(), gjson.GetBytes(body, "id").String())
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())
			ensureFieldsExist(t, body)
			assert.Equal(t, "Property password is missing.", gjson.GetBytes(body, "methods.password.config.fields.#(name==password).messages.0.text").String(), "%s", body)

			if !isAPI {
				assert.Equal(t, x.FakeCSRFToken, gjson.GetBytes(body, "methods.password.config.fields.#(name==csrf_token).value").String())
			}
			assert.Equal(t, "identifier", gjson.GetBytes(body, "methods.password.config.fields.#(name==identifier).value").String(), "%s", body)

			// This must not include the password!
			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==password).value").String())
			return res
		}

		t.Run("type=browser", func(t *testing.T) {
			require.Contains(t, run(t, false,
				url.Values{"identifier": {"identifier"}}.Encode()).Request.URL.Path, "login-ts")
		})

		t.Run("type=api", func(t *testing.T) {
			require.Contains(t, run(t, true, x.MustEncodeJSON(t,
				&password.LoginFormPayload{Identifier: "identifier"})).Request.URL.Path, password.RouteLogin)
		})
	})

	t.Run("should return an error because the credentials are invalid (password not correct)", func(t *testing.T) {
		run := func(t *testing.T, isAPI bool) *http.Response {
			identifier, pwd := fmt.Sprintf("login-identifier-6-%v", isAPI), "password"
			createIdentity(identifier, pwd)

			payload := url.Values{"identifier": {identifier}, "password": {"not-password"}}.Encode()
			if isAPI {
				payload = x.MustEncodeJSON(t, &password.LoginFormPayload{
					Identifier: identifier, Password: "not-password"})
			}

			lr := nlr(time.Hour, isAPI)
			res, body := fakeRequest(t, lr, isAPI, payload, nil, nil, expectStatusCode(isAPI, http.StatusBadRequest))

			assert.Equal(t, lr.ID.String(), gjson.GetBytes(body, "id").String())
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())
			ensureFieldsExist(t, body)
			assert.Equal(t,
				errorsx.Cause(schema.NewInvalidCredentialsError()).(*schema.ValidationError).Messages[0].Text,
				gjson.GetBytes(body, "methods.password.config.messages.0.text").String(),
				"%s", body,
			)

			// This must not include the password!
			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==password).value").String())
			return res
		}

		t.Run("type=browser", func(t *testing.T) {
			require.Contains(t, run(t, false).Request.URL.Path, "login-ts")
		})

		t.Run("type=api", func(t *testing.T) {
			require.Contains(t, run(t, true).Request.URL.Path, password.RouteLogin)
		})
	})

	t.Run("should pass with fake request", func(t *testing.T) {
		var identifier, pwd string
		run := func(t *testing.T, isAPI bool) (*http.Response, []byte) {
			identifier, pwd = fmt.Sprintf("login-identifier-7-%v", isAPI), "password"
			createIdentity(identifier, pwd)

			payload := url.Values{"identifier": {identifier}, "password": {pwd}}.Encode()
			if isAPI {
				payload = x.MustEncodeJSON(t, &password.LoginFormPayload{
					Identifier: identifier, Password: pwd})
			}

			lr := nlr(time.Hour, isAPI)
			return fakeRequest(t, lr, isAPI, payload, nil, nil, expectStatusCode(isAPI, http.StatusOK))
		}

		t.Run("type=browser", func(t *testing.T) {
			res, body := run(t, false)
			require.Contains(t, res.Request.URL.Path, "return-ts", "%s", res.Request.URL.String())
			assert.Equal(t, identifier, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
		})

		t.Run("type=api", func(t *testing.T) {
			res, body := run(t, true)
			require.Contains(t, res.Request.URL.Path, password.RouteLogin, "%s", res.Request.URL.String())
			assert.Equal(t, identifier, gjson.GetBytes(body, "session.identity.traits.subject").String(), "%s", body)
			assert.NotEmpty(t, gjson.GetBytes(body, "session_token").String(), "%s", body)
		})
	})
	t.Run("should pass with real request", func(t *testing.T) {
		// _= func(t *testing.T, isAPI bool, jar *cookiejar.Jar) {
		// 	identifier, pwd := fmt.Sprintf("login-identifier-8-%v", isAPI), "password"
		// 	createIdentity(identifier, pwd)
		//
		// 	payload := url.Values{"identifier": {identifier}, "password": {pwd}}.Encode()
		// 	if isAPI {
		// 		payload = x.MustEncodeJSON(t, &password.LoginFormPayload{
		// 			Identifier: identifier, Password: pwd})
		// 	}
		//
		// 	res, body := initBrowserFlow(t, payload, jar, true, http.StatusOK)
		// 	if isAPI {
		// 		require.Contains(t, res.Request.URL.Path, password.RouteLogin, "%s", res.Request.URL.String())
		// 		assert.Equal(t, identifier, gjson.GetBytes(body, "session.identity.traits.subject").String(), "%s", body)
		// 		assert.NotEmpty(t, gjson.GetBytes(body, "session_token").String(), "%s", body)
		// 	} else {
		// 		require.Contains(t, res.Request.URL.Path, "return-ts", "%s", res.Request.URL.String())
		// 		assert.Equal(t, identifier, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
		// 	}
		// }

		t.Run("type=browser", func(t *testing.T) {
			identifier, pwd := "login-identifier-8-browser", "password"
			createIdentity(identifier, pwd)
			payload := url.Values{"identifier": {identifier}, "password": {pwd}}.Encode()

			jar, _ := cookiejar.New(nil)

			res, body := initBrowserFlow(t, payload, jar, true, http.StatusOK)
			require.Contains(t, res.Request.URL.Path, "return-ts", "%s", res.Request.URL.String())
			assert.Equal(t, identifier, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)

			t.Run("retry with different refresh", func(t *testing.T) {
				c := &http.Client{Jar: jar}

				t.Run("redirect to returnTS if refresh is missing", func(t *testing.T) {
					res, err := c.Get(ts.URL + login.RouteInitBrowserFlow)
					require.NoError(t, err)
					require.EqualValues(t, http.StatusOK, res.StatusCode)
				})

				t.Run("show UI and hint at username", func(t *testing.T) {
					res, err := c.Get(ts.URL + login.RouteInitBrowserFlow + "?refresh=true")
					require.NoError(t, err)
					require.EqualValues(t, http.StatusOK, res.StatusCode)

					rid := res.Request.URL.Query().Get("flow")
					assert.NotEmpty(t, rid, "%s", res.Request.URL)

					res, err = c.Get(ts.URL + login.RouteGetFlow + "?id=" + rid)
					require.NoError(t, err)
					require.EqualValues(t, http.StatusOK, res.StatusCode)

					body, err := ioutil.ReadAll(res.Body)
					require.NoError(t, err)
					assert.True(t, gjson.GetBytes(body, "forced").Bool())
					assert.Equal(t, identifier, gjson.GetBytes(body, "methods.password.config.fields.#(name==identifier).value").String(), "%s", body)
					assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==password).value").String(), "%s", body)
				})
			})
		})

		t.Run("type=api", func(t *testing.T) {
			identifier, pwd := "login-identifier-8-api", "password"
			createIdentity(identifier, pwd)
			payload := x.MustEncodeJSON(t, &password.LoginFormPayload{Identifier: identifier, Password: pwd})

			res, body := initAPIFlow(t, payload, true, http.StatusOK)
			require.Contains(t, res.Request.URL.Path, password.RouteLogin, "%s", res.Request.URL.String())
			assert.Equal(t, identifier, gjson.GetBytes(body, "session.identity.traits.subject").String(), "%s", body)
			st := gjson.GetBytes(body, "session_token").String()
			assert.NotEmpty(t, st, "%s", body)

			t.Run("retry with different refresh", func(t *testing.T) {
				c := &http.Client{}

				t.Run("redirect to returnTS if refresh is missing", func(t *testing.T) {
					req := testhelpers.NewHTTPGetJSONRequest(t, ts.URL+login.RouteInitAPIFlow)
					req.Header.Add("Authorization", "Bearer "+st)
					res, err := c.Do(req)
					require.NoError(t, err)

					defer res.Body.Close()
					body, err := ioutil.ReadAll(res.Body)
					require.Nil(t, err)
					require.EqualValues(t, http.StatusBadRequest, res.StatusCode)
					assertx.EqualAsJSON(t, login.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
				})

				t.Run("show UI and hint at username", func(t *testing.T) {
					req := testhelpers.NewHTTPGetJSONRequest(t, ts.URL+login.RouteInitAPIFlow+"?refresh=true")
					req.Header.Add("Authorization", "Bearer "+st)
					res, err := c.Do(req)
					require.NoError(t, err)
					require.EqualValues(t, http.StatusOK, res.StatusCode)

					body, err := ioutil.ReadAll(res.Body)
					require.NoError(t, err)

					assert.True(t, gjson.GetBytes(body, "forced").Bool())
					assert.Equal(t, identifier, gjson.GetBytes(body, "methods.password.config.fields.#(name==identifier).value").String(), "%s", body)
					assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==password).value").String(), "%s", body)
				})
			})
		})
	})

	t.Run("should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
		run := func(t *testing.T, isAPI bool) {
			ft := flow.TypeBrowser
			if isAPI {
				ft = flow.TypeAPI
			}
			lr := &login.Flow{
				ID:        x.NewUUID(),
				Type:      ft,
				ExpiresAt: time.Now().Add(time.Minute),
				Methods: map[identity.CredentialsType]*login.FlowMethod{
					identity.CredentialsTypePassword: {
						Method: identity.CredentialsTypePassword,
						Config: &login.FlowMethodConfig{
							FlowMethodConfigurator: &password.RequestMethod{
								HTMLForm: &form.HTMLForm{
									Method:   "POST",
									Action:   "/action",
									Messages: text.Messages{{Text: "some error"}},
									Fields: form.Fields{
										{
											Value:    "baz",
											Name:     "identifier",
											Messages: text.Messages{{Text: "err"}},
										},
										{
											Value:    "bar",
											Name:     "password",
											Messages: text.Messages{{Text: "err"}},
										},
									},
								},
							},
						},
					},
				},
			}

			identifier := fmt.Sprintf("registration-identifier-9-%v", isAPI)
			payload := url.Values{"identifier": {identifier}}.Encode()
			if isAPI {
				payload = x.MustEncodeJSON(t, &password.LoginFormPayload{
					Identifier: identifier})
			}

			res, body := fakeRequest(t, lr, isAPI, payload, nil, nil, expectStatusCode(isAPI, http.StatusBadRequest))
			if isAPI {
				require.Contains(t, res.Request.URL.Path, password.RouteLogin)
				checkFormContent(t, body, "identifier", "password")
			} else {
				require.Contains(t, res.Request.URL.Path, "login-ts")
				checkFormContent(t, body, "identifier", "password", "csrf_token")
			}

			assert.Equal(t, lr.ID.String(), gjson.GetBytes(body, "id").String())
			assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String())

			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==identity).value"))
			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==identity).error"))
			assert.Empty(t, gjson.GetBytes(body, "methods.password.config.error"))
			assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==password).messages.0").String(), "Property password is missing.", "%s", body)
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=api", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("should be a new session with forced flag", func(t *testing.T) {
		identifier, pwd := "login-identifier-reauth", "password"
		createIdentity(identifier, pwd)

		jar, err := cookiejar.New(&cookiejar.Options{})
		require.NoError(t, err)
		_, body1 := fakeRequest(t, nlr(time.Hour, false), false, url.Values{
			"identifier": {identifier},
			"password":   {pwd},
		}.Encode(), nil, jar, http.StatusOK)

		lr2 := nlr(time.Hour, false)
		lr2.Forced = true
		res, body2 := fakeRequest(t, lr2, false, url.Values{
			"identifier": {identifier},
			"password":   {pwd},
		}.Encode(), nil, jar, http.StatusOK)

		require.Contains(t, res.Request.URL.Path, "return-ts", "%s", res.Request.URL.String())
		assert.Equal(t, identifier, gjson.GetBytes(body2, "identity.traits.subject").String(), "%s", body2)
		assert.NotEqual(t, gjson.GetBytes(body1, "id").String(), gjson.GetBytes(body2, "id").String(), "%s\n\n%s\n", body1, body2)
	})

	t.Run("should be the same session without forced flag", func(t *testing.T) {
		identifier, pwd := "login-identifier-no-reauth", "password"
		createIdentity(identifier, pwd)

		jar, err := cookiejar.New(&cookiejar.Options{})
		require.NoError(t, err)
		_, body1 := fakeRequest(t, nlr(time.Hour, false), false, url.Values{
			"identifier": {identifier},
			"password":   {pwd},
		}.Encode(), nil, jar, http.StatusOK)

		lr2 := nlr(time.Hour, false)
		res, body2 := fakeRequest(t, lr2, false, url.Values{
			"identifier": {identifier}, "password": {pwd}}.Encode(), nil, jar, http.StatusOK)

		require.Contains(t, res.Request.URL.Path, "return-ts", "%s", res.Request.URL.String())
		assert.Equal(t, identifier, gjson.GetBytes(body2, "identity.traits.subject").String(), "%s", body2)
		assert.Equal(t, gjson.GetBytes(body1, "id").String(), gjson.GetBytes(body2, "id").String(), "%s\n\n%s\n", body1, body2)
	})
}
