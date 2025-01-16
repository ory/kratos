// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/x/pointerx"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/herodot"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

func (s *Strategy) RecoveryStrategyID() string {
	return string(recovery.RecoveryStrategyCode)
}

func (s *Strategy) PopulateRecoveryMethod(r *http.Request, f *recovery.Flow) error {
	f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		node.NewInputField("email", nil, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
			WithMetaLabel(text.NewInfoNodeInputEmail()),
	)
	f.UI.
		GetNodes().
		Append(node.NewInputField("method", s.RecoveryStrategyID(), node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(text.NewInfoNodeLabelContinue()))

	return nil
}

// Update Recovery Flow with Code Method
//
// swagger:model updateRecoveryFlowWithCodeMethod
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateRecoveryFlowWithCodeMethod struct {
	// The email address of the account to recover
	//
	// If the email belongs to a valid account, a recovery email will be sent.
	//
	// If you want to notify the email address if the account does not exist, see
	// the [notify_unknown_recipients flag](https://www.ory.sh/docs/kratos/self-service/flows/account-recovery-password-reset#attempted-recovery-notifications)
	//
	// If a code was already sent, including this field in the payload will invalidate the sent code and re-send a new code.
	//
	// format: email
	// required: false
	Email string `json:"email" form:"email"`

	// Code from the recovery email
	//
	// If you want to submit a code, use this field, but make sure to _not_ include the email field, as well.
	//
	// required: false
	Code string `json:"code" form:"code"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `form:"csrf_token" json:"csrf_token"`

	// Method is the method that should be used for this recovery flow
	//
	// Allowed values are `link` and `code`.
	//
	// required: true
	Method recovery.RecoveryMethod `json:"method"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (s *Strategy) isCodeFlow(f *recovery.Flow) bool {
	value, err := f.Active.Value()
	if err != nil {
		return false
	}
	return value == s.NodeGroup().String()
}

func (s *Strategy) Recover(w http.ResponseWriter, r *http.Request, f *recovery.Flow) (err error) {
	ctx, span := s.deps.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.code.Strategy.Recover")
	span.SetAttributes(attribute.String("selfservice_flows_recovery_use", s.deps.Config().SelfServiceFlowRecoveryUse(ctx)))
	defer otelx.End(span, &err)

	if !s.isCodeFlow(f) {
		span.SetAttributes(attribute.String("not_responsible_reason", "not code flow"))
		return errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	body, err := s.decodeRecovery(r)
	if err != nil {
		return s.HandleRecoveryError(w, r, nil, body, err)
	}

	f.TransientPayload = body.TransientPayload

	if f.DangerousSkipCSRFCheck {
		s.deps.Logger().
			WithRequest(r).
			Debugf("A recovery flow with `DangerousSkipCSRFCheck` set has been submitted, skipping anti-CSRF measures.")
	} else if err := flow.EnsureCSRF(s.deps, r, f.Type, s.deps.Config().DisableAPIFlowEnforcement(ctx), s.deps.GenerateCSRFToken, body.CSRFToken); err != nil {
		// If a CSRF violation occurs the flow is most likely FUBAR, as the user either lost the CSRF token, or an attack occured.
		// In this case, we just issue a new flow and "abandon" the old flow.
		return s.retryRecoveryFlow(w, r, flow.TypeBrowser, RetryWithError(err))
	}

	sID := s.RecoveryStrategyID()

	f.UI.ResetMessages()

	// If the email is present in the submission body, the user needs a new code via resend
	if f.State != flow.StateChooseMethod && len(body.Email) == 0 {
		if err := flow.MethodEnabledAndAllowed(ctx, flow.RecoveryFlow, sID, sID, s.deps); err != nil {
			return s.HandleRecoveryError(w, r, nil, body, err)
		}
		return s.recoveryUseCode(w, r, body, f)
	}

	if _, err := s.deps.SessionManager().FetchFromRequest(ctx, r); err == nil {
		// User is already logged in
		if x.IsJSONRequest(r) {
			session.RespondWithJSONErrorOnAuthenticated(s.deps.Writer(), recovery.ErrAlreadyLoggedIn)(w, r, nil)
		} else {
			session.RedirectOnAuthenticated(s.deps)(w, r, nil)
		}
		return errors.WithStack(flow.ErrCompletedByStrategy)
	}

	if err := flow.MethodEnabledAndAllowed(ctx, flow.RecoveryFlow, sID, body.Method, s.deps); err != nil {
		return s.HandleRecoveryError(w, r, nil, body, err)
	}

	recoveryFlow, err := s.deps.RecoveryFlowPersister().GetRecoveryFlow(ctx, x.ParseUUID(body.Flow))
	if err != nil {
		return s.HandleRecoveryError(w, r, recoveryFlow, body, err)
	}

	if err := recoveryFlow.Valid(); err != nil {
		return s.HandleRecoveryError(w, r, recoveryFlow, body, err)
	}

	switch recoveryFlow.State {
	case flow.StateChooseMethod,
		flow.StateEmailSent:
		return s.recoveryHandleFormSubmission(w, r, recoveryFlow, body)
	case flow.StatePassedChallenge:
		// was already handled, do not allow retry
		return s.retryRecoveryFlow(w, r, recoveryFlow.Type, RetryWithMessage(text.NewErrorValidationRecoveryRetrySuccess()))
	default:
		return s.retryRecoveryFlow(w, r, recoveryFlow.Type, RetryWithMessage(text.NewErrorValidationRecoveryStateFailure()))
	}
}

func (s *Strategy) recoveryIssueSession(w http.ResponseWriter, r *http.Request, f *recovery.Flow, id *identity.Identity) error {
	ctx := r.Context()

	f.UI.Messages.Clear()
	f.State = flow.StatePassedChallenge
	f.RecoveredIdentityID = uuid.NullUUID{
		UUID:  id.ID,
		Valid: true,
	}

	if f.Type == flow.TypeBrowser {
		f.SetCSRFToken(s.deps.CSRFHandler().RegenerateToken(w, r))
	}

	if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(ctx, f); err != nil {
		return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
	}

	sess := session.NewInactiveSession()
	sess.CompletedLoginFor(identity.CredentialsTypeRecoveryCode, identity.AuthenticatorAssuranceLevel1)
	if err := s.deps.SessionManager().ActivateSession(r, sess, id, time.Now().UTC()); err != nil {
		return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
	}

	if err := s.deps.RecoveryExecutor().PostRecoveryHook(w, r, f, sess); err != nil {
		return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
	}

	switch f.Type {
	case flow.TypeBrowser:
		if err := s.deps.SessionManager().UpsertAndIssueCookie(ctx, w, r, sess); err != nil {
			return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
		}
	case flow.TypeAPI:
		if err := s.deps.SessionPersister().UpsertSession(r.Context(), sess); err != nil {
			return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
		}
		f.ContinueWith = append(f.ContinueWith, flow.NewContinueWithSetToken(sess.Token))
	}

	sf, err := s.deps.SettingsHandler().NewFlow(ctx, w, r, sess.Identity, f.Type)
	if err != nil {
		return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
	}

	returnToURL := s.deps.Config().SelfServiceFlowRecoveryReturnTo(r.Context(), nil)
	returnTo := ""
	if returnToURL != nil {
		returnTo = returnToURL.String()
	}
	sf.RequestURL, err = x.TakeOverReturnToParameter(f.RequestURL, sf.RequestURL, returnTo)
	if err != nil {
		return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
	}

	config := s.deps.Config()

	sf.UI.Messages.Set(text.NewRecoverySuccessful(time.Now().Add(config.SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx))))
	if err := s.deps.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), sf); err != nil {
		return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
	}

	if s.deps.Config().UseContinueWithTransitions(ctx) {
		redirectTo := sf.AppendTo(s.deps.Config().SelfServiceFlowSettingsUI(r.Context())).String()
		switch {
		case f.Type.IsAPI(), x.IsJSONRequest(r):
			f.ContinueWith = append(f.ContinueWith, flow.NewContinueWithSettingsUI(sf, redirectTo))
			s.deps.Writer().Write(w, r, f)
		default:
			http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		}
	} else {
		if x.IsJSONRequest(r) {
			s.deps.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(sf.AppendTo(s.deps.Config().SelfServiceFlowSettingsUI(r.Context())).String()))
		} else {
			http.Redirect(w, r, sf.AppendTo(s.deps.Config().SelfServiceFlowSettingsUI(r.Context())).String(), http.StatusSeeOther)
		}
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) recoveryUseCode(w http.ResponseWriter, r *http.Request, body *recoverySubmitPayload, f *recovery.Flow) error {
	ctx := r.Context()
	code, err := s.deps.RecoveryCodePersister().UseRecoveryCode(ctx, f.ID, body.Code)
	if errors.Is(err, ErrCodeNotFound) {
		f.UI.Messages.Clear()
		f.UI.Messages.Add(text.NewErrorValidationRecoveryCodeInvalidOrAlreadyUsed())
		if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(ctx, f); err != nil {
			return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
		}

		if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
			http.Redirect(w, r, f.AppendTo(s.deps.Config().SelfServiceFlowRecoveryUI(r.Context())).String(), http.StatusSeeOther)
		} else {
			s.deps.Writer().Write(w, r, f)
		}
		return errors.WithStack(flow.ErrCompletedByStrategy)
	} else if err != nil {
		return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
	}

	// Important to expand everything here, as we need the data for recovery.
	recovered, err := s.deps.IdentityPool().GetIdentity(ctx, code.IdentityID, identity.ExpandEverything)
	if err != nil {
		return s.HandleRecoveryError(w, r, f, nil, err)
	}

	// mark address as verified only for a self-service flow
	if code.CodeType == RecoveryCodeTypeSelfService {
		if err := s.markRecoveryAddressVerified(w, r, f, recovered, code.RecoveryAddress); err != nil {
			return s.HandleRecoveryError(w, r, f, body, err)
		}
	}

	return s.recoveryIssueSession(w, r, f, recovered)
}

