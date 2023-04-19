// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/x/httpx"
	"github.com/ory/x/otelx/semconv"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	PreHookExecutor interface {
		ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error
	}
	PreHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow) error

	PostHookPostPersistExecutor interface {
		ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error
	}
	PostHookPostPersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error

	PostHookPrePersistExecutor interface {
		ExecutePostRegistrationPrePersistHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error
	}
	PostHookPrePersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error

	HooksProvider interface {
		PreRegistrationHooks(ctx context.Context) []PreHookExecutor
		PostRegistrationPrePersistHooks(ctx context.Context, credentialsType identity.CredentialsType) []PostHookPrePersistExecutor
		PostRegistrationPostPersistHooks(ctx context.Context, credentialsType identity.CredentialsType) []PostHookPostPersistExecutor
	}
)

func ExecutorNames[T any](e []T) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
}

func (f PreHookExecutorFunc) ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	return f(w, r, a)
}

func (f PostHookPostPersistExecutorFunc) ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error {
	return f(w, r, a, s)
}

func (f PostHookPrePersistExecutorFunc) ExecutePostRegistrationPrePersistHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error {
	return f(w, r, a, i)
}

type (
	executorDependencies interface {
		config.Provider
		identity.ManagementProvider
		identity.ValidationProvider
		session.PersistenceProvider
		session.ManagementProvider
		HooksProvider
		hydra.HydraProvider
		x.CSRFTokenGeneratorProvider
		x.HTTPClientProvider
		x.LoggingProvider
		x.WriterProvider
		sessiontokenexchange.PersistenceProvider
	}
	HookExecutor struct {
		d executorDependencies
	}
	HookExecutorProvider interface {
		RegistrationExecutor() *HookExecutor
	}
)

func NewHookExecutor(d executorDependencies) *HookExecutor {
	return &HookExecutor{d: d}
}

