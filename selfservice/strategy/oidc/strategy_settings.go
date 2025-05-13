// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/stringsx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/strategy"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

//go:embed .schema/settings.schema.json
var settingsSchema []byte

var (
	_                                settings.Strategy = new(Strategy)
	UnknownConnectionValidationError                   = &jsonschema.ValidationError{
		Message: "can not unlink non-existing OpenID Connect connection", InstancePtr: "#/",
	}
)

var ConnectionExistValidationError = &jsonschema.ValidationError{
	Message: "can not link unknown or already existing OpenID Connect connection", InstancePtr: "#/",
}

var UnlinkAllFirstFactorConnectionsError = &jsonschema.ValidationError{
	Message: "can not unlink OpenID Connect connection because it is the last remaining first factor credential", InstancePtr: "#/",
}

func (s *Strategy) RegisterSettingsRoutes(*x.RouterPublic) {}

func (s *Strategy) SettingsStrategyID() string {
	return s.ID().String()
}

func (s *Strategy) decoderSettings(ctx context.Context, p *updateSettingsFlowWithOidcMethod, r *http.Request) error {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(ctx)
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

func (s *Strategy) linkedProviders(conf *ConfigurationCollection, confidential *identity.Identity) ([]Provider, error) {
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

func (s *Strategy) linkableProviders(conf *ConfigurationCollection, confidential *identity.Identity) ([]Provider, error) {
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

func (s *Strategy) PopulateSettingsMethod(ctx context.Context, r *http.Request, id *identity.Identity, sr *settings.Flow) (err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.Strategy.PopulateSettingsMethod")
	defer otelx.End(span, &err)

	if sr.Type != flow.TypeBrowser {
		return nil
	}

	conf, err := s.Config(ctx)
	if err != nil {
		return err
	}

	linkable, err := s.linkableProviders(conf, id)
	if err != nil {
		return err
	}

	linked, err := s.linkedProviders(conf, id)
	if err != nil {
		return err
	}

	sr.UI.GetNodes().Remove("unlink", "link")
	sr.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	for _, l := range linkable {
		// We do not want to offer to link SSO providers in the settings.
		if l.Config().OrganizationID != "" {
			continue
		}
		sr.UI.GetNodes().Append(NewLinkNode(l.Config().ID, stringsx.Coalesce(l.Config().Label, l.Config().ID)))
	}

	count, err := s.d.IdentityManager().CountActiveFirstFactorCredentials(ctx, id)
	if err != nil {
		return err
	}

	if count > 1 {
		// This means that we're able to remove a connection because it is the last configured credential. If it is
		// removed, the identity is no longer able to sign in.
		for _, l := range linked {
			sr.UI.GetNodes().Append(NewUnlinkNode(l.Config().ID, stringsx.Coalesce(l.Config().Label, l.Config().ID)))
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

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (p *updateSettingsFlowWithOidcMethod) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.FlowID)
}

func (p *updateSettingsFlowWithOidcMethod) SetFlowID(rid uuid.UUID) {
	p.FlowID = rid.String()
}

func (s *Strategy) Settings(ctx context.Context, w http.ResponseWriter, r *http.Request, f *settings.Flow, ss *session.Session) (_ *settings.UpdateContext, err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.Strategy.Settings")
	defer otelx.End(span, &err)

	var p updateSettingsFlowWithOidcMethod
	if err := s.decoderSettings(ctx, &p, r); err != nil {
		return nil, err
	}
	f.TransientPayload = p.TransientPayload

	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, f, ss, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		if !s.d.Config().SelfServiceStrategy(ctx, s.SettingsStrategyID()).Enabled {
			return nil, s.handleMethodNotAllowedError(errors.WithStack(herodot.ErrNotFound.WithReason(strategy.EndpointDisabledMessage)))
		}

		if len(p.Link) > 0 {
			if err := s.initLinkProvider(ctx, w, r, ctxUpdate, &p); err != nil {
				return nil, err
			}

			return ctxUpdate, nil
		} else if len(p.Unlink) > 0 {
			if err := s.unlinkProvider(ctx, w, r, ctxUpdate, &p); err != nil {
				return nil, err
			}

			return ctxUpdate, nil
		}

		return nil, s.handleSettingsError(ctx, w, r, ctxUpdate, &p, errors.WithStack(herodot.ErrInternalServerError.WithReason("Expected either link or unlink to be set when continuing flow but both are unset.")))
	} else if err != nil {
		return nil, s.handleSettingsError(ctx, w, r, ctxUpdate, &p, err)
	}

	if len(p.Link)+len(p.Unlink) == 0 {
		span.SetAttributes(attribute.String("not_responsible_reason", "neither link nor unlink set"))
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if !s.d.Config().SelfServiceStrategy(ctx, s.SettingsStrategyID()).Enabled {
		return nil, s.handleMethodNotAllowedError(errors.WithStack(herodot.ErrNotFound.WithReason(strategy.EndpointDisabledMessage)))
	}

	switch l, u := len(p.Link), len(p.Unlink); {
	case l > 0 && u > 0:
		return nil, s.handleSettingsError(ctx, w, r, ctxUpdate, &p, errors.WithStack(&jsonschema.ValidationError{
			Message:     "it is not possible to link and unlink providers in the same request",
			InstancePtr: "#/",
		}))
	case l > 0:
		if err := s.initLinkProvider(ctx, w, r, ctxUpdate, &p); err != nil {
			return nil, err
		}
		return ctxUpdate, nil
	case u > 0:
		if err := s.unlinkProvider(ctx, w, r, ctxUpdate, &p); err != nil {
			return nil, err
		}
		return ctxUpdate, nil
	}
	// this case should never be reached as we previously checked whether link and unlink are both empty
	return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
}

func (s *Strategy) isLinkable(ctx context.Context, ctxUpdate *settings.UpdateContext, toLink string) (*identity.Identity, error) {
	providers, err := s.Config(ctx)
	if err != nil {
		return nil, err
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(ctx, ctxUpdate.Session.Identity.ID)
	if err != nil {
		return nil, err
	}

	linkable, err := s.linkableProviders(providers, i)
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

func (s *Strategy) initLinkProvider(ctx context.Context, w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithOidcMethod) error {
	if _, err := s.isLinkable(ctx, ctxUpdate, p.Link); err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)).Before(time.Now()) {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
	}

	provider, err := s.Provider(ctx, p.Link)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	req, err := s.validateFlow(ctx, r, ctxUpdate.Flow.ID)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	state, pkce, err := s.GenerateState(ctx, provider, ctxUpdate.Flow.ID)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}
	if err := s.d.ContinuityManager().Pause(ctx, w, r, sessionName,
		continuity.WithPayload(&AuthCodeContainer{
			State:  state,
			FlowID: ctxUpdate.Flow.ID.String(),
			Traits: p.Traits,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	var up map[string]string
	if err := json.NewDecoder(bytes.NewBuffer(p.UpstreamParameters)).Decode(&up); err != nil {
		return err
	}

	codeURL, err := getAuthRedirectURL(ctx, provider, req, state, up, pkce)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(codeURL))
	} else {
		http.Redirect(w, r, codeURL, http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) linkProvider(ctx context.Context, w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, token *identity.CredentialsOIDCEncryptedTokens, claims *Claims, provider Provider) error {
	p := &updateSettingsFlowWithOidcMethod{
		Link: provider.Config().ID, FlowID: ctxUpdate.Flow.ID.String(),
	}

	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)).Before(time.Now()) {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
	}

	i, err := s.isLinkable(ctx, ctxUpdate, p.Link)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	if err := s.linkCredentials(ctx, i, token, provider.Config().ID, claims.Subject, provider.Config().OrganizationID); err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	if err := s.d.SettingsHookExecutor().PostSettingsHook(ctx, w, r, s.SettingsStrategyID(), ctxUpdate, i, settings.WithCallback(func(ctxUpdate *settings.UpdateContext) error {
		// Credential population is done by PostSettingsHook on ctxUpdate.Session.Identity
		return s.PopulateSettingsMethod(ctx, r, ctxUpdate.Session.Identity, ctxUpdate.Flow)
	})); err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	return nil
}

func (s *Strategy) unlinkProvider(ctx context.Context, w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithOidcMethod) error {
	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)).Before(time.Now()) {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
	}

	providers, err := s.Config(ctx)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(ctx, ctxUpdate.Session.Identity.ID)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	availableProviders, err := s.linkedProviders(providers, i)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	var cc identity.CredentialsOIDC
	creds, err := i.ParseCredentials(s.ID(), &cc)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	count, err := s.d.IdentityManager().CountActiveFirstFactorCredentials(ctx, i)
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	if count < 2 {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, errors.WithStack(UnlinkAllFirstFactorConnectionsError))
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
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, errors.WithStack(UnknownConnectionValidationError))
	}

	creds.Identifiers = updatedIdentifiers
	creds.Config, err = json.Marshal(&identity.CredentialsOIDC{Providers: updatedProviders})
	if err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, errors.WithStack(err))
	}

	i.Credentials[s.ID()] = *creds
	if err := s.d.SettingsHookExecutor().PostSettingsHook(ctx, w, r, s.SettingsStrategyID(), ctxUpdate, i, settings.WithCallback(func(ctxUpdate *settings.UpdateContext) error {
		// Credential population is done by PostSettingsHook on ctxUpdate.Session.Identity
		return s.PopulateSettingsMethod(ctx, r, ctxUpdate.Session.Identity, ctxUpdate.Flow)
	})); err != nil {
		return s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) handleSettingsError(ctx context.Context, w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *updateSettingsFlowWithOidcMethod, err error) error {
	if e := new(settings.FlowNeedsReAuth); errors.As(err, &e) {
		if err := s.d.ContinuityManager().Pause(ctx, w, r,
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

func (s *Strategy) Link(ctx context.Context, i *identity.Identity, credentialsConfig sqlxx.JSONRawMessage) (err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.Strategy.Link")
	defer otelx.End(span, &err)

	var credentialsOIDCConfig identity.CredentialsOIDC
	if err := json.Unmarshal(credentialsConfig, &credentialsOIDCConfig); err != nil {
		return err
	}
	if len(credentialsOIDCConfig.Providers) != 1 {
		return errors.New("no oidc provider was set")
	}
	credentialsOIDCProvider := credentialsOIDCConfig.Providers[0]

	if err := s.linkCredentials(
		ctx,
		i,
		// The tokens in this credential are coming from the existing identity. Hence, the values are already encrypted.
		credentialsOIDCProvider.GetTokens(),
		credentialsOIDCProvider.Provider,
		credentialsOIDCProvider.Subject,
		credentialsOIDCProvider.Organization,
	); err != nil {
		return err
	}

	if err := s.d.IdentityManager().Update(ctx, i, identity.ManagerAllowWriteProtectedTraits); err != nil {
		return err
	}

	return nil
}

func (s *Strategy) CompletedLogin(sess *session.Session, data *flow.DuplicateCredentialsData) error {
	var credentialsOIDCConfig identity.CredentialsOIDC
	if err := json.Unmarshal(data.CredentialsConfig, &credentialsOIDCConfig); err != nil {
		return err
	}
	if len(credentialsOIDCConfig.Providers) != 1 {
		return errors.New("no oidc provider was set")
	}
	credentialsOIDCProvider := credentialsOIDCConfig.Providers[0]

	sess.CompletedLoginForWithProvider(
		s.ID(),
		identity.AuthenticatorAssuranceLevel1,
		credentialsOIDCProvider.Provider,
		credentialsOIDCProvider.Organization,
	)

	return nil
}

func (s *Strategy) SetDuplicateCredentials(f flow.InternalContexter, duplicateIdentifier string, credentials identity.Credentials, provider string) error {
	var credentialsOIDCConfig identity.CredentialsOIDC
	if err := json.Unmarshal(credentials.Config, &credentialsOIDCConfig); err != nil {
		return err
	}

	// We want to only set the provider in the credentials config that was used to authenticate the user.
	for _, p := range credentialsOIDCConfig.Providers {
		if p.Provider == provider {
			credentialsOIDCConfig.Providers = []identity.CredentialsOIDCProvider{p}
			config, err := json.Marshal(credentialsOIDCConfig)
			if err != nil {
				return err
			}
			return flow.SetDuplicateCredentials(f, flow.DuplicateCredentialsData{
				CredentialsType:     s.ID(),
				CredentialsConfig:   config,
				DuplicateIdentifier: duplicateIdentifier,
			})
		}
	}
	return fmt.Errorf("provider %q not found in credentials", provider)
}
