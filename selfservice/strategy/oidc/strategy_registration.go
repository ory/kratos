// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/fetcher"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlxx"
)

var _ registration.Strategy = new(Strategy)

var jsonnetCache, _ = ristretto.NewCache(&ristretto.Config{
	MaxCost:     100 << 20, // 100MB,
	NumCounters: 1_000_000, // 1kB per snippet -> 100k snippets -> 1M counters
	BufferItems: 64,
})

type MetadataType string

type VerifiedAddress struct {
	Value string                         `json:"value"`
	Via   identity.VerifiableAddressType `json:"via"`
}

const (
	VerifiedAddressesKey = "identity.verified_addresses"

	PublicMetadata MetadataType = "identity.metadata_public"
	AdminMetadata  MetadataType = "identity.metadata_admin"
)

func (s *Strategy) RegisterRegistrationRoutes(r *x.RouterPublic) {
	s.setRoutes(r)
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, f *registration.Flow) error {
	return s.populateMethod(r, f, text.NewInfoRegistrationWith)
}

// Update Registration Flow with OpenID Connect Method
//
// swagger:model updateRegistrationFlowWithOidcMethod
type UpdateRegistrationFlowWithOidcMethod struct {
	// The provider to register with
	//
	// required: true
	Provider string `json:"provider"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token"`

	// The identity traits
	Traits json.RawMessage `json:"traits"`

	// Method to use
	//
	// This field must be set to `oidc` when using the oidc method.
	//
	// required: true
	Method string `json:"method"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty"`

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

	// IDToken is an optional id token provided by an OIDC provider
	//
	// If submitted, it is verified using the OIDC provider's public key set and the claims are used to populate
	// the OIDC credentials of the identity.
	// If the OIDC provider does not store additional claims (such as name, etc.) in the IDToken itself, you can use
	// the `traits` field to populate the identity's traits. Note, that Apple only includes the users email in the IDToken.
	//
	// Supported providers are
	// - Apple
	// required: false
	IDToken string `json:"id_token,omitempty"`

	// IDTokenNonce is the nonce, used when generating the IDToken.
	// If the provider supports nonce validation, the nonce will be validated against this value and is required.
	//
	// required: false
	IDTokenNonce string `json:"id_token_nonce,omitempty"`
}

func (s *Strategy) newLinkDecoder(p interface{}, r *http.Request) error {
	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return err
	}

	raw, err := sjson.SetBytes(linkSchema, "properties.traits.$ref", ds.String()+"#/properties/traits")
	if err != nil {
		return errors.WithStack(err)
	}

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.dec.Decode(r, &p, compiler,
		decoderx.HTTPKeepRequestBody(true),
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderUseQueryAndBody(),
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.oidc.strategy.Register")
	defer otelx.End(span, &err)

	var p UpdateRegistrationFlowWithOidcMethod
	if err := s.newLinkDecoder(&p, r); err != nil {
		return s.handleError(w, r, f, "", nil, err)
	}

	f.TransientPayload = p.TransientPayload
	f.IDToken = p.IDToken
	f.RawIDTokenNonce = p.IDTokenNonce

	pid := p.Provider // this can come from both url query and post body
	if pid == "" {
		return errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if !strings.EqualFold(strings.ToLower(p.Method), s.SettingsStrategyID()) && p.Method != "" {
		// the user is sending a method that is not oidc, but the payload includes a provider
		s.d.Audit().
			WithRequest(r).
			WithField("provider", p.Provider).
			WithField("method", p.Method).
			Warn("The payload includes a `provider` field but is using a method other than `oidc`. Therefore, social sign in will not be executed.")
		return errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if err := flow.MethodEnabledAndAllowed(ctx, f.GetFlowName(), s.SettingsStrategyID(), s.SettingsStrategyID(), s.d); err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	}

	provider, err := s.provider(ctx, r, pid)
	if err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	}

	c, err := provider.OAuth2(ctx)
	if err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	}

	req, err := s.validateFlow(ctx, r, f.ID)
	if err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	}

	if authenticated, err := s.alreadyAuthenticated(w, r, req); err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	} else if authenticated {
		return errors.WithStack(registration.ErrAlreadyLoggedIn)
	}

	if p.IDToken != "" {
		claims, err := s.processIDToken(w, r, provider, p.IDToken, p.IDTokenNonce)
		if err != nil {
			return s.handleError(w, r, f, pid, nil, err)
		}
		_, err = s.processRegistration(w, r, f, nil, claims, provider, &AuthCodeContainer{
			FlowID:           f.ID.String(),
			Traits:           p.Traits,
			TransientPayload: f.TransientPayload,
		}, p.IDToken)
		if err != nil {
			return s.handleError(w, r, f, pid, nil, err)
		}
		return errors.WithStack(flow.ErrCompletedByStrategy)
	}

	state := generateState(f.ID.String())
	if code, hasCode, _ := s.d.SessionTokenExchangePersister().CodeForFlow(ctx, f.ID); hasCode {
		state.setCode(code.InitCode)
	}
	if err := s.d.ContinuityManager().Pause(ctx, w, r, sessionName,
		continuity.WithPayload(&AuthCodeContainer{
			State:            state.String(),
			FlowID:           f.ID.String(),
			Traits:           p.Traits,
			TransientPayload: f.TransientPayload,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	}

	var up map[string]string
	if err := json.NewDecoder(bytes.NewBuffer(p.UpstreamParameters)).Decode(&up); err != nil {
		return err
	}

	codeURL := c.AuthCodeURL(state.String(), append(UpstreamParameters(provider, up), provider.AuthCodeURLOptions(req)...)...)
	if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(codeURL))
	} else {
		http.Redirect(w, r, codeURL, http.StatusSeeOther)
	}

	return errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) registrationToLogin(w http.ResponseWriter, r *http.Request, rf *registration.Flow, providerID string) (*login.Flow, error) {
	// If return_to was set before, we need to preserve it.
	var opts []login.FlowOption
	if len(rf.ReturnTo) > 0 {
		opts = append(opts, login.WithFlowReturnTo(rf.ReturnTo))
	}

	if len(rf.UI.Messages) > 0 {
		opts = append(opts, login.WithFormErrorMessage(rf.UI.Messages))
	}

	opts = append(opts, login.WithInternalContext(rf.InternalContext))

	lf, _, err := s.d.LoginHandler().NewLoginFlow(w, r, rf.Type, opts...)
	if err != nil {
		return nil, err
	}

	err = s.d.SessionTokenExchangePersister().MoveToNewFlow(r.Context(), rf.ID, lf.ID)
	if err != nil {
		return nil, err
	}

	lf.RequestURL, err = x.TakeOverReturnToParameter(rf.RequestURL, lf.RequestURL)
	if err != nil {
		return nil, err
	}

	return lf, nil
}

