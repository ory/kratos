// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/webauthnx"
)

// Update Registration Flow with WebAuthn Method
//
// swagger:model updateRegistrationFlowWithWebAuthnMethod
type updateRegistrationFlowWithWebAuthnMethod struct {
	// Register a WebAuthn Security Key
	//
	// It is expected that the JSON returned by the WebAuthn registration process
	// is included here.
	Register string `json:"webauthn_register"`

	// Name of the WebAuthn Security Key to be Added
	//
	// A human-readable name for the security key which will be added.
	RegisterDisplayName string `json:"webauthn_register_displayname"`

	// CSRFToken is the anti-CSRF token
	CSRFToken string `json:"csrf_token"`

	// The identity's traits
	//
	// required: true
	Traits json.RawMessage `json:"traits"`

	// Method
	//
	// Should be set to "webauthn" when trying to add, update, or remove a webAuthn pairing.
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
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (s *Strategy) RegisterRegistrationRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) handleRegistrationError(_ http.ResponseWriter, r *http.Request, f *registration.Flow, p *updateRegistrationFlowWithWebAuthnMethod, err error) error {
	if f != nil {
		if p != nil {
			for _, n := range container.NewFromJSON("", node.DefaultGroup, p.Traits, "traits").Nodes {
				// we only set the value and not the whole field because we want to keep types from the initial form generation
				f.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
			}
		}

		f.UI.Nodes.SetValueAttribute(node.WebAuthnRegisterDisplayName, p.RegisterDisplayName)
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

func (s *Strategy) decode(p *updateRegistrationFlowWithWebAuthnMethod, r *http.Request) error {
	return registration.DecodeBody(p, r, s.hd, s.d.Config(), registrationSchema)
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, regFlow *registration.Flow, i *identity.Identity) (err error) {
	ctx := r.Context()

	if regFlow.Type != flow.TypeBrowser || !s.d.Config().WebAuthnForPasswordless(r.Context()) {
		return flow.ErrStrategyNotResponsible
	}

	var p updateRegistrationFlowWithWebAuthnMethod
	if err := s.decode(&p, r); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, err)
	}

	regFlow.TransientPayload = p.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, regFlow.Type, s.d.Config().DisableAPIFlowEnforcement(ctx), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, err)
	}

	if len(p.Register) == 0 {
		return flow.ErrStrategyNotResponsible
	}

	p.Method = s.SettingsStrategyID()
	if err := flow.MethodEnabledAndAllowed(ctx, regFlow.GetFlowName(), s.SettingsStrategyID(), p.Method, s.d); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, err)
	}

	if len(p.Traits) == 0 {
		p.Traits = json.RawMessage("{}")
	}
	i.Traits = identity.Traits(p.Traits)

	webAuthnSession := gjson.GetBytes(regFlow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if !webAuthnSession.IsObject() {
		return s.handleRegistrationError(w, r, regFlow, &p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object.")))
	}

	var webAuthnSess webauthn.SessionData
	if err := json.Unmarshal([]byte(webAuthnSession.Raw), &webAuthnSess); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object but got: %s", err)))
	}

	webAuthnResponse, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(p.Register))
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, errors.WithStack(
			herodot.ErrBadRequest.WithReasonf("Unable to parse WebAuthn response: %s", err)))
	}

	web, err := webauthn.New(s.d.Config().WebAuthnConfig(r.Context()))
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Unable to get webAuthn config.").WithDebug(err.Error())))
	}

	credential, err := web.CreateCredential(webauthnx.NewUser(webAuthnSess.UserID, nil, web.Config), webAuthnSess, webAuthnResponse)
	if err != nil {
		if devErr := new(protocol.Error); errors.As(err, &devErr) {
			s.d.Logger().WithError(err).WithField("error_devinfo", devErr.DevInfo).Error("Failed to create WebAuthn credential")
		}
		return s.handleRegistrationError(w, r, regFlow, &p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Unable to create WebAuthn credential: %s", err)))
	}

	credentialWebAuthn := identity.CredentialFromWebAuthn(credential, true)
	credentialWebAuthn.DisplayName = p.RegisterDisplayName
	credentialWebAuthnConfig, err := json.Marshal(identity.CredentialsWebAuthnConfig{
		Credentials: identity.CredentialsWebAuthn{*credentialWebAuthn},
		UserHandle:  webAuthnSess.UserID,
	})
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, errors.WithStack(
			herodot.ErrInternalServerError.WithReasonf("Unable to encode identity credentials.").WithDebug(err.Error())))
	}

	i.UpsertCredentialsConfig(s.ID(), credentialWebAuthnConfig, 1)
	if err := s.validateCredentials(ctx, i); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, err)
	}

	// Remove the WebAuthn URL from the internal context now that it is set!
	regFlow.InternalContext, err = sjson.DeleteBytes(regFlow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, err)
	}

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, regFlow); err != nil {
		return s.handleRegistrationError(w, r, regFlow, &p, err)
	}

	return nil
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	ctx := r.Context()

	if f.Type != flow.TypeBrowser || !s.d.Config().WebAuthnForPasswordless(ctx) {
		return nil
	}

	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	nodes, err := container.NodesFromJSONSchema(ctx, node.DefaultGroup, ds.String(), "", nil)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		f.UI.SetNode(n)
	}

	web, err := webauthn.New(s.d.Config().WebAuthnConfig(ctx))
	if err != nil {
		return errors.WithStack(err)
	}

	webauthID := x.NewUUID()
	user := webauthnx.NewUser(webauthID[:], nil, s.d.Config().WebAuthnConfig(ctx))
	option, sessionData, err := web.BeginRegistration(user)
	if err != nil {
		return errors.WithStack(err)
	}

	f.InternalContext, err = sjson.SetBytes(f.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData), sessionData)
	if err != nil {
		return errors.WithStack(err)
	}

	injectWebAuthnOptions, err := json.Marshal(option)
	if err != nil {
		return errors.WithStack(err)
	}

	f.UI.Nodes.Upsert(webauthnx.NewWebAuthnScript(s.d.Config().SelfPublicURL(ctx)))
	f.UI.Nodes.Upsert(webauthnx.NewWebAuthnConnectionName())
	f.UI.Nodes.Upsert(webauthnx.NewWebAuthnConnectionInput())
	f.UI.Nodes.Upsert(webauthnx.NewWebAuthnConnectionTrigger(string(injectWebAuthnOptions)).
		WithMetaLabel(text.NewInfoSelfServiceRegistrationRegisterWebAuthn()))

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	return nil
}
