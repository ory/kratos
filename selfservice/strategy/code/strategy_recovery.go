// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

const (
	RouteAdminCreateRecoveryCode = "/recovery/code"
)

func (s *Strategy) RecoveryStrategyID() string {
	return string(recovery.RecoveryStrategyCode)
}

func (s *Strategy) RegisterPublicRecoveryRoutes(public *x.RouterPublic) {
	s.deps.CSRFHandler().IgnorePath(RouteAdminCreateRecoveryCode)
	public.POST(RouteAdminCreateRecoveryCode, x.RedirectToAdminRoute(s.deps))
}

func (s *Strategy) RegisterAdminRecoveryRoutes(admin *x.RouterAdmin) {
	wrappedCreateRecoveryCode := strategy.IsDisabled(s.deps, s.RecoveryStrategyID(), s.createRecoveryCodeForIdentity)
	admin.POST(RouteAdminCreateRecoveryCode, wrappedCreateRecoveryCode)
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
			WithMetaLabel(text.NewInfoNodeLabelSubmit()))

	return nil
}

// Create Recovery Code for Identity Parameters
//
// swagger:parameters createRecoveryCodeForIdentity
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createRecoveryCodeForIdentity struct {
	// in: body
	Body createRecoveryCodeForIdentityBody
}

// Create Recovery Code for Identity Request Body
//
// swagger:model createRecoveryCodeForIdentityBody
type createRecoveryCodeForIdentityBody struct {
	// Identity to Recover
	//
	// The identity's ID you wish to recover.
	//
	// required: true
	IdentityID uuid.UUID `json:"identity_id"`

	// Code Expires In
	//
	// The recovery code will expire after that amount of time has passed. Defaults to the configuration value of
	// `selfservice.methods.code.config.lifespan`.
	//
	//
	// pattern: ^([0-9]+(ns|us|ms|s|m|h))*$
	// example:
	//	- 1h
	//	- 1m
	//	- 1s
	ExpiresIn string `json:"expires_in"`
}

