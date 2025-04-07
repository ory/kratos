// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
	"github.com/ory/x/otelx/semconv"
)

//go:embed .schema/registration.schema.json
var registrationSchema []byte

// The RegistrationScreen
// swagger:enum RegistrationScreen
type RegistrationScreen string

const (
	//nolint:gosec // not a credential
	RegistrationScreenCredentialSelection RegistrationScreen = "credential-selection"
	RegistrationScreenPrevious            RegistrationScreen = "previous"
)

var _ registration.Strategy = new(Strategy)
var _ registration.FormHydrator = new(Strategy)

// Update Registration Flow with Profile Method
//
// swagger:model updateRegistrationFlowWithProfileMethod
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateRegistrationFlowWithProfileMethod struct {
	// Traits
	//
	// The identity's traits.
	//
	// required: true
	Traits json.RawMessage `json:"traits"`

	// Method
	//
	// Should be set to profile when trying to update a profile.
	//
	// required: true
	Method string `json:"method"`

	// Screen requests navigation to a previous screen.
	//
	// This must be set to credential-selection to go back to the credential
	// selection screen.
	//
	// required: false
	Screen RegistrationScreen `json:"screen" form:"screen"`

	// FlowIDRequestID is the flow ID.
	//
	// swagger:ignore
	FlowID string `json:"flow"`

	// The Anti-CSRF Token
	//
	// This token is only required when performing browser flows.
	CSRFToken string `json:"csrf_token"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty"`
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeProfile
}

func (s *Strategy) RegisterRegistrationRoutes(*x.RouterPublic) {}

func (s *Strategy) PopulateRegistrationMethodCredentials(r *http.Request, f *registration.Flow, options ...registration.FormHydratorModifier) error {
	f.UI.Nodes.Append(nodePreviousScreen())
	f.UI.Nodes.RemoveMatching(nodeSubmitProfile())

	for _, n := range f.UI.Nodes {
		if n.Group != node.DefaultGroup || n.Type != node.Input {
			continue
		}
		if attr, ok := n.Attributes.(*node.InputAttributes); ok {
			attr.Type = node.InputAttributeTypeHidden
		}
	}

	f.UI.Messages.Add(text.NewInfoSelfServiceChooseCredentials())
	return nil
}

func (s *Strategy) PopulateRegistrationMethodProfile(r *http.Request, f *registration.Flow, options ...registration.FormHydratorModifier) error {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	conf := new(registration.FormHydratorOptions)
	for _, o := range options {
		o(conf)
	}

	f.UI.Nodes.RemoveMatching(nodePreviousScreen())
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	nodes, err := container.NodesFromJSONSchema(r.Context(), node.DefaultGroup, ds.String(), "", nil)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		f.UI.Nodes.Upsert(n)
	}

	if len(conf.WithTraits) > 0 {
		f.UI.UpdateNodeValuesFromJSON(conf.WithTraits, "traits", node.DefaultGroup)
	}

	f.UI.Nodes.Append(nodeSubmitProfile())
	return nil
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	nodes, err := container.NodesFromJSONSchema(r.Context(), node.DefaultGroup, ds.String(), "", nil)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		f.UI.SetNode(n)
	}

	return nil
}

