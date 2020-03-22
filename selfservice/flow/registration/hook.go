package registration

import (
	"net/http"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	PreHookExecutor interface {
		ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, a *Request) error
	}
	PostHookExecutor interface {
		ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, a *Request, s *session.Session) error
	}
	HooksProvider interface {
		PreRegistrationHooks() []PreHookExecutor
		PostRegistrationHooks(credentialsType identity.CredentialsType) []PostHookExecutor
	}
)

type (
	executorDependencies interface {
		identity.ManagementProvider
		identity.ValidationProvider
		HooksProvider
		x.LoggingProvider
	}
	HookExecutor struct {
		d executorDependencies
		c configuration.Provider
	}
	HookExecutorProvider interface {
		RegistrationExecutor() *HookExecutor
	}
)

func NewHookExecutor(
	d executorDependencies,
	c configuration.Provider,
) *HookExecutor {
	return &HookExecutor{
		d: d,
		c: c,
	}
}

func (e *HookExecutor) PostRegistrationHook(w http.ResponseWriter, r *http.Request, hooks []PostHookExecutor, a *Request, i *identity.Identity) error {
	s := session.NewSession(i, r, e.c)

	// We need to make sure that the identity has a valid schema before passing it down to the identity pool.
	if err := e.d.IdentityValidator().Validate(s.Identity); err != nil {
		return err
		// We're now creating the identity because any of the hooks could trigger a "redirect" or a "session" which
		// would imply that the identity has to exist already.
	} else if err := e.d.IdentityManager().Create(r.Context(), s.Identity); err != nil {
		if errorsx.Cause(err) == sqlcon.ErrUniqueViolation {
			return schema.NewDuplicateCredentialsError()
		}
		return err
	}

	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("A new identity has registered using self-service registration. Running post execution hooks.")

	// Now we execute the post-registration hooks!
	for _, executor := range hooks {
		if err := executor.ExecuteRegistrationPostHook(w, r, a, s); err != nil {
			// TODO https://github.com/ory/kratos/issues/51 #51
			return err
		}
	}

	// We need to make sure that the identity has a valid schema before passing it down to the identity pool.
	if err := e.d.IdentityValidator().Validate(s.Identity); err != nil {
		return err
		// We're now creating the identity because any of the hooks could trigger a "redirect" or a "session" which
		// would imply that the identity has to exist already.
	} else if err := e.d.IdentityManager().Update(r.Context(), s.Identity); err != nil {
		return err
	}

	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("Post registration execution hooks completed successfully.")

	return nil
}

func (e *HookExecutor) PreRegistrationHook(w http.ResponseWriter, r *http.Request, a *Request) error {
	for _, executor := range e.d.PreRegistrationHooks() {
		if err := executor.ExecuteRegistrationPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
