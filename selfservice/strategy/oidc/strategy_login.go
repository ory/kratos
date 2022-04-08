package oidc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/schema"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/fetcher"

	"golang.org/x/oauth2"

	"github.com/ory/kratos/session"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/registration"

	"github.com/ory/kratos/text"

	"github.com/ory/kratos/continuity"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

var _ login.Strategy = new(Strategy)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	s.setRoutes(r)
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, l *login.Flow) error {
	// This strategy can only solve AAL1
	if requestedAAL > identity.AuthenticatorAssuranceLevel1 {
		return nil
	}

	return s.populateMethod(r, l.UI, text.NewInfoLoginWith)
}

// SubmitSelfServiceLoginFlowWithOidcMethodBody is used to decode the login form payload
// when using the oidc method.
//
// swagger:model submitSelfServiceLoginFlowWithOidcMethodBody
type SubmitSelfServiceLoginFlowWithOidcMethodBody struct {
	// The provider to register with
	//
	// required: true
	Provider string `json:"provider"`

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

	// Only used in API-type flows, when an id token has been received by mobile app directly from oidc provider.
	//
	// required: false
	IdToken string `json:"id_token"`
}

func (s *Strategy) processLogin(w http.ResponseWriter, r *http.Request, a *login.Flow, token *oauth2.Token, claims *Claims, provider Provider, container *authCodeContainer) (*registration.Flow, error) {
	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), identity.CredentialsTypeOIDC, identity.OIDCUniqueID(provider.Config().ID, claims.Subject))
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			// If no account was found we're "manually" creating a new registration flow and redirecting the browser
			// to that endpoint.

			// That will execute the "pre registration" hook which allows to e.g. disallow this request. The registration
			// ui however will NOT be shown, instead the user is directly redirected to the auth path. That should then
			// do a silent re-request. While this might be a bit excessive from a network perspective it should usually
			// happen without any downsides to user experience as the flow has already been authorized and should
			// not need additional consent/login.

			// This is kinda hacky but the only way to ensure seamless login/registration flows when using OIDC.

			s.d.Logger().WithField("provider", provider.Config().ID).WithField("subject", claims.Subject).Debug("Received successful OpenID Connect callback but user is not registered. Re-initializing registration flow now.")

			// This flow only works for browsers anyways.
			aa, err := s.d.RegistrationHandler().NewRegistrationFlow(w, r, flow.TypeBrowser)
			if err != nil {
				return nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
			}

			if _, err := s.processRegistration(w, r, aa, token, claims, provider, container); err != nil {
				return aa, err
			}

			return nil, nil
		}

		return nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	var o identity.CredentialsOIDC
	if err := json.NewDecoder(bytes.NewBuffer(c.Config)).Decode(&o); err != nil {
		return nil, s.handleError(w, r, a, provider.Config().ID, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error())))
	}

	sess := session.NewInactiveSession()
	sess.CompletedLoginFor(s.ID(), identity.AuthenticatorAssuranceLevel1)
	for _, c := range o.Providers {
		if c.Subject == claims.Subject && c.Provider == provider.Config().ID {
			if err = s.d.LoginHookExecutor().PostLoginHook(w, r, a, i, sess); err != nil {
				return nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
			}
			return nil, nil
		}
	}

	return nil, s.handleError(w, r, a, provider.Config().ID, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to find matching OpenID Connect Credentials.").WithDebugf(`Unable to find credentials that match the given provider "%s" and subject "%s".`, provider.Config().ID, claims.Subject)))
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, ss *session.Session) (i *identity.Identity, err error) {
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, err
	}

	var pid = ""
	var idToken = ""

	var p SubmitSelfServiceLoginFlowWithOidcMethodBody
	if f.Type == flow.TypeBrowser {
		if err := s.newLinkDecoder(&p, r); err != nil {
			return nil, s.handleError(w, r, f, "", nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error())))
		}
		pid = p.Provider // this can come from both url query and post body
	} else {
		if err := flow.MethodEnabledAndAllowedFromRequest(r, s.ID().String(), s.d); err != nil {
			return nil, err
		}

		if err := s.dec.Decode(r, &p,
			decoderx.HTTPDecoderSetValidatePayloads(true),
			decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
			decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
			return nil, s.handleError(w, r, f, "", nil, err)
		}

		idToken = p.IdToken
		pid = p.Provider
	}

	if pid == "" {
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), s.SettingsStrategyID(), s.SettingsStrategyID(), s.d); err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	provider, err := s.provider(r.Context(), r, pid)
	if err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	c, err := provider.OAuth2(r.Context())
	if err != nil {
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
	if f.Type == flow.TypeBrowser {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r, sessionName,
			continuity.WithPayload(&authCodeContainer{
				State:  state,
				FlowID: f.ID.String(),
				Traits: p.Traits,
			}),
			continuity.WithLifespan(time.Minute*30)); err != nil {
			return nil, s.handleError(w, r, f, pid, nil, err)
		}

		http.Redirect(w, r, c.AuthCodeURL(state, provider.AuthCodeURLOptions(req)...), http.StatusFound)
	} else if f.Type == flow.TypeAPI {
		var claims *Claims
		if apiFlowProvider, ok := provider.(APIFlowProvider); ok {
			if len(idToken) > 0 {
				claims, err = apiFlowProvider.ClaimsFromIdToken(r.Context(), idToken)
				if err != nil {
					return nil, errors.WithStack(err)
				}
			} else {
				return nil, s.handleError(w, r, f, p.Provider, nil, ErrApiTokenMissing)
			}
		} else {
			return nil, s.handleError(w, r, f, p.Provider, nil, ErrProviderNoAPISupport)
		}

		i, _, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(),
			identity.CredentialsTypeOIDC,
			identity.OIDCUniqueID(provider.Config().ID, claims.Subject))
		if err != nil {
			if !errors.Is(err, sqlcon.ErrNoRows) {
				return nil, err
			}
			i = identity.NewIdentity(s.d.Config(r.Context()).DefaultIdentityTraitsSchemaID())

			fetch := fetcher.NewFetcher(fetcher.WithClient(s.d.HTTPClient(r.Context())))
			jn, err := fetch.Fetch(provider.Config().Mapper)
			if err != nil {
				return nil, s.handleError(w, r, f, provider.Config().ID, nil, err)
			}

			var jsonClaims bytes.Buffer
			if err := json.NewEncoder(&jsonClaims).Encode(claims); err != nil {
				return nil, s.handleError(w, r, f, provider.Config().ID, nil, err)
			}

			vm := jsonnet.MakeVM()
			vm.ExtCode("claims", jsonClaims.String())
			evaluated, err := vm.EvaluateAnonymousSnippet(provider.Config().Mapper, jn.String())
			if err != nil {
				return nil, s.handleError(w, r, f, provider.Config().ID, nil, err)
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

			//option, err := decoderRegistration(s.d.Config(r.Context()).DefaultIdentityTraitsSchemaURL().String())
			//if err != nil {
			//	return nil, s.handleError(w, r, f, provider.Config().ID, nil, err)
			//}
			//
			//i.Traits, err = merge(container.Form.Encode(), json.RawMessage(i.Traits), option)
			//if err != nil {
			//	return nil, s.handleError(w, r, f, provider.Config().ID, nil, err)
			//}

			// Validate the identity itself
			if err := s.d.IdentityValidator().Validate(r.Context(), i); err != nil {
				return nil, s.handleError(w, r, f, provider.Config().ID, i.Traits, err)
			}

			creds, err := identity.NewCredentialsOIDC(
				idToken,
				"",
				"",
				provider.Config().ID,
				claims.Subject)
			if err != nil {
				return nil, s.handleError(w, r, f, provider.Config().ID, i.Traits, err)
			}

			i.SetCredentials(s.ID(), *creds)

			if err := s.d.IdentityManager().Create(r.Context(), i); err != nil {
				if errors.Is(err, sqlcon.ErrUniqueViolation) {
					return nil, schema.NewDuplicateCredentialsError()
				}
				return nil, errors.WithStack(err)
			}

			//i.Traits = identity.Traits(fmt.Sprintf("{\"phone\": \"%s\"}", p.Phone))
			//
			//if err := s.d.IdentityValidator().Validate(r.Context(), i); err != nil {
			//	return nil, err
			//} else if err := s.d.IdentityManager().Create(r.Context(), i); err != nil {
			//	if errors.Is(err, sqlcon.ErrUniqueViolation) {
			//		return nil, schema.NewDuplicateCredentialsError()
			//	}
			//	return nil, err
			//}
			//
			//s.d.Audit().
			//	WithRequest(r).
			//	WithField("identity_id", i.ID).
			//	Info("A new identity has registered using self-service login auto-provisioning.")

		}
		return i, nil
	} else {
		return nil, errors.WithStack(errors.New(fmt.Sprintf("Not supported flow type: %s", f.Type)))
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		return nil, s.handleError(w, r, f, pid, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithDebug(err.Error())))
	}

	codeURL := c.AuthCodeURL(state, provider.AuthCodeURLOptions(req)...)
	if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(codeURL))
	} else {
		http.Redirect(w, r, codeURL, http.StatusSeeOther)
	}

	return nil, errors.WithStack(flow.ErrCompletedByStrategy)
}
