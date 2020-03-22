package profile

import (
	"net/http"
	"time"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	PostHookExecutor interface {
		ExecuteProfileManagementPostHook(w http.ResponseWriter, r *http.Request, a *Request, s *session.Session) error
	}
	HooksProvider interface {
		PostProfileManagementHooks(credentialsType string) []PostHookExecutor
	}
)

type (
	executorDependencies interface {
		identity.ManagementProvider
		identity.ValidationProvider
		HooksProvider
		x.LoggingProvider
		RequestPersistenceProvider
	}
	HookExecutor struct {
		d executorDependencies
		c configuration.Provider
	}
	HookExecutorProvider interface {
		ProfileManagementExecutor() *HookExecutor
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

func (e *HookExecutor) PostProfileManagementHook(w http.ResponseWriter, r *http.Request, hooks []PostHookExecutor, a *Request, ss *session.Session, i *identity.Identity) error {
	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("An identity's profile was updated, running post hooks.")

	// Now we execute the post-ProfileManagement hooks!
	for _, executor := range hooks {
		if err := executor.ExecuteProfileManagementPostHook(w, r, a, ss); err != nil {
			return err
		}
	}

	options := []identity.ManagerOption{identity.ManagerExposeValidationErrors}
	if ss.AuthenticatedAt.Add(e.c.SelfServicePrivilegedSessionMaxAge()).After(time.Now()) {
		options = append(options, identity.ManagerAllowWriteProtectedTraits)
	}

	if err := e.d.IdentityManager().Update(r.Context(), i, options...); err != nil {
		return err
	}

	a.UpdateSuccessful = true
	if err := e.d.ProfileRequestPersister().UpdateProfileRequest(r.Context(), a); err != nil {
		return err
	}

	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("Post profile management execution hooks completed successfully.")

	return nil
}
