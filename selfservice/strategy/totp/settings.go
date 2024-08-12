// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package totp

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/schema"

	"github.com/ory/kratos/text"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/session"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

const InternalContextKeyURL = "url"

func (s *Strategy) RegisterSettingsRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) SettingsStrategyID() string {
	return identity.CredentialsTypeTOTP.String()
}

// Update Settings Flow with TOTP Method
//
// swagger:model updateSettingsFlowWithTotpMethod
type updateSettingsFlowWithTotpMethod struct {
	// ValidationTOTP must contain a valid TOTP based on the
	ValidationTOTP string `json:"totp_code"`

	// UnlinkTOTP if true will remove the TOTP pairing,
	// effectively removing the credential. This can be used
	// to set up a new TOTP device.
	UnlinkTOTP bool `json:"totp_unlink"`

	// CSRFToken is the anti-CSRF token
	CSRFToken string `json:"csrf_token"`

	// Method
	//
	// Should be set to "totp" when trying to add, update, or remove a totp pairing.
	//
	// required: true
	Method string `json:"method"`

	// Flow is flow ID.
	//
	// swagger:ignore
	Flow string `json:"flow"`
}

func (p *updateSettingsFlowWithTotpMethod) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.Flow)
}

func (p *updateSettingsFlowWithTotpMethod) SetFlowID(rid uuid.UUID) {
	p.Flow = rid.String()
}

func (s *Strategy) Settings(w http.ResponseWriter, r *http.Request, f *settings.Flow, ss *session.Session) (*settings.UpdateContext, error) {
	var p updateSettingsFlowWithTotpMethod
	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, f, ss, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		return ctxUpdate, s.continueSettingsFlow(w, r, ctxUpdate, &p)
	} else if err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if err := s.decodeSettingsFlow(r, &p); err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if p.UnlinkTOTP {
		// This is a submit so we need to manually set the type to TOTP
		p.Method = s.SettingsStrategyID()
		if err := flow.MethodEnabledAndAllowed(r.Context(), f.GetFlowName(), s.SettingsStrategyID(), p.Method, s.d); err != nil {
			return nil, s.handleSettingsError(w, r, ctxUpdate, &p, err)
		}
	} else if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.SettingsStrategyID(), s.d); err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	// This does not come from the payload!
	p.Flow = ctxUpdate.Flow.ID.String()
	if err := s.continueSettingsFlow(w, r, ctxUpdate, &p); err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	return ctxUpdate, nil
}

func (s *Strategy) decodeSettingsFlow(r *http.Request, dest interface{}) error {
	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(settingsSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	return decoderx.NewHTTP().Decode(r, dest, compiler,
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	)
}

func (s *Strategy) continueSettingsFlow(
	w http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithTotpMethod,
) error {
	if err := flow.MethodEnabledAndAllowed(r.Context(), flow.SettingsFlow, s.SettingsStrategyID(), p.Method, s.d); err != nil {
		return err
	}

	if err := flow.EnsureCSRF(s.d, r, ctxUpdate.Flow.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return err
	}

	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context())).Before(time.Now()) {
		return errors.WithStack(settings.NewFlowNeedsReAuth())
	}

	hasTOTP, err := s.identityHasTOTP(r.Context(), ctxUpdate.Session.IdentityID)
	if err != nil {
		return err
	}

	// We have now two cases:
	//
	// 1. TOTP should be removed -> we have it already
	// 2. TOTP should be added -> we do not have it yet
	var i *identity.Identity
	if hasTOTP {
		i, err = s.continueSettingsFlowRemoveTOTP(w, r, ctxUpdate, p)
	} else {
		i, err = s.continueSettingsFlowAddTOTP(w, r, ctxUpdate, p)
	}

	if err != nil {
		return err
	}

	ctxUpdate.UpdateIdentity(i)
	return nil
}

