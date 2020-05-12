package oidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gobuffalo/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	SettingsPath = BasePath + "/settings/connections"

	continuityPrefix = "ory_kratos_settings_oidc"
)

var _ settings.Strategy = new(Strategy)
var UnknownConnectionValidationError = &jsonschema.ValidationError{
	Message: fmt.Sprintf("can not unlink non-existing OpenID Connect connection"), InstancePtr: "#/unlink"}
var ConnectionExistValidationError = &jsonschema.ValidationError{
	Message: fmt.Sprintf("can not override existing OpenID Connect connection"), InstancePtr: "#/link"}

func (s *Strategy) RegisterSettingsRoutes(router *x.RouterPublic) {
	router.POST(SettingsPath, s.completeSettingsFlow)
}

func (s *Strategy) SettingsStrategyID() string {
	return s.ID().String()
}

func (s *Strategy) linkedProviders(conf *ConfigurationCollection, confidential *identity.Identity) ([]Provider, error) {
	creds, ok := confidential.GetCredentials(s.ID())
	if !ok {
		return nil, nil
	}

	var available CredentialsConfig
	if err := json.Unmarshal(creds.Config, &available); err != nil {
		return nil, errors.WithStack(err)
	}

	var count int
	for _, strategy := range s.d.ActiveCredentialsCounterStrategies() {
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
		prov, err := conf.Provider(p.Provider, s.c.SelfPublicURL())
		if err != nil {
			return nil, err
		}
		result = append(result, prov)
	}

	return result, nil
}

func (s *Strategy) linkableProviders(conf *ConfigurationCollection, confidential *identity.Identity) ([]Provider, error) {
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
			prov, err := conf.Provider(p.ID, s.c.SelfPublicURL())
			if err != nil {
				return nil, err
			}
			result = append(result, prov)
		}
	}

	return result, nil
}

func (s *Strategy) PopulateSettingsMethod(r *http.Request, ss *session.Session, sr *settings.Request) error {
	conf, err := s.Config()
	if err != nil {
		return err
	}

	confidential, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ss.IdentityID)
	if err != nil {
		return err
	}

	linkable, err := s.linkableProviders(conf, confidential)
	if err != nil {
		return err
	}

	linked, err := s.linkedProviders(conf, confidential)
	if err != nil {
		return err
	}

	f := form.NewHTMLForm(urlx.CopyWithQuery(urlx.AppendPaths(
		s.c.SelfPublicURL(), SettingsPath), url.Values{"request": {sr.ID.String()}}).String())
	f.SetCSRF(s.d.GenerateCSRFToken(r))

	for _, l := range linkable {
		f.Fields = append(f.Fields, form.Field{
			Name:  "link",
			Type:  "submit",
			Value: l.Config().ID,
		})
	}

	for _, l := range linked {
		f.Fields = append(f.Fields, form.Field{
			Name:  "unlink",
			Type:  "submit",
			Value: l.Config().ID,
		})
	}

	sr.Methods[s.SettingsStrategyID()] = &settings.RequestMethod{
		Method: s.SettingsStrategyID(),
		Config: &settings.RequestMethodConfig{RequestMethodConfigurator: NewRequestMethodConfig(f)},
	}

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

	// RequestID is request ID.
	//
	// in: query
	RequestID string `json:"request_id"`
}

func (p *completeSelfServiceBrowserSettingsOIDCFlowPayload) GetRequestID() uuid.UUID {
	return x.ParseUUID(p.RequestID)
}

func (p *completeSelfServiceBrowserSettingsOIDCFlowPayload) SetRequestID(rid uuid.UUID) {
	p.RequestID = rid.String()
}

