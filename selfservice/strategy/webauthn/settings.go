// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ory/kratos/text"

	"github.com/ory/x/urlx"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"

	"github.com/ory/kratos/session"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
)

func (s *Strategy) RegisterSettingsRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) SettingsStrategyID() string {
	return identity.CredentialsTypeWebAuthn.String()
}

const (
	InternalContextKeySessionData = "session_data"
)

// Update Settings Flow with WebAuthn Method
//
// swagger:model updateSettingsFlowWithWebAuthnMethod
type updateSettingsFlowWithWebAuthnMethod struct {
	// Register a WebAuthn Security Key
	//
	// It is expected that the JSON returned by the WebAuthn registration process
	// is included here.
	Register string `json:"webauthn_register"`

	// Name of the WebAuthn Security Key to be Added
	//
	// A human-readable name for the security key which will be added.
	RegisterDisplayName string `json:"webauthn_register_displayname"`

	// Remove a WebAuthn Security Key
	//
	// This must contain the ID of the WebAuthN connection.
	Remove string `json:"webauthn_remove"`

	// CSRFToken is the anti-CSRF token
	CSRFToken string `json:"csrf_token"`

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
}

func (p *updateSettingsFlowWithWebAuthnMethod) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.Flow)
}

func (p *updateSettingsFlowWithWebAuthnMethod) SetFlowID(rid uuid.UUID) {
	p.Flow = rid.String()
}

func (s *Strategy) Settings(w http.ResponseWriter, r *http.Request, f *settings.Flow, ss *session.Session) (*settings.UpdateContext, error) {
	if f.Type != flow.TypeBrowser {
		return nil, flow.ErrStrategyNotResponsible
	}
	var p updateSettingsFlowWithWebAuthnMethod
	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, f, ss, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		return ctxUpdate, s.continueSettingsFlow(w, r, ctxUpdate, &p)
	} else if err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if err := s.decodeSettingsFlow(r, &p); err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if len(p.Register+p.Remove) > 0 {
		// This method has only two submit buttons
		p.Method = s.SettingsStrategyID()
		if err := flow.MethodEnabledAndAllowed(r.Context(), f.GetFlowName(), s.SettingsStrategyID(), p.Method, s.d); err != nil {
			return nil, s.handleSettingsError(w, r, ctxUpdate, &p, err)
		}
	} else {
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	// This does not come from the payload!
	p.Flow = ctxUpdate.Flow.ID.String()
	if err := s.continueSettingsFlow(w, r, ctxUpdate, &p); err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	return ctxUpdate, nil
}

func (s *Strategy) decodeSettingsFlow(r *http.Request, dest interface{}) error {
	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(settingsSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	return decoderx.NewHTTP().Decode(r, dest, compiler,
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	)
}

func (s *Strategy) continueSettingsFlow(
	w http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithWebAuthnMethod,
) error {
	if len(p.Register+p.Remove) > 0 {
		if err := flow.MethodEnabledAndAllowed(r.Context(), flow.SettingsFlow, s.SettingsStrategyID(), s.SettingsStrategyID(), s.d); err != nil {
			return err
		}

		if err := flow.EnsureCSRF(s.d, r, ctxUpdate.Flow.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
			return err
		}

		if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context())).Before(time.Now()) {
			return errors.WithStack(settings.NewFlowNeedsReAuth())
		}
	} else {
		return errors.New("ended up in unexpected state")
	}

	if len(p.Register) > 0 {
		return s.continueSettingsFlowAdd(w, r, ctxUpdate, p)
	} else if len(p.Remove) > 0 {
		return s.continueSettingsFlowRemove(w, r, ctxUpdate, p)
	}

	return errors.New("ended up in unexpected state")
}

func (s *Strategy) continueSettingsFlowRemove(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithWebAuthnMethod) error {
	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.IdentityID)
	if err != nil {
		return err
	}

	cred, ok := i.GetCredentials(s.ID())
	if !ok {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("You tried to remove a WebAuthn but you have no WebAuthn set up."))
	}

	var cc identity.CredentialsWebAuthnConfig
	if err := json.Unmarshal(cred.Config, &cc); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to decode identity credentials.").WithDebug(err.Error()))
	}

	var wasPasswordless bool
	updated := make([]identity.CredentialWebAuthn, 0)
	for k, cred := range cc.Credentials {
		if fmt.Sprintf("%x", cred.ID) != p.Remove {
			updated = append(updated, cc.Credentials[k])
		} else if cred.IsPasswordless {
			wasPasswordless = true
		}
	}

	if len(updated) == len(cc.Credentials) {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("You tried to remove a WebAuthn credential which does not exist."))
	}

	count, err := s.d.IdentityManager().CountActiveFirstFactorCredentials(r.Context(), i)
	if err != nil {
		return err
	}

	if count < 2 && wasPasswordless {
		return s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(ErrNotEnoughCredentials))
	}

	if len(updated) == 0 {
		i.DeleteCredentialsType(identity.CredentialsTypeWebAuthn)
		ctxUpdate.UpdateIdentity(i)
		return nil
	}

	cc.Credentials = updated
	cred.Config, err = json.Marshal(cc)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode identity credentials.").WithDebug(err.Error()))
	}

	i.SetCredentials(s.ID(), *cred)
	ctxUpdate.UpdateIdentity(i)
	return nil
}

