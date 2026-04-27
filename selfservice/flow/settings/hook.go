// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	stderrors "errors"

	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/kratos/x/redir"

	"github.com/ory/pop/v6"
	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/x/events"

	"github.com/ory/kratos/session"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/schema"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
)

type (
	PreHookExecutor interface {
		ExecuteSettingsPreHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error
	}
	PreHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error

	PostHookPrePersistExecutor interface {
		ExecuteSettingsPrePersistHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity, s *session.Session) error
	}
	PostHookPrePersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity, s *session.Session) error

	PostHookPostPersistExecutor interface {
		ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *Flow, id *identity.Identity, s *session.Session) error
	}
	PostHookPostPersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, id *identity.Identity, s *session.Session) error

	HooksProvider interface {
		PreSettingsHooks(ctx context.Context) ([]PreHookExecutor, error)
		PostSettingsPrePersistHooks(ctx context.Context, settingsType string) ([]PostHookPrePersistExecutor, error)
		PostSettingsPostPersistHooks(ctx context.Context, settingsType string) ([]PostHookPostPersistExecutor, error)
	}

	executorDependencies interface {
		identity.ManagementProvider
		identity.ValidationProvider
		identity.PoolProvider
		identity.PrivilegedPoolProvider
		identity.PendingTraitsChangePersistenceProvider
		session.ManagementProvider
		session.PersistenceProvider
		config.Provider

		HandlerProvider
		HooksProvider
		FlowPersistenceProvider

		nosurfx.CSRFTokenGeneratorProvider
		logrusx.Provider
		httpx.WriterProvider
		otelx.Provider
		x.TransactionPersistenceProvider
	}
	HookExecutor struct {
		d executorDependencies
	}
	HookExecutorProvider interface {
		SettingsHookExecutor() *HookExecutor
	}
)

func (f PreHookExecutorFunc) ExecuteSettingsPreHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error {
	return f(w, r, a, s)
}

func (f PostHookPrePersistExecutorFunc) ExecuteSettingsPrePersistHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity, s *session.Session) error {
	return f(w, r, a, i, s)
}

func (f PostHookPostPersistExecutorFunc) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *Flow, id *identity.Identity, s *session.Session) error {
	return f(w, r, a, id, s)
}

func PostHookPostPersistExecutorNames(e []PostHookPostPersistExecutor) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
}

func PostHookPrePersistExecutorNames(e []PostHookPrePersistExecutor) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
}

func NewHookExecutor(d executorDependencies) *HookExecutor {
	return &HookExecutor{d: d}
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

func (e *HookExecutor) handleSettingsError(_ context.Context, _ http.ResponseWriter, r *http.Request, settingsType string, f *Flow, i *identity.Identity, flowError error) error {
	if f != nil {
		if i != nil {
			var group node.UiNodeGroup
			switch settingsType {
			case "password":
				group = node.PasswordGroup
			case "oidc":
				group = node.OpenIDConnectGroup
			}

			cont, err := container.NewFromStruct("", group, i.Traits, "traits")
			if err != nil {
				e.d.Logger().WithError(err).Error("could not update flow UI")
				return err
			}

			for _, n := range cont.Nodes {
				// we only set the value and not the whole field because we want to keep types from the initial form generation
				f.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
			}
		}

		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(e.d.GenerateCSRFToken(r))
		}
	}

	return flowError
}

