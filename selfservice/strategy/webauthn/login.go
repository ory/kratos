// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ory/kratos/selfservice/flowhelpers"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"

	"github.com/ory/x/urlx"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/x/decoderx"
)

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *login.Flow) error {
	if sr.Type != flow.TypeBrowser {
		return nil
	}

	if s.d.Config().WebAuthnForPasswordless(r.Context()) && (requestedAAL == identity.AuthenticatorAssuranceLevel1) {
		if err := s.populateLoginMethodForPasswordless(r, sr); errors.Is(err, ErrNoCredentials) {
			return nil
		} else if err != nil {
			return err
		}
		return nil
	} else if sr.IsForced() {
		if err := s.populateLoginMethodForPasswordless(r, sr); errors.Is(err, ErrNoCredentials) {
			return nil
		} else if err != nil {
			return err
		}
		return nil
	} else if !s.d.Config().WebAuthnForPasswordless(r.Context()) && (requestedAAL == identity.AuthenticatorAssuranceLevel2) {
		// We have done proper validation before so this should never error
		sess, err := s.d.SessionManager().FetchFromRequest(r.Context(), r)
		if err != nil {
			return err
		}

		if err := s.populateLoginMethod(r, sr, sess.Identity, text.NewInfoSelfServiceLoginWebAuthn(), identity.AuthenticatorAssuranceLevel2); errors.Is(err, ErrNoCredentials) {
			return nil
		} else if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (s *Strategy) populateLoginMethodForPasswordless(r *http.Request, sr *login.Flow) error {
	if sr.IsForced() {
		identifier, id, _ := flowhelpers.GuessForcedLoginIdentifier(r, s.d, sr, s.ID())
		if identifier == "" {
			return nil
		}

		if err := s.populateLoginMethod(r, sr, id, text.NewInfoSelfServiceLoginWebAuthn(), ""); errors.Is(err, ErrNoCredentials) {
			return nil
		} else if err != nil {
			return err
		}

		sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		sr.UI.SetNode(node.NewInputField("identifier", identifier, node.DefaultGroup, node.InputAttributeTypeHidden))
		return nil
	}

	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	sr.UI.SetNode(node.NewInputField("identifier", "", node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeLabelID()))
	sr.UI.GetNodes().Append(node.NewInputField("method", "webauthn", node.WebAuthnGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoLoginPasswordlessWebAuthn()))
	return nil
}

func (s *Strategy) populateLoginMethod(r *http.Request, sr *login.Flow, i *identity.Identity, label *text.Message, aal identity.AuthenticatorAssuranceLevel) error {
	id, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), i.ID)
	if err != nil {
		return err
	}

	cred, ok := id.GetCredentials(s.ID())
	if !ok {
		// Identity has no webauth
		return nil
	}

	var conf identity.CredentialsWebAuthnConfig
	if err := json.Unmarshal(cred.Config, &conf); err != nil {
		return errors.WithStack(err)
	}

	webAuthCreds := conf.Credentials.ToWebAuthn()
	if !sr.IsForced() {
		webAuthCreds = conf.Credentials.ToWebAuthnFiltered(aal)
	}

	if len(webAuthCreds) == 0 {
		// Identity has no webauthn
		return ErrNoCredentials
	}

	web, err := webauthn.New(s.d.Config().WebAuthnConfig(r.Context()))
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to initiate WebAuth.").WithDebug(err.Error()))
	}

	options, sessionData, err := web.BeginLogin(NewUser(conf.UserHandle, webAuthCreds, web.Config))
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to initiate WebAuth login.").WithDebug(err.Error()))
	}

	// Remove the WebAuthn URL from the internal context now that it is set!
	sr.InternalContext, err = sjson.SetBytes(sr.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData), sessionData)
	if err != nil {
		return errors.WithStack(err)
	}

	injectWebAuthnOptions, err := json.Marshal(options)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(cred.Identifiers) > 0 {
		sr.UI.SetNode(node.NewInputField("identifier", cred.Identifiers[0], node.DefaultGroup, node.InputAttributeTypeHidden))
	}

	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	sr.UI.Nodes.Upsert(NewWebAuthnScript(urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()), webAuthnRoute).String(), jsOnLoad))
	sr.UI.SetNode(NewWebAuthnLoginTrigger(string(injectWebAuthnOptions)).
		WithMetaLabel(label))
	sr.UI.Nodes.Upsert(NewWebAuthnLoginInput())

	return nil
}

