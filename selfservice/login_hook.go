package selfservice

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/session"
)

var ErrBreak = errors.New("break")

type HookLoginPreExecutor interface {
	ExecuteLoginPreHook(w http.ResponseWriter, r *http.Request, a *LoginRequest) error
}

type HookLoginPostExecutor interface {
	ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *LoginRequest, s *session.Session) error
}

type HookLoginPreExecutionProvider interface {
	AuthHookLoginPreExecutors() []HookLoginPreExecutor
}

type loginExecutorDependencies interface {
	identity.PoolProvider
	HookLoginPreExecutionProvider
}

type LoginExecutor struct {
	d loginExecutorDependencies
	c configuration.Provider
}

func NewLoginExecutor(d loginExecutorDependencies, c configuration.Provider) *LoginExecutor {
	return &LoginExecutor{d: d, c: c}
}

type LoginExecutionProvider interface {
	LoginExecutor() *LoginExecutor
}

type PostLoginHookProvider interface {
	PostLoginHooks(credentialsType identity.CredentialsType) []HookLoginPostExecutor
}

func (e *LoginExecutor) PostLoginHook(w http.ResponseWriter, r *http.Request, hooks []HookLoginPostExecutor, a *LoginRequest, i *identity.Identity) error {
	s := session.NewSession(i, r, e.c)

	for _, executor := range hooks {
		if err := executor.ExecuteLoginPostHook(w, r, a, s); err != nil {
			return err
		}
	}

	if s.WasIdentityModified() {
		if _, err := e.d.IdentityPool().Update(r.Context(), s.Identity); err != nil {
			return err
		}
	}

	s.ResetModifiedIdentityFlag()
	return nil
}

func (e *LoginExecutor) PreLoginHook(w http.ResponseWriter, r *http.Request, a *LoginRequest) error {
	for _, executor := range e.d.AuthHookLoginPreExecutors() {
		if err := executor.ExecuteLoginPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
