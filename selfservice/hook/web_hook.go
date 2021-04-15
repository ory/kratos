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
	"github.com/google/go-jsonnet"
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

func (e *WebHook) executeWebHook(f interface{}, s interface{}) error {
	var conf webHookConfig
	if err := json.Unmarshal(e.c, &conf); err != nil {
		return err
	}

	if body, err := e.createBody(conf.Body, f, s); err != nil {
		e.r.Logger().WithError(err).Warn("Failed to create web hook payload")
		return err
	} else if err = e.doHttpCall(conf, body); err != nil {
		e.r.Logger().WithError(err).Warn("Failed to call the web hook")
		return err
	}

	return nil
}

func (e *WebHook) createBody(jsonnetFile string, f interface{}, s interface{}) (io.Reader, error) {
	vm := jsonnet.MakeVM()

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	if err := enc.Encode(f); err != nil {
		return nil, err
	}
	vm.TLACode("flow", buf.String())

	buf.Reset()
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	vm.TLACode("session", buf.String())

	if res, err := vm.EvaluateFile(jsonnetFile); err != nil {
		return nil, err
	} else {
		return bytes.NewReader([]byte(res)), nil
	}
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
		return err
	} else if resp.StatusCode >= 400 {
		return fmt.Errorf("web hook failed with status code %v", resp.StatusCode)
	}

	return nil
}
