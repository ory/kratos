package selfservice

import (
	"net/http"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/session"
)

type HookRegistrationPreExecutor interface {
	ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, a *RegistrationRequest) error
}

type HookRegistrationPostExecutor interface {
	ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, a *RegistrationRequest, s *session.Session) error
}

type HookRegistrationPreExecutionProvider interface {
	AuthHookRegistrationPreExecutors() []HookRegistrationPreExecutor
}

type registrationExecutorDependencies interface {
	identity.PoolProvider
	HookRegistrationPreExecutionProvider
}

type RegistrationExecutor struct {
	d registrationExecutorDependencies
	c configuration.Provider
}

func NewRegistrationExecutor(
	d registrationExecutorDependencies,
	c configuration.Provider,
) *RegistrationExecutor {
	return &RegistrationExecutor{
		d: d,
		c: c,
	}
}

type RegistrationExecutionProvider interface {
	RegistrationExecutor() *RegistrationExecutor
}

type PostRegistrationHookProvider interface {
	PostRegistrationHooks(credentialsType identity.CredentialsType) []HookRegistrationPostExecutor
}

func (e *RegistrationExecutor) PostRegistrationHook(w http.ResponseWriter, r *http.Request, hooks []HookRegistrationPostExecutor, a *RegistrationRequest, i *identity.Identity) error {
	s := session.NewSession(i, r, e.c)

	if _, err := e.d.IdentityPool().Create(r.Context(), s.Identity); err != nil {
		return err
	}

	for _, executor := range hooks {
		if err := executor.ExecuteRegistrationPostHook(w, r, a, s); err != nil {
			return err
		}
	}

	if _, err := e.d.IdentityPool().Update(r.Context(), s.Identity); err != nil {
		return err
	}

	s.ResetModifiedIdentityFlag()
	return nil
}

func (e *RegistrationExecutor) PreRegistrationHook(w http.ResponseWriter, r *http.Request, a *RegistrationRequest) error {
	for _, executor := range e.d.AuthHookRegistrationPreExecutors() {
		if err := executor.ExecuteRegistrationPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
