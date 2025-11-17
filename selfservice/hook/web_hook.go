// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.11.0"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/request"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/events"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/otelx"
)

var _ interface {
	login.PreHookExecutor
	login.PostHookExecutor

	registration.PostHookPostPersistExecutor
	registration.PostHookPrePersistExecutor
	registration.PreHookExecutor

	verification.PreHookExecutor
	verification.PostHookExecutor

	recovery.PreHookExecutor
	recovery.PostHookExecutor

	settings.PreHookExecutor
	settings.PostHookPrePersistExecutor
	settings.PostHookPostPersistExecutor
} = (*WebHook)(nil)

var jsonnetCache, _ = ristretto.NewCache(&ristretto.Config[[]byte, []byte]{
	MaxCost:     100 << 20, // 100MB,
	NumCounters: 1_000_000, // 1kB per snippet -> 100k snippets -> 1M counters
	BufferItems: 64,
})

type (
	webHookDependencies interface {
		x.LoggingProvider
		x.HTTPClientProvider
		x.TracingProvider
		jsonnetsecure.VMProvider
		config.Provider
	}

	templateContext struct {
		Flow           flow.Flow          `json:"flow"`
		RequestHeaders http.Header        `json:"request_headers"`
		RequestMethod  string             `json:"request_method"`
		RequestURL     string             `json:"request_url"`
		RequestCookies map[string]string  `json:"request_cookies"`
		Identity       *identity.Identity `json:"identity,omitempty"`
		Session        *session.Session   `json:"session,omitempty"`
	}

	WebHook struct {
		deps webHookDependencies
		conf *request.Config
	}

	detailedMessage struct {
		ID      int             `json:"id"`
		Text    string          `json:"text"`
		Type    string          `json:"type"`
		Context json.RawMessage `json:"context,omitempty"`
	}

	errorMessage struct {
		InstancePtr      string            `json:"instance_ptr"`
		DetailedMessages []detailedMessage `json:"messages"`
	}

	rawHookResponse struct {
		Messages []errorMessage `json:"messages"`
	}
)

func cookies(req *http.Request) map[string]string {
	cookies := make(map[string]string)
	for _, c := range req.Cookies() {
		if c.Name != "" {
			cookies[c.Name] = c.Value
		}
	}
	return cookies
}

func NewWebHook(r webHookDependencies, c *request.Config) *WebHook {
	return &WebHook{deps: r, conf: c}
}

func (e *WebHook) ExecuteLoginPreHook(_ http.ResponseWriter, req *http.Request, flow *login.Flow) error {
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecuteLoginPreHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
		})
	})
}

func (e *WebHook) ExecuteLoginPostHook(_ http.ResponseWriter, req *http.Request, _ node.UiNodeGroup, flow *login.Flow, session *session.Session) error {
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecuteLoginPostHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
			Identity:       session.Identity,
			Session:        session,
		})
	})
}

func (e *WebHook) ExecuteVerificationPreHook(_ http.ResponseWriter, req *http.Request, flow *verification.Flow) error {
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecuteVerificationPreHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
		})
	})
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, req *http.Request, flow *verification.Flow, id *identity.Identity) error {
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecutePostVerificationHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
			Identity:       id,
		})
	})
}

func (e *WebHook) ExecuteRecoveryPreHook(_ http.ResponseWriter, req *http.Request, flow *recovery.Flow) error {
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecuteRecoveryPreHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestCookies: cookies(req),
			RequestURL:     x.RequestURL(req).String(),
		})
	})
}

func (e *WebHook) ExecutePostRecoveryHook(_ http.ResponseWriter, req *http.Request, flow *recovery.Flow, session *session.Session) error {
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecutePostRecoveryHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
			Identity:       session.Identity,
		})
	})
}

func (e *WebHook) ExecuteRegistrationPreHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow) error {
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecuteRegistrationPreHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
		})
	})
}

