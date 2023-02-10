package pin

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ory/herodot"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/pkg/errors"
	"net/http"
)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *login.Flow) error {

	sess, err := s.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		return nil
	}

	id, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), sess.IdentityID)
	if err != nil {
		return err
	}

	_, ok := id.GetCredentials(s.ID())
	if !ok {
		// Identity has no TOTP
		return nil
	}

	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	sr.UI.SetNode(node.NewInputField("pin", "", node.PinGroup, node.InputAttributeTypePassword, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoLoginTOTPLabel()))

	return nil
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, ss *session.Session) (i *identity.Identity, err error) {

	if err := flow.MethodEnabledAndAllowedFromRequest(r, s.ID().String(), s.d); err != nil {
		return nil, err
	}

	if !ss.Active {
		return nil, s.handleLoginError(w, r, f, nil,
			errors.WithStack(session.NewErrNoActiveSessionFound().WithReason("This endpoint can only be accessed with a valid session. Please log in and try again.")))
	}

	var p submitSelfServiceLoginFlowWithPinMethodBody
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}

	i, err = s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ss.IdentityID)
	if err != nil {
		return nil, err
	}
	c, found := i.GetCredentials(s.ID())
	if !found {
		return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(schema.NewInvalidPinError()))
	}
	var o identity.CredentialsPin
	d := json.NewDecoder(bytes.NewBuffer(c.Config))
	if err := d.Decode(&o); err != nil {
		return nil, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)
	}

	if err := hash.Compare(r.Context(), []byte(p.Pin), []byte(o.HashedPin)); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(schema.NewInvalidPinError()))
	}

	f.Active = identity.CredentialsTypePin
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithDebug(err.Error())))
	}

	return i, nil
}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    identity.NoAuthenticatorAssuranceLevel,
	}
}

func (s *Strategy) handleLoginError(w http.ResponseWriter, r *http.Request, f *login.Flow, payload *submitSelfServiceLoginFlowWithPinMethodBody, err error) error {
	if f != nil {
		f.UI.Nodes.ResetNodes("pin")
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}
