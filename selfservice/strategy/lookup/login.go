// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package lookup

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/x/sqlxx"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *login.Flow) error {
	// This strategy can only solve AAL2
	if requestedAAL != identity.AuthenticatorAssuranceLevel2 {
		return nil
	}

	// We have done proper validation before so this should never error
	sess, err := s.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		return err
	}

	id, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), sess.IdentityID)
	if err != nil {
		return err
	}

	_, ok := id.GetCredentials(s.ID())
	if !ok {
		// Identity has no lookup codes
		return nil
	}

	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	sr.UI.SetNode(node.NewInputField(node.LookupCodeEnter, "", node.LookupGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoLoginLookupLabel()))
	sr.UI.GetNodes().Append(node.NewInputField("method", s.ID(), node.LookupGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoLoginLookup()))

	return nil
}

func (s *Strategy) handleLoginError(r *http.Request, f *login.Flow, err error) error {
	if f != nil {
		f.UI.Nodes.ResetNodes(node.LookupCodeEnter)
		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
		}
	}

	return err
}

// Update Login Flow with Lookup Secret Method
//
// swagger:model updateLoginFlowWithLookupSecretMethod
type updateLoginFlowWithLookupSecretMethod struct {
	// Method should be set to "lookup_secret" when logging in using the lookup_secret strategy.
	//
	// required: true
	Method string `json:"method"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `json:"csrf_token"`

	// The lookup secret.
	//
	// required: true
	Code string `json:"lookup_secret"`
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID) (i *identity.Identity, err error) {
	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel2); err != nil {
		return nil, err
	}

	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.ID().String(), s.d); err != nil {
		return nil, err
	}

	var p updateLoginFlowWithLookupSecretMethod
	if err := s.hd.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginSchema),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	if err := flow.EnsureCSRF(s.d, r, f.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), identityID.String())
	if errors.Is(err, sqlcon.ErrNoRows) {
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewNoLookupDefined()))
	} else if err != nil {
		return nil, s.handleLoginError(r, f, err)
	}

	var o identity.CredentialsLookupConfig
	if err := json.Unmarshal(c.Config, &o); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("The lookup secrets could not be decoded properly").WithDebug(err.Error()).WithWrap(err))
	}

	var found bool
	for k, rc := range o.RecoveryCodes {
		if rc.Code == p.Code {
			if time.Time(rc.UsedAt).IsZero() {
				o.RecoveryCodes[k].UsedAt = sqlxx.NullTime(time.Now().UTC().Round(time.Second))
				found = true
			} else {
				return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewLookupAlreadyUsed()))
			}
		}
	}

	if !found {
		return nil, s.handleLoginError(r, f, errors.WithStack(schema.NewErrorValidationLookupInvalid()))
	}

	toUpdate, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), identityID)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(&o)
	if err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to encoded updated lookup secrets.").WithDebug(err.Error())))
	}

	c.Config = encoded
	toUpdate.SetCredentials(s.ID(), *c)

	if err := s.d.PrivilegedIdentityPool().UpdateIdentity(r.Context(), toUpdate); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to update identity.").WithDebug(err.Error())))
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		return nil, s.handleLoginError(r, f, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow.").WithDebug(err.Error())))
	}

	return i, nil
}