type retry struct {
	err     error
	message *text.Message
}

type RetryOption func(*retry)

func RetryWithError(err error) RetryOption {
	return func(r *retry) {
		r.err = err
	}
}

func RetryWithMessage(msg *text.Message) RetryOption {
	return func(r *retry) {
		r.message = msg
	}
}

func (s *Strategy) retryRecoveryFlow(w http.ResponseWriter, r *http.Request, ft flow.Type, opts ...RetryOption) error {
	retryOptions := retry{}

	for _, o := range opts {
		o(&retryOptions)
	}

	logger := s.deps.Logger().WithRequest(r).WithError(retryOptions.err)
	if retryOptions.message != nil {
		logger = logger.WithField("message", retryOptions.message)
	}
	logger.Debug("A recovery flow is being retried because a validation error occurred.")

	ctx := r.Context()
	config := s.deps.Config()

	f, err := recovery.NewFlow(config, config.SelfServiceFlowRecoveryRequestLifespan(ctx), s.deps.CSRFHandler().RegenerateToken(w, r), r, s, ft)
	if err != nil {
		return err
	}

	if retryOptions.message != nil {
		f.UI.Messages.Add(retryOptions.message)
	}

	if retryOptions.err != nil {
		if expired := new(flow.ExpiredError); errors.As(retryOptions.err, &expired) {
			f.UI.Messages.Add(text.NewErrorValidationRecoveryFlowExpired(expired.ExpiredAt))
		} else if err := f.UI.ParseError(node.CodeGroup, retryOptions.err); err != nil {
			return err
		}
	}
	if err := s.deps.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f); err != nil {
		return err
	}

	if s.deps.Config().UseContinueWithTransitions(ctx) {
		switch {
		case x.IsJSONRequest(r):
			rErr := new(herodot.DefaultError)
			if !errors.As(retryOptions.err, &rErr) {
				rErr = rErr.WithError(retryOptions.err.Error())
			}
			s.deps.Writer().WriteError(w, r, flow.ErrorWithContinueWith(rErr, flow.NewContinueWithRecoveryUI(f)))
		default:
			http.Redirect(w, r, f.AppendTo(config.SelfServiceFlowRecoveryUI(ctx)).String(), http.StatusSeeOther)
		}
	} else {
		if x.IsJSONRequest(r) {
			http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(config.SelfPublicURL(ctx),
				recovery.RouteGetFlow), url.Values{"id": {f.ID.String()}}).String(), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, f.AppendTo(config.SelfServiceFlowRecoveryUI(ctx)).String(), http.StatusSeeOther)
		}
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

