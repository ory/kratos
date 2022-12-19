// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.11.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/jsonnetsecure"

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
	"github.com/ory/kratos/x"
)

var (
	_ registration.PostHookPostPersistExecutor = new(WebHook)
	_ registration.PostHookPrePersistExecutor  = new(WebHook)

	_ verification.PostHookExecutor = new(WebHook)

	_ recovery.PostHookExecutor = new(WebHook)

	_ settings.PostHookPostPersistExecutor = new(WebHook)
	_ settings.PostHookPrePersistExecutor  = new(WebHook)
)

type (
	webHookDependencies interface {
		x.LoggingProvider
		x.HTTPClientProvider
		x.TracingProvider
		jsonnetsecure.VMProvider
	}

	templateContext struct {
		Flow           flow.Flow          `json:"flow"`
		RequestHeaders http.Header        `json:"request_headers"`
		RequestMethod  string             `json:"request_method"`
		RequestURL     string             `json:"request_url"`
		RequestCookies map[string]string  `json:"request_cookies"`
		Identity       *identity.Identity `json:"identity,omitempty"`
	}

	WebHook struct {
		deps webHookDependencies
		conf json.RawMessage
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

func NewWebHook(r webHookDependencies, c json.RawMessage) *WebHook {
	return &WebHook{deps: r, conf: c}
}

func (e *WebHook) ExecuteLoginPreHook(_ http.ResponseWriter, req *http.Request, flow *login.Flow) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePreLoginHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
	})
}

func (e *WebHook) ExecuteLoginPostHook(_ http.ResponseWriter, req *http.Request, _ node.UiNodeGroup, flow *login.Flow, session *session.Session) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePostLoginHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecuteVerificationPreHook(_ http.ResponseWriter, req *http.Request, flow *verification.Flow) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePreVerificationHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
	})
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, req *http.Request, flow *verification.Flow, id *identity.Identity) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePostVerificationHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
		Identity:       id,
	})
}

func (e *WebHook) ExecuteRecoveryPreHook(_ http.ResponseWriter, req *http.Request, flow *recovery.Flow) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePreRecoveryHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestCookies: cookies(req),
		RequestURL:     x.RequestURL(req).String(),
	})
}

func (e *WebHook) ExecutePostRecoveryHook(_ http.ResponseWriter, req *http.Request, flow *recovery.Flow, session *session.Session) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePostRecoveryHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecuteRegistrationPreHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecuteRegistrationPreHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
	})
}

func (e *WebHook) ExecutePostRegistrationPrePersistHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow, id *identity.Identity) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePostRegistrationPrePersistHook")
	if !gjson.GetBytes(e.conf, "can_interrupt").Bool() {
		return nil
	}

	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
		Identity:       id,
	})
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow, session *session.Session) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePostRegistrationPostPersistHook")
	if gjson.GetBytes(e.conf, "can_interrupt").Bool() {
		return nil
	}

	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecuteSettingsPreHook(_ http.ResponseWriter, req *http.Request, flow *settings.Flow) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePreSettingsHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
	})
}

func (e *WebHook) ExecuteSettingsPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *settings.Flow, id *identity.Identity) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecuteSettingsPostPersistHook")
	if gjson.GetBytes(e.conf, "can_interrupt").Bool() {
		return nil
	}

	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
		Identity:       id,
	})
}

func (e *WebHook) ExecuteSettingsPrePersistHook(_ http.ResponseWriter, req *http.Request, flow *settings.Flow, id *identity.Identity) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecuteSettingsPrePersistHook")
	if !gjson.GetBytes(e.conf, "can_interrupt").Bool() {
		return nil
	}

	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
		Identity:       id,
	})
}

func (e *WebHook) execute(ctx context.Context, data *templateContext) error {
	builder, err := request.NewBuilder(e.conf, e.deps)
	if err != nil {
		return err
	}

	req, err := builder.BuildRequest(ctx, data)
	if errors.Is(err, request.ErrCancel) {
		return nil
	} else if err != nil {
		return err
	}

	attrs := semconv.HTTPClientAttributesFromHTTPRequest(req.Request)
	if data.Identity != nil {
		attrs = append(attrs,
			attribute.String("webhook.identity.id", data.Identity.ID.String()),
			attribute.String("webhook.identity.nid", data.Identity.NID.String()),
		)
	}
	var (
		httpClient                       = e.deps.HTTPClient(ctx)
		async                            = gjson.GetBytes(e.conf, "response.ignore").Bool()
		parseResponse                    = gjson.GetBytes(e.conf, "can_interrupt").Bool()
		tracer                           = trace.SpanFromContext(ctx).TracerProvider().Tracer("kratos-webhooks")
		cancel        context.CancelFunc = func() {}
		spanOpts                         = []trace.SpanStartOption{trace.WithAttributes(attrs...)}
		errChan                          = make(chan error, 1)
	)
	if async {
		// dissociate the context from the one passed into this function
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
		spanOpts = append(spanOpts, trace.WithNewRoot())
	}
	ctx, span := tracer.Start(ctx, "Webhook", spanOpts...)
	e.deps.Logger().WithRequest(req.Request).Info("Dispatching webhook")
	t0 := time.Now()
	go func() {
		defer close(errChan)
		defer cancel()
		defer span.End()

		resp, err := httpClient.Do(req.WithContext(ctx))
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			errChan <- errors.WithStack(err)
			return
		}
		defer resp.Body.Close()
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)...)

		if resp.StatusCode >= http.StatusBadRequest {
			span.SetStatus(codes.Error, "HTTP status code >= 400")
			if parseResponse {
				if err := parseWebhookResponse(resp); err != nil {
					span.SetStatus(codes.Error, err.Error())
					errChan <- err
				}
			}
			errChan <- fmt.Errorf("webhook failed with status code %v", resp.StatusCode)
			return
		}

		errChan <- nil
	}()

	if async {
		traceID, spanID := span.SpanContext().TraceID(), span.SpanContext().SpanID()
		go func() {
			if err := <-errChan; err != nil {
				e.deps.Logger().WithField("otel", map[string]string{
					"trace_id": traceID.String(),
					"span_id":  spanID.String(),
				}).WithError(err).Warning("Webhook request failed but the error was ignored because the configuration indicated that the upstream response should be ignored.")
			} else {
				e.deps.Logger().WithField("duration", time.Since(t0)).Info("Webhook request succeeded")
			}
		}()
		return nil
	}

	return <-errChan
}

func parseWebhookResponse(resp *http.Response) (err error) {
	if resp == nil {
		return errors.Errorf("empty response provided from the webhook")
	}
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
