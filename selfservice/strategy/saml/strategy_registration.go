package saml

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"

	"github.com/tidwall/sjson"

	"github.com/ory/kratos/x"
)

// Implement the interface
var _ registration.Strategy = new(Strategy)

// Call at the creation of Kratos, when Kratos implement all authentication routes
func (s *Strategy) RegisterRegistrationRoutes(r *x.RouterPublic) {
	s.setRoutes(r)
}

func (s *Strategy) createIdentity(w http.ResponseWriter, r *http.Request, a *registration.Flow, claims *Claims, provider Provider) (*identity.Identity, error) {
	var jsonClaims bytes.Buffer
	if err := json.NewEncoder(&jsonClaims).Encode(claims); err != nil {
		return nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	i := identity.NewIdentity(s.d.Config().DefaultIdentityTraitsSchemaID(r.Context()))
	if err := s.setTraits(w, r, a, claims, provider, jsonClaims, i); err != nil {
		return nil, s.handleError(w, r, a, provider.Config().ID, i.Traits, err)
	}

	s.d.Logger().
		WithRequest(r).
		WithField("saml_provider", provider.Config().ID).
		WithSensitiveField("saml_claims", claims).
		Debug("SAML Connect completed.")
	return i, nil
}

func (s *Strategy) setTraits(w http.ResponseWriter, r *http.Request, a *registration.Flow, claims *Claims, provider Provider, jsonClaims bytes.Buffer, i *identity.Identity) error {

	traitsMap := make(map[string]interface{})
	json.Unmarshal(jsonClaims.Bytes(), &traitsMap)
	delete(traitsMap, "iss")
	delete(traitsMap, "email_verified")
	delete(traitsMap, "sub")
	traits, err := json.Marshal(traitsMap)
	if err != nil {
		return s.handleError(w, r, a, provider.Config().ID, i.Traits, err)
	}
	i.Traits = identity.Traits(traits)

	s.d.Logger().
		WithRequest(r).
		WithField("oidc_provider", provider.Config().ID).
		WithSensitiveField("identity_traits", i.Traits).
		WithField("mapper_jsonnet_url", provider.Config().Mapper).
		Debug("Merged form values and OpenID Connect Jsonnet output.")
	return nil
}

func (s *Strategy) processRegistration(w http.ResponseWriter, r *http.Request, a *registration.Flow, provider Provider, claims *Claims) error {
	i, err := s.createIdentity(w, r, a, claims, provider)
	if err != nil {
		return s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	// Verify the identity
	if err := s.d.IdentityValidator().Validate(r.Context(), i); err != nil {
		return s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	// Create new uniq credentials identifier for user is database
	creds, err := identity.NewCredentialsSAML(claims.Subject, provider.Config().ID)
	if err != nil {
		return s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	// Set the identifiers to the identity
	i.SetCredentials(s.ID(), *creds)
	if err := s.d.RegistrationExecutor().PostRegistrationHook(w, r, identity.CredentialsTypeSAML, a, i); err != nil {
		return s.handleError(w, r, a, provider.Config().ID, i.Traits, err)
	}

	return nil
}

// Method not used but necessary to implement the interface
func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	if f.Type != flow.TypeBrowser {
		return nil
	}

	return s.populateMethod(r, f.UI, text.NewInfoRegistrationWith)
}

func (s *Strategy) newLinkDecoder(p interface{}, r *http.Request) error {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	raw, err := sjson.SetBytes(linkSchema, "properties.traits.$ref", ds.String()+"#/properties/traits")
	if err != nil {
		return errors.WithStack(err)
	}

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.hd.Decode(r, &p, compiler,
		decoderx.HTTPKeepRequestBody(true),
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderUseQueryAndBody(),
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Not needed in SAML
func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	return flow.ErrStrategyNotResponsible
}