func (e *WebHook) ExecutePostRegistrationPrePersistHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow, id *identity.Identity) error {
	if !e.conf.CanInterrupt && !e.conf.Response.Parse {
		return nil
	}

	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecutePostRegistrationPrePersistHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
			Identity:       id,
		})
	})
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow, session *session.Session) error {
	if e.conf.CanInterrupt || e.conf.Response.Parse {
		return nil
	}

	// We want to decouple the request from the hook execution, so that the hooks still execute even
	// if the request is canceled.
	ctx := context.WithoutCancel(req.Context())

	return otelx.WithSpan(ctx, "selfservice.hook.WebHook.ExecutePostRegistrationPostPersistHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
			Identity:       session.Identity,
		})
	})
}

func (e *WebHook) ExecuteSettingsPreHook(_ http.ResponseWriter, req *http.Request, flow *settings.Flow, s *session.Session) error {
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecuteSettingsPreHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
			Session:        s,
		})
	})
}

func (e *WebHook) ExecuteSettingsPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *settings.Flow, id *identity.Identity, s *session.Session) error {
	if e.conf.CanInterrupt || e.conf.Response.Parse {
		return nil
	}
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecuteSettingsPostPersistHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
			Identity:       id,
			Session:        s,
		})
	})
}

func (e *WebHook) ExecuteSettingsPrePersistHook(_ http.ResponseWriter, req *http.Request, flow *settings.Flow, id *identity.Identity, s *session.Session) error {
	if !e.conf.CanInterrupt && !e.conf.Response.Parse {
		return nil
	}
	return otelx.WithSpan(req.Context(), "selfservice.hook.WebHook.ExecuteSettingsPrePersistHook", func(ctx context.Context) error {
		return e.execute(ctx, &templateContext{
			Flow:           flow,
			RequestHeaders: req.Header,
			RequestMethod:  req.Method,
			RequestURL:     x.RequestURL(req).String(),
			RequestCookies: cookies(req),
			Identity:       id,
			Session:        s,
		})
	})
}

