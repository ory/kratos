package hydra

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	} else if conf.SelfServiceFlowHydraAdminURL(r.Context()) == nil {
		return uuid.NullUUID{}, errors.WithStack(herodot.ErrInternalServerError.WithReason("refusing to parse login_challenge query parameter because " + config.ViperKeySelfServiceHydraAdminURL + " is invalid or unset"))
	}

	hlc, err := uuid.FromString(r.URL.Query().Get("login_challenge"))
	if err != nil || hlc.IsNil() {
		return uuid.NullUUID{}, errors.WithStack(herodot.ErrBadRequest.WithReason("the login_challenge parameter is present but invalid or zero UUID"))
	} else {
		return uuid.NullUUID{UUID: hlc, Valid: !hlc.IsNil()}, nil
	}
}

func (h *DefaultHydra) getAdminURL(ctx context.Context) (string, error) {
	u := h.d.Config().SelfServiceFlowHydraAdminURL(ctx)
	if u == nil {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason(config.ViperKeySelfServiceHydraAdminURL + " is not configured"))
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
	configuration.HTTPClient = h.d.HTTPClient(ctx).StandardClient()

	return hydraclientgo.NewAPIClient(configuration).AdminApi, nil
}

func (h *DefaultHydra) AcceptLoginRequest(ctx context.Context, hlc uuid.UUID, sub string, amr session.AuthenticationMethods) (string, error) {
	remember := h.d.Config().SessionPersistentCookie(ctx)
	remember_for := int64(h.d.Config().SessionLifespan(ctx) / time.Second)

	alr := hydraclientgo.NewAcceptLoginRequest(sub)
	alr.Remember = &remember
	alr.RememberFor = &remember_for
	alr.Amr = []string{}
	for _, r := range amr {
		alr.Amr = append(alr.Amr, string(r.Method))
	}

	aa, err := h.getAdminAPIClient(ctx)
	if err != nil {
		return "", err
	}

	resp, r, err := aa.AcceptLoginRequest(context.Background()).LoginChallenge(fmt.Sprintf("%x", hlc)).AcceptLoginRequest(*alr).Execute()
	if err != nil {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithError(err.Error()).WithDetail("status_code", r.StatusCode))
	}

	if resp == nil {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("AcceptLoginRequest produced an empty response with no error").WithDetail("status_code", r.StatusCode))
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

	hlr, _, err := aa.GetLoginRequest(context.Background()).LoginChallenge(fmt.Sprintf("%x", hlc.UUID)).Execute()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("failed to retrieve login request from Hydra"))
	} else if hlr == nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Hydra returned an empty login request"))
	}

	return hlr, nil
}
