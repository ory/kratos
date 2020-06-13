package hook

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
)

var (
	_ login.PostHookExecutor                  = new(Redirector)
	_ registration.PostHookPrePersistExecutor = new(Redirector)
	_ registration.PreHookExecutor            = new(Redirector)
	_ login.PreHookExecutor                   = new(Redirector)
	_ settings.PostHookPostPersistExecutor    = new(Redirector)
)

func NewRedirector(config json.RawMessage) *Redirector {
	return &Redirector{config: config}
}

type Redirector struct {
	config json.RawMessage
}

func (e *Redirector) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, _ *settings.Request, _ *identity.Identity) error {
	if err := e.do(w, r); err != nil {
		return err
	}
	return errors.WithStack(settings.ErrHookAbortRequest)
}

func (e *Redirector) ExecuteLoginPreHook(w http.ResponseWriter, r *http.Request, _ *login.Request) error {
	if err := e.do(w, r); err != nil {
		return err
	}
	return errors.WithStack(login.ErrHookAbortRequest)
}

func (e *Redirector) ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, _ *registration.Request) error {
	if err := e.do(w, r); err != nil {
		return err
	}
	return errors.WithStack(registration.ErrHookAbortRequest)
}

func (e *Redirector) ExecutePostRegistrationPrePersistHook(w http.ResponseWriter, r *http.Request, _ *registration.Request, _ *identity.Identity) error {
	if err := e.do(w, r); err != nil {
		return err
	}
	return errors.WithStack(registration.ErrHookAbortRequest)
}

func (e *Redirector) ExecuteSettingsPrePersistHook(w http.ResponseWriter, r *http.Request, _ *settings.Request, _ *identity.Identity) error {
	if err := e.do(w, r); err != nil {
		return err
	}
	return errors.WithStack(settings.ErrHookAbortRequest)
}

func (e *Redirector) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, _ *login.Request, _ *session.Session) error {
	if err := e.do(w, r); err != nil {
		return err
	}
	return errors.WithStack(login.ErrHookAbortRequest)
}

func (e *Redirector) do(w http.ResponseWriter, r *http.Request) error {
	rt := gjson.GetBytes(e.config, "default_redirect_url").String()
	if rt == "" {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("A redirector hook was configured without a default_redirect_url value set."))
	}

	http.Redirect(w, r, rt, http.StatusFound)
	return nil
}