func (e *HookExecutor) PostRegistrationHook(w http.ResponseWriter, r *http.Request, ct identity.CredentialsType, a *Flow, i *identity.Identity) error {
	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", ct).
		Debug("Running PostRegistrationPrePersistHooks.")
	for k, executor := range e.d.PostRegistrationPrePersistHooks(r.Context(), ct) {
		if err := executor.ExecutePostRegistrationPrePersistHook(w, r, a, i); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().
					WithRequest(r).
					WithField("executor", fmt.Sprintf("%T", executor)).
					WithField("executor_position", k).
					WithField("executors", ExecutorNames(e.d.PostRegistrationPrePersistHooks(r.Context(), ct))).
					WithField("identity_id", i.ID).
					WithField("flow_method", ct).
					Debug("A ExecutePostRegistrationPrePersistHook hook aborted early.")
				return nil
			}

			e.d.Logger().
				WithRequest(r).
				WithField("executor", fmt.Sprintf("%T", executor)).
				WithField("executor_position", k).
				WithField("executors", ExecutorNames(e.d.PostRegistrationPrePersistHooks(r.Context(), ct))).
				WithField("identity_id", i.ID).
				WithField("flow_method", ct).
				WithError(err).
				Error("ExecutePostRegistrationPostPersistHook hook failed with an error.")

			traits := i.Traits
			return flow.HandleHookError(w, r, a, traits, ct.ToUiNodeGroup(), err, e.d, e.d)
		}

		e.d.Logger().WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", ExecutorNames(e.d.PostRegistrationPrePersistHooks(r.Context(), ct))).
			WithField("identity_id", i.ID).
			WithField("flow_method", ct).
			Debug("ExecutePostRegistrationPrePersistHook completed successfully.")
	}

	// We need to make sure that the identity has a valid schema before passing it down to the identity pool.
	if err := e.d.IdentityValidator().Validate(r.Context(), i); err != nil {
		return err
		// We're now creating the identity because any of the hooks could trigger a "redirect" or a "session" which
		// would imply that the identity has to exist already.
	} else if err := e.d.IdentityManager().Create(r.Context(), i); err != nil {
		if errors.Is(err, sqlcon.ErrUniqueViolation) {
			// In this case the user is already registered through another method.
			// We handle this case by returning a spcial error that is handled by
			// the caller.
			return ErrDuplicateCredentials
		}
		return err
	}

	// Verify the redirect URL before we do any other processing.
	c := e.d.Config()
	returnTo, err := x.SecureRedirectTo(r, c.SelfServiceBrowserDefaultReturnTo(r.Context()),
		x.SecureRedirectReturnTo(a.ReturnTo),
		x.SecureRedirectUseSourceURL(a.RequestURL),
		x.SecureRedirectAllowURLs(c.SelfServiceBrowserAllowedReturnToDomains(r.Context())),
		x.SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL(r.Context())),
		x.SecureRedirectOverrideDefaultReturnTo(c.SelfServiceFlowRegistrationReturnTo(r.Context(), ct.String())),
	)
	if err != nil {
		return err
	}

	e.d.Audit().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Info("A new identity has registered using self-service registration.")
	trace.SpanFromContext(r.Context()).AddEvent(
		semconv.EventIdentityCreated,
		trace.WithAttributes(
			attribute.String(semconv.AttrIdentityID, i.ID.String()),
			attribute.String(semconv.AttrNID, i.NID.String()),
			attribute.String(semconv.AttrClientIP, httpx.ClientIP(r)),
			attribute.String("flow", string(a.Type)),
		),
	)

	s, err := session.NewActiveSession(r, i, e.d.Config(), time.Now().UTC(), ct, identity.AuthenticatorAssuranceLevel1)
	if err != nil {
		return err
	}

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", ct).
		Debug("Running PostRegistrationPostPersistHooks.")
	for k, executor := range e.d.PostRegistrationPostPersistHooks(r.Context(), ct) {
		if err := executor.ExecutePostRegistrationPostPersistHook(w, r, a, s); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().
					WithRequest(r).
					WithField("executor", fmt.Sprintf("%T", executor)).
					WithField("executor_position", k).
					WithField("executors", ExecutorNames(e.d.PostRegistrationPostPersistHooks(r.Context(), ct))).
					WithField("identity_id", i.ID).
					WithField("flow_method", ct).
					Debug("A ExecutePostRegistrationPostPersistHook hook aborted early.")
				return nil
			}

			e.d.Logger().
				WithRequest(r).
				WithField("executor", fmt.Sprintf("%T", executor)).
				WithField("executor_position", k).
				WithField("executors", ExecutorNames(e.d.PostRegistrationPostPersistHooks(r.Context(), ct))).
				WithField("identity_id", i.ID).
				WithField("flow_method", ct).
				WithError(err).
				Error("ExecutePostRegistrationPostPersistHook hook failed with an error.")

			traits := i.Traits
			return flow.HandleHookError(w, r, a, traits, ct.ToUiNodeGroup(), err, e.d, e.d)
		}

		e.d.Logger().WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", ExecutorNames(e.d.PostRegistrationPostPersistHooks(r.Context(), ct))).
			WithField("identity_id", i.ID).
			WithField("flow_method", ct).
			Debug("ExecutePostRegistrationPostPersistHook completed successfully.")
	}

	e.d.Logger().
		WithRequest(r).
		WithField("flow_method", ct).
		WithField("identity_id", i.ID).
		Debug("Post registration execution hooks completed successfully.")

	if a.Type == flow.TypeAPI || x.IsJSONRequest(r) {
		if _, ok, _ := e.d.SessionTokenExchangePersister().CodeForFlow(r.Context(), a.ID); ok {
			if err = e.d.SessionTokenExchangePersister().UpdateSessionOnExchanger(r.Context(), a.ID, s.ID); err != nil {
				return errors.WithStack(err)
			}
			http.Redirect(w, r, returnTo.String(), http.StatusFound)
			return nil
		}

		e.d.Writer().Write(w, r, &APIFlowResponse{
			Identity:     i,
			ContinueWith: a.ContinueWith(),
		})
		return nil
	}

	finalReturnTo := returnTo.String()
	if a.OAuth2LoginChallenge.Valid {
		cr, err := e.d.Hydra().AcceptLoginRequest(r.Context(), a.OAuth2LoginChallenge.UUID, i.ID.String(), s.AMR)
		if err != nil {
			return err
		}
		finalReturnTo = cr
	} else if a.ReturnToVerification != "" {
		finalReturnTo = a.ReturnToVerification
	}

	x.ContentNegotiationRedirection(w, r, s.Declassified(), e.d.Writer(), finalReturnTo)
	return nil
}

func (e *HookExecutor) PreRegistrationHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	for _, executor := range e.d.PreRegistrationHooks(r.Context()) {
		if err := executor.ExecuteRegistrationPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
