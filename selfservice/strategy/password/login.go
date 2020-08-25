package password

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

const (
	RouteLogin = "/self-service/login/methods/password"
)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	s.d.CSRFHandler().ExemptPath(RouteLogin)
	r.POST(RouteLogin, s.handleLogin)
}

func (s *Strategy) handleLoginError(w http.ResponseWriter, r *http.Request, rr *login.Flow, payload *LoginFormPayload, err error) {
	if rr != nil {
		if method, ok := rr.Methods[identity.CredentialsTypePassword]; ok {
			method.Config.Reset()
			method.Config.SetValue("identifier", payload.Identifier)
			if rr.Type == flow.TypeBrowser {
				method.Config.SetCSRF(s.d.GenerateCSRFToken(r))
			}

			rr.Methods[identity.CredentialsTypePassword] = method
		}
	}

	s.d.LoginFlowErrorHandler().WriteFlowError(w, r, identity.CredentialsTypePassword, rr, err)
}

// nolint:deadcode,unused
// swagger:parameters completeSelfServiceLoginFlowWithPasswordMethod
type completeSelfServiceLoginFlowWithPasswordMethod struct {
	// The Flow ID
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// in: body
	LoginFormPayload
}

// swagger:route GET /self-service/login/methods/password public completeSelfServiceLoginFlowWithPasswordMethod
//
// Complete Login Flow with Username/Email Password Method
//
// Use this endpoint to complete a login flow by sending an identity's identifier and password. This endpoint
// behaves differently for API and browser flows.
//
// API flows expect `application/json` to be sent in the body and responds with
//   - HTTP 200 and a application/json body with the session token on success;
//   - HTTP 302 redirect to a fresh login flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//
// Browser flows expect `application/x-www-form-urlencoded` to be sent in the body and responds with
//   - a HTTP 302 redirect to the post/after login URL or the `return_to` value if it was set and if the login succeeded;
//   - a HTTP 302 redirect to the login UI URL with the flow ID containing the validation errors otherwise.
//
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Schemes: http, https
//
//     Consumes:
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Produces:
//     - application/json
//
//     Responses:
//       200: loginViaApiResponse
//       302: emptyResponse
//       400: genericError
//       500: genericError
func (s *Strategy) handleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid := x.ParseUUID(r.URL.Query().Get("flow"))
	if x.IsZeroUUID(rid) {
		s.handleLoginError(w, r, nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The flow query parameter is missing or invalid.")))
		return
	}

	ar, err := s.d.LoginFlowPersister().GetLoginFlow(r.Context(), rid)
	if err != nil {
		s.handleLoginError(w, r, nil, nil, err)
		return
	}

	var p LoginFormPayload
	if err := s.hd.Decode(r, &p, decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema)); err != nil {
		s.handleLoginError(w, r, ar, &p, err)
		return
	}

	if ar.Type == flow.TypeBrowser && !nosurf.VerifyToken(s.d.GenerateCSRFToken(r), p.CSRFToken) {
		s.handleLoginError(w, r, ar, &p, x.ErrInvalidCSRFToken)
		return
	}

	if _, err := s.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil && !ar.Forced {
		if ar.Type == flow.TypeBrowser {
			http.Redirect(w, r, s.c.SelfServiceBrowserDefaultReturnTo().String(), http.StatusFound)
			return
		}

		s.d.Writer().WriteError(w, r, errors.WithStack(login.ErrAlreadyLoggedIn))
		return
	}

	if err := ar.Valid(); err != nil {
		s.handleLoginError(w, r, ar, &p, err)
		return
	}

	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), p.Identifier)
	if err != nil {
		s.handleLoginError(w, r, ar, &p, errors.WithStack(schema.NewInvalidCredentialsError()))
		return
	}

	var o CredentialsConfig
	d := json.NewDecoder(bytes.NewBuffer(c.Config))
	if err := d.Decode(&o); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()))
		return
	}

	if err := s.d.Hasher().Compare([]byte(p.Password), []byte(o.HashedPassword)); err != nil {
		s.handleLoginError(w, r, ar, &p, errors.WithStack(schema.NewInvalidCredentialsError()))
		return
	}

	if err := s.d.LoginHookExecutor().PostLoginHook(w, r, identity.CredentialsTypePassword, ar, i); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, sr *login.Flow) error {
	// This block adds the identifier to the method when the request is forced - as a hint for the user.
	var identifier string
	if !sr.IsForced() {
		// do nothing
	} else if sess, err := s.d.SessionManager().FetchFromRequest(r.Context(), r); err != nil {
		// do nothing
	} else if id, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), sess.IdentityID); err != nil {
		// do nothing
	} else if creds, ok := id.GetCredentials(s.ID()); !ok {
		// do nothing
	} else if len(creds.Identifiers) == 0 {
		// do nothing
	} else {
		identifier = creds.Identifiers[0]
	}

	f := &form.HTMLForm{
		Action: sr.AppendTo(urlx.AppendPaths(s.c.SelfPublicURL(), RouteLogin)).String(),
		Method: "POST",
		Fields: form.Fields{{
			Name:     "identifier",
			Type:     "text",
			Value:    identifier,
			Required: true,
		}, {
			Name:     "password",
			Type:     "password",
			Required: true,
		}}}
	f.SetCSRF(s.d.GenerateCSRFToken(r))

	sr.Methods[identity.CredentialsTypePassword] = &login.FlowMethod{
		Method: identity.CredentialsTypePassword,
		Config: &login.FlowMethodConfig{FlowMethodConfigurator: &FlowMethod{HTMLForm: f}}}
	return nil
}
