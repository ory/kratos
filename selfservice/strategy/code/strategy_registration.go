// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"

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
var _ registration.FormHydrator = new(Strategy)

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
			for _, n := range container.NewFromJSON(action, node.DefaultGroup, body.Traits, "traits").Nodes {
				// we only set the value and not the whole field because we want to keep types from the initial form generation
				f.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
			}
		}

		f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
	}

	return err
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	if !s.deps.Config().SelfServiceCodeStrategy(r.Context()).PasswordlessEnabled {
		return nil
	}

	f.GetUI().Nodes.Append(nodeSubmitRegistration())
	return nil
}

func (s *Strategy) PopulateRegistrationMethodCredentials(r *http.Request, f *registration.Flow, options ...registration.FormHydratorModifier) error {
	if !s.deps.Config().SelfServiceCodeStrategy(r.Context()).PasswordlessEnabled {
		return nil
	}

	f.GetUI().Nodes.RemoveMatching(nodeRegistrationResendNode())
	f.GetUI().Nodes.RemoveMatching(nodeRegistrationSelectCredentialsNode())
	f.GetUI().Nodes.RemoveMatching(nodeContinueButton())
	f.GetUI().Nodes.RemoveMatching(nodeCodeInputFieldHidden())
	f.GetUI().Nodes.RemoveMatching(nodeCodeInputField())

	f.GetUI().Nodes.Append(nodeSubmitRegistration())
	f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
	return nil
}

func (s *Strategy) PopulateRegistrationMethodProfile(r *http.Request, f *registration.Flow, options ...registration.FormHydratorModifier) error {
	if !s.deps.Config().SelfServiceCodeStrategy(r.Context()).PasswordlessEnabled {
		return nil
	}

	f.GetUI().Nodes.RemoveMatching(nodeSubmitRegistration())
	f.GetUI().Nodes.RemoveMatching(nodeRegistrationResendNode())
	f.GetUI().Nodes.RemoveMatching(nodeRegistrationSelectCredentialsNode())
	f.GetUI().Nodes.RemoveMatching(nodeContinueButton())
	f.GetUI().Nodes.RemoveMatching(nodeCodeInputFieldHidden())
	f.GetUI().Nodes.RemoveMatching(nodeCodeInputField())

	f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
	return nil
}

func (s *Strategy) validateTraits(ctx context.Context, traits json.RawMessage, i *identity.Identity) error {
	i.Traits = []byte("{}")
	if gjson.ValidBytes(traits) {
		i.Traits = identity.Traits(traits)
	}

	// Validate the identity
	if err := s.deps.IdentityValidator().Validate(ctx, i); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) validateAndGetCredentialsFromTraits(ctx context.Context, i *identity.Identity, traits json.RawMessage) (*identity.Credentials, *identity.CredentialsCode, error) {
	if err := s.validateTraits(ctx, traits, i); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	cred, ok := i.GetCredentials(identity.CredentialsTypeCodeAuth)
	if !ok {
		return nil, nil, errors.WithStack(schema.NewMissingIdentifierError())
	} else if len(strings.Join(cred.Identifiers, "")) == 0 {
		return nil, nil, errors.WithStack(schema.NewMissingIdentifierError())
	}

	var conf identity.CredentialsCode
	if len(cred.Config) > 0 {
		if err := json.Unmarshal(cred.Config, &conf); err != nil {
			return nil, nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to unmarshal credentials config: %s", err))
		}
	}

	return cred, &conf, nil
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	ctx, span := s.deps.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.code.Strategy.Register")
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
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.Strategy.registrationSendEmail")
	defer otelx.End(span, &err)

	if len(p.Traits) == 0 {
		return errors.WithStack(schema.NewRequiredError("#/traits", "traits"))
	}

	// Create the Registration code

	// Step 1: validate the identity's traits
	_, conf, err := s.validateAndGetCredentialsFromTraits(ctx, i, p.Traits)
	if err != nil {
		return err
	}

	// Step 2: Delete any previous registration codes for this flow ID
	if err := s.deps.RegistrationCodePersister().DeleteRegistrationCodesOfFlow(ctx, f.ID); err != nil {
		return errors.WithStack(err)
	}

	// Step 3: Get the identity email and send the code
	var addresses []Address
	for _, address := range conf.Addresses {
		addresses = append(addresses, Address{To: address.Address, Via: address.Channel})
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
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.Strategy.registrationVerifyCode")
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
	cred, _, err := s.validateAndGetCredentialsFromTraits(ctx, i, p.Traits)
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

	// Step 4: Verify the address
	if err := s.verifyAddress(ctx, i, Address{
		To:  registrationCode.Address,
		Via: registrationCode.AddressType,
	}); err != nil {
		return err
	}

	// since nothing has errored yet, we can assume that the code is correct
	// and we can update the registration flow
	f.SetState(flow.NextState(f.GetState()))

	if err := s.deps.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, f); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
