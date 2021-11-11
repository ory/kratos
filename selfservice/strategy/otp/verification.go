package otp

import (
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func (s *Strategy) VerificationStrategyID() string {
	return verification.StrategyVerificationOTPName
}

func (s *Strategy) RegisterPublicVerificationRoutes(public *x.RouterPublic) {
}

func (s *Strategy) RegisterAdminVerificationRoutes(admin *x.RouterAdmin) {
}

func (s *Strategy) PopulateVerificationMethod(r *http.Request, f *verification.Flow) error {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true}
		node.NewInputField("phone", nil, node.VerificationOTPGroup, node.InputAttributeTypePhone, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeLabelVerifyOTP()),
	)
	f.UI.GetNodes().Append(node.NewInputField("method", s.VerificationStrategyID(), node.VerificationOTPGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSubmit()))
	return nil
}

func (s *Strategy) decodeVerification(r *http.Request) (*payloadBody, error) {
	var body payloadBody

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(verificationMethodSchema)
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

// handleVerificationError is a convenience function for handling all types of errors that may occur (e.g. validation error).
func (s *Strategy) handleVerificationError(w http.ResponseWriter, r *http.Request, f *verification.Flow, body *payloadBody, err error) error {
	if f != nil {
		f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		f.UI.GetNodes().Upsert(
			node.NewInputField("phone", body.Phone, node.VerificationOTPGroup, node.InputAttributeTypePhone, node.WithRequiredInputAttribute),
		)
	}

	return err
}

func (s *Strategy) Verify(w http.ResponseWriter, r *http.Request, f *verification.Flow) (err error) {
	body, err := s.decodeVerification(r)
	if err != nil {
		return s.handleVerificationError(w, r, nil, body, err)
	}

	if body.Token != "" {
		if err := flow.MethodEnabledAndAllowed(r.Context(), s.VerificationStrategyID(), s.VerificationStrategyID(), s.d); err != nil {
			return s.handleVerificationError(w, r, nil, body, err)
		}

		return s.verificationUseToken(w, r, body)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), s.VerificationStrategyID(), body.Method, s.d); err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	if err := f.Valid(); err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	switch f.State {
	case verification.StateChooseMethod:
		fallthrough
	case verification.StateSmsSent:
		// Do nothing (continue with execution after this switch statement)
		return s.verificationHandleFormSubmission(w, r, f)
	case verification.StatePassedChallenge:
		return s.retryVerificationFlowWithMessage(w, r, f.Type, text.NewErrorValidationVerificationRetrySuccess())
	default:
		return s.retryVerificationFlowWithMessage(w, r, f.Type, text.NewErrorValidationVerificationStateFailure())
	}
}

func (s *Strategy) verificationHandleFormSubmission(w http.ResponseWriter, r *http.Request, f *verification.Flow) error {
	var body = new(payloadBody)
	body, err := s.decodeVerification(r)
	if err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	if len(body.Phone) == 0 {
		return s.handleVerificationError(w, r, f, body, schema.NewRequiredError("#/phone", "phone"))
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, body.CSRFToken); err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	if err := s.d.OTPSender().SendVerificationOTP(r.Context(), f, identity.VerifiableAddressTypePhone, body.Phone); err != nil {
		if !errors.Is(err, ErrUnknownPhone) {
			return s.handleVerificationError(w, r, f, body, err)
		}
		// Continue execution
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		node.NewInputField("phone", body.Phone, node.VerificationOTPGroup, node.InputAttributeTypePhone, node.WithRequiredInputAttribute),
	)

	f.Active = sqlxx.NullString(s.VerificationNodeGroup())
	f.State = verification.StateSmsSent
	f.UI.Messages.Set(text.NewVerificationOTPSent())
	if err := s.d.VerificationFlowPersister().UpdateVerificationFlow(r.Context(), f); err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	return nil
}

