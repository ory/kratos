package link

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"
	"github.com/ory/x/randx"
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
	PublicRecoveryLinkPath = "/self-service/browser/flows/recovery/link"
)

var _ recovery.Strategy = new(StrategyLink)

type (
	// swagger:model strategyLinkMethodConfig
	StrategyLinkMethodConfig struct {
		*form.HTMLForm
	}

	strategyEmailDependencies interface {
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

	StrategyLink struct {
		c configuration.Provider
		d strategyEmailDependencies
	}
)

func NewStrategyLink(d strategyEmailDependencies, c configuration.Provider) *StrategyLink {
	return &StrategyLink{c: c, d: d}
}

func (s *StrategyLink) RecoveryStrategyID() string {
	return recovery.StrategyRecoveryTokenName
}

func (s *StrategyLink) RegisterRecoveryRoutes(public *x.RouterPublic) {
	redirect := session.RedirectOnAuthenticated(s.c)
	public.GET(PublicRecoveryLinkPath, s.d.SessionHandler().IsNotAuthenticated(s.handleSubmit, redirect))
	public.POST(PublicRecoveryLinkPath, s.d.SessionHandler().IsNotAuthenticated(s.handleSubmit, redirect))
}

func (s *StrategyLink) PopulateRecoveryMethod(r *http.Request, req *recovery.Request) error {
	f := form.NewHTMLForm(urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), PublicRecoveryLinkPath),
		url.Values{"request": {req.ID.String()}},
	).String())

	f.SetCSRF(s.d.GenerateCSRFToken(r))
	f.SetField(form.Field{Name: "email", Type: "email", Required: true})

	req.Methods[s.RecoveryStrategyID()] = &recovery.RequestMethod{
		Method: s.RecoveryStrategyID(),
		Config: &recovery.RequestMethodConfig{RequestMethodConfigurator: &StrategyLinkMethodConfig{HTMLForm: f}},
	}
	return nil
}

// swagger:model completeSelfServiceBrowserRecoveryLinkStrategyFlowPayload
//
// nolint
type completeSelfServiceBrowserRecoveryLinkStrategyFlowPayload struct {
	// Email
	//
	// in: body
	Email string `json:"email"`

	// RequestID is request ID.
	//
	// in: query
	RequestID string `json:"request_id"`
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
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (s *StrategyLink) handleSubmit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func (s *StrategyLink) issueSession(w http.ResponseWriter, r *http.Request, req *recovery.Request) {
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
	if err := s.d.SessionManager().CreateToRequest(r.Context(), w, r, sess); err != nil {
		s.handleError(w, r, req, err)
		return
	}

	sr := settings.NewRequest(s.c.SelfServiceSettingsRequestLifespan(), r, sess)
	sr.Messages.Set(text.NewRecoverySuccessful(time.Now().Add(s.c.SelfServicePrivilegedSessionMaxAge())))
	if err := s.d.SettingsHandler().CreateRequest(w, r, sess, sr); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r, sr.URL(s.c.SettingsURL()).String(), http.StatusFound)
}

func (s *StrategyLink) verifyToken(w http.ResponseWriter, r *http.Request) {
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

func (s *StrategyLink) retryFlowWithMessage(w http.ResponseWriter, r *http.Request, message *text.Message) {
	s.d.Logger().WithRequest(r).WithField("message", message).Debug("A recovery flow is being retried because a validation error occurred.")

	req, err := recovery.NewRequest(s.c.SelfServiceRecoveryRequestLifespan(), s.d.GenerateCSRFToken(r), r, s.d.RecoveryStrategies())
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
		urlx.CopyWithQuery(s.c.RecoveryURL(), url.Values{"request": {req.ID.String()}}).String(),
		http.StatusFound,
	)
}

func (s *StrategyLink) issueAndSendRecoveryToken(w http.ResponseWriter, r *http.Request, req *recovery.Request) {
	email := r.PostForm.Get("email")
	if len(email) == 0 {
		s.handleError(w, r, req, schema.NewRequiredError("#/email", "email"))
		return
	}

	s.d.Logger().
		WithField("via", identity.RecoveryAddressTypeEmail).
		WithSensitiveField("address", email).
		Debug("Preparing account recovery token.")

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

	http.Redirect(w, r, req.URL(s.c.RecoveryURL()).String(), http.StatusFound)
}

func (s *StrategyLink) sendToUnknownAddress(ctx context.Context, address string) error {
	s.d.Logger().
		WithField("via", identity.RecoveryAddressTypeEmail).
		WithSensitiveField("email_address", address).
		Debug("Sending out stub recovery email because address is not linked to any account.")
	return s.run(identity.RecoveryAddressTypeEmail, func() error {
		_, err := s.d.Courier().QueueEmail(ctx,
			templates.NewRecoveryInvalid(s.c, &templates.RecoveryInvalidModel{To: address}))
		return err
	})
}

func (s *StrategyLink) sendCodeToKnownAddress(ctx context.Context, req *recovery.Request, address *identity.RecoveryAddress) error {
	token := randx.MustString(32, randx.AlphaNum)
	if err := s.d.RecoveryTokenPersister().CreateRecoveryToken(ctx, NewToken(token, address, req)); err != nil {
		return err
	}

	s.d.Logger().
		WithField("via", address.Via).
		WithField("identity_id", address.IdentityID).
		WithSensitiveField("email_address", address.Value).
		WithSensitiveField("token", token).
		Debug("Sending out recovery email with recovery link.")
	return s.run(address.Via, func() error {
		_, err := s.d.Courier().QueueEmail(ctx, templates.NewRecoveryValid(s.c,
			&templates.RecoveryValidModel{To: address.Value, RecoveryURL: urlx.CopyWithQuery(
				urlx.AppendPaths(s.c.SelfPublicURL(), PublicRecoveryLinkPath),
				url.Values{"token": {token}}).String()}))
		return err
	})
}

func (s *StrategyLink) run(via identity.RecoveryAddressType, emailFunc func() error) error {
	switch via {
	case identity.RecoveryAddressTypeEmail:
		return emailFunc()
	default:
		return errors.Errorf("received unexpected via type: %s", via)
	}
}

func (s *StrategyLink) handleError(w http.ResponseWriter, r *http.Request, req *recovery.Request, err error) {
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
