package hook

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
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

func (e *Redirector) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, f *settings.Request, _ *identity.Identity) error {
	return e.do(w, r, f.Type, settings.ErrHookAbortRequest)
}

func (e *Redirector) ExecuteLoginPreHook(w http.ResponseWriter, r *http.Request, f *login.Flow) error {
	return e.do(w, r, f.Type, login.ErrHookAbortFlow)
}

func (e *Redirector) ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, f *registration.Flow) error {
	return e.do(w, r, f.Type, registration.ErrHookAbortFlow)
}

func (e *Redirector) ExecutePostRegistrationPrePersistHook(w http.ResponseWriter, r *http.Request, f *registration.Flow, _ *identity.Identity) error {
	return e.do(w, r, f.Type, registration.ErrHookAbortFlow)
}

func (e *Redirector) ExecuteSettingsPrePersistHook(w http.ResponseWriter, r *http.Request, f *settings.Request, _ *identity.Identity) error {
	return e.do(w, r, f.Type, settings.ErrHookAbortRequest)
}

func (e *Redirector) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, f *login.Flow, _ *session.Session) error {
	return e.do(w, r, f.Type, login.ErrHookAbortFlow)
}

func (e *Redirector) do(w http.ResponseWriter, r *http.Request, ft flow.Type, abort error) error {
	if ft == flow.TypeAPI {
		// do nothing
		return nil
	}

	rt := gjson.GetBytes(e.config, "to").String()
	if rt == "" {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("A redirector hook was configured without a redirect_to value set."))
	}

	http.Redirect(w, r, rt, http.StatusFound)
	return abort
}
