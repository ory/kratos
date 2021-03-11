package password

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/errorsx"
)

type RegistrationFormPayload struct {
	Password  RegistrationFormPasswordPayload `json:"password"`
	CSRFToken string                          `json:"csrf_token"`
}

type RegistrationFormPasswordPayload struct {
	Password string          `json:"password"`
	Traits   json.RawMessage `json:"traits"`
}

func (s *Strategy) RegisterRegistrationRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) handleRegistrationError(_ http.ResponseWriter, r *http.Request, f *registration.Flow, p *RegistrationFormPayload, err error) error {
	if f != nil {
		f.UI.Nodes.Reset()
		if p != nil {
			for _, n := range container.NewFromJSON("", node.PasswordGroup, p.Password.Traits, "password.traits").Nodes {
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

func (s *Strategy) decode(p *RegistrationFormPayload, r *http.Request) error {
	raw, err := sjson.SetBytes(registrationSchema,
		"properties.password.properties.traits.$ref", s.d.Config(r.Context()).DefaultIdentityTraitsSchemaURL().String()+"#/properties/traits")
	if err != nil {
		return errors.WithStack(err)
	}

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return errors.WithStack(err)
	}

	return s.hd.Decode(r, p, compiler, decoderx.HTTPDecoderSetValidatePayloads(true), decoderx.HTTPDecoderJSONFollowsFormFormat())
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	if err := flow.MethodEnabledAndAllowed(r, s.ID().String(), s.d); err != nil {
		return err
	}

	var p RegistrationFormPayload
	if err := s.decode(&p, r); err != nil {
		return s.handleRegistrationError(w, r, f, &p, err)
	}

	if err := flow.EnsureCSRF(r, f.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return s.handleRegistrationError(w, r, f, &p, err)
	}

	if len(p.Password.Password) == 0 {
		return s.handleRegistrationError(w, r, f, &p, schema.NewRequiredError("#/password/password", "password"))
	}

	if len(p.Password.Traits) == 0 {
		p.Password.Traits = json.RawMessage("{}")
	}

	hpw, err := s.d.Hasher().Generate(r.Context(), []byte(p.Password.Password))
	if err != nil {
		return s.handleRegistrationError(w, r, f, &p, err)
	}

	co, err := json.Marshal(&CredentialsConfig{HashedPassword: string(hpw)})
	if err != nil {
		return s.handleRegistrationError(w, r, f, &p, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode password options to JSON: %s", err)))
	}

	i.Traits = identity.Traits(p.Password.Traits)
	i.SetCredentials(s.ID(), identity.Credentials{Type: s.ID(), Identifiers: []string{}, Config: co})

	if err := s.validateCredentials(r.Context(), i, p.Password.Password); err != nil {
		return s.handleRegistrationError(w, r, f, &p, err)
	}

	if err := s.d.RegistrationExecutor().PostRegistrationHook(w, r, identity.CredentialsTypePassword, f, i); err != nil {
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
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("No login identifiers (e.g. email, phone number, username) were set. Contact an administrator, the identity schema is misconfigured."))
	}

	for _, id := range c.Identifiers {
		if err := s.d.PasswordValidator().Validate(ctx, id, pw); err != nil {
			if _, ok := errorsx.Cause(err).(*herodot.DefaultError); ok {
				return err
			}
			return schema.NewPasswordPolicyViolationError("#/password/password", err.Error())
		}
	}

	return nil
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	nodes, err := container.NodesFromJSONSchema(node.PasswordGroup, s.d.Config(r.Context()).DefaultIdentityTraitsSchemaURL().String(), "password", nil)
	if err != nil {
		return err
	}

	nodes.Upsert(node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit))
	nodes.Upsert(NewPasswordNode("password.password"))

	for _, n := range nodes {
		f.UI.SetNode(n)
	}
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	// TODO: check prefix only password group?
	if err := f.UI.SortNodes(s.d.Config(r.Context()).DefaultIdentityTraitsSchemaURL().String(), "", []string{
		"method",
		x.CSRFTokenName,
		"password.password",
	}); err != nil {
		return err
	}

	return nil
}
