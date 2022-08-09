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
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

const (
	RouteAdminCreateRecoveryCode = "/recovery/code"
)

func (s *Strategy) RecoveryStrategyID() string {
	return recovery.StrategyRecoveryCodeName
}

func (s *Strategy) RegisterPublicRecoveryRoutes(public *x.RouterPublic) {
	s.deps.CSRFHandler().IgnorePath(RouteAdminCreateRecoveryCode)
	public.POST(RouteAdminCreateRecoveryCode, x.RedirectToAdminRoute(s.deps))

}

func (s *Strategy) RegisterAdminRecoveryRoutes(admin *x.RouterAdmin) {
	wrappedCreateRecoveryCode := strategy.IsDisabled(s.deps, s.RecoveryStrategyID(), s.createRecoveryCode)
	admin.POST(RouteAdminCreateRecoveryCode, wrappedCreateRecoveryCode)
}

func (s *Strategy) PopulateRecoveryMethod(r *http.Request, f *recovery.Flow) error {
	f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		node.NewInputField("email", nil, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputEmail()),
	)
	f.UI.
		GetNodes().
		Append(node.NewInputField("method", s.RecoveryStrategyID(), node.CodeGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSubmit()))

	return nil
}

// swagger:parameters adminCreateSelfServiceRecoveryCode
//
// nolint
type adminCreateSelfServiceRecoveryCode struct {
	// in: body
	Body adminCreateSelfServiceRecoveryCodeBody
}

// swagger:model adminCreateSelfServiceRecoveryCodeBody
type adminCreateSelfServiceRecoveryCodeBody struct {
	// Identity to Recover
	//
	// The identity's ID you wish to recover.
	//
	// required: true
	IdentityID uuid.UUID `json:"identity_id"`

	// Code Expires In
	//
	// The recovery code will expire at that point in time. Defaults to the configuration value of
	// `selfservice.flows.recovery.request_lifespan`.
	//
	//
	// pattern: ^[0-9]+(ns|us|ms|s|m|h)$
	// example:
	//	- 1h
	//	- 1m
	//	- 1s
	ExpiresIn string `json:"expires_in"`
}

// swagger:model selfServiceRecoveryCode
// nolint
type selfServiceRecoveryCode struct {
	// Recovery Link
	//
	// This link can be used to recover the account.
	//
	// required: true
	// format: uri
	RecoveryCode string `json:"recovery_code"`

	// Recovery Link Expires At
	//
	// The timestamp when the recovery link expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// swagger:route POST /admin/recovery/code v0alpha2 adminCreateSelfServiceRecoveryCode
//
// # Create a Recovery Link
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
//	Responses:
//		200: selfServiceRecoveryCode
//		400: jsonError
//		404: jsonError
//		500: jsonError
func (s *Strategy) createRecoveryCode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p adminCreateSelfServiceRecoveryCodeBody
	if err := s.dx.Decode(r, &p, decoderx.HTTPJSONDecoder()); err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	expiresIn := s.deps.Config(r.Context()).SelfServiceLinkMethodLifespan()
	if len(p.ExpiresIn) > 0 {
		var err error
		expiresIn, err = time.ParseDuration(p.ExpiresIn)
		if err != nil {
			s.deps.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to parse "expires_in" whose format should match "[0-9]+(ns|us|ms|s|m|h)" but did not: %s`, p.ExpiresIn)))
			return
		}
	}

	if time.Now().Add(expiresIn).Before(time.Now()) {
		s.deps.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Value from "expires_in" must be result to a future time: %s`, p.ExpiresIn)))
		return
	}

	req, err := recovery.NewFlow(s.deps.Config(r.Context()), expiresIn, s.deps.GenerateCSRFToken(r),
		r, s.deps.RecoveryStrategies(r.Context()), flow.TypeBrowser)
	if err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	if err := s.deps.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	id, err := s.deps.IdentityPool().GetIdentity(r.Context(), p.IdentityID)
	if errors.Is(err, sqlcon.ErrNoRows) {
		s.deps.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The requested identity id does not exist.").WithWrap(err)))
		return
	} else if err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	code := NewRecoveryCode(id.ID, expiresIn)
	if err := s.deps.RecoveryCodePersister().CreateRecoveryCode(r.Context(), code); err != nil {
		s.deps.Writer().WriteError(w, r, err)
		return
	}

	s.deps.Audit().
		WithField("identity_id", id.ID).
		WithSensitiveField("recovery_link_code", code).
		Info("A recovery code has been created.")

	s.deps.Writer().Write(w, r, &selfServiceRecoveryCode{
		ExpiresAt:    req.ExpiresAt.UTC(),
		RecoveryCode: code.Code,
	})
}

