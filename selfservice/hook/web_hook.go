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
	WebHook struct {
		r   webHookDependencies
		url string
	}
)

func NewWebHook(r webHookDependencies, url string) *WebHook {
	return &WebHook{r: r, url: url}
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, _ *http.Request, _ *verification.Flow, s *session.Session) error {
	return e.executeWebHook(s)
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, _ *http.Request, _ *registration.Flow, s *session.Session) error {
	return e.executeWebHook(s)
}

func (e *WebHook) executeWebHook(s *session.Session) error {
	payloadBuf := new(bytes.Buffer)
	err := json.NewEncoder(payloadBuf).Encode(s.Identity)
	if err != nil {
		e.r.Logger().WithError(err).Error("Error while marshalling web_hook content json")
	}

	response, err := http.Post(e.url, "application/json", payloadBuf)
	if err != nil {
		e.r.Logger().WithError(err).Warn("Web hook failed")
		return err
	}
	if response.StatusCode >= 400 {
		return fmt.Errorf("web hook failed with status code %v", response.StatusCode)
	}

	return nil
}
