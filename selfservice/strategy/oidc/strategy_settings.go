package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ory/kratos/session"

	"github.com/ory/kratos/ui/container"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"

	"github.com/ory/kratos/x"
)

const (
	SettingsPath = RouteBase + "/settings/connections"
)

var _ settings.Strategy = new(Strategy)
var UnknownConnectionValidationError = &jsonschema.ValidationError{
	Message: "can not unlink non-existing OpenID Connect connection", InstancePtr: "#/"}
var ConnectionExistValidationError = &jsonschema.ValidationError{
	Message: "can not link unknown or already existing OpenID Connect connection", InstancePtr: "#/"}

func (s *Strategy) RegisterSettingsRoutes(router *x.RouterPublic) {}

func (s *Strategy) SettingsStrategyID() string {
	return s.ID().String()
}

func (s *Strategy) linkedProviders(ctx context.Context, r *http.Request, conf *ConfigurationCollection, confidential *identity.Identity) ([]Provider, error) {
	creds, ok := confidential.GetCredentials(s.ID())
	if !ok {
		return nil, nil
	}

	var available CredentialsConfig
	if err := json.Unmarshal(creds.Config, &available); err != nil {
		return nil, errors.WithStack(err)
	}

	var count int
	for _, strategy := range s.d.ActiveCredentialsCounterStrategies(ctx) {
		current, err := strategy.CountActiveCredentials(confidential.Credentials)
		if err != nil {
			return nil, err
		}

		count += current
		if count > 1 {
			break
		}
	}

	if count < 2 {
		// This means that we're able to remove a connection because it is the last configured credential. If it is
		// removed, the identity is no longer able to sign in.
		return nil, nil
	}

	var result []Provider
	for _, p := range available.Providers {
		prov, err := conf.Provider(p.Provider, s.d.Config(ctx).SelfPublicURL(r))
		if err != nil {
			return nil, err
		}
		result = append(result, prov)
	}

	return result, nil
}

func (s *Strategy) linkableProviders(ctx context.Context, r *http.Request, conf *ConfigurationCollection, confidential *identity.Identity) ([]Provider, error) {
	var available CredentialsConfig
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
			prov, err := conf.Provider(p.ID, s.d.Config(ctx).SelfPublicURL(r))
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

	f := container.New(urlx.CopyWithQuery(urlx.AppendPaths(
		s.d.Config(r.Context()).SelfPublicURL(r), SettingsPath), url.Values{"flow": {sr.ID.String()}}).String())
	f.SetCSRF(s.d.GenerateCSRFToken(r))

	for _, l := range linkable {
		f.GetNodes().Append(NewLinkNode(l.Config().ID))
	}

	for _, l := range linked {
		f.GetNodes().Append(NewUnlinkNode(l.Config().ID))
	}

	panic(`
	sr.Methods[s.SettingsStrategyID()] = &settings.FlowMethod{
		Method: s.SettingsStrategyID(),
		Config: &settings.FlowMethodConfig{FlowMethodConfigurator: NewFlowMethod(f)},
	}`)

	return nil
}

// swagger:model completeSelfServiceBrowserSettingsOIDCFlowPayload
type completeSelfServiceBrowserSettingsOIDCFlowPayload struct {
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
}

func (p *completeSelfServiceBrowserSettingsOIDCFlowPayload) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.FlowID)
}

func (p *completeSelfServiceBrowserSettingsOIDCFlowPayload) SetFlowID(rid uuid.UUID) {
	p.FlowID = rid.String()
}

