package otp

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
)

const (
	RouteAdminCreateRecoveryOTP = "/recovery/otp"
)

func (s *Strategy) RecoveryStrategyID() string {
	return recovery.StrategyRecoveryOTPName
}

func (s *Strategy) RegisterPublicRecoveryRoutes(public *x.RouterPublic) {
	s.d.CSRFHandler().IgnorePath(RouteAdminCreateRecoveryOTP)
	public.POST(RouteAdminCreateRecoveryOTP, x.RedirectToAdminRoute(s.d))
}

func (s *Strategy) RegisterAdminRecoveryRoutes(_ *x.RouterAdmin) {
}

func (s *Strategy) PopulateRecoveryMethod(r *http.Request, f *recovery.Flow) error {
	if s.d.Config(r.Context()).SelfServiceStrategy(s.RecoveryStrategyID()).Enabled {
		return nil
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		node.NewInputField("identifier", nil, node.OTPGroup, node.InputAttributeTypeIdentifier),
	)
	f.UI.GetNodes().Append(node.NewInputField("method", s.RecoveryStrategyID(), node.OTPGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSubmit()))

	return nil
}

func (s *Strategy) Recover(w http.ResponseWriter, r *http.Request, _ *recovery.Flow) (err error) {
	body, err := s.decodeRecovery(r)
	if err != nil {
		return s.HandleRecoveryError(r, nil, body, err)
	}

	if body.Token != "" {
		if err := flow.MethodEnabledAndAllowed(r.Context(), s.RecoveryStrategyID(), s.RecoveryStrategyID(), s.d); err != nil {
			return s.HandleRecoveryError(r, nil, body, err)
		}

		return s.recoveryUseToken(w, r, body)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), s.RecoveryStrategyID(), body.Method, s.d); err != nil {
		return s.HandleRecoveryError(r, nil, body, err)
	}

	req, err := s.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), x.ParseUUID(body.Flow))
	if err != nil {
		return s.HandleRecoveryError(r, req, body, err)
	}

	if err := req.Valid(); err != nil {
		return s.HandleRecoveryError(r, req, body, err)
	}

	switch req.State {
	case recovery.StateChooseMethod:
		fallthrough
	case recovery.StateOTPSent:
		return s.recoveryHandleFormSubmission(r, req)
	case recovery.StatePassedChallenge:
		// was already handled, do not allow retry
		return s.retryRecoveryFlowWithMessage(w, r, req.Type, text.NewErrorValidationRecoveryRetrySuccess())
	default:
		return s.retryRecoveryFlowWithMessage(w, r, req.Type, text.NewErrorValidationRecoveryStateFailure())
	}
}

func (s *Strategy) recoveryUseToken(w http.ResponseWriter, r *http.Request, body *payloadBody) error {
	tkn, err := s.d.RecoveryTokenPersister().UseRecoveryToken(r.Context(), body.Token)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			return s.retryRecoveryFlowWithMessage(w, r, flow.TypeBrowser, text.NewErrorValidationRecoveryTokenInvalidOrAlreadyUsed())
		}

		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	var f *recovery.Flow
	if !tkn.FlowID.Valid {
		f, err = recovery.NewFlow(s.d.Config(r.Context()), time.Until(tkn.ExpiresAt), s.d.GenerateCSRFToken(r),
			r, s.d.RecoveryStrategies(r.Context()), flow.TypeBrowser)
		if err != nil {
			return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
		}

		if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), f); err != nil {
			return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
		}
	} else {
		f, err = s.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), tkn.FlowID.UUID)
		if err != nil {
			return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
		}
	}

	if err := tkn.Valid(); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	recovered, err := s.d.IdentityPool().GetIdentity(r.Context(), tkn.IdentityID)
	if err != nil {
		return s.HandleRecoveryError(r, f, nil, err)
	}

	// mark address as verified only for a self-service flow
	if tkn.FlowID.Valid {
		if err := s.markRecoveryAddressVerified(r, f, recovered, tkn.RecoveryAddress); err != nil {
			return s.HandleRecoveryError(r, f, body, err)
		}
	}

	return s.recoveryIssueSession(w, r, f, recovered)
}

