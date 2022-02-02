package hook_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/hook"

	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"

	"github.com/ory/kratos/session"

	"github.com/ory/kratos/selfservice/flow/login"

	"github.com/stretchr/testify/assert"
)

func TestWebHooks(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
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
		callWebHook  func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error
		expectedBody func(req *http.Request, f flow.Flow, s *session.Session) string
		createFlow   func() flow.Flow
	}{
		{
			uc:         "Pre Login Hook",
			createFlow: func() flow.Flow { return &login.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, _ *session.Session) error {
				return wh.ExecuteLoginPreHook(nil, req, f.(*login.Flow))
			},
			expectedBody: func(req *http.Request, f flow.Flow, _ *session.Session) string {
				return bodyWithFlowOnly(req, f)
			},
		},
		{
			uc:         "Post Login Hook",
			createFlow: func() flow.Flow { return &login.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteLoginPostHook(nil, req, f.(*login.Flow), s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
			},
		},
		{
			uc:         "Pre Registration Hook",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, _ *session.Session) error {
				return wh.ExecuteRegistrationPreHook(nil, req, f.(*registration.Flow))
			},
			expectedBody: func(req *http.Request, f flow.Flow, _ *session.Session) string {
				return bodyWithFlowOnly(req, f)
			},
		},
		{
			uc:         "Post Registration Hook",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRegistrationPostPersistHook(nil, req, f.(*registration.Flow), s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
			},
		},
		{
			uc:         "Post Recovery Hook",
			createFlow: func() flow.Flow { return &recovery.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRecoveryHook(nil, req, f.(*recovery.Flow), s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
			},
		},
		{
			uc:         "Post Verification Hook",
			createFlow: func() flow.Flow { return &verification.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostVerificationHook(nil, req, f.(*verification.Flow), s.Identity)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
			},
		},
		{
			uc:         "Post Settings Hook",
			createFlow: func() flow.Flow { return &settings.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
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

							wh := hook.NewWebHook(reg, conf)

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
		conf := json.RawMessage("not valid json")
		wh := hook.NewWebHook(reg, conf)

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
		wh := hook.NewWebHook(reg, conf)

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
			wh := hook.NewWebHook(reg, conf)

			err := wh.ExecuteLoginPreHook(nil, req, f)
			if tc.mustSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestDisallowPrivateIPRanges(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeyClientHTTPNoPrivateIPRanges, true)

	req := &http.Request{
		Header:     map[string][]string{"Some-Header": {"Some-Value"}},
		RequestURI: "https://www.ory.sh/some_end_point",
		Method:     http.MethodPost,
	}
	s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{ID: x.NewUUID()}}
	f := &login.Flow{ID: x.NewUUID()}
	wh := hook.NewWebHook(reg, json.RawMessage(`{
  "url": "https://localhost:1234/",
  "method": "GET",
  "body": "file://stub/test_body.jsonnet"
}`))
	err := wh.ExecuteLoginPostHook(nil, req, f, s)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ip 127.0.0.1 is in the 127.0.0.0/8 range")
}
