// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import (
	_ "embed"
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
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flowhelpers"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/webauthnx"
	"github.com/ory/x/decoderx"
)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	webauthnx.RegisterWebauthnRoute(r)
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, aal identity.AuthenticatorAssuranceLevel, sr *login.Flow) error {
	if sr.Type != flow.TypeBrowser || aal != identity.AuthenticatorAssuranceLevel1 {
		return nil
	}

	return s.populateLoginMethodForPasskeys(r, sr)
}

func (s *Strategy) populateLoginMethodForPasskeys(r *http.Request, loginFlow *login.Flow) error {
	if loginFlow.IsForced() {
		return s.populateLoginMethodForRefresh(r, loginFlow)
	}

	ctx := r.Context()

	loginFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	identifierLabel, err := login.GetIdentifierLabelFromSchema(r.Context(), ds.String())
	if err != nil {
		return err
	}

	webAuthn, err := webauthn.New(s.d.Config().PasskeyConfig(ctx))
	if err != nil {
		return errors.WithStack(err)
	}
	option, sessionData, err := webAuthn.BeginDiscoverableLogin()
	if err != nil {
		return errors.WithStack(err)
	}

	loginFlow.InternalContext, err = sjson.SetBytes(
		loginFlow.InternalContext,
		flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData),
		sessionData,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	injectWebAuthnOptions, err := json.Marshal(option)
	if err != nil {
		return errors.WithStack(err)
	}

	loginFlow.UI.Nodes.Upsert(node.NewInputField(
		"identifier",
		"",
		node.DefaultGroup,
		node.InputAttributeTypeText,
		node.WithRequiredInputAttribute,
		func(attributes *node.InputAttributes) { attributes.Autocomplete = "username webauthn" },
	).WithMetaLabel(identifierLabel))

	loginFlow.UI.Nodes.Upsert(&node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name:       node.PasskeyChallenge,
			Type:       node.InputAttributeTypeHidden,
			FieldValue: string(injectWebAuthnOptions),
		}})

	loginFlow.UI.Nodes.Upsert(webauthnx.NewWebAuthnScript(s.d.Config().SelfPublicURL(ctx)))

	loginFlow.UI.Nodes.Upsert(&node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name: node.PasskeyLogin,
			Type: node.InputAttributeTypeHidden,
		}})

	loginFlow.UI.Nodes.Append(node.NewInputField(
		node.PasskeyLoginTrigger,
		"",
		node.PasskeyGroup,
		node.InputAttributeTypeButton,
		node.WithInputAttributes(func(attr *node.InputAttributes) {
			attr.OnClick = "window.__oryPasskeyLogin()"                // this function is defined in webauthn.js
			attr.OnLoad = "window.__oryPasskeyLoginAutocompleteInit()" // same here
		}),
	).WithMetaLabel(text.NewInfoSelfServiceLoginPasskey()))

	return nil
}

func (s *Strategy) populateLoginMethodForRefresh(r *http.Request, loginFlow *login.Flow) error {
	ctx := r.Context()

	identifier, id, _ := flowhelpers.GuessForcedLoginIdentifier(r, s.d, loginFlow, s.ID())
	if identifier == "" {
		return nil
	}

	id, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), id.ID)
	if err != nil {
		return err
	}

	cred, ok := id.GetCredentials(s.ID())
	if !ok {
		// Identity has no passkey
		return nil
	}

	var conf identity.CredentialsWebAuthnConfig
	if err := json.Unmarshal(cred.Config, &conf); err != nil {
		return errors.WithStack(err)
	}

	webAuthCreds := conf.Credentials.ToWebAuthn()
	if len(webAuthCreds) == 0 {
		// Identity has no webauthn
		return nil
	}

	passkeyIdentifier := s.PasskeyDisplayNameFromIdentity(ctx, id)

	webAuthn, err := webauthn.New(s.d.Config().PasskeyConfig(ctx))
	if err != nil {
		return errors.WithStack(err)
	}
	option, sessionData, err := webAuthn.BeginLogin(&webauthnx.User{
		Name:        passkeyIdentifier,
		ID:          conf.UserHandle,
		Credentials: webAuthCreds,
		Config:      webAuthn.Config,
	})
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to initiate passkey login.").WithDebug(err.Error()))
	}

	loginFlow.InternalContext, err = sjson.SetBytes(
		loginFlow.InternalContext,
		flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData),
		sessionData,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	injectWebAuthnOptions, err := json.Marshal(option)
	if err != nil {
		return errors.WithStack(err)
	}

	loginFlow.UI.Nodes.Upsert(&node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name:       node.PasskeyChallenge,
			Type:       node.InputAttributeTypeHidden,
			FieldValue: string(injectWebAuthnOptions),
		}})

	loginFlow.UI.Nodes.Append(webauthnx.NewWebAuthnScript(s.d.Config().SelfPublicURL(ctx)))

	loginFlow.UI.Nodes.Upsert(&node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name: node.PasskeyLogin,
			Type: node.InputAttributeTypeHidden,
		}})

	loginFlow.UI.Nodes.Append(node.NewInputField(
		node.PasskeyLoginTrigger,
		"",
		node.PasskeyGroup,
		node.InputAttributeTypeButton,
		node.WithInputAttributes(func(attr *node.InputAttributes) {
			attr.OnClick = "window.__oryPasskeyLogin()" // this function is defined in webauthn.js
		}),
	).WithMetaLabel(text.NewInfoSelfServiceLoginPasskey()))

	loginFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	loginFlow.UI.SetNode(node.NewInputField(
		"identifier",
		passkeyIdentifier,
		node.DefaultGroup,
		node.InputAttributeTypeHidden,
	))

	return nil
}

