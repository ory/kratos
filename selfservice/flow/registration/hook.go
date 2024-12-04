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

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/events"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
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
		identity.PrivilegedPoolProvider
		identity.ValidationProvider
		login.FlowPersistenceProvider
		login.StrategyProvider
		session.PersistenceProvider
		session.ManagementProvider
		HooksProvider
		FlowPersistenceProvider
		hydra.Provider
		x.CSRFTokenGeneratorProvider
		x.HTTPClientProvider
		x.LoggingProvider
		x.WriterProvider
		x.TracingProvider
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

func (e *HookExecutor) PostRegistrationHook(w http.ResponseWriter, r *http.Request, ct identity.CredentialsType, provider, organizationID string, registrationFlow *Flow, i *identity.Identity) (err error) {
	ctx := r.Context()
	ctx, span := e.d.Tracer(ctx).Tracer().Start(ctx, "HookExecutor.PostRegistrationHook")
	r = r.WithContext(ctx)
	defer otelx.End(span, &err)

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", ct).
		Debug("Running PostRegistrationPrePersistHooks.")
	for k, executor := range e.d.PostRegistrationPrePersistHooks(ctx, ct) {
		if err := executor.ExecutePostRegistrationPrePersistHook(w, r, registrationFlow, i); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().
					WithRequest(r).
					WithField("executor", fmt.Sprintf("%T", executor)).
					WithField("executor_position", k).
					WithField("executors", ExecutorNames(e.d.PostRegistrationPrePersistHooks(ctx, ct))).
					WithField("identity_id", i.ID).
					WithField("flow_method", ct).
					Debug("A ExecutePostRegistrationPrePersistHook hook aborted early.")
				return nil
			}

			e.d.Logger().
				WithRequest(r).
				WithField("executor", fmt.Sprintf("%T", executor)).
				WithField("executor_position", k).
				WithField("executors", ExecutorNames(e.d.PostRegistrationPrePersistHooks(ctx, ct))).
				WithField("identity_id", i.ID).
				WithField("flow_method", ct).
				WithError(err).
				Error("ExecutePostRegistrationPostPersistHook hook failed with an error.")

			traits := i.Traits
			return flow.HandleHookError(w, r, registrationFlow, traits, ct.ToUiNodeGroup(), err, e.d, e.d)
		}

		e.d.Logger().WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", ExecutorNames(e.d.PostRegistrationPrePersistHooks(ctx, ct))).
			WithField("identity_id", i.ID).
			WithField("flow_method", ct).
			Debug("ExecutePostRegistrationPrePersistHook completed successfully.")
	}

	// We need to make sure that the identity has a valid schema before passing it down to the identity pool.
	if err := e.d.IdentityValidator().Validate(ctx, i); err != nil {
		return err
	}
	// We're now creating the identity because any of the hooks could trigger a "redirect" or a "session" which
	// would imply that the identity has to exist already.
	if err := e.d.IdentityManager().Create(ctx, i); err != nil {
		if errors.Is(err, sqlcon.ErrUniqueViolation) {
			strategy, err := e.d.AllLoginStrategies().Strategy(ct)
			if err != nil {
				return err
			}

			if strategy, ok := strategy.(login.LinkableStrategy); ok {
				duplicateIdentifier, err := e.getDuplicateIdentifier(ctx, i)
				if err != nil {
					return err
				}

				if err := strategy.SetDuplicateCredentials(
					registrationFlow,
					duplicateIdentifier,
					i.Credentials[ct],
					provider,
				); err != nil {
					return err
				}
			}
		}
		return err
	}

	// At this point the identity is already created and will not be rolled back, so
	// we want all PostPersist hooks to be able to continue even when the client cancels the request.
	ctx = context.WithoutCancel(ctx)
	r = r.WithContext(ctx)

	// Verify the redirect URL before we do any other processing.
	c := e.d.Config()
	returnTo, err := x.SecureRedirectTo(r, c.SelfServiceBrowserDefaultReturnTo(ctx),
		x.SecureRedirectReturnTo(registrationFlow.ReturnTo),
		x.SecureRedirectUseSourceURL(registrationFlow.RequestURL),
		x.SecureRedirectAllowURLs(c.SelfServiceBrowserAllowedReturnToDomains(ctx)),
		x.SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL(ctx)),
		x.SecureRedirectOverrideDefaultReturnTo(c.SelfServiceFlowRegistrationReturnTo(ctx, ct.String())),
	)
	if err != nil {
		return err
	}

	span.SetAttributes(otelx.StringAttrs(map[string]string{
		"return_to":       returnTo.String(),
		"flow_type":       string(registrationFlow.Type),
		"redirect_reason": "registration successful",
	})...)

	if registrationFlow.Type == flow.TypeBrowser && x.IsJSONRequest(r) {
		registrationFlow.AddContinueWith(flow.NewContinueWithRedirectBrowserTo(returnTo.String()))
	}

	e.d.Audit().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Info("A new identity has registered using self-service registration.")

	span.AddEvent(events.NewRegistrationSucceeded(ctx, registrationFlow.ID, i.ID, string(registrationFlow.Type), registrationFlow.Active.String(), provider))

	s := session.NewInactiveSession()

	s.CompletedLoginForWithProvider(ct, identity.AuthenticatorAssuranceLevel1, provider, organizationID)
	if err := e.d.SessionManager().ActivateSession(r, s, i, time.Now().UTC()); err != nil {
		return err
	}

	// We persist the session here so that subsequent hooks (like verification) can use it.
	if err := e.d.SessionPersister().UpsertSession(ctx, s); err != nil {
		return err
	}

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", ct).
		Debug("Running PostRegistrationPostPersistHooks.")
	for k, executor := range e.d.PostRegistrationPostPersistHooks(ctx, ct) {
		if err := executor.ExecutePostRegistrationPostPersistHook(w, r, registrationFlow, s); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().
					WithRequest(r).
					WithField("executor", fmt.Sprintf("%T", executor)).
					WithField("executor_position", k).
					WithField("executors", ExecutorNames(e.d.PostRegistrationPostPersistHooks(ctx, ct))).
					WithField("identity_id", i.ID).
					WithField("flow_method", ct).
					Debug("A ExecutePostRegistrationPostPersistHook hook aborted early.")

				span.SetAttributes(attribute.String("redirect_reason", "aborted by hook"), attribute.String("executor", fmt.Sprintf("%T", executor)))

				return nil
			}

			e.d.Logger().
				WithRequest(r).
				WithField("executor", fmt.Sprintf("%T", executor)).
				WithField("executor_position", k).
				WithField("executors", ExecutorNames(e.d.PostRegistrationPostPersistHooks(ctx, ct))).
				WithField("identity_id", i.ID).
				WithField("flow_method", ct).
				WithError(err).
				Error("ExecutePostRegistrationPostPersistHook hook failed with an error.")

			span.SetAttributes(attribute.String("redirect_reason", "hook error"), attribute.String("executor", fmt.Sprintf("%T", executor)))

			traits := i.Traits
			return flow.HandleHookError(w, r, registrationFlow, traits, ct.ToUiNodeGroup(), err, e.d, e.d)
		}

		e.d.Logger().WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", ExecutorNames(e.d.PostRegistrationPostPersistHooks(ctx, ct))).
			WithField("identity_id", i.ID).
			WithField("flow_method", ct).
			Debug("ExecutePostRegistrationPostPersistHook completed successfully.")
	}

	e.d.Logger().
		WithRequest(r).
		WithField("flow_method", ct).
		WithField("identity_id", i.ID).
		Debug("Post registration execution hooks completed successfully.")

	if registrationFlow.Type == flow.TypeAPI || x.IsJSONRequest(r) {
		span.SetAttributes(attribute.String("flow_type", string(flow.TypeAPI)))

		if registrationFlow.IDToken != "" {
			// We don't want to redirect with the code, if the flow was submitted with an ID token.
			// This is the case for Sign in with native Apple SDK or Google SDK.
		} else if handled, err := e.d.SessionManager().MaybeRedirectAPICodeFlow(w, r, registrationFlow, s.ID, ct.ToUiNodeGroup()); err != nil {
			return errors.WithStack(err)
		} else if handled {
			return nil
		}

		e.d.Writer().Write(w, r, &APIFlowResponse{
			Identity:     i,
			ContinueWith: registrationFlow.ContinueWith(),
		})
		return nil
	}

	finalReturnTo := returnTo.String()
	if registrationFlow.OAuth2LoginChallenge != "" {
		if registrationFlow.ReturnToVerification != "" {
			// Special case: If Kratos is used as a login UI *and* we want to show the verification UI,
			// redirect to the verification URL first and then return to Hydra.
			finalReturnTo = registrationFlow.ReturnToVerification
		} else {
			callbackURL, err := e.d.Hydra().AcceptLoginRequest(ctx,
				hydra.AcceptLoginRequestParams{
					LoginChallenge:        string(registrationFlow.OAuth2LoginChallenge),
					IdentityID:            i.ID.String(),
					SessionID:             s.ID.String(),
					AuthenticationMethods: s.AMR,
				})
			if err != nil {
				return err
			}
			finalReturnTo = callbackURL
		}
		span.SetAttributes(attribute.String("redirect_reason", "oauth2 login challenge"))
	} else if registrationFlow.ReturnToVerification != "" {
		finalReturnTo = registrationFlow.ReturnToVerification
		span.SetAttributes(attribute.String("redirect_reason", "verification requested"))
	}
	span.SetAttributes(attribute.String("return_to", finalReturnTo))

	x.ContentNegotiationRedirection(w, r, s.Declassified(), e.d.Writer(), finalReturnTo)
	return nil
}

func (e *HookExecutor) getDuplicateIdentifier(ctx context.Context, i *identity.Identity) (string, error) {
	_, id, err := e.d.IdentityManager().ConflictingIdentity(ctx, i)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (e *HookExecutor) PreRegistrationHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	for _, executor := range e.d.PreRegistrationHooks(r.Context()) {
		if err := executor.ExecuteRegistrationPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
