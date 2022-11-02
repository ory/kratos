package hydra

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ory/x/httpx"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	hydraclientgo "github.com/ory/hydra-client-go"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	hydraDependencies interface {
		config.Provider
		x.HTTPClientProvider
	}
	HydraProvider interface {
		Hydra() Hydra
	}
	Hydra interface {
		AcceptLoginRequest(ctx context.Context, hlc uuid.UUID, sub string, amr session.AuthenticationMethods) (string, error)
		GetLoginRequest(ctx context.Context, hlc uuid.NullUUID) (*hydraclientgo.LoginRequest, error)
	}
	DefaultHydra struct {
		d hydraDependencies
	}
)

func NewDefaultHydra(d hydraDependencies) *DefaultHydra {
	return &DefaultHydra{
		d: d,
	}
}

func GetLoginChallengeID(conf *config.Config, r *http.Request) (uuid.NullUUID, error) {
	if !r.URL.Query().Has("login_challenge") {
		return uuid.NullUUID{}, nil
	} else if conf.OAuth2ProviderURL(r.Context()) == nil {
		return uuid.NullUUID{}, errors.WithStack(herodot.ErrInternalServerError.WithReason("refusing to parse login_challenge query parameter because " + config.ViperKeyOAuth2ProviderURL + " is invalid or unset"))
	}

	hlc, err := uuid.FromString(r.URL.Query().Get("login_challenge"))
	if err != nil || hlc.IsNil() {
		return uuid.NullUUID{}, errors.WithStack(herodot.ErrBadRequest.WithReason("the login_challenge parameter is present but invalid or zero UUID"))
	} else {
		return uuid.NullUUID{UUID: hlc, Valid: true}, nil
	}
}

func (h *DefaultHydra) getAdminURL(ctx context.Context) (string, error) {
	u := h.d.Config().OAuth2ProviderURL(ctx)
	if u == nil {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason(config.ViperKeyOAuth2ProviderURL + " is not configured"))
	}
	return u.String(), nil
}

func (h *DefaultHydra) getAdminAPIClient(ctx context.Context) (hydraclientgo.AdminApi, error) {
	url, err := h.getAdminURL(ctx)
	if err != nil {
		return nil, err
	}

	configuration := hydraclientgo.NewConfiguration()
	configuration.Servers = hydraclientgo.ServerConfigurations{{URL: url}}

	client := h.d.HTTPClient(ctx).StandardClient()
	if header := h.d.Config().OAuth2ProviderHeader(ctx); header != nil {
		client.Transport = httpx.WrapTransportWithHeader(client.Transport, header)
	}

	configuration.HTTPClient = client
	return hydraclientgo.NewAPIClient(configuration).AdminApi, nil
}

func (h *DefaultHydra) AcceptLoginRequest(ctx context.Context, hlc uuid.UUID, sub string, amr session.AuthenticationMethods) (string, error) {
	remember := h.d.Config().SessionPersistentCookie(ctx)
	rememberFor := int64(h.d.Config().SessionLifespan(ctx) / time.Second)

	alr := hydraclientgo.NewAcceptLoginRequest(sub)
	alr.Remember = &remember
	alr.RememberFor = &rememberFor
	alr.Amr = []string{}
	for _, r := range amr {
		alr.Amr = append(alr.Amr, string(r.Method))
	}

	aa, err := h.getAdminAPIClient(ctx)
	if err != nil {
		return "", err
	}

	resp, r, err := aa.AcceptLoginRequest(ctx).LoginChallenge(fmt.Sprintf("%x", hlc)).AcceptLoginRequest(*alr).Execute()
	if err != nil {
		innerErr := herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to accept OAuth 2.0 Login Challenge.")
		if r != nil {
			innerErr = innerErr.
				WithDetail("status_code", r.StatusCode).
				WithDebug(err.Error())
		}
		return "", errors.WithStack(innerErr)
	}

	return resp.RedirectTo, nil
}

func (h *DefaultHydra) GetLoginRequest(ctx context.Context, hlc uuid.NullUUID) (*hydraclientgo.LoginRequest, error) {
	if !hlc.Valid {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReason("invalid login_challenge"))
	}

	aa, err := h.getAdminAPIClient(ctx)
	if err != nil {
		return nil, err
	}

	hlr, r, err := aa.GetLoginRequest(ctx).LoginChallenge(fmt.Sprintf("%x", hlc.UUID)).Execute()
	if err != nil {
		innerErr := herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to get OAuth 2.0 Login Challenge.")
		if r != nil {
			innerErr = innerErr.
				WithDetail("status_code", r.StatusCode).
				WithDebug(err.Error())
		}
		return nil, errors.WithStack(innerErr)
	}

	return hlr, nil
}