// recoveryHandleFormSubmission handles the submission of an Email for recovery
func (s *Strategy) recoveryHandleFormSubmission(w http.ResponseWriter, r *http.Request, f *recovery.Flow, body *recoverySubmitPayload) error {
	if len(body.Email) == 0 {
		return s.HandleRecoveryError(w, r, f, body, schema.NewRequiredError("#/email", "email"))
	}

	ctx := r.Context()
	config := s.deps.Config()

	if err := flow.EnsureCSRF(s.deps, r, f.Type, config.DisableAPIFlowEnforcement(ctx), s.deps.GenerateCSRFToken, body.CSRFToken); err != nil {
		return s.HandleRecoveryError(w, r, f, body, err)
	}

	if err := s.deps.RecoveryCodePersister().DeleteRecoveryCodesOfFlow(ctx, f.ID); err != nil {
		return s.HandleRecoveryError(w, r, f, body, err)
	}

	f.TransientPayload = body.TransientPayload
	if err := s.deps.CodeSender().SendRecoveryCode(ctx, f, identity.VerifiableAddressTypeEmail, body.Email); err != nil {
		if !errors.Is(err, ErrUnknownAddress) {
			return s.HandleRecoveryError(w, r, f, body, err)
		}
		// Continue execution
	}

	// re-initialize the UI with a "clean" new state
	f.UI = &container.Container{
		Method: "POST",
		Action: flow.AppendFlowTo(urlx.AppendPaths(s.deps.Config().SelfPublicURL(r.Context()), recovery.RouteSubmitFlow), f.ID).String(),
	}

	f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))

	f.Active = sqlxx.NullString(s.NodeGroup())
	f.State = flow.StateEmailSent
	f.UI.Messages.Set(text.NewRecoveryEmailWithCodeSent())
	f.UI.Nodes.Append(node.NewInputField("code", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithInputAttributes(func(a *node.InputAttributes) {
		a.Required = true
		a.Pattern = "[0-9]+"
		a.MaxLength = CodeLength
	})).
		WithMetaLabel(text.NewInfoNodeLabelRecoveryCode()),
	)
	f.UI.Nodes.Append(node.NewInputField("method", s.NodeGroup(), node.CodeGroup, node.InputAttributeTypeHidden))

	f.UI.
		GetNodes().
		Append(node.NewInputField("method", s.RecoveryStrategyID(), node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(text.NewInfoNodeLabelContinue()))

	f.UI.Nodes.Append(node.NewInputField("email", body.Email, node.CodeGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoNodeResendOTP()),
	)
	if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return s.HandleRecoveryError(w, r, f, body, err)
	}

	return nil
}

