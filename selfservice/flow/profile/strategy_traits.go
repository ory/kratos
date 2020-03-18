package profile

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const FormTraitsID = "traits"

var _ Strategy = new(StrategyTraits)

type (
	strategyDependencies interface {
		x.CSRFTokenGeneratorProvider
	}
	StrategyTraits struct {
		c configuration.Provider
		d strategyDependencies
	}
)

// swagger:model traitsFormConfig
type TraitsRequestMethod struct {
	*form.HTMLForm
}

func NewStrategyTraits(d strategyDependencies, c configuration.Provider) *StrategyTraits {
	return &StrategyTraits{c: c, d: d}
}

func (s *StrategyTraits) ID() string {
	return FormTraitsID
}

func (s *StrategyTraits) RegisterProfileManagementRoutes(*x.RouterPublic) {
	return
}

func (s *StrategyTraits) PopulateProfileManagementMethod(r *http.Request, ss *session.Session, pr *Request) error {
	traitsSchema, err := s.c.IdentityTraitsSchemas().FindSchemaByID(ss.Identity.TraitsSchemaID)
	if err != nil {
		return err
	}

	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()
	registerNewDisableIdentifiersExtension(schemaCompiler)

	f, err := form.NewHTMLFormFromJSONSchema(urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), PublicProfileManagementUpdatePath),
		url.Values{"request": {pr.ID.String()}},
	).String(), traitsSchema.URL, "traits", schemaCompiler)
	if err != nil {
		return err
	}

	f.SetValuesFromJSON(json.RawMessage(ss.Identity.Traits), "traits")
	f.SetCSRF(s.d.GenerateCSRFToken(r))

	if err := f.SortFields(traitsSchema.URL, "traits"); err != nil {
		return err
	}

	pr.Methods[s.ID()] = &RequestMethod{
		Method: s.ID(),
		Config: &RequestMethodConfig{RequestMethodConfigurator: &TraitsRequestMethod{HTMLForm: f}},
	}
	return nil
}
