// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"time"

	"github.com/tidwall/sjson"

	"golang.org/x/oauth2"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/selfservice/strategy"
	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/session"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"

	"github.com/ory/kratos/x"
)

//go:embed .schema/settings.schema.json
var settingsSchema []byte

var _ settings.Strategy = new(Strategy)
var UnknownConnectionValidationError = &jsonschema.ValidationError{
	Message: "can not unlink non-existing OpenID Connect connection", InstancePtr: "#/"}
var ConnectionExistValidationError = &jsonschema.ValidationError{
	Message: "can not link unknown or already existing OpenID Connect connection", InstancePtr: "#/"}
var UnlinkAllFirstFactorConnectionsError = &jsonschema.ValidationError{
	Message: "can not unlink OpenID Connect connection because it is the last remaining first factor credential", InstancePtr: "#/"}

func (s *Strategy) RegisterSettingsRoutes(router *x.RouterPublic) {}

func (s *Strategy) SettingsStrategyID() string {
	return s.ID().String()
}

func (s *Strategy) decoderSettings(p *updateSettingsFlowWithOidcMethod, r *http.Request) error {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}
	raw, err := sjson.SetBytes(settingsSchema,
		"properties.traits.$ref", ds.String()+"#/properties/traits")
	if err != nil {
		return errors.WithStack(err)
	}

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.dec.Decode(r, &p, compiler,
		decoderx.HTTPKeepRequestBody(true),
		decoderx.HTTPDecoderUseQueryAndBody(),
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Strategy) linkedProviders(ctx context.Context, r *http.Request, conf *ConfigurationCollection, confidential *identity.Identity) ([]Provider, error) {
	creds, ok := confidential.GetCredentials(s.ID())
	if !ok {
		return nil, nil
	}

	var available identity.CredentialsOIDC
	if err := json.Unmarshal(creds.Config, &available); err != nil {
		return nil, errors.WithStack(err)
	}

	var result []Provider
	for _, p := range available.Providers {
		prov, err := conf.Provider(p.Provider, s.d)
		if errors.Is(err, herodot.ErrNotFound) {
			continue
		} else if err != nil {
			return nil, err
		}
		result = append(result, prov)
	}

	return result, nil
}

func (s *Strategy) linkableProviders(ctx context.Context, r *http.Request, conf *ConfigurationCollection, confidential *identity.Identity) ([]Provider, error) {
	var available identity.CredentialsOIDC
	creds, ok := confidential.GetCredentials(s.ID())
	if ok {
		if err := json.Unmarshal(creds.Config, &available); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	var result []Provider
	for _, p := range conf.Providers {
		var found bool
		for _, pp := range available.Providers {
			if pp.Provider == p.ID {
				found = true
				break
			}
		}

		if !found {
			prov, err := conf.Provider(p.ID, s.d)
			if err != nil {
				return nil, err
			}
			result = append(result, prov)
		}
	}

	return result, nil
}

func (s *Strategy) PopulateSettingsMethod(r *http.Request, id *identity.Identity, sr *settings.Flow) error {
	if sr.Type != flow.TypeBrowser {
		return nil
	}

	conf, err := s.Config(r.Context())
	if err != nil {
		return err
	}

	confidential, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), id.ID)
	if err != nil {
		return err
	}

	linkable, err := s.linkableProviders(r.Context(), r, conf, confidential)
	if err != nil {
		return err
	}

	linked, err := s.linkedProviders(r.Context(), r, conf, confidential)
	if err != nil {
		return err
	}

	sr.UI.GetNodes().Remove("unlink", "link")
	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	for _, l := range linkable {
		sr.UI.GetNodes().Append(NewLinkNode(l.Config().ID))
	}

	count, err := s.d.IdentityManager().CountActiveFirstFactorCredentials(r.Context(), confidential)
	if err != nil {
		return err
	}

	if count > 1 {
		// This means that we're able to remove a connection because it is the last configured credential. If it is
		// removed, the identity is no longer able to sign in.
		for _, l := range linked {
			sr.UI.GetNodes().Append(NewUnlinkNode(l.Config().ID))
		}
	}

	return nil
}

// Update Settings Flow with OpenID Connect Method
//
// swagger:model updateSettingsFlowWithOidcMethod
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateSettingsFlowWithOidcMethod struct {
	// Method
	//
	// Should be set to profile when trying to update a profile.
	//
	// required: true
	Method string `json:"method"`

	// Link this provider
	//
	// Either this or `unlink` must be set.
	//
	// type: string
	// in: body
	Link string `json:"link"`

	// Unlink this provider
	//
	// Either this or `link` must be set.
	//
	// type: string
	// in: body
	Unlink string `json:"unlink"`

	// Flow ID is the flow's ID.
	//
	// in: query
	FlowID string `json:"flow"`

	// The identity's traits
	//
	// in: body
	Traits json.RawMessage `json:"traits"`

	// UpstreamParameters are the parameters that are passed to the upstream identity provider.
	//
	// These parameters are optional and depend on what the upstream identity provider supports.
	// Supported parameters are:
	// - `login_hint` (string): The `login_hint` parameter suppresses the account chooser and either pre-fills the email box on the sign-in form, or selects the proper session.
	// - `hd` (string): The `hd` parameter limits the login/registration process to a Google Organization, e.g. `mycollege.edu`.
	// - `prompt` (string): The `prompt` specifies whether the Authorization Server prompts the End-User for reauthentication and consent, e.g. `select_account`.
	//
	// required: false
	UpstreamParameters json.RawMessage `json:"upstream_parameters"`
}

