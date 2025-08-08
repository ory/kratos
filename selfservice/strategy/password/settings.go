// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/ory/x/otelx"

	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/text"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/session"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

func (s *Strategy) RegisterSettingsRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) SettingsStrategyID() string {
	return identity.CredentialsTypePassword.String()
}

// Update Settings Flow with Password Method
//
// swagger:model updateSettingsFlowWithPasswordMethod
type updateSettingsFlowWithPasswordMethod struct {
	// Password is the updated password
	//
	// required: true
	Password string `json:"password"`

	// CSRFToken is the anti-CSRF token
	CSRFToken string `json:"csrf_token"`

	// Method
	//
	// Should be set to password when trying to update a password.
	//
	// required: true
	Method string `json:"method"`

	// Flow is flow ID.
	//
	// swagger:ignore
	Flow string `json:"flow"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (p *updateSettingsFlowWithPasswordMethod) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.Flow)
}

func (p *updateSettingsFlowWithPasswordMethod) SetFlowID(rid uuid.UUID) {
	p.Flow = rid.String()
}

func (s *Strategy) Settings(ctx context.Context, w http.ResponseWriter, r *http.Request, f *settings.Flow, ss *session.Session) (_ *settings.UpdateContext, err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.password.Strategy.Settings")
	defer otelx.End(span, &err)

	var p updateSettingsFlowWithPasswordMethod
	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, f, ss, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		return ctxUpdate, s.continueSettingsFlow(ctx, r, ctxUpdate, p)
	} else if err != nil {
		return ctxUpdate, s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.SettingsStrategyID(), s.d); err != nil {
		return ctxUpdate, s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	if err := s.decodeSettingsFlow(r, &p); err != nil {
		return ctxUpdate, s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	// This does not come from the payload!
	p.Flow = ctxUpdate.Flow.ID.String()
	if err := s.continueSettingsFlow(ctx, r, ctxUpdate, p); err != nil {
		return ctxUpdate, s.handleSettingsError(ctx, w, r, ctxUpdate, p, err)
	}

	return ctxUpdate, nil
}

func (s *Strategy) decodeSettingsFlow(r *http.Request, dest interface{}) error {
	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(settingsSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	return decoderx.NewHTTP().Decode(r, dest, compiler,
		decoderx.HTTPKeepRequestBody(true),
		decoderx.HTTPDecoderAllowedMethods("POST", "GET"),
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	)
}

// Try to find a password hash in the credentials. Returns it if found, otherwise return an empty string.
func getPasswordHashFromCredential(creds map[identity.CredentialsType]identity.Credentials) string {
	if creds == nil {
		return ""
	}

	cred, ok := creds[identity.CredentialsTypePassword]
	if !ok {
		return ""
	}

	var hashedPassword identity.CredentialsPassword
	if err := json.Unmarshal(cred.Config, &hashedPassword); err != nil {
		return ""
	}
	return hashedPassword.HashedPassword
}

// Detect whether the new password is the same as the old password.
// This is helpful to a user, e.g. in the case of a password leak: they want to change their password,
// and unknowingly set the new password to be the same as the old one (that leaked). We force them to
// set a different password in that case.
func isNewPasswordSameAsOld(ctx context.Context, oldHashedPassword string, newPassword string) bool {
	if oldHashedPassword == "" {
		return false
	}

	// `hash.Compare` returns `nil` on 'success' i.e. old and new are the same.
	return hash.Compare(ctx, []byte(newPassword), []byte(oldHashedPassword)) == nil
}

func (s *Strategy) continueSettingsFlow(ctx context.Context, r *http.Request, ctxUpdate *settings.UpdateContext, p updateSettingsFlowWithPasswordMethod) error {
	if err := flow.MethodEnabledAndAllowed(ctx, flow.SettingsFlow, s.SettingsStrategyID(), p.Method, s.d); err != nil {
		return err
	}

	if err := flow.EnsureCSRF(s.d, r, ctxUpdate.Flow.Type, s.d.Config().DisableAPIFlowEnforcement(ctx), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return err
	}

	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)).Before(time.Now()) {
		return errors.WithStack(settings.NewFlowNeedsReAuth())
	}

	if len(p.Password) == 0 {
		return schema.NewRequiredError("#/password", "password")
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(ctx, ctxUpdate.Session.Identity.ID)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	var newPasswordHash []byte
	// Extract an immutable value to avoid data races between goroutines.
	oldHashedPassword := getPasswordHashFromCredential(i.Credentials)

	// Do in parallel due to limitations of the `bcrypt` library and for performance:
	// - `hash(newPassword)` (expensive).
	// - Check that the new password is not the same as the old password,
	//   which internally computes `hash(newPassword)` (expensive).
	// - `validateCredentials` which may call the HaveIBeenPawned external API.
	g.Go(func() error {
		var err error
		newPasswordHash, err = s.d.Hasher(ctx).Generate(ctx, []byte(p.Password))
		return err
	})
	g.Go(func() error {
		if isNewPasswordSameAsOld(ctx, oldHashedPassword, p.Password) {
			return schema.NewPasswordPolicyViolationError("#/password", text.NewErrorValidationPasswordNewSameAsOld())
		}
		return nil
	})
	g.Go(func() error {
		// Note: this goroutine mutates `i` so careful not to share it with other goroutines!

		// The credentials could have been modified in many ways possible. To keep it simple, we reset, and the validators
		// will populate it correctly.
		i.UpsertCredentialsConfig(s.ID(), []byte("{}"), 0)
		return s.validateCredentials(ctx, i, p.Password)
	})
	if err := g.Wait(); err != nil {
		return err
	}

	co, err := json.Marshal(&identity.CredentialsPassword{HashedPassword: string(newPasswordHash)})
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode password options to JSON: %s", err))
	}
	i.UpsertCredentialsConfig(s.ID(), co, 0)
	ctxUpdate.UpdateIdentity(i)

	return nil
}

func (s *Strategy) PopulateSettingsMethod(ctx context.Context, r *http.Request, _ *identity.Identity, f *settings.Flow) (err error) {
	_, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.password.Strategy.PopulateSettingsMethod")
	defer otelx.End(span, &err)

	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.Nodes.Upsert(NewPasswordNode("password", node.InputAttributeAutocompleteNewPassword).WithMetaLabel(text.NewInfoNodeInputPassword()))
	f.UI.Nodes.Append(node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSave()))

	return nil
}

func (s *Strategy) handleSettingsError(ctx context.Context, w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p updateSettingsFlowWithPasswordMethod, err error) error {
	// Do not pause flow if the flow type is an API flow as we can't save cookies in those flows.
	if e := new(settings.FlowNeedsReAuth); errors.As(err, &e) && ctxUpdate.Flow != nil && ctxUpdate.Flow.Type == flow.TypeBrowser {
		if err := s.d.ContinuityManager().Pause(ctx, w, r, settings.ContinuityKey(s.SettingsStrategyID()), settings.ContinuityOptions(p, ctxUpdate.GetSessionIdentity())...); err != nil {
			return err
		}
	}

	if ctxUpdate.Flow != nil {
		ctxUpdate.Flow.UI.ResetMessages()
		ctxUpdate.Flow.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	}

	return err
}
