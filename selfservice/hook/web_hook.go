package hook

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/go-jsonnet"
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

	basicAuthConfig struct {
		User     string
		Password string
	}

	apiKeyConfig struct {
		Name  string
		Value string
		In    string
	}

	AuthConfig interface {
		apply(req *http.Request)
	}

	Auth struct {
		Type       string
		AuthConfig AuthConfig      `json:"-"`
		RawConfig  json.RawMessage `json:"config"`
	}

	webHookConfig struct {
		Method string
		Url    string
		Body   string
		Auth   Auth
	}

	WebHook struct {
		r webHookDependencies
		c json.RawMessage
	}
)

func (c *basicAuthConfig) apply(req *http.Request) {
	credentials := base64.RawStdEncoding.EncodeToString([]byte(c.User + ":" + c.Password))
	req.Header.Set("Authorization", "Basic "+credentials)
}

func (c *apiKeyConfig) apply(req *http.Request) {
	switch c.In {
	case "cookie":
		req.Header.Add("Cookie", c.Name + "=" + c.Value)
	default:
		req.Header.Set(c.Name, c.Value)
	}
}

func (a *Auth) UnmarshalJSON(bytes []byte) error {
	type auth Auth
	err := json.Unmarshal(bytes, (*auth)(a))
	if err != nil {
		return err
	}
	println("type: " + a.Type)
	switch a.Type {
	case "basic-auth":
		var authConfig basicAuthConfig
		json.Unmarshal(a.RawConfig, &authConfig)
		a.AuthConfig = &authConfig
	case "api-key":
		var authConfig apiKeyConfig
		json.Unmarshal(a.RawConfig, &authConfig)
		a.AuthConfig = &authConfig
	default:
		return fmt.Errorf("unknown auth type %v", a.Type)
	}
	return nil
}

func NewWebHook(r webHookDependencies, c json.RawMessage) *WebHook {
	return &WebHook{r: r, c: c}
}

func (e *WebHook) ExecutePostVerificationHook(_ http.ResponseWriter, _ *http.Request, f *verification.Flow, s *session.Session) error {
	return e.executeWebHook(f, s)
}

func (e *WebHook) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, _ *http.Request, f *registration.Flow, s *session.Session) error {
	return e.executeWebHook(f, s)
}

func (e *WebHook) executeWebHook(f interface{}, s interface{}) (err error) {
	var conf webHookConfig
	if err := json.Unmarshal(e.c, &conf); err != nil {
		return err
	}
	var body io.Reader
	if len(conf.Body) != 0 {
		body, err = e.createBody(conf.Body, f, s)
		if err != nil {
			return fmt.Errorf("failed to create web hook %w", err)
		}
	}
	if err = e.doHttpCall(conf, body); err != nil {
		return fmt.Errorf("failed to call web hook %w", err)
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

	conf.Auth.AuthConfig.apply(req)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	} else if resp.StatusCode >= 400 {
		return fmt.Errorf("web hook failed with status code %v", resp.StatusCode)
	}

	return nil
}
