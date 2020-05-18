package settings

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type (
	PostHookPrePersistExecutor interface {
		ExecuteSettingsPrePersistHook(w http.ResponseWriter, r *http.Request, a *Request, s *identity.Identity) error
	}
	PostHookPrePersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Request, s *identity.Identity) error

	PostHookPostPersistExecutor interface {
		ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *Request, s *identity.Identity) error
	}
	PostHookPostPersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Request, s *identity.Identity) error

	HooksProvider interface {
		PostSettingsPrePersistHooks(settingsType string) []PostHookPrePersistExecutor
		PostSettingsPostPersistHooks(settingsType string) []PostHookPostPersistExecutor
	}
)

func (f PostHookPrePersistExecutorFunc) ExecuteSettingsPrePersistHook(w http.ResponseWriter, r *http.Request, a *Request, s *identity.Identity) error {
	return f(w, r, a, s)
}

func (f PostHookPostPersistExecutorFunc) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *Request, s *identity.Identity) error {
	return f(w, r, a, s)
}

type (
	executorDependencies interface {
		identity.ManagementProvider
		identity.ValidationProvider
		HooksProvider
		x.LoggingProvider
		RequestPersistenceProvider
		x.WriterProvider
	}
	HookExecutor struct {
		d executorDependencies
		c configuration.Provider
	}
	HookExecutorProvider interface {
		SettingsHookExecutor() *HookExecutor
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

type PostSettingsHookOption func(o *postSettingsHookOptions)

type postSettingsHookOptions struct {
	cb func(ctxUpdate *UpdateContext) error
}

func WithCallback(cb func(ctxUpdate *UpdateContext) error) func(o *postSettingsHookOptions) {
	return func(o *postSettingsHookOptions) {
		o.cb = cb
	}
}

func (e *HookExecutor) PostSettingsHook(w http.ResponseWriter, r *http.Request, settingsType string, ctxUpdate *UpdateContext, i *identity.Identity, opts ...PostSettingsHookOption) error {
	config := new(postSettingsHookOptions)
	for _, f := range opts {
		f(config)
	}

	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("An identity's settings have been updated, running post hooks.")

	for _, executor := range e.d.PostSettingsPrePersistHooks(settingsType) {
		if err := executor.ExecuteSettingsPrePersistHook(w, r, ctxUpdate.Request, i); err != nil {
			if errors.Is(err, ErrHookAbortRequest) {
				return nil
			}
			return err
		}
	}

	options := []identity.ManagerOption{identity.ManagerExposeValidationErrors}
	ttl := e.c.SelfServicePrivilegedSessionMaxAge()
	if ctxUpdate.Session.AuthenticatedAt.Add(ttl).After(time.Now()) {
		options = append(options, identity.ManagerAllowWriteProtectedTraits)
	}

	if err := e.d.IdentityManager().Update(r.Context(), i, options...); err != nil {
		if errors.Is(err, identity.ErrProtectedFieldModified) {
			e.d.Logger().WithField("error", fmt.Sprintf("%+v", err)).Debug("Modifying protected field requires re-authentication.")
			return errors.WithStack(ErrRequestNeedsReAuthentication)
		}
		return err
	}

	ctxUpdate.Session.Identity = i
	ctxUpdate.Request.UpdateSuccessful = true

	if config.cb != nil {
		if err := config.cb(ctxUpdate); err != nil {
			return err
		}
	}

	if method, ok := ctxUpdate.Request.Methods[settingsType]; ok {
		method.Config.ResetErrors()
	}

	if err := e.d.SettingsRequestPersister().UpdateSettingsRequest(r.Context(), ctxUpdate.Request); err != nil {
		return err
	}

	for _, executor := range e.d.PostSettingsPostPersistHooks(settingsType) {
		if err := executor.ExecuteSettingsPostPersistHook(w, r, ctxUpdate.Request, i); err != nil {
			if errors.Is(err, ErrHookAbortRequest) {
				return nil
			}
			return err
		}
	}

	e.d.Logger().
		WithField("identity_id", i.ID).
		Debug("Post settings execution hooks completed successfully.")

	return x.SecureContentNegotiationRedirection(w, r, ctxUpdate.Session.Declassify(), ctxUpdate.Request.RequestURL, e.d.Writer(), e.c,
		x.SecureRedirectOverrideDefaultReturnTo(
			e.c.SelfServiceSettingsReturnTo(settingsType,
				urlx.CopyWithQuery(
					e.c.SettingsURL(),
					url.Values{"request": {ctxUpdate.Request.ID.String()}}))))
}
