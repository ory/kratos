// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link

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
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

const (
	RouteAdminCreateRecoveryLink = "/recovery/link"
)

func (s *Strategy) RecoveryStrategyID() string {
	return recovery.StrategyRecoveryLinkName
}

func (s *Strategy) RegisterPublicRecoveryRoutes(public *x.RouterPublic) {
	s.d.CSRFHandler().IgnorePath(RouteAdminCreateRecoveryLink)
	public.POST(RouteAdminCreateRecoveryLink, x.RedirectToAdminRoute(s.d))

}

func (s *Strategy) RegisterAdminRecoveryRoutes(admin *x.RouterAdmin) {
	wrappedCreateRecoveryLink := strategy.IsDisabled(s.d, s.RecoveryStrategyID(), s.createRecoveryLinkForIdentity)
	admin.POST(RouteAdminCreateRecoveryLink, wrappedCreateRecoveryLink)
}

func (s *Strategy) PopulateRecoveryMethod(r *http.Request, f *recovery.Flow) error {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true},
		node.NewInputField("email", nil, node.LinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputEmail()),
	)
	f.UI.GetNodes().Append(node.NewInputField("method", s.RecoveryStrategyID(), node.LinkGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSubmit()))

	return nil
}

// Create Recovery Link for Identity Parameters
//
// swagger:parameters createRecoveryLinkForIdentity
//
// nolint
type createRecoveryLinkForIdentity struct {
	// in: body
	Body createRecoveryLinkForIdentityBody
}

// Create Recovery Link for Identity Request Body
//
// swagger:model createRecoveryLinkForIdentityBody
type createRecoveryLinkForIdentityBody struct {
	// Identity to Recover
	//
	// The identity's ID you wish to recover.
	//
	// required: true
	IdentityID uuid.UUID `json:"identity_id"`

	// Link Expires In
	//
	// The recovery link will expire after that amount of time has passed. Defaults to the configuration value of
	// `selfservice.methods.code.config.lifespan`.
	//
	//
	// pattern: ^[0-9]+(ns|us|ms|s|m|h)$
	// example:
	//	- 1h
	//	- 1m
	//	- 1s
	ExpiresIn string `json:"expires_in"`
}

// Identity Recovery Link
//
// Used when an administrator creates a recovery link for an identity.
//
// swagger:model recoveryLinkForIdentity
// nolint
type recoveryLinkForIdentity struct {
	// Recovery Link
	//
	// This link can be used to recover the account.
	//
	// required: true
	// format: uri
	RecoveryLink string `json:"recovery_link"`

	// Recovery Link Expires At
	//
	// The timestamp when the recovery link expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// swagger:route POST /admin/recovery/link identity createRecoveryLinkForIdentity
//
// # Create a Recovery Link
//
// This endpoint creates a recovery link which should be given to the user in order for them to recover
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
//	  oryAccessToken:
//
//	Responses:
//	  200: recoveryLinkForIdentity
//	  400: errorGeneric
//	  404: errorGeneric
//	  default: errorGeneric
func (s *Strategy) createRecoveryLinkForIdentity(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p createRecoveryLinkForIdentityBody
	if err := s.dx.Decode(r, &p, decoderx.HTTPJSONDecoder()); err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	expiresIn := s.d.Config().SelfServiceLinkMethodLifespan(r.Context())
	if len(p.ExpiresIn) > 0 {
		var err error
		expiresIn, err = time.ParseDuration(p.ExpiresIn)
		if err != nil {
			s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to parse "expires_in" whose format should match "[0-9]+(ns|us|ms|s|m|h)" but did not: %s`, p.ExpiresIn)))
			return
		}
	}

	if time.Now().Add(expiresIn).Before(time.Now()) {
		s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Value from "expires_in" must be result to a future time: %s`, p.ExpiresIn)))
		return
	}

	req, err := recovery.NewFlow(s.d.Config(), expiresIn, s.d.GenerateCSRFToken(r), r, s, flow.TypeBrowser)
	if err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	id, err := s.d.IdentityPool().GetIdentity(r.Context(), p.IdentityID)
	if errors.Is(err, sqlcon.ErrNoRows) {
		s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The requested identity id does not exist.").WithWrap(err)))
		return
	} else if err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	token := NewAdminRecoveryToken(id.ID, req.ID, expiresIn)
	if err := s.d.RecoveryTokenPersister().CreateRecoveryToken(r.Context(), token); err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	s.d.Audit().
		WithField("identity_id", id.ID).
		WithSensitiveField("recovery_link_token", token).
		Info("A recovery link has been created.")

	s.d.Writer().Write(w, r, &recoveryLinkForIdentity{
		ExpiresAt: req.ExpiresAt.UTC(),
		RecoveryLink: urlx.CopyWithQuery(
			urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()), recovery.RouteSubmitFlow),
			url.Values{
				"token": {token.Token},
				"flow":  {req.ID.String()},
			}).String()},
		herodot.UnescapedHTML)
}

