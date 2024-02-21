// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/ory/herodot"
	"github.com/ory/x/otelx"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/urlx"
)

var _ registration.Strategy = new(Strategy)

// Update Registration Flow with Code Method
//
// swagger:model updateRegistrationFlowWithCodeMethod
type updateRegistrationFlowWithCodeMethod struct {
	// The identity's traits
	//
	// required: true
	Traits json.RawMessage `json:"traits" form:"traits"`

	// The OTP Code sent to the user
	//
	// required: false
	Code string `json:"code" form:"code"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token" form:"csrf_token"`

	// Method to use
	//
	// This field must be set to `code` when using the code method.
	//
	// required: true
	Method string `json:"method" form:"method"`

	// Resend restarts the flow with a new code
	//
	// required: false
	Resend string `json:"resend" form:"resend"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (p *updateRegistrationFlowWithCodeMethod) GetResend() string {
	return p.Resend
}

func (s *Strategy) RegisterRegistrationRoutes(*x.RouterPublic) {}

func (s *Strategy) HandleRegistrationError(ctx context.Context, r *http.Request, f *registration.Flow, body *updateRegistrationFlowWithCodeMethod, err error) error {
	if errors.Is(err, flow.ErrCompletedByStrategy) {
		return err
	}

	if f != nil {
		if body != nil {
			action := f.AppendTo(urlx.AppendPaths(s.deps.Config().SelfPublicURL(ctx), registration.RouteSubmitFlow)).String()
			for _, n := range container.NewFromJSON(action, node.CodeGroup, body.Traits, "traits").Nodes {
				// we only set the value and not the whole field because we want to keep types from the initial form generation
				f.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
			}
		}

		f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
	}

	return err
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, rf *registration.Flow) error {
	return s.PopulateMethod(r, rf)
}

type options func(*identity.Identity) error

func WithCredentials(via identity.CodeAddressType, usedAt sql.NullTime) options {
	return func(i *identity.Identity) error {
		return i.SetCredentialsWithConfig(identity.CredentialsTypeCodeAuth, identity.Credentials{Type: identity.CredentialsTypePassword, Identifiers: []string{}}, &identity.CredentialsCode{AddressType: via, UsedAt: usedAt})
	}
}

func (s *Strategy) handleIdentityTraits(ctx context.Context, f *registration.Flow, traits, transientPayload json.RawMessage, i *identity.Identity, opts ...options) error {
	f.TransientPayload = transientPayload
	if len(traits) == 0 {
		traits = json.RawMessage("{}")
	}

	// we explicitly set the Code credentials type
	i.Traits = identity.Traits(traits)
	if err := i.SetCredentialsWithConfig(s.ID(), identity.Credentials{Type: s.ID(), Identifiers: []string{}}, &identity.CredentialsCode{UsedAt: sql.NullTime{}}); err != nil {
		return err
	}

	for _, opt := range opts {
		if err := opt(i); err != nil {
			return err
		}
	}

	// Validate the identity
	if err := s.deps.IdentityValidator().Validate(ctx, i); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) getCredentialsFromTraits(ctx context.Context, f *registration.Flow, i *identity.Identity, traits, transientPayload json.RawMessage) (*identity.Credentials, error) {
	if err := s.handleIdentityTraits(ctx, f, traits, transientPayload, i); err != nil {
		return nil, errors.WithStack(err)
	}

	cred, ok := i.GetCredentials(identity.CredentialsTypeCodeAuth)
	if !ok {
		return nil, errors.WithStack(schema.NewMissingIdentifierError())
	} else if len(cred.Identifiers) == 0 {
		return nil, errors.WithStack(schema.NewMissingIdentifierError())
	}
	return cred, nil
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	ctx, span := s.deps.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.code.strategy.Register")
	defer otelx.End(span, &err)

	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.ID().String(), s.deps); err != nil {
		return err
	}

	var p updateRegistrationFlowWithCodeMethod
	if err := registration.DecodeBody(&p, r, s.dx, s.deps.Config(), registrationSchema); err != nil {
		return s.HandleRegistrationError(ctx, r, f, &p, err)
	}

	f.TransientPayload = p.TransientPayload

	if err := flow.EnsureCSRF(s.deps, r, f.Type, s.deps.Config().DisableAPIFlowEnforcement(ctx), s.deps.GenerateCSRFToken, p.CSRFToken); err != nil {
		return s.HandleRegistrationError(ctx, r, f, &p, err)
	}

	// By Default the flow should be in the 'choose method' state.
	SetDefaultFlowState(f, p.Resend)

	switch f.GetState() {
	case flow.StateChooseMethod:
		return s.HandleRegistrationError(ctx, r, f, &p, s.registrationSendEmail(ctx, w, r, f, &p, i))
	case flow.StateEmailSent:
		return s.HandleRegistrationError(ctx, r, f, &p, s.registrationVerifyCode(ctx, f, &p, i))
	case flow.StatePassedChallenge:
		return s.HandleRegistrationError(ctx, r, f, &p, errors.WithStack(schema.NewNoRegistrationStrategyResponsible()))
	}

	return s.HandleRegistrationError(ctx, r, f, &p, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unexpected flow state: %s", f.GetState())))
}

func (s *Strategy) registrationSendEmail(ctx context.Context, w http.ResponseWriter, r *http.Request, f *registration.Flow, p *updateRegistrationFlowWithCodeMethod, i *identity.Identity) (err error) {
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.strategy.registrationSendEmail")
	defer otelx.End(span, &err)

	if len(p.Traits) == 0 {
		return errors.WithStack(schema.NewRequiredError("#/traits", "traits"))
	}

	// Create the Registration code

	// Step 1: validate the identity's traits
	cred, err := s.getCredentialsFromTraits(ctx, f, i, p.Traits, p.TransientPayload)
	if err != nil {
		return err
	}

	// Step 2: Delete any previous registration codes for this flow ID
	if err := s.deps.RegistrationCodePersister().DeleteRegistrationCodesOfFlow(ctx, f.ID); err != nil {
		return errors.WithStack(err)
	}

	// Step 3: Get the identity email and send the code
	var addresses []Address
	for _, identifier := range cred.Identifiers {
		addresses = append(addresses, Address{To: identifier, Via: identity.AddressTypeEmail})
	}
	// kratos only supports `email` identifiers at the moment with the code method
	// this is validated in the identity validation step above
	if err := s.deps.CodeSender().SendCode(ctx, f, i, addresses...); err != nil {
		return errors.WithStack(err)
	}

	// sets the flow state to code sent
	f.SetState(flow.NextState(f.GetState()))

	// Step 4: Generate the UI for the `code` input form
	// re-initialize the UI with a "clean" new state
	// this should also provide a "resend" button and an option to change the email address
	if err := s.NewCodeUINodes(r, f, p.Traits); err != nil {
		return errors.WithStack(err)
	}

	f.Active = identity.CredentialsTypeCodeAuth
	if err := s.deps.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, f); err != nil {
		return errors.WithStack(err)
	}

	if x.IsJSONRequest(r) {
		s.deps.Writer().WriteCode(w, r, http.StatusBadRequest, f)
	} else {
		http.Redirect(w, r, f.AppendTo(s.deps.Config().SelfServiceFlowRegistrationUI(ctx)).String(), http.StatusSeeOther)
	}

	// we return an error to the flow handler so that it does not continue execution of the hooks.
	// we are not done with the registration flow yet. The user needs to verify the code and then we need to persist the identity.
	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) registrationVerifyCode(ctx context.Context, f *registration.Flow, p *updateRegistrationFlowWithCodeMethod, i *identity.Identity) (err error) {
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.strategy.registrationVerifyCode")
	defer otelx.End(span, &err)

	if len(p.Code) == 0 {
		return errors.WithStack(schema.NewRequiredError("#/code", "code"))
	}

	if len(p.Traits) == 0 {
		return errors.WithStack(schema.NewRequiredError("#/traits", "traits"))
	}

	// Step 1: Re-validate the identity's traits
	// this is important since the client could have switched out the identity's traits
	// this method also returns the credentials for a temporary identity
	cred, err := s.getCredentialsFromTraits(ctx, f, i, p.Traits, p.TransientPayload)
	if err != nil {
		return err
	}

	// Step 2: Check if the flow traits match the identity traits
	for _, n := range container.NewFromJSON("", node.DefaultGroup, p.Traits, "traits").Nodes {
		if f.GetUI().GetNodes().Find(n.ID()).Attributes.GetValue() != n.Attributes.GetValue() {
			return errors.WithStack(schema.NewTraitsMismatch())
		}
	}

	// Step 3: Attempt to use the code
	registrationCode, err := s.deps.RegistrationCodePersister().UseRegistrationCode(ctx, f.ID, p.Code, cred.Identifiers...)
	if err != nil {
		if errors.Is(err, ErrCodeNotFound) {
			return errors.WithStack(schema.NewRegistrationCodeInvalid())
		}
		return errors.WithStack(err)
	}

	// Step 4: The code was correct, populate the Identity credentials and traits
	if err := s.handleIdentityTraits(ctx, f, p.Traits, p.TransientPayload, i, WithCredentials(registrationCode.AddressType, registrationCode.UsedAt)); err != nil {
		return errors.WithStack(err)
	}

	// since nothing has errored yet, we can assume that the code is correct
	// and we can update the registration flow
	f.SetState(flow.NextState(f.GetState()))

	if err := s.deps.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, f); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
