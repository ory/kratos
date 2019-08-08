package password

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/gojsonschema"
	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/schema"
	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/session"
	"github.com/ory/hive/x"
)

const (
	RegistrationPath = "/auth/browser/methods/password/registration"
)

type RegistrationFormPayload struct {
	Password string          `json:"password"`
	Traits   json.RawMessage `json:"traits"`
}

func (s *Strategy) setRegistrationRoutes(r *x.RouterPublic) {
	if _, _, ok := r.Lookup("POST", RegistrationPath); !ok {
		r.POST(RegistrationPath, s.d.SessionHandler().IsNotAuthenticated(s.handleRegistration, session.RedirectOnAuthenticated(s.c)))
	}
}

func (s *Strategy) handleRegistrationError(w http.ResponseWriter, r *http.Request, rr *selfservice.RegistrationRequest, err error) {
	s.d.SelfServiceRequestErrorHandler().HandleRegistrationError(w, r, CredentialsType, rr, err,
		&selfservice.ErrorHandlerOptions{
			AdditionalKeys: map[string]interface{}{
				selfservice.CSRFTokenName: s.cg(r),
			},
			IgnoreValuesForKeys: []string{"password"},
		},
	)
}

func (s *Strategy) handleRegistration(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid := r.URL.Query().Get("request")
	if len(rid) == 0 {
		s.handleRegistrationError(w, r, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request Code is missing.")))
		return
	}

	ar, err := s.d.RegistrationRequestManager().GetRegistrationRequest(r.Context(), rid)
	if err != nil {
		s.handleRegistrationError(w, r, nil, err)
		return
	}

	if err := ar.Valid(); err != nil {
		s.handleRegistrationError(w, r, ar, err)
		return
	}

	var p RegistrationFormPayload
	if err := s.dec.Decode(r, &p); err != nil {
		s.handleRegistrationError(w, r, ar, err)
		return
	}

	if len(p.Password) == 0 {
		s.handleRegistrationError(w, r, ar, errors.WithStack(schema.NewRequiredError("", gojsonschema.NewJsonContext("password", nil))))
		return
	}

	if len(p.Traits) == 0 {
		p.Traits = json.RawMessage("{}")
	}

	hpw, err := s.d.PasswordHasher().Generate([]byte(p.Password))
	if err != nil {
		s.handleRegistrationError(w, r, ar, err)
		return
	}

	co, err := json.Marshal(&CredentialsConfig{HashedPassword: string(hpw)})
	if err != nil {
		s.handleRegistrationError(w, r, ar, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode password options to JSON: %s", err)))
		return
	}

	i := identity.NewIdentity(s.c.DefaultIdentityTraitsSchemaURL().String())
	i.Traits = p.Traits
	i.SetCredentials(s.ID(), identity.Credentials{
		ID:          s.ID(),
		Identifiers: []string{},
		Options:     json.RawMessage(co),
	})

	if err := s.validateCredentials(i, p.Password); err != nil {
		s.handleRegistrationError(w, r, ar, err)
		return
	}

	if err := s.d.RegistrationExecutor().PostRegistrationHook(w, r,
		s.d.PostRegistrationHooks(CredentialsType),
		ar,
		i,
	); errors.Cause(err) == selfservice.ErrBreak {
		return
	} else if err != nil {
		s.handleRegistrationError(w, r, ar, err)
		return
	}
}

func (s *Strategy) validateCredentials(i *identity.Identity, pw string) error {
	ext := NewValidationExtension()
	if err := s.d.IdentityValidator().Validate(i, ext); err != nil {
		return err
	}

	c, ok := i.GetCredentials(CredentialsType)
	if !ok {
		// This should never happen
		panic(fmt.Sprintf("identity object did not provide the %s CredentialType unexpectedly", CredentialsType))
	} else if len(c.Identifiers) == 0 {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("No login identifiers (e.g. email, phone number, username) were set. Contact an administrator, the identity schema is misconfigured."))
	}

	for _, id := range c.Identifiers {
		if err := s.d.PasswordValidator().Validate(id, pw); err != nil {
			if _, ok := errors.Cause(err).(*herodot.DefaultError); ok {
				return err
			}
			return schema.NewPasswordPolicyValidation(
				"",
				err.Error(),
				gojsonschema.NewJsonContext("password", nil),
			)
		}
	}

	return nil
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, sr *selfservice.RegistrationRequest) error {
	action := urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), RegistrationPath),
		url.Values{"request": {sr.ID}},
	)

	sr.Methods[CredentialsType] = &selfservice.DefaultRequestMethod{
		Method: CredentialsType,
		Config: &RequestMethodConfig{
			Action: action.String(),
			Fields: selfservice.FormFields{
				"password": {
					Name:     "password",
					Type:     "password",
					Required: true,
				},
				selfservice.CSRFTokenName: {
					Name:     selfservice.CSRFTokenName,
					Type:     "hidden",
					Required: true,
					Value:    s.cg(r),
				},
			},
		},
	}
	return nil
}
