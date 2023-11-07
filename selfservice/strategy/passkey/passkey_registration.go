package passkey

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
)

//go:embed .schema/registration.schema.json
var registrationSchema []byte

func (s *Strategy) RegisterRegistrationRoutes(r *x.RouterPublic) {
	webauthnx.RegisterWebauthnRoute(r)
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, regFlow *registration.Flow) error {
	ctx := r.Context()

	if regFlow.Type != flow.TypeBrowser {
		return nil
	}

	nodes, err := s.registrationNodes(ctx)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		regFlow.UI.SetNode(n)
	}

	regFlow.UI.Nodes.Append(node.NewInputField(
		"method",
		"passkey",
		node.WebAuthnGroup,
		node.InputAttributeTypeSubmit,
	).WithMetaLabel(text.NewInfoSelfServiceRegistrationRegisterPasskey()))

	regFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	return nil
}

func (s *Strategy) registrationNodes(ctx context.Context) (node.Nodes, error) {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(ctx)
	if err != nil {
		return nil, err
	}

	runner, err := schema.NewExtensionRunner(ctx, identity.NewSchemaExtensionCredentials(nil))
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	runner.Register(c)

	nodes, err := container.NodesFromJSONSchema(ctx, node.DefaultGroup, ds.String(), "", c)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (s *Strategy) identifierNode(ctx context.Context) (*node.Node, error) {
	nodes, err := s.registrationNodes(ctx)
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

	return nil, errors.New("identifier node not found")
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, regFlow *registration.Flow, i *identity.Identity) (err error) {
	ctx := r.Context()

	if regFlow.Type != flow.TypeBrowser {
		return flow.ErrStrategyNotResponsible
	}

	params, err := s.decode(r)
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, err)
	}

	regFlow.TransientPayload = params.TransientPayload

	if err := flow.EnsureCSRF(s.d, r, regFlow.Type, s.d.Config().DisableAPIFlowEnforcement(ctx), s.d.GenerateCSRFToken, params.CSRFToken); err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, err)
	}

	if len(params.Register) == 0 {
		regFlow.UI.Messages.Clear()
		idNode, err := s.identifierNode(ctx)
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
			//if attr, ok := n.Attributes.(*node.InputAttributes); ok {
			//	attr.Type = node.InputAttributeTypeHidden
			//}
			regFlow.UI.SetNode(n)
			if n.ID() == idNode.ID() {
				identifier, _ = n.Attributes.GetValue().(string)
			}

		}

		regFlow.UI.SetCSRF(s.d.GenerateCSRFToken(r))

		webAuthn, err := webauthn.New(s.d.Config().PasskeyConfig(ctx))
		if err != nil {
			return errors.WithStack(err)
		}
		webauthID := x.NewUUID()
		user := &webauthnx.User{
			Name:   identifier,
			ID:     webauthID.Bytes(),
			Config: s.d.Config().PasskeyConfig(ctx),
		}
		option, sessionData, err := webAuthn.BeginRegistration(user)
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

		injectWebAuthnOptions, err := json.Marshal(option)
		if err != nil {
			return errors.WithStack(err)
		}

		regFlow.UI.Nodes.Append(webauthnx.NewCreatePasskeyScript(s.d.Config().SelfPublicURL(ctx)))
		regFlow.UI.Nodes.Upsert(&node.Node{
			Type:  node.Input,
			Group: node.PasskeyGroup,
			Meta:  &node.Meta{},
			Attributes: &node.InputAttributes{
				Name: "passkey_register",
				Type: node.InputAttributeTypeHidden,
			}})

		regFlow.UI.Nodes.Upsert(&node.Node{
			Type:  node.Input,
			Group: node.WebAuthnGroup,
			Meta:  &node.Meta{},
			Attributes: &node.InputAttributes{
				Name:       "create_passkey_data",
				Type:       node.InputAttributeTypeHidden,
				FieldValue: string(injectWebAuthnOptions),
			}})

		redirectTo := regFlow.AppendTo(s.d.Config().SelfServiceFlowRegistrationUI(r.Context())).String()

		if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(r.Context(), regFlow); err != nil {
			return s.handleRegistrationError(w, r, regFlow, params, err)
		}

		x.AcceptToRedirectOrJSON(w, r, s.d.Writer(), err, redirectTo)
		return nil
	}

	params.Method = s.SettingsStrategyID()
	if err := flow.MethodEnabledAndAllowed(ctx, regFlow.GetFlowName(), string(s.ID()), params.Method, s.d); err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, err)
	}

	if len(params.Traits) == 0 {
		params.Traits = json.RawMessage("{}")
	}
	i.Traits = identity.Traits(params.Traits)

	webAuthnSession := gjson.GetBytes(regFlow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if !webAuthnSession.IsObject() {
		return s.handleRegistrationError(w, r, regFlow, params, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object.")))
	}

	var webAuthnSess webauthn.SessionData
	if err := json.Unmarshal([]byte(gjson.GetBytes(regFlow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData)).Raw), &webAuthnSess); err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected WebAuthN in internal context to be an object but got: %s", err)))
	}

	webAuthnResponse, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(params.Register))
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse WebAuthn response: %s", err)))
	}

	webAuthn, err := webauthn.New(s.d.Config().PasskeyConfig(ctx))
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to get webAuthn config.").WithDebug(err.Error())))
	}

	credential, err := webAuthn.CreateCredential(webauthnx.NewUser(webAuthnSess.UserID, nil, webAuthn.Config), webAuthnSess, webAuthnResponse)
	if err != nil {
		if devErr := new(protocol.Error); errors.As(err, &devErr) {
			s.d.Logger().WithError(err).WithField("error_devinfo", devErr.DevInfo).Error("Failed to create WebAuthn credential")
		}
		return s.handleRegistrationError(w, r, regFlow, params, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to create WebAuthn credential: %s", err)))
	}

	var cc identity.CredentialsWebAuthnConfig
	wc := identity.CredentialFromWebAuthn(credential, true)
	wc.AddedAt = time.Now().UTC().Round(time.Second)
	//wc.DisplayName = params.RegisterDisplayName
	wc.IsPasswordless = s.d.Config().WebAuthnForPasswordless(ctx)
	cc.UserHandle = webAuthnSess.UserID

	cc.Credentials = append(cc.Credentials, *wc)
	co, err := json.Marshal(cc)
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode identity credentials.").WithDebug(err.Error())))
	}

	i.UpsertCredentialsConfig(s.ID(), co, 1)
	if err := s.validateCredentials(ctx, i); err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, err)
	}

	// Remove the WebAuthn URL from the internal context now that it is set!
	regFlow.InternalContext, err = sjson.DeleteBytes(regFlow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeySessionData))
	if err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, err)
	}

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(ctx, regFlow); err != nil {
		return s.handleRegistrationError(w, r, regFlow, params, err)
	}

	return nil
}

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

func (s *Strategy) decode(r *http.Request) (*updateRegistrationFlowWithPasskeyMethod, error) {
	var p updateRegistrationFlowWithPasskeyMethod
	err := registration.DecodeBody(&p, r, s.hd, s.d.Config(), registrationSchema)

	return &p, err
}

func (s *Strategy) handleRegistrationError(_ http.ResponseWriter, r *http.Request, f *registration.Flow, p *updateRegistrationFlowWithPasskeyMethod, err error) error {
	if f != nil {
		if p != nil {
			for _, n := range container.NewFromJSON("", node.DefaultGroup, p.Traits, "traits").Nodes {
				// we only set the value and not the whole field because we want to keep types from the initial form generation
				f.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
			}
		}

		//f.UI.Nodes.SetValueAttribute(node.WebAuthnRegisterDisplayName, p.RegisterDisplayName)
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
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
