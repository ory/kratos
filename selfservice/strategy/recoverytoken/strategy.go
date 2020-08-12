package recoverytoken

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"

	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

const (
	PublicPath = "/self-service/browser/flows/recovery/link" // #nosec G101
	AdminPath  = "/recovery/link"
)

var _ recovery.Strategy = new(Strategy)
var _ recovery.AdminHandler = new(Strategy)
var _ recovery.PublicHandler = new(Strategy)

type (
	// swagger:model strategyRecoveryLinkMethodConfig
	StrategyMethodConfig struct {
		*form.HTMLForm
	}

	strategyDependencies interface {
		x.CSRFProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider

		session.HandlerProvider
		session.ManagementProvider
		settings.HandlerProvider

		identity.ValidationProvider
		identity.ManagementProvider
		identity.PoolProvider

		courier.Provider

		errorx.ManagementProvider

		recovery.ErrorHandlerProvider
		recovery.RequestPersistenceProvider
		recovery.StrategyProvider
		PersistenceProvider

		IdentityTraitsSchemas() schema.Schemas
	}

	Strategy struct {
		c  configuration.Provider
		d  strategyDependencies
		dx *decoderx.HTTP
	}
)

func NewStrategy(d strategyDependencies, c configuration.Provider) *Strategy {
	return &Strategy{c: c, d: d, dx: decoderx.NewHTTP()}
}

func (s *Strategy) RecoveryStrategyID() string {
	return recovery.StrategyRecoveryTokenName
}

func (s *Strategy) RegisterPublicRecoveryRoutes(public *x.RouterPublic) {
	redirect := session.RedirectOnAuthenticated(s.c)
	public.GET(PublicPath, s.d.SessionHandler().IsNotAuthenticated(s.handleSubmit, redirect))
	public.POST(PublicPath, s.d.SessionHandler().IsNotAuthenticated(s.handleSubmit, redirect))
}

func (s *Strategy) RegisterAdminRecoveryRoutes(admin *x.RouterAdmin) {
	admin.POST(AdminPath, s.createRecoveryLink)
}

func (s *Strategy) PopulateRecoveryMethod(r *http.Request, req *recovery.Request) error {
	f := form.NewHTMLForm(urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), PublicPath),
		url.Values{"request": {req.ID.String()}},
	).String())

	f.SetCSRF(s.d.GenerateCSRFToken(r))
	f.SetField(form.Field{Name: "email", Type: "email", Required: true})

	req.Methods[s.RecoveryStrategyID()] = &recovery.RequestMethod{
		Method: s.RecoveryStrategyID(),
		Config: &recovery.RequestMethodConfig{RequestMethodConfigurator: &StrategyMethodConfig{HTMLForm: f}},
	}
	return nil
}