func (s *Strategy) processRegistration(w http.ResponseWriter, r *http.Request, rf *registration.Flow, token *oauth2.Token, claims *Claims, provider Provider, container *AuthCodeContainer, idToken string) (*login.Flow, error) {
	if _, _, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), identity.CredentialsTypeOIDC, identity.OIDCUniqueID(provider.Config().ID, claims.Subject)); err == nil {
		// If the identity already exists, we should perform the login flow instead.

		// That will execute the "pre registration" hook which allows to e.g. disallow this flow. The registration
		// ui however will NOT be shown, instead the user is directly redirected to the auth path. That should then
		// do a silent re-request. While this might be a bit excessive from a network perspective it should usually
		// happen without any downsides to user experience as the request has already been authorized and should
		// not need additional consent/login.

		// This is kinda hacky but the only way to ensure seamless login/registration flows when using OIDC.
		s.d.Logger().WithRequest(r).WithField("provider", provider.Config().ID).
			WithField("subject", claims.Subject).
			Debug("Received successful OpenID Connect callback but user is already registered. Re-initializing login flow now.")

		lf, err := s.registrationToLogin(w, r, rf, provider.Config().ID)
		if err != nil {
			return nil, s.handleError(w, r, rf, provider.Config().ID, nil, err)
		}

		if _, err := s.processLogin(w, r, lf, token, claims, provider, container); err != nil {
			return lf, s.handleError(w, r, rf, provider.Config().ID, nil, err)
		}

		return nil, nil
	}

	fetch := fetcher.NewFetcher(fetcher.WithClient(s.d.HTTPClient(r.Context())), fetcher.WithCache(jsonnetCache, 60*time.Minute))
	jsonnetMapperSnippet, err := fetch.FetchContext(r.Context(), provider.Config().Mapper)
	if err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, nil, err)
	}

	i, va, err := s.createIdentity(w, r, rf, claims, provider, container, jsonnetMapperSnippet.Bytes())
	if err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, nil, err)
	}

	// Validate the identity itself
	if err := s.d.IdentityValidator().Validate(r.Context(), i); err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
	}

	for n := range i.VerifiableAddresses {
		verifiable := &i.VerifiableAddresses[n]
		for _, verified := range va {
			if verifiable.Via == verified.Via && verifiable.Value == verified.Value {
				verifiable.Status = identity.VerifiableAddressStatusCompleted
				verifiable.Verified = true
				t := sqlxx.NullTime(time.Now().UTC().Round(time.Second))
				verifiable.VerifiedAt = &t
			}
		}
	}

	var it string = idToken
	var cat, crt string
	if token != nil {
		if idToken, ok := token.Extra("id_token").(string); ok {
			if it, err = s.d.Cipher(r.Context()).Encrypt(r.Context(), []byte(idToken)); err != nil {
				return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
			}
		}

		cat, err = s.d.Cipher(r.Context()).Encrypt(r.Context(), []byte(token.AccessToken))
		if err != nil {
			return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
		}

		crt, err = s.d.Cipher(r.Context()).Encrypt(r.Context(), []byte(token.RefreshToken))
		if err != nil {
			return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
		}
	}

	creds, err := identity.NewCredentialsOIDC(it, cat, crt, provider.Config().ID, claims.Subject, provider.Config().OrganizationID)
	if err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
	}

	i.SetCredentials(s.ID(), *creds)
	if err := s.d.RegistrationExecutor().PostRegistrationHook(w, r, identity.CredentialsTypeOIDC, provider.Config().ID, rf, i); err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
	}

	return nil, nil
}