func (s *Strategy) Settings(w http.ResponseWriter, r *http.Request, f *settings.Flow, ss *session.Session) (*settings.UpdateContext, error) {
	if err := flow.MethodEnabledAndAllowedFromRequest(r, s.SettingsStrategyID(), s.d); err != nil {
		return nil, err
	}

	var p completeSelfServiceBrowserSettingsOIDCFlowPayload
	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, f, ss, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		if l := len(p.Link); l > 0 {
			panic("s.initLinkProvider(w, r, ctxUpdate, &p)")
			return nil, nil
		} else if u := len(p.Unlink); u > 0 {
			panic("s.unlinkProvider(w, r, ctxUpdate, &p)")
			return nil, nil
		}

		return nil, s.handleSettingsError(w, r, ctxUpdate, &p, errors.WithStack(herodot.ErrInternalServerError.WithReason("Expected either link or unlink to be set when continuing flow but both are unset.")))
	} else if err != nil {
		return nil, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	if err := r.ParseForm(); err != nil {
		return nil, s.handleSettingsError(w, r, ctxUpdate, &p, err)
	}

	p.Link = r.Form.Get("link")
	p.Unlink = r.Form.Get("unlink")
	if l, u := len(p.Link), len(p.Unlink); l > 0 && u > 0 {
		return nil, s.handleSettingsError(w, r, ctxUpdate, &p, errors.WithStack(&jsonschema.ValidationError{
			Message:     "it is not possible to link and unlink providers in the same request",
			InstancePtr: "#/",
		}))
	} else if l > 0 {
		panic("s.initLinkProvider(w, r, ctxUpdate, &p)")
		return nil, nil
	} else if u > 0 {
		panic("s.unlinkProvider(w, r, ctxUpdate, &p)")
		return nil, nil
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

func (s *Strategy) initLinkProvider(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext,
	p *completeSelfServiceBrowserSettingsOIDCFlowPayload) {
	if _, err := s.isLinkable(r, ctxUpdate, p.Link); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config(r.Context()).SelfServiceFlowSettingsPrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
		return
	}

	http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(r),
		strings.Replace(RouteAuth, ":flow", p.FlowID, 1)),
		url.Values{"provider": {p.Link}}).String(), http.StatusFound)
}

func (s *Strategy) linkProvider(w http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, claims *Claims, provider Provider) {
	p := &completeSelfServiceBrowserSettingsOIDCFlowPayload{
		Link: provider.Config().ID, FlowID: ctxUpdate.Flow.ID.String()}
	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config(r.Context()).SelfServiceFlowSettingsPrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
		return
	}

	i, err := s.isLinkable(r, ctxUpdate, p.Link)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	var conf CredentialsConfig
	creds, err := i.ParseCredentials(s.ID(), &conf)
	if errors.Is(err, herodot.ErrNotFound) {
		var err error
		if creds, err = NewCredentials(provider.Config().ID, claims.Subject); err != nil {
			s.handleSettingsError(w, r, ctxUpdate, p, err)
			return
		}
	} else if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	} else {
		creds.Identifiers = append(creds.Identifiers, uid(provider.Config().ID, claims.Subject))
		conf.Providers = append(conf.Providers, ProviderCredentialsConfig{
			Subject: claims.Subject, Provider: provider.Config().ID})
		creds.Config, err = json.Marshal(conf)
		if err != nil {
			s.handleSettingsError(w, r, ctxUpdate, p, err)
			return
		}
	}

	i.Credentials[s.ID()] = *creds
	if err := s.d.SettingsHookExecutor().PostSettingsHook(w, r, s.SettingsStrategyID(), ctxUpdate, i, settings.WithCallback(func(ctxUpdate *settings.UpdateContext) error {
		return s.PopulateSettingsMethod(r, ctxUpdate.Session.Identity, ctxUpdate.Flow)
	})); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}
}

func (s *Strategy) unlinkProvider(w http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, p *completeSelfServiceBrowserSettingsOIDCFlowPayload) {
	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config(r.Context()).SelfServiceFlowSettingsPrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
		return
	}

	providers, err := s.Config(r.Context())
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	availableProviders, err := s.linkedProviders(r.Context(), r, providers, i)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	var cc CredentialsConfig
	creds, err := i.ParseCredentials(s.ID(), &cc)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(UnknownConnectionValidationError))
		return
	}

	var found bool
	var updatedProviders []ProviderCredentialsConfig
	var updatedIdentifiers []string
	for _, available := range availableProviders {
		if p.Unlink == available.Config().ID {
			for _, link := range cc.Providers {
				if link.Provider != p.Unlink {
					updatedIdentifiers = append(updatedIdentifiers, uid(link.Provider, link.Subject))
					updatedProviders = append(updatedProviders, link)
				} else {
					found = true
				}
			}
		}
	}

	if !found {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(UnknownConnectionValidationError))
		return
	}

	creds.Identifiers = updatedIdentifiers
	creds.Config, err = json.Marshal(&CredentialsConfig{updatedProviders})
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(err))
		return
	}

	i.Credentials[s.ID()] = *creds
	if err := s.d.SettingsHookExecutor().PostSettingsHook(w, r, s.SettingsStrategyID(), ctxUpdate, i, settings.WithCallback(func(ctxUpdate *settings.UpdateContext) error {
		return s.PopulateSettingsMethod(r, ctxUpdate.Session.Identity, ctxUpdate.Flow)
	})); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *completeSelfServiceBrowserSettingsOIDCFlowPayload, err error) error {
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