// swagger:parameters createRecoveryLink
//
// nolint
type createRecoveryLinkParams struct {
	// in: body
	Body struct {
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
}

// swagger:model createRecoveryLinkResponse
//
// nolint
type createRecoveryLinkResponse struct {
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

// swagger:route POST /recovery/link admin createRecoveryLink
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
//       200: createRecoveryLinkResponse
//       404: genericError
//       400: genericError
//       500: genericError
func (s *Strategy) createRecoveryLink(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p createRecoveryLinkParams
	if err := s.dx.Decode(r, &p.Body, decoderx.HTTPJSONDecoder()); err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	expiresIn := s.c.SelfServiceFlowRecoveryRequestLifespan()
	if len(p.Body.ExpiresIn) > 0 {
		var err error
		expiresIn, err = time.ParseDuration(p.Body.ExpiresIn)
		if err != nil {
			s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to parse "expires_in" whose format should match "[0-9]+(ns|us|ms|s|m|h)" but did not: %s`, p.Body.ExpiresIn)))
			return
		}
	}

	if time.Now().Add(expiresIn).Before(time.Now()) {
		s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Value from "expires_in" must be result to a future time: %s`, p.Body.ExpiresIn)))
		return
	}

	req, err := recovery.NewRequest(expiresIn, "", r, nil)
	if err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	if err := s.d.RecoveryRequestPersister().CreateRecoveryRequest(r.Context(), req); err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	id, err := s.d.IdentityPool().GetIdentity(r.Context(), p.Body.IdentityID)
	if err != nil {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	if len(id.RecoveryAddresses) == 0 {
		s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The identity does not have any recovery addresses set.")))
		return
	}

	address := id.RecoveryAddresses[0]
	token := NewToken(&address, req)
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

	s.d.Writer().Write(w, r, &createRecoveryLinkResponse{
		ExpiresAt: req.ExpiresAt.UTC(),
		RecoveryLink: urlx.CopyWithQuery(
			urlx.AppendPaths(s.c.SelfPublicURL(), PublicPath),
			url.Values{
				"token":   {token.Token},
				"request": {req.ID.String()},
			}).String()})
}

// swagger:parameters completeSelfServiceBrowserRecoveryLinkStrategyFlow
//
// nolint
type completeSelfServiceBrowserRecoveryLinkStrategyFlowParams struct {
	// Email to Recover
	//
	// Needs to be set when initiating the flow. If the email is a registered
	// recovery email, a recovery link will be sent. If the email is not known,
	// a email with details on what happened will be sent instead.
	//
	// in: body
	Email string `json:"email"`

	// Recovery Token
	//
	// The recovery token which completes the recovery request. If the token
	// is invalid (e.g. expired) an error will be shown to the end-user.
	//
	// in: query
	Token string `json:"token"`

	// Recovery Request ID
	//
	// in: query
	Request string `json:"request"`
}

// swagger:route POST /self-service/browser/flows/recovery/link public completeSelfServiceBrowserRecoveryLinkStrategyFlow
//
// Complete the browser-based recovery flow using a recovery link
//
// > This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...) and HTML Forms.
//
// More information can be found at [ORY Kratos Account Recovery Documentation](../self-service/flows/password-reset-account-recovery).
//
//     Consumes:
//     - application/x-www-form-urlencoded
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (s *Strategy) handleSubmit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		s.handleError(w, r, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse the request: %s", err)))
		return
	}

	if len(r.Form.Get("token")) > 0 {
		s.verifyToken(w, r)
		return
	}

	rid := r.URL.Query().Get("request")
	req, err := s.d.RecoveryRequestPersister().GetRecoveryRequest(r.Context(), x.ParseUUID(rid))
	if err != nil {
		s.handleError(w, r, req, err)
		return
	}

	if err := req.Valid(); err != nil {
		s.handleError(w, r, req, err)
		return
	}

	switch req.State {
	case recovery.StateChooseMethod:
		fallthrough
	case recovery.StateEmailSent:
		s.issueAndSendRecoveryToken(w, r, req)
		return
	case recovery.StatePassedChallenge:
		// was already handled, do not allow retry
		s.retryFlowWithMessage(w, r, text.NewErrorValidationRecoveryRetrySuccess())
		return
	default:
		s.retryFlowWithMessage(w, r, text.NewErrorValidationRecoveryStateFailure())
		return
	}
}

func (s *Strategy) issueSession(w http.ResponseWriter, r *http.Request, req *recovery.Request) {
	req.State = recovery.StatePassedChallenge
	if err := s.d.RecoveryRequestPersister().UpdateRecoveryRequest(r.Context(), req); err != nil {
		s.handleError(w, r, req, err)
		return
	}

	recovered, err := s.d.IdentityPool().GetIdentity(r.Context(), req.RecoveredIdentityID.UUID)
	if err != nil {
		s.handleError(w, r, req, err)
		return
	}

	sess := session.NewSession(recovered, s.c, time.Now().UTC())
	if err := s.d.SessionManager().CreateAndIssueCookie(r.Context(), w, r, sess); err != nil {
		s.handleError(w, r, req, err)
		return
	}

	sr := settings.NewRequest(s.c.SelfServiceFlowSettingsRequestLifespan(), r, sess)
	sr.Messages.Set(text.NewRecoverySuccessful(time.Now().Add(s.c.SelfServiceFlowSettingsPrivilegedSessionMaxAge())))
	if err := s.d.SettingsHandler().CreateRequest(w, r, sess, sr); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r, sr.URL(s.c.SelfServiceFlowSettingsUI()).String(), http.StatusFound)
}

func (s *Strategy) verifyToken(w http.ResponseWriter, r *http.Request) {
	token, err := s.d.RecoveryTokenPersister().UseRecoveryToken(r.Context(), r.Form.Get("token"))
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			s.retryFlowWithMessage(w, r, text.NewErrorValidationRecoveryRecoveryTokenInvalidOrAlreadyUsed())
			return
		}

		s.handleError(w, r, nil, err)
		return
	}

	if err := token.Request.Valid(); err != nil {
		s.handleError(w, r, token.Request, err)
		return
	}

	req := token.Request
	req.Messages.Clear()
	req.State = recovery.StatePassedChallenge
	req.RecoveredIdentityID = uuid.NullUUID{UUID: token.RecoveryAddress.IdentityID, Valid: true}
	if err := s.d.RecoveryRequestPersister().UpdateRecoveryRequest(r.Context(), req); err != nil {
		s.handleError(w, r, req, err)
		return
	}

	s.issueSession(w, r, req)
}

func (s *Strategy) retryFlowWithMessage(w http.ResponseWriter, r *http.Request, message *text.Message) {
	s.d.Logger().WithRequest(r).WithField("message", message).Debug("A recovery flow is being retried because a validation error occurred.")

	req, err := recovery.NewRequest(s.c.SelfServiceFlowRecoveryRequestLifespan(), s.d.GenerateCSRFToken(r), r, s.d.RecoveryStrategies())
	if err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	req.Messages.Add(message)
	if err := s.d.RecoveryRequestPersister().CreateRecoveryRequest(r.Context(), req); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.SelfServiceFlowRecoveryUI(), url.Values{"request": {req.ID.String()}}).String(),
		http.StatusFound,
	)
}

