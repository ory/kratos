package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/request"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
)

var _ registration.PostHookPostPersistExecutor = new(WebHook)
var _ verification.PostHookExecutor = new(WebHook)
var _ recovery.PostHookExecutor = new(WebHook)

type (
	webHookDependencies interface {
		x.LoggingProvider
		x.HTTPClientProvider
		x.TracingProvider
	}

	templateContext struct {
		Flow           flow.Flow          `json:"flow"`
		RequestHeaders http.Header        `json:"request_headers"`
		RequestMethod  string             `json:"request_method"`
		RequestUrl     string             `json:"request_url"`
		Identity       *identity.Identity `json:"identity"`
	}

	WebHook struct {
		deps webHookDependencies
		conf json.RawMessage
	}
)

func NewWebHook(r webHookDependencies, c json.RawMessage) *WebHook {
	return &WebHook{deps: r, conf: c}
}

func (e *WebHook) ExecuteLoginPreHook(_ http.ResponseWriter, req *http.Request, flow *login.Flow) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePreLoginHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
	})
}

func (e *WebHook) ExecuteLoginPostHook(_ http.ResponseWriter, req *http.Request, flow *login.Flow, session *session.Session) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePostLoginHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, req *http.Request, flow *verification.Flow, identity *identity.Identity) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePostVerificationHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       identity,
	})
}

func (e *WebHook) ExecutePostRecoveryHook(_ http.ResponseWriter, req *http.Request, flow *recovery.Flow, session *session.Session) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePostRecoveryHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecuteRegistrationPreHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecuteRegistrationPreHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
	})
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow, session *session.Session) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecutePostRegistrationPostPersistHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecuteSettingsPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *settings.Flow, identity *identity.Identity) error {
	ctx, _ := e.deps.Tracer(req.Context()).Tracer().Start(req.Context(), "selfservice.hook.ExecuteSettingsPostPersistHook")
	return e.execute(ctx, &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       identity,
	})
}

func (e *WebHook) execute(ctx context.Context, data *templateContext) error {
	span := trace.SpanFromContext(ctx)
	attrs := map[string]string{
		"webhook.http.method":  data.RequestMethod,
		"webhook.http.url":     data.RequestUrl,
		"webhook.http.headers": fmt.Sprintf("%#v", data.RequestHeaders),
		"webhook.identity":     fmt.Sprintf("%#v", data.Identity),
	}
	span.SetAttributes(otelx.StringAttrs(attrs)...)
	defer span.End()

	builder, err := request.NewBuilder(e.conf, e.deps.HTTPClient(ctx), e.deps.Logger())
	if err != nil {
		return err
	}

	req, err := builder.BuildRequest(data)
	if errors.Is(err, request.ErrCancel) {
		return nil
	} else if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)

		resp, err := e.deps.HTTPClient(ctx).Do(req)
		if err != nil {
			errChan <- err
			return
		}

		if resp.StatusCode >= http.StatusBadRequest {
			errChan <- fmt.Errorf("web hook failed with status code %v", resp.StatusCode)
			span.SetStatus(codes.Error, fmt.Sprintf("web hook failed with status code %v", resp.StatusCode))
			return
		}

		errChan <- nil
	}()

	if gjson.GetBytes(e.conf, "response.ignore").Bool() {
		go func() {
			err := <-errChan
			e.deps.Logger().WithError(err).Warning("A web hook request failed but the error was ignored because the configuration indicated that the upstream response should be ignored.")
		}()
		return nil
	}

	return <-errChan
}