func (e *WebHook) execute(ctx context.Context, data *templateContext) error {
	var (
		httpClient     = e.deps.HTTPClient(ctx)
		ignoreResponse = e.conf.Response.Ignore
		canInterrupt   = e.conf.CanInterrupt
		parseResponse  = e.conf.Response.Parse
		emitEvent      = e.conf.EmitAnalyticsEvent == nil || *e.conf.EmitAnalyticsEvent // default true
		webhookID      = e.conf.ID
		// The trigger ID is a random ID. It can be used to correlate webhook requests across retries.
		triggerID = x.NewUUID()
		tracer    = trace.SpanFromContext(ctx).TracerProvider().Tracer("kratos-webhooks")
	)
	if ignoreResponse && (parseResponse || canInterrupt) {
		return errors.WithStack(herodot.ErrMisconfiguration.WithReasonf("A webhook is configured to ignore the response but also to parse the response. This is not possible."))
	}

	makeRequest := func() (finalErr error) {
		if ignoreResponse {
			// This means we want to run this closure asynchronously and not be
			// canceled when the parent context is canceled.
			//
			// The webhook will still cancel after 30 seconds as that is the
			// configured timeout for the HTTP client.
			ctx = context.WithoutCancel(ctx)
		}
		ctx, span := tracer.Start(ctx, "selfservice.webhook")
		defer otelx.End(span, &finalErr)

		if emitEvent {
			InstrumentHTTPClientForEvents(ctx, httpClient, triggerID, webhookID)
		}

		defer func(startTime time.Time) {
			traceID, spanID := span.SpanContext().TraceID(), span.SpanContext().SpanID()
			logger := e.deps.Logger().WithField("otel", map[string]string{
				"trace_id": traceID.String(),
				"span_id":  spanID.String(),
			}).WithField("duration", time.Since(startTime))
			if finalErr != nil {
				if emitEvent && !errors.Is(finalErr, context.Canceled) {
					span.AddEvent(events.NewWebhookFailed(ctx, finalErr, triggerID, webhookID))
				}
				if ignoreResponse {
					logger.WithError(finalErr).Warning("Webhook request failed but the error was ignored because the configuration indicated that the upstream response should be ignored")
				} else {
					logger.WithError(finalErr).Error("Webhook request failed")
				}
			} else {
				logger.Info("Webhook request succeeded")
				if emitEvent {
					span.AddEvent(events.NewWebhookSucceeded(ctx, triggerID, webhookID))
				}
			}
		}(time.Now())

		builder, err := request.NewBuilder(ctx, e.conf, e.deps, request.WithCache(jsonnetCache))
		if err != nil {
			return err
		}

		span.SetAttributes(
			attribute.String("webhook.jsonnet.template-uri", builder.Config.TemplateURI),
			attribute.Bool("webhook.can_interrupt", canInterrupt),
			attribute.Bool("webhook.response.ignore", ignoreResponse),
			attribute.Bool("webhook.response.parse", parseResponse),
		)

		data.RequestHeaders = RemoveDisallowedHeaders(data.RequestHeaders, e.deps.Config().WebhookHeaderAllowlist(ctx))

		req, err := builder.BuildRequest(ctx, data)
		if errors.Is(err, request.ErrCancel) {
			span.SetAttributes(attribute.Bool("webhook.jsonnet.canceled", true))
			return nil
		} else if err != nil {
			return err
		}

		if data.Identity != nil {
			span.SetAttributes(
				attribute.String("webhook.identity.id", data.Identity.ID.String()),
				attribute.String("webhook.identity.nid", data.Identity.NID.String()),
			)
		}

		e.deps.Logger().WithRequest(req.Request).Info("Dispatching webhook")

		req = req.WithContext(ctx)

		resp, err := httpClient.Do(req)
		if err != nil {
			if IsTimeoutError(err) {
				return herodot.DefaultError{
					CodeField:     http.StatusGatewayTimeout,
					StatusField:   http.StatusText(http.StatusGatewayTimeout),
					GRPCCodeField: grpccodes.DeadlineExceeded,
					ErrorField:    err.Error(),
					ReasonField:   "A third-party upstream service could not be reached. Please try again later.",
				}.WithWrap(errors.WithStack(err))
			}
			return errors.WithStack(err)
		}
		defer func() { _ = resp.Body.Close() }()
		resp.Body = io.NopCloser(io.LimitReader(resp.Body, 5<<20)) // read at most 5 MB from the response
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)...)

		if resp.StatusCode >= http.StatusBadRequest {
			span.SetStatus(codes.Error, "HTTP status code >= 400")
			if canInterrupt || parseResponse {
				if err := parseWebhookResponse(resp, data.Identity); err != nil {
					return err
				}
			}
			return herodot.DefaultError{
				CodeField:     http.StatusBadGateway,
				StatusField:   http.StatusText(http.StatusBadGateway),
				GRPCCodeField: grpccodes.Aborted,
				ReasonField:   "A third-party upstream service responded improperly. Please try again later.",
				ErrorField:    fmt.Sprintf("webhook failed with status code %v", resp.StatusCode),
			}
		}

		if parseResponse {
			return parseWebhookResponse(resp, data.Identity)
		}
		return nil
	}

	if !ignoreResponse {
		return makeRequest()
	}
	go func() {
		// we cannot handle the error as we are running async, and it is logged anyway
		_ = makeRequest()
	}()
	return nil
}

func RemoveDisallowedHeaders(httpHeaders http.Header, headerAllowlist []string) http.Header {
	res := make(http.Header, len(headerAllowlist))
	for _, allowed := range headerAllowlist {
		h, present := httpHeaders[textproto.CanonicalMIMEHeaderKey(allowed)]
		if present {
			res[allowed] = h
		}
	}
	return res
}

