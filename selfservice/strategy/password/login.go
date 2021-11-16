package password

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	errors2 "github.com/ory/kratos/schema/errors"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/session"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
}

func (s *Strategy) handleLoginError(w http.ResponseWriter, r *http.Request, f *login.Flow, payload *submitSelfServiceLoginFlowWithPasswordMethodBody, err error) error {
	if f != nil {
		f.UI.Nodes.ResetNodes("password")
		f.UI.Nodes.SetValueAttribute("password_identifier", payload.Identifier)
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, ss *session.Session) (i *identity.Identity, err error) {
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, err
	}

	if err := flow.MethodEnabledAndAllowedFromRequest(r, s.ID().String(), s.d); err != nil {
		return nil, err
	}

	var p submitSelfServiceLoginFlowWithPasswordMethodBody
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}

	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), p.Identifier)
	if err != nil {
		time.Sleep(x.RandomDelay(s.d.Config(r.Context()).HasherArgon2().ExpectedDuration, s.d.Config(r.Context()).HasherArgon2().ExpectedDeviation))
		return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(errors2.NewInvalidCredentialsError()))
	}

	var o identity.CredentialsConfig
	d := json.NewDecoder(bytes.NewBuffer(c.Config))
	if err := d.Decode(&o); err != nil {
		return nil, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)
	}

	if err := hash.Compare(r.Context(), []byte(p.Password), []byte(o.HashedPassword)); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(errors2.NewInvalidCredentialsError()))
	}

	if !s.d.Hasher().Understands([]byte(o.HashedPassword)) {
		// Migrate password hash if not configured hasher.
		// But Kratos doesn't have ability to import credentials now.
		// see https://github.com/ory/kratos/issues/605
		if err := s.migratePasswordHash(r.Context(), i.ID, []byte(p.Password)); err != nil {
			return nil, s.handleLoginError(w, r, f, &p, err)
		}
	}

	f.Active = identity.CredentialsTypePassword
	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithDebug(err.Error())))
	}

	return i, nil
}

func (s *Strategy) migratePasswordHash(ctx context.Context, identifier uuid.UUID, password []byte) error {
	hpw, err := s.d.Hasher().Generate(ctx, password)
	if err != nil {
		return err
	}
	co, err := json.Marshal(&identity.CredentialsConfig{HashedPassword: string(hpw)})
	if err != nil {
		return errors.Wrap(err, "unable to encode password configuration to JSON")
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(ctx, identifier)
	if err != nil {
		return err
	}

	c, ok := i.GetCredentials(s.ID())
	if !ok {
		return errors.New("expected to find password credential but could not")
	}

	c.Config = co
	i.SetCredentials(s.ID(), *c)

	return s.d.PrivilegedIdentityPool().UpdateIdentity(ctx, i)
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *login.Flow) error {
	// This strategy can only solve AAL1
	if requestedAAL > identity.AuthenticatorAssuranceLevel1 {
		return nil
	}

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
	sr.UI.SetNode(node.NewInputField("password_identifier", identifier, node.PasswordGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeLabelID()))
	sr.UI.SetNode(NewPasswordNode("password"))
	sr.UI.GetNodes().Append(node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoLogin()))

	return nil
}
