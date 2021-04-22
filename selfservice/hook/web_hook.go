package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-jsonnet"

	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var _ registration.PostHookPostPersistExecutor = new(WebHook)
var _ verification.PostHookExecutor = new(WebHook)

type (
	AuthStrategy interface {
		apply(req *http.Request)
	}

	authStrategyFactory func(c json.RawMessage) (AuthStrategy, error)

	noopAuthStrategy struct{}

	basicAuthStrategy struct {
		user     string
		password string
	}

	apiKeyStrategy struct {
		name  string
		value string
		in    string
	}

	webHookConfig struct {
		method       string
		url          string
		templatePath string
		auth         AuthStrategy
	}

	webHookDependencies interface {
		x.LoggingProvider
	}

	WebHook struct {
		r webHookDependencies
		c json.RawMessage
	}
)

var strategyFactories = map[string]authStrategyFactory{
	"":           newNoopAuthStrategy,
	"api-key":    newApiKeyStrategy,
	"basic-auth": newBasicAuthStrategy,
}

func newAuthStrategy(name string, c json.RawMessage) (as AuthStrategy, err error) {
	if f, ok := strategyFactories[name]; ok {
		as, err = f(c)
	} else {
		err = fmt.Errorf("unsupported auth type: %s", name)
	}
	return
}

func newNoopAuthStrategy(_ json.RawMessage) (AuthStrategy, error) {
	return &noopAuthStrategy{}, nil
}

func (c *noopAuthStrategy) apply(_ *http.Request) {}

func newBasicAuthStrategy(raw json.RawMessage) (AuthStrategy, error) {
	type config struct {
		User     string
		Password string
	}

	var c config
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}

	return &basicAuthStrategy{
		user:     c.User,
		password: c.Password,
	}, nil
}

func (c *basicAuthStrategy) apply(req *http.Request) {
	req.SetBasicAuth(c.user, c.password)
}

func newApiKeyStrategy(raw json.RawMessage) (AuthStrategy, error) {
	type config struct {
		In    string
		Name  string
		Value string
	}

	var c config
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}

	return &apiKeyStrategy{
		in:    c.In,
		name:  c.Name,
		value: c.Value,
	}, nil
}

func (c *apiKeyStrategy) apply(req *http.Request) {
	switch c.in {
	case "cookie":
		req.AddCookie(&http.Cookie{Name: c.name, Value: c.value})
	default:
		req.Header.Set(c.name, c.value)
	}
}

func newWebHookConfig(r json.RawMessage) (*webHookConfig, error) {
	type rawWebHookConfig struct {
		Method string
		Url    string
		Body   string
		Auth   struct {
			Type   string
			Config json.RawMessage
		}
	}

	var rc rawWebHookConfig
	err := json.Unmarshal(r, &rc)
	if err != nil {
		return nil, err
	}

	as, err := newAuthStrategy(rc.Auth.Type, rc.Auth.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create web hook auth strategy: %w", err)
	}

	return &webHookConfig{
		method:       rc.Method,
		url:          rc.Url,
		templatePath: rc.Body,
		auth:         as,
	}, nil
}

func NewWebHook(r webHookDependencies, c json.RawMessage) *WebHook {
	return &WebHook{r: r, c: c}
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, _ *http.Request, f *verification.Flow, s *session.Session) error {
	return e.execute(f, s)
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, _ *http.Request, f *registration.Flow, s *session.Session) error {
	return e.execute(f, s)
}

func (e *WebHook) execute(f interface{}, s interface{}) error {
	conf, err := newWebHookConfig(e.c)
	if err != nil {
		return err
	}

	body, err := createBody(conf.templatePath, f, s)
	if err != nil {
		return fmt.Errorf("failed to create web hook body: %w", err)
	}

	if err = doHttpCall(conf.method, conf.url, conf.auth, body); err != nil {
		return fmt.Errorf("failed to call web hook %w", err)
	}
	return nil
}

func createBody(templatePath string, f, s interface{}) (io.Reader, error) {
	var body io.Reader
	if len(templatePath) == 0 {
		return body, nil
	}

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

	if res, err := vm.EvaluateFile(templatePath); err != nil {
		return nil, err
	} else {
		return bytes.NewReader([]byte(res)), nil
	}
}

func doHttpCall(method string, url string, as AuthStrategy, body io.Reader) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	as.apply(req)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	} else if resp.StatusCode >= 400 {
		return fmt.Errorf("web hook failed with status code %v", resp.StatusCode)
	}

	return nil
}