func (s *Strategy) createIdentity(w http.ResponseWriter, r *http.Request, a *registration.Flow, claims *Claims, provider Provider, container *AuthCodeContainer, jsonnetSnippet []byte) (*identity.Identity, []VerifiedAddress, error) {
	var jsonClaims bytes.Buffer
	if err := json.NewEncoder(&jsonClaims).Encode(claims); err != nil {
		return nil, nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	vm, err := s.d.JsonnetVM(r.Context())
	if err != nil {
		return nil, nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	vm.ExtCode("claims", jsonClaims.String())
	evaluated, err := vm.EvaluateAnonymousSnippet(provider.Config().Mapper, string(jsonnetSnippet))
	if err != nil {
		return nil, nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	i := identity.NewIdentity(s.d.Config().DefaultIdentityTraitsSchemaID(r.Context()))
	if err := s.setTraits(w, r, a, claims, provider, container, evaluated, i); err != nil {
		return nil, nil, s.handleError(w, r, a, provider.Config().ID, i.Traits, err)
	}

	if err := s.setMetadata(evaluated, i, PublicMetadata); err != nil {
		return nil, nil, s.handleError(w, r, a, provider.Config().ID, i.Traits, err)
	}

	if err := s.setMetadata(evaluated, i, AdminMetadata); err != nil {
		return nil, nil, s.handleError(w, r, a, provider.Config().ID, i.Traits, err)
	}

	va, err := s.extractVerifiedAddresses(evaluated)
	if err != nil {
		return nil, nil, s.handleError(w, r, a, provider.Config().ID, i.Traits, err)
	}

	if orgID := httprouter.ParamsFromContext(r.Context()).ByName("organization"); orgID != "" {
		i.OrganizationID = uuid.NullUUID{UUID: x.ParseUUID(orgID), Valid: true}
	}

	s.d.Logger().
		WithRequest(r).
		WithField("oidc_provider", provider.Config().ID).
		WithSensitiveField("oidc_claims", claims).
		WithSensitiveField("mapper_jsonnet_output", evaluated).
		WithField("mapper_jsonnet_url", provider.Config().Mapper).
		Debug("OpenID Connect Jsonnet mapper completed.")
	return i, va, nil
}

func (s *Strategy) setTraits(w http.ResponseWriter, r *http.Request, a *registration.Flow, claims *Claims, provider Provider, container *AuthCodeContainer, evaluated string, i *identity.Identity) error {
	jsonTraits := gjson.Get(evaluated, "identity.traits")
	if !jsonTraits.IsObject() {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("OpenID Connect Jsonnet mapper did not return an object for key identity.traits. Please check your Jsonnet code!"))
	}

	if container != nil {
		traits, err := merge(container.Traits, json.RawMessage(jsonTraits.Raw))
		if err != nil {
			return s.handleError(w, r, a, provider.Config().ID, nil, err)
		}

		i.Traits = traits
	} else {
		i.Traits = identity.Traits(json.RawMessage(jsonTraits.Raw))
	}
	s.d.Logger().
		WithRequest(r).
		WithField("oidc_provider", provider.Config().ID).
		WithSensitiveField("identity_traits", i.Traits).
		WithSensitiveField("mapper_jsonnet_output", evaluated).
		WithField("mapper_jsonnet_url", provider.Config().Mapper).
		Debug("Merged form values and OpenID Connect Jsonnet output.")
	return nil
}

func (s *Strategy) setMetadata(evaluated string, i *identity.Identity, m MetadataType) error {
	if m != PublicMetadata && m != AdminMetadata {
		return errors.Errorf("undefined metadata type: %s", m)
	}

	metadata := gjson.Get(evaluated, string(m))
	if metadata.Exists() && !metadata.IsObject() {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("OpenID Connect Jsonnet mapper did not return an object for key %s. Please check your Jsonnet code!", m))
	}

	switch m {
	case PublicMetadata:
		i.MetadataPublic = []byte(metadata.Raw)
	case AdminMetadata:
		i.MetadataAdmin = []byte(metadata.Raw)
	}

	return nil
}

func (s *Strategy) extractVerifiedAddresses(evaluated string) ([]VerifiedAddress, error) {
	if verifiedAddresses := gjson.Get(evaluated, VerifiedAddressesKey); verifiedAddresses.Exists() {
		if !verifiedAddresses.IsArray() {
			return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("OpenID Connect Jsonnet mapper did not return an array for key %s. Please check your Jsonnet code!", VerifiedAddressesKey))
		}

		var va []VerifiedAddress
		if err := json.Unmarshal([]byte(verifiedAddresses.Raw), &va); err != nil {
			return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Failed to unmarshal value for key %s. Please check your Jsonnet code!", VerifiedAddressesKey).WithDebugf("%s", err))
		}

		for i := range va {
			va := &va[i]
			if va.Via == identity.VerifiableAddressTypeEmail {
				va.Value = strings.ToLower(strings.TrimSpace(va.Value))
			}
		}

		return va, nil
	}

	return nil, nil
}
