package password

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang/gddo/httputil"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/gojsonschema"
	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/schema"
	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/x"
)

const (
	RegistrationPath = "/auth/browser/registration/methods/password"
)

func (s *Strategy) setRegistrationRoutes(r *x.RouterPublic) {
	if _, _, ok := r.Lookup("POST", RegistrationPath); !ok {
		r.POST(RegistrationPath, s.handleRegistration)
	}
}

func (s *Strategy) handleRegistrationError(w http.ResponseWriter, r *http.Request, rr *selfservice.RegistrationRequest, err error) {
	if rr == nil {
		rr = NewBlankRegistrationRequest("")
	}

	if httputil.NegotiateContentType(
		r,
		[]string{"application/json", "text/html", "text/*", "*/*"},
		"text/*",
	) == "application/json" {
		switch errors.Cause(err).(type) {
		case schema.ResultErrors:
			s.d.Writer().WriteErrorCode(w, r, http.StatusBadRequest, err)
		default:
			s.d.Writer().WriteError(w, r, err)
		}
		return
	}

	var tc = func() *RegistrationRequestMethodConfig {
		if method, ok := rr.Methods[CredentialsType]; !ok {
			panic(fmt.Sprintf(`*selfservice.RegistrationRequest.Methods must have CredentialsType "%s" but did not: %+v`, CredentialsType, rr.Methods))
		} else if mc, ok := method.Config.(*RegistrationRequestMethodConfig); !ok {
			panic(fmt.Sprintf(`*selfservice.RegistrationRequest.Methods[%s].Config must be of type "*RegistrationRequestMethodConfig" but got: %T`, CredentialsType, method.Config))
		} else {
			return mc
		}
	}

	switch et := errors.Cause(err).(type) {
	case *herodot.DefaultError:
		if et.Error() == selfservice.ErrRegistrationRequestExpired.Error() {
			config := tc()
			config.Reset()
			config.Error = et.Reason()
			if err := s.d.RegistrationRequestManager().UpdateRegistrationRequest(r.Context(), rr.ID, CredentialsType, config); err != nil {
				s.d.ErrorManager().ForwardError(w, r, err)
				return
			}

			http.Redirect(w,
				r,
				urlx.CopyWithQuery(s.c.RegisterURL(), url.Values{"request": {rr.ID}}).String(),
				http.StatusFound,
			)
			return
		}
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	case schema.ResultErrors:
		config := tc()
		config.Reset()
		for k := range tidyForm(r.PostForm) {
			config.Fields.SetValue(k, r.PostForm.Get(k))
		}
		config.Fields.SetValue(csrfTokenName, s.cg(r))

		for k, e := range et {
			herodot.DefaultErrorLogger(s.d.Logger(), err).Warnf("A form error occurred during registration (%d of %d): %s", k+1, len(et), e.String())
			name := e.Field()
			if trimmed := strings.TrimPrefix(e.Field(), "traits."); trimmed != e.Field() {
				// has traits prefix
				name = fmt.Sprintf("traits[%s]", trimmed)
			}
			config.Fields.SetError(
				name,
				e.String(),
			)
		}

		if err := s.d.RegistrationRequestManager().UpdateRegistrationRequest(r.Context(), rr.ID, CredentialsType, config); err != nil {
			s.d.ErrorManager().ForwardError(w, r, err)
			return
		}

		http.Redirect(w,
			r,
			urlx.CopyWithQuery(s.c.RegisterURL(), url.Values{"request": {rr.ID}}).String(),
			http.StatusFound,
		)
	default:
		s.d.ErrorManager().ForwardError(w, r, err)
	}
}

func (s *Strategy) handleRegistration(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid := r.URL.Query().Get("request")
	if len(rid) == 0 {
		s.handleRegistrationError(w, r, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request ID is missing.")))
		return
	}

	ar, err := s.d.RegistrationRequestManager().GetRegistrationRequest(r.Context(), rid)
	if err != nil {
		s.handleRegistrationError(w, r, NewBlankRegistrationRequest(rid), err)
		return
	}

	p, err := s.dec.Decode(r)
	if err != nil {
		s.handleRegistrationError(w, r, ar, err)
		return
	}

	if err := ar.Valid(); err != nil {
		s.handleRegistrationError(w, r, ar, err)
		return
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

	sr.Methods[CredentialsType] = &selfservice.RegistrationRequestMethod{
		Method: CredentialsType,
		Config: &RegistrationRequestMethodConfig{
			Action: action.String(),
			Fields: FormFields{
				"password": {
					Name:     "password",
					Type:     "password",
					Required: true,
				},
				csrfTokenName: {
					Name:     csrfTokenName,
					Type:     "hidden",
					Required: true,
					Value:    s.cg(r),
				},
			},
		},
	}
	return nil
}
