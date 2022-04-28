package otp

import (
	"context"
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

func (s *Strategy) RegisterPublicVerificationRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) RegisterAdminVerificationRoutes(_ *x.RouterAdmin) {
}

func (s *Strategy) Verify(w http.ResponseWriter, r *http.Request, f *verification.Flow) (err error) {
	body, err := s.decodeVerification(r)
	if err != nil {
		return s.verificationHandleError(err, f, r, body.Identifier)
	}

	if body.Token != "" {
		if err := flow.MethodEnabledAndAllowed(r.Context(), s.VerificationStrategyID(), s.VerificationStrategyID(), s.d); err != nil {
			return s.verificationHandleError(err, f, r, body.Identifier)
		}

		return s.verificationUseToken(w, r, body)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), s.VerificationStrategyID(), body.Method, s.d); err != nil {
		return s.verificationHandleError(err, f, r, body.Identifier)
	}

	if err := f.Valid(); err != nil {
		return s.verificationHandleError(err, f, r, body.Identifier)
	}

	switch f.State {
	case verification.StateChooseMethod:
		fallthrough
	case verification.StateOTPSent:
		// Do nothing (continue with execution after this switch statement)
		return s.verificationHandleFormSubmission(r, f)
	case verification.StatePassedChallenge:
		return s.retryVerificationFlowWithMessage(w, r, f.Type, text.NewErrorValidationVerificationRetrySuccess())
	default:
		return s.retryVerificationFlowWithMessage(w, r, f.Type, text.NewErrorValidationVerificationStateFailure())
	}
}

