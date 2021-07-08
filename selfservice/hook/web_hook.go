package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/settings"

	"github.com/ory/kratos/selfservice/flow/login"

	"github.com/ory/kratos/selfservice/flow/recovery"

	"github.com/google/go-jsonnet"

	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var _ registration.PostHookPostPersistExecutor = new(WebHook)
var _ registration.PostHookPrePersistExecutor = new(WebHook)
var _ verification.PostHookExecutor = new(WebHook)
var _ recovery.PostHookExecutor = new(WebHook)

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
		interrupt    bool
	}

	webHookDependencies interface {
		x.LoggingProvider
	}

	templateContext struct {
		Flow           flow.Flow             `json:"flow"`
		RequestHeaders http.Header           `json:"request_headers"`
		RequestMethod  string                `json:"request_method"`
		RequestUrl     string                `json:"request_url"`
		Identity       *identity.Identity    `json:"identity,omitempty"`
		Credentials    *identity.Credentials `json:"credentials,omitempty"`
	}

	WebHook struct {
		r webHookDependencies
		c json.RawMessage
	}

	detailedMessage struct {
		ID      int
		Text    string
		Type    string
		Context json.RawMessage `json:"context,omitempty"`
	}

	rawHookResponse struct {
		InstancePtr      string
		Message          string
		DetailedMessages []detailedMessage
	}
)

var strategyFactories = map[string]authStrategyFactory{
	"":           newNoopAuthStrategy,
	"api_key":    newApiKeyStrategy,
	"basic_auth": newBasicAuthStrategy,
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
		Interrupt bool
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
		interrupt:    rc.Interrupt,
	}, nil
}

func NewWebHook(r webHookDependencies, c json.RawMessage) *WebHook {
	return &WebHook{r: r, c: c}
}

func (e *WebHook) ExecuteLoginPreHook(_ http.ResponseWriter, req *http.Request, flow *login.Flow) error {
	return e.execute(&templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
	})
}

func (e *WebHook) ExecuteLoginPostHook(_ http.ResponseWriter, req *http.Request, flow *login.Flow, session *session.Session) error {
	return e.execute(&templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, req *http.Request, flow *verification.Flow, id *identity.Identity) error {
	return e.execute(&templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       id,
	})
}

func (e *WebHook) ExecutePostRecoveryHook(_ http.ResponseWriter, req *http.Request, flow *recovery.Flow, session *session.Session) error {
	return e.execute(&templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       session.Identity,
	})
}

func (e *WebHook) ExecuteRegistrationPreHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow) error {
	return e.execute(&templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
	})
}

func (e *WebHook) ExecutePostRegistrationPrePersistHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow, id *identity.Identity, ct identity.CredentialsType) error {
	credentials, _ := id.GetCredentials(ct)
	return e.execute(&templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       id,
		Credentials:    credentials,
	})
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *registration.Flow, session *session.Session, ct identity.CredentialsType) error {
	credentials, _ := session.Identity.GetCredentials(ct)
	return e.execute(&templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       session.Identity,
		Credentials:    credentials,
	})
}

func (e *WebHook) ExecuteSettingsPostPersistHook(_ http.ResponseWriter, req *http.Request, flow *settings.Flow, id *identity.Identity, ct identity.CredentialsType) error {
	credentials, _ := id.GetCredentials(ct)
	return e.execute(&templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestUrl:     req.RequestURI,
		Identity:       id,
		Credentials:    credentials,
	})
}

func (e *WebHook) execute(data *templateContext) error {
	// TODO: reminder for the future: move parsing of config to the web hook initialization
	conf, err := newWebHookConfig(e.c)
	if err != nil {
		return fmt.Errorf("failed to parse web hook config: %w", err)
	}

	var body io.Reader
	if conf.method != "TRACE" {
		// According to the HTTP spec any request method, but TRACE is allowed to
		// have a body. Even this is a really bad practice for some of them, like for
		// GET
		body, err = createBody(conf.templatePath, data)
		if err != nil {
			return fmt.Errorf("failed to create web hook body: %w", err)
		}
	}

	resp, err := doHttpCall(conf.method, conf.url, conf.auth, body)
	if err != nil {
		return fmt.Errorf("failed to call web hook %w", err)
	}

	if !conf.interrupt || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	err = parseResponse(resp)
	return err
}

func createBody(templatePath string, data *templateContext) (io.Reader, error) {
	var body io.Reader
	if len(templatePath) == 0 {
		return body, nil
	}

	vm := jsonnet.MakeVM()

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	vm.TLACode("ctx", buf.String())

	if res, err := vm.EvaluateFile(templatePath); err != nil {
		return nil, err
	} else {
		return bytes.NewReader([]byte(res)), nil
	}
}

func doHttpCall(method string, url string, as AuthStrategy, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	as.apply(req)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	} else if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("web hook failed with status code %v", resp.StatusCode)
	}

	return resp, nil
}

func parseResponse(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("empty response provided from the webhook")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read response body")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	hookResponse := rawHookResponse{
		DetailedMessages: []detailedMessage{},
	}

	if err = json.Unmarshal(body, &hookResponse); err != nil {
		return errors.Wrap(err, "hook response could not be unmarshalled properly")
	}

	detailedMessages := text.Messages{}
	for _, msg := range hookResponse.DetailedMessages {
		detailedMessages.Add(&text.Message{
			ID:      text.ID(msg.ID),
			Text:    msg.Text,
			Type:    text.Type(msg.Type),
			Context: msg.Context,
		})
	}

	return schema.NewValidationHookError(hookResponse.InstancePtr, hookResponse.Message, detailedMessages)
}
