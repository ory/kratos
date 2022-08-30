package saml

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

// Implement the interface
var _ login.Strategy = new(Strategy)

// Call at the creation of Kratos, when Kratos implement all authentication routes
func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	s.setRoutes(r)
}

// SubmitSelfServiceLoginFlowWithSAMLMethodBody is used to decode the login form payload
// when using the saml method.
//
// swagger:model SubmitSelfServiceLoginFlowWithSAMLMethodBody
type SubmitSelfServiceLoginFlowWithSAMLMethodBody struct {
	// The provider to register with
	//
	// required: true
	Provider string `json:"samlProvider"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token"`

	// Method to use
	//
	// This field must be set to `oidc` when using the oidc method.
	//
	// required: true
	Method string `json:"method"`

	// The identity traits. This is a placeholder for the registration flow.
	Traits json.RawMessage `json:"traits"`
}

// Login and give a session to the user
func (s *Strategy) processLogin(w http.ResponseWriter, r *http.Request, a *login.Flow, provider Provider, c *identity.Credentials, i *identity.Identity, claims *Claims) (*registration.Flow, error) {

	s.updateIdentityTraits(i, provider, claims)

	var o identity.CredentialsSAML
	if err := json.NewDecoder(bytes.NewBuffer(c.Config)).Decode(&o); err != nil {
		return nil, s.handleError(w, r, a, provider.Config().ID, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error())))
	}

	sess := session.NewInactiveSession()                                  // Creation of an inactive session
	sess.CompletedLoginFor(s.ID(), identity.AuthenticatorAssuranceLevel1) // Add saml to the Authentication Method References

	if err := s.d.LoginHookExecutor().PostLoginHook(w, r, node.SAMLGroup, a, i, sess); err != nil {
		return nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	return nil, nil
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, ss *session.Session) (i *identity.Identity, err error) {
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, err
	}

	var p SubmitSelfServiceLoginFlowWithSAMLMethodBody
	if err := s.newLinkDecoder(&p, r); err != nil {
		return nil, s.handleError(w, r, f, "", nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error())))
	}

	var pid = p.Provider // This can come from both url query and post body
	if pid == "" {
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), s.ID().String(), s.ID().String(), s.d); err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	req, err := s.validateFlow(r.Context(), r, f.ID)
	if err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	if s.alreadyAuthenticated(w, r, req) {
		return
	}

	state := x.NewUUID().String()
	if err := s.d.RelayStateContinuityManager().Pause(r.Context(), w, r, sessionName,
		continuity.WithPayload(&authCodeContainer{
			State:  state,
			FlowID: f.ID.String(),
			Traits: p.Traits,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		return nil, s.handleError(w, r, f, pid, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithDebug(err.Error())))
	}

	if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(RouteBaseAuth))
	} else {
		http.Redirect(w, r, RouteBaseAuth+"/"+pid, http.StatusSeeOther)
	}

	return nil, errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, l *login.Flow) error {
	if l.Type != flow.TypeBrowser {
		return nil
	}

	// This strategy can only solve AAL1
	if requestedAAL > identity.AuthenticatorAssuranceLevel1 {
		return nil
	}

	return s.populateMethod(r, l.UI, text.NewInfoLoginWith)
}

// In order to do a JustInTimeProvisioning, it is important to update the identity traits at each new SAML connection
func (s *Strategy) updateIdentityTraits(i *identity.Identity, provider Provider, claims *Claims) error {
	jn, err := s.f.Fetch(provider.Config().Mapper)
	if err != nil {
		return nil
	}

	var jsonClaims bytes.Buffer
	if err := json.NewEncoder(&jsonClaims).Encode(claims); err != nil {
		return nil
	}

	vm := jsonnet.MakeVM()
	vm.ExtCode("claims", jsonClaims.String())
	evaluated, err := vm.EvaluateAnonymousSnippet(provider.Config().Mapper, jn.String())
	if err != nil {
		return err
	} else if traits := gjson.Get(evaluated, "identity.traits"); !traits.IsObject() {
		i.Traits = []byte{'{', '}'}
		return errors.New("SAML Jsonnet mapper did not return an object for key identity.traits. Please check your Jsonnet code!")
	} else {
		i.Traits = []byte(traits.Raw)
		return nil
	}

}