func (p *updateSettingsFlowWithOidcMethod) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.FlowID)
}

func (p *updateSettingsFlowWithOidcMethod) SetFlowID(rid uuid.UUID) {
	p.FlowID = rid.String()
}

func (s *Strategy) Settings(w http.ResponseWriter, r *http.Request, f *settings.Flow, ss *session.Session) (*settings.UpdateContext, error) {
	var p updateSettingsFlowWithOidcMethod
	if err := s.decoderSettings(&p, r); err != nil {
		return nil, err
	}

	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, f, ss, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		if !s.d.Config().SelfServiceStrategy(r.Context(), s.SettingsStrategyID()).Enabled {
			return nil, errors.WithStack(herodot.ErrNotFound.WithReason(strategy.EndpointDisabledMessage))
		}

		if l := len(p.Link); l > 0 {
			if err := s.initLinkProvider(w, r, ctxUpdate, &p); err != nil {
				return nil, err
			}

			return ctxUpdate, nil
		} else if u := len(p.Unlink); u > 0 {
			if err := s.unlinkProvider(w, r, ctxUpdate, &p); err != nil {
				return nil, err
			}

			return ctxUpdate, nil
		}

		return nil, s.handleSettingsError(w, r, ctxUpdate, &p, errors.WithStack(herodot.ErrInternalServerError.WithReason("Expected either link or unlink to be set when continuing flow but both are unset.")))
	} else if err != nil {
		return nil, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if len(p.Link+p.Unlink) == 0 {
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if !s.d.Config().SelfServiceStrategy(r.Context(), s.SettingsStrategyID()).Enabled {
		return nil, errors.WithStack(herodot.ErrNotFound.WithReason(strategy.EndpointDisabledMessage))
	}

	if l, u := len(p.Link), len(p.Unlink); l > 0 && u > 0 {
		return nil, s.handleSettingsError(w, r, ctxUpdate, &p, errors.WithStack(&jsonschema.ValidationError{
			Message:     "it is not possible to link and unlink providers in the same request",
			InstancePtr: "#/",
		}))
	} else if l > 0 {
		if err := s.initLinkProvider(w, r, ctxUpdate, &p); err != nil {
			return nil, err
		}
		return ctxUpdate, nil
	} else if u > 0 {
		if err := s.unlinkProvider(w, r, ctxUpdate, &p); err != nil {
			return nil, err
		}

		return ctxUpdate, nil
	}

	return nil, s.handleSettingsError(w, r, ctxUpdate, &p, errors.WithStack(errors.WithStack(&jsonschema.ValidationError{
		Message: "missing properties: link, unlink", InstancePtr: "#/",
		Context: &jsonschema.ValidationErrorContextRequired{Missing: []string{"link", "unlink"}}})))
}

func (s *Strategy) isLinkable(r *http.Request, ctxUpdate *settings.UpdateContext, toLink string) (*identity.Identity, error) {
	providers, err := s.Config(r.Context())
	if err != nil {
		return nil, err
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		return nil, err
	}

	linkable, err := s.linkableProviders(r.Context(), r, providers, i)
	if err != nil {
		return nil, err
	}

	var found bool
	for _, available := range linkable {
		if toLink == available.Config().ID {
			found = true
		}
	}

	if !found {
		return nil, errors.WithStack(ConnectionExistValidationError)
	}

	return i, nil
}

func (s *Strategy) initLinkProvider(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithOidcMethod) error {
	if _, err := s.isLinkable(r, ctxUpdate, p.Link); err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context())).Before(time.Now()) {
		return s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
	}

	provider, err := s.provider(r.Context(), r, p.Link)
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	c, err := provider.OAuth2(r.Context())
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	req, err := s.validateFlow(r.Context(), r, ctxUpdate.Flow.ID)
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	state := generateState(ctxUpdate.Flow.ID.String()).String()
	if err := s.d.ContinuityManager().Pause(r.Context(), w, r, sessionName,
		continuity.WithPayload(&authCodeContainer{
			State:  state,
			FlowID: ctxUpdate.Flow.ID.String(),
			Traits: p.Traits,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	var up map[string]string
	if err := json.NewDecoder(bytes.NewBuffer(p.UpstreamParameters)).Decode(&up); err != nil {
		return err
	}

	codeURL := c.AuthCodeURL(state, append(UpstreamParameters(provider, up), provider.AuthCodeURLOptions(req)...)...)
	if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(codeURL))
	} else {
		http.Redirect(w, r, codeURL, http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) linkProvider(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, token *oauth2.Token, claims *Claims, provider Provider) error {
	p := &updateSettingsFlowWithOidcMethod{
		Link: provider.Config().ID, FlowID: ctxUpdate.Flow.ID.String()}
	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context())).Before(time.Now()) {
		return s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
	}

	i, err := s.isLinkable(r, ctxUpdate, p.Link)
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	var it string
	if idToken, ok := token.Extra("id_token").(string); ok {
		if it, err = s.d.Cipher(r.Context()).Encrypt(r.Context(), []byte(idToken)); err != nil {
			return s.handleSettingsError(w, r, ctxUpdate, p, err)
		}
	}

	cat, err := s.d.Cipher(r.Context()).Encrypt(r.Context(), []byte(token.AccessToken))
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	crt, err := s.d.Cipher(r.Context()).Encrypt(r.Context(), []byte(token.RefreshToken))
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	var conf identity.CredentialsOIDC
	creds, err := i.ParseCredentials(s.ID(), &conf)
	if errors.Is(err, herodot.ErrNotFound) {
		var err error
		if creds, err = identity.NewCredentialsOIDC(it, cat, crt, provider.Config().ID, claims.Subject); err != nil {
			return s.handleSettingsError(w, r, ctxUpdate, p, err)
		}
	} else if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	} else {
		creds.Identifiers = append(creds.Identifiers, identity.OIDCUniqueID(provider.Config().ID, claims.Subject))
		conf.Providers = append(conf.Providers, identity.CredentialsOIDCProvider{
			Subject: claims.Subject, Provider: provider.Config().ID,
			InitialAccessToken:  cat,
			InitialRefreshToken: crt,
			InitialIDToken:      it,
		})

		creds.Config, err = json.Marshal(conf)
		if err != nil {
			return s.handleSettingsError(w, r, ctxUpdate, p, err)
		}
	}

	i.Credentials[s.ID()] = *creds
	if err := s.d.SettingsHookExecutor().PostSettingsHook(w, r, s.SettingsStrategyID(), ctxUpdate, i, settings.WithCallback(func(ctxUpdate *settings.UpdateContext) error {
		return s.PopulateSettingsMethod(r, ctxUpdate.Session.Identity, ctxUpdate.Flow)
	})); err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	return nil
}