func (s *Strategy) verificationUseToken(w http.ResponseWriter, r *http.Request, body *payloadBody) error {
	token, err := s.d.VerificationTokenPersister().UseVerificationToken(r.Context(), body.Token)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			return s.retryVerificationFlowWithMessage(w, r, flow.TypeBrowser, text.NewErrorValidationVerificationTokenInvalidOrAlreadyUsed())
		}

		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	var f *verification.Flow
	if !token.FlowID.Valid {
		f, err = verification.NewFlow(s.d.Config(r.Context()), s.d.Config(r.Context()).SelfServiceFlowVerificationRequestLifespan(), s.d.GenerateCSRFToken(r), r, s.d.VerificationStrategies(r.Context()), flow.TypeBrowser)
		if err != nil {
			return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
		}

		if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), f); err != nil {
			return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
		}
	} else {
		f, err = s.d.VerificationFlowPersister().GetVerificationFlow(r.Context(), token.FlowID.UUID)
		if err != nil {
			return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
		}
	}

	if err := token.Valid(); err != nil {
		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	f.UI.Messages.Clear()
	f.State = verification.StatePassedChallenge
	// See https://github.com/ory/kratos/issues/1547
	f.SetCSRFToken(flow.GetCSRFToken(s.d, w, r, f.Type))
	f.UI.Messages.Set(text.NewInfoSelfServiceVerificationSuccessful())
	if err := s.d.VerificationFlowPersister().UpdateVerificationFlow(r.Context(), f); err != nil {
		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	i, err := s.d.IdentityPool().GetIdentity(r.Context(), token.VerifiableAddress.IdentityID)
	if err != nil {
		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	if err := s.d.VerificationExecutor().PostVerificationHook(w, r, f, i); err != nil {
		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	address := token.VerifiableAddress
	address.Verified = true
	verifiedAt := sqlxx.NullTime(time.Now().UTC())
	address.VerifiedAt = &verifiedAt
	address.Status = identity.VerifiableAddressStatusCompleted
	if err := s.d.PrivilegedIdentityPool().UpdateVerifiableAddress(r.Context(), address); err != nil {
		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	defaultRedirectURL := s.d.Config(r.Context()).SelfServiceFlowVerificationReturnTo(f.AppendTo(s.d.Config(r.Context()).SelfServiceFlowVerificationUI()))

	verificationRequestURL, err := urlx.Parse(f.GetRequestURL())
	if err != nil {
		s.d.Logger().Debugf("error parsing verification requestURL: %s\n", err)
		http.Redirect(w, r, defaultRedirectURL.String(), http.StatusSeeOther)
		return errors.WithStack(flow.ErrCompletedByStrategy)
	}
	verificationRequest := http.Request{URL: verificationRequestURL}

	returnTo, err := x.SecureRedirectTo(&verificationRequest, defaultRedirectURL,
		x.SecureRedirectAllowSelfServiceURLs(s.d.Config(r.Context()).SelfPublicURL()),
		x.SecureRedirectAllowURLs(s.d.Config(r.Context()).SelfServiceBrowserWhitelistedReturnToDomains()),
	)
	if err != nil {
		s.d.Logger().Debugf("error parsing redirectTo from verification: %s\n", err)
		http.Redirect(w, r, defaultRedirectURL.String(), http.StatusSeeOther)
		return errors.WithStack(flow.ErrCompletedByStrategy)
	}

	http.Redirect(w, r, returnTo.String(), http.StatusSeeOther)
	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryVerificationFlowWithMessage(w http.ResponseWriter, r *http.Request, ft flow.Type, message *text.Message) error {
	s.d.Logger().WithRequest(r).WithField("message", message).Debug("A verification flow is being retried because a validation error occurred.")

	f, err := verification.NewFlow(s.d.Config(r.Context()),
		s.d.Config(r.Context()).SelfServiceFlowVerificationRequestLifespan(), s.d.CSRFHandler().RegenerateToken(w, r), r, s.d.VerificationStrategies(r.Context()), ft)
	if err != nil {
		return s.handleVerificationError(w, r, f, nil, err)
	}

	f.UI.Messages.Add(message)
	if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), f); err != nil {
		return s.handleVerificationError(w, r, f, nil, err)
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, f.AppendTo(s.d.Config(r.Context()).SelfServiceFlowVerificationUI()).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(),
			verification.RouteGetFlow), url.Values{"id": {f.ID.String()}}).String(), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryVerificationFlowWithError(w http.ResponseWriter, r *http.Request, ft flow.Type, verErr error) error {
	s.d.Logger().WithRequest(r).WithError(verErr).Debug("A verification flow is being retried because an error occurred.")

	f, err := verification.NewFlow(s.d.Config(r.Context()),
		s.d.Config(r.Context()).SelfServiceFlowVerificationRequestLifespan(), s.d.CSRFHandler().RegenerateToken(w, r), r, s.d.VerificationStrategies(r.Context()), ft)
	if err != nil {
		return s.handleVerificationError(w, r, f, nil, err)
	}

	if expired := new(flow.ExpiredError); errors.As(verErr, &expired) {
		return s.retryVerificationFlowWithMessage(w, r, ft, text.NewErrorValidationVerificationFlowExpired(expired.Ago))
	} else {
		if err := f.UI.ParseError(node.RecoveryOTPGroup, verErr); err != nil {
			return err
		}
	}

	if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), f); err != nil {
		return s.handleVerificationError(w, r, f, nil, err)
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, f.AppendTo(s.d.Config(r.Context()).SelfServiceFlowVerificationUI()).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(),
			verification.RouteGetFlow), url.Values{"id": {f.ID.String()}}).String(), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}