// swagger:model submitSelfServiceRecoveryFlowWithCodeMethodBody
// nolint:deadcode,unused
type submitSelfServiceRecoveryFlowWithCodeMethodBody struct {
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

	if len(body.Code) > 0 {
		if err := flow.MethodEnabledAndAllowed(r.Context(), s.RecoveryStrategyID(), s.RecoveryStrategyID(), s.deps); err != nil {
			return s.HandleRecoveryError(w, r, nil, body, err)
		}

		return s.recoveryUseCode(w, r, body)
	}

	if _, err := s.deps.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
		if x.IsJSONRequest(r) {
			session.RespondWithJSONErrorOnAuthenticated(s.deps.Writer(), recovery.ErrAlreadyLoggedIn)(w, r, nil)
		} else {
			session.RedirectOnAuthenticated(s.deps)(w, r, nil)
		}
		return errors.WithStack(flow.ErrCompletedByStrategy)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), s.RecoveryStrategyID(), body.Method, s.deps); err != nil {
		return s.HandleRecoveryError(w, r, nil, body, err)
	}

	req, err := s.deps.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), x.ParseUUID(body.Flow))
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
	f.SetCSRFToken(s.deps.CSRFHandler().RegenerateToken(w, r))
	f.RecoveredIdentityID = uuid.NullUUID{
		UUID:  id.ID,
		Valid: true,
	}
	if err := s.deps.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sess, err := session.NewActiveSession(id, s.deps.Config(r.Context()), time.Now().UTC(), identity.CredentialsTypeRecoveryLink, identity.AuthenticatorAssuranceLevel1)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	if err := s.deps.SessionManager().UpsertAndIssueCookie(r.Context(), w, r, sess); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sf, err := s.deps.SettingsHandler().NewFlow(w, r, sess.Identity, flow.TypeBrowser)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	// Take over `return_to` parameter from recovery flow
	sfRequestURL, err := url.Parse(sf.RequestURL)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}
	fRequestURL, err := url.Parse(f.RequestURL)
	if err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}
	sfQuery := sfRequestURL.Query()
	sfQuery.Set("return_to", fRequestURL.Query().Get("return_to"))
	sfRequestURL.RawQuery = sfQuery.Encode()
	sf.RequestURL = sfRequestURL.String()

	if err := s.deps.RecoveryExecutor().PostRecoveryHook(w, r, f, sess); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sf.UI.Messages.Set(text.NewRecoverySuccessful(time.Now().Add(s.deps.Config(r.Context()).SelfServiceFlowSettingsPrivilegedSessionMaxAge())))
	if err := s.deps.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), sf); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	http.Redirect(w, r, sf.AppendTo(s.deps.Config(r.Context()).SelfServiceFlowSettingsUI()).String(), http.StatusSeeOther)
	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) recoveryUseCode(w http.ResponseWriter, r *http.Request, body *recoverySubmitPayload) error {
	token, err := s.deps.RecoveryCodePersister().UseRecoveryCode(r.Context(), body.Code)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			return s.retryRecoveryFlowWithMessage(w, r, flow.TypeBrowser, text.NewErrorValidationRecoveryTokenInvalidOrAlreadyUsed())
		}

		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	var f *recovery.Flow
	if !token.FlowID.Valid {
		f, err = recovery.NewFlow(s.deps.Config(r.Context()), time.Until(token.ExpiresAt), s.deps.GenerateCSRFToken(r),
			r, s.deps.RecoveryStrategies(r.Context()), flow.TypeBrowser)
		if err != nil {
			return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
		}

		if err := s.deps.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), f); err != nil {
			return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
		}
	} else {
		f, err = s.deps.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), token.FlowID.UUID)
		if err != nil {
			return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
		}
	}

	if err := token.Valid(); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	recovered, err := s.deps.IdentityPool().GetIdentity(r.Context(), token.IdentityID)
	if err != nil {
		return s.HandleRecoveryError(w, r, f, nil, err)
	}

	// mark address as verified only for a self-service flow
	if token.FlowID.Valid {
		if err := s.markRecoveryAddressVerified(w, r, f, recovered, token.RecoveryAddress); err != nil {
			return s.HandleRecoveryError(w, r, f, body, err)
		}
	}

	return s.recoveryIssueSession(w, r, f, recovered)
}

