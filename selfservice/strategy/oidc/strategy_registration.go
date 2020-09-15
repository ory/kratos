package oidc

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/google/go-jsonnet"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/configuration"
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

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, sr *registration.Flow) error {
	if sr.Type != flow.TypeBrowser {
		return nil
	}

	config, err := s.populateMethod(r, sr.ID)
	if err != nil {
		return err
	}
	sr.Methods[s.ID()] = &registration.FlowMethod{
		Method: s.ID(),
		Config: &registration.FlowMethodConfig{FlowMethodConfigurator: config},
	}
	return nil
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

	i := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)

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

	option, err := decoderRegistration(s.c.DefaultIdentityTraitsSchemaURL().String())
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
	if err := s.d.IdentityValidator().Validate(i); err != nil {
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
