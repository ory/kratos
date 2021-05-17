package hook

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"

	"github.com/ory/kratos/session"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/flow/login"

	"github.com/stretchr/testify/assert"
)

func TestNoopAuthStrategy(t *testing.T) {
	req := http.Request{Header: map[string][]string{}}
	auth := noopAuthStrategy{}

	auth.apply(&req)

	assert.Empty(t, req.Header, "Empty auth strategy shall not modify any request headers")
}

func TestBasicAuthStrategy(t *testing.T) {
	req := http.Request{Header: map[string][]string{}}
	auth := basicAuthStrategy{
		user:     "test-user",
		password: "test-pass",
	}

	auth.apply(&req)

	assert.Len(t, req.Header, 1)

	user, pass, _ := req.BasicAuth()
	assert.Equal(t, "test-user", user)
	assert.Equal(t, "test-pass", pass)
}

func TestApiKeyInHeaderStrategy(t *testing.T) {
	req := http.Request{Header: map[string][]string{}}
	auth := apiKeyStrategy{
		in:    "header",
		name:  "my-api-key-name",
		value: "my-api-key-value",
	}

	auth.apply(&req)

	require.Len(t, req.Header, 1)

	actualValue := req.Header.Get("my-api-key-name")
	assert.Equal(t, "my-api-key-value", actualValue)
}

func TestApiKeyInCookieStrategy(t *testing.T) {
	req := http.Request{Header: map[string][]string{}}
	auth := apiKeyStrategy{
		in:    "cookie",
		name:  "my-api-key-name",
		value: "my-api-key-value",
	}

	auth.apply(&req)

	cookies := req.Cookies()
	assert.Len(t, cookies, 1)

	assert.Equal(t, "my-api-key-name", cookies[0].Name)
	assert.Equal(t, "my-api-key-value", cookies[0].Value)
}

func TestJsonNetSupport(t *testing.T) {
	f := login.Flow{ID: x.NewUUID()}
	i := identity.NewIdentity("")

	headers := http.Header{}
	headers.Add("Some-Header", "Some-Value")
	headers.Add("Cookie", "c1=v1")
	headers.Add("Cookie", "c2=v2")
	data := &templateContext{
		Flow:           &f,
		RequestHeaders: headers,
		RequestMethod:  "POST",
		RequestUrl:     "https://test.kratos.ory.sh/some-test-path",
		Identity:       i,
	}

	b, err := createBody("./stub/test_body.jsonnet", data)
	assert.NoError(t, err)

	buf := new(strings.Builder)
	io.Copy(buf, b)

	expected := fmt.Sprintf(`
		{
			"flow_id": "%s",
			"identity_id": "%s",
			"headers": {
				"Cookie": ["%s", "%s"],
				"Some-Header": ["%s"]
			},
			"method": "%s",
			"url": "%s"
		}`,
		f.ID, i.ID,
		data.RequestHeaders.Values("Cookie")[0],
		data.RequestHeaders.Values("Cookie")[1],
		data.RequestHeaders.Get("Some-Header"),
		data.RequestMethod,
		data.RequestUrl)

	assert.JSONEq(t, expected, buf.String())
}

func TestWebHookConfig(t *testing.T) {
	for _, tc := range []struct {
		strategy     string
		method       string
		url          string
		body         string
		rawConfig    string
		authStrategy AuthStrategy
	}{
		{
			strategy: "empty",
			method:   "POST",
			url:      "https://test.kratos.ory.sh/my_hook1",
			body:     "/path/to/my/jsonnet1.file",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_hook1",
				"method": "POST",
				"body": "/path/to/my/jsonnet1.file"
			}`,
			authStrategy: &noopAuthStrategy{},
		},
		{
			strategy: "basic_auth",
			method:   "GET",
			url:      "https://test.kratos.ory.sh/my_hook2",
			body:     "/path/to/my/jsonnet2.file",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_hook2",
				"method": "GET",
				"body": "/path/to/my/jsonnet2.file",
				"auth": {
					"type": "basic_auth",
					"config": {
						"user": "test-api-user",
						"password": "secret"
					}
				}
			}`,
			authStrategy: &basicAuthStrategy{},
		},
		{
			strategy: "api-key/header",
			method:   "DELETE",
			url:      "https://test.kratos.ory.sh/my_hook3",
			body:     "/path/to/my/jsonnet3.file",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_hook3",
				"method": "DELETE",
				"body": "/path/to/my/jsonnet3.file",
				"auth": {
					"type": "api_key",
					"config": {
						"in": "header",
						"name": "my-api-key",
						"value": "secret"
					}
				}
			}`,
			authStrategy: &apiKeyStrategy{},
		},
		{
			strategy: "api-key/cookie",
			method:   "POST",
			url:      "https://test.kratos.ory.sh/my_hook4",
			body:     "/path/to/my/jsonnet4.file",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_hook4",
				"method": "POST",
				"body": "/path/to/my/jsonnet4.file",
				"auth": {
					"type": "api_key",
					"config": {
						"in": "cookie",
						"name": "my-api-key",
						"value": "secret"
					}
				}
			}`,
			authStrategy: &apiKeyStrategy{},
		},
	} {
		t.Run("auth-strategy="+tc.strategy, func(t *testing.T) {
			conf, err := newWebHookConfig([]byte(tc.rawConfig))
			assert.Nil(t, err)

			assert.Equal(t, tc.url, conf.url)
			assert.Equal(t, tc.method, conf.method)
			assert.Equal(t, tc.body, conf.templatePath)
			assert.NotNil(t, conf.auth)
			assert.IsTypef(t, tc.authStrategy, conf.auth, "Auth should be of the expected type")
		})
	}
}

