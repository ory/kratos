// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"

	"github.com/ory/x/fetcher"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/x/decoderx"

	"golang.org/x/oauth2"

	"github.com/ory/kratos/selfservice/flow/login"

	"github.com/ory/kratos/text"

	"github.com/pkg/errors"

	"github.com/ory/kratos/continuity"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/x"
)

var _ registration.Strategy = new(Strategy)

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
	return s.populateMethod(r, f.UI, text.NewInfoRegistrationWith)
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
	var p UpdateRegistrationFlowWithOidcMethod
	if err := s.newLinkDecoder(&p, r); err != nil {
		return s.handleError(w, r, f, "", nil, err)
	}

	f.TransientPayload = p.TransientPayload

	var pid = p.Provider // this can come from both url query and post body
	if pid == "" {
		return errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), s.SettingsStrategyID(), s.SettingsStrategyID(), s.d); err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	}

	provider, err := s.provider(r.Context(), r, pid)
	if err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	}

	c, err := provider.OAuth2(r.Context())
	if err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	}

	req, err := s.validateFlow(r.Context(), r, f.ID)
	if err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	}

	if authenticated, err := s.alreadyAuthenticated(w, r, req); err != nil {
		return s.handleError(w, r, f, pid, nil, err)
	} else if authenticated {
		return errors.WithStack(registration.ErrAlreadyLoggedIn)
	}

	state := generateState(f.ID.String())
	if code, hasCode, _ := s.d.SessionTokenExchangePersister().CodeForFlow(r.Context(), f.ID); hasCode {
		state.setCode(code.InitCode)
	}
	if err := s.d.ContinuityManager().Pause(r.Context(), w, r, sessionName,
		continuity.WithPayload(&authCodeContainer{
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

func (s *Strategy) processRegistration(w http.ResponseWriter, r *http.Request, rf *registration.Flow, token *oauth2.Token, claims *Claims, provider Provider, container *authCodeContainer) (*login.Flow, error) {
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

	fetch := fetcher.NewFetcher(fetcher.WithClient(s.d.HTTPClient(r.Context())))
	jn, err := fetch.FetchContext(r.Context(), provider.Config().Mapper)
	if err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, nil, err)
	}

	i, va, err := s.createIdentity(w, r, rf, claims, provider, container, jn)
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

	var it string
	if idToken, ok := token.Extra("id_token").(string); ok {
		if it, err = s.d.Cipher(r.Context()).Encrypt(r.Context(), []byte(idToken)); err != nil {
			return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
		}
	}

	cat, err := s.d.Cipher(r.Context()).Encrypt(r.Context(), []byte(token.AccessToken))
	if err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
	}

	crt, err := s.d.Cipher(r.Context()).Encrypt(r.Context(), []byte(token.RefreshToken))
	if err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
	}

	creds, err := identity.NewCredentialsOIDC(it, cat, crt, provider.Config().ID, claims.Subject)
	if err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
	}

	i.SetCredentials(s.ID(), *creds)
	if err := s.d.RegistrationExecutor().PostRegistrationHook(w, r, identity.CredentialsTypeOIDC, provider.Config().ID, rf, i); err != nil {
		return nil, s.handleError(w, r, rf, provider.Config().ID, i.Traits, err)
	}

	return nil, nil
}

func (s *Strategy) createIdentity(w http.ResponseWriter, r *http.Request, a *registration.Flow, claims *Claims, provider Provider, container *authCodeContainer, jn *bytes.Buffer) (*identity.Identity, []VerifiedAddress, error) {
	var jsonClaims bytes.Buffer
	if err := json.NewEncoder(&jsonClaims).Encode(claims); err != nil {
		return nil, nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	vm, err := s.d.JsonnetVM(r.Context())
	if err != nil {
		return nil, nil, s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	vm.ExtCode("claims", jsonClaims.String())
	evaluated, err := vm.EvaluateAnonymousSnippet(provider.Config().Mapper, jn.String())
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

	s.d.Logger().
		WithRequest(r).
		WithField("oidc_provider", provider.Config().ID).
		WithSensitiveField("oidc_claims", claims).
		WithSensitiveField("mapper_jsonnet_output", evaluated).
		WithField("mapper_jsonnet_url", provider.Config().Mapper).
		Debug("OpenID Connect Jsonnet mapper completed.")
	return i, va, nil
}

func (s *Strategy) setTraits(w http.ResponseWriter, r *http.Request, a *registration.Flow, claims *Claims, provider Provider, container *authCodeContainer, evaluated string, i *identity.Identity) error {
	jsonTraits := gjson.Get(evaluated, "identity.traits")
	if !jsonTraits.IsObject() {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("OpenID Connect Jsonnet mapper did not return an object for key identity.traits. Please check your Jsonnet code!"))
	}

	traits, err := merge(container.Traits, json.RawMessage(jsonTraits.Raw))
	if err != nil {
		return s.handleError(w, r, a, provider.Config().ID, nil, err)
	}

	i.Traits = traits
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
