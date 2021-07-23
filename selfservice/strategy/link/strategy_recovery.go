package link

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
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
	wrappedCreateRecoveryLink := strategy.IsDisabled(s.d, s.RecoveryStrategyID(), s.createRecoveryLink)
	admin.POST(RouteAdminCreateRecoveryLink, wrappedCreateRecoveryLink)
}

func (s *Strategy) PopulateRecoveryMethod(r *http.Request, f *recovery.Flow) error {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true},
		node.NewInputField("email", nil, node.RecoveryLinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute),
	)
	f.UI.GetNodes().Append(node.NewInputField("method", s.RecoveryStrategyID(), node.RecoveryLinkGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSubmit()))

	return nil
}

// swagger:parameters adminCreateSelfServiceRecoveryLink
//
// nolint
type adminCreateSelfServiceRecoveryLink struct {
	// in: body
	Body adminCreateSelfServiceRecoveryLinkBody
}

// swagger:model adminCreateSelfServiceRecoveryLinkBody
type adminCreateSelfServiceRecoveryLinkBody struct {
	// Identity to Recover
	//
	// The identity's ID you wish to recover.
	//
	// required: true
	IdentityID uuid.UUID `json:"identity_id"`

	// Link Expires In
	//
	// The recovery link will expire at that point in time. Defaults to the configuration value of
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

// swagger:model selfServiceRecoveryLink
// nolint
type selfServiceRecoveryLink struct {
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

// swagger:route POST /recovery/link v0alpha1 adminCreateSelfServiceRecoveryLink
//
// Create a Recovery Link
//
// This endpoint creates a recovery link which should be given to the user in order for them to recover
// (or activate) their account.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: selfServiceRecoveryLink
//       404: jsonError
//       400: jsonError
//       500: jsonError
func (s *Strategy) createRecoveryLink(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p adminCreateSelfServiceRecoveryLinkBody
	if err := s.dx.Decode(r, &p, decoderx.HTTPJSONDecoder()); err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	expiresIn := s.d.Config(r.Context()).SelfServiceLinkMethodLifespan()
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

	req, err := recovery.NewFlow(s.d.Config(r.Context()), expiresIn, s.d.GenerateCSRFToken(r),
		r, s.d.RecoveryStrategies(r.Context()), flow.TypeBrowser)
	if err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	id, err := s.d.IdentityPool().GetIdentity(r.Context(), p.IdentityID)
	if err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	if len(id.RecoveryAddresses) == 0 {
		s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The identity does not have any recovery addresses set.")))
		return
	}

	address := id.RecoveryAddresses[0]
	token := NewRecoveryToken(&address, expiresIn)
	if err := s.d.RecoveryTokenPersister().CreateRecoveryToken(r.Context(), token); err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	s.d.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithSensitiveField("email_address", address.Value).
		WithSensitiveField("recovery_link_token", token).
		Info("A recovery link has been created.")

	s.d.Writer().Write(w, r, &selfServiceRecoveryLink{
		ExpiresAt: req.ExpiresAt.UTC(),
		RecoveryLink: urlx.CopyWithQuery(
			urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(r), recovery.RouteSubmitFlow),
			url.Values{
				"token": {token.Token},
				"flow":  {req.ID.String()},
			}).String()},
		herodot.UnescapedHTML)
}

// swagger:model submitSelfServiceRecoveryFlowWithLinkMethodBody
// nolint:deadcode,unused
type submitSelfServiceRecoveryFlowWithLinkMethodBody struct {
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

		return s.recoveryUseToken(w, r, body)
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

func (s *Strategy) recoveryIssueSession(w http.ResponseWriter, r *http.Request, f *recovery.Flow, recoveredID uuid.UUID) error {
	recovered, err := s.d.IdentityPool().GetIdentity(r.Context(), recoveredID)
	if err != nil {
		return s.HandleRecoveryError(w, r, f, nil, err)
	}

	f.UI.Messages.Clear()
	f.State = recovery.StatePassedChallenge
	f.SetCSRFToken(flow.GetCSRFToken(s.d, w, r, f.Type))
	f.RecoveredIdentityID = uuid.NullUUID{
		UUID:  recoveredID,
		Valid: true,
	}
	if err := s.d.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return s.HandleRecoveryError(w, r, f, nil, err)
	}

	sess, _ := session.NewActiveSession(recovered, s.d.Config(r.Context()), time.Now().UTC())
	if err := s.d.SessionManager().CreateAndIssueCookie(r.Context(), w, r, sess); err != nil {
		return s.HandleRecoveryError(w, r, f, nil, err)
	}

	sf, err := s.d.SettingsHandler().NewFlow(w, r, sess.Identity, flow.TypeBrowser)
	if err != nil {
		return s.HandleRecoveryError(w, r, f, nil, err)
	}

	if err := s.d.RecoveryExecutor().PostRecoveryHook(w, r, f, sess); err != nil {
		return s.HandleRecoveryError(w, r, f, nil, err)
	}

	sf.UI.Messages.Set(text.NewRecoverySuccessful(time.Now().Add(s.d.Config(r.Context()).SelfServiceFlowSettingsPrivilegedSessionMaxAge())))
	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), sf); err != nil {
		return s.HandleRecoveryError(w, r, f, nil, err)
	}

	http.Redirect(w, r, sf.AppendTo(s.d.Config(r.Context()).SelfServiceFlowSettingsUI()).String(), http.StatusFound)
	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) recoveryUseToken(w http.ResponseWriter, r *http.Request, body *recoverySubmitPayload) error {
	token, err := s.d.RecoveryTokenPersister().UseRecoveryToken(r.Context(), body.Token)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			return s.retryRecoveryFlowWithMessage(w, r, flow.TypeBrowser, text.NewErrorValidationRecoveryTokenInvalidOrAlreadyUsed())
		}

		return s.HandleRecoveryError(w, r, nil, body, err)
	}

	var f *recovery.Flow
	if !token.FlowID.Valid {
		f, err = recovery.NewFlow(s.d.Config(r.Context()), time.Until(token.ExpiresAt), s.d.GenerateCSRFToken(r),
			r, s.d.RecoveryStrategies(r.Context()), flow.TypeBrowser)
		if err != nil {
			return s.HandleRecoveryError(w, r, nil, body, err)
		}

		if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), f); err != nil {
			return s.HandleRecoveryError(w, r, nil, body, err)
		}
	} else {
		f, err = s.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), token.FlowID.UUID)
		if err != nil {
			return s.HandleRecoveryError(w, r, nil, body, err)
		}
	}

	if err := token.Valid(); err != nil {
		return s.HandleRecoveryError(w, r, f, body, err)
	}

	return s.recoveryIssueSession(w, r, f, token.RecoveryAddress.IdentityID)
}