func (s *Strategy) issueAndSendRecoveryToken(w http.ResponseWriter, r *http.Request, req *recovery.Request) {
	email := r.PostForm.Get("email")
	if len(email) == 0 {
		s.handleError(w, r, req, schema.NewRequiredError("#/email", "email"))
		return
	}

	s.d.Audit().
		WithField("via", identity.RecoveryAddressTypeEmail).
		WithSensitiveField("address", email).
		Info("Preparing account recovery token.")

	a, err := s.d.IdentityPool().FindRecoveryAddressByValue(r.Context(), identity.RecoveryAddressTypeEmail, email)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			if err := s.sendToUnknownAddress(r.Context(), email); err != nil {
				s.handleError(w, r, req, err)
				return
			}
		} else {
			s.handleError(w, r, req, err)
			return
		}
	} else if err := s.sendCodeToKnownAddress(r.Context(), req, a); err != nil {
		s.handleError(w, r, req, err)
		return
	}

	config, err := req.MethodToForm(s.RecoveryStrategyID())
	if err != nil {
		s.handleError(w, r, req, err)
		return
	}

	config.Reset()
	config.SetCSRF(s.d.GenerateCSRFToken(r))
	config.SetField(form.Field{Name: "email", Type: "email", Required: true, Value: r.PostForm.Get("email")})

	req.Active = sqlxx.NullString(s.RecoveryStrategyID())
	req.State = recovery.StateEmailSent
	req.Messages.Set(text.NewRecoveryEmailSent())
	if err := s.d.RecoveryRequestPersister().UpdateRecoveryRequest(r.Context(), req); err != nil {
		s.handleError(w, r, req, err)
		return
	}

	http.Redirect(w, r, req.URL(s.c.SelfServiceFlowRecoveryUI()).String(), http.StatusFound)
}

func (s *Strategy) sendToUnknownAddress(ctx context.Context, address string) error {
	s.d.Audit().
		WithField("via", identity.RecoveryAddressTypeEmail).
		WithSensitiveField("email_address", address).
		Info("Sending out stub recovery email because address is not linked to any account.")
	return s.run(identity.RecoveryAddressTypeEmail, func() error {
		_, err := s.d.Courier().QueueEmail(ctx,
			templates.NewRecoveryInvalid(s.c, &templates.RecoveryInvalidModel{To: address}))
		return err
	})
}

func (s *Strategy) sendCodeToKnownAddress(ctx context.Context, req *recovery.Request, address *identity.RecoveryAddress) error {
	token := NewToken(address, req)
	if err := s.d.RecoveryTokenPersister().CreateRecoveryToken(ctx, token); err != nil {
		return err
	}

	s.d.Audit().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithField("recovery_link_id", token.ID).
		WithSensitiveField("email_address", address.Value).
		WithSensitiveField("recovery_link_token", token.Token).
		Info("Sending out recovery email with recovery link.")
	return s.run(address.Via, func() error {
		_, err := s.d.Courier().QueueEmail(ctx, templates.NewRecoveryValid(s.c,
			&templates.RecoveryValidModel{To: address.Value, RecoveryURL: urlx.CopyWithQuery(
				urlx.AppendPaths(s.c.SelfPublicURL(), PublicPath),
				url.Values{"token": {token.Token}}).String()}))
		return err
	})
}

func (s *Strategy) run(via identity.RecoveryAddressType, emailFunc func() error) error {
	switch via {
	case identity.RecoveryAddressTypeEmail:
		return emailFunc()
	default:
		return errors.Errorf("received unexpected via type: %s", via)
	}
}

func (s *Strategy) handleError(w http.ResponseWriter, r *http.Request, req *recovery.Request, err error) {
	if errors.Is(err, recovery.ErrRequestExpired) {
		s.retryFlowWithMessage(w, r, text.NewErrorValidationRecoveryRecoveryTokenInvalidOrAlreadyUsed())
		return
	}

	if req != nil {
		config, err := req.MethodToForm(s.RecoveryStrategyID())
		if err != nil {
			s.d.RecoveryRequestErrorHandler().HandleRecoveryError(w, r, req, err, s.RecoveryStrategyID())
			return
		}

		config.Reset()
		config.SetCSRF(s.d.GenerateCSRFToken(r))
		config.SetField(form.Field{Name: "email", Type: "email", Required: true, Value: r.PostForm.Get("email")})
	}

	s.d.RecoveryRequestErrorHandler().HandleRecoveryError(w, r, req, err, s.RecoveryStrategyID())
}
