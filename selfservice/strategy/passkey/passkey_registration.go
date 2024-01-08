// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	jsonschema "github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/webauthnx"
	"github.com/ory/x/randx"
)

// Update Registration Flow with Passkey Method
//
// swagger:model updateRegistrationFlowWithPasskeyMethod
type updateRegistrationFlowWithPasskeyMethod struct {
	// Register a WebAuthn Security Key
	//
	// It is expected that the JSON returned by the WebAuthn registration process
	// is included here.
	Register string `json:"passkey_register"`

	// CSRFToken is the anti-CSRF token
	CSRFToken string `json:"csrf_token"`

	// The identity's traits
	//
	// required: true
	Traits json.RawMessage `json:"traits"`

	// Method
	//
	// Should be set to "passkey" when trying to add, update, or remove a Passkey.
	//
	// required: true
	Method string `json:"method"`

	// Flow is flow ID.
	//
	// swagger:ignore
	Flow string `json:"flow"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty"`
}

func (s *Strategy) RegisterRegistrationRoutes(r *x.RouterPublic) {
	webauthnx.RegisterWebauthnRoute(r)
}

func (s *Strategy) handleRegistrationError(_ http.ResponseWriter, r *http.Request, f *registration.Flow, p *updateRegistrationFlowWithPasskeyMethod, err error) error {
	if f != nil {
		if p != nil {
			for _, n := range container.NewFromJSON("", node.DefaultGroup, p.Traits, "traits").Nodes {
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

func (s *Strategy) decode(r *http.Request) (*updateRegistrationFlowWithPasskeyMethod, error) {
	var p updateRegistrationFlowWithPasskeyMethod
	err := registration.DecodeBody(&p, r, s.hd, s.d.Config(), registrationSchema)
	return &p, err
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, regFlow *registration.Flow, i *identity.Identity) (err error) {
	ctx := r.Context()

	if regFlow.Type != flow.TypeBrowser {
		return flow.ErrStrategyNotResponsible
	}

	p, err := s.decode(r)
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, err)
	}

	regFlow.TransientPayload = p.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, regFlow.Type, s.d.Config().DisableAPIFlowEnforcement(ctx), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, err)
	}

	if p.Register == "" && p.Method != "passkey" {
		return flow.ErrStrategyNotResponsible
	}

	if len(p.Register) == 0 {
		return s.addPassKeyNodes(r, w, regFlow, p)
	}

	p.Method = s.ID().String()
	if err := flow.MethodEnabledAndAllowed(ctx, regFlow.GetFlowName(), p.Method, p.Method, s.d); err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, err)
	}

	if len(p.Traits) == 0 {
		p.Traits = json.RawMessage("{}")
	}
	i.Traits = identity.Traits(p.Traits)

	webAuthnSession := gjson.GetBytes(regFlow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if !webAuthnSession.IsObject() {
		return s.handleRegistrationError(w, r, regFlow, p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object.")))
	}

	var webAuthnSess webauthn.SessionData
	if err := json.Unmarshal([]byte(webAuthnSession.Raw), &webAuthnSess); err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object but got: %s", err)))
	}

	webAuthnResponse, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(p.Register))
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, errors.WithStack(
			herodot.ErrBadRequest.WithReasonf("Unable to parse WebAuthn response: %s", err)))
	}

	webAuthn, err := webauthn.New(s.d.Config().PasskeyConfig(ctx))
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Unable to get webAuthn config.").WithDebug(err.Error())))
	}

	credential, err := webAuthn.CreateCredential(webauthnx.NewUser(webAuthnSess.UserID, nil, webAuthn.Config), webAuthnSess, webAuthnResponse)
	if err != nil {
		if devErr := new(protocol.Error); errors.As(err, &devErr) {
			s.d.Logger().WithError(err).WithField("error_devinfo", devErr.DevInfo).Error("Failed to create WebAuthn credential")
		}
		return s.handleRegistrationError(w, r, regFlow, p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Unable to create WebAuthn credential: %s", err)))
	}

	credentialWebAuthn := identity.CredentialFromWebAuthn(credential, true)
	credentialWebAuthnConfig, err := json.Marshal(identity.CredentialsWebAuthnConfig{
		Credentials: identity.CredentialsWebAuthn{*credentialWebAuthn},
		UserHandle:  webAuthnSess.UserID,
	})
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Unable to encode identity credentials.").WithDebug(err.Error())))
	}

	i.UpsertCredentialsConfig(s.ID(), credentialWebAuthnConfig, 1)
	passkeyCred, _ := i.GetCredentials(s.ID())
	passkeyCred.Identifiers = []string{string(webAuthnSess.UserID)}
	i.SetCredentials(s.ID(), *passkeyCred)
	if err := s.validateCredentials(ctx, i); err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, err)
	}

	// Remove the WebAuthn URL from the internal context now that it is set!
	regFlow.InternalContext, err = sjson.DeleteBytes(regFlow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, err)
	}

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, regFlow); err != nil {
		return s.handleRegistrationError(w, r, regFlow, p, err)
	}

	return nil
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, regFlow *registration.Flow) error {
	ctx := r.Context()

	if regFlow.Type != flow.TypeBrowser {
		return nil
	}

	defaultSchemaURL, err := s.d.Config().DefaultIdentityTraitsSchemaURL(ctx)
	if err != nil {
		return err
	}
	nodes, err := s.populateRegistrationNodes(ctx, defaultSchemaURL)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		regFlow.UI.SetNode(n)
	}

	regFlow.UI.Nodes.Append(node.NewInputField(
		"method",
		"passkey",
		node.PasskeyGroup,
		node.InputAttributeTypeSubmit,
	).WithMetaLabel(text.NewInfoSelfServiceRegistrationRegisterPasskey()))

	regFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	return nil
}