func (s *Strategy) decode(p *updateRegistrationFlowWithProfileMethod, r *http.Request) error {
	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(registrationSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.dc.Decode(r, p, compiler, decoderx.HTTPKeepRequestBody(true), decoderx.HTTPDecoderSetValidatePayloads(false), decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return err
	}

	if p.Method != "profile" && p.Method != "profile:back" && len(p.Screen) == 0 {
		return errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	return registration.DecodeBody(p, r, s.dc, s.d.Config(), registrationSchema)
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, regFlow *registration.Flow, i *identity.Identity) (err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.profile.Strategy.Register")
	defer otelx.End(span, &err)

	if !s.d.Config().SelfServiceFlowRegistrationTwoSteps(ctx) {
		return flow.ErrStrategyNotResponsible
	}

	var params updateRegistrationFlowWithProfileMethod

	if err = s.decode(&params, r); err != nil {
		return s.handleRegistrationError(r, regFlow, params, err)
	}

	if params.Method == "profile" || len(params.Screen) > 0 {
		switch params.Screen {
		case RegistrationScreenCredentialSelection:
			return s.showCredentialsSelection(ctx, w, r, regFlow, i, params)
		case RegistrationScreenPrevious:
			return s.returnToProfileForm(ctx, w, r, regFlow, params)
		default:
			// FIXME In this scenario we are on the first step of the registration flow and the user clicked on "continue".
			// FIXME The appropriate solution would be to also have `screen=credential-selection` available, but that
			// FIXME is not the case right now. So instead, we fall back.
			span.AddEvent(semconv.NewDeprecatedFeatureUsedEvent(ctx, "profile:missing_screen_parameter"))
			return s.showCredentialsSelection(ctx, w, r, regFlow, i, params)
		}
	} else if params.Method == "profile:back" {
		// "profile:back" is kept for backwards compatibility.
		// FIXME remove this at some point.
		span.AddEvent(semconv.NewDeprecatedFeatureUsedEvent(ctx, "profile:back"))
		return s.returnToProfileForm(ctx, w, r, regFlow, params)
	}

	return flow.ErrStrategyNotResponsible
}

func (s *Strategy) returnToProfileForm(ctx context.Context, w http.ResponseWriter, r *http.Request, regFlow *registration.Flow, params updateRegistrationFlowWithProfileMethod) error {
	regFlow.UI.ResetMessages()
	regFlow.UI.UpdateNodeValuesFromJSON(params.Traits, "traits", node.DefaultGroup)

	for _, ls := range s.d.RegistrationStrategies(ctx) {
		populator, ok := ls.(registration.FormHydrator)
		if !ok {
			continue
		}

		if err := populator.PopulateRegistrationMethodProfile(r, regFlow, registration.WithTraits(params.Traits)); err != nil {
			return s.handleRegistrationError(r, regFlow, params, err)
		}
	}

	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return s.handleRegistrationError(r, regFlow, params, err)
	}

	if err := registration.SortNodes(r.Context(), regFlow.UI.Nodes, ds.String()); err != nil {
		return s.handleRegistrationError(r, regFlow, params, err)
	}

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, regFlow); err != nil {
		return s.handleRegistrationError(r, regFlow, params, err)
	}

	redirectTo := regFlow.AppendTo(s.d.Config().SelfServiceFlowRegistrationUI(ctx)).String()
	if x.IsJSONRequest(r) {
		s.d.Writer().WriteCode(w, r, http.StatusBadRequest, regFlow)
	} else {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}

	return flow.ErrCompletedByStrategy
}

func (s *Strategy) showCredentialsSelection(ctx context.Context, w http.ResponseWriter, r *http.Request, regFlow *registration.Flow, i *identity.Identity, params updateRegistrationFlowWithProfileMethod) error {
	// Reset state-esque flow fields
	regFlow.Active = ""
	regFlow.State = "choose_method"

	regFlow.UI.ResetMessages()
	regFlow.TransientPayload = params.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, regFlow.Type, s.d.Config().DisableAPIFlowEnforcement(ctx), s.d.GenerateCSRFToken, params.CSRFToken); err != nil {
		return s.handleRegistrationError(r, regFlow, params, err)
	}

	if len(params.Traits) == 0 {
		params.Traits = json.RawMessage("{}")
	}

	i.Traits = identity.Traits(params.Traits)
	if err := s.d.IdentityValidator().Validate(ctx, i); err != nil {
		return s.handleRegistrationError(r, regFlow, params, err)
	}

	for _, ls := range s.d.RegistrationStrategies(ctx) {
		populator, ok := ls.(registration.FormHydrator)
		if !ok {
			continue
		}

		if err := populator.PopulateRegistrationMethodCredentials(r, regFlow); err != nil {
			return s.handleRegistrationError(r, regFlow, params, err)
		}
	}

	regFlow.UI.UpdateNodeValuesFromJSON(json.RawMessage(i.Traits), "traits", node.DefaultGroup)

	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return s.handleRegistrationError(r, regFlow, params, err)
	}

	if err := registration.SortNodes(r.Context(), regFlow.UI.Nodes, ds.String()); err != nil {
		return s.handleRegistrationError(r, regFlow, params, err)
	}

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, regFlow); err != nil {
		return s.handleRegistrationError(r, regFlow, params, err)
	}

	redirectTo := regFlow.AppendTo(s.d.Config().SelfServiceFlowRegistrationUI(ctx)).String()
	if x.IsJSONRequest(r) {
		s.d.Writer().WriteCode(w, r, http.StatusBadRequest, regFlow)
	} else {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}

	return flow.ErrCompletedByStrategy
}

func (s *Strategy) handleRegistrationError(r *http.Request, regFlow *registration.Flow, params updateRegistrationFlowWithProfileMethod, err error) error {
	if regFlow != nil {
		for _, n := range container.NewFromJSON("", node.DefaultGroup, params.Traits, "traits").Nodes {
			// we only set the value and not the whole field because we want to keep types from the initial form generation
			regFlow.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
		}

		regFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	}

	return err
}
