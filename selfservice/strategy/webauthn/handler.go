package webauthn

import (
	_ "embed"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/x"
)

//go:embed js/webauthn.js
var jsOnLoad []byte

const webAuthnRoute = "/.well-known/ory/webauthn.js"

// swagger:model webAuthnJavaScript
type webAuthnJavaScript string

// swagger:route GET /.well-known/ory/webauthn.js v0alpha2 getWebAuthnJavaScript
//
// # Get WebAuthn JavaScript
//
// This endpoint provides JavaScript which is needed in order to perform WebAuthn login and registration.
//
// If you are building a JavaScript Browser App (e.g. in ReactJS or AngularJS) you will need to load this file:
//
//	```html
//	<script src="https://public-kratos.example.org/.well-known/ory/webauthn.js" type="script" async />
//	```
//
// More information can be found at [Ory Kratos User Login](https://www.ory.sh/docs/kratos/self-service/flows/user-login) and [User Registration Documentation](https://www.ory.sh/docs/kratos/self-service/flows/user-registration).
//
//	Produces:
//	- text/javascript
//
//	Schemes: http, https
//
//	Responses:
//	  200: webAuthnJavaScript
func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	if handle, _, _ := r.Lookup("GET", webAuthnRoute); handle == nil {
		r.GET(webAuthnRoute, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
			_, _ = w.Write([]byte(webAuthnJavaScript(jsOnLoad)))
		})
	}
}
