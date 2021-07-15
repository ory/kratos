package hook

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ory/kratos/selfservice/flow/verification"

	"github.com/ory/kratos/selfservice/flow/recovery"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
)

var (
	_ registration.PostHookPrePersistExecutor  = new(Error)
	_ registration.PostHookPostPersistExecutor = new(Error)
	_ registration.PreHookExecutor             = new(Error)

	_ login.PreHookExecutor  = new(Error)
	_ login.PostHookExecutor = new(Error)

	_ settings.PostHookPostPersistExecutor = new(Error)
	_ settings.PostHookPrePersistExecutor  = new(Error)
)

type Error struct {
	Config json.RawMessage
}

func (e Error) err(path string, abort error) error {
	switch gjson.GetBytes(e.Config, path).String() {
	case "err":
		return errors.New("err")
	case "abort":
		return abort
	}
	return nil
}

func (e Error) ExecuteSettingsPrePersistHook(w http.ResponseWriter, r *http.Request, a *settings.Flow, s *identity.Identity) error {
	return e.err("ExecuteSettingsPrePersistHook", settings.ErrHookAbortRequest)
}

func (e Error) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *settings.Flow, s *identity.Identity) error {
	return e.err("ExecuteSettingsPostPersistHook", settings.ErrHookAbortRequest)
}

func (e Error) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *login.Flow, s *session.Session) error {
	return e.err("ExecuteLoginPostHook", login.ErrHookAbortFlow)
}

func (e Error) ExecuteLoginPreHook(w http.ResponseWriter, r *http.Request, a *login.Flow) error {
	return e.err("ExecuteLoginPreHook", login.ErrHookAbortFlow)
}

func (e Error) ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, a *registration.Flow) error {
	return e.err("ExecuteRegistrationPreHook", registration.ErrHookAbortFlow)
}

func (e Error) ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, a *registration.Flow, s *session.Session, ct identity.CredentialsType) error {
	return e.err("ExecutePostRegistrationPostPersistHook", registration.ErrHookAbortFlow)
}

func (e Error) ExecutePostRegistrationPrePersistHook(w http.ResponseWriter, r *http.Request, a *registration.Flow, i *identity.Identity, ct identity.CredentialsType) error {
	return e.err("ExecutePostRegistrationPrePersistHook", registration.ErrHookAbortFlow)
}

func (e Error) ExecutePostRecoveryHook(w http.ResponseWriter, r *http.Request, a *recovery.Flow, s *session.Session) error {
	return e.err("ExecutePostRecoveryHook", recovery.ErrHookAbortFlow)
}

func (e Error) ExecutePostVerificationHook(w http.ResponseWriter, r *http.Request, a *verification.Flow, i *identity.Identity) error {
	return e.err("ExecutePostVerificationHook", verification.ErrHookAbortFlow)
}
