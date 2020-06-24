package registration

import (
	"errors"
	"net/http"
	"time"

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
	PreHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Request) error

	PostHookPostPersistExecutor interface {
		ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, a *Request, s *session.Session) error
	}
	PostHookPostPersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Request, s *session.Session) error

	PostHookPrePersistExecutor interface {
		ExecutePostRegistrationPrePersistHook(w http.ResponseWriter, r *http.Request, a *Request, i *identity.Identity) error
	}
	PostHookPrePersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Request, i *identity.Identity) error

	HooksProvider interface {
		PreRegistrationHooks() []PreHookExecutor
		PostRegistrationPrePersistHooks(credentialsType identity.CredentialsType) []PostHookPrePersistExecutor
		PostRegistrationPostPersistHooks(credentialsType identity.CredentialsType) []PostHookPostPersistExecutor
	}
)

func (f PreHookExecutorFunc) ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, a *Request) error {
	return f(w, r, a)
}
func (f PostHookPostPersistExecutorFunc) ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, a *Request, s *session.Session) error {
	return f(w, r, a, s)
}
func (f PostHookPrePersistExecutorFunc) ExecutePostRegistrationPrePersistHook(w http.ResponseWriter, r *http.Request, a *Request, i *identity.Identity) error {
	return f(w, r, a, i)
}

type (
	executorDependencies interface {
		identity.ManagementProvider
		identity.ValidationProvider
		HooksProvider
		x.LoggingProvider
		x.WriterProvider
	}
	HookExecutor struct {
		d executorDependencies
		c configuration.Provider
	}
	HookExecutorProvider interface {
		RegistrationExecutor() *HookExecutor
	}
)

func NewHookExecutor(d executorDependencies, c configuration.Provider) *HookExecutor {
	return &HookExecutor{
		d: d,
		c: c,
	}
}

func (e *HookExecutor) PostRegistrationHook(w http.ResponseWriter, r *http.Request, ct identity.CredentialsType, a *Request, i *identity.Identity) error {
	for _, executor := range e.d.PostRegistrationPrePersistHooks(ct) {
		if err := executor.ExecutePostRegistrationPrePersistHook(w, r, a, i); err != nil {
			if errors.Is(err, ErrHookAbortRequest) {
				return nil
			}
			return err
		}
	}

	// We need to make sure that the identity has a valid schema before passing it down to the identity pool.
	if err := e.d.IdentityValidator().Validate(i); err != nil {
		return err
		// We're now creating the identity because any of the hooks could trigger a "redirect" or a "session" which
		// would imply that the identity has to exist already.
	} else if err := e.d.IdentityManager().Create(r.Context(), i); err != nil {
		if errors.Is(err, sqlcon.ErrUniqueViolation) {
			return schema.NewDuplicateCredentialsError()
		}
		return err
	}

	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("A new identity has registered using self-service registration. Running post execution hooks.")

	s := session.NewSession(i, e.c, time.Now().UTC())
	for _, executor := range e.d.PostRegistrationPostPersistHooks(ct) {
		if err := executor.ExecutePostRegistrationPostPersistHook(w, r, a, s); err != nil {
			if errors.Is(err, ErrHookAbortRequest) {
				return nil
			}
			return err
		}
	}

	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("Post registration execution hooks completed successfully.")

	return x.SecureContentNegotiationRedirection(w, r, s.Declassify(), a.RequestURL,
		e.d.Writer(), e.c, x.SecureRedirectOverrideDefaultReturnTo(e.c.SelfServiceRegistrationReturnTo(ct.String())))
}

func (e *HookExecutor) PreRegistrationHook(w http.ResponseWriter, r *http.Request, a *Request) error {
	for _, executor := range e.d.PreRegistrationHooks() {
		if err := executor.ExecuteRegistrationPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
