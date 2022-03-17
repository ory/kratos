package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/ui/node"

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

var _ registration.PostHookPostPersistExecutor = new(WebHook)
var _ registration.PostHookPrePersistExecutor = new(WebHook)
var _ verification.PostHookExecutor = new(WebHook)
var _ recovery.PostHookExecutor = new(WebHook)
var _ settings.PostHookPostPersistExecutor = new(WebHook)

type (
	webHookDependencies interface {
		x.LoggingProvider
		x.HTTPClientProvider
	}

	templateContext struct {
		Flow           flow.Flow          `json:"flow"`
		RequestHeaders http.Header        `json:"request_headers"`
		RequestMethod  string             `json:"request_method"`
		RequestUrl     string             `json:"request_url"`
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
		Message          string            `json:"message"`
		DetailedMessages []detailedMessage `json:"detailed_messages"`
	}

	rawHookResponse struct {
		Messages []errorMessage `json:"messages"`
	}
)

func NewWebHook(r webHookDependencies, c json.RawMessage) *WebHook {
	return &WebHook{deps: r, conf: c}
}

func (e *WebHook) ExecuteLoginPreHook(_ http.ResponseWriter, req *http.Request, flow *login.Flow) error {
	return e.execute(req.Context(), &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
	})
}

func (e *WebHook) ExecuteLoginPostHook(_ http.ResponseWriter, req *http.Request, _ node.Group, flow *login.Flow, session *session.Session) error {
	return e.execute(req.Context(), &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, req *http.Request, flow *verification.Flow, id *identity.Identity) error {
	return e.execute(req.Context(), &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       id,
	})
}

func (e *WebHook) ExecutePostRecoveryHook(_ http.ResponseWriter, req *http.Request, flow *recovery.Flow, session *session.Session) error {
	return e.execute(req.Context(), &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecuteRegistrationPreHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow) error {
	return e.execute(req.Context(), &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
	})
}

func (e *WebHook) ExecutePostRegistrationPrePersistHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow, id *identity.Identity) error {
	return e.execute(req.Context(), &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       id,
	})
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow, session *session.Session) error {
	return e.execute(req.Context(), &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecuteSettingsPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *settings.Flow, id *identity.Identity) error {
	return e.execute(req.Context(), &templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       id,
	})
}

func (e *WebHook) execute(ctx context.Context, data *templateContext) error {
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
			if gjson.GetBytes(e.conf, "can_interrupt").Bool() {
				if err := parseWebhookResponse(resp); err != nil {
					errChan <- err
				}
			}
			errChan <- fmt.Errorf("web hook failed with status code %v", resp.StatusCode)
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

func parseWebhookResponse(resp *http.Response) (err error) {
	if resp == nil {
		return errors.Errorf("empty response provided from the webhook")
	}
	var hookResponse rawHookResponse
	if err := json.NewDecoder(resp.Body).Decode(&hookResponse); err != nil {
		return errors.Wrap(err, "hook response could not be unmarshalled properly from JSON")
	}

	validationErr := schema.NewValidationListError()
	for _, msg := range hookResponse.Messages {
		messages := text.Messages{}
		for _, detail := range msg.DetailedMessages {
			var msgType text.Type
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
		validationErr.Add(schema.NewHookValidationError(msg.InstancePtr, msg.Message, messages))
	}

	if !validationErr.HasErrors() {
		return errors.New("error while parsing hook response: got no validation errors")
	}

	return errors.WithStack(validationErr)
}
