// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ory/kratos/selfservice/flowhelpers"

	"github.com/ory/x/stringsx"

	"github.com/gofrs/uuid"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
}

func (s *Strategy) handleLoginError(w http.ResponseWriter, r *http.Request, f *login.Flow, payload *updateLoginFlowWithPasswordMethod, err error) error {
	if f != nil {
		f.UI.Nodes.ResetNodes("password")
		f.UI.Nodes.SetValueAttribute("identifier", stringsx.Coalesce(payload.Identifier, payload.LegacyIdentifier))
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID) (i *identity.Identity, err error) {
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, err
	}

	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.ID().String(), s.d); err != nil {
		return nil, err
	}

	var p updateLoginFlowWithPasswordMethod
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, err)
	}

	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), stringsx.Coalesce(p.Identifier, p.LegacyIdentifier))
	if err != nil {
		time.Sleep(x.RandomDelay(s.d.Config().HasherArgon2(r.Context()).ExpectedDuration, s.d.Config().HasherArgon2(r.Context()).ExpectedDeviation))
		return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(schema.NewInvalidCredentialsError()))
	}

	var o identity.CredentialsPassword
	d := json.NewDecoder(bytes.NewBuffer(c.Config))
	if err := d.Decode(&o); err != nil {
		return nil, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()).WithWrap(err)
	}

	if err := hash.Compare(r.Context(), []byte(p.Password), []byte(o.HashedPassword)); err != nil {
		return nil, s.handleLoginError(w, r, f, &p, errors.WithStack(schema.NewInvalidCredentialsError()))
	}

	if !s.d.Hasher(r.Context()).Understands([]byte(o.HashedPassword)) {
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
	hpw, err := s.d.Hasher(ctx).Generate(ctx, password)
	if err != nil {
		return err
	}
	co, err := json.Marshal(&identity.CredentialsPassword{HashedPassword: string(hpw)})
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

	if sr.IsForced() {
		// We only show this method on a refresh request if the user has indeed a password set.
		identifier, id, _ := flowhelpers.GuessForcedLoginIdentifier(r, s.d, sr, s.ID())
		if identifier == "" {
			return nil
		}

		count, err := s.CountActiveFirstFactorCredentials(id.Credentials)
		if err != nil {
			return err
		} else if count == 0 {
			return nil
		}

		sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		sr.UI.SetNode(node.NewInputField("identifier", identifier, node.DefaultGroup, node.InputAttributeTypeHidden))
	} else {
		sr.UI.SetNode(node.NewInputField("identifier", "", node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeLabelID()))
	}

	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	sr.UI.SetNode(NewPasswordNode("password", node.InputAttributeAutocompleteCurrentPassword))
	sr.UI.GetNodes().Append(node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoLogin()))

	return nil
}