func (s *Strategy) retryRecoveryFlowWithMessage(w http.ResponseWriter, r *http.Request, ft flow.Type, message *text.Message) error {
	s.d.Logger().WithRequest(r).WithField("message", message).Debug("A recovery flow is being retried because a validation error occurred.")

	req, err := recovery.NewFlow(s.d.Config(r.Context()), s.d.Config(r.Context()).SelfServiceFlowRecoveryRequestLifespan(), s.d.GenerateCSRFToken(r), r, s.d.RecoveryStrategies(r.Context()), ft)
	if err != nil {
		return err
	}

	req.UI.Messages.Add(message)
	if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		return err
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, req.AppendTo(s.d.Config(r.Context()).SelfServiceFlowRecoveryUI()).String(), http.StatusFound)
	} else {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(r),
			recovery.RouteGetFlow), url.Values{"id": {req.ID.String()}}).String(), http.StatusFound)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) recoveryHandleFormSubmission(w http.ResponseWriter, r *http.Request, req *recovery.Flow) error {
	body, err := s.decodeRecovery(r)
	if err != nil {
		return s.HandleRecoveryError(w, r, req, body, err)
	}

	if len(body.Email) == 0 {
		return s.HandleRecoveryError(w, r, req, body, schema.NewRequiredError("#/email", "email"))
	}

	if err := flow.EnsureCSRF(s.d, r, req.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, body.CSRFToken); err != nil {
		return s.HandleRecoveryError(w, r, req, body, err)
	}

	if err := s.d.LinkSender().SendRecoveryLink(r.Context(), r, req, identity.VerifiableAddressTypeEmail, body.Email); err != nil {
		if !errors.Is(err, ErrUnknownAddress) {
			return s.HandleRecoveryError(w, r, req, body, err)
		}
		// Continue execution
	}

	req.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	req.UI.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true, Value: body.Body.Email}
		node.NewInputField("email", body.Email, node.RecoveryLinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute),
	)

	req.Active = sqlxx.NullString(s.RecoveryNodeGroup())
	req.State = recovery.StateEmailSent
	req.UI.Messages.Set(text.NewRecoveryEmailSent())
	if err := s.d.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), req); err != nil {
		return s.HandleRecoveryError(w, r, req, body, err)
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
			node.NewInputField("email", email, node.RecoveryLinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute),
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
