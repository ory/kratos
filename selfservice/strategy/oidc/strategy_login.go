// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"
	oidcv1 "github.com/ory/kratos/gen/oidc/v1"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flowhelpers"
	"github.com/ory/kratos/selfservice/strategy/idfirst"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
)

var (
	_ login.AAL1FormHydrator = (*Strategy)(nil)
	_ login.Strategy         = (*Strategy)(nil)
)

// Update Login Flow with OpenID Connect Method
//
// swagger:model updateLoginFlowWithOidcMethod
type UpdateLoginFlowWithOidcMethod struct {
	// The provider to register with
	//
	// required: true
	Provider string `json:"provider"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token"`

	// Method to use
	//
	// This field must be set to `oidc` when using the oidc method.
	//
	// required: true
	Method string `json:"method"`

	// The identity traits. This is a placeholder for the registration flow.
	Traits json.RawMessage `json:"traits"`

	// UpstreamParameters are the parameters that are passed to the upstream identity provider.
	//
	// These parameters are optional and depend on what the upstream identity provider supports.
	// Supported parameters are:
	// - `login_hint` (string): The `login_hint` parameter suppresses the account chooser and either pre-fills the email box on the sign-in form, or selects the proper session.
	// - `hd` (string): The `hd` parameter limits the login/registration process to a Google Organization, e.g. `mycollege.edu`.
	// - `prompt` (string): The `prompt` specifies whether the Authorization Server prompts the End-User for reauthentication and consent, e.g. `select_account`.
	// - `acr_values` (string): The `acr_values` specifies the Authentication Context Class Reference values for the authorization request.
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
	// - Google
	// required: false
	IDToken string `json:"id_token,omitempty"`

	// IDTokenNonce is the nonce, used when generating the IDToken.
	// If the provider supports nonce validation, the nonce will be validated against this value and required.
	//
	// required: false
	IDTokenNonce string `json:"id_token_nonce,omitempty"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (s *Strategy) handleConflictingIdentity(ctx context.Context, loginFlow *login.Flow, token *identity.CredentialsOIDCEncryptedTokens, claims *Claims, provider Provider, container *AuthCodeContainer) (verdict ConflictingIdentityVerdict, id *identity.Identity, credentials *identity.Credentials, err error) {
	if s.conflictingIdentityPolicy == nil {
		return ConflictingIdentityVerdictReject, nil, nil, nil
	}

	// Find out if there is a conflicting identity
	newIdentity, va, err := s.newIdentityFromClaims(ctx, claims, provider, container, loginFlow.IdentitySchema)
	if err != nil {
		return ConflictingIdentityVerdictReject, nil, nil, nil
	}

	// Validate the identity itself
	// We ignore the error here because the claims may not fulfil the requirements
	// of the identity schema.
	//
	// However, this is not a problem because the identity will be merged with the existing
	// identity and the existing identity will be updated with the new credentials, but not any traits.
	//
	// We do need the validation step however, to "hydrate" the verifiable address of the user, which is then
	// used in subsequent calls to match the existing with the new identity.
	_ = s.d.IdentityValidator().Validate(ctx, newIdentity)

	for n := range newIdentity.VerifiableAddresses {
		verifiable := &newIdentity.VerifiableAddresses[n]
		for _, verified := range va {
			if verifiable.Via == verified.Via && verifiable.Value == verified.Value {
				verifiable.Status = identity.VerifiableAddressStatusCompleted
				verifiable.Verified = true
				t := sqlxx.NullTime(time.Now().UTC().Round(time.Second))
				verifiable.VerifiedAt = &t
			}
		}
	}

	creds, err := identity.NewOIDCLikeCredentials(token, s.ID(), provider.Config().ID, claims.Subject, provider.Config().OrganizationID)
	if err != nil {
		return ConflictingIdentityVerdictUnknown, nil, nil, err
	}

	newIdentity.SetCredentials(s.ID(), *creds)

	existingIdentity, _, _, err := s.d.IdentityManager().ConflictingIdentity(ctx, newIdentity)
	if err != nil {
		return ConflictingIdentityVerdictReject, nil, nil, nil
	}

	verdict = s.conflictingIdentityPolicy(ctx, existingIdentity, newIdentity, provider, claims)
	if verdict == ConflictingIdentityVerdictMerge {
		if err = existingIdentity.MergeOIDCCredentials(s.ID(), *creds); err != nil {
			return ConflictingIdentityVerdictUnknown, nil, nil, err
		}

		if err = s.d.PrivilegedIdentityPool().UpdateIdentity(ctx, existingIdentity); err != nil {
			return ConflictingIdentityVerdictUnknown, nil, nil, err
		}
	}

	return verdict, existingIdentity, creds, nil
}

func verifiableAddressHash(i *identity.Identity) [sha256.Size]byte {
	h := sha256.New()
	for _, a := range i.VerifiableAddresses {
		h.Write([]byte(a.Signature())) // sha256 Write never returns an error
	}
	return [sha256.Size]byte(h.Sum(nil))
}

func recoveryAddressHash(i *identity.Identity) [sha256.Size]byte {
	h := sha256.New()
	for _, a := range i.RecoveryAddresses {
		h.Write([]byte(a.Signature())) // sha256 Write never returns an error
	}
	return [sha256.Size]byte(h.Sum(nil))
}

// jsonEqual reports whether two JSON values are deeply equal, ignoring key
// order and whitespace differences. Array element order is significant.
func jsonEqual(a, b json.RawMessage) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	var aVal, bVal any
	if err := json.Unmarshal(a, &aVal); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bVal); err != nil {
		return false
	}
	return reflect.DeepEqual(aVal, bVal)
}

// UpdateIdentityFromClaims re-runs the Jsonnet claims mapper and applies the
// result to an existing identity. It returns true if traits or metadata changed.
// Unlike registration, user-supplied form traits are not merged — only the
// mapper output is applied.
func (s *Strategy) UpdateIdentityFromClaims(ctx context.Context, claims *Claims, provider Provider, i *identity.Identity) (changed bool, err error) {
	evaluated, _, err := s.EvaluateClaimsMapper(ctx, claims, provider, i)
	if err != nil {
		return false, err
	}

	// Save the current state for comparison.
	oldTraits := json.RawMessage(i.Traits)
	oldMetadataPublic := json.RawMessage(i.MetadataPublic)
	oldMetadataAdmin := json.RawMessage(i.MetadataAdmin)
	oldVerifiableHash := verifiableAddressHash(i)
	oldRecoveryHash := recoveryAddressHash(i)

	// Merge the mapper output with the existing identity traits. The mapper
	// output takes precedence, but existing traits that the mapper does not
	// output are preserved.
	jsonTraits := gjson.Get(evaluated, "identity.traits")
	if !jsonTraits.IsObject() {
		return false, errors.WithStack(herodot.ErrInternalServerError().WithReasonf("OpenID Connect Jsonnet mapper did not return an object for key identity.traits. Please check your Jsonnet code!"))
	}
	// merge(override, base) merges base into override, with override winning.
	// Pass mapper output as override so it takes precedence over existing traits,
	// while traits not in the mapper output are preserved from the base.
	mergedTraits, err := merge(json.RawMessage(jsonTraits.Raw), json.RawMessage(i.Traits))
	if err != nil {
		return false, err
	}
	i.Traits = mergedTraits

	// Only update metadata if the mapper explicitly outputs the key. When the
	// mapper omits metadata_public or metadata_admin (common for mappers
	// written for registration only), we preserve the existing values rather
	// than wiping them.
	if gjson.Get(evaluated, string(PublicMetadata)).Exists() {
		if err = s.setMetadata(evaluated, i, PublicMetadata); err != nil {
			return false, err
		}
	}
	if gjson.Get(evaluated, string(AdminMetadata)).Exists() {
		if err = s.setMetadata(evaluated, i, AdminMetadata); err != nil {
			return false, err
		}
	}

	// Save existing addresses before Validate regenerates them from the schema.
	oldVerifiableAddresses := make([]identity.VerifiableAddress, len(i.VerifiableAddresses))
	copy(oldVerifiableAddresses, i.VerifiableAddresses)
	oldRecoveryAddresses := make([]identity.RecoveryAddress, len(i.RecoveryAddresses))
	copy(oldRecoveryAddresses, i.RecoveryAddresses)

	// Validate the updated identity against the schema to prevent persisting
	// invalid traits produced by a broken mapper.
	if err = s.d.IdentityValidator().Validate(ctx, i); err != nil {
		return false, err
	}

	// Restore existing verification and recovery state for addresses that
	// survived validation. Validate regenerates both VerifiableAddresses and
	// RecoveryAddresses from the schema, losing IDs, verified/verifiedAt
	// status, etc. Carry over the state for matching addresses.
	for n := range i.VerifiableAddresses {
		addr := &i.VerifiableAddresses[n]
		for _, old := range oldVerifiableAddresses {
			if addr.Via == old.Via && addr.Value == old.Value {
				*addr = old
				break
			}
		}
	}
	for n := range i.RecoveryAddresses {
		addr := &i.RecoveryAddresses[n]
		for _, old := range oldRecoveryAddresses {
			if addr.Via == old.Via && addr.Value == old.Value {
				*addr = old
				break
			}
		}
	}

	// Apply verified addresses from the mapper output, matching the
	// registration behavior. This lets the mapper carry over the verified
	// status from the SSO provider on every login (e.g., when the user's
	// email changes upstream).
	va, err := s.extractVerifiedAddresses(evaluated)
	if err != nil {
		return false, err
	}
	for n := range i.VerifiableAddresses {
		verifiable := &i.VerifiableAddresses[n]
		for _, verified := range va {
			if verifiable.Via == verified.Via && verifiable.Value == verified.Value {
				if !verifiable.Verified {
					verifiable.Status = identity.VerifiableAddressStatusCompleted
					verifiable.Verified = true
					t := sqlxx.NullTime(time.Now().UTC().Round(time.Second))
					verifiable.VerifiedAt = &t
				}
			}
		}
	}

	changed = !jsonEqual(oldTraits, json.RawMessage(i.Traits)) ||
		!jsonEqual(oldMetadataPublic, json.RawMessage(i.MetadataPublic)) ||
		!jsonEqual(oldMetadataAdmin, json.RawMessage(i.MetadataAdmin)) ||
		oldVerifiableHash != verifiableAddressHash(i) ||
		oldRecoveryHash != recoveryAddressHash(i)

	s.d.Logger().
		WithField("oidc_provider", provider.Config().ID).
		WithField("identity_id", i.ID).
		WithField("identity_changed", changed).
		Debug("Re-evaluated OpenID Connect claims mapper on login.")

	return changed, nil
}

func (s *Strategy) ProcessLogin(ctx context.Context, w http.ResponseWriter, r *http.Request, loginFlow *login.Flow, token *identity.CredentialsOIDCEncryptedTokens, claims *Claims, provider Provider, container *AuthCodeContainer) (_ *registration.Flow, err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.Strategy.processLogin")
	defer otelx.End(span, &err)

	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, s.ID(), identity.OIDCUniqueID(provider.Config().ID, claims.Subject))
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows()) {
			var verdict ConflictingIdentityVerdict
			verdict, i, c, err = s.handleConflictingIdentity(ctx, loginFlow, token, claims, provider, container)
			switch verdict {
			case ConflictingIdentityVerdictMerge:
				// Do nothing
			case ConflictingIdentityVerdictReject:
				// If no account was found we're "manually" creating a new registration flow and redirecting the browser
				// to that endpoint.

				// That will execute the "pre registration" hook which allows to e.g. disallow this request. The registration
				// ui however will NOT be shown, instead the user is directly redirected to the auth path. That should then
				// do a silent re-request. While this might be a bit excessive from a network perspective it should usually
				// happen without any downsides to user experience as the flow has already been authorized and should
				// not need additional consent/login.

				// This is kinda hacky but the only way to ensure seamless login/registration flows when using OIDC.
				s.d.
					Logger().
					WithField("provider", provider.Config().ID).
					WithField("subject", claims.Subject).
					Debug("Received successful OpenID Connect callback but user is not registered. Re-initializing registration flow now.")

				// If return_to was set before, we need to preserve it.
				var opts []registration.FlowOption
				if len(loginFlow.ReturnTo) > 0 {
					opts = append(opts, registration.WithFlowReturnTo(loginFlow.ReturnTo))
				}

				if loginFlow.OAuth2LoginChallenge.String() != "" {
					opts = append(opts, registration.WithFlowOAuth2LoginChallenge(loginFlow.OAuth2LoginChallenge.String()))
				}

				registrationFlow, err := s.d.RegistrationHandler().NewRegistrationFlow(w, r, loginFlow.Type, opts...)
				if err != nil {
					return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err)
				}

				err = s.d.SessionTokenExchangePersister().MoveToNewFlow(ctx, loginFlow.ID, registrationFlow.ID)
				if err != nil {
					return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err)
				}

				registrationFlow.OrganizationID = loginFlow.OrganizationID
				registrationFlow.IDToken = loginFlow.IDToken
				registrationFlow.RawIDTokenNonce = loginFlow.RawIDTokenNonce
				registrationFlow.TransientPayload = loginFlow.TransientPayload
				registrationFlow.Active = s.ID()
				registrationFlow.IdentitySchema = loginFlow.IdentitySchema

				// We are converting the flow here, but want to retain the original request URL.
				registrationFlow.RequestURL = loginFlow.RequestURL

				if _, err := s.processRegistration(ctx, w, r, registrationFlow, token, claims, provider, container); err != nil {
					return registrationFlow, err
				}

				return nil, nil
			case ConflictingIdentityVerdictUnknown:
				fallthrough
			default:
				// This should never happen if err == nil, but just for safety:
				if err != nil {
					return nil, err
				}
				return nil, errors.WithStack(herodot.ErrInternalServerError().WithReason("The OpenID Connect identity merge policy returned an unknown verdict without other error details, which prevents the sign up from completing. Please report this as a bug."))
			}

		} else {
			return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err)
		}
	}

	var oidcCredentials identity.CredentialsOIDC
	if err := json.NewDecoder(bytes.NewBuffer(c.Config)).Decode(&oidcCredentials); err != nil {
		return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, x.WrapWithIdentityIDError(errors.WithStack(herodot.ErrInternalServerError().WithReason("The OpenID Connect credentials could not be decoded properly").WithDebug(err.Error())), i.ID))
	}

	sess := session.NewInactiveSession()
	sess.CompletedLoginForWithProvider(s.ID(), identity.AuthenticatorAssuranceLevel1, provider.Config().ID, provider.Config().OrganizationID)

	for _, c := range oidcCredentials.Providers {
		if c.Subject == claims.Subject && c.Provider == provider.Config().ID {
			if provider.Config().UpdateIdentityOnLogin == UpdateIdentityOnLoginAutomatic {
				identityChanged, err := s.UpdateIdentityFromClaims(ctx, claims, provider, i)
				if err != nil {
					return nil, x.WrapWithIdentityIDError(s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err), i.ID)
				}
				if identityChanged {
					if err := s.d.PrivilegedIdentityPool().UpdateIdentity(ctx, i); err != nil {
						return nil, x.WrapWithIdentityIDError(s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err), i.ID)
					}
				}
			}

			if err = s.d.LoginHookExecutor().PostLoginHook(w, r, node.OpenIDConnectGroup, loginFlow, i, sess, provider.Config().ID); err != nil {
				return nil, x.WrapWithIdentityIDError(s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err), i.ID)
			}
			return nil, nil
		}
	}

	return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, x.WrapWithIdentityIDError(errors.WithStack(herodot.ErrInternalServerError().WithReason("Unable to find matching OpenID Connect credentials.").WithDebugf(`Unable to find credentials that match the given provider "%s" and subject "%s".`, provider.Config().ID, claims.Subject)), i.ID))
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, _ *session.Session) (i *identity.Identity, err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.oidc.Strategy.Login")
	defer otelx.End(span, &err)

	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		span.SetAttributes(attribute.String("not_responsible_reason", "requested AAL is not AAL1"))
		return nil, err
	}

	var p UpdateLoginFlowWithOidcMethod
	if err := s.newLinkDecoder(ctx, &p, r, &f.IdentitySchema); err != nil {
		return nil, s.HandleError(ctx, w, r, f, "", nil, err)
	}

	f.IDToken = p.IDToken
	f.RawIDTokenNonce = p.IDTokenNonce
	f.TransientPayload = p.TransientPayload

	pid := p.Provider // this can come from both url query and post body
	if pid == "" {
		span.SetAttributes(attribute.String("not_responsible_reason", "provider ID missing"))
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if !strings.EqualFold(strings.ToLower(p.Method), s.SettingsStrategyID()) && p.Method != "" {
		// the user is sending a method that is not oidc, but the payload includes a provider
		s.d.Logger().
			WithRequest(r).
			WithField("provider", p.Provider).
			WithField("method", p.Method).
			Warn("The payload includes a `provider` field but is using a method other than `oidc`. Therefore, social sign in will not be executed.")
		span.SetAttributes(attribute.String("not_responsible_reason", "method is not oidc"))
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if err := flow.MethodEnabledAndAllowed(ctx, f.GetFlowName(), s.SettingsStrategyID(), s.SettingsStrategyID(), s.d); err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, s.handleMethodNotAllowedError(err))
	}

	provider, err := s.Provider(ctx, pid)
	if err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}

	req, err := s.validateFlow(ctx, r, f.ID, oidcv1.FlowKind_FLOW_KIND_LOGIN)
	if err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}

	if authenticated, err := s.alreadyAuthenticated(ctx, w, r, req); err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	} else if authenticated {
		return i, nil
	}

	if p.IDToken != "" {
		claims, err := s.ProcessIDToken(r, provider, p.IDToken, p.IDTokenNonce)
		if err != nil {
			return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
		}
		_, err = s.ProcessLogin(ctx, w, r, f, nil, claims, provider, &AuthCodeContainer{
			FlowID: f.ID.String(),
			Traits: p.Traits,
		})
		if errors.Is(err, flow.ErrCompletedByStrategy) {
			return nil, err
		} else if err != nil {
			return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
		}
		return nil, errors.WithStack(flow.ErrCompletedByStrategy)
	}

	state, pkce, err := s.GenerateState(ctx, provider, f)
	if err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}
	if err := s.d.ContinuityManager().Pause(ctx, w, r, sessionName,
		continuity.WithPayload(&AuthCodeContainer{
			State:            state,
			FlowID:           f.ID.String(),
			Traits:           p.Traits,
			TransientPayload: f.TransientPayload,
			IdentitySchema:   f.IdentitySchema,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}

	// For API/native flows, persist TransientPayload in InternalContext so it
	// survives the OIDC redirect. Browser flows restore it from the continuity
	// cookie instead, which is not available in native flows because the
	// callback comes from a different user agent (system browser/webview).
	if f.Type == flow.TypeAPI && len(f.TransientPayload) > 0 {
		f.EnsureInternalContext()
		ic, err := sjson.SetRawBytes(f.InternalContext, "transient_payload", f.TransientPayload)
		if err != nil {
			return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
		}
		f.InternalContext = ic
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, errors.WithStack(herodot.ErrInternalServerError().WithReason("Could not update flow").WithWrap(err)))
	}

	var up map[string]string
	if err := json.NewDecoder(bytes.NewBuffer(p.UpstreamParameters)).Decode(&up); err != nil {
		return nil, err
	}

	codeURL, err := getAuthRedirectURL(ctx, provider, f, state, up, pkce)
	if err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}

	if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(codeURL))
	} else {
		http.Redirect(w, r, codeURL, http.StatusSeeOther)
	}

	return nil, errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) PopulateLoginMethodFirstFactorRefresh(r *http.Request, lf *login.Flow, _ *session.Session) error {
	conf, err := s.Config(r.Context())
	if err != nil {
		return err
	}

	var providers []Configuration
	_, id, c := flowhelpers.GuessForcedLoginIdentifier(r, s.d, lf, s.ID())
	if id == nil || c == nil {
		providers = nil
	} else {
		var credentials identity.CredentialsOIDC
		if err := json.Unmarshal(c.Config, &credentials); err != nil {
			// failed to read OIDC credentials, don't add any providers
			providers = nil
		} else {
			// add only providers that can actually be used to log in as this identity
			providers = make([]Configuration, 0, len(conf.Providers))
			for i := range conf.Providers {
				for j := range credentials.Providers {
					if conf.Providers[i].ID == credentials.Providers[j].Provider {
						providers = append(providers, conf.Providers[i])
						break
					}
				}
			}
		}
	}

	lf.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	AddProviders(lf.UI, providers, text.NewInfoLoginWith, s.ID())
	return nil
}

func (s *Strategy) PopulateLoginMethodFirstFactor(r *http.Request, f *login.Flow) error {
	return s.populateMethod(r, f, text.NewInfoLoginWith)
}

func (s *Strategy) removeProviders(conf *ConfigurationCollection, f *login.Flow) {
	for _, l := range conf.Providers {
		group := node.OpenIDConnectGroup
		if s.ID() == identity.CredentialsTypeSAML {
			group = node.SAMLGroup
		}

		if l.OrganizationID != "" {
			continue
		}

		f.GetUI().Nodes.RemoveMatching(&node.Node{
			Group: group,
			Type:  node.Input,
			Attributes: &node.InputAttributes{
				Name:       "provider",
				FieldValue: l.ID,
			},
		})
	}
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstCredentials(r *http.Request, f *login.Flow, mods ...login.FormHydratorModifier) (err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.oidc.Strategy.PopulateLoginMethodIdentifierFirstCredentials")
	defer otelx.End(span, &err)

	conf, err := s.Config(ctx)
	if err != nil {
		return err
	}

	o := login.NewFormHydratorOptions(mods)

	var linked []Provider
	if o.IdentityHint != nil {
		var err error
		// If we have an identity hint we check if the identity has any providers configured.
		if linked, err = s.linkedProviders(conf, o.IdentityHint); err != nil {
			return err
		}
	}

	if len(linked) == 0 {
		// If we found no credentials:
		if s.d.Config().SecurityAccountEnumerationMitigate(ctx) {
			// We found no credentials but do not want to leak that we know that. So we return early and do not
			// modify the initial provider list.
			return nil
		}

		if o.IdentityHint != nil {
			// We found no credentials. We remove all the providers and tell the strategy that we found nothing.
			// We only execute this, if the identity hint is set, otherwise we do not know if the user has any credentials and we likely stay on the `provide_credentials` screen.
			// The OIDC method is special in that regard, as it's the only method showing buttons on that screen.
			s.removeProviders(conf, f)
		}
		return idfirst.ErrNoCredentialsFound
	}

	if !s.d.Config().SecurityAccountEnumerationMitigate(ctx) {
		// Account enumeration is disabled, so we show all providers that are linked to the identity.
		// User is found and enumeration mitigation is disabled. Filter the list!
		s.removeProviders(conf, f)

		for _, l := range linked {
			lc := l.Config()

			// Organizations are handled differently.
			if lc.OrganizationID != "" {
				continue
			}
			AddProvider(f.UI, lc.ID, text.NewInfoLoginWith(cmp.Or(lc.Label, lc.ID), lc.ID), s.ID())
		}
	}

	return nil
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstIdentification(r *http.Request, f *login.Flow) error {
	return s.populateMethod(r, f, text.NewInfoLoginWith)
}
