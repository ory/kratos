package webauthn

import (
	"encoding/json"
	"net/http"
	"strings"

	errors2 "github.com/ory/kratos/schema/errors"

	"github.com/ory/x/urlx"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/x/decoderx"
)

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *login.Flow) error {
	// AAL is configurable for webauth
	if requestedAAL != identity.AuthenticatorAssuranceLevel2 || sr.Type != flow.TypeBrowser {
		return nil
	}

	// We have done proper validation before so this should never error
	sess, err := s.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		return err
	}

	id, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), sess.IdentityID)
	if err != nil {
		return err
	}

	cred, ok := id.GetCredentials(s.ID())
	if !ok {
		// Identity has no webauth
		return nil
	}

	var conf CredentialsConfig
	if err := json.Unmarshal(cred.Config, &conf); err != nil {
		return errors.WithStack(err)
	}

	if len(conf.Credentials) == 0 {
		// Identity has no webauth
		return nil
	}

	web, err := s.newWebAuthn(r.Context())
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to initiate WebAuth.").WithDebug(err.Error()))
	}

	options, sessionData, err := web.BeginLogin(&wrappedUser{id: sess.IdentityID, c: conf.Credentials.ToWebAuthn()})
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to initiate WebAuth login.").WithDebug(err.Error()))
	}

	// Remove the WebAuthn URL from the internal context now that it is set!
	sr.InternalContext, err = sjson.SetBytes(sr.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData), sessionData)
	if err != nil {
		return err
	}

	injectWebAuthnOptions, err := json.Marshal(options)
	if err != nil {
		return errors.WithStack(err)
	}

	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	sr.UI.Nodes.Upsert(NewWebAuthnScript(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(), webAuthnRoute).String(), jsOnLoad))
	sr.UI.SetNode(NewWebAuthnLoginTrigger(string(injectWebAuthnOptions)))
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

// submitSelfServiceLoginFlowWithWebAuthnMethodBody is used to decode the login form payload.
//
// swagger:model submitSelfServiceLoginFlowWithWebAuthnMethodBody
type submitSelfServiceLoginFlowWithWebAuthnMethodBody struct {
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

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, ss *session.Session) (i *identity.Identity, err error) {
	if f.Type != flow.TypeBrowser {
		return nil, flow.ErrStrategyNotResponsible
	}

	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel2); err != nil {
		return nil, err
	}

	var p submitSelfServiceLoginFlowWithWebAuthnMethodBody
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if len(p.Login) > 0 {
		// This method has only two submit buttons
		p.Method = s.SettingsStrategyID()
		if err := flow.MethodEnabledAndAllowed(r.Context(), s.SettingsStrategyID(), p.Method, s.d); err != nil {
			return nil, s.handleLoginError(r, f, err)
		}
	} else {
		return nil, flow.ErrStrategyNotResponsible
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), ss.IdentityID.String())
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(errors2.NewNoWebAuthnRegistered()))
	}

	var o CredentialsConfig
	if err := json.Unmarshal(c.Config, &o); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReason("The WebAuthn credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)))
	}

	web, err := s.newWebAuthn(r.Context())
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

	if _, err := web.ValidateLogin(&wrappedUser{id: i.ID, c: o.Credentials.ToWebAuthn()}, webAuthnSess, webAuthnResponse); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(errors2.NewWebAuthnVerifierWrongError("#/")))
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