func (s *Strategy) PopulateVerificationMethod(r *http.Request, f *verification.Flow) error {
	if s.d.Config(r.Context()).SelfServiceStrategy(s.VerificationStrategyID()).Enabled {
		return nil
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		node.NewInputField("identifier", nil, node.OTPGroup, node.InputAttributeTypeIdentifier, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeLabelVerifyOTP()),
	)
	f.UI.GetNodes().Append(
		node.NewInputField("method", s.VerificationStrategyID(), node.OTPGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSubmit()),
	)

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

func (s *Strategy) verificationHandleFormSubmission(r *http.Request, f *verification.Flow) error {
	body, err := s.decodeVerification(r)
	if err != nil {
		return s.verificationHandleError(err, f, r, body.Identifier)
	}

	if body.Identifier == "" {
		requiredFieldErr := schema.NewRequiredError("#/identifier", "identifier")
		return s.verificationHandleError(requiredFieldErr, f, r, body.Identifier)
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, body.CSRFToken); err != nil {
		return s.verificationHandleError(err, f, r, body.Identifier)
	}

	if err := s.d.OTPSender().SendVerificationOTP(r.Context(), f, body.Identifier); err != nil {
		if !errors.Is(err, ErrUnknownIdentifier) {
			return s.verificationHandleError(err, f, r, body.Identifier)
		}
		// Continue execution
	}

	s.updateVerificationFlowUI(f, r, body.Identifier)

	f.State = verification.StateOTPSent
	f.Active = sqlxx.NullString(s.VerificationNodeGroup())
	f.UI.Messages.Set(text.NewVerificationOTPSent())

	if err := s.d.VerificationFlowPersister().UpdateVerificationFlow(r.Context(), f); err != nil {
		return s.verificationHandleError(err, f, r, "")
	}

	return nil
}

func (s *Strategy) verificationUseToken(w http.ResponseWriter, r *http.Request, body *payloadBody) error {
	conf := s.d.Config(r.Context())

	token, err := s.d.VerificationTokenPersister().UseVerificationToken(r.Context(), body.Token)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			return s.retryVerificationFlowWithMessage(w, r, flow.TypeBrowser, text.NewErrorValidationVerificationTokenInvalidOrAlreadyUsed())
		}

		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	var f *verification.Flow
	if !token.FlowID.Valid {
		f, err = verification.NewFlow(conf, conf.SelfServiceFlowVerificationRequestLifespan(), s.d.GenerateCSRFToken(r), r, s.d.VerificationStrategies(r.Context()), flow.TypeBrowser)
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

	err = s.redirectVerificationSuccess(w, r, f)
	if err != nil {
		return err
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryVerificationFlowWithMessage(w http.ResponseWriter, r *http.Request, ft flow.Type, message *text.Message) error {
	s.d.Logger().WithRequest(r).WithField("message", message).Debug("A verification flow is being retried because a validation error occurred.")

	conf := s.d.Config(r.Context())

	f, err := verification.NewFlow(conf, conf.SelfServiceFlowVerificationRequestLifespan(),
		s.d.CSRFHandler().RegenerateToken(w, r), r, s.d.VerificationStrategies(r.Context()), ft)
	if err != nil {
		return s.verificationHandleError(err, f, r, "")
	}

	f.UI.Messages.Add(message)
	if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), f); err != nil {
		return s.verificationHandleError(err, f, r, "")
	}

	var targetURL string
	if ft == flow.TypeBrowser {
		targetURL = f.AppendTo(conf.SelfServiceFlowVerificationUI()).String()
	} else {
		targetURL = s.createGetFlowURL(r.Context(), verification.RouteGetFlow, f.ID.String())
	}

	http.Redirect(w, r, targetURL, http.StatusSeeOther)

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryVerificationFlowWithError(w http.ResponseWriter, r *http.Request, ft flow.Type, verErr error) error {
	s.d.Logger().WithRequest(r).WithError(verErr).Debug("A verification flow is being retried because an error occurred.")

	conf := s.d.Config(r.Context())

	f, err := verification.NewFlow(conf, conf.SelfServiceFlowVerificationRequestLifespan(),
		s.d.CSRFHandler().RegenerateToken(w, r), r, s.d.VerificationStrategies(r.Context()), ft)
	if err != nil {
		return s.verificationHandleError(err, f, r, "")
	}

	if expired := new(flow.ExpiredError); errors.As(verErr, &expired) {
		return s.retryVerificationFlowWithMessage(w, r, ft, text.NewErrorValidationVerificationFlowExpired(expired.Ago))
	} else {
		if err := f.UI.ParseError(node.OTPGroup, verErr); err != nil {
			return err
		}
	}

	if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), f); err != nil {
		return s.verificationHandleError(err, f, r, "")
	}

	var targetURL string
	if ft == flow.TypeBrowser {
		targetURL = f.AppendTo(conf.SelfServiceFlowVerificationUI()).String()
	} else {
		targetURL = s.createGetFlowURL(r.Context(), verification.RouteGetFlow, f.ID.String())
	}

	http.Redirect(w, r, targetURL, http.StatusSeeOther)

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) redirectVerificationSuccess(w http.ResponseWriter, r *http.Request, f *verification.Flow) error {
	conf := s.d.Config(r.Context())

	defaultRedirectURL := conf.SelfServiceFlowVerificationReturnTo(f.AppendTo(conf.SelfServiceFlowVerificationUI()))

	verificationRequestURL, err := urlx.Parse(f.GetRequestURL())
	if err != nil {
		s.d.Logger().Debugf("error parsing verification requestURL: %s\n", err)
		http.Redirect(w, r, defaultRedirectURL.String(), http.StatusSeeOther)
		return errors.WithStack(flow.ErrCompletedByStrategy)
	}

	verificationRequest := http.Request{URL: verificationRequestURL}

	returnTo, err := x.SecureRedirectTo(&verificationRequest, defaultRedirectURL,
		x.SecureRedirectAllowSelfServiceURLs(conf.SelfPublicURL()),
		x.SecureRedirectAllowURLs(conf.SelfServiceBrowserAllowedReturnToDomains()),
	)
	if err != nil {
		s.d.Logger().Debugf("error parsing redirectTo from verification: %s\n", err)
		http.Redirect(w, r, defaultRedirectURL.String(), http.StatusSeeOther)
		return errors.WithStack(flow.ErrCompletedByStrategy)
	}

	http.Redirect(w, r, returnTo.String(), http.StatusSeeOther)

	return nil
}

func (s *Strategy) verificationHandleError(err error, f *verification.Flow, r *http.Request, identifier string) error {
	if f != nil {
		s.updateVerificationFlowUI(f, r, identifier)
	}

	return err
}

func (s *Strategy) updateVerificationFlowUI(f *verification.Flow, r *http.Request, identifier string) {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		node.NewInputField("identifier", identifier, node.OTPGroup, node.InputAttributeTypeIdentifier, node.WithRequiredInputAttribute),
	)
}

func (s *Strategy) createGetFlowURL(ctx context.Context, path, flowID string) string {
	addr := urlx.AppendPaths(s.d.Config(ctx).SelfPublicURL(), path)
	return urlx.CopyWithQuery(addr, url.Values{"id": {flowID}}).String()
}
