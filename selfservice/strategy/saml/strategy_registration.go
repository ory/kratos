package saml

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/go-jsonnet"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/kratos/x"
)

// Implement the interface
var _ registration.Strategy = new(Strategy)

// Call at the creation of Kratos, when Kratos implement all authentication routes
func (s *Strategy) RegisterRegistrationRoutes(r *x.RouterPublic) {
	s.setRoutes(r)
}

func (s *Strategy) GetRegistrationIdentity(r *http.Request, ctx context.Context, provider Provider, claims *Claims, logsEnabled bool) (*identity.Identity, error) {
	// Fetch fetches the file contents from the mapper file.
	jn, err := s.f.Fetch(provider.Config().Mapper)
	if err != nil {
		return nil, err
	}

	var jsonClaims bytes.Buffer
	if err := json.NewEncoder(&jsonClaims).Encode(claims); err != nil {
		return nil, err
	}

	// Identity Creation
	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)

	vm := jsonnet.MakeVM()
	vm.ExtCode("claims", jsonClaims.String())
	evaluated, err := vm.EvaluateAnonymousSnippet(provider.Config().Mapper, jn.String())
	if err != nil {
		return nil, err
	} else if traits := gjson.Get(evaluated, "identity.traits"); !traits.IsObject() {
		i.Traits = []byte{'{', '}'}
		if logsEnabled {
			s.d.Logger().
				WithRequest(r).
				WithField("Provider", provider.Config().ID).
				WithSensitiveField("saml_claims", claims).
				WithField("mapper_jsonnet_output", evaluated).
				WithField("mapper_jsonnet_url", provider.Config().Mapper).
				Error("SAML Jsonnet mapper did not return an object for key identity.traits. Please check your Jsonnet code!")
		}
	} else {
		i.Traits = []byte(traits.Raw)
	}

	if logsEnabled {
		s.d.Logger().
			WithRequest(r).
			WithField("saml_provider", provider.Config().ID).
			WithSensitiveField("saml_claims", claims).
			WithSensitiveField("mapper_jsonnet_output", evaluated).
			WithField("mapper_jsonnet_url", provider.Config().Mapper).
			Debug("SAML Jsonnet mapper completed.")

		s.d.Logger().
			WithRequest(r).
			WithField("saml_provider", provider.Config().ID).
			WithSensitiveField("identity_traits", i.Traits).
			WithSensitiveField("mapper_jsonnet_output", evaluated).
			WithField("mapper_jsonnet_url", provider.Config().Mapper).
			Debug("Merged form values and SAML Jsonnet output.")
	}

	// Verify the identity
	if err := s.d.IdentityValidator().Validate(ctx, i); err != nil {
		return i, err
	}

	// Create new uniq credentials identifier for user is database
	creds, err := identity.NewCredentialsSAML(claims.Subject, provider.Config().ID)
	if err != nil {
		return i, err
	}

	// Set the identifiers to the identity
	i.SetCredentials(s.ID(), *creds)

	return i, nil
}

func (s *Strategy) processRegistration(w http.ResponseWriter, r *http.Request, a *registration.Flow, provider Provider, claims *Claims) error {

	i, err := s.GetRegistrationIdentity(r, r.Context(), provider, claims, true)
	if err != nil {
		if i == nil {
			return s.handleError(w, r, a, provider.Config().ID, nil, err)
		} else {
			return s.handleError(w, r, a, provider.Config().ID, i.Traits, err)
		}
	}

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