// swagger:route POST /self-service/browser/flows/registration/strategies/oidc/settings/connections public completeSelfServiceBrowserSettingsOIDCSettingsFlow
//
// Complete the browser-based settings flow for the OpenID Connect strategy
//
// This endpoint completes a browser-based settings flow. This is usually achieved by POSTing data to this
// endpoint.
//
// > This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...) and HTML Forms.
//
// More information can be found at [ORY Kratos User Settings & Profile Management Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-settings-profile-management).
//
//     Consumes:
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (s *Strategy) completeSettingsFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p completeSelfServiceBrowserSettingsOIDCFlowPayload
	ctxUpdate, err := settings.PrepareUpdate(s.d, r, continuityPrefix, &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		if l := len(p.Link); l > 0 {
			http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.c.SelfPublicURL(),
				strings.Replace(AuthPath, ":request", p.RequestID, 1)),
				url.Values{"provider": {p.Link}}).String(), http.StatusFound)
			return
		} else if u := len(p.Unlink); u > 0 {
			s.unlinkProvider(w, r, ctxUpdate, &p)
			return
		}

		s.handleSettingsError(w, r, ctxUpdate, &p, errors.WithStack(herodot.ErrInternalServerError.WithReason("Expected either link or unlink to be set when continuing flow but both are unset.")))
		return
	} else if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, &p, err)
		return
	}

	if l := len(p.Link); l > 0 {
		http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.c.SelfPublicURL(),
			strings.Replace(AuthPath, ":request", p.RequestID, 1)),
			url.Values{"provider": {p.Link}}).String(), http.StatusFound)
		return
	} else if u := len(p.Unlink); u > 0 {
		s.unlinkProvider(w, r, ctxUpdate, &p)
		return
	} else if l+u == 0 {
		s.handleSettingsError(w, r, ctxUpdate, &p, errors.WithStack(&jsonschema.ValidationError{
			Message:     fmt.Sprintf("it is not possible to link and unlink providers in the same request"),
			InstancePtr: "#/unlink",
		}))
		return
	}

	s.handleSettingsError(w, r, ctxUpdate, &p, errors.WithStack(errors.WithStack(&jsonschema.ValidationError{
		Message: fmt.Sprintf("missing properties: link, unlink"), InstancePtr: "#/",
		Context: &jsonschema.ValidationErrorContextRequired{Missing: []string{"link", "unlink"}}})))
	return
}

func (s *Strategy) linkProvider(w http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, claims *Claims, provider Provider) {
	p := &completeSelfServiceBrowserSettingsOIDCFlowPayload{
		Link: provider.Config().ID, RequestID: ctxUpdate.Request.ID.String()}
	if ctxUpdate.Session.AuthenticatedAt.Add(s.c.SelfServicePrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.ErrRequestNeedsReAuthentication))
		return
	}

	providers, err := s.Config()
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	linkable, err := s.linkableProviders(providers, i)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	var found bool
	for _, available := range linkable {
		if p.Link == available.Config().ID {
			found = true
		}
	}

	if !found {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(ConnectionExistValidationError))
		return
	}

	var conf CredentialsConfig
	creds, err := i.ParseCredentials(s.ID(), conf)
	if errors.Is(err, herodot.ErrNotFound) {
		var err error
		if creds, err = NewCredentials(provider.Config().ID, claims.Subject); err != nil {
			s.handleSettingsError(w, r, ctxUpdate, p, err)
			return
		}
	} else if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	creds.Identifiers = append(creds.Identifiers, uid(provider.Config().ID, claims.Subject))
	conf.Providers = append(conf.Providers, ProviderCredentialsConfig{
		Subject:  claims.Subject,
		Provider: provider.Config().ID,
	})

	creds.Config, err = json.Marshal(conf)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	i.Credentials[s.ID()] = *creds
	if err := s.d.SettingsHookExecutor().PostSettingsHook(w, r, s.SettingsStrategyID(), ctxUpdate, i); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}
}

func (s *Strategy) unlinkProvider(w http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, p *completeSelfServiceBrowserSettingsOIDCFlowPayload) {
	if ctxUpdate.Session.AuthenticatedAt.Add(s.c.SelfServicePrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.ErrRequestNeedsReAuthentication))
		return
	}

	providers, err := s.Config()
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	availableProviders, err := s.linkedProviders(providers, i)
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
	if err := s.d.SettingsHookExecutor().PostSettingsHook(w, r, s.SettingsStrategyID(), ctxUpdate, i); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *completeSelfServiceBrowserSettingsOIDCFlowPayload, err error) {
	if errors.Is(err, settings.ErrRequestNeedsReAuthentication) {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r,
			continuityPrefix+"."+r.URL.Query().Get("request"), settings.ContinuityOptions(p, ctxUpdate.Session.Identity)...); err != nil {
			s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, ctxUpdate.Request, err, s.SettingsStrategyID())
			return
		}
	}

	if ctxUpdate.Request != nil {
		ctxUpdate.Request.Methods[s.SettingsStrategyID()].Config.ResetErrors()
		ctxUpdate.Request.Methods[s.SettingsStrategyID()].Config.SetCSRF(s.d.GenerateCSRFToken(r))
	}

	s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, ctxUpdate.Request, err, s.SettingsStrategyID())
}
