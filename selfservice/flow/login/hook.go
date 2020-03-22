package login

import (
	"net/http"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
)

type (
	PreHookExecutor interface {
		ExecuteLoginPreHook(w http.ResponseWriter, r *http.Request, a *Request) error
	}
	PostHookExecutor interface {
		ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *Request, s *session.Session) error
	}
	HooksProvider interface {
		PreLoginHooks() []PreHookExecutor
		PostLoginHooks(credentialsType identity.CredentialsType) []PostHookExecutor
	}
)

type (
	executorDependencies interface {
		identity.ManagementProvider
		HooksProvider
	}
	HookExecutor struct {
		d executorDependencies
		c configuration.Provider
	}
	HookExecutorProvider interface {
		LoginHookExecutor() *HookExecutor
	}
)

func NewHookExecutor(d executorDependencies, c configuration.Provider) *HookExecutor {
	return &HookExecutor{d: d, c: c}
}

func (e *HookExecutor) PostLoginHook(w http.ResponseWriter, r *http.Request, hooks []PostHookExecutor, a *Request, i *identity.Identity) error {
	s := session.NewSession(i, r, e.c)

	for _, executor := range hooks {
		if err := executor.ExecuteLoginPostHook(w, r, a, s); err != nil {
			return err
		}
	}

	if s.WasIdentityModified() {
		if err := e.d.IdentityManager().Update(r.Context(), s.Identity); err != nil {
			return err
		}
	}

	s.ResetModifiedIdentityFlag()
	return nil
}

func (e *HookExecutor) PreLoginHook(w http.ResponseWriter, r *http.Request, a *Request) error {
	for _, executor := range e.d.PreLoginHooks() {
		if err := executor.ExecuteLoginPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
