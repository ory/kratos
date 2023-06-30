package code

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/stringsx"
)

var _ login.Strategy = new(Strategy)

type loginSubmitPayload struct {
	Method     string `json:"method"`
	CSRFToken  string `json:"csrf_token"`
	Code       string `json:"code"`
	Identifier string `json:"identifier"`
}

func (s *Strategy) RegisterLoginRoutes(*x.RouterPublic) {
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeCodeAuth
}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: identity.CredentialsTypeCodeAuth,
		AAL:    identity.AuthenticatorAssuranceLevel2,
	}
}

func (s *Strategy) HandleLoginError(w http.ResponseWriter, r *http.Request, flow *login.Flow, body *loginSubmitPayload, err error) error {
	if flow != nil {
		email := ""
		if body != nil {
			email = body.Identifier
		}

		flow.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		flow.UI.GetNodes().Upsert(
			node.NewInputField("identifier", email, node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeInputEmail()),
		)
	}

	return err
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, lf *login.Flow) error {
	if lf.Type != flow.TypeBrowser {
		return nil
	}

	if requestedAAL == identity.AuthenticatorAssuranceLevel2 {
		return nil
	}

	lf.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
	lf.UI.GetNodes().Upsert(
		node.NewInputField("identifier", "", node.CodeGroup, node.InputAttributeTypeEmail, node.WithRequiredInputAttribute).
			WithMetaLabel(text.NewInfoNodeInputEmail()),
	)
	return nil
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID) (i *identity.Identity, err error) {
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, err
	}

	if err := flow.MethodEnabledAndAllowedFromRequest(r, s.ID().String(), s.deps); err != nil {
		return nil, err
	}

	var p loginSubmitPayload
	if err := s.dx.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginMethodSchema),
		decoderx.HTTPDecoderAllowedMethods("POST"),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.HandleLoginError(w, r, f, &p, err)
	}

	if err := flow.EnsureCSRF(s.deps, r, f.Type, s.deps.Config().DisableAPIFlowEnforcement(r.Context()), s.deps.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.HandleLoginError(w, r, f, &p, err)
	}

	i, c, err := s.deps.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), p.Identifier)

	if err != nil {
		return nil, s.HandleLoginError(w, r, f, &p, err)
	}

	var o identity.CredentialsOTP
	d := json.NewDecoder(bytes.NewBuffer(c.Config))
	if err := d.Decode(&o); err != nil {
		return nil, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)
	}

	f.Active = identity.CredentialsTypeCodeAuth
	if err = s.deps.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		return nil, s.HandleLoginError(w, r, f, &p, err)
	}

	return i, nil
}
