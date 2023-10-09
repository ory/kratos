// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package lookup

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/x/randx"

	"github.com/ory/herodot"
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

func (s *Strategy) RegisterSettingsRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) SettingsStrategyID() string {
	return identity.CredentialsTypeLookup.String()
}

const (
	internalContextKeyRevealed    = "revealed"
	InternalContextKeyRegenerated = "regenerated"
)

const numCodes = 12

var allSettingsNodes = []string{
	node.LookupRegenerate,
	node.LookupReveal,
	node.LookupRegenerate,
	node.LookupDisable,
	node.LookupCodes,
	node.LookupConfirm,
}

// Update Settings Flow with Lookup Method
//
// swagger:model updateSettingsFlowWithLookupMethod
type updateSettingsFlowWithLookupMethod struct {
	// If set to true will reveal the lookup secrets
	RevealLookup bool `json:"lookup_secret_reveal"`

	// If set to true will regenerate the lookup secrets
	RegenerateLookup bool `json:"lookup_secret_regenerate"`

	// If set to true will save the regenerated lookup secrets
	ConfirmLookup bool `json:"lookup_secret_confirm"`

	// Disables this method if true.
	DisableLookup bool `json:"lookup_secret_disable"`

	// CSRFToken is the anti-CSRF token
	CSRFToken string `json:"csrf_token"`

	// Method
	//
	// Should be set to "lookup" when trying to add, update, or remove a lookup pairing.
	//
	// required: true
	Method string `json:"method"`

	// Flow is flow ID.
	//
	// swagger:ignore
	Flow string `json:"flow"`
}

func (p *updateSettingsFlowWithLookupMethod) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.Flow)
}

func (p *updateSettingsFlowWithLookupMethod) SetFlowID(rid uuid.UUID) {
	p.Flow = rid.String()
}

func (s *Strategy) Settings(w http.ResponseWriter, r *http.Request, f *settings.Flow, ss *session.Session) (*settings.UpdateContext, error) {
	var p updateSettingsFlowWithLookupMethod
	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, f, ss, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		return ctxUpdate, s.continueSettingsFlow(w, r, ctxUpdate, &p)
	} else if err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if err := s.decodeSettingsFlow(r, &p); err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if p.RegenerateLookup || p.RevealLookup || p.ConfirmLookup || p.DisableLookup {
		// This method has only two submit buttons
		p.Method = s.SettingsStrategyID()
		if err := flow.MethodEnabledAndAllowed(r.Context(), f.GetFlowName(), s.SettingsStrategyID(), p.Method, s.d); err != nil {
			return nil, s.handleSettingsError(w, r, ctxUpdate, &p, err)
		}
	} else {
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
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
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	)
}

func (s *Strategy) continueSettingsFlow(
	w http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithLookupMethod,
) error {
	if p.ConfirmLookup || p.RevealLookup || p.RegenerateLookup || p.DisableLookup {
		if err := flow.MethodEnabledAndAllowed(r.Context(), flow.SettingsFlow, s.SettingsStrategyID(), s.SettingsStrategyID(), s.d); err != nil {
			return err
		}

		if err := flow.EnsureCSRF(s.d, r, ctxUpdate.Flow.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
			return err
		}

		if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context())).Before(time.Now()) {
			return errors.WithStack(settings.NewFlowNeedsReAuth())
		}
	} else {
		return errors.New("ended up in unexpected state")
	}

	if p.ConfirmLookup {
		return s.continueSettingsFlowConfirm(w, r, ctxUpdate, p)
	} else if p.RevealLookup {
		if err := s.continueSettingsFlowReveal(w, r, ctxUpdate, p); err != nil {
			return err
		}
		return flow.ErrStrategyAsksToReturnToUI
	} else if p.DisableLookup {
		return s.continueSettingsFlowDisable(w, r, ctxUpdate, p)
	} else if p.RegenerateLookup {
		if err := s.continueSettingsFlowRegenerate(w, r, ctxUpdate, p); err != nil {
			return err
		}
		// regen
		return flow.ErrStrategyAsksToReturnToUI
	}

	return errors.New("ended up in unexpected state")
}

func (s *Strategy) continueSettingsFlowDisable(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithLookupMethod) error {
	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		return err
	}

	i.DeleteCredentialsType(s.ID())

	for _, n := range allSettingsNodes {
		ctxUpdate.Flow.UI.Nodes.Remove(n)
	}

	ctxUpdate.Flow.UI.Nodes.Upsert(NewRegenerateLookupNode())
	ctxUpdate.Flow.InternalContext, err = sjson.SetBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), internalContextKeyRevealed), false)
	if err != nil {
		return err
	}
	ctxUpdate.Flow.InternalContext, err = sjson.SetRawBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeyRegenerated), []byte("{}"))
	if err != nil {
		return err
	}

	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), ctxUpdate.Flow); err != nil {
		return err
	}

	ctxUpdate.UpdateIdentity(i)
	return nil
}

