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
	identity.ValidationProvider
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

	// We need to make sure that the identity has a valid schema before passing it down to the identity pool.
	if err := e.d.IdentityValidator().Validate(s.Identity); err != nil {
		return err
	// We're now creating the identity because any of the hooks could trigger a "redirect" or a "session" which
	// would imply that the identity has to exist already.
	} else if _, err := e.d.IdentityPool().Create(r.Context(), s.Identity); err != nil {
		return err
	}

	// Now we execute the post-registration hooks!
	for _, executor := range hooks {
		if err := executor.ExecuteRegistrationPostHook(w, r, a, s); err != nil {
			// TODO https://github.com/ory/hive/issues/51 #51
			return err
		}
	}

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