// Recovery Code for Identity
//
// Used when an administrator creates a recovery code for an identity.
//
// swagger:model recoveryCodeForIdentity
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type recoveryCodeForIdentity struct {
	// RecoveryLink with flow
	//
	// This link opens the recovery UI with an empty `code` field.
	//
	// required: true
	// format: uri
	RecoveryLink string `json:"recovery_link"`

	// RecoveryCode is the code that can be used to recover the account
	//
	// required: true
	RecoveryCode string `json:"recovery_code"`

	// Expires At is the timestamp of when the recovery flow expires
	//
	// The timestamp when the recovery link expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// swagger:route POST /admin/recovery/code identity createRecoveryCodeForIdentity
//
// # Create a Recovery Code
//
// This endpoint creates a recovery code which should be given to the user in order for them to recover
// (or activate) their account.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Security:
//		oryAccessToken:
//
//	Responses:
//		201: recoveryCodeForIdentity
//		400: errorGeneric
//		404: errorGeneric
//		default: errorGeneric
func (s *Strategy) createRecoveryCodeForIdentity(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p createRecoveryCodeForIdentityBody
	if err := s.dx.Decode(r, &p, decoderx.HTTPJSONDecoder()); err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	ctx := r.Context()
	config := s.deps.Config()

	expiresIn := config.SelfServiceCodeMethodLifespan(ctx)
	if len(p.ExpiresIn) > 0 {
		// If an expiration of the code was supplied use it instead of the default duration
		var err error
		expiresIn, err = time.ParseDuration(p.ExpiresIn)
		if err != nil {
			s.deps.Writer().WriteError(w, r, errors.WithStack(herodot.
				ErrBadRequest.
				WithReasonf(`Unable to parse "expires_in" whose format should match "[0-9]+(ns|us|ms|s|m|h)" but did not: %s`, p.ExpiresIn)))
			return
		}
	}

	if time.Now().Add(expiresIn).Before(time.Now()) {
		s.deps.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Value from "expires_in" must result to a future time: %s`, p.ExpiresIn)))
		return
	}

	flow, err := recovery.NewFlow(config, expiresIn, s.deps.GenerateCSRFToken(r), r, s, flow.TypeBrowser)
	if err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}
	flow.DangerousSkipCSRFCheck = true
	flow.State = recovery.StateEmailSent
	flow.UI.Nodes = node.Nodes{}
	flow.UI.Nodes.Append(node.NewInputField("code", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).
		WithMetaLabel(text.NewInfoNodeLabelRecoveryCode()),
	)

	flow.UI.Nodes.
		Append(node.NewInputField("method", s.RecoveryStrategyID(), node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(text.NewInfoNodeLabelSubmit()))

	if err := s.deps.RecoveryFlowPersister().CreateRecoveryFlow(ctx, flow); err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	id, err := s.deps.IdentityPool().GetIdentity(ctx, p.IdentityID, identity.ExpandDefault)
	if notFoundErr := sqlcon.ErrNoRows; errors.As(err, &notFoundErr) {
		s.deps.Writer().WriteError(w, r, notFoundErr.WithReasonf("could not find identity"))
		return
	} else if err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	rawCode := GenerateCode()

	if _, err := s.deps.RecoveryCodePersister().CreateRecoveryCode(ctx, &CreateRecoveryCodeParams{
		RawCode:    rawCode,
		CodeType:   RecoveryCodeTypeAdmin,
		ExpiresIn:  expiresIn,
		FlowID:     flow.ID,
		IdentityID: id.ID,
	}); err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	s.deps.Audit().
		WithField("identity_id", id.ID).
		WithSensitiveField("recovery_code", rawCode).
		Info("A recovery code has been created.")

	body := &recoveryCodeForIdentity{
		ExpiresAt: flow.ExpiresAt.UTC(),
		RecoveryLink: urlx.CopyWithQuery(
			s.deps.Config().SelfServiceFlowRecoveryUI(ctx),
			url.Values{
				"flow": {flow.ID.String()},
			}).String(),
		RecoveryCode: rawCode,
	}

	s.deps.Writer().WriteCode(w, r, http.StatusCreated, body, herodot.UnescapedHTML)
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
}

func (s Strategy) isCodeFlow(f *recovery.Flow) bool {
	value, err := f.Active.Value()
	if err != nil {
		return false
	}
	return value == s.RecoveryNodeGroup().String()
}

func (s *Strategy) Recover(w http.ResponseWriter, r *http.Request, f *recovery.Flow) (err error) {
	if !s.isCodeFlow(f) {
		return errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	body, err := s.decodeRecovery(r)
	if err != nil {
		return s.HandleRecoveryError(w, r, nil, body, err)
	}
	ctx := r.Context()

	if f.DangerousSkipCSRFCheck {
		s.deps.Logger().
			WithRequest(r).
			Debugf("A recovery flow with `DangerousSkipCSRFCheck` set has been submitted, skipping anti-CSRF measures.")
	} else if err := flow.EnsureCSRF(s.deps, r, f.Type, s.deps.Config().DisableAPIFlowEnforcement(ctx), s.deps.GenerateCSRFToken, body.CSRFToken); err != nil {
		// If a CSRF violation occurs the flow is most likely FUBAR, as the user either lost the CSRF token, or an attack occured.
		// In this case, we just issue a new flow and "abandon" the old flow.
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sID := s.RecoveryStrategyID()

	f.UI.ResetMessages()

	// If the email is present in the submission body, the user needs a new code via resend
	if f.State != recovery.StateChooseMethod && len(body.Email) == 0 {
		if err := flow.MethodEnabledAndAllowed(ctx, sID, sID, s.deps); err != nil {
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

	if err := flow.MethodEnabledAndAllowed(ctx, sID, body.Method, s.deps); err != nil {
		return s.HandleRecoveryError(w, r, nil, body, err)
	}

	flow, err := s.deps.RecoveryFlowPersister().GetRecoveryFlow(ctx, x.ParseUUID(body.Flow))
	if err != nil {
		return s.HandleRecoveryError(w, r, flow, body, err)
	}

	if err := flow.Valid(); err != nil {
		return s.HandleRecoveryError(w, r, flow, body, err)
	}

	switch flow.State {
	case recovery.StateChooseMethod:
		fallthrough
	case recovery.StateEmailSent:
		return s.recoveryHandleFormSubmission(w, r, flow, body)
	case recovery.StatePassedChallenge:
		// was already handled, do not allow retry
		return s.retryRecoveryFlowWithMessage(w, r, flow.Type, text.NewErrorValidationRecoveryRetrySuccess())
	default:
		return s.retryRecoveryFlowWithMessage(w, r, flow.Type, text.NewErrorValidationRecoveryStateFailure())
	}
}

func (s *Strategy) recoveryIssueSession(w http.ResponseWriter, r *http.Request, f *recovery.Flow, id *identity.Identity) error {
	ctx := r.Context()

	f.UI.Messages.Clear()
	f.State = recovery.StatePassedChallenge
	f.SetCSRFToken(s.deps.CSRFHandler().RegenerateToken(w, r))
	f.RecoveredIdentityID = uuid.NullUUID{
		UUID:  id.ID,
		Valid: true,
	}
	if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(ctx, f); err != nil {
		return s.retryRecoveryFlowWithError(w, r, f.Type, err)
	}

	sess, err := session.NewActiveSession(r, id, s.deps.Config(), time.Now().UTC(),
		identity.CredentialsTypeRecoveryCode, identity.AuthenticatorAssuranceLevel1)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, f.Type, err)
	}

	// TODO: How does this work with Mobile?
	if err := s.deps.SessionManager().UpsertAndIssueCookie(ctx, w, r, sess); err != nil {
		return s.retryRecoveryFlowWithError(w, r, f.Type, err)
	}

	sf, err := s.deps.SettingsHandler().NewFlow(w, r, sess.Identity, f.Type)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, f.Type, err)
	}

	returnToURL := s.deps.Config().SelfServiceFlowRecoveryReturnTo(r.Context(), nil)
	returnTo := ""
	if returnToURL != nil {
		returnTo = returnToURL.String()
	}
	sf.RequestURL, err = x.TakeOverReturnToParameter(f.RequestURL, sf.RequestURL, returnTo)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	if err := s.deps.RecoveryExecutor().PostRecoveryHook(w, r, f, sess); err != nil {
		return s.retryRecoveryFlowWithError(w, r, f.Type, err)
	}

	config := s.deps.Config()

	sf.UI.Messages.Set(text.NewRecoverySuccessful(time.Now().Add(config.SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx))))
	if err := s.deps.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), sf); err != nil {
		return s.retryRecoveryFlowWithError(w, r, f.Type, err)
	}

	if x.IsJSONRequest(r) {
		s.deps.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(sf.AppendTo(s.deps.Config().SelfServiceFlowSettingsUI(r.Context())).String()))
	} else {
		http.Redirect(w, r, sf.AppendTo(s.deps.Config().SelfServiceFlowSettingsUI(r.Context())).String(), http.StatusSeeOther)
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
			return s.retryRecoveryFlowWithError(w, r, f.Type, err)
		}

		// No error
		return nil
	} else if err != nil {
		return s.retryRecoveryFlowWithError(w, r, f.Type, err)
	}

	recovered, err := s.deps.IdentityPool().GetIdentity(ctx, code.IdentityID, identity.ExpandDefault)
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

func (s *Strategy) retryRecoveryFlowWithMessage(w http.ResponseWriter, r *http.Request, ft flow.Type, message *text.Message) error {
	s.deps.Logger().
		WithRequest(r).
		WithField("message", message).
		Debug("A recovery flow is being retried because a validation error occurred.")

	ctx := r.Context()
	config := s.deps.Config()

	f, err := recovery.NewFlow(config, config.SelfServiceFlowRecoveryRequestLifespan(ctx), s.deps.CSRFHandler().RegenerateToken(w, r), r, s, ft)
	if err != nil {
		return err
	}

	f.UI.Messages.Add(message)
	if err := s.deps.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f); err != nil {
		return err
	}

	if x.IsJSONRequest(r) {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(config.SelfPublicURL(ctx),
			recovery.RouteGetFlow), url.Values{"id": {f.ID.String()}}).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, f.AppendTo(config.SelfServiceFlowRecoveryUI(ctx)).String(), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryRecoveryFlowWithError(w http.ResponseWriter, r *http.Request, ft flow.Type, recErr error) error {
	s.deps.Logger().
		WithRequest(r).
		WithError(recErr).
		Debug("A recovery flow is being retried because a validation error occurred.")

	ctx := r.Context()
	config := s.deps.Config()

	if expired := new(flow.ExpiredError); errors.As(recErr, &expired) {
		return s.retryRecoveryFlowWithMessage(w, r, ft, text.NewErrorValidationRecoveryFlowExpired(expired.ExpiredAt))
	}

	f, err := recovery.NewFlow(config, config.SelfServiceFlowRecoveryRequestLifespan(ctx), s.deps.CSRFHandler().RegenerateToken(w, r), r, s, ft)
	if err != nil {
		return err
	}
	if err := f.UI.ParseError(node.CodeGroup, recErr); err != nil {
		return err
	}
	if err := s.deps.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f); err != nil {
		return err
	}

	if x.IsJSONRequest(r) {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(config.SelfPublicURL(ctx),
			recovery.RouteGetFlow), url.Values{"id": {f.ID.String()}}).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, f.AppendTo(config.SelfServiceFlowRecoveryUI(ctx)).String(), http.StatusSeeOther)
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

	f.Active = sqlxx.NullString(s.RecoveryNodeGroup())
	f.State = recovery.StateEmailSent
	f.UI.Messages.Set(text.NewRecoveryEmailWithCodeSent())
	f.UI.Nodes.Append(node.NewInputField("code", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).
		WithMetaLabel(text.NewInfoNodeLabelRecoveryCode()),
	)
	f.UI.Nodes.Append(node.NewInputField("method", s.RecoveryNodeGroup(), node.CodeGroup, node.InputAttributeTypeHidden))

	f.UI.
		GetNodes().
		Append(node.NewInputField("method", s.RecoveryStrategyID(), node.CodeGroup, node.InputAttributeTypeSubmit).
			WithMetaLabel(text.NewInfoNodeLabelSubmit()))

	f.UI.Nodes.Append(node.NewInputField("email", body.Email, node.CodeGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoNodeResendOTP()),
	)
	if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return s.HandleRecoveryError(w, r, f, body, err)
	}

	return nil
}

func (s *Strategy) markRecoveryAddressVerified(w http.ResponseWriter, r *http.Request, f *recovery.Flow, id *identity.Identity, recoveryAddress *identity.RecoveryAddress) error {
	var address *identity.VerifiableAddress
	for idx := range id.VerifiableAddresses {
		va := id.VerifiableAddresses[idx]
		if va.Value == recoveryAddress.Value {
			address = &va
			break
		}
	}

	if address != nil && !address.Verified { // can it be that the address is nil?
		address.Verified = true
		verifiedAt := sqlxx.NullTime(time.Now().UTC())
		address.VerifiedAt = &verifiedAt
		address.Status = identity.VerifiableAddressStatusCompleted
		if err := s.deps.PrivilegedIdentityPool().UpdateVerifiableAddress(r.Context(), address); err != nil {
			return s.HandleRecoveryError(w, r, f, nil, err)
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
	Method    string `json:"method" form:"method"`
	Code      string `json:"code" form:"code"`
	CSRFToken string `json:"csrf_token" form:"csrf_token"`
	Flow      string `json:"flow" form:"flow"`
	Email     string `json:"email" form:"email"`
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
