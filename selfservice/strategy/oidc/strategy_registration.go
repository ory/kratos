package oidc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"

	"github.com/google/go-jsonnet"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/x"
)

const (
	registrationFormPayloadSchema = `{
  "$id": "https://schemas.ory.sh/kratos/selfservice/oidc/registration/config.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {}
  }
}`
)

var _ registration.Strategy = new(Strategy)

func (s *Strategy) RegisterRegistrationRoutes(r *x.RouterPublic) {
	s.setRoutes(r)
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	if f.Type != flow.TypeBrowser {
		return nil
	}

	return s.populateMethod(r, f.UI)
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	if err := flow.MethodEnabledAndAllowedFromRequest(r, s.ID().String(), s.d); err != nil {
		return err
	}

	if err := r.ParseForm(); err != nil {
		return s.handleError(w, r, f.ID, "", nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error())))
	}

	var pid = r.Form.Get(s.SettingsStrategyID() + ".provider") // this can come from both url query and post body
	if pid == "" {
		return s.handleError(w, r, f.ID, pid, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`The HTTP request did not contain the required "%s.provider" form field`, s.SettingsStrategyID())))
	}

	provider, err := s.provider(r.Context(), r, pid)
	if err != nil {
		return s.handleError(w, r, f.ID, pid, nil, err)
	}

	c, err := provider.OAuth2(r.Context())
	if err != nil {
		return s.handleError(w, r, f.ID, pid, nil, err)
	}

	req, err := s.validateFlow(r.Context(), r, f.ID)
	if err != nil {
		return s.handleError(w, r, f.ID, pid, nil, err)
	}

	if s.alreadyAuthenticated(w, r, req) {
		return errors.WithStack(registration.ErrAlreadyLoggedIn)
	}

	state := x.NewUUID().String()
	if err := s.d.ContinuityManager().Pause(r.Context(), w, r, sessionName,
		continuity.WithPayload(&authCodeContainer{
			State:  state,
			FlowID: f.ID.String(),
			Form:   r.PostForm,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		return s.handleError(w, r, f.ID, pid, nil, err)
	}

	http.Redirect(w, r, c.AuthCodeURL(state, provider.AuthCodeURLOptions(req)...), http.StatusFound)

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) processRegistration(w http.ResponseWriter, r *http.Request, a *registration.Flow, claims *Claims, provider Provider, container *authCodeContainer) {
	if _, _, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), identity.CredentialsTypeOIDC, uid(provider.Config().ID, claims.Subject)); err == nil {
		// If the identity already exists, we should perform the login flow instead.

		// That will execute the "pre registration" hook which allows to e.g. disallow this flow. The registration
		// ui however will NOT be shown, instead the user is directly redirected to the auth path. That should then
		// do a silent re-request. While this might be a bit excessive from a network perspective it should usually
		// happen without any downsides to user experience as the request has already been authorized and should
		// not need additional consent/login.

		// This is kinda hacky but the only way to ensure seamless login/registration flows when using OIDC.
		s.d.Logger().WithRequest(r).WithField("provider", provider.Config().ID).
			WithField("subject", claims.Subject).
			Debug("Received successful OpenID Connect callback but user is already registered. Re-initializing login flow now.")

		// This endpoint only handles browser flow at the moment.
		ar, err := s.d.LoginHandler().NewLoginFlow(w, r, flow.TypeBrowser)
		if err != nil {
			s.handleError(w, r, a.GetID(), provider.Config().ID, nil, err)
			return
		}

		s.processLogin(w, r, ar, claims, provider, container)
		return
	}

	jn, err := s.f.Fetch(provider.Config().Mapper)
	if err != nil {
		s.handleError(w, r, a.GetID(), provider.Config().ID, nil, err)
		return
	}

	var jsonClaims bytes.Buffer
	if err := json.NewEncoder(&jsonClaims).Encode(claims); err != nil {
		s.handleError(w, r, a.GetID(), provider.Config().ID, nil, err)
		return
	}

	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)

	vm := jsonnet.MakeVM()
	vm.ExtCode("claims", jsonClaims.String())
	evaluated, err := vm.EvaluateSnippet(provider.Config().Mapper, jn.String())
	if err != nil {
		s.handleError(w, r, a.GetID(), provider.Config().ID, nil, err)
		return
	} else if traits := gjson.Get(evaluated, "identity.traits"); !traits.IsObject() {
		i.Traits = []byte{'{', '}'}
		s.d.Logger().
			WithRequest(r).
			WithField("oidc_provider", provider.Config().ID).
			WithSensitiveField("oidc_claims", claims).
			WithField("mapper_jsonnet_output", evaluated).
			WithField("mapper_jsonnet_url", provider.Config().Mapper).
			Error("OpenID Connect Jsonnet mapper did not return an object for key identity.traits. Please check your Jsonnet code!")
	} else {
		i.Traits = []byte(traits.Raw)
	}

	s.d.Logger().
		WithRequest(r).
		WithField("oidc_provider", provider.Config().ID).
		WithSensitiveField("oidc_claims", claims).
		WithField("mapper_jsonnet_output", evaluated).
		WithField("mapper_jsonnet_url", provider.Config().Mapper).
		Debug("OpenID Connect Jsonnet mapper completed.")

	option, err := decoderRegistration(s.d.Config(r.Context()).DefaultIdentityTraitsSchemaURL().String())
	if err != nil {
		s.handleError(w, r, a.GetID(), provider.Config().ID, nil, err)
		return
	}

	i.Traits, err = merge(container.Form.Encode(), json.RawMessage(i.Traits), option)
	if err != nil {
		s.handleError(w, r, a.GetID(), provider.Config().ID, nil, err)
		return
	}

	// Validate the identity itself
	if err := s.d.IdentityValidator().Validate(r.Context(), i); err != nil {
		s.handleError(w, r, a.GetID(), provider.Config().ID, i.Traits, err)
		return
	}

	creds, err := NewCredentials(provider.Config().ID, claims.Subject)
	if err != nil {
		s.handleError(w, r, a.GetID(), provider.Config().ID, i.Traits, err)
		return
	}

	i.SetCredentials(s.ID(), *creds)
	if err := s.d.RegistrationExecutor().PostRegistrationHook(w, r, identity.CredentialsTypeOIDC, a, i); err != nil {
		s.handleError(w, r, a.GetID(), provider.Config().ID, i.Traits, err)
		return
	}
}