func (s *Strategy) continueSettingsFlowAdd(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithWebAuthnMethod) error {
	webAuthnSession := gjson.GetBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if !webAuthnSession.IsObject() {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object."))
	}

	var webAuthnSess webauthn.SessionData
	if err := json.Unmarshal([]byte(gjson.GetBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData)).Raw), &webAuthnSess); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object but got: %s", err))
	}

	webAuthnResponse, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(p.Register))
	if err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse WebAuthn response: %s", err))
	}

	web, err := webauthn.New(s.d.Config().WebAuthnConfig(r.Context()))
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to get webAuthn config.").WithDebug(err.Error()))
	}

	credential, err := web.CreateCredential(NewUser(ctxUpdate.Session.IdentityID[:], nil, web.Config), webAuthnSess, webAuthnResponse)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to create WebAuthn credential: %s", err))
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.IdentityID)
	if err != nil {
		return err
	}

	cred := i.GetCredentialsOr(s.ID(), &identity.Credentials{Config: sqlxx.JSONRawMessage("{}")})

	var cc identity.CredentialsWebAuthnConfig
	if err := json.Unmarshal(cred.Config, &cc); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to decode identity credentials.").WithDebug(err.Error()))
	}

	wc := identity.CredentialFromWebAuthn(credential, s.d.Config().WebAuthnForPasswordless(r.Context()))
	wc.AddedAt = time.Now().UTC().Round(time.Second)
	wc.DisplayName = p.RegisterDisplayName
	wc.IsPasswordless = s.d.Config().WebAuthnForPasswordless(r.Context())
	cc.UserHandle = ctxUpdate.Session.IdentityID[:]

	cc.Credentials = append(cc.Credentials, *wc)
	co, err := json.Marshal(cc)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode identity credentials.").WithDebug(err.Error()))
	}

	i.UpsertCredentialsConfig(s.ID(), co, 1)
	if err := s.validateCredentials(r.Context(), i); err != nil {
		return err
	}

	// Remove the WebAuthn URL from the internal context now that it is set!
	ctxUpdate.Flow.InternalContext, err = sjson.DeleteBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if err != nil {
		return err
	}

	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), ctxUpdate.Flow); err != nil {
		return err
	}

	aal := identity.AuthenticatorAssuranceLevel1
	if !s.d.Config().WebAuthnForPasswordless(r.Context()) {
		aal = identity.AuthenticatorAssuranceLevel2
	}

	// Since we added the method, it also means that we have authenticated it
	if err := s.d.SessionManager().SessionAddAuthenticationMethods(r.Context(), ctxUpdate.Session.ID, session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    aal,
	}); err != nil {
		return err
	}

	ctxUpdate.UpdateIdentity(i)
	return nil
}

func (s *Strategy) identityListWebAuthn(id *identity.Identity) (*identity.CredentialsWebAuthnConfig, error) {
	cred, ok := id.GetCredentials(s.ID())
	if !ok {
		return nil, errors.WithStack(sqlcon.ErrNoRows)
	}

	var cc identity.CredentialsWebAuthnConfig
	if err := json.Unmarshal(cred.Config, &cc); err != nil {
		return nil, errors.WithStack(err)
	}

	return &cc, nil
}

func (s *Strategy) PopulateSettingsMethod(r *http.Request, id *identity.Identity, f *settings.Flow) error {
	if f.Type != flow.TypeBrowser {
		return nil
	}

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	confidentialIdentity, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), id.ID)
	if err != nil {
		return err
	}

	count, err := s.d.IdentityManager().CountActiveFirstFactorCredentials(r.Context(), confidentialIdentity)
	if err != nil {
		return err
	}

	if webAuthns, err := s.identityListWebAuthn(confidentialIdentity); errors.Is(err, sqlcon.ErrNoRows) {
		// Do nothing
	} else if err != nil {
		return err
	} else {
		for k := range webAuthns.Credentials {
			// We only show the option to remove a credential, if it is not the last one when passwordless,
			// or, if it is for MFA we show it always.
			cred := &webAuthns.Credentials[k]
			if cred.IsPasswordless && count < 2 {
				// Do not remove this node because it is the last credential the identity can sign in with.
				continue
			}
			f.UI.Nodes.Append(NewWebAuthnUnlink(cred))
		}
	}

	web, err := webauthn.New(s.d.Config().WebAuthnConfig(r.Context()))
	if err != nil {
		return errors.WithStack(err)
	}

	option, sessionData, err := web.BeginRegistration(NewUser(id.ID.Bytes(), nil, web.Config))
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

	f.UI.Nodes.Upsert(NewWebAuthnScript(urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()), webAuthnRoute).String(), jsOnLoad))
	f.UI.Nodes.Upsert(NewWebAuthnConnectionName())
	f.UI.Nodes.Upsert(NewWebAuthnConnectionTrigger(string(injectWebAuthnOptions)).
		WithMetaLabel(text.NewInfoSelfServiceSettingsRegisterWebAuthn()))
	f.UI.Nodes.Upsert(NewWebAuthnConnectionInput())
	return nil
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithWebAuthnMethod, err error) error {
	// Do not pause flow if the flow type is an API flow as we can't save cookies in those flows.
	if e := new(settings.FlowNeedsReAuth); errors.As(err, &e) && ctxUpdate.Flow != nil && ctxUpdate.Flow.Type == flow.TypeBrowser {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r, settings.ContinuityKey(s.SettingsStrategyID()), settings.ContinuityOptions(p, ctxUpdate.GetSessionIdentity())...); err != nil {
			return err
		}
	}

	if ctxUpdate.Flow != nil {
		ctxUpdate.Flow.UI.ResetMessages()
		ctxUpdate.Flow.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	}

	return err
}
