package saml

import (
	"errors"
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/x/sqlcon"
)

// Handle SAML Assertion and process to either login or register
func (s *Strategy) processLoginOrRegister(w http.ResponseWriter, r *http.Request, loginFlow *login.Flow, provider Provider, claims *Claims) (*flow.Flow, error) {

	// If the user's ID is null, we have to handle error
	if claims.Subject == "" {
		return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, errors.New("the user ID is empty: the problem probably comes from the mapping between the SAML attributes and the identity attributes"))
	}

	// This is a check to see if the user exists in the database
	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), identity.CredentialsTypeSAML, identity.SAMLUniqueID(provider.Config().ID, claims.Subject))

	if err != nil {
		// ErrNoRows is returned when a SQL SELECT statement returns no rows.
		if errors.Is(err, sqlcon.ErrNoRows) {

			// The user doesn't net existe yet so we register him
			registerFlow, err := s.d.RegistrationHandler().NewRegistrationFlow(w, r, flow.TypeBrowser)
			if err != nil {
				if i == nil {
					return nil, s.handleError(w, r, registerFlow, provider.Config().ID, nil, err)
				} else {
					return nil, s.handleError(w, r, registerFlow, provider.Config().ID, i.Traits, err)
				}
			}

			if err = s.processRegistration(w, r, registerFlow, provider, claims); err != nil {
				if i == nil {
					return nil, s.handleError(w, r, registerFlow, provider.Config().ID, nil, err)
				} else {
					return nil, s.handleError(w, r, registerFlow, provider.Config().ID, i.Traits, err)
				}
			}

			return nil, nil

		} else {
			return nil, err
		}
	} else {
		// The user already exist in database so we log him
		// loginFlow, err := s.d.LoginHandler().NewLoginFlow(w, r, flow.TypeBrowser)
		if err != nil {
			if i == nil {
				return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, err)
			} else {
				return nil, s.handleError(w, r, loginFlow, provider.Config().ID, i.Traits, err)
			}
		}
		if _, err = s.processLogin(w, r, loginFlow, provider, c, i, claims); err != nil {
			if i == nil {
				return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, err)
			} else {
				return nil, s.handleError(w, r, loginFlow, provider.Config().ID, i.Traits, err)
			}
		}
		return nil, nil
	}
}