func (s *Strategy) markRecoveryAddressVerified(w http.ResponseWriter, r *http.Request, f *recovery.Flow, id *identity.Identity, recoveryAddress *identity.RecoveryAddress) error {
	for k, v := range id.VerifiableAddresses {
		if v.Value == recoveryAddress.Value {
			id.VerifiableAddresses[k].Verified = true
			id.VerifiableAddresses[k].VerifiedAt = pointerx.Ptr(sqlxx.NullTime(time.Now().UTC()))
			id.VerifiableAddresses[k].Status = identity.VerifiableAddressStatusCompleted
			if err := s.deps.PrivilegedIdentityPool().UpdateVerifiableAddress(r.Context(), &id.VerifiableAddresses[k]); err != nil {
				return s.HandleRecoveryError(w, r, f, nil, err)
			}
		}
	}

	return nil
}

func (s *Strategy) HandleRecoveryError(w http.ResponseWriter, r *http.Request, flow *recovery.Flow, body *recoverySubmitPayload, err error) error {
	if flow != nil {
		email := ""
		if body != nil {
			email = body.Email
		}

		flow.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		flow.UI.GetNodes().Upsert(
			node.NewInputField("email", email, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeInputEmail()),
		)
	}

	return err
}

type recoverySubmitPayload struct {
	Method           string          `json:"method" form:"method"`
	Code             string          `json:"code" form:"code"`
	CSRFToken        string          `json:"csrf_token" form:"csrf_token"`
	Flow             string          `json:"flow" form:"flow"`
	Email            string          `json:"email" form:"email"`
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (s *Strategy) decodeRecovery(r *http.Request) (*recoverySubmitPayload, error) {
	var body recoverySubmitPayload

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(recoveryMethodSchema)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := s.dx.Decode(r, &body, compiler,
		decoderx.HTTPDecoderUseQueryAndBody(),
		decoderx.HTTPKeepRequestBody(true),
		decoderx.HTTPDecoderAllowedMethods("POST"),
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return &body, nil
}