func (s *Strategy) handleLoginError(r *http.Request, f *login.Flow, err error) error {
	if f != nil {
		f.UI.Nodes.ResetNodes("webauth_login")
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

// Update Login Flow with WebAuthn Method
//
// swagger:model updateLoginFlowWithWebAuthnMethod
type updateLoginFlowWithWebAuthnMethod struct {
	// Identifier is the email or username of the user trying to log in.
	//
	// required: true
	Identifier string `json:"identifier"`

	// Method should be set to "webAuthn" when logging in using the WebAuthn strategy.
	//
	// required: true
	Method string `json:"method"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `json:"csrf_token"`

	// Login a WebAuthn Security Key
	//
	// This must contain the ID of the WebAuthN connection.
	Login string `json:"webauthn_login"`
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID) (i *identity.Identity, err error) {
	if f.Type != flow.TypeBrowser {
		return nil, flow.ErrStrategyNotResponsible
	}

	var p updateLoginFlowWithWebAuthnMethod
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

	if s.d.Config().WebAuthnForPasswordless(r.Context()) || f.IsForced() && f.RequestedAAL == identity.AuthenticatorAssuranceLevel1 {
		return s.loginPasswordless(w, r, f, &p)
	}

	return s.loginMultiFactor(w, r, f, identityID, &p)
}

func (s *Strategy) loginPasswordless(w http.ResponseWriter, r *http.Request, f *login.Flow, p *updateLoginFlowWithWebAuthnMethod) (i *identity.Identity, err error) {
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if p.Identifier == "" {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrBadRequest.WithReason("identifier is required")))
	}

	i, _, err = s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), p.Identifier)
	if err != nil {
		time.Sleep(x.RandomDelay(s.d.Config().HasherArgon2(r.Context()).ExpectedDuration, s.d.Config().HasherArgon2(r.Context()).ExpectedDeviation))
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewNoWebAuthnCredentials()))
	}

	if len(p.Login) == 0 {
		// Reset all nodes to not confuse users.
		// This is kinda hacky and will probably need to be updated at some point.
		previousNodes := f.UI.Nodes
		f.UI.Nodes = node.Nodes{}

		if err := s.populateLoginMethod(r, f, i, text.NewInfoSelfServiceLoginContinue(), identity.AuthenticatorAssuranceLevel1); errors.Is(err, ErrNoCredentials) {
			f.UI.Nodes = previousNodes
			return nil, s.handleLoginError(r, f, schema.NewNoWebAuthnCredentials())
		} else if err != nil {
			return nil, s.handleLoginError(r, f, err)
		}

		// Adds the "Continue" button
		f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		f.UI.Messages.Add(text.NewInfoLoginWebAuthnPasswordless())
		f.UI.SetNode(node.NewInputField("identifier", p.Identifier, node.DefaultGroup, node.InputAttributeTypeHidden, node.WithRequiredInputAttribute))
		if err := s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
			return nil, s.handleLoginError(r, f, err)
		}

		redirectTo := f.AppendTo(s.d.Config().SelfServiceFlowLoginUI(r.Context())).String()
		if x.IsJSONRequest(r) {
			s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(redirectTo))
		} else {
			http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowLoginUI(r.Context())).String(), http.StatusSeeOther)
		}

		return nil, errors.WithStack(flow.ErrCompletedByStrategy)
	}

	return s.loginAuthenticate(w, r, f, i.ID, p, identity.AuthenticatorAssuranceLevel1)
}

func (s *Strategy) loginAuthenticate(_ http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID, p *updateLoginFlowWithWebAuthnMethod, aal identity.AuthenticatorAssuranceLevel) (*identity.Identity, error) {
	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), identityID)
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewNoWebAuthnRegistered()))
	}

	c, ok := i.GetCredentials(s.ID())
	if !ok {
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewNoWebAuthnRegistered()))
	}

	var o identity.CredentialsWebAuthnConfig
	if err := json.Unmarshal(c.Config, &o); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReason("The WebAuthn credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)))
	}

	web, err := webauthn.New(s.d.Config().WebAuthnConfig(r.Context()))
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to get webAuthn config.").WithDebug(err.Error())))
	}

	webAuthnResponse, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(p.Login))
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse WebAuthn response.").WithDebug(err.Error())))
	}

	var webAuthnSess webauthn.SessionData
	if err := json.Unmarshal([]byte(gjson.GetBytes(f.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData)).Raw), &webAuthnSess); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object but got: %s", err)))
	}

	webAuthCreds := o.Credentials.ToWebAuthnFiltered(aal)
	if f.IsForced() {
		webAuthCreds = o.Credentials.ToWebAuthn()
	}

	if _, err := web.ValidateLogin(NewUser(o.UserHandle, webAuthCreds, web.Config), webAuthnSess, webAuthnResponse); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewWebAuthnVerifierWrongError("#/")))
	}

	// Remove the WebAuthn URL from the internal context now that it is set!
	f.InternalContext, err = sjson.DeleteBytes(f.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(err))
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithDebug(err.Error())))
	}

	return i, nil
}

func (s *Strategy) loginMultiFactor(w http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID, p *updateLoginFlowWithWebAuthnMethod) (*identity.Identity, error) {
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel2); err != nil {
		return nil, err
	}
	return s.loginAuthenticate(w, r, f, identityID, p, identity.AuthenticatorAssuranceLevel2)
}