func parseWebhookResponse(resp *http.Response, id *identity.Identity) (err error) {
	if resp == nil {
		return errors.Errorf("empty response provided from the webhook")
	}

	if resp.StatusCode == http.StatusOK {
		type localIdentity identity.Identity
		var hookResponse struct {
			Identity *localIdentity `json:"identity"`
		}
		// io.ReadAll is safe, because resp.Body is already a limited reader.
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "webhook response body could not be read")
		}
		if err = json.Unmarshal(body, &hookResponse); err != nil {
			return errors.Wrap(err, "webhook response could not be unmarshalled properly from JSON")
		}

		if hookResponse.Identity == nil {
			return nil
		}

		if len(hookResponse.Identity.Traits) > 0 {
			id.Traits = hookResponse.Identity.Traits
		}

		if len(hookResponse.Identity.SchemaID) > 0 {
			id.SchemaID = hookResponse.Identity.SchemaID
		}

		if len(hookResponse.Identity.State) > 0 {
			id.State = hookResponse.Identity.State
		}

		if len(hookResponse.Identity.VerifiableAddresses) > 0 {
			id.VerifiableAddresses = hookResponse.Identity.VerifiableAddresses
		}

		if len(hookResponse.Identity.VerifiableAddresses) > 0 {
			id.VerifiableAddresses = hookResponse.Identity.VerifiableAddresses
		}

		if len(hookResponse.Identity.RecoveryAddresses) > 0 {
			id.RecoveryAddresses = hookResponse.Identity.RecoveryAddresses
		}

		if len(hookResponse.Identity.MetadataPublic) > 0 {
			id.MetadataPublic = hookResponse.Identity.MetadataPublic
		}

		if len(hookResponse.Identity.MetadataAdmin) > 0 {
			id.MetadataAdmin = hookResponse.Identity.MetadataAdmin
		}

		if gjson.GetBytes(body, "identity.external_id").Exists() {
			id.ExternalID = hookResponse.Identity.ExternalID
		}

		return nil
	} else if resp.StatusCode == http.StatusNoContent {
		return nil
	} else if resp.StatusCode >= http.StatusBadRequest {
		var hookResponse rawHookResponse
		if err := json.NewDecoder(resp.Body).Decode(&hookResponse); err != nil {
			return errors.Wrap(err, "webhook response could not be unmarshalled properly from JSON")
		}

		var validationErrs []*schema.ValidationError
		for _, msg := range hookResponse.Messages {
			messages := text.Messages{}
			for _, detail := range msg.DetailedMessages {
				var msgType text.UITextType
				if detail.Type == "error" {
					msgType = text.Error
				} else {
					msgType = text.Info
				}
				messages.Add(&text.Message{
					ID:      text.ID(detail.ID),
					Text:    detail.Text,
					Type:    msgType,
					Context: detail.Context,
				})
			}
			validationErrs = append(validationErrs, schema.NewHookValidationError(msg.InstancePtr, "a webhook target returned an error", messages))
		}

		if len(validationErrs) == 0 {
			return errors.New("error while parsing webhook response: got no validation errors")
		}

		return schema.NewValidationListError(validationErrs)
	}

	return nil
}

func IsTimeoutError(err error) bool {
	var te interface{ Timeout() bool }
	return errors.As(err, &te) && te.Timeout() || errors.Is(err, context.DeadlineExceeded)
}

func InstrumentHTTPClientForEvents(ctx context.Context, httpClient *retryablehttp.Client, triggerID uuid.UUID, webhookID string) {
	// TODO(@alnr): improve this implementation to redact sensitive data
	var (
		attempt   = 0
		requestID uuid.UUID
		reqBody   []byte
	)
	httpClient.RequestLogHook = func(_ retryablehttp.Logger, req *http.Request, retryNumber int) {
		attempt = retryNumber + 1
		requestID = x.NewUUID()
		req.Header.Set("Ory-Webhook-Request-ID", requestID.String())
		req.Header.Set("Ory-Webhook-Trigger-ID", triggerID.String())
		// TODO(@alnr): redact sensitive data
		// reqBody, _ = httputil.DumpRequestOut(req, true)
		reqBody = []byte("<redacted>")
	}
	httpClient.ResponseLogHook = func(_ retryablehttp.Logger, res *http.Response) {
		// res.Body = io.NopCloser(io.LimitReader(res.Body, 5<<20)) // read at most 5 MB from the response
		// resBody, _ := httputil.DumpResponse(res, true)
		// resBody = resBody[:min(len(resBody), 2<<10)] // truncate response body to 2 kB for event
		// TODO(@alnr): redact sensitive data
		resBody := []byte("<redacted>")
		trace.SpanFromContext(ctx).AddEvent(events.NewWebhookDelivered(ctx, res.Request.URL, reqBody, res.StatusCode, resBody, attempt, requestID, triggerID, webhookID))
	}
}