func (s *Strategy) handleLoginError(r *http.Request, f *login.Flow, err error) error {
	if f != nil {
		f.UI.Nodes.ResetNodes(node.PasskeyLogin)
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

// Update Login Flow with Passkey Method
//
// swagger:model updateLoginFlowWithPasskeyMethod
type updateLoginFlowWithPasskeyMethod struct {
	// Method should be set to "passkey" when logging in using the Passkey strategy.
	//
	// required: true
	Method string `json:"method"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `json:"csrf_token"`

	// Login a WebAuthn Security Key
	//
	// This must contain the ID of the WebAuthN connection.
	Login string `json:"passkey_login"`
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, _ *session.Session) (i *identity.Identity, err error) {
	if f.Type != flow.TypeBrowser {
		return nil, flow.ErrStrategyNotResponsible
	}

	var p updateLoginFlowWithPasskeyMethod
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if len(p.Login) > 0 || p.Method == s.SettingsStrategyID() {
		// This method has only two submit buttons
		p.Method = s.SettingsStrategyID()
	} else {
		return nil, flow.ErrStrategyNotResponsible
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), f.GetFlowName(), s.SettingsStrategyID(), p.Method, s.d); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	return s.loginPasswordless(w, r, f, &p)
}

func (s *Strategy) loginPasswordless(w http.ResponseWriter, r *http.Request, f *login.Flow, p *updateLoginFlowWithPasskeyMethod) (i *identity.Identity, err error) {
	if err = login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if err = flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if len(p.Login) == 0 {
		// Reset all nodes to not confuse users.
		f.UI.Nodes = node.Nodes{}

		if err = s.populateLoginMethodForPasskeys(r, f); err != nil {
			return nil, s.handleLoginError(r, f, err)
		}

		redirectTo := f.AppendTo(s.d.Config().SelfServiceFlowLoginUI(r.Context())).String()
		if x.IsJSONRequest(r) {
			s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(redirectTo))
		} else {
			http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		}

		return nil, errors.WithStack(flow.ErrCompletedByStrategy)
	}

	return s.loginAuthenticate(w, r, f, p, identity.AuthenticatorAssuranceLevel1)
}

func (s *Strategy) loginAuthenticate(_ http.ResponseWriter, r *http.Request, f *login.Flow, p *updateLoginFlowWithPasskeyMethod, _ identity.AuthenticatorAssuranceLevel) (*identity.Identity, error) {
	ctx := r.Context()

	web, err := webauthn.New(s.d.Config().PasskeyConfig(ctx))
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to get webAuthn config.").WithDebug(err.Error())))
	}

	webAuthnResponse, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(p.Login))
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse WebAuthn response.").WithDebug(err.Error())))
	}

	var webAuthnSess webauthn.SessionData
	if err := json.Unmarshal([]byte(gjson.GetBytes(f.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData)).Raw), &webAuthnSess); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.
			WithReasonf("Expected WebAuthN in internal context to be an object but got: %s", err)))
	}
	webAuthnSess.UserID = nil

	userHandle := webAuthnResponse.Response.UserHandle
	credentialType := identity.CredentialsTypePasskey
	i, _, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, identity.CredentialsTypePasskey, string(userHandle))
	if err != nil {
		// Migration strategy: Don't give up yet! If we don't find a "passkey" credential
		// here, look for a "webauthn" credential next
		if i, err = s.d.PrivilegedIdentityPool().FindIdentityByWebauthnUserHandle(ctx, userHandle); err != nil {
			return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewNoWebAuthnCredentials()))
		}
		credentialType = identity.CredentialsTypeWebAuthn
	}
	err = s.d.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, i, identity.ExpandCredentials)
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.
			WithReason("Could not load identity credentials").
			WithWrap(err)))
	}

	c, ok := i.GetCredentials(credentialType)
	if !ok {
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewNoWebAuthnRegistered()))
	}

	var o identity.CredentialsWebAuthnConfig
	if err := json.Unmarshal(c.Config, &o); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.
			WithReason("The WebAuthn credentials could not be decoded properly").
			WithDebug(err.Error()).
			WithWrap(err)))
	}

	webAuthCreds := o.Credentials.PasswordlessOnly()

	_, err = web.ValidateDiscoverableLogin(
		func(rawID, userHandle []byte) (user webauthn.User, err error) {
			return webauthnx.NewUser(userHandle, webAuthCreds, web.Config), nil
		}, webAuthnSess, webAuthnResponse)
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewWebAuthnVerifierWrongError("#/")))
	}

	// Remove the WebAuthn URL from the internal context now that it is set!
	f.InternalContext, err = sjson.DeleteBytes(f.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(err))
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithDebug(err.Error())))
	}

	return i, nil
}