func (s *Strategy) continueSettingsFlowAddTOTP(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithTotpMethod) (*identity.Identity, error) {
	keyURL := gjson.GetBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeyURL)).String()
	if len(keyURL) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Could not find they TOTP key in the internal context. This is a code bug and should be reported to https://github.com/ory/kratos/."))
	}

	key, err := otp.NewKeyFromURL(keyURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithTrace(err).WithReasonf("Could not decode TOTP key from the internal context. This is a code bug and should be reported to https://github.com/ory/kratos/."))
	}

	if len(key.Secret()) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("TOTP secret is not set. This is a code bug and should be reported to https://github.com/ory/kratos/."))
	}

	if p.ValidationTOTP == "" {
		return nil, schema.NewRequiredError("#/totp_code", "totp_code")
	}

	if !totp.Validate(p.ValidationTOTP, key.Secret()) {
		return nil, schema.NewTOTPVerifierWrongError("#/totp_code")
	}

	co, err := json.Marshal(&identity.CredentialsTOTPConfig{TOTPURL: key.URL()})
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode totp options to JSON: %s", err))
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		return nil, err
	}

	// We do not really need the identifier, so we add the identity's ID
	c := &identity.Credentials{Type: s.ID(), Identifiers: []string{i.ID.String()}, Config: co}
	c.Config = co
	i.SetCredentials(s.ID(), *c)

	// Remove the TOTP URL from the internal context now that it is set!
	ctxUpdate.Flow.InternalContext, err = sjson.DeleteBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeyURL))
	if err != nil {
		return nil, err
	}

	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), ctxUpdate.Flow); err != nil {
		return nil, err
	}

	// Since we added the method, it also means that we have authenticated it
	if err := s.d.SessionManager().SessionAddAuthenticationMethods(r.Context(), ctxUpdate.Session.ID, session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    identity.AuthenticatorAssuranceLevel2,
	}); err != nil {
		return nil, err
	}

	return i, nil
}

func (s *Strategy) continueSettingsFlowRemoveTOTP(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithTotpMethod) (*identity.Identity, error) {
	if !p.UnlinkTOTP {
		return ctxUpdate.Session.Identity, nil
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		return nil, err
	}

	i.DeleteCredentialsType(identity.CredentialsTypeTOTP)
	return i, nil
}

func (s *Strategy) identityHasTOTP(ctx context.Context, id uuid.UUID) (bool, error) {
	confidential, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(ctx, id)
	if err != nil {
		return false, err
	}

	count, err := s.CountActiveMultiFactorCredentials(confidential.Credentials)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Strategy) PopulateSettingsMethod(r *http.Request, id *identity.Identity, f *settings.Flow) error {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))

	hasTOTP, err := s.identityHasTOTP(r.Context(), id.ID)
	if err != nil {
		return err
	}

	// OTP already set up, just add an unlink option
	if hasTOTP {
		f.UI.Nodes.Upsert(NewUnlinkTOTPNode())
	} else {
		e := NewSchemaExtension(id.ID.String())
		_ = s.d.IdentityValidator().ValidateWithRunner(r.Context(), id, e)

		// No TOTP set up yet, add nodes allowing us to add it.
		key, err := NewKey(r.Context(), e.AccountName, s.d)
		if err != nil {
			return err
		}

		f.InternalContext, err = sjson.SetBytes(f.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeyURL), key.URL())
		if err != nil {
			return err
		}

		qr, err := NewTOTPImageQRNode(key)
		if err != nil {
			return err
		}

		f.UI.Nodes.Upsert(NewTOTPSourceURLNode(key))
		f.UI.Nodes.Upsert(qr)
		f.UI.Nodes.Upsert(NewVerifyTOTPNode())
		f.UI.Nodes.Append(node.NewInputField("method", "totp", node.TOTPGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSave()))
	}

	return nil
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithTotpMethod, err error) error {
	// Do not pause flow if the flow type is an API flow as we can't save cookies in those flows.
	if e := new(settings.FlowNeedsReAuth); errors.As(err, &e) && ctxUpdate.Flow != nil && ctxUpdate.Flow.Type == flow.TypeBrowser {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r, settings.ContinuityKey(s.SettingsStrategyID()), settings.ContinuityOptions(p, ctxUpdate.GetSessionIdentity())...); err != nil {
			return err
		}
	}

	if ctxUpdate.Flow != nil {
		ctxUpdate.Flow.UI.ResetMessages()
		ctxUpdate.Flow.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	}

	return err
}
