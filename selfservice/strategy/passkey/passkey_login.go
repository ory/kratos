package passkey

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/webauthnx"
	"github.com/ory/x/decoderx"
)

//go:embed .schema/login.schema.json
var loginSchema []byte

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	webauthnx.RegisterWebauthnRoute(r)
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *login.Flow) error {
	if sr.Type != flow.TypeBrowser {
		return nil
	}

	if err := s.populateLoginMethodForPasswordless(r, sr); errors.Is(err, webauthnx.ErrNoCredentials) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (s *Strategy) populateLoginMethodForPasswordless(r *http.Request, loginFlow *login.Flow) error {
	ctx := r.Context()

	loginFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	loginFlow.UI.Nodes.Upsert(node.NewInputField(
		"identifier",
		"",
		node.DefaultGroup,
		node.InputAttributeTypeText,
		node.WithRequiredInputAttribute,
		func(attributes *node.InputAttributes) { attributes.Autocomplete = "username webauthn" },
	).WithMetaLabel(text.NewInfoNodeInputEmail()))

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

	loginFlow.UI.Nodes.Upsert(&node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name:       "passkey_challenge",
			Type:       node.InputAttributeTypeHidden,
			FieldValue: string(injectWebAuthnOptions),
		}})

	loginFlow.UI.Nodes.Append(webauthnx.NewWebAuthnScript(s.d.Config().SelfPublicURL(ctx)))

	loginFlow.UI.Nodes.Upsert(&node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name: "passkey_login",
			Type: node.InputAttributeTypeHidden,
		}})

	loginFlow.UI.Nodes.Append(node.NewInputField(
		"login_with_passkey",
		"",
		node.PasskeyGroup,
		node.InputAttributeTypeButton,
	).WithMetaLabel(text.NewInfoSelfServiceLoginPasskey()))

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

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID) (i *identity.Identity, err error) {
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
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if len(p.Login) == 0 {
		// Reset all nodes to not confuse users.
		f.UI.Nodes = node.Nodes{}

		err := s.populateLoginMethodForPasswordless(r, f)
		if err != nil {
			return nil, s.handleLoginError(r, f, err)
		}
		return nil, errors.WithStack(flow.ErrCompletedByStrategy)
	}

	return s.loginAuthenticate(w, r, f, p, identity.AuthenticatorAssuranceLevel1)
}

func (s *Strategy) loginAuthenticate(_ http.ResponseWriter, r *http.Request, f *login.Flow, p *updateLoginFlowWithPasskeyMethod, aal identity.AuthenticatorAssuranceLevel) (*identity.Identity, error) {
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
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object but got: %s", err)))
	}
	webAuthnSess.UserID = webAuthnResponse.Response.UserHandle

	i, _, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), string(webAuthnResponse.Response.UserHandle))
	if err != nil {
		time.Sleep(x.RandomDelay(s.d.Config().HasherArgon2(r.Context()).ExpectedDuration, s.d.Config().HasherArgon2(r.Context()).ExpectedDeviation))
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewNoWebAuthnCredentials()))
	}
	err = s.d.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, i, identity.ExpandCredentials)
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewNoWebAuthnCredentials()))
	}

	c, ok := i.GetCredentials(s.ID())
	if !ok {
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewNoWebAuthnRegistered()))
	}

	var o identity.CredentialsWebAuthnConfig
	if err := json.Unmarshal(c.Config, &o); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReason("The WebAuthn credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)))
	}

	webAuthCreds := o.Credentials.ToWebAuthnFiltered(aal)
	if f.IsForced() {
		webAuthCreds = o.Credentials.ToWebAuthn()
	}

	if _, err := web.ValidateLogin(webauthnx.NewUser(o.UserHandle, webAuthCreds, web.Config), webAuthnSess, webAuthnResponse); err != nil {
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
