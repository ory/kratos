package password

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/ory/kratos/driver/configuration"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/x/errorsx"

	_ "github.com/ory/jsonschema/v3/fileloader"
	_ "github.com/ory/jsonschema/v3/httploader"
	"github.com/ory/x/decoderx"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RegistrationPath              = "/self-service/browser/flows/registration/strategies/password"
	registrationFormPayloadSchema = `{
  "$id": "https://schemas.ory.sh/kratos/selfservice/password/registration/config.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": [
    "password",
    "traits"
  ],
  "properties": {
    "password": {
      "type": "string",
      "minLength": 1
    },
    "traits": {}
  }
}`
)

type RegistrationFormPayload struct {
	Password string          `json:"password"`
	Traits   json.RawMessage `json:"traits"`
}

func (s *Strategy) RegisterRegistrationRoutes(r *x.RouterPublic) {
	r.POST(RegistrationPath, s.d.SessionHandler().IsNotAuthenticated(s.handleRegistration, session.RedirectOnAuthenticated(s.c)))
}

func (s *Strategy) handleRegistrationError(w http.ResponseWriter, r *http.Request, rr *registration.Request, p *RegistrationFormPayload, err error) {
	if rr != nil {
		if method, ok := rr.Methods[identity.CredentialsTypePassword]; ok {
			method.Config.Reset()

			if p != nil {
				for _, field := range form.NewHTMLFormFromJSON("", p.Traits, "traits").Fields {
					method.Config.SetField(field)
				}
			}

			method.Config.SetCSRF(s.d.GenerateCSRFToken(r))
			rr.Methods[identity.CredentialsTypePassword] = method
			if errSec := method.Config.SortFields(s.c.DefaultIdentityTraitsSchemaURL().String(), "traits"); errSec != nil {
				s.d.RegistrationRequestErrorHandler().HandleRegistrationError(w, r, identity.CredentialsTypePassword, rr, errors.Wrap(err, errSec.Error()))
				return
			}
		}
	}

	s.d.RegistrationRequestErrorHandler().HandleRegistrationError(w, r, identity.CredentialsTypePassword, rr, err)
}

func (s *Strategy) decoderRegistration() (decoderx.HTTPDecoderOption, error) {
	raw, err := sjson.SetBytes([]byte(registrationFormPayloadSchema), "properties.traits.$ref", s.c.DefaultIdentityTraitsSchemaURL().String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	o, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return o, nil
}

func (s *Strategy) handleRegistration(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid := x.ParseUUID(r.URL.Query().Get("request"))
	if x.IsZeroUUID(rid) {
		s.handleRegistrationError(w, r, nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request Code is missing.")))
		return
	}

	ar, err := s.d.RegistrationRequestPersister().GetRegistrationRequest(r.Context(), rid)
	if err != nil {
		s.handleRegistrationError(w, r, nil, nil, err)
		return
	}

	if err := ar.Valid(); err != nil {
		s.handleRegistrationError(w, r, ar, nil, err)
		return
	}

	var p RegistrationFormPayload
	option, err := s.decoderRegistration()
	if err != nil {
		s.handleRegistrationError(w, r, ar, nil, err)
		return
	}

	if err := decoderx.NewHTTP().Decode(r, &p,
		decoderx.HTTPFormDecoder(),
		option,
		decoderx.HTTPDecoderSetIgnoreParseErrorsStrategy(decoderx.ParseErrorIgnore),
		decoderx.HTTPDecoderSetValidatePayloads(false),
	); err != nil {
		s.handleRegistrationError(w, r, ar, &p, err)
		return
	}

	if len(p.Password) == 0 {
		s.handleRegistrationError(w, r, ar, &p, schema.NewRequiredError("#/", "password"))
		return
	}

	if len(p.Traits) == 0 {
		p.Traits = json.RawMessage("{}")
	}

	hpw, err := s.d.Hasher().Generate([]byte(p.Password))
	if err != nil {
		s.handleRegistrationError(w, r, ar, &p, err)
		return
	}

	co, err := json.Marshal(&CredentialsConfig{HashedPassword: string(hpw)})
	if err != nil {
		s.handleRegistrationError(w, r, ar, &p, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode password options to JSON: %s", err)))
		return
	}

	i := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
	i.Traits = identity.Traits(p.Traits)
	i.SetCredentials(s.ID(), identity.Credentials{Type: s.ID(), Identifiers: []string{}, Config: co})

	if err := s.validateCredentials(i, p.Password); err != nil {
		s.handleRegistrationError(w, r, ar, &p, err)
		return
	}

	if err := s.d.RegistrationExecutor().PostRegistrationHook(w, r, identity.CredentialsTypePassword, ar, i); err != nil {
		s.handleRegistrationError(w, r, ar, &p, err)
		return
	}
}

func (s *Strategy) validateCredentials(i *identity.Identity, pw string) error {
	if err := s.d.IdentityValidator().Validate(i); err != nil {
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
		if err := s.d.PasswordValidator().Validate(id, pw); err != nil {
			if _, ok := errorsx.Cause(err).(*herodot.DefaultError); ok {
				return err
			}
			return schema.NewPasswordPolicyViolationError("#/password", err.Error())
		}
	}

	return nil
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, sr *registration.Request) error {
	action := urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), RegistrationPath),
		url.Values{"request": {sr.ID.String()}},
	)

	htmlf, err := form.NewHTMLFormFromJSONSchema(action.String(), s.c.DefaultIdentityTraitsSchemaURL().String(), "traits", nil)
	if err != nil {
		return err
	}

	htmlf.Method = "POST"
	htmlf.SetCSRF(s.d.GenerateCSRFToken(r))
	htmlf.SetField(form.Field{Name: "password", Type: "password", Required: true})

	if err := htmlf.SortFields(s.c.DefaultIdentityTraitsSchemaURL().String(), "traits"); err != nil {
		return err
	}

	sr.Methods[identity.CredentialsTypePassword] = &registration.RequestMethod{
		Method: identity.CredentialsTypePassword,
		Config: &registration.RequestMethodConfig{RequestMethodConfigurator: &RequestMethod{HTMLForm: htmlf}},
	}

	return nil
}
