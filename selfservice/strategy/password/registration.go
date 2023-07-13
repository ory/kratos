// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import (
	"context"
	"encoding/json"
	"net/http"

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
	"github.com/ory/x/errorsx"
)

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
	TransientPayload json.RawMessage `json:"transient_payload,omitempty"`
}

func (s *Strategy) RegisterRegistrationRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) handleRegistrationError(_ http.ResponseWriter, r *http.Request, f *registration.Flow, p *UpdateRegistrationFlowWithPasswordMethod, err error) error {
	if f != nil {
		if p != nil {
			for _, n := range container.NewFromJSON("", node.ProfileGroup, p.Traits, "traits").Nodes {
				// we only set the value and not the whole field because we want to keep types from the initial form generation
				f.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
			}
		}

		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

func (s *Strategy) decode(p *UpdateRegistrationFlowWithPasswordMethod, r *http.Request) error {
	return registration.DecodeBody(p, r, s.hd, s.d.Config(), registrationSchema)
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.ID().String(), s.d); err != nil {
		return err
	}

	var p UpdateRegistrationFlowWithPasswordMethod
	if err := s.decode(&p, r); err != nil {
		return s.handleRegistrationError(w, r, f, &p, err)
	}

	f.TransientPayload = p.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return s.handleRegistrationError(w, r, f, &p, err)
	}

	if len(p.Password) == 0 {
		return s.handleRegistrationError(w, r, f, &p, schema.NewRequiredError("#/password", "password"))
	}

	if len(p.Traits) == 0 {
		p.Traits = json.RawMessage("{}")
	}

	hpw, err := s.d.Hasher(r.Context()).Generate(r.Context(), []byte(p.Password))
	if err != nil {
		return s.handleRegistrationError(w, r, f, &p, err)
	}

	i.Traits = identity.Traits(p.Traits)
	if err := i.SetCredentialsWithConfig(s.ID(), identity.Credentials{Type: s.ID(), Identifiers: []string{}}, &identity.CredentialsPassword{HashedPassword: string(hpw)}); err != nil {
		return s.handleRegistrationError(w, r, f, &p, err)
	}

	if err := s.validateCredentials(r.Context(), i, p.Password); err != nil {
		return s.handleRegistrationError(w, r, f, &p, err)
	}

	return nil
}

func (s *Strategy) validateCredentials(ctx context.Context, i *identity.Identity, pw string) error {
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
			if _, ok := errorsx.Cause(err).(*herodot.DefaultError); ok {
				return err
			}
			return schema.NewPasswordPolicyViolationError("#/password", err.Error())
		}
	}

	return nil
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	nodes, err := container.NodesFromJSONSchema(r.Context(), node.PasswordGroup, ds.String(), "", nil)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		f.UI.SetNode(n)
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.Nodes.Upsert(NewPasswordNode("password", node.InputAttributeAutocompleteNewPassword))
	f.UI.Nodes.Append(node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoRegistration()))

	return nil
}