func TestWebHooks(t *testing.T) {
	type WebHookRequest struct {
		Body    string
		Headers http.Header
		Method  string
	}

	webHookEndPoint := func(whr *WebHookRequest) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			whr.Body = string(body)
			whr.Headers = r.Header
			whr.Method = r.Method
		}
	}

	webHookHttpCodeEndPoint := func(code int) httprouter.Handle {
		return func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			w.WriteHeader(code)
		}
	}

	path := "/web_hook"
	newServer := func(f httprouter.Handle) *httptest.Server {
		r := httprouter.New()

		r.Handle("CONNECT", path, f)
		r.DELETE(path, f)
		r.GET(path, f)
		r.OPTIONS(path, f)
		r.PATCH(path, f)
		r.POST(path, f)
		r.PUT(path, f)
		r.Handle("TRACE", path, f)

		ts := httptest.NewServer(r)
		t.Cleanup(ts.Close)
		return ts
	}

	bodyWithFlowOnly := func(req *http.Request, f flow.Flow) string {
		h, _ := json.Marshal(req.Header)
		return fmt.Sprintf(`{
   					"flow_id": "%s",
					"identity_id": null,
   					"headers": %s,
					"method": "%s",
					"url": "%s"
				}`, f.GetID(), string(h), req.Method, req.RequestURI)
	}

	bodyWithFlowAndIdentity := func(req *http.Request, f flow.Flow, s *session.Session) string {
		h, _ := json.Marshal(req.Header)
		return fmt.Sprintf(`{
   					"flow_id": "%s",
					"identity_id": "%s",
   					"headers": %s,
					"method": "%s",
					"url": "%s"
				}`, f.GetID(), s.Identity.ID, string(h), req.Method, req.RequestURI)
	}

	for _, tc := range []struct {
		uc           string
		callWebHook  func(wh *WebHook, req *http.Request, f flow.Flow, s *session.Session) error
		expectedBody func(req *http.Request, f flow.Flow, s *session.Session) string
		createFlow   func() flow.Flow
	}{
		{
			uc:         "Pre Login Hook",
			createFlow: func() flow.Flow { return &login.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *WebHook, req *http.Request, f flow.Flow, _ *session.Session) error {
				return wh.ExecuteLoginPreHook(nil, req, f.(*login.Flow))
			},
			expectedBody: func(req *http.Request, f flow.Flow, _ *session.Session) string {
				return bodyWithFlowOnly(req, f)
			},
		},
		{
			uc:         "Post Login Hook",
			createFlow: func() flow.Flow { return &login.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteLoginPostHook(nil, req, f.(*login.Flow), s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
			},
		},
		{
			uc:         "Pre Registration Hook",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *WebHook, req *http.Request, f flow.Flow, _ *session.Session) error {
				return wh.ExecuteRegistrationPreHook(nil, req, f.(*registration.Flow))
			},
			expectedBody: func(req *http.Request, f flow.Flow, _ *session.Session) string {
				return bodyWithFlowOnly(req, f)
			},
		},
		{
			uc:         "Post Registration Hook",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRegistrationPostPersistHook(nil, req, f.(*registration.Flow), s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
			},
		},
		{
			uc:         "Post Recovery Hook",
			createFlow: func() flow.Flow { return &recovery.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRecoveryHook(nil, req, f.(*recovery.Flow), s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
			},
		},
		{
			uc:         "Post Verification Hook",
			createFlow: func() flow.Flow { return &verification.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostVerificationHook(nil, req, f.(*verification.Flow), s.Identity)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
			},
		},
		{
			uc:         "Post Settings Hook",
			createFlow: func() flow.Flow { return &settings.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteSettingsPostPersistHook(nil, req, f.(*settings.Flow), s.Identity)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
			},
		},
	} {
		t.Run("uc="+tc.uc, func(t *testing.T) {
			for _, auth := range []struct {
				uc               string
				createAuthConfig func() string
				expectedHeader   func(header http.Header)
			}{
				{
					uc:               "no auth",
					createAuthConfig: func() string { return "{}" },
					expectedHeader:   func(header http.Header) {},
				},
				{
					uc: "api key in header",
					createAuthConfig: func() string {
						return `{
							"type": "api_key",
							"config": {
								"name": "My-Key",
								"value": "My-Key-Value",
								"in": "header"
							}
						}`
					},
					expectedHeader: func(header http.Header) {
						header.Set("My-Key", "My-Key-Value")
					},
				},
				{
					uc: "api key in cookie",
					createAuthConfig: func() string {
						return `{
							"type": "api_key",
							"config": {
								"name": "My-Key",
								"value": "My-Key-Value",
								"in": "cookie"
							}
						}`
					},
					expectedHeader: func(header http.Header) {
						header.Set("Cookie", "My-Key=My-Key-Value")
					},
				},
				{
					uc: "basic auth",
					createAuthConfig: func() string {
						return `{
							"type": "basic_auth",
							"config": {
								"user": "My-User",
								"password": "Super-Secret"
							}
						}`
					},
					expectedHeader: func(header http.Header) {
						header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("My-User:Super-Secret")))
					},
				},
			} {
				t.Run("auth="+auth.uc, func(t *testing.T) {
					for _, method := range []string{"CONNECT", "DELETE", "GET", "OPTIONS", "PATCH", "POST", "PUT", "TRACE", "GARBAGE"} {
						t.Run("method="+method, func(t *testing.T) {
							f := tc.createFlow()
							req := &http.Request{
								Header:     map[string][]string{"Some-Header": {"Some-Value"}},
								RequestURI: "https://www.ory.sh/some_end_point",
								Method:     http.MethodPost,
							}
							s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{ID: x.NewUUID()}}
							whr := &WebHookRequest{}
							ts := newServer(webHookEndPoint(whr))
							conf := json.RawMessage(fmt.Sprintf(`{
								"url": "%s",
								"method": "%s",
								"body": "%s",
								"auth": %s
							}`, ts.URL+path, method, "./stub/test_body.jsonnet", auth.createAuthConfig()))
							wh := NewWebHook(nil, conf)

							err := tc.callWebHook(wh, req, f, s)
							if method == "GARBAGE" {
								assert.Error(t, err)
								return
							}

							assert.NoError(t, err)

							assert.Equal(t, method, whr.Method)

							expectedHeader := http.Header{}
							expectedHeader.Set("Content-Type", "application/json")
							auth.expectedHeader(expectedHeader)
							for k, v := range expectedHeader {
								vals := whr.Headers.Values(k)
								assert.Equal(t, v, vals)
							}

							if method != "TRACE" {
								// According to the HTTP spec any request method, but TRACE is allowed to
								// have a body. Even this is a really bad practice for some of them, like for
								// GET
								assert.JSONEq(t, tc.expectedBody(req, f, s), whr.Body)
							} else {
								assert.Emptyf(t, whr.Body, "HTTP %s is not allowed to have a body", method)
							}
						})
					}
				})
			}
		})
	}

	t.Run("Must error when config is erroneous", func(t *testing.T) {
		req := &http.Request{
			Header:     map[string][]string{"Some-Header": {"Some-Value"}},
			RequestURI: "https://www.ory.sh/some_end_point",
			Method:     http.MethodPost,
		}
		f := &login.Flow{ID: x.NewUUID()}
		conf := json.RawMessage(fmt.Sprintf(`{
					"url": "/foo",
					"method": "BAR"
				`)) // closing } is missing
		wh := NewWebHook(nil, conf)

		err := wh.ExecuteLoginPreHook(nil, req, f)
		assert.Error(t, err)
	})

	t.Run("Must error when template is erroneous", func(t *testing.T) {
		ts := newServer(webHookHttpCodeEndPoint(200))
		req := &http.Request{
			Header:     map[string][]string{"Some-Header": {"Some-Value"}},
			RequestURI: "https://www.ory.sh/some_end_point",
			Method:     http.MethodPost,
		}
		f := &login.Flow{ID: x.NewUUID()}
		conf := json.RawMessage(fmt.Sprintf(`{
					"url": "%s",
					"method": "%s",
					"body": "%s"
				}`, ts.URL+path, "POST", "./stub/bad_template.jsonnet"))
		wh := NewWebHook(nil, conf)

		err := wh.ExecuteLoginPreHook(nil, req, f)
		assert.Error(t, err)
	})

	boolToString := func(f bool) string {
		if f {
			return " not"
		} else {
			return ""
		}
	}

	for _, tc := range []struct {
		code        int
		mustSuccess bool
	}{
		{200, true},
		{299, true},
		{300, true},
		{399, true},
		{400, false},
		{499, false},
		{500, false},
		{599, false},
	} {
		t.Run("Must"+boolToString(tc.mustSuccess)+" error when end point is returning "+strconv.Itoa(tc.code), func(t *testing.T) {
			ts := newServer(webHookHttpCodeEndPoint(tc.code))
			req := &http.Request{
				Header:     map[string][]string{"Some-Header": {"Some-Value"}},
				RequestURI: "https://www.ory.sh/some_end_point",
				Method:     http.MethodPost,
			}
			f := &login.Flow{ID: x.NewUUID()}
			conf := json.RawMessage(fmt.Sprintf(`{
					"url": "%s",
					"method": "%s",
					"body": "%s"
				}`, ts.URL+path, "POST", "./stub/test_body.jsonnet"))
			wh := NewWebHook(nil, conf)

			err := wh.ExecuteLoginPreHook(nil, req, f)
			if tc.mustSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
