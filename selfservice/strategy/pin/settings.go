package pin

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

func (s *Strategy) RegisterSettingsRoutes(_ *x.RouterPublic) {
}

func (s *Strategy) SettingsStrategyID() string {
	return identity.CredentialsTypePin.String()
}

// swagger:model submitSelfServiceSettingsFlowWithPinMethodBody
type submitSelfServiceSettingsFlowWithPinMethodBody struct {
	// Pin is the updated pin code
	//
	// required: true
	Pin string `json:"pin"`

	// CSRFToken is the anti-CSRF token
	CSRFToken string `json:"csrf_token"`

	// Method
	//
	// Should be set to pin when trying to update a pin.
	//
	// required: true
	Method string `json:"method"`

	// Flow is flow ID.
	//
	// swagger:ignore
	Flow string `json:"flow"`
}

func (p *submitSelfServiceSettingsFlowWithPinMethodBody) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.Flow)
}

func (p *submitSelfServiceSettingsFlowWithPinMethodBody) SetFlowID(rid uuid.UUID) {
	p.Flow = rid.String()
}

func (s *Strategy) Settings(w http.ResponseWriter, r *http.Request, f *settings.Flow, ss *session.Session) (*settings.UpdateContext, error) {
	var p submitSelfServiceSettingsFlowWithPinMethodBody

	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, f, ss, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		return ctxUpdate, s.continueSettingsFlow(w, r, ctxUpdate, &p)
	} else if err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if err := flow.MethodEnabledAndAllowedFromRequest(r, s.SettingsStrategyID(), s.d); err != nil {
		return ctxUpdate, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if err := s.decodeSettingsFlow(r, &p); err != nil {
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
	_ http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, p *submitSelfServiceSettingsFlowWithPinMethodBody,
) error {
	if err := flow.MethodEnabledAndAllowed(r.Context(), s.SettingsStrategyID(), p.Method, s.d); err != nil {
		return err
	}

	if err := flow.EnsureCSRF(s.d, r, ctxUpdate.Flow.Type, s.d.Config().DisableAPIFlowEnforcement(r.Context()), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		return err
	}

	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context())).Before(time.Now()) {
		return errors.WithStack(settings.NewFlowNeedsReAuth())
	}

	if len(p.Pin) == 0 {
		return schema.NewRequiredError("#/password", "password")
	}

	hpin, err := s.d.Hasher(r.Context()).Generate(r.Context(), []byte(p.Pin))
	if err != nil {
		return err
	}

	co, err := json.Marshal(&identity.CredentialsPin{HashedPin: string(hpin)})
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode pin options to JSON: %s", err))
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		return err
	}

	i.UpsertCredentialsConfig(s.ID(), co, 0)
	ctxUpdate.UpdateIdentity(i)

	return nil
}

func (s *Strategy) PopulateSettingsMethod(r *http.Request, _ *identity.Identity, f *settings.Flow) error {
	f.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	f.UI.Nodes.Upsert(NewPinNode("pin").WithMetaLabel(text.NewInfoNodeInputPin()))
	f.UI.Nodes.Append(node.NewInputField("method", "pin", node.PinGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoNodeLabelSave()))

	return nil
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *submitSelfServiceSettingsFlowWithPinMethodBody, err error) error {
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
