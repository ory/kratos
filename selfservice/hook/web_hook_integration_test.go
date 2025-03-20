// Copyright © 2023 Ory Corp
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
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/events"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/otelx/semconv"
	"github.com/ory/x/snapshotx"
)

var transientPayload = json.RawMessage(`{
	"stuff": {
		"name": "fubar",
		"numbers": [42, 12345, 3.1415]
	}
}`)

func TestWebHooks(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	logger := logrusx.New("kratos", "test")

	conf.Set(ctx, config.ViperKeyWebhookHeaderAllowlist, []string{
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"Content-Length",
		"Content-Type",
		"Origin",
		"Priority",
		"Referer",
		"Sec-Ch-Ua",
		"Sec-Ch-Ua-Mobile",
		"Sec-Ch-Ua-Platform",
		"Sec-Fetch-Dest",
		"Sec-Fetch-Mode",
		"Sec-Fetch-Site",
		"Sec-Fetch-User",
		"True-Client-Ip",
		"User-Agent",
		"Valid-Header",
	})

	whDeps := struct {
		x.SimpleLoggerWithClient
		*jsonnetsecure.TestProvider
		config.Provider
	}{
		x.SimpleLoggerWithClient{L: logger, C: reg.HTTPClient(ctx), T: otelx.NewNoop(logger, &otelx.Config{ServiceName: "kratos"})},
		jsonnetsecure.NewTestProvider(t),
		reg,
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
		body := fmt.Sprintf(`{
					"flow_id": "%s",
					"method": "%s",
					"url": "%s",
					"cookies": {
						"Some-Cookie-1": "Some-Cookie-Value",
						"Some-Cookie-2": "Some-other-Cookie-Value",
						"Some-Cookie-3": "Third-Cookie-Value"
					}
				}`, f.GetID(), req.Method, "http://www.ory.sh/some_end_point")
		if len(req.Header) != 0 {
			if ua := req.Header.Get("User-Agent"); ua != "" {
				body, _ = sjson.Set(body, "headers.User-Agent", []string{ua})
			}
		}

		return body
	}

	bodyWithFlowAndIdentityAndTransientPayload := func(req *http.Request, f flow.Flow, s *session.Session, tp json.RawMessage) string {
		body := fmt.Sprintf(`{
					"flow_id": "%s",
					"identity_id": "%s",
					"method": "%s",
					"url": "%s",
					"cookies": {
						"Some-Cookie-1": "Some-Cookie-Value",
						"Some-Cookie-2": "Some-other-Cookie-Value",
						"Some-Cookie-3": "Third-Cookie-Value"
					},
					"transient_payload": %s
				}`, f.GetID(), s.Identity.ID, req.Method, "http://www.ory.sh/some_end_point", string(tp))
		if len(req.Header) != 0 {
			if ua := req.Header.Get("User-Agent"); ua != "" {
				body, _ = sjson.Set(body, "headers.User-Agent", []string{ua})
			}
		}

		return body
	}

	bodyWithFlowAndIdentityAndSessionAndTransientPayload := func(req *http.Request, f flow.Flow, s *session.Session, tp json.RawMessage) string {
		body := fmt.Sprintf(`{
					"flow_id": "%s",
					"identity_id": "%s",
					"session_id": "%s",
					"method": "%s",
					"url": "%s",
					"cookies": {
						"Some-Cookie-1": "Some-Cookie-Value",
						"Some-Cookie-2": "Some-other-Cookie-Value",
						"Some-Cookie-3": "Third-Cookie-Value"
					},
					"transient_payload": %s
				}`, f.GetID(), s.Identity.ID, s.ID, req.Method, "http://www.ory.sh/some_end_point", string(tp))
		if len(req.Header) != 0 {
			if ua := req.Header.Get("User-Agent"); ua != "" {
				body, _ = sjson.Set(body, "headers.User-Agent", []string{ua})
			}
		}

		return body
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
			createFlow: func() flow.Flow { return &login.Flow{ID: x.NewUUID(), TransientPayload: transientPayload} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteLoginPostHook(nil, req, node.PasswordGroup, f.(*login.Flow), s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentityAndSessionAndTransientPayload(req, f, s, transientPayload)
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
			uc: "Post Registration Hook",
			createFlow: func() flow.Flow {
				return &registration.Flow{
					ID:               x.NewUUID(),
					TransientPayload: transientPayload,
				}
			},
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRegistrationPostPersistHook(nil, req, f.(*registration.Flow), s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentityAndTransientPayload(req, f, s, transientPayload)
			},
		},
		{
			uc:         "Post Recovery Hook",
			createFlow: func() flow.Flow { return &recovery.Flow{ID: x.NewUUID(), TransientPayload: transientPayload} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostRecoveryHook(nil, req, f.(*recovery.Flow), s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentityAndTransientPayload(req, f, s, transientPayload)
			},
		},
		{
			uc:         "Post Verification Hook",
			createFlow: func() flow.Flow { return &verification.Flow{ID: x.NewUUID(), TransientPayload: transientPayload} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecutePostVerificationHook(nil, req, f.(*verification.Flow), s.Identity)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentityAndTransientPayload(req, f, s, transientPayload)
			},
		},
		{
			uc:         "Post Settings Hook",
			createFlow: func() flow.Flow { return &settings.Flow{ID: x.NewUUID(), TransientPayload: transientPayload} },
			callWebHook: func(wh *hook.WebHook, req *http.Request, f flow.Flow, s *session.Session) error {
				return wh.ExecuteSettingsPostPersistHook(nil, req, f.(*settings.Flow), s.Identity, s)
			},
			expectedBody: func(req *http.Request, f flow.Flow, s *session.Session) string {
				return bodyWithFlowAndIdentityAndTransientPayload(req, f, s, transientPayload)
			},
		},
	} {
		tc := tc
		t.Run("uc="+tc.uc, func(t *testing.T) {
			t.Parallel()
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
							req := (&http.Request{
								Host: "www.ory.sh",
								Header: map[string][]string{
									"Some-Header":    {"Some-Value"},
									"User-Agent":     {"Foo-Bar-Browser"},
									"Invalid-Header": {"should be ignored"},
									"Valid-Header":   {"should not be ignored"},
									"Cookie":         {"Some-Cookie-1=Some-Cookie-Value; Some-Cookie-2=Some-other-Cookie-Value", "Some-Cookie-3=Third-Cookie-Value"},
								},
								RequestURI: "/some_end_point",
								Method:     http.MethodPost,
								URL:        &url.URL{Path: "/some_end_point"},
							}).WithContext(ctx)
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
							assert.NotZero(t, whr.Headers.Get("Ory-Webhook-Request-ID"))
							assert.NotZero(t, whr.Headers.Get("Ory-Webhook-Trigger-ID"))

							if method != "TRACE" {
								// According to the HTTP spec any request method, but TRACE is allowed to
								// have a body. Even this is a really bad practice for some of them, like for
								// GET
								assert.Zero(t, gjson.Get(whr.Body, "headers.Invalid-Header"))
								assert.NotZero(t, gjson.Get(whr.Body, "headers.Valid-Header"))
								whr.Body, err = sjson.Delete(whr.Body, "headers.Valid-Header")
								assert.NoError(t, err)

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

	webhookError := schema.NewValidationListError([]*schema.ValidationError{schema.NewHookValidationError("#/traits/username", "a webhook target returned an error", text.Messages{{ID: 1234, Type: "info", Text: "error message"}})})
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
				return wh.ExecuteSettingsPostPersistHook(nil, req, f.(*settings.Flow), s.Identity, s)
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
				return wh.ExecuteSettingsPostPersistHook(nil, req, f.(*settings.Flow), s.Identity, s)
			},
			webHookResponse: func() (int, []byte) {
				return http.StatusBadRequest, webHookResponse
			},
			expectedError: nil,
		},
	} {
		tc := tc
		t.Run("uc="+tc.uc, func(t *testing.T) {
			t.Parallel()
			for _, method := range []string{"CONNECT", "DELETE", "GET", "OPTIONS", "PATCH", "POST", "PUT", "TRACE"} {
				t.Run("method="+method, func(t *testing.T) {
					f := tc.createFlow()
					req := &http.Request{
						Host: "www.ory.sh",
						Header: map[string][]string{
							"Some-Header":       {"Some-Value"},
							"X-Forwarded-Proto": {"https"},
							"Cookie":            {"Some-Cookie-1=Some-Cookie-Value; Some-Cookie-2=Some-other-Cookie-Value", "Some-Cookie-3=Third-Cookie-Value"},
						},
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

	t.Run("update identity fields", func(t *testing.T) {
		t.Parallel()
		run := func(t *testing.T, id identity.Identity, responseCode int, response []byte) *identity.WithCredentialsAndAdminMetadataInJSON {
			f := &registration.Flow{ID: x.NewUUID()}
			req := &http.Request{
				Host:       "www.ory.sh",
				Header:     map[string][]string{},
				RequestURI: "/some_end_point",
				Method:     http.MethodPost,
				URL:        &url.URL{Path: "some_end_point"},
			}
			ts := newServer(webHookHttpCodeWithBodyEndPoint(t, responseCode, response))
			conf := json.RawMessage(fmt.Sprintf(`{"url": "%s", "method": "POST", "body": "%s", "response": {"parse":true}}`, ts.URL+path, "file://./stub/test_body.jsonnet"))
			wh := hook.NewWebHook(&whDeps, conf)
			in := &id
			err := wh.ExecutePostRegistrationPrePersistHook(nil, req, f, in)
			require.NoError(t, err)
			result := identity.WithCredentialsAndAdminMetadataInJSON(*in)
			return &result
		}

		t.Run("case=update identity fields", func(t *testing.T) {
			expected := identity.Identity{
				Credentials: map[identity.CredentialsType]identity.Credentials{identity.CredentialsTypePassword: {Type: "password", Identifiers: []string{"test"}, Config: []byte(`{"hashed_password":"$argon2id$v=19$m=65536,t=1,p=1$Z3JlZW5hbmRlcnNlY3JldA$Z3JlZW5hbmRlcnNlY3JldA"}`)}},
				SchemaID:    "default",
				SchemaURL:   "file://stub/default.schema.json",
				State:       identity.StateActive,
				Traits:      []byte(`{"email":"some@example.org"}`),
				VerifiableAddresses: []identity.VerifiableAddress{{
					Value:    "some@example.org",
					Verified: false,
					Via:      "email",
					Status:   identity.VerifiableAddressStatusPending,
				}},
				RecoveryAddresses: []identity.RecoveryAddress{{
					Value: "some@example.org",
					Via:   "email",
				}},
				MetadataPublic: []byte(`{"public":"data"}`),
				MetadataAdmin:  []byte(`{"admin":"data"}`),
			}

			t.Run("case=body is empty", func(t *testing.T) {
				actual := run(t, expected, http.StatusOK, []byte(`{}`))
				snapshotx.SnapshotT(t, &actual)
			})

			t.Run("case=identity is present but empty", func(t *testing.T) {
				actual := run(t, expected, http.StatusOK, []byte(`{"identity":{}}`))
				snapshotx.SnapshotT(t, &actual)
			})

			t.Run("case=identity has updated traits", func(t *testing.T) {
				actual := run(t, expected, http.StatusOK, []byte(`{"identity":{"traits":{"email":"some@other-example.org"}}}`))
				snapshotx.SnapshotT(t, &actual)
			})

			t.Run("case=identity has updated state", func(t *testing.T) {
				actual := run(t, expected, http.StatusOK, []byte(`{"identity":{"state":"inactive"}}`))
				snapshotx.SnapshotT(t, &actual)
			})

			t.Run("case=identity has updated public metadata", func(t *testing.T) {
				actual := run(t, expected, http.StatusOK, []byte(`{"identity":{"metadata_public":{"useful":"metadata"}}}`))
				snapshotx.SnapshotT(t, &actual)
			})

			t.Run("case=identity has updated admin metadata", func(t *testing.T) {
				actual := run(t, expected, http.StatusOK, []byte(`{"identity":{"metadata_admin":{"useful":"admin metadata"}}}`))
				snapshotx.SnapshotT(t, &actual)
			})

			t.Run("case=identity has updated verified addresses", func(t *testing.T) {
				actual := run(t, expected, http.StatusOK, []byte(`{"identity":{"traits":{"email":"some@other-example.org"},"verifiable_addresses":[{"value":"some@other-example.org","via":"email"}]}}`))
				snapshotx.SnapshotT(t, &actual)
			})

			t.Run("case=identity has updated recovery addresses", func(t *testing.T) {
				actual := run(t, expected, http.StatusOK, []byte(`{"identity":{"traits":{"email":"some@other-example.org"},"recovery_addresses":[{"value":"some@other-example.org","via":"email"}]}}`))
				snapshotx.SnapshotT(t, &actual)
			})
		})
	})

	t.Run("must error when config is erroneous", func(t *testing.T) {
		t.Parallel()
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

	t.Run("cannot have parse and ignore both set", func(t *testing.T) {
		t.Parallel()
		ts := newServer(webHookHttpCodeEndPoint(200))
		req := &http.Request{
			Header: map[string][]string{"Some-Header": {"Some-Value"}},
			Host:   "www.ory.sh",
			TLS:    new(tls.ConnectionState),
			URL:    &url.URL{Path: "/some_end_point"},

			Method: http.MethodPost,
		}
		f := &login.Flow{ID: x.NewUUID()}
		conf := json.RawMessage(fmt.Sprintf(`{"url": "%s", "method": "GET", "body": "./stub/test_body.jsonnet", "response": {"ignore": true, "parse": true}}`, ts.URL+path))
		wh := hook.NewWebHook(&whDeps, conf)

		err := wh.ExecuteLoginPreHook(nil, req, f)
		assert.Error(t, err)
	})

	for _, tc := range []struct {
		uc    string
		parse bool
	}{
		{uc: "Post Settings Hook - parse true", parse: true},
		{uc: "Post Settings Hook - parse false", parse: false},
	} {
		tc := tc
		t.Run("uc="+tc.uc, func(t *testing.T) {
			t.Parallel()
			ts := newServer(webHookHttpCodeWithBodyEndPoint(t, 200, []byte(`{"identity":{"traits":{"email":"some@other-example.org"}}}`)))
			req := &http.Request{
				Header: map[string][]string{"Some-Header": {"Some-Value"}},
				Host:   "www.ory.sh",
				TLS:    new(tls.ConnectionState),
				URL:    &url.URL{Path: "/some_end_point"},

				Method: http.MethodPost,
			}
			f := &settings.Flow{ID: x.NewUUID()}
			conf := json.RawMessage(fmt.Sprintf(`{"url": "%s", "method": "POST", "body": "%s", "response": {"parse":%t}}`, ts.URL+path, "file://./stub/test_body.jsonnet", tc.parse))
			wh := hook.NewWebHook(&whDeps, conf)
			uuid := x.NewUUID()
			in := &identity.Identity{ID: uuid}
			s := &session.Session{ID: x.NewUUID(), Identity: in}

			postPersistErr := wh.ExecuteSettingsPostPersistHook(nil, req, f, in, s)
			assert.NoError(t, postPersistErr)
			assert.Equal(t, in, &identity.Identity{ID: uuid})

			prePersistErr := wh.ExecuteSettingsPrePersistHook(nil, req, f, in)
			assert.NoError(t, prePersistErr)
			if tc.parse == true {
				assert.Equal(t, in, &identity.Identity{ID: uuid, Traits: identity.Traits(`{"email":"some@other-example.org"}`)})
			} else {
				assert.Equal(t, in, &identity.Identity{ID: uuid})
			}
		})
	}

	t.Run("must error when template is erroneous", func(t *testing.T) {
		t.Parallel()
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

	t.Run("must not error when template is erroneous and responses are ignored", func(t *testing.T) {
		ts := newServer(webHookHttpCodeEndPoint(200))
		req := &http.Request{
			Header: map[string][]string{"Some-Header": {"Some-Value"}},
			Host:   "www.ory.sh",
			TLS:    new(tls.ConnectionState),
			URL:    &url.URL{Path: "/some_end_point"},
			Method: http.MethodPost,
		}
		f := &login.Flow{ID: x.NewUUID()}
		conf := json.RawMessage(fmt.Sprintf(`{"url": "%s", "method": "GET", "body": "file://./stub/bad_template.jsonnet", "response": {"ignore": true}}`, ts.URL+path))
		wh := hook.NewWebHook(&whDeps, conf)

		err := wh.ExecuteLoginPreHook(nil, req, f)
		assert.NoError(t, err)
	})

	t.Run("must not make request", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		assert.Less(t, time.Since(start), waitTime)

		wg.Wait()
	})

	t.Run("does not error on 500 request with retry", func(t *testing.T) {
		t.Parallel()
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
		tc := tc
		t.Run("Must"+boolToString(tc.mustSuccess)+" error when end point is returning "+strconv.Itoa(tc.code), func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyClientHTTPNoPrivateIPRanges, true)
	conf.MustSet(ctx, config.ViperKeyClientHTTPPrivateIPExceptionURLs, []string{"http://localhost/exception"})
	logger := logrusx.New("kratos", "test")
	whDeps := struct {
		x.SimpleLoggerWithClient
		*jsonnetsecure.TestProvider
		config.Provider
	}{
		x.SimpleLoggerWithClient{L: logger, C: reg.HTTPClient(context.Background()), T: otelx.NewNoop(logger, &otelx.Config{ServiceName: "kratos"})},
		jsonnetsecure.NewTestProvider(t),
		reg,
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
		t.Parallel()
		wh := hook.NewWebHook(&whDeps, json.RawMessage(`{
  "url": "https://localhost:1234/",
  "method": "GET",
  "body": "file://stub/test_body.jsonnet"
}`))
		err := wh.ExecuteLoginPostHook(nil, req, node.DefaultGroup, f, s)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not a permitted destination")
	})

	t.Run("allowed to call exempt url", func(t *testing.T) {
		t.Parallel()
		wh := hook.NewWebHook(&whDeps, json.RawMessage(`{
  "url": "http://localhost/exception",
  "method": "GET",
  "body": "file://stub/test_body.jsonnet"
}`))
		err := wh.ExecuteLoginPostHook(nil, req, node.DefaultGroup, f, s)
		require.Error(t, err, "the target does not exist and we still receive an error")
		require.NotContains(t, err.Error(), "is not a permitted destination", "but the error is not related to the IP range.")
	})

	t.Run("not allowed to load from source", func(t *testing.T) {
		t.Parallel()
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
		require.Contains(t, err.Error(), "is not a permitted destination")
	})
}

func TestAsyncWebhook(t *testing.T) {
	t.Parallel()
	_, reg := internal.NewFastRegistryWithMocks(t)
	logger := logrusx.New("kratos", "test")
	logHook := new(test.Hook)
	logger.Logger.Hooks.Add(logHook)
	whDeps := struct {
		x.SimpleLoggerWithClient
		*jsonnetsecure.TestProvider
		config.Provider
	}{
		x.SimpleLoggerWithClient{L: logger, C: reg.HTTPClient(context.Background()), T: otelx.NewNoop(logger, &otelx.Config{ServiceName: "kratos"})},
		jsonnetsecure.NewTestProvider(t),
		reg,
	}

	req := &http.Request{
		Header: map[string][]string{"Some-Header": {"Some-Value"}},
		Host:   "www.ory.sh",
		TLS:    new(tls.ConnectionState),
		URL:    &url.URL{Path: "/some_end_point"},
		Method: http.MethodPost,
	}

	incomingCtx, incomingCancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		// cancel this context one second before test timeout for clean shutdown
		var cleanup context.CancelFunc
		incomingCtx, cleanup = context.WithDeadline(incomingCtx, deadline.Add(-time.Second))
		defer cleanup()
	}

	req = req.WithContext(incomingCtx)
	s := &session.Session{ID: x.NewUUID(), Identity: &identity.Identity{ID: x.NewUUID()}}
	f := &login.Flow{ID: x.NewUUID()}

	handlerEntered, blockHandlerOnExit := make(chan struct{}), make(chan struct{})
	webhookReceiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(handlerEntered)
		<-blockHandlerOnExit
		w.Write([]byte("ok"))
	}))
	t.Cleanup(webhookReceiver.Close)

	wh := hook.NewWebHook(&whDeps, json.RawMessage(fmt.Sprintf(`
		{
			"url": %q,
			"method": "GET",
			"body": "file://stub/test_body.jsonnet",
			"response": {
				"ignore": true
			}
		}`, webhookReceiver.URL)))
	err := wh.ExecuteLoginPostHook(nil, req, node.DefaultGroup, f, s)
	require.NoError(t, err) // execution returns immediately for async webhook
	select {
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for webhook request to reach test handler")
	case <-handlerEntered:
		// ok
	}
	// at this point, a goroutine is in the middle of the call to our test handler and waiting for a response
	incomingCancel() // simulate the incoming Kratos request having finished
	close(blockHandlerOnExit)
	timeout := time.After(1 * time.Second)
	var found bool
	for !found {
		for _, entry := range logHook.AllEntries() {
			if entry.Message == "Webhook request succeeded" {
				found = true
				break
			}
		}

		select {
		case <-timeout:
			t.Fatal("timed out waiting for successful webhook completion")
		case <-time.After(50 * time.Millisecond):
			// continue loop
		}
	}
	require.True(t, found)
}

func TestWebhookEvents(t *testing.T) {
	t.Parallel()
	_, reg := internal.NewFastRegistryWithMocks(t)
	logger := logrusx.New("kratos", "test")
	whDeps := struct {
		x.SimpleLoggerWithClient
		*jsonnetsecure.TestProvider
		config.Provider
	}{
		x.SimpleLoggerWithClient{L: logger, C: reg.HTTPClient(context.Background()), T: otelx.NewNoop(logger, &otelx.Config{ServiceName: "kratos"})},
		jsonnetsecure.NewTestProvider(t),
		reg,
	}

	req := &http.Request{
		Header: map[string][]string{"Some-Header": {"Some-Value"}},
		Host:   "www.ory.sh",
		TLS:    new(tls.ConnectionState),
		URL:    &url.URL{Path: "/some_end_point"},
		Method: http.MethodPost,
	}
	f := &login.Flow{ID: x.NewUUID()}

	webhookReceiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ok" {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		} else {
			w.WriteHeader(500)
			w.Write([]byte("fail"))
		}
	}))
	t.Cleanup(webhookReceiver.Close)

	getAttributes := func(attrs []attribute.KeyValue) (webhookID, triggerID, requestID string) {
		for _, kv := range attrs {
			switch semconv.AttributeKey(kv.Key) {
			case events.AttributeKeyWebhookID:
				webhookID = kv.Value.Emit()
			case events.AttributeKeyWebhookTriggerID:
				triggerID = kv.Value.Emit()
			case events.AttributeKeyWebhookRequestID:
				requestID = kv.Value.Emit()
			}
		}
		return
	}

	t.Run("success", func(t *testing.T) {
		whID := x.NewUUID()
		wh := hook.NewWebHook(&whDeps, json.RawMessage(fmt.Sprintf(`
		{
			"id": %q,
			"url": %q,
			"method": "GET",
			"body": "file://stub/test_body.jsonnet",
			"response": {
				"ignore": false,
				"parse": false
			}
		}`, whID, webhookReceiver.URL+"/ok")))

		recorder := tracetest.NewSpanRecorder()
		tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)).Tracer("test")
		ctx, span := tracer.Start(context.Background(), "parent")
		defer span.End()

		r1 := req.Clone(ctx)

		require.NoError(t, wh.ExecuteLoginPreHook(nil, r1, f))

		ended := recorder.Ended()
		require.NotEmpty(t, ended)

		i := slices.IndexFunc(ended, func(sp sdktrace.ReadOnlySpan) bool { return sp.Name() == "selfservice.webhook" })
		require.GreaterOrEqual(t, i, 0)

		evs := ended[i].Events()
		i = slices.IndexFunc(evs, func(ev sdktrace.Event) bool { return ev.Name == events.WebhookDelivered.String() })
		require.GreaterOrEqual(t, i, 0)

		actualWhID, deliveredTriggerID, deliveredRequestID := getAttributes(evs[i].Attributes)
		require.Equal(t, whID.String(), actualWhID)
		require.NotEmpty(t, deliveredTriggerID)
		require.NotEmpty(t, deliveredRequestID)
		assert.NotEqual(t, deliveredTriggerID, deliveredRequestID)

		i = slices.IndexFunc(evs, func(ev sdktrace.Event) bool { return ev.Name == events.WebhookSucceeded.String() })
		require.GreaterOrEqual(t, i, 0)

		actualWhID, succeededTriggerID, _ := getAttributes(evs[i].Attributes)
		require.Equal(t, whID.String(), actualWhID)
		require.NotEmpty(t, succeededTriggerID)

		assert.Equal(t, deliveredTriggerID, succeededTriggerID)

		i = slices.IndexFunc(evs, func(ev sdktrace.Event) bool { return ev.Name == events.WebhookFailed.String() })
		require.Equal(t, -1, i)
	})

	t.Run("failed", func(t *testing.T) {
		whID := x.NewUUID()
		wh := hook.NewWebHook(&whDeps, json.RawMessage(fmt.Sprintf(`
		{
			"id": %q,
			"url": %q,
			"method": "GET",
			"body": "file://stub/test_body.jsonnet",
			"response": {
				"ignore": false,
				"parse": false
			}
		}`, whID, webhookReceiver.URL+"/fail")))

		recorder := tracetest.NewSpanRecorder()
		tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)).Tracer("test")
		ctx, span := tracer.Start(context.Background(), "parent")
		defer span.End()

		r1 := req.Clone(ctx)
		require.Error(t, wh.ExecuteLoginPreHook(nil, r1, f))

		ended := recorder.Ended()
		require.NotEmpty(t, ended)

		i := slices.IndexFunc(ended, func(sp sdktrace.ReadOnlySpan) bool { return sp.Name() == "selfservice.webhook" })
		require.GreaterOrEqual(t, i, 0)

		evs := ended[i].Events()

		var deliveredEvents []sdktrace.Event
		deliveredTriggerIDs := map[string]struct{}{}
		deliveredRequestIDs := map[string]struct{}{}
		for _, ev := range evs {
			if ev.Name == events.WebhookDelivered.String() {
				deliveredEvents = append(deliveredEvents, ev)
				actualWhID, triggerID, requestID := getAttributes(ev.Attributes)
				require.Equal(t, whID.String(), actualWhID)
				require.NotEmpty(t, triggerID)
				require.NotEmpty(t, requestID)
				deliveredTriggerIDs[triggerID] = struct{}{}
				deliveredRequestIDs[requestID] = struct{}{}
			}
		}

		assert.Len(t, deliveredEvents, 3)
		assert.Len(t, deliveredTriggerIDs, 1)
		assert.Len(t, deliveredRequestIDs, 3)

		i = slices.IndexFunc(evs, func(ev sdktrace.Event) bool { return ev.Name == "WebhookFailed" })
		require.GreaterOrEqual(t, i, 0)

		actualWhID, failedTriggerID, _ := getAttributes(evs[i].Attributes)
		require.Equal(t, whID.String(), actualWhID)
		require.NotEmpty(t, failedTriggerID)

		assert.Contains(t, deliveredTriggerIDs, failedTriggerID)

		i = slices.IndexFunc(evs, func(ev sdktrace.Event) bool { return ev.Name == "WebhookSucceeded" })
		require.Equal(t, i, -1)
	})

	t.Run("event disabled", func(t *testing.T) {
		wh := hook.NewWebHook(&whDeps, json.RawMessage(fmt.Sprintf(`
		{
			"url": %q,
			"method": "GET",
			"body": "file://stub/test_body.jsonnet",
			"response": {
				"ignore": false,
				"parse": false
			},
			"emit_analytics_event": false
		}`, webhookReceiver.URL+"/fail")))

		recorder := tracetest.NewSpanRecorder()
		tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)).Tracer("test")
		ctx, span := tracer.Start(context.Background(), "parent")
		defer span.End()

		r1 := req.Clone(ctx)
		require.Error(t, wh.ExecuteLoginPreHook(nil, r1, f))

		ended := recorder.Ended()
		require.NotEmpty(t, ended)

		i := slices.IndexFunc(ended, func(sp sdktrace.ReadOnlySpan) bool {
			return sp.Name() == "selfservice.webhook"
		})
		require.GreaterOrEqual(t, i, 0)

		evs := ended[i].Events()
		i = slices.IndexFunc(evs, func(ev sdktrace.Event) bool {
			return ev.Name == "WebhookDelivered"
		})
		require.Equal(t, -1, i)

		i = slices.IndexFunc(evs, func(ev sdktrace.Event) bool {
			return ev.Name == "WebhookFailed"
		})
		require.Equal(t, -1, i)

		i = slices.IndexFunc(evs, func(ev sdktrace.Event) bool {
			return ev.Name == "WebhookSucceeded"
		})
		require.Equal(t, i, -1)
	})
}