func (s *Strategy) retryRecoveryFlowWithMessage(w http.ResponseWriter, r *http.Request, ft flow.Type, message *text.Message) error {
	s.deps.Logger().WithRequest(r).WithField("message", message).Debug("A recovery flow is being retried because a validation error occurred.")

	req, err := recovery.NewFlow(s.deps.Config(r.Context()), s.deps.Config(r.Context()).SelfServiceFlowRecoveryRequestLifespan(), s.deps.CSRFHandler().RegenerateToken(w, r), r, s.deps.RecoveryStrategies(r.Context()), ft)
	if err != nil {
		return err
	}

	req.UI.Messages.Add(message)
	if err := s.deps.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		return err
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, req.AppendTo(s.deps.Config(r.Context()).SelfServiceFlowRecoveryUI()).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.deps.Config(r.Context()).SelfPublicURL(),
			recovery.RouteGetFlow), url.Values{"id": {req.ID.String()}}).String(), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryRecoveryFlowWithError(w http.ResponseWriter, r *http.Request, ft flow.Type, recErr error) error {
	s.deps.Logger().WithRequest(r).WithError(recErr).Debug("A recovery flow is being retried because a validation error occurred.")

	req, err := recovery.NewFlow(s.deps.Config(r.Context()), s.deps.Config(r.Context()).SelfServiceFlowRecoveryRequestLifespan(), s.deps.CSRFHandler().RegenerateToken(w, r), r, s.deps.RecoveryStrategies(r.Context()), ft)
	if err != nil {
		return err
	}

	if expired := new(flow.ExpiredError); errors.As(recErr, &expired) {
		return s.retryRecoveryFlowWithMessage(w, r, ft, text.NewErrorValidationRecoveryFlowExpired(expired.Ago))
	} else {
		if err := req.UI.ParseError(node.CodeGroup, recErr); err != nil {
			return err
		}
	}

	if err := s.deps.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		return err
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, req.AppendTo(s.deps.Config(r.Context()).SelfServiceFlowRecoveryUI()).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.deps.Config(r.Context()).SelfPublicURL(),
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

	if err := flow.EnsureCSRF(s.deps, r, f.Type, s.deps.Config(r.Context()).DisableAPIFlowEnforcement(), s.deps.GenerateCSRFToken, body.CSRFToken); err != nil {
		return s.HandleRecoveryError(w, r, f, body, err)
	}

	if err := s.deps.RecoveryCodeSender().SendRecoveryCode(r.Context(), r, f, identity.VerifiableAddressTypeEmail, body.Email); err != nil {
		if !errors.Is(err, ErrUnknownAddress) {
			return s.HandleRecoveryError(w, r, f, body, err)
		}
		// Continue execution
	}

	f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true, Value: body.Body.Email}
		node.NewInputField("email", body.Email, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputEmail()),
	)

	f.Active = sqlxx.NullString(s.RecoveryNodeGroup())
	f.State = recovery.StateEmailSent
	f.UI.Messages.Set(text.NewRecoveryEmailSent())
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

func (s *Strategy) HandleRecoveryError(w http.ResponseWriter, r *http.Request, req *recovery.Flow, body *recoverySubmitPayload, err error) error {
	if req != nil {
		email := ""
		if body != nil {
			email = body.Email
		}

		req.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		req.UI.GetNodes().Upsert(
			node.NewInputField("email", email, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputEmail()),
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
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return &body, nil
}
