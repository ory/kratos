package password

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
}

func (s *Strategy) handleLoginError(w http.ResponseWriter, r *http.Request, f *login.Flow, payload *CompleteSelfServiceLoginFlowWithPasswordMethod, err error) error {
	if f != nil {
		f.UI.Nodes.Reset()
		f.UI.Nodes.SetValueAttribute(node.ToID(s.NodeGroup(), "password.identifier"), payload.Password.Identifier)
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

// nolint:deadcode,unused
// swagger:parameters completeSelfServiceLoginFlowWithPasswordMethod
type completeSelfServiceLoginFlowWithPasswordMethodParameters struct {
	// The Flow ID
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// in: body
	Body CompleteSelfServiceLoginFlowWithPasswordMethod
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow) (i *identity.Identity, err error) {
	if err := flow.MethodEnabledAndAllowed(r, s.ID().String(), s.d); err != nil {
		return nil, err
	}

	var p CompleteSelfServiceLoginFlowWithPasswordMethod
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}

	if err := flow.EnsureCSRF(r, f.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}

	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), p.Password.Identifier)
	if err != nil {
		time.Sleep(x.RandomDelay(s.d.Config(r.Context()).HasherArgon2().ExpectedDuration, s.d.Config(r.Context()).HasherArgon2().ExpectedDeviation))
		return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(schema.NewInvalidCredentialsError()))
	}

	var o CredentialsConfig
	d := json.NewDecoder(bytes.NewBuffer(c.Config))
	if err := d.Decode(&o); err != nil {
		return nil, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)
	}

	if err := hash.Compare(r.Context(), []byte(p.Password.Password), []byte(o.HashedPassword)); err != nil {
		s.handleLoginError(w, r, ar, &p, errors.WithStack(schema.NewInvalidCredentialsError()))
		return
	}

	return i, nil
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

	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	sr.UI.SetNode(node.NewInputField("password.identifier", identifier, node.PasswordGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute))
	sr.UI.SetNode(node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit))
	sr.UI.SetNode(NewPasswordNode("password.password"))

	return nil
}
