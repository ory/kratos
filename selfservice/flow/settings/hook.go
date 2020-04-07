package settings

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	PostHookExecutor interface {
		ExecuteSettingsPostHook(w http.ResponseWriter, r *http.Request, a *Request, s *session.Session) error
	}
	HooksProvider interface {
		PostSettingsHooks(credentialsType string) []PostHookExecutor
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
		SettingsExecutor() *HookExecutor
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

func (e *HookExecutor) PostSettingsHook(w http.ResponseWriter, r *http.Request, hooks []PostHookExecutor, a *Request, ss *session.Session, i *identity.Identity) error {
	options := []identity.ManagerOption{identity.ManagerExposeValidationErrors}
	if ss.AuthenticatedAt.Add(e.c.SelfServicePrivilegedSessionMaxAge()).After(time.Now()) {
		options = append(options, identity.ManagerAllowWriteProtectedTraits)
	}

	if err := e.d.IdentityManager().Update(r.Context(), i, options...); err != nil {
		if errors.Is(err, identity.ErrProtectedFieldModified) {
			return errors.WithStack(ErrRequestNeedsReAuthentication)
		}
		return err
	}

	a.UpdateSuccessful = true
	if err := e.d.SettingsRequestPersister().UpdateSettingsRequest(r.Context(), a); err != nil {
		return err
	}

	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("An identity's settings have been updated, running post hooks.")

	// Now we execute the post-Settings hooks!
	for _, executor := range hooks {
		if err := executor.ExecuteSettingsPostHook(w, r, a, ss); err != nil {
			return err
		}
	}

	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("Post settings execution hooks completed successfully.")

	return nil
}
