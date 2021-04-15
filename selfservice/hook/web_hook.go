package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"io"
	"net/http"
)

var _ registration.PostHookPostPersistExecutor = new(WebHook)
var _ verification.PostHookExecutor = new(WebHook)

type (
	webHookDependencies interface {
		x.LoggingProvider
	}

	webHookConfig struct {
		Method string
		Url    string
		Body   string
	}

	WebHook struct {
		r webHookDependencies
		c json.RawMessage
	}
)

func NewWebHook(r webHookDependencies, c json.RawMessage) *WebHook {
	return &WebHook{r: r, c: c}
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, _ *http.Request, f *verification.Flow, s *session.Session) error {
	return e.executeWebHook(f, s)
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, _ *http.Request, f *registration.Flow, s *session.Session) error {
	return e.executeWebHook(f, s)
}

func (e *WebHook) executeWebHook(f interface{}, s *session.Session) error {
	var conf webHookConfig
	if err := json.Unmarshal(e.c, &conf); err != nil {
		e.r.Logger().WithError(err).Error("Error while unmarshalling web_hook config json")
		return err
	}

	if body, err := e.createBody(conf, f, s); err != nil {
		return err
	} else if err = e.doHttpCall(conf, body); err != nil {
		return err
	}

	return nil
}

func (e *WebHook) createBody(_ webHookConfig, _ interface{}, s *session.Session) (io.Reader, error) {
	// TODO: make use of JSONNet
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(s.Identity); err != nil {
		e.r.Logger().WithError(err).Error("Error while marshalling web_hook payload json")
		return nil, err
	}

	return payloadBuf, nil
}

func (e *WebHook) doHttpCall(conf webHookConfig, body io.Reader) error {
	req, err := http.NewRequest(conf.Method, conf.Url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// TODO: Make use of authentication/authorization
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		e.r.Logger().WithError(err).Warn("Web hook failed")
		return err
	} else if resp.StatusCode >= 400 {
		return fmt.Errorf("web hook failed with status code %v", resp.StatusCode)
	}

	return nil
}
