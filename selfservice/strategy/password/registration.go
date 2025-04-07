// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ory/x/otelx/semconv"

	"github.com/ory/x/otelx"

	"github.com/ory/kratos/text"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

func nodePasswordInput() *node.Node {
	return NewPasswordNode("password", node.InputAttributeAutocompleteNewPassword)
}

func nodeSubmit() *node.Node {
	return node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoRegistration())
}

var _ registration.FormHydrator = new(Strategy)

// Update Registration Flow with Password Method
//
// swagger:model updateRegistrationFlowWithPasswordMethod
type UpdateRegistrationFlowWithPasswordMethod struct {
	// Password to sign the user up with
	//
	// required: true
	Password string `json:"password"`

	// The identity's traits
	//
	// required: true
	Traits json.RawMessage `json:"traits"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token"`

	// Method to use
	//
	// This field must be set to `password` when using the password method.
	//
	// required: true
	Method string `json:"method"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (s *Strategy) RegisterRegistrationRoutes(*x.RouterPublic) {
}

func (s *Strategy) handleRegistrationError(r *http.Request, f *registration.Flow, p UpdateRegistrationFlowWithPasswordMethod, err error) error {
	if f != nil {
		for _, n := range container.NewFromJSON("", node.DefaultGroup, p.Traits, "traits").Nodes {
			// we only set the value and not the whole field because we want to keep types from the initial form generation
			f.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
		}

		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

func (s *Strategy) decode(p *UpdateRegistrationFlowWithPasswordMethod, r *http.Request) (err error) {
	return registration.DecodeBody(p, r, s.hd, s.d.Config(), registrationSchema)
}

func (s *Strategy) Register(_ http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.password.Strategy.Register")
	defer otelx.End(span, &err)

	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.ID().String(), s.d); err != nil {
		return err
	}

	var p UpdateRegistrationFlowWithPasswordMethod
	if err := s.decode(&p, r); err != nil {
		return s.handleRegistrationError(r, f, p, err)
	}

	f.TransientPayload = p.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(ctx), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return s.handleRegistrationError(r, f, p, err)
	}

	if len(p.Password) == 0 {
		return s.handleRegistrationError(r, f, p, schema.NewRequiredError("#/password", "password"))
	}

	if len(p.Traits) == 0 {
		p.Traits = json.RawMessage("{}")
	}

	hpw := make(chan []byte)
	errC := make(chan error)
	go func() {
		defer close(hpw)
		defer close(errC)

		h, err := s.d.Hasher(ctx).Generate(ctx, []byte(p.Password))
		if err != nil {
			errC <- err
			return
		}
		hpw <- h
	}()

	if err != nil {
		return s.handleRegistrationError(r, f, p, err)
	}

	i.Traits = identity.Traits(p.Traits)
	// We have to set the credential here, so the identity validator can populate the identifiers.
	// The password hash is computed in parallel and set later.
	if err := i.SetCredentialsWithConfig(s.ID(), identity.Credentials{Type: s.ID(), Identifiers: []string{}}, json.RawMessage("{}")); err != nil {
		return s.handleRegistrationError(r, f, p, err)
	}

	if err := s.validateCredentials(ctx, i, p.Password); err != nil {
		return s.handleRegistrationError(r, f, p, err)
	}

	select {
	case err := <-errC:
		return s.handleRegistrationError(r, f, p, err)
	case h := <-hpw:
		co, err := json.Marshal(&identity.CredentialsPassword{HashedPassword: string(h)})
		if err != nil {
			return s.handleRegistrationError(r, f, p, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode password options to JSON: %s", err)))
		}
		i.UpsertCredentialsConfig(s.ID(), co, 0)
	}

	return nil
}

func (s *Strategy) validateCredentials(ctx context.Context, i *identity.Identity, pw string) (err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.password.Strategy.validateCredentials")
	defer otelx.End(span, &err)

	if err := s.d.IdentityValidator().Validate(ctx, i); err != nil {
		return err
	}

	c, ok := i.GetCredentials(identity.CredentialsTypePassword)
	if !ok {
		// This should never happen
		return errors.WithStack(x.PseudoPanic.WithReasonf("identity object did not provide the %s CredentialType unexpectedly", identity.CredentialsTypePassword))
	} else if len(c.Identifiers) == 0 {
		return schema.NewMissingIdentifierError()
	}

	for _, id := range c.Identifiers {
		if err := s.d.PasswordValidator().Validate(ctx, id, pw); err != nil {
			if herodotErr := new(herodot.DefaultError); errors.As(err, &herodotErr) {
				return err
			}
			if message := new(text.Message); errors.As(err, &message) {
				return schema.NewPasswordPolicyViolationError("#/password", message)
			}
			return schema.NewPasswordPolicyViolationError("#/password", text.NewErrorValidationPasswordPolicyViolationGeneric(err.Error()))
		}
	}

	return nil
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) (err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.password.Strategy.PopulateRegistrationMethod")
	defer otelx.End(span, &err)

	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	// The group used to be `password`, but to make it consistent with other methods and two-step registration,
	// it is now `default`. To make this switch backwards compatible, this feature flag is used.
	//
	// Previously, the behavior would be that the group of NodesFromJSONSchema would be `password`, but as soon
	// as any other method (code, passkeys, two-step) would be enabled, the group would be `default`.
	//
	// Going forward, the default is that the group is `default` and the feature flag is not set.
	//
	// TODO remove me when everyone has migrated.
	group := node.DefaultGroup
	if !s.d.Config().SelfServiceFlowRegistrationTwoSteps(r.Context()) && node.UiNodeGroup(s.d.Config().SelfServiceFlowRegistrationPasswordMethodProfileGroup(r.Context())) == node.PasswordGroup {
		span.AddEvent(semconv.NewDeprecatedFeatureUsedEvent(ctx, "password_profile_registration_node_group=password"))

		// This is the legacy code path. In the new code path, the profile method is responsible for hydrating the form
		// nodes. In the old code path, the password method is responsible for hydrating the form nodes if it is
		// the only method enabled.
		group = node.PasswordGroup
		nodes, err := container.NodesFromJSONSchema(r.Context(), group, ds.String(), "", nil)
		if err != nil {
			return err
		}

		for _, n := range nodes {
			f.UI.SetNode(n)
		}
	}
	// TODO end

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.Nodes.Upsert(nodePasswordInput())
	f.UI.Nodes.Upsert(nodeSubmit())

	return nil
}

func (s *Strategy) PopulateRegistrationMethodCredentials(r *http.Request, f *registration.Flow, options ...registration.FormHydratorModifier) error {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.Nodes.Upsert(nodePasswordInput())
	f.UI.Nodes.Upsert(nodeSubmit())
	return nil
}

func (s *Strategy) PopulateRegistrationMethodProfile(r *http.Request, f *registration.Flow, options ...registration.FormHydratorModifier) error {
	// The profile method is responsible for rendering the profile form fields.
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.Nodes.RemoveMatching(nodePasswordInput())
	f.UI.Nodes.RemoveMatching(nodeSubmit())
	return nil
}
