// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hydra

import (
	"context"
	"net/http"
	"time"

	"github.com/ory/x/httpx"
	"github.com/ory/x/sqlxx"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	hydraclientgo "github.com/ory/hydra-client-go/v2"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	hydraDependencies interface {
		config.Provider
		x.HTTPClientProvider
	}
	Provider interface {
		Hydra() Hydra
	}
	Hydra interface {
		AcceptLoginRequest(ctx context.Context, loginChallenge string, subject string, amr session.AuthenticationMethods) (string, error)
		GetLoginRequest(ctx context.Context, loginChallenge string) (*hydraclientgo.OAuth2LoginRequest, error)
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

func GetLoginChallengeID(conf *config.Config, r *http.Request) (sqlxx.NullString, error) {
	if !r.URL.Query().Has("login_challenge") {
		return "", nil
	} else if conf.OAuth2ProviderURL(r.Context()) == nil {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason("refusing to parse login_challenge query parameter because " + config.ViperKeyOAuth2ProviderURL + " is invalid or unset"))
	}

	loginChallenge := r.URL.Query().Get("login_challenge")
	if loginChallenge == "" {
		return "", errors.WithStack(herodot.ErrBadRequest.WithReason("the login_challenge parameter is present but empty"))
	}

	return sqlxx.NullString(loginChallenge), nil
}

func (h *DefaultHydra) getAdminURL(ctx context.Context) (string, error) {
	u := h.d.Config().OAuth2ProviderURL(ctx)
	if u == nil {
		return "", errors.WithStack(herodot.ErrInternalServerError.WithReason(config.ViperKeyOAuth2ProviderURL + " is not configured"))
	}
	return u.String(), nil
}

func (h *DefaultHydra) getAdminAPIClient(ctx context.Context) (hydraclientgo.OAuth2Api, error) {
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
	return hydraclientgo.NewAPIClient(configuration).OAuth2Api, nil
}

func (h *DefaultHydra) AcceptLoginRequest(ctx context.Context, loginChallenge string, subject string, amr session.AuthenticationMethods) (string, error) {
	remember := h.d.Config().SessionPersistentCookie(ctx)
	rememberFor := int64(h.d.Config().SessionLifespan(ctx) / time.Second)

	alr := hydraclientgo.NewAcceptOAuth2LoginRequest(subject)
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

	resp, r, err := aa.AcceptOAuth2LoginRequest(ctx).LoginChallenge(loginChallenge).AcceptOAuth2LoginRequest(*alr).Execute()
	if err != nil {
		innerErr := herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to accept OAuth 2.0 Login Challenge.")
		if r != nil {
			innerErr = innerErr.
				WithDetail("status_code", r.StatusCode).
				WithDebug(err.Error())
		}

		if openApiErr := new(hydraclientgo.GenericOpenAPIError); errors.As(err, &openApiErr) {
			switch oauth2Err := openApiErr.Model().(type) {
			case hydraclientgo.ErrorOAuth2:
				innerErr = innerErr.WithDetail("oauth2_error_hint", oauth2Err.GetErrorHint())
			case *hydraclientgo.ErrorOAuth2:
				innerErr = innerErr.WithDetail("oauth2_error_hint", oauth2Err.GetErrorHint())
			}
		}

		return "", errors.WithStack(innerErr)
	}

	return resp.RedirectTo, nil
}

func (h *DefaultHydra) GetLoginRequest(ctx context.Context, loginChallenge string) (*hydraclientgo.OAuth2LoginRequest, error) {
	if loginChallenge == "" {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReason("invalid login_challenge"))
	}

	aa, err := h.getAdminAPIClient(ctx)
	if err != nil {
		return nil, err
	}

	hlr, r, err := aa.GetOAuth2LoginRequest(ctx).LoginChallenge(loginChallenge).Execute()
	if err != nil {
		innerErr := herodot.ErrInternalServerError.WithWrap(err).WithReasonf("Unable to get OAuth 2.0 Login Challenge.")
		if r != nil {
			innerErr = innerErr.
				WithDetail("status_code", r.StatusCode).
				WithDebug(err.Error())
		}

		if openApiErr := new(hydraclientgo.GenericOpenAPIError); errors.As(err, &openApiErr) {
			switch oauth2Err := openApiErr.Model().(type) {
			case hydraclientgo.ErrorOAuth2:
				innerErr = innerErr.WithDetail("oauth2_error_hint", oauth2Err.GetErrorHint())
			case *hydraclientgo.ErrorOAuth2:
				innerErr = innerErr.WithDetail("oauth2_error_hint", oauth2Err.GetErrorHint())
			}
		}

		return nil, errors.WithStack(innerErr)
	}

	return hlr, nil
}
