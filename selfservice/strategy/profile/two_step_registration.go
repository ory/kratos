// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

//go:embed .schema/registration.schema.json
var registrationSchema []byte

var _ registration.Strategy = new(Strategy)

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeProfile
}

func (s *Strategy) RegisterRegistrationRoutes(*x.RouterPublic) {}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	if !s.d.Config().SelfServiceFlowRegistrationTwoSteps(r.Context()) {
		return nil
	}

	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	nodes, err := container.NodesFromJSONSchema(r.Context(), node.DefaultGroup, ds.String(), "", nil)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		f.UI.SetNode(n)
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.Nodes.Append(node.NewInputField(
		"method",
		"profile",
		node.ProfileGroup,
		node.InputAttributeTypeSubmit,
	).WithMetaLabel(text.NewInfoRegistration()))

	return nil
}

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
	Screen string `json:"screen" form:"screen"`

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

func (s *Strategy) decode(p *updateRegistrationFlowWithProfileMethod, r *http.Request) error {
	return registration.DecodeBody(p, r, s.dc, s.d.Config(), registrationSchema)
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, regFlow *registration.Flow, i *identity.Identity) (err error) {
	if !s.d.Config().SelfServiceFlowRegistrationTwoSteps(r.Context()) {
		return flow.ErrStrategyNotResponsible
	}

	var params updateRegistrationFlowWithProfileMethod

	if err = s.decode(&params, r); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &params, err)
	}

	if params.Screen == "credential-selection" {
		params.Method = "profile"
	}

	switch params.Method {
	case "profile":
		return s.displayStepTwoNodes(w, r, regFlow, i, params)
	case "profile:back":
		return s.displayStepOneNodes(w, r, regFlow, i, params)
	}
	// Default case
	return flow.ErrStrategyNotResponsible
}

func (s *Strategy) displayStepOneNodes(w http.ResponseWriter, r *http.Request, regFlow *registration.Flow, _ *identity.Identity, params updateRegistrationFlowWithProfileMethod) error {
	ctx := r.Context()
	regFlow.UI.ResetMessages()
	err := json.Unmarshal([]byte(gjson.GetBytes(regFlow.InternalContext, "stepOneNodes").Raw), &regFlow.UI.Nodes)
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, &params, err)
	}
	regFlow.UI.UpdateNodeValuesFromJSON(params.Traits, "traits", node.DefaultGroup)

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, regFlow); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &params, err)
	}

	redirectTo := regFlow.AppendTo(s.d.Config().SelfServiceFlowRegistrationUI(ctx)).String()
	if x.IsJSONRequest(r) {
		s.d.Writer().WriteCode(w, r, http.StatusBadRequest, regFlow)
	} else {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}

	return flow.ErrCompletedByStrategy
}

func (s *Strategy) displayStepTwoNodes(w http.ResponseWriter, r *http.Request, regFlow *registration.Flow, i *identity.Identity, params updateRegistrationFlowWithProfileMethod) error {
	ctx := r.Context()

	// Reset state-esque flow fields
	regFlow.Active = ""
	regFlow.State = "choose_method"

	regFlow.UI.ResetMessages()
	regFlow.TransientPayload = params.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, regFlow.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, params.CSRFToken); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &params, err)
	}

	if len(params.Traits) == 0 {
		params.Traits = json.RawMessage("{}")
	}
	i.Traits = identity.Traits(params.Traits)
	if err := s.d.IdentityValidator().Validate(ctx, i); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &params, err)
	}

	err := json.Unmarshal([]byte(gjson.GetBytes(regFlow.InternalContext, "stepTwoNodes").Raw), &regFlow.UI.Nodes)
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, &params, err)
	}

	regFlow.UI.Messages.Add(text.NewInfoSelfServiceChooseCredentials())

	regFlow.UI.Nodes.Append(node.NewInputField(
		"method",
		"profile:back",
		node.ProfileGroup,
		node.InputAttributeTypeSubmit,
	).WithMetaLabel(text.NewInfoRegistrationBack()))

	regFlow.UI.UpdateNodeValuesFromJSON(json.RawMessage(i.Traits), "traits", node.DefaultGroup)
	for _, n := range regFlow.UI.Nodes {
		if n.Group != node.DefaultGroup || n.Type != node.Input {
			continue
		}
		if attr, ok := n.Attributes.(*node.InputAttributes); ok {
			attr.Type = node.InputAttributeTypeHidden
		}
	}

	if regFlow.Type == flow.TypeBrowser {
		regFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	}

	if err = s.d.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, regFlow); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &params, err)
	}

	redirectTo := regFlow.AppendTo(s.d.Config().SelfServiceFlowRegistrationUI(ctx)).String()
	if x.IsJSONRequest(r) {
		s.d.Writer().WriteCode(w, r, http.StatusBadRequest, regFlow)
	} else {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}

	return flow.ErrCompletedByStrategy
}

func (s *Strategy) handleRegistrationError(_ http.ResponseWriter, r *http.Request, regFlow *registration.Flow, params *updateRegistrationFlowWithProfileMethod, err error) error {
	if regFlow != nil {
		if params != nil {
			for _, n := range container.NewFromJSON("", node.ProfileGroup, params.Traits, "traits").Nodes {
				// we only set the value and not the whole field because we want to keep types from the initial form generation
				regFlow.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
			}
		}

		if regFlow.Type == flow.TypeBrowser {
			regFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}