func (e *HookExecutor) PostSettingsHook(ctx context.Context, w http.ResponseWriter, r *http.Request, settingsType string, ctxUpdate *UpdateContext, i *identity.Identity, opts ...PostSettingsHookOption) (err error) {
	ctx, span := e.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.flow.settings.HookExecutor.PostSettingsHook")
	defer otelx.End(span, &err)

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", settingsType).
		Debug("Running PostSettingsPrePersistHooks.")

	// Verify the redirect URL before we do any other processing.
	c := e.d.Config()
	returnTo, err := redir.SecureRedirectTo(r, c.SelfServiceBrowserDefaultReturnTo(ctx),
		redir.SecureRedirectUseSourceURL(ctxUpdate.Flow.RequestURL),
		redir.SecureRedirectAllowURLs(c.SelfServiceBrowserAllowedReturnToDomains(ctx)),
		redir.SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL(ctx)),
		redir.SecureRedirectOverrideDefaultReturnTo(
			e.d.Config().SelfServiceFlowSettingsReturnTo(ctx, settingsType,
				ctxUpdate.Flow.AppendTo(e.d.Config().SelfServiceFlowSettingsUI(ctx)))),
	)
	if err != nil {
		return err
	}

	hookOptions := new(postSettingsHookOptions)
	for _, f := range opts {
		f(hookOptions)
	}

	preHooks, err := e.d.PostSettingsPrePersistHooks(ctx, settingsType)
	if err != nil {
		return err
	}
	for k, executor := range preHooks {
		logFields := logrus.Fields{
			"executor":          fmt.Sprintf("%T", executor),
			"executor_position": k,
			"executors":         PostHookPrePersistExecutorNames(preHooks),
			"identity_id":       i.ID,
			"flow_method":       settingsType,
		}

		if err := executor.ExecuteSettingsPrePersistHook(w, r, ctxUpdate.Flow, i, ctxUpdate.Session); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().WithRequest(r).WithFields(logFields).
					Debug("A ExecuteSettingsPrePersistHook hook aborted early.")
				return nil
			}
			var group node.UiNodeGroup
			switch settingsType {
			case "password":
				group = node.PasswordGroup
			case "oidc":
				group = node.OpenIDConnectGroup
			}
			return flow.HandleHookError(w, r, ctxUpdate.Flow, i.Traits, group, err, e.d, e.d)
		}

		e.d.Logger().WithRequest(r).WithFields(logFields).Debug("ExecuteSettingsPrePersistHook completed successfully.")
	}

	options := []identity.ManagerOption{identity.ManagerExposeValidationErrorsForInternalTypeAssertion}
	if e.d.SessionManager().IsPrivileged(ctx, ctxUpdate.Session) {
		options = append(options, identity.ManagerAllowWriteProtectedTraits)
	}

	if err := e.d.IdentityManager().Update(ctx, i, options...); err != nil {
		if errors.Is(err, identity.ErrProtectedFieldModified()) {
			e.d.Logger().WithError(err).Debug("Modifying protected field requires re-authentication.")
			return errors.WithStack(NewFlowNeedsReAuth())
		}
		if errors.Is(err, sqlcon.ErrUniqueViolation()) {
			return schema.NewDuplicateCredentialsError(err)
		}
		return err
	}
	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Debug("An identity's settings have been updated.")

	ctxUpdate.UpdateIdentity(i)
	ctxUpdate.Flow.State = flow.StateSuccess
	if hookOptions.cb != nil {
		if err := hookOptions.cb(ctxUpdate); err != nil {
			return err
		}
	}

	newFlow, err := e.d.SettingsHandler().NewFlow(ctx, w, r, i, ctxUpdate.Session, ctxUpdate.Flow.Type)
	if err != nil {
		return err
	}

	ctxUpdate.Flow.UI = newFlow.UI
	ctxUpdate.Flow.UI.ResetMessages()
	ctxUpdate.Flow.UI.AddMessage(node.DefaultGroup, text.NewInfoSelfServiceSettingsUpdateSuccess())
	ctxUpdate.Flow.InternalContext = newFlow.InternalContext
	if err := e.d.SettingsFlowPersister().UpdateSettingsFlow(ctx, ctxUpdate.Flow); err != nil {
		return err
	}

	postHooks, err := e.d.PostSettingsPostPersistHooks(ctx, settingsType)
	if err != nil {
		return err
	}
	for k, executor := range postHooks {
		if err := executor.ExecuteSettingsPostPersistHook(w, r, ctxUpdate.Flow, i, ctxUpdate.Session); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().
					WithRequest(r).
					WithField("executor", fmt.Sprintf("%T", executor)).
					WithField("executor_position", k).
					WithField("executors", PostHookPostPersistExecutorNames(postHooks)).
					WithField("identity_id", i.ID).
					WithField("flow_method", settingsType).
					Debug("A ExecuteSettingsPostPersistHook hook aborted early.")
				return nil
			}
			return e.handleSettingsError(ctx, w, r, settingsType, ctxUpdate.Flow, i, err)
		}

		e.d.Logger().WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", PostHookPostPersistExecutorNames(postHooks)).
			WithField("identity_id", i.ID).
			WithField("flow_method", settingsType).
			Debug("ExecuteSettingsPostPersistHook completed successfully.")
	}

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", settingsType).
		Debug("Completed all PostSettingsPrePersistHooks and PostSettingsPostPersistHooks.")

	trace.SpanFromContext(ctx).AddEvent(events.NewSettingsSucceeded(
		ctx, ctxUpdate.Flow.ID, i.ID, string(ctxUpdate.Flow.Type), settingsType))

	if ctxUpdate.Flow.Type == flow.TypeAPI {
		updatedFlow, err := e.d.SettingsFlowPersister().GetSettingsFlow(ctx, ctxUpdate.Flow.ID)
		if err != nil {
			return err
		}
		// ContinueWith items are transient items, not stored in the database, and need to be carried over here, so
		// they can be returned to the client.
		updatedFlow.ContinueWithItems = ctxUpdate.Flow.ContinueWithItems

		e.d.Writer().Write(w, r, updatedFlow)
		return nil
	}

	if err := e.d.SessionManager().IssueCookie(ctx, w, r, ctxUpdate.Session); err != nil {
		return errors.WithStack(err)
	}

	if x.IsJSONRequest(r) {
		updatedFlow, err := e.d.SettingsFlowPersister().GetSettingsFlow(ctx, ctxUpdate.Flow.ID)
		if err != nil {
			return err
		}
		// ContinueWith items are transient items, not stored in the database, and need to be carried over here, so
		// they can be returned to the client.
		ctxUpdate.Flow.AddContinueWith(flow.NewContinueWithRedirectBrowserTo(returnTo.String()))
		updatedFlow.ContinueWithItems = ctxUpdate.Flow.ContinueWithItems

		e.d.Writer().Write(w, r, updatedFlow)
		return nil
	}

	redir.ContentNegotiationRedirection(w, r, i.CopyWithoutCredentials(), e.d.Writer(), returnTo.String())
	return nil
}

