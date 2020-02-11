package hook

import (
	"encoding/json"
	"net/http"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/verify"
	"github.com/ory/kratos/session"
)

var _ registration.PostHookExecutor = new(Verifier)

type (
	verifierDependencies interface {
		verify.PersistenceProvider
		verify.ManagementProvider
		identity.ValidationProvider
	}
	Verifier struct {
		v *schema.Validator
		r verifierDependencies
		c configuration.Provider
	}
)

func NewVerifier(r verifierDependencies, c configuration.Provider) *Verifier {
	return &Verifier{
		v: schema.NewValidator(),
		r: r,
		c: c,
	}
}

func (e *Verifier) ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, _ *registration.Request, s *session.Session) error {
	sc, err := e.c.IdentityTraitsSchemas().FindSchemaByID(s.Identity.TraitsSchemaID)
	if err != nil {
		return err
	}

	extension := verify.NewValidationExtensionRunner(s.Identity, e.c.SelfServiceVerificationLinkLifespan())
	er, err := schema.NewExtensionRunner(schema.ExtensionRunnerIdentityMetaSchema, extension.Runner)
	if err != nil {
		return err
	}

	if err := e.v.Validate(sc.URL, json.RawMessage(s.Identity.Traits), schema.WithExtensionRunner(er)); err != nil {
		return err
	}

	return e.r.VerificationManager().TrackAndSend(r.Context(), extension.Addresses())
}
