package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"net/http"
)

var _ registration.PostHookPostPersistExecutor = new(WebHook)
var _ verification.PostHookExecutor = new(WebHook)

type (
	webHookDependencies interface {
		x.LoggingProvider
	}

	webHookConfig struct {
		method string
		url    string
	}

	WebHook struct {
		r webHookDependencies
		c json.RawMessage
	}
)

func NewWebHook(r webHookDependencies, c json.RawMessage) *WebHook {
	return &WebHook{r: r, c: c}
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, _ *http.Request, _ *verification.Flow, s *session.Session) error {
	return e.executeWebHook(s)
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, _ *http.Request, _ *registration.Flow, s *session.Session) error {
	return e.executeWebHook(s)
}

func (e *WebHook) executeWebHook(s *session.Session) error {
	var conf webHookConfig
	if err := json.Unmarshal(e.c, &conf); err != nil {
		e.r.Logger().WithError(err).Error("Error while unmarshalling web_hook config json")
		return err
	}

	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(s.Identity); err != nil {
		e.r.Logger().WithError(err).Error("Error while marshalling web_hook payload json")
		return err
	}

	if response, err := http.Post(conf.url, "application/json", payloadBuf); err != nil {
		e.r.Logger().WithError(err).Warn("Web hook failed")
		return err
	} else if response.StatusCode >= 400 {
		return fmt.Errorf("web hook failed with status code %v", response.StatusCode)
	}

	return nil
}
