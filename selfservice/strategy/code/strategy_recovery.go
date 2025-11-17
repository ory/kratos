// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/ory/kratos/x/redir"

	"github.com/ory/x/otelx/semconv"
	"github.com/ory/x/pointerx"
	"github.com/ory/x/sqlcon"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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

// This builds the initial UI (first recovery screen).
func (s *Strategy) PopulateRecoveryMethod(r *http.Request, f *recovery.Flow) error {
	switch f.State {
	case flow.StateChooseMethod:
		f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		f.UI.GetNodes().Upsert(
			node.NewInputField("email", nil, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeInputEmail()),
		)
	case flow.StateRecoveryAwaitingAddress:
		// re-initialize the UI with a "clean" new state
		f.UI = &container.Container{
			Method: "POST",
			Action: flow.AppendFlowTo(urlx.AppendPaths(s.deps.Config().SelfPublicURL(r.Context()), recovery.RouteSubmitFlow), f.ID).String(),
		}
		f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		f.UI.GetNodes().Append(
			node.NewInputField("recovery_address", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewRecoveryAskAnyRecoveryAddress()),
		)
	default:
		// Unreachable.
		return errors.Errorf("unreachable state: %s", f.State)
	}

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

	// A recovery address that is registered for the user.
	// It can be an email, a phone number (to receive the code via SMS), etc.
	// Used in RecoveryV2.
	RecoveryAddress string `json:"recovery_address" form:"recovery_address"`

	// If there are multiple addresses registered for the user, a choice is presented and this field
	// stores the result of this choice.
	// Addresses are 'masked' (never sent in full to the client and shown partially in the UI) since at this point in the recovery flow,
	// the user has not yet proven that it knows the full address and we want to avoid
	// information exfiltration.
	// So for all intents and purposes, the value of this field should be treated as an opaque identifier.
	// Used in RecoveryV2.
	RecoverySelectAddress string `json:"recovery_select_address" form:"recovery_select_address"`

	// If there are multiple recovery addresses registered for the user, and the initially provided address
	// is different from the address chosen when the choice (of masked addresses) is presented, then we need to make sure
	// that the user actually knows the full address to avoid information exfiltration, so we ask for the full address.
	// Used in RecoveryV2.
	RecoveryConfirmAddress string `json:"recovery_confirm_address" form:"recovery_confirm_address"`

	// Set to "previous" to go back in the flow, meaningfully.
	// Used in RecoveryV2.
	Screen string `json:"screen" form:"screen"`
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

	// NOTE: This is implicitly looking at the state machine (for Recovery v1), by inspecting which fields are present,
	// instead of inspecting the state explicitly.
	// For Recovery v2 we inspect the state explicitly, a few lines below.
	if !flow.IsStateRecoveryV2(f.State) {
		// If the email is not present in the submission body, the user needs a new code via resend
		if f.State != flow.StateChooseMethod && len(body.Email) == 0 {
			if err := flow.MethodEnabledAndAllowed(ctx, flow.RecoveryFlow, sID, sID, s.deps); err != nil {
				return s.HandleRecoveryError(w, r, nil, body, err)
			}
			return s.recoveryUseCode(w, r, body, f)
		}
	}

	if _, err := s.deps.SessionManager().FetchFromRequest(ctx, r); err == nil {
		// User is already logged in
		if x.IsJSONRequest(r) {
			session.RespondWithJSONErrorOnAuthenticated(s.deps.Writer(), recovery.ErrAlreadyLoggedIn)(w, r)
		} else {
			session.RedirectOnAuthenticated(s.deps)(w, r)
		}
		return errors.WithStack(flow.ErrCompletedByStrategy)
	}

	// Recovery V1 sets some magic fields in the UI and inspects them in the body, e.g. `method`.
	// This is brittle and rendered unnecessary in Recovery V2 by properly inspecting the `state` (and the CSRF token).
	if !flow.IsStateRecoveryV2(f.State) {
		if err := flow.MethodEnabledAndAllowed(ctx, flow.RecoveryFlow, sID, body.Method, s.deps); err != nil {
			return s.HandleRecoveryError(w, r, nil, body, err)
		}
	}

	recoveryFlow, err := s.deps.RecoveryFlowPersister().GetRecoveryFlow(ctx, x.ParseUUID(body.Flow))
	if err != nil {
		return s.HandleRecoveryError(w, r, recoveryFlow, body, err)
	}

	if err := recoveryFlow.Valid(); err != nil {
		return s.HandleRecoveryError(w, r, recoveryFlow, body, err)
	}

	if body.Screen == "previous" {
		return s.recoveryV2HandleGoBack(r, f, body)
	}

	switch recoveryFlow.State {
	case flow.StateChooseMethod,
		flow.StateEmailSent:
		return s.recoveryHandleFormSubmission(w, r, recoveryFlow, body)
	case flow.StatePassedChallenge:
		// was already handled, do not allow retry
		return s.retryRecoveryFlow(w, r, recoveryFlow.Type, RetryWithMessage(text.NewErrorValidationRecoveryRetrySuccess()))

		// Recovery V2.
	case flow.StateRecoveryAwaitingAddress:
		return s.recoveryV2HandleStateAwaitingAddress(r, recoveryFlow, body)
	case flow.StateRecoveryAwaitingAddressChoice:
		return s.recoveryV2HandleStateAwaitingAddressChoice(r, recoveryFlow, body)
	case flow.StateRecoveryAwaitingAddressConfirm:
		return s.recoveryV2HandleStateConfirmingAddress(r, recoveryFlow, body)
	case flow.StateRecoveryAwaitingCode:
		return s.recoveryV2HandleStateAwaitingCode(w, r, recoveryFlow, body)

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

	sf, err := s.deps.SettingsHandler().NewFlow(ctx, w, r, sess.Identity, sess, f.Type)
	if err != nil {
		return s.retryRecoveryFlow(w, r, f.Type, RetryWithError(err))
	}

	returnToURL := s.deps.Config().SelfServiceFlowRecoveryReturnTo(r.Context(), nil)
	returnTo := ""
	if returnToURL != nil {
		returnTo = returnToURL.String()
	}
	sf.RequestURL, err = redir.TakeOverReturnToParameter(f.RequestURL, sf.RequestURL, returnTo)
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
		trace.SpanFromContext(r.Context()).AddEvent(semconv.NewDeprecatedFeatureUsedEvent(r.Context(), "no_continue_with_transition_recovery_issue_session"))
		if x.IsJSONRequest(r) {
			s.deps.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(sf.AppendTo(s.deps.Config().SelfServiceFlowSettingsUI(r.Context())).String()))
		} else {
			http.Redirect(w, r, sf.AppendTo(s.deps.Config().SelfServiceFlowSettingsUI(r.Context())).String(), http.StatusSeeOther)
		}
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

// NOTE: This function handles two cases:
//   - A code was submitted: try to use it
//   - No code was submitted: delete all existing codes, re-generate a new one, send it.
//     This corresponds to the user clicking on the 're-send code' button.
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
		trace.SpanFromContext(r.Context()).AddEvent(semconv.NewDeprecatedFeatureUsedEvent(r.Context(), "no_continue_with_transition_recovery_retry_flow_handler"))
		if x.IsJSONRequest(r) {
			http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(config.SelfPublicURL(ctx),
				recovery.RouteGetFlow), url.Values{"id": {f.ID.String()}}).String(), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, f.AppendTo(config.SelfServiceFlowRecoveryUI(ctx)).String(), http.StatusSeeOther)
		}
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func AddressToHashBase64(address string) string {
	hash := sha256.Sum256([]byte(address))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (s *Strategy) recoveryV2HandleStateAwaitingAddress(r *http.Request, f *recovery.Flow, body *recoverySubmitPayload) error {
	if f.State != flow.StateRecoveryAwaitingAddress {
		return errors.Errorf("unreachable state: %s", f.State)
	}

	if len(body.RecoveryAddress) == 0 {
		return schema.NewRequiredError("#/recovery_address", "recovery_address")
	}

	// Need to retrieve all possible recovery addresses and present a choice.
	recoveryAddresses, err := s.deps.IdentityPool().FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(r.Context(), body.RecoveryAddress)
	// Real error.
	if err != nil && !errors.Is(err, sqlcon.ErrNoRows) {
		return err
	}

	// No rows returned.
	if len(recoveryAddresses) == 0 {
		// To avoid an attacker from using this case to probe for existing addresses, we pretend it exists.
		// This is the same behavior as in Recovery V1.
		recoveryAddresses = append(recoveryAddresses, identity.RecoveryAddress{Value: body.RecoveryAddress})
	}

	f.State = flow.StateRecoveryAwaitingAddressChoice

	if len(recoveryAddresses) == 1 && recoveryAddresses[0].Value == body.RecoveryAddress {
		// Skip two states for convenience:
		// - No need to present a choice with only one option
		// - No need to ask for the full address if there is only one and it was just provided in full

		body.RecoveryConfirmAddress = body.RecoveryAddress
		f.State = flow.StateRecoveryAwaitingAddressConfirm
		if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
			return err
		}
		return s.recoveryV2HandleStateConfirmingAddress(r, f, body)
	}

	// re-initialize the UI with a "clean" new state
	f.UI = &container.Container{
		Method: "POST",
		Action: flow.AppendFlowTo(urlx.AppendPaths(s.deps.Config().SelfPublicURL(r.Context()), recovery.RouteSubmitFlow), f.ID).String(),
	}

	f.UI.SetCSRF(f.CSRFToken)

	f.State = flow.StateRecoveryAwaitingAddressChoice
	f.UI.Messages.Set(text.NewRecoveryAskToChooseAddress())

	slices.SortFunc(recoveryAddresses, func(a, b identity.RecoveryAddress) int {
		return strings.Compare(a.Value, b.Value)
	})

	for _, a := range recoveryAddresses {
		// NOTE: Only send the masked value and the hash, to avoid information exfiltration.
		// Why the hash? So that we can recognize later, when the user chooses the masked address in the list,
		// that the chosen masked address is the `recovery_address` provided in the beginning,
		// and then we do not ask again the user to provide it in full.
		hashBase64 := AddressToHashBase64(a.Value)
		f.UI.GetNodes().Append(node.NewInputField("recovery_select_address", hashBase64, node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(&text.Message{
				ID:   text.InfoNodeLabel,
				Text: MaskAddress(a.Value),
				Type: text.Info,
			}))
	}

	f.UI.Nodes.Append(node.NewInputField("method", s.NodeGroup(), node.CodeGroup, node.InputAttributeTypeHidden))
	f.UI.Nodes.Append(node.NewInputField("recovery_address", body.RecoveryAddress, node.CodeGroup, node.InputAttributeTypeHidden))
	// No back button here because there is no point for the user.

	if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) recoveryV2HandleStateAwaitingAddressChoice(r *http.Request, f *recovery.Flow, body *recoverySubmitPayload) error {
	if f.State != flow.StateRecoveryAwaitingAddressChoice {
		return errors.Errorf("unreachable state: %s", f.State)
	}

	if len(body.RecoverySelectAddress) == 0 {
		return schema.NewRequiredError("#/recovery_select_address", "recovery_select_address")
	}

	if len(body.RecoveryAddress) == 0 {
		return schema.NewRequiredError("#/recovery_address", "recovery_address")
	}

	// Is the chosen masked address the same as the address provided in full at the beginning?
	// If yes, then do not ask it again in full.
	// Technically we check `hash(recovery_address) == recovery_select_address` and
	// `recovery_select_address` is `hash(recovery_address)`.
	hashBase64 := AddressToHashBase64(body.RecoveryAddress)

	// Better safe than sorry, use constant time comparison.
	if subtle.ConstantTimeCompare([]byte(hashBase64), []byte(body.RecoverySelectAddress)) == 1 {
		// Skip a state: do not ask the user again to provide the full address.
		body.RecoveryConfirmAddress = body.RecoveryAddress
		f.State = flow.StateRecoveryAwaitingAddressConfirm
		return s.recoveryV2HandleStateConfirmingAddress(r, f, body)
	}

	// re-initialize the UI with a "clean" new state
	f.UI = &container.Container{
		Method: "POST",
		Action: flow.AppendFlowTo(urlx.AppendPaths(s.deps.Config().SelfPublicURL(r.Context()), recovery.RouteSubmitFlow), f.ID).String(),
	}
	f.UI.SetCSRF(f.CSRFToken)

	f.State = flow.StateRecoveryAwaitingAddressConfirm
	f.UI.Messages.Set(text.NewRecoveryAskForFullAddress())

	// Retrieve the selected recovery address in plaintext to determine the input label and type.
	var plaintextRecoveryAddress string
	recoveryAddresses, err := s.deps.IdentityPool().FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(r.Context(), body.RecoveryAddress)
	if err == nil {
		for _, a := range recoveryAddresses {
			if subtle.ConstantTimeCompare([]byte(AddressToHashBase64(a.Value)), []byte(body.RecoverySelectAddress)) == 1 {
				plaintextRecoveryAddress = a.Value
				break
			}
		}
	}
	if plaintextRecoveryAddress == "" {
		return herodot.ErrBadRequest.
			WithReason("The selected recovery address is not valid.").
			WithDebug("The selected recovery address does not match any of the known recovery addresses.")
	}

	var inputType node.UiNodeInputAttributeType
	var label *text.Message
	if strings.ContainsRune(plaintextRecoveryAddress, '@') {
		inputType = node.InputAttributeTypeEmail
		label = text.NewInfoNodeInputEmail()
	} else {
		inputType = node.InputAttributeTypeTel
		label = text.NewInfoNodeInputPhoneNumber()
	}

	f.UI.Nodes.Append(node.NewInputField("recovery_confirm_address", body.RecoveryConfirmAddress, node.CodeGroup, inputType, node.WithRequiredInputAttribute).
		WithMetaLabel(label),
	)
	f.UI.Nodes.Append(node.NewInputField("recovery_address", body.RecoveryAddress, node.CodeGroup, node.InputAttributeTypeHidden))

	f.UI.
		GetNodes().
		Append(node.NewInputField("method", s.RecoveryStrategyID(), node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(text.NewInfoNodeLabelContinue()))

	f.UI.Nodes.Append(node.NewInputField("method", s.NodeGroup(), node.CodeGroup, node.InputAttributeTypeHidden))

	buttonScreen := node.NewInputField("screen", "previous", node.CodeGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewRecoveryBack())
	f.UI.GetNodes().Append(buttonScreen)

	if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) recoveryV2HandleStateConfirmingAddress(r *http.Request, f *recovery.Flow, body *recoverySubmitPayload) error {
	if f.State != flow.StateRecoveryAwaitingAddressConfirm {
		return errors.Errorf("unreachable state: %s", f.State)
	}

	if len(body.RecoveryConfirmAddress) == 0 {
		return schema.NewRequiredError("#/recovery_confirm_address", "recovery_confirm_address")
	}

	if err := s.deps.RecoveryCodePersister().DeleteRecoveryCodesOfFlow(r.Context(), f.ID); err != nil {
		return err
	}

	f.TransientPayload = body.TransientPayload

	var addressType identity.RecoveryAddressType
	// Inferring the address type like this is a bit hacky, and actually not really necessary.
	// That's because `SendRecoveryCode` expects it, but not because it fundamentally is required.
	if strings.ContainsRune(body.RecoveryConfirmAddress, '@') {
		addressType = identity.RecoveryAddressTypeEmail
	} else {
		addressType = identity.RecoveryAddressTypeSMS
	}

	// NOTE: We do not fetch the db address here. We only (try to) send the code to the user provided address.
	// That way we avoid information exfiltration.
	// `SendRecoveryCode` will anyway check by itself if the provided address is a known address or not.
	if err := s.deps.CodeSender().SendRecoveryCode(r.Context(), f, addressType, body.RecoveryConfirmAddress); err != nil {
		if !errors.Is(err, ErrUnknownAddress) {
			return err
		}

		// Continue execution
	}

	// re-initialize the UI with a "clean" new state
	f.UI = &container.Container{
		Method: "POST",
		Action: flow.AppendFlowTo(urlx.AppendPaths(s.deps.Config().SelfPublicURL(r.Context()), recovery.RouteSubmitFlow), f.ID).String(),
	}
	f.UI.SetCSRF(f.CSRFToken)

	f.State = flow.StateRecoveryAwaitingCode

	uiText := text.NewRecoveryCodeRecoverySelectAddressSent(MaskAddress(body.RecoveryConfirmAddress))

	f.UI.Messages.Set(uiText)
	f.UI.Nodes.Append(node.NewInputField("code", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithInputAttributes(func(a *node.InputAttributes) {
		a.Required = true
		a.Pattern = "[0-9]+"
		a.MaxLength = CodeLength
	})).
		WithMetaLabel(text.NewInfoNodeLabelRecoveryCode()),
	)

	f.UI.Nodes.Append(node.NewInputField("method", s.NodeGroup(), node.CodeGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoNodeLabelContinue()),
	)

	f.UI.Nodes.Append(node.NewInputField("method", s.NodeGroup(), node.CodeGroup, node.InputAttributeTypeHidden))

	// Required to make 'resend' work.
	f.UI.Nodes.Append(node.NewInputField("recovery_confirm_address", body.RecoveryConfirmAddress, node.CodeGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoNodeResendOTP()),
	)
	f.UI.Nodes.Append(node.NewInputField("recovery_address", body.RecoveryAddress, node.CodeGroup, node.InputAttributeTypeHidden))

	buttonScreen := node.NewInputField("screen", "previous", node.CodeGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewRecoveryBack())
	f.UI.GetNodes().Append(buttonScreen)

	if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) recoveryV2HandleStateAwaitingCode(w http.ResponseWriter, r *http.Request, f *recovery.Flow, body *recoverySubmitPayload) error {
	if f.State != flow.StateRecoveryAwaitingCode {
		return errors.Errorf("unreachable state: %s", f.State)
	}

	if len(body.Code) == 0 {
		// The 're-send' button was clicked. We handle it as if the user first arrived at the state `RecoveryV2StateAwaitingAddressConfirm`.
		// That will invalidate all existing codes and send a new code.
		f.State = flow.StateRecoveryAwaitingAddressConfirm
		return s.recoveryV2HandleStateConfirmingAddress(r, f, body)
	} else {
		return s.recoveryUseCode(w, r, body, f)
	}
}

func (s *Strategy) recoveryV2HandleGoBack(r *http.Request, f *recovery.Flow, body *recoverySubmitPayload) error {
	// If no address choice needs to take place, just go to the first screen.
	recoveryAddresses, _ := s.deps.IdentityPool().FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(r.Context(), body.RecoveryAddress)
	if len(recoveryAddresses) <= 1 {
		f.State = flow.StateRecoveryAwaitingAddress
		err := s.PopulateRecoveryMethod(r, f)
		if err != nil {
			return err
		}

		if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
			return err
		}

	}

	switch f.State {
	// Go back to the second screen (choose an address) by essentially going to the first screen
	// and re-submitting the form (to arrive at the second screen).
	// This contraption is necessary since the UI nodes are stored in the database and not generated on the fly.
	// So simply redirecting to a previous screen (as in: 'web page') would do nothing, it would just show the same UI.
	// This way we force the UI generation code to re-run and the new UI nodes to be stored to the database.
	case flow.StateRecoveryAwaitingCode:
		fallthrough
	case flow.StateRecoveryAwaitingAddressConfirm:
		// Reset some body fields since we are going to (almost) the beginning of the flow.
		body.RecoveryConfirmAddress = ""
		body.RecoverySelectAddress = ""
		body.Screen = ""

		f.State = flow.StateRecoveryAwaitingAddress

		return s.recoveryV2HandleStateAwaitingAddress(r, f, body)
	default:
		// Should not trigger, but do something sensible: start from scratch.
		return s.PopulateRecoveryMethod(r, f)
	}
}

// recoveryHandleFormSubmission handles the submission of an address for recovery
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
	if err := s.deps.CodeSender().SendRecoveryCode(ctx, f, identity.RecoveryAddressTypeEmail, body.Email); err != nil {
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
			if err := s.deps.PrivilegedIdentityPool().UpdateVerifiableAddress(r.Context(), &id.VerifiableAddresses[k], "verified", "verified_at", "status"); err != nil {
				return s.HandleRecoveryError(w, r, f, nil, err)
			}
		}
	}

	return nil
}

func (s *Strategy) HandleRecoveryError(w http.ResponseWriter, r *http.Request, fl *recovery.Flow, body *recoverySubmitPayload, err error) error {
	if fl != nil {
		if flow.IsStateRecoveryV2(fl.State) {
			// Unreachable: RecoveryV2 never uses this function.
			return err
		}
		email := ""
		if body != nil {
			email = body.Email
		}

		fl.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		fl.UI.GetNodes().Upsert(
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

	// Used in RecoveryV2.
	RecoveryAddress        string `json:"recovery_address" form:"recovery_address"`
	RecoverySelectAddress  string `json:"recovery_select_address" form:"recovery_select_address"`
	RecoveryConfirmAddress string `json:"recovery_confirm_address" form:"recovery_confirm_address"`
	Screen                 string `json:"screen" form:"screen"`
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