func (s *Strategy) continueSettingsFlowReveal(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithLookupMethod) error {
	hasLookup, err := s.identityHasLookup(r.Context(), ctxUpdate.Session.IdentityID)
	if err != nil {
		return err
	}

	if !hasLookup {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("Can not reveal lookup codes because you have none."))
	}

	_, cred, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), ctxUpdate.Session.IdentityID.String())
	if err != nil {
		return err
	}

	var creds identity.CredentialsLookupConfig
	if err := json.Unmarshal(cred.Config, &creds); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to decode lookup codes from JSON.").WithDebug(err.Error()))
	}

	for _, n := range allSettingsNodes {
		ctxUpdate.Flow.UI.Nodes.Remove(n)
	}

	ctxUpdate.Flow.UI.Nodes.Upsert(creds.ToNode())
	ctxUpdate.Flow.UI.Nodes.Upsert(NewDisableLookupNode())
	ctxUpdate.Flow.UI.Nodes.Upsert(NewRegenerateLookupNode())

	ctxUpdate.Flow.InternalContext, err = sjson.SetBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), internalContextKeyRevealed), true)
	if err != nil {
		return err
	}

	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), ctxUpdate.Flow); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) continueSettingsFlowRegenerate(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithLookupMethod) error {
	codes := make([]identity.RecoveryCode, numCodes)
	for k := range codes {
		codes[k] = identity.RecoveryCode{Code: randx.MustString(8, randx.AlphaLowerNum)}
	}

	for _, n := range allSettingsNodes {
		ctxUpdate.Flow.UI.Nodes.Remove(n)
	}

	ctxUpdate.Flow.UI.Nodes.Upsert((&identity.CredentialsLookupConfig{RecoveryCodes: codes}).ToNode())
	ctxUpdate.Flow.UI.Nodes.Upsert(NewConfirmLookupNode())

	var err error
	ctxUpdate.Flow.InternalContext, err = sjson.SetBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeyRegenerated), codes)
	if err != nil {
		return err
	}

	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), ctxUpdate.Flow); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) continueSettingsFlowConfirm(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithLookupMethod) error {
	codes := gjson.GetBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeyRegenerated)).Array()
	if len(codes) != numCodes {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("You must (re-)generate recovery backup codes before you can save them."))
	}

	rc := make([]identity.RecoveryCode, len(codes))
	for k := range rc {
		rc[k] = identity.RecoveryCode{Code: codes[k].Get("code").String()}
	}

	co, err := json.Marshal(&identity.CredentialsLookupConfig{RecoveryCodes: rc})
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode totp options to JSON: %s", err))
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		return err
	}

	// We do not really need the identifier, so we add the identity's ID
	c := &identity.Credentials{Type: s.ID(), Identifiers: []string{i.ID.String()}, Config: co}
	c.Config = co
	i.SetCredentials(s.ID(), *c)

	// Remove the TOTP URL from the internal context now that it is set!
	ctxUpdate.Flow.InternalContext, err = sjson.DeleteBytes(ctxUpdate.Flow.InternalContext, flow.PrefixInternalContextKey(s.ID(), InternalContextKeyRegenerated))
	if err != nil {
		return err
	}

	if err := s.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), ctxUpdate.Flow); err != nil {
		return err
	}

	// Since we added the method, it also means that we have authenticated it
	if err := s.d.SessionManager().SessionAddAuthenticationMethods(r.Context(), ctxUpdate.Session.ID, session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    identity.AuthenticatorAssuranceLevel2,
	}); err != nil {
		return err
	}

	ctxUpdate.UpdateIdentity(i)
	return nil
}

func (s *Strategy) identityHasLookup(ctx context.Context, id uuid.UUID) (bool, error) {
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

	hasLookup, err := s.identityHasLookup(r.Context(), id.ID)
	if err != nil {
		return err
	}

	if hasLookup {
		f.UI.Nodes.Upsert(NewRevealLookupNode())
		f.UI.Nodes.Upsert(NewDisableLookupNode())
	} else {
		f.UI.Nodes.Upsert(NewRegenerateLookupNode())
	}

	return nil
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithLookupMethod, err error) error {
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
