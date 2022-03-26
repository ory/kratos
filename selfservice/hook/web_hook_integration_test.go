// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook_test

import (
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/text"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/otelx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/logrusx"

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
	logger := logrusx.New("kratos", "test")
	whDeps := struct {
		x.SimpleLoggerWithClient
		*jsonnetsecure.TestProvider
	}{
		x.SimpleLoggerWithClient{L: logger, C: reg.HTTPClient(context.Background()), T: otelx.NewNoop(logger, &otelx.Config{ServiceName: "kratos"})},
		jsonnetsecure.NewTestProvider(t),
	}
	type WebHookRequest struct {
		Body    string
		Headers http.Header
		Method  string
	}

	webHookEndPoint := func(whr *WebHookRequest) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			body, err := io.ReadAll(r.Body)
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

	webHookHttpCodeWithBodyEndPoint := func(t *testing.T, code int, body []byte) httprouter.Handle {
		return func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			w.WriteHeader(code)
			_, err := w.Write(body)
			assert.NoError(t, err, "error while returning response from webHookHttpCodeWithBodyEndPoint")
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
					"url": "%s",
					"cookies": {
						"Some-Cookie-1": "Some-Cookie-Value",
						"Some-Cookie-2": "Some-other-Cookie-Value",
						"Some-Cookie-3": "Third-Cookie-Value"
					}
				}`, f.GetID(), string(h), req.Method, "http://www.ory.sh/some_end_point")
	}

	bodyWithFlowAndIdentity := func(req *http.Request, f flow.Flow, s *session.Session) string {
		h, _ := json.Marshal(req.Header)
		return fmt.Sprintf(`{
   					"flow_id": "%s",
					"identity_id": "%s",
   					"headers": %s,
					"method": "%s",
					"url": "%s",
					"cookies": {
						"Some-Cookie-1": "Some-Cookie-Value",
						"Some-Cookie-2": "Some-other-Cookie-Value",
						"Some-Cookie-3": "Third-Cookie-Value"
					}
				}`, f.GetID(), s.Identity.ID, string(h), req.Method, "http://www.ory.sh/some_end_point")
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
				return wh.ExecuteLoginPostHook(nil, req, node.PasswordGroup, f.(*login.Flow), s)
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
			uc:         "Pre Persist Registration Hook",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRegistrationPrePersistHook(nil, req, f.(*registration.Flow), s.Identity)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentity(req, f, s)
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
								Host: "www.ory.sh",
								Header: map[string][]string{
									"Some-Header": {"Some-Value"},
									"Cookie":      {"Some-Cookie-1=Some-Cookie-Value; Some-Cookie-2=Some-other-Cookie-Value", "Some-Cookie-3=Third-Cookie-Value"},
								},
								RequestURI: "/some_end_point",
								Method:     http.MethodPost,
								URL:        &url.URL{Path: "/some_end_point"},
							}
							cookie, err := req.Cookie("Some-Cookie-1")
							require.NoError(t, err)
							require.Equal(t, cookie.Name, "Some-Cookie-1")
							require.Equal(t, cookie.Value, "Some-Cookie-Value")
							cookie, err = req.Cookie("Some-Cookie-2")
							require.NoError(t, err)
							require.Equal(t, cookie.Name, "Some-Cookie-2")
							require.Equal(t, cookie.Value, "Some-other-Cookie-Value")
							cookie, err = req.Cookie("Some-Cookie-3")
							require.NoError(t, err)
							require.Equal(t, cookie.Name, "Some-Cookie-3")
							require.Equal(t, cookie.Value, "Third-Cookie-Value")

							s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{ID: x.NewUUID()}}
							whr := &WebHookRequest{}
							ts := newServer(webHookEndPoint(whr))
							conf := json.RawMessage(fmt.Sprintf(`{
								"url": "%s",
								"method": "%s",
								"body": "%s",
								"auth": %s
							}`, ts.URL+path, method, "file://./stub/test_body.jsonnet", auth.createAuthConfig()))

							wh := hook.NewWebHook(&whDeps, conf)

							err = tc.callWebHook(wh, req, f, s)
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

	webHookResponse := []byte(
		`{
			"messages": [{
				"instance_ptr": "#/traits/username",
				"messages": [{
					"id": 1234,
					"text": "error message",
					"type": "info"
				}]
			}]
		}`,
	)

	webhookError := schema.NewValidationListError([]*schema.ValidationError{schema.NewHookValidationError("#/traits/username", "a web-hook target returned an error", text.Messages{{ID: 1234, Type: "info", Text: "error message"}})})
	for _, tc := range []struct {
		uc              string
		callWebHook     func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error
		webHookResponse func() (int, []byte)
		createFlow      func() flow.Flow
		expectedError   error
	}{
		{
			uc:         "Pre Login Hook - no block",
			createFlow: func() flow.Flow { return &login.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, _ *session.Session) error {
				return wh.ExecuteLoginPreHook(nil, req, f.(*login.Flow))
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusOK, []byte{}
			},
			expectedError: nil,
		},
		{
			uc:         "Pre Login Hook - block",
			createFlow: func() flow.Flow { return &login.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, _ *session.Session) error {
				return wh.ExecuteLoginPreHook(nil, req, f.(*login.Flow))
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: webhookError,
		},
		{
			uc:         "Post Login Hook - no block",
			createFlow: func() flow.Flow { return &login.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteLoginPostHook(nil, req, node.PasswordGroup, f.(*login.Flow), s)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusOK, []byte{}
			},
			expectedError: nil,
		},
		{
			uc:         "Post Login Hook - block",
			createFlow: func() flow.Flow { return &login.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteLoginPostHook(nil, req, node.PasswordGroup, f.(*login.Flow), s)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: webhookError,
		},
		{
			uc:         "Pre Registration Hook - no block",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, _ *session.Session) error {
				return wh.ExecuteRegistrationPreHook(nil, req, f.(*registration.Flow))
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusOK, []byte{}
			},
			expectedError: nil,
		},
		{
			uc:         "Pre Registration Hook - block",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, _ *session.Session) error {
				return wh.ExecuteRegistrationPreHook(nil, req, f.(*registration.Flow))
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: webhookError,
		},
		{
			uc:         "Post Registration Post Persist Hook - no block",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRegistrationPostPersistHook(nil, req, f.(*registration.Flow), s)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusOK, []byte{}
			},
			expectedError: nil,
		},
		{
			uc:         "Post Registration Post Persist Hook - block has no effect",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRegistrationPostPersistHook(nil, req, f.(*registration.Flow), s)
			},
			// This would usually error, but post persist does not execute blocking web hooks, so we expect no error.
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: nil,
		},
		{
			uc:         "Post Registration Pre Persist Hook - no block",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRegistrationPrePersistHook(nil, req, f.(*registration.Flow), s.Identity)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusOK, []byte{}
			},
			expectedError: nil,
		},
		{
			uc:         "Post Registration Pre Persist Hook - block",
			createFlow: func() flow.Flow { return &registration.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRegistrationPrePersistHook(nil, req, f.(*registration.Flow), s.Identity)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: webhookError,
		},
		{
			uc:         "Post Recovery Hook - no block",
			createFlow: func() flow.Flow { return &recovery.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRecoveryHook(nil, req, f.(*recovery.Flow), s)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusOK, []byte{}
			},
			expectedError: nil,
		},
		{
			uc:         "Post Recovery Hook - block",
			createFlow: func() flow.Flow { return &recovery.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRecoveryHook(nil, req, f.(*recovery.Flow), s)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: webhookError,
		},
		{
			uc:         "Post Verification Hook - no block",
			createFlow: func() flow.Flow { return &verification.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostVerificationHook(nil, req, f.(*verification.Flow), s.Identity)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusOK, []byte{}
			},
			expectedError: nil,
		},
		{
			uc:         "Post Verification Hook - block",
			createFlow: func() flow.Flow { return &verification.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostVerificationHook(nil, req, f.(*verification.Flow), s.Identity)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: webhookError,
		},
		{
			uc:         "Post Settings Hook - no block",
			createFlow: func() flow.Flow { return &settings.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteSettingsPostPersistHook(nil, req, f.(*settings.Flow), s.Identity)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusOK, []byte{}
			},
			expectedError: nil,
		},
		{
			uc:         "Post Settings Hook Pre Persist - block",
			createFlow: func() flow.Flow { return &settings.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteSettingsPrePersistHook(nil, req, f.(*settings.Flow), s.Identity)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: webhookError,
		},
		{
			uc:         "Post Settings Hook Post Persist - block has no effect",
			createFlow: func() flow.Flow { return &settings.Flow{ID: x.NewUUID()} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteSettingsPostPersistHook(nil, req, f.(*settings.Flow), s.Identity)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: nil,
		},
	} {
		t.Run("uc="+tc.uc, func(t *testing.T) {
			for _, method := range []string{"CONNECT", "DELETE", "GET", "OPTIONS", "PATCH", "POST", "PUT", "TRACE"} {
				t.Run("method="+method, func(t *testing.T) {
					f := tc.createFlow()
					req := &http.Request{
						Host: "www.ory.sh",
						Header: map[string][]string{
							"Some-Header":       {"Some-Value"},
							"X-Forwarded-Proto": {"https"},
							"Cookie":            {"Some-Cookie-1=Some-Cookie-Value; Some-Cookie-2=Some-other-Cookie-Value", "Some-Cookie-3=Third-Cookie-Value"}},
						RequestURI: "/some_end_point",
						Method:     http.MethodPost,
						URL: &url.URL{
							Path: "some_end_point",
						},
					}
					s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{ID: x.NewUUID()}}
					code, res := tc.webHookResponse()
					ts := newServer(webHookHttpCodeWithBodyEndPoint(t, code, res))
					conf := json.RawMessage(fmt.Sprintf(`{
								"url": "%s",
								"method": "%s",
								"body": "%s",
								"can_interrupt": true
							}`, ts.URL+path, method, "file://./stub/test_body.jsonnet"))

					wh := hook.NewWebHook(&whDeps, conf)

					err := tc.callWebHook(wh, req, f, s)
					if tc.expectedError == nil {
						assert.NoError(t, err)
						return
					}

					var validationError *schema.ValidationListError
					var expectedError *schema.ValidationListError
					if assert.ErrorAs(t, err, &validationError) && assert.ErrorAs(t, tc.expectedError, &expectedError) {
						assert.Equal(t, expectedError, validationError)
					}
				})
			}
		})
	}

	t.Run("must error when config is erroneous", func(t *testing.T) {
		req := &http.Request{
			Header: map[string][]string{"Some-Header": {"Some-Value"}},
			Host:   "www.ory.sh",
			TLS:    new(tls.ConnectionState),
			URL:    &url.URL{Path: "/some_end_point"},

			Method: http.MethodPost,
		}
		f := &login.Flow{ID: x.NewUUID()}
		conf := json.RawMessage("not valid json")
		wh := hook.NewWebHook(&whDeps, conf)

		err := wh.ExecuteLoginPreHook(nil, req, f)
		assert.Error(t, err)
	})

	t.Run("must error when template is erroneous", func(t *testing.T) {
		ts := newServer(webHookHttpCodeEndPoint(200))
		req := &http.Request{
			Header: map[string][]string{"Some-Header": {"Some-Value"}},
			Host:   "www.ory.sh",
			TLS:    new(tls.ConnectionState),
			URL:    &url.URL{Path: "/some_end_point"},
			Method: http.MethodPost,
		}
		f := &login.Flow{ID: x.NewUUID()}
		conf := json.RawMessage(fmt.Sprintf(`{
					"url": "%s",
					"method": "%s",
					"body": "%s"
				}`, ts.URL+path, "POST", "file://./stub/bad_template.jsonnet"))
		wh := hook.NewWebHook(&whDeps, conf)

		err := wh.ExecuteLoginPreHook(nil, req, f)
		assert.Error(t, err)
	})

	t.Run("must not make request", func(t *testing.T) {
		req := &http.Request{
			Header: map[string][]string{"Some-Header": {"Some-Value"}},
			Host:   "www.ory.sh",
			TLS:    new(tls.ConnectionState),
			URL:    &url.URL{Path: "/some_end_point"},

			Method: http.MethodPost,
		}
		f := &login.Flow{ID: x.NewUUID()}
		conf := json.RawMessage(`{
	"url": "https://i-do-not-exist/",
	"method": "POST",
	"body": "./stub/cancel_template.jsonnet"
}`)
		wh := hook.NewWebHook(&whDeps, conf)

		err := wh.ExecuteLoginPreHook(nil, req, f)
		assert.NoError(t, err)
	})

	boolToString := func(f bool) string {
		if f {
			return " not"
		} else {
			return ""
		}
	}

	t.Run("ignores the response and is async", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		waitTime := time.Millisecond * 100
		ts := newServer(func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			defer wg.Done()
			time.Sleep(waitTime)
			w.WriteHeader(http.StatusBadRequest)
		})

		req := &http.Request{
			Header: map[string][]string{"Some-Header": {"Some-Value"}},
			Host:   "www.ory.sh",
			TLS:    new(tls.ConnectionState),
			URL:    &url.URL{Path: "/some_end_point"},
			Method: http.MethodPost,
		}
		f := &login.Flow{ID: x.NewUUID()}
		conf := json.RawMessage(fmt.Sprintf(`{"url": "%s", "method": "GET", "body": "./stub/test_body.jsonnet", "response": {"ignore": true}}`, ts.URL+path))
		wh := hook.NewWebHook(&whDeps, conf)

		start := time.Now()
		err := wh.ExecuteLoginPreHook(nil, req, f)
		assert.NoError(t, err)
		assert.True(t, time.Since(start) < waitTime)

		wg.Wait()
	})

	t.Run("does not error on 500 request with retry", func(t *testing.T) {
		// This test essentially ensures that we do not regress on the bug we had where 500 status code
		// would cause a retry, but because the body was incorrectly set we ended up with a ContentLength
		// error.

		var wg sync.WaitGroup
		wg.Add(3) // HTTP client does 3 attempts
		ts := newServer(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			defer wg.Done()
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{"error":"some error"}`))
		})

		req := &http.Request{
			Header: map[string][]string{"Some-Header": {"Some-Value"}},
			Host:   "www.ory.sh",
			TLS:    new(tls.ConnectionState),
			URL:    &url.URL{Path: "/some_end_point"},
			Method: http.MethodPost,
		}
		f := &login.Flow{ID: x.NewUUID()}
		conf := json.RawMessage(fmt.Sprintf(`{"url": "%s", "method": "GET", "body": "./stub/test_body.jsonnet"}`, ts.URL+path))
		wh := hook.NewWebHook(&whDeps, conf)

		err := wh.ExecuteLoginPreHook(nil, req, f)
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "ContentLength")

		wg.Wait()
	})

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
				Header: map[string][]string{"Some-Header": {"Some-Value"}},
				Host:   "www.ory.sh",
				TLS:    new(tls.ConnectionState),
				URL:    &url.URL{Path: "/some_end_point"},
				Method: http.MethodPost,
			}
			f := &login.Flow{ID: x.NewUUID()}
			conf := json.RawMessage(fmt.Sprintf(`{
					"url": "%s",
					"method": "%s",
					"body": "%s"
				}`, ts.URL+path, "POST", "file://./stub/test_body.jsonnet"))
			wh := hook.NewWebHook(&whDeps, conf)

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
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyClientHTTPNoPrivateIPRanges, true)
	conf.MustSet(ctx, config.ViperKeyClientHTTPPrivateIPExceptionURLs, []string{"http://localhost/exception"})
	logger := logrusx.New("kratos", "test")
	whDeps := struct {
		x.SimpleLoggerWithClient
		*jsonnetsecure.TestProvider
	}{
		x.SimpleLoggerWithClient{L: logger, C: reg.HTTPClient(context.Background()), T: otelx.NewNoop(logger, &otelx.Config{ServiceName: "kratos"})},
		jsonnetsecure.NewTestProvider(t),
	}

	req := &http.Request{
		Header: map[string][]string{"Some-Header": {"Some-Value"}},
		Host:   "www.ory.sh",
		TLS:    new(tls.ConnectionState),
		URL:    &url.URL{Path: "/some_end_point"},
		Method: http.MethodPost,
	}
	s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{ID: x.NewUUID()}}
	f := &login.Flow{ID: x.NewUUID()}

	t.Run("not allowed to call url", func(t *testing.T) {
		wh := hook.NewWebHook(&whDeps, json.RawMessage(`{
  "url": "https://localhost:1234/",
  "method": "GET",
  "body": "file://stub/test_body.jsonnet"
}`))
		err := wh.ExecuteLoginPostHook(nil, req, node.DefaultGroup, f, s)
		require.Error(t, err)
		require.Contains(t, err.Error(), "private, loopback, or unspecified IP range")
	})

	t.Run("allowed to call exempt url", func(t *testing.T) {
		wh := hook.NewWebHook(&whDeps, json.RawMessage(`{
  "url": "http://localhost/exception",
  "method": "GET",
  "body": "file://stub/test_body.jsonnet"
}`))
		err := wh.ExecuteLoginPostHook(nil, req, node.DefaultGroup, f, s)
		require.Error(t, err, "the target does not exist and we still receive an error")
		require.NotContains(t, err.Error(), "is in the private, loopback, or unspecified IP range", "but the error is not related to the IP range.")
	})

	t.Run("not allowed to load from source", func(t *testing.T) {
		req := &http.Request{
			Header: map[string][]string{"Some-Header": {"Some-Value"}},
			Host:   "www.ory.sh",
			TLS:    new(tls.ConnectionState),
			URL:    &url.URL{Path: "/some_end_point"},
			Method: http.MethodPost,
		}
		s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{ID: x.NewUUID()}}
		f := &login.Flow{ID: x.NewUUID()}
		wh := hook.NewWebHook(&whDeps, json.RawMessage(`{
  "url": "https://www.google.com/",
  "method": "GET",
  "body": "http://192.168.178.0/test_body.jsonnet"
}`))
		err := wh.ExecuteLoginPostHook(nil, req, node.DefaultGroup, f, s)
		require.Error(t, err)
		require.Contains(t, err.Error(), "ip 192.168.178.0 is in the private, loopback, or unspecified IP range")
	})
}