func (s *Strategy) recoveryIssueSession(w http.ResponseWriter, r *http.Request, f *recovery.Flow, id *identity.Identity) error {
	f.UI.Messages.Clear()
	f.State = recovery.StatePassedChallenge
	f.SetCSRFToken(flow.GetCSRFToken(s.d, w, r, f.Type))
	f.RecoveredIdentityID = uuid.NullUUID{
		UUID:  id.ID,
		Valid: true,
	}
	if err := s.d.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sess, err := session.NewActiveSession(id, s.d.Config(r.Context()), time.Now().UTC(),
		identity.CredentialsTypeRecoveryOTP, identity.AuthenticatorAssuranceLevel1)
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

	if err := s.d.RecoveryExecutor().PostRecoveryHook(w, r, f, sess); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	sf.UI.Messages.Set(text.NewRecoverySuccessful(time.Now().Add(s.d.Config(r.Context()).SelfServiceFlowSettingsPrivilegedSessionMaxAge())))
	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), sf); err != nil {
		return s.retryRecoveryFlowWithError(w, r, flow.TypeBrowser, err)
	}

	http.Redirect(w, r, sf.AppendTo(s.d.Config(r.Context()).SelfServiceFlowSettingsUI()).String(), http.StatusSeeOther)

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryRecoveryFlowWithMessage(w http.ResponseWriter, r *http.Request, ft flow.Type, message *text.Message) error {
	s.d.Logger().WithRequest(r).WithField("message", message).Debug("A recovery flow is being retried because a validation error occurred.")

	config := s.d.Config(r.Context())

	f, err := recovery.NewFlow(config, config.SelfServiceFlowRecoveryRequestLifespan(), s.d.CSRFHandler().RegenerateToken(w, r), r, s.d.RecoveryStrategies(r.Context()), ft)
	if err != nil {
		return err
	}

	f.UI.Messages.Add(message)
	if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), f); err != nil {
		return err
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, f.AppendTo(config.SelfServiceFlowRecoveryUI()).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, s.createGetFlowURL(r.Context(), recovery.RouteGetFlow, f.ID.String()), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryRecoveryFlowWithError(w http.ResponseWriter, r *http.Request, ft flow.Type, recErr error) error {
	s.d.Logger().WithRequest(r).WithError(recErr).Debug("A recovery flow is being retried because a validation error occurred.")

	config := s.d.Config(r.Context())

	f, err := recovery.NewFlow(config, config.SelfServiceFlowRecoveryRequestLifespan(),
		s.d.CSRFHandler().RegenerateToken(w, r), r, s.d.RecoveryStrategies(r.Context()), ft)
	if err != nil {
		return err
	}

	if expired := new(flow.ExpiredError); errors.As(recErr, &expired) {
		return s.retryRecoveryFlowWithMessage(w, r, ft, text.NewErrorValidationRecoveryFlowExpired(expired.Ago))
	} else {
		if err := f.UI.ParseError(node.OTPGroup, recErr); err != nil {
			return err
		}
	}

	if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), f); err != nil {
		return err
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, f.AppendTo(config.SelfServiceFlowRecoveryUI()).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, s.createGetFlowURL(r.Context(), recovery.RouteGetFlow, f.ID.String()), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) recoveryHandleFormSubmission(r *http.Request, f *recovery.Flow) error {
	body, err := s.decodeRecovery(r)
	if err != nil {
		return s.HandleRecoveryError(r, f, body, err)
	}

	if body.Identifier == "" {
		return s.HandleRecoveryError(r, f, body, schema.NewRequiredError("#/phone", "phone"))
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, body.CSRFToken); err != nil {
		return s.HandleRecoveryError(r, f, body, err)
	}

	if err := s.d.OTPSender().SendRecoveryOTP(r.Context(), r, f, body.Identifier); err != nil {
		if !errors.Is(err, ErrUnknownIdentifier) {
			return s.HandleRecoveryError(r, f, body, err)
		}
		// Continue execution
	}

	s.updateRecoveryFlowUI(f, r, body.Identifier)

	f.Active = sqlxx.NullString(s.RecoveryNodeGroup())

	f.State = recovery.StateOTPSent
	f.UI.Messages.Set(text.NewRecoveryOTPSent())
	if err := s.d.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		return s.HandleRecoveryError(r, f, body, err)
	}

	return nil
}

func (s *Strategy) markRecoveryAddressVerified(r *http.Request, f *recovery.Flow, id *identity.Identity, recoveryAddress *identity.RecoveryAddress) error {
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
			return s.HandleRecoveryError(r, f, nil, err)
		}
	}

	return nil
}

func (s *Strategy) HandleRecoveryError(r *http.Request, f *recovery.Flow, body *payloadBody, err error) error {
	if f != nil {
		var phone string
		if body != nil && body.Identifier != "" {
			phone = body.Identifier
		}

		s.updateRecoveryFlowUI(f, r, phone)
	}

	return err
}

func (s *Strategy) decodeRecovery(r *http.Request) (*payloadBody, error) {
	var body payloadBody

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(recoveryMethodSchema)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := s.dx.Decode(r, &body, compiler,
		decoderx.HTTPDecoderUseQueryAndBody(),
		decoderx.HTTPKeepRequestBody(true),
		decoderx.HTTPDecoderAllowedMethods(http.MethodPost, http.MethodGet),
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return &body, nil
}

func (s *Strategy) updateRecoveryFlowUI(f *recovery.Flow, r *http.Request, identifier string) {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		node.NewInputField("identifier", identifier, node.OTPGroup, node.InputAttributeTypeIdentifier, node.WithRequiredInputAttribute),
	)
}