// Update Recovery Flow with Link Method
//
// swagger:model updateRecoveryFlowWithLinkMethod
// nolint:deadcode,unused
type updateRecoveryFlowWithLinkMethod struct {
	// Email to Recover
	//
	// Needs to be set when initiating the flow. If the email is a registered
	// recovery email, a recovery link will be sent. If the email is not known,
	// a email with details on what happened will be sent instead.
	//
	// format: email
	// required: true
	Email string `json:"email" form:"email"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `form:"csrf_token" json:"csrf_token"`

	// Method supports `link` only right now.
	//
	// required: true
	Method string `json:"method"`
}

func (s *Strategy) Recover(w http.ResponseWriter, r *http.Request, f *recovery.Flow) (err error) {
	body, err := s.decodeRecovery(r)
	if err != nil {
		return s.HandleRecoveryError(w, r, nil, body, err)
	}

	if len(body.Token) > 0 {
		if err := flow.MethodEnabledAndAllowed(r.Context(), s.RecoveryStrategyID(), s.RecoveryStrategyID(), s.d); err != nil {
			return s.HandleRecoveryError(w, r, nil, body, err)
		}

		return s.recoveryUseToken(w, r, f.ID, body)
	}

	if _, err := s.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
		if x.IsJSONRequest(r) {
			session.RespondWithJSONErrorOnAuthenticated(s.d.Writer(), recovery.ErrAlreadyLoggedIn)(w, r, nil)
		} else {
			session.RedirectOnAuthenticated(s.d)(w, r, nil)
		}
		return errors.WithStack(flow.ErrCompletedByStrategy)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), s.RecoveryStrategyID(), body.Method, s.d); err != nil {
		return s.HandleRecoveryError(w, r, nil, body, err)
	}

	req, err := s.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), x.ParseUUID(body.Flow))
	if err != nil {
		return s.HandleRecoveryError(w, r, req, body, err)
	}

	if err := req.Valid(); err != nil {
		return s.HandleRecoveryError(w, r, req, body, err)
	}

	switch req.State {
	case recovery.StateChooseMethod:
		fallthrough
	case recovery.StateEmailSent:
		return s.recoveryHandleFormSubmission(w, r, req)
	case recovery.StatePassedChallenge:
		// was already handled, do not allow retry
		return s.retryRecoveryFlowWithMessage(w, r, req.Type, text.NewErrorValidationRecoveryRetrySuccess())
	default:
		return s.retryRecoveryFlowWithMessage(w, r, req.Type, text.NewErrorValidationRecoveryStateFailure())
	}
}

func (s *Strategy) recoveryIssueSession(w http.ResponseWriter, r *http.Request, f *recovery.Flow, id *identity.Identity) error {
	f.UI.Messages.Clear()
	f.State = recovery.StatePassedChallenge
	f.SetCSRFToken(s.d.CSRFHandler().RegenerateToken(w, r))
	f.RecoveredIdentityID = uuid.NullUUID{
		UUID:  id.ID,
		Valid: true,
	}
	if err := s.d.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sess, err := session.NewActiveSession(r, id, s.d.Config(), time.Now().UTC(), identity.CredentialsTypeRecoveryLink, identity.AuthenticatorAssuranceLevel1)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	if err := s.d.SessionManager().UpsertAndIssueCookie(r.Context(), w, r, sess); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sf, err := s.d.SettingsHandler().NewFlow(w, r, sess.Identity, flow.TypeBrowser)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sf.RequestURL, err = x.TakeOverReturnToParameter(f.RequestURL, sf.RequestURL)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	if err := s.d.RecoveryExecutor().PostRecoveryHook(w, r, f, sess); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sf.UI.Messages.Set(text.NewRecoverySuccessful(time.Now().Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context()))))
	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), sf); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	http.Redirect(w, r, sf.AppendTo(s.d.Config().SelfServiceFlowSettingsUI(r.Context())).String(), http.StatusSeeOther)
	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) recoveryUseToken(w http.ResponseWriter, r *http.Request, fID uuid.UUID, body *recoverySubmitPayload) error {
	token, err := s.d.RecoveryTokenPersister().UseRecoveryToken(r.Context(), fID, body.Token)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			return s.retryRecoveryFlowWithMessage(w, r, flow.TypeBrowser, text.NewErrorValidationRecoveryTokenInvalidOrAlreadyUsed())
		}

		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	var f *recovery.Flow
	if !token.FlowID.Valid {
		f, err = recovery.NewFlow(s.d.Config(), time.Until(token.ExpiresAt), s.d.GenerateCSRFToken(r), r, s, flow.TypeBrowser)
		if err != nil {
			return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
		}

		if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), f); err != nil {
			return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
		}
	} else {
		f, err = s.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), token.FlowID.UUID)
		if err != nil {
			return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
		}
	}

	if err := token.Valid(); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	recovered, err := s.d.IdentityPool().GetIdentity(r.Context(), token.IdentityID)
	if err != nil {
		return s.HandleRecoveryError(w, r, f, nil, err)
	}

	// mark address as verified only for a self-service flow
	if token.TokenType == RecoveryTokenTypeSelfService {
		if err := s.markRecoveryAddressVerified(w, r, f, recovered, token.RecoveryAddress); err != nil {
			return s.HandleRecoveryError(w, r, f, body, err)
		}
	}

	return s.recoveryIssueSession(w, r, f, recovered)
}

func (s *Strategy) retryRecoveryFlowWithMessage(w http.ResponseWriter, r *http.Request, ft flow.Type, message *text.Message) error {
	s.d.Logger().WithRequest(r).WithField("message", message).Debug("A recovery flow is being retried because a validation error occurred.")

	req, err := recovery.NewFlow(s.d.Config(), s.d.Config().SelfServiceFlowRecoveryRequestLifespan(r.Context()), s.d.CSRFHandler().RegenerateToken(w, r), r, s, ft)
	if err != nil {
		return err
	}

	req.UI.Messages.Add(message)
	if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		return err
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, req.AppendTo(s.d.Config().SelfServiceFlowRecoveryUI(r.Context())).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()),
			recovery.RouteGetFlow), url.Values{"id": {req.ID.String()}}).String(), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryRecoveryFlowWithError(w http.ResponseWriter, r *http.Request, ft flow.Type, recErr error) error {
	s.d.Logger().WithRequest(r).WithError(recErr).Debug("A recovery flow is being retried because a validation error occurred.")

	req, err := recovery.NewFlow(s.d.Config(), s.d.Config().SelfServiceFlowRecoveryRequestLifespan(r.Context()), s.d.CSRFHandler().RegenerateToken(w, r), r, s, ft)
	if err != nil {
		return err
	}

	if expired := new(flow.ExpiredError); errors.As(recErr, &expired) {
		return s.retryRecoveryFlowWithMessage(w, r, ft, text.NewErrorValidationRecoveryFlowExpired(expired.ExpiredAt))
	} else {
		if err := req.UI.ParseError(node.LinkGroup, recErr); err != nil {
			return err
		}
	}

	if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		return err
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, req.AppendTo(s.d.Config().SelfServiceFlowRecoveryUI(r.Context())).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()),
			recovery.RouteGetFlow), url.Values{"id": {req.ID.String()}}).String(), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) recoveryHandleFormSubmission(w http.ResponseWriter, r *http.Request, f *recovery.Flow) error {
	body, err := s.decodeRecovery(r)
	if err != nil {
		return s.HandleRecoveryError(w, r, f, body, err)
	}

	if len(body.Email) == 0 {
		return s.HandleRecoveryError(w, r, f, body, schema.NewRequiredError("#/email", "email"))
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, body.CSRFToken); err != nil {
		return s.HandleRecoveryError(w, r, f, body, err)
	}

	if err := s.d.LinkSender().SendRecoveryLink(r.Context(), r, f, identity.VerifiableAddressTypeEmail, body.Email); err != nil {
		if !errors.Is(err, ErrUnknownAddress) {
			return s.HandleRecoveryError(w, r, f, body, err)
		}
		// Continue execution
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true, Value: body.Body.Email}
		node.NewInputField("email", body.Email, node.LinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputEmail()),
	)

	f.Active = sqlxx.NullString(s.RecoveryNodeGroup())
	f.State = recovery.StateEmailSent
	f.UI.Messages.Set(text.NewRecoveryEmailSent())
	if err := s.d.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
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
		if err := s.d.PrivilegedIdentityPool().UpdateVerifiableAddress(r.Context(), address); err != nil {
			return s.HandleRecoveryError(w, r, f, nil, err)
		}
	}

	return nil
}

func (s *Strategy) HandleRecoveryError(w http.ResponseWriter, r *http.Request, req *recovery.Flow, body *recoverySubmitPayload, err error) error {
	if req != nil {
		email := ""
		if body != nil {
			email = body.Email
		}

		req.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		req.UI.GetNodes().Upsert(
			// v0.5: form.Field{Name: "email", Type: "email", Required: true, Value: body.Body.Email}
			node.NewInputField("email", email, node.LinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputEmail()),
		)
	}

	return err
}

type recoverySubmitPayload struct {
	Method    string `json:"method" form:"method"`
	Token     string `json:"token" form:"token"`
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
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return &body, nil
}