func (s *Strategy) unlinkProvider(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithOidcMethod) error {
	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context())).Before(time.Now()) {
		return s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
	}

	providers, err := s.Config(r.Context())
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	availableProviders, err := s.linkedProviders(r.Context(), r, providers, i)
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	var cc identity.CredentialsOIDC
	creds, err := i.ParseCredentials(s.ID(), &cc)
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	count, err := s.d.IdentityManager().CountActiveFirstFactorCredentials(r.Context(), i)
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	if count < 2 {
		return s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(UnlinkAllFirstFactorConnectionsError))
	}

	var found bool
	var updatedProviders []identity.CredentialsOIDCProvider
	var updatedIdentifiers []string
	for _, available := range availableProviders {
		if p.Unlink == available.Config().ID {
			for _, link := range cc.Providers {
				if link.Provider != p.Unlink {
					updatedIdentifiers = append(updatedIdentifiers, identity.OIDCUniqueID(link.Provider, link.Subject))
					updatedProviders = append(updatedProviders, link)
				} else {
					found = true
				}
			}
		}
	}

	if !found {
		return s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(UnknownConnectionValidationError))
	}

	creds.Identifiers = updatedIdentifiers
	creds.Config, err = json.Marshal(&identity.CredentialsOIDC{Providers: updatedProviders})
	if err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(err))

	}

	i.Credentials[s.ID()] = *creds
	if err := s.d.SettingsHookExecutor().PostSettingsHook(w, r, s.SettingsStrategyID(), ctxUpdate, i, settings.WithCallback(func(ctxUpdate *settings.UpdateContext) error {
		return s.PopulateSettingsMethod(r, ctxUpdate.Session.Identity, ctxUpdate.Flow)
	})); err != nil {
		return s.handleSettingsError(w, r, ctxUpdate, p, err)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithOidcMethod, err error) error {
	if e := new(settings.FlowNeedsReAuth); errors.As(err, &e) {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r,
			settings.ContinuityKey(s.SettingsStrategyID()), settings.ContinuityOptions(p, ctxUpdate.Session.Identity)...); err != nil {
			return err
		}
	}

	if ctxUpdate.Flow != nil {
		ctxUpdate.Flow.UI.ResetMessages()
		ctxUpdate.Flow.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	}

	return err
}
