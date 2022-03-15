package flowhelpers

import (
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
)

// GuessForcedLoginIdentifier returns the identifier for login flows where the identity needs to refresh the session.
func GuessForcedLoginIdentifier(r *http.Request, d interface {
	session.ManagementProvider
	identity.PrivilegedPoolProvider
}, f interface {
	IsForced() bool
}, ct identity.CredentialsType) (identifier string, id *identity.Identity, creds *identity.Credentials) {
	var ok bool
	// This block adds the identifier to the method when the request is forced - as a hint for the user.
	if !f.IsForced() {
		// do nothing
	} else if sess, err := d.SessionManager().FetchFromRequest(r.Context(), r); err != nil {
		// do nothing
	} else if id, err = d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), sess.IdentityID); err != nil {
		// do nothing
	} else if creds, ok = id.GetCredentials(ct); !ok {
		// do nothing
	} else if len(creds.Identifiers) == 0 {
		// do nothing
	} else {
		identifier = creds.Identifiers[0]
	}
	return
}