// settingsTypePendingTraitsChange is the settings strategy the verify_new_address
// hook is allowed to be configured under. The JSON config schema
// (selfServiceAfterSettingsProfileMethod) only admits verify_new_address inside
// the profile hook list, so a pending traits change can only ever originate
// from the profile strategy.
const settingsTypePendingTraitsChange = "profile"

// ErrPendingTraitsChangeSessionInvalid is returned by ApplyPendingTraitsChange
// when the session that created the pending change is missing, revoked,
// expired, disabled, or belongs to a different tenant. Callers should treat
// this as "the verification attempt is invalid" — the change is not applied.
var ErrPendingTraitsChangeSessionInvalid = stderrors.New("pending traits change session is no longer valid")

// ApplyPendingTraitsChange commits a pending traits change to the identity and
// fires the "after profile settings, post-persist" hook chain. It is the
// single entry point verification strategies use to resolve a pending change
// created by a settings flow.
//
// The originating session is validated inside the same transaction that
// applies the change, so a session that is revoked between the caller's
// token/code consumption and this call is still rejected. If validation
// succeeds, the loaded session is reused for the hook chain — no second
// GetSession round-trip.
//
// Returns ErrPendingTraitsChangeSessionInvalid when the session is not usable.
// Returns identity.ErrConcurrentModification when the identity's traits have
// drifted since the change was proposed.
func (e *HookExecutor) ApplyPendingTraitsChange(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ptc *identity.PendingTraitsChange,
) (i *identity.Identity, err error) {
	ctx, span := e.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.flow.settings.HookExecutor.ApplyPendingTraitsChange")
	defer otelx.End(span, &err)

	// Reject PTCs whose FK references were nulled or never written. The
	// session FK is ON DELETE SET NULL and the origin-flow FK is ON
	// DELETE CASCADE, so a NULL here means either a revoked session or
	// a writer that failed to set the link.
	if !ptc.SessionID.Valid || !ptc.OriginSettingsFlowID.Valid {
		return nil, errors.WithStack(ErrPendingTraitsChangeSessionInvalid)
	}

	var sess *session.Session
	err = e.d.TransactionalPersisterProvider().Transaction(ctx, func(ctx context.Context, _ *pop.Connection) error {
		// ExpandDefault loads the session's identity in the same query,
		// so IsActive() below checks the identity's active state and we
		// avoid a second round-trip to IdentityPool.GetIdentity for the
		// hash comparison.
		s, sErr := e.d.SessionPersister().GetSession(ctx, ptc.SessionID.UUID, session.ExpandDefault)
		if sErr != nil {
			if errors.Is(sErr, sqlcon.ErrNoRows()) {
				return errors.WithStack(ErrPendingTraitsChangeSessionInvalid)
			}
			return sErr
		}
		if s.NID != ptc.NID || !s.IsActive() {
			return errors.WithStack(ErrPendingTraitsChangeSessionInvalid)
		}
		// TODO(jonas): We can likely drop the identity id column entirely, since the session's identity is always loaded and we check the ID matches here. This would simplify the schema and remove a potential source of FK inconsistency.
		if s.Identity == nil || s.Identity.ID != ptc.IdentityID {
			return errors.WithStack(ErrPendingTraitsChangeSessionInvalid)
		}
		sess = s

		if identity.HashTraits(json.RawMessage(s.Identity.Traits)) != ptc.OriginalTraitsHash {
			return identity.ErrConcurrentModification
		}

		if err := e.d.IdentityManager().UpdateTraits(
			ctx, ptc.IdentityID, identity.Traits(ptc.ProposedTraits),
			identity.ManagerAllowWriteProtectedTraits,
		); err != nil {
			return err
		}

		// The PTC Status column is informational only; the real
		// single-submit guard is the verification flow state machine
		// (StatePassedChallenge) plus atomic token/code consumption.
		// We intentionally do not write Status here.

		// Re-read the identity after UpdateTraits so the returned
		// snapshot reflects the freshly-committed state, including any
		// VerifiableAddress row UpdateTraits may have created for the
		// new address.
		updated, err := e.d.IdentityPool().GetIdentity(ctx, ptc.IdentityID, identity.ExpandDefault)
		if err != nil {
			return err
		}

		for idx := range updated.VerifiableAddresses {
			a := &updated.VerifiableAddresses[idx]
			if a.Value == ptc.NewAddressValue && a.Via == ptc.NewAddressVia {
				a.Verified = true
				verifiedAt := sqlxx.NullTime(time.Now().UTC())
				a.VerifiedAt = &verifiedAt
				a.Status = identity.VerifiableAddressStatusCompleted
				if err := e.d.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, a, "verified", "verified_at", "status"); err != nil {
					return err
				}
				break
			}
		}

		i = updated
		// Keep Session.Identity in sync so post-persist webhook templates render the committed traits and the newly-verified address.
		// TODO: pointer aliasing could cause issues if a downstream hook implementation modifies the identity in-place instead of treating it as immutable. We should either enforce immutability or document this behavior.
		sess.Identity = updated
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Fire the "profile" post-persist hooks. Load the originating settings flow
	// so the hook chain receives a real Flow (honest ID, type, RequestURL)
	// instead of a synthetic one. The FK is ON DELETE CASCADE, so if the flow
	// was cleaned up the PTC was already deleted and we would not reach this
	// call. Apply has already committed, so a hook error cannot roll the change
	// back — we propagate errors so callers can surface them in the
	// verification UI.
	originFlow, err := e.d.SettingsFlowPersister().GetSettingsFlow(ctx, ptc.OriginSettingsFlowID.UUID)
	if err != nil {
		return nil, err
	}
	postHooks, err := e.d.PostSettingsPostPersistHooks(ctx, settingsTypePendingTraitsChange)
	if err != nil {
		return nil, err
	}
	for k, executor := range postHooks {
		logFields := logrus.Fields{
			"executor":          fmt.Sprintf("%T", executor),
			"executor_position": k,
			"executors":         PostHookPostPersistExecutorNames(postHooks),
			"identity_id":       i.ID,
			"flow_method":       settingsTypePendingTraitsChange,
		}
		if err := executor.ExecuteSettingsPostPersistHook(w, r, originFlow, i, sess); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().WithRequest(r).WithFields(logFields).
					Debug("A post-persist hook aborted the pending-traits-change continuation flow.")
				return i, nil
			}
			return nil, err
		}
		e.d.Logger().WithRequest(r).WithFields(logFields).
			Debug("ExecuteSettingsPostPersistHook completed successfully after pending traits change.")
	}

	return i, nil
}

func (e *HookExecutor) PreSettingsHook(ctx context.Context, w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) (err error) {
	ctx, span := e.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.flow.settings.HookExecutor.PreSettingsHook")
	defer otelx.End(span, &err)

	hooks, err := e.d.PreSettingsHooks(ctx)
	if err != nil {
		return err
	}
	for _, executor := range hooks {
		if err := executor.ExecuteSettingsPreHook(w, r, a, s); err != nil {
			return err
		}
	}

	return nil
}