func (s *Strategy) populateRegistrationNodes(ctx context.Context, schemaURL *url.URL) (node.Nodes, error) {
	runner, err := schema.NewExtensionRunner(ctx)
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	runner.Register(c)

	nodes, err := container.NodesFromJSONSchema(ctx, node.DefaultGroup, schemaURL.String(), "", c)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// webauthnIdentifierNode returns the node that is used to identify the user in the WebAuthn flow.
func (s *Strategy) webauthnIdentifierNode(ctx context.Context, schemaURL *url.URL) (*node.Node, error) {
	nodes, err := s.populateRegistrationNodes(ctx, schemaURL)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if attr, ok := n.Attributes.(*node.InputAttributes); ok {
			if attr.DataWebauthnIdentifier {
				return n, nil
			}
		}
	}

	return nil, schema.NewMissingIdentifierError()
}

func (s *Strategy) addPassKeyNodes(r *http.Request, w http.ResponseWriter, regFlow *registration.Flow, params *updateRegistrationFlowWithPasskeyMethod) error {
	ctx := r.Context()

	defaultSchemaURL, err := s.d.Config().DefaultIdentityTraitsSchemaURL(ctx)
	if err != nil {
		return err
	}
	idNode, err := s.webauthnIdentifierNode(ctx, defaultSchemaURL)
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, err)
	}

	// Render default nodes as hidden fields, also create passkey
	c, err := container.NewFromStruct("", node.DefaultGroup, params.Traits, "traits")
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, err)
	}
	var identifier string
	for _, n := range c.Nodes {
		regFlow.UI.SetValue(n.ID(), n)
		if n.ID() == idNode.ID() {
			identifier, _ = n.Attributes.GetValue().(string)
		}
	}

	regFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	webAuthn, err := webauthn.New(s.d.Config().PasskeyConfig(ctx))
	if err != nil {
		return errors.WithStack(err)
	}
	user := &webauthnx.User{
		Name:   identifier,
		ID:     []byte(randx.MustString(64, randx.AlphaNum)),
		Config: s.d.Config().PasskeyConfig(ctx),
	}
	option, sessionData, err := webAuthn.BeginRegistration(user)
	if err != nil {
		return errors.WithStack(err)
	}

	injectWebAuthnOptions, err := json.Marshal(option)
	if err != nil {
		return errors.WithStack(err)
	}

	regFlow.InternalContext, err = sjson.SetBytes(
		regFlow.InternalContext,
		flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData),
		sessionData,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	regFlow.UI.Nodes.Append(webauthnx.NewWebAuthnScript(s.d.Config().SelfPublicURL(ctx)))

	regFlow.UI.Nodes.Upsert(&node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name: node.PasskeyRegister,
			Type: node.InputAttributeTypeHidden,
		}})

	regFlow.UI.Nodes.Upsert(&node.Node{
		Type:  node.Input,
		Group: node.WebAuthnGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name:       node.PasskeyCreateData,
			Type:       node.InputAttributeTypeHidden,
			FieldValue: string(injectWebAuthnOptions),
		}})

	redirectTo := regFlow.AppendTo(s.d.Config().SelfServiceFlowRegistrationUI(r.Context())).String()

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(r.Context(), regFlow); err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, err)
	}

	x.AcceptToRedirectOrJSON(w, r, s.d.Writer(), err, redirectTo)
	return flow.ErrCompletedByStrategy
}

func (s *Strategy) validateCredentials(ctx context.Context, i *identity.Identity) error {
	if err := s.d.IdentityValidator().Validate(ctx, i); err != nil {
		return err
	}

	c := i.GetCredentialsOr(identity.CredentialsTypePasskey, &identity.Credentials{})
	if len(c.Identifiers) == 0 {
		return schema.NewMissingIdentifierError()
	}

	return nil
}
