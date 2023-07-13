// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link

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
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func (s *Strategy) VerificationStrategyID() string {
	return string(verification.VerificationStrategyLink)
}

func (s *Strategy) RegisterPublicVerificationRoutes(public *x.RouterPublic) {
}

func (s *Strategy) RegisterAdminVerificationRoutes(admin *x.RouterAdmin) {
}

func (s *Strategy) PopulateVerificationMethod(r *http.Request, f *verification.Flow) error {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true}
		node.NewInputField("email", nil, node.LinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputEmail()),
	)
	f.UI.GetNodes().Append(node.NewInputField("method", s.VerificationStrategyID(), node.LinkGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSubmit()))
	return nil
}

type verificationSubmitPayload struct {
	Method    string `json:"method" form:"method"`
	Token     string `json:"token" form:"token"`
	CSRFToken string `json:"csrf_token" form:"csrf_token"`
	Flow      string `json:"flow" form:"flow"`
	Email     string `json:"email" form:"email"`
}

func (s *Strategy) decodeVerification(r *http.Request) (*verificationSubmitPayload, error) {
	var body verificationSubmitPayload

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(verificationMethodSchema)
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

// handleVerificationError is a convenience function for handling all types of errors that may occur (e.g. validation error).
func (s *Strategy) handleVerificationError(w http.ResponseWriter, r *http.Request, f *verification.Flow, body *verificationSubmitPayload, err error) error {
	if f != nil {
		f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		f.UI.GetNodes().Upsert(
			// v0.5: form.Field{Name: "email", Type: "email", Required: true, Value: body.Body.Email}
			node.NewInputField("email", body.Email, node.LinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputEmail()),
		)
	}

	return err
}

// Update Verification Flow with Link Method
//
// swagger:model updateVerificationFlowWithLinkMethod
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateVerificationFlowWithLinkMethod struct {
	// Email to Verify
	//
	// Needs to be set when initiating the flow. If the email is a registered
	// verification email, a verification link will be sent. If the email is not known,
	// a email with details on what happened will be sent instead.
	//
	// format: email
	// required: true
	Email string `json:"email"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `form:"csrf_token" json:"csrf_token"`

	// Method is the method that should be used for this verification flow
	//
	// Allowed values are `link` and `code`
	//
	// required: true
	Method verification.VerificationStrategy `json:"method"`
}

func (s *Strategy) Verify(w http.ResponseWriter, r *http.Request, f *verification.Flow) (err error) {
	body, err := s.decodeVerification(r)
	if err != nil {
		return s.handleVerificationError(w, r, nil, body, err)
	}

	if len(body.Token) > 0 {
		if err := flow.MethodEnabledAndAllowed(r.Context(), f.GetFlowName(), s.VerificationStrategyID(), s.VerificationStrategyID(), s.d); err != nil {
			return s.handleVerificationError(w, r, nil, body, err)
		}

		return s.verificationUseToken(w, r, body, f)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), f.GetFlowName(), s.VerificationStrategyID(), body.Method, s.d); err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	if err := f.Valid(); err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	switch f.State {
	case flow.StateChooseMethod:
		fallthrough
	case flow.StateEmailSent:
		// Do nothing (continue with execution after this switch statement)
		return s.verificationHandleFormSubmission(w, r, f)
	case flow.StatePassedChallenge:
		return s.retryVerificationFlowWithMessage(w, r, f.Type, text.NewErrorValidationVerificationRetrySuccess())
	default:
		return s.retryVerificationFlowWithMessage(w, r, f.Type, text.NewErrorValidationVerificationStateFailure())
	}
}

func (s *Strategy) verificationHandleFormSubmission(w http.ResponseWriter, r *http.Request, f *verification.Flow) error {
	body := new(verificationSubmitPayload)
	body, err := s.decodeVerification(r)
	if err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	if len(body.Email) == 0 {
		return s.handleVerificationError(w, r, f, body, schema.NewRequiredError("#/email", "email"))
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, body.CSRFToken); err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	if err := s.d.LinkSender().SendVerificationLink(r.Context(), f, identity.VerifiableAddressTypeEmail, body.Email); err != nil {
		if !errors.Is(err, ErrUnknownAddress) {
			return s.handleVerificationError(w, r, f, body, err)
		}
		// Continue execution
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.GetNodes().Upsert(
		// v0.5: form.Field{Name: "email", Type: "email", Required: true, Value: body.Body.Email}
		node.NewInputField("email", body.Email, node.LinkGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputEmail()),
	)

	f.Active = sqlxx.NullString(s.NodeGroup())
	f.State = flow.StateEmailSent
	f.UI.Messages.Set(text.NewVerificationEmailSent())
	if err := s.d.VerificationFlowPersister().UpdateVerificationFlow(r.Context(), f); err != nil {
		return s.handleVerificationError(w, r, f, body, err)
	}

	return nil
}

func (s *Strategy) verificationUseToken(w http.ResponseWriter, r *http.Request, body *verificationSubmitPayload, f *verification.Flow) error {
	token, err := s.d.VerificationTokenPersister().UseVerificationToken(r.Context(), f.ID, body.Token)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			return s.retryVerificationFlowWithMessage(w, r, flow.TypeBrowser, text.NewErrorValidationVerificationTokenInvalidOrAlreadyUsed())
		}

		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	if err := token.Valid(); err != nil {
		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	i, err := s.d.IdentityPool().GetIdentity(r.Context(), token.VerifiableAddress.IdentityID, identity.ExpandDefault)
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

	returnTo := f.ContinueURL(r.Context(), s.d.Config())

	f.UI.
		Nodes.
		Append(node.NewAnchorField("continue", returnTo.String(), node.CodeGroup, text.NewInfoNodeLabelContinue()).
			WithMetaLabel(text.NewInfoNodeLabelContinue()))

	f.UI = &container.Container{
		Method: "GET",
		Action: returnTo.String(),
	}
	f.UI.Messages.Clear()
	f.State = flow.StatePassedChallenge
	// See https://github.com/ory/kratos/issues/1547
	f.SetCSRFToken(flow.GetCSRFToken(s.d, w, r, f.Type))
	f.UI.Messages.Set(text.NewInfoSelfServiceVerificationSuccessful())
	f.UI.
		Nodes.
		Append(node.NewAnchorField("continue", returnTo.String(), node.LinkGroup, text.NewInfoNodeLabelContinue()).
			WithMetaLabel(text.NewInfoNodeLabelContinue()))

	if err := s.d.VerificationFlowPersister().UpdateVerificationFlow(r.Context(), f); err != nil {
		return s.retryVerificationFlowWithError(w, r, flow.TypeBrowser, err)
	}

	return nil
}

func (s *Strategy) retryVerificationFlowWithMessage(w http.ResponseWriter, r *http.Request, ft flow.Type, message *text.Message) error {
	s.d.Logger().WithRequest(r).WithField("message", message).Debug("A verification flow is being retried because a validation error occurred.")

	f, err := verification.NewFlow(s.d.Config(),
		s.d.Config().SelfServiceFlowVerificationRequestLifespan(r.Context()), s.d.CSRFHandler().RegenerateToken(w, r), r, s, ft)
	if err != nil {
		return s.handleVerificationError(w, r, f, nil, err)
	}

	f.UI.Messages.Add(message)
	if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), f); err != nil {
		return s.handleVerificationError(w, r, f, nil, err)
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowVerificationUI(r.Context())).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()),
			verification.RouteGetFlow), url.Values{"id": {f.ID.String()}}).String(), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) retryVerificationFlowWithError(w http.ResponseWriter, r *http.Request, ft flow.Type, verErr error) error {
	s.d.Logger().WithRequest(r).WithError(verErr).Debug("A verification flow is being retried because an error occurred.")

	f, err := verification.NewFlow(s.d.Config(),
		s.d.Config().SelfServiceFlowVerificationRequestLifespan(r.Context()), s.d.CSRFHandler().RegenerateToken(w, r), r, s, ft)
	if err != nil {
		return s.handleVerificationError(w, r, f, nil, err)
	}

	if expired := new(flow.ExpiredError); errors.As(verErr, &expired) {
		return s.retryVerificationFlowWithMessage(w, r, ft, text.NewErrorValidationVerificationFlowExpired(expired.ExpiredAt))
	} else {
		if err := f.UI.ParseError(node.LinkGroup, verErr); err != nil {
			return err
		}
	}

	if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), f); err != nil {
		return s.handleVerificationError(w, r, f, nil, err)
	}

	if ft == flow.TypeBrowser {
		http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowVerificationUI(r.Context())).String(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()),
			verification.RouteGetFlow), url.Values{"id": {f.ID.String()}}).String(), http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) SendVerificationEmail(ctx context.Context, f *verification.Flow, i *identity.Identity, a *identity.VerifiableAddress) error {
	token := NewSelfServiceVerificationToken(a, f, s.d.Config().SelfServiceLinkMethodLifespan(ctx))
	if err := s.d.VerificationTokenPersister().CreateVerificationToken(ctx, token); err != nil {
		return err
	}

	return s.d.LinkSender().SendVerificationTokenTo(ctx, f, i, a, token)
}
