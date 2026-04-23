// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/x/httpx"
	"github.com/ory/x/randx"

	"github.com/ory/herodot"
)

var _ OAuth2Provider = (*ProviderSber)(nil)

const sberAPITimeout = 10 * time.Second

type ProviderSber struct {
	config *Configuration
	reg    Dependencies
}

func NewProviderSber(config *Configuration, reg Dependencies) Provider {
	return &ProviderSber{
		config: config,
		reg:    reg,
	}
}

func (g *ProviderSber) Config() *Configuration {
	return g.config
}

func (g *ProviderSber) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("nonce", sberNonceFromRequest(g.config.ID, r)),
		oauth2.SetAuthURLParam("client_type", "PRIVATE"),
		oauth2.SetAuthURLParam("response_type", "code"),
	}
}

func (g *ProviderSber) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	oauthConfig := &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://id.sber.ru/CSAFront/oidc/sberbank_id/authorize.do",
			TokenURL: "https://oauth.sber.ru/ru/prod/tokens/v2/oidc",
		},
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}

	g.reg.Logger().
		WithField("oidc_provider", g.config.ID).
		WithField("oidc_stage", "provider_config").
		WithField("token_url", oauthConfig.Endpoint.TokenURL).
		Debug("OIDC provider config loaded")

	return oauthConfig, nil
}

func (g *ProviderSber) Claims(ctx context.Context, exchange *oauth2.Token, _ url.Values) (*Claims, error) {
	ctx, cancel := context.WithTimeout(ctx, sberAPITimeout)
	defer cancel()

	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, err
	}

	ctx, client := httpx.SetOAuth2(ctx, g.reg.HTTPClient(ctx), o, exchange)
	userinfoURL := sberUserinfoURL("sber", g.config)
	stageStart := time.Now()
	if oauthHTTPClient := client.HTTPClient; oauthHTTPClient != nil {
		mtlsTransport, mtlsCertPath, mtlsKeyPath, mtlsBaseType, mtlsErr := withSberMTLS(baseRoundTripper(oauthHTTPClient.Transport))
		oauthHTTPClient.Transport = mtlsTransport
		fields := g.reg.Logger().
			WithField("oidc_provider", g.config.ID).
			WithField("oidc_stage", "userinfo_claims").
			WithField("sber_debug_version", sberTokenDebugVersion).
			WithField("mtls_cert_path", mtlsCertPath).
			WithField("mtls_key_path", mtlsKeyPath).
			WithField("mtls_base_transport_type", mtlsBaseType)
		if mtlsErr != nil {
			fields.WithError(mtlsErr).Error("Failed to attach mTLS certificate for Sber userinfo request")
		} else {
			fields.Debug("Attached mTLS certificate for Sber userinfo request")
		}
	}

	req, err := retryablehttp.NewRequestWithContext(
		ctx,
		"GET",
		userinfoURL,
		nil,
	)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReasonf("%s", err))
	}

	var hexRunes = []rune("0123456789ABCDEF")
	requestID := randx.MustString(32, hexRunes)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", exchange.AccessToken))
	req.Header.Set("x-introspect-rquid", requestID)

	g.reg.Logger().
		WithField("oidc_provider", g.config.ID).
		WithField("oidc_stage", "userinfo_claims").
		WithField("sber_debug_version", sberTokenDebugVersion).
		WithField("request_id", requestID).
		WithField("userinfo_url", userinfoURL).
		WithField("access_token_present", exchange.AccessToken != "").
		WithField("access_token_len", len(exchange.AccessToken)).
		Debug("Starting OIDC userinfo request")

	resp, err := client.Do(req)
	if err != nil {
		g.reg.Logger().
			WithError(err).
			WithField("oidc_provider", g.config.ID).
			WithField("oidc_stage", "userinfo_claims").
			WithField("request_id", requestID).
			WithField("userinfo_url", userinfoURL).
			WithField("latency_ms", time.Since(stageStart).Milliseconds()).
			Debug("OIDC userinfo request failed")
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}
	defer func() { _ = resp.Body.Close() }()

	g.reg.Logger().
		WithField("oidc_provider", g.config.ID).
		WithField("oidc_stage", "userinfo_claims").
		WithField("request_id", requestID).
		WithField("userinfo_url", userinfoURL).
		WithField("http_status", resp.StatusCode).
		WithField("latency_ms", time.Since(stageStart).Milliseconds()).
		Debug("OIDC userinfo response received")

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if readErr != nil {
			g.reg.Logger().
				WithError(readErr).
				WithField("oidc_provider", g.config.ID).
				WithField("oidc_stage", "userinfo_claims").
				WithField("request_id", requestID).
				WithField("userinfo_url", userinfoURL).
				WithField("http_status", resp.StatusCode).
				WithField("latency_ms", time.Since(stageStart).Milliseconds()).
				Error("OIDC userinfo response read failed")
		}

		bodyFragment := safeBodyForLog(body, maxUpstreamBodyLogBytes)
		fields := g.reg.Logger().
			WithField("oidc_provider", g.config.ID).
			WithField("oidc_stage", "userinfo_claims").
			WithField("request_id", requestID).
			WithField("userinfo_url", userinfoURL).
			WithField("http_status", resp.StatusCode).
			WithField("latency_ms", time.Since(stageStart).Milliseconds())
		if bodyFragment != "" {
			fields = fields.WithField("response_body_fragment", bodyFragment)
		}
		fields.Error("OIDC userinfo failed with upstream response")

		return nil, errors.WithStack(
			herodot.ErrUpstreamError.
				WithReasonf("OpenID Connect provider returned a %d status code but 200 is expected. debug_version=%s stage=userinfo_claims provider=%s request_id=%s userinfo_url=%s access_token_present=%t access_token_len=%d response=%q",
					resp.StatusCode,
					sberTokenDebugVersion,
					g.config.ID,
					requestID,
					userinfoURL,
					exchange.AccessToken != "",
					len(exchange.AccessToken),
					bodyFragment,
				),
		)
	}

	var user struct {
		Sub         string `json:"sub"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
		GivenName   string `json:"given_name"`
		FamilyName  string `json:"family_name"`
		MiddleName  string `json:"middle_name"`
		BirthDate   string `json:"birthdate"`
		Gender      int    `json:"gender"`
		Picture     string `json:"picture"`
		AvatarURL   string `json:"avatar_url"`
		City        string `json:"city"`
		Address     string `json:"address"`
		School      string `json:"school"`
		University  string `json:"university"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}

	gender := ""
	switch user.Gender {
	case 1:
		gender = "female"
	case 2:
		gender = "male"
	}

	picture := user.Picture
	if picture == "" {
		picture = user.AvatarURL
	}

	claims := &Claims{
		Issuer:      "https://oauth.sber.ru/ru/prod/sberbankid/v2.1/userinfo",
		Subject:     user.Sub,
		GivenName:   normalizeNameTitle(user.GivenName),
		FamilyName:  normalizeNameTitle(user.FamilyName),
		LastName:    normalizeNameTitle(user.FamilyName),
		MiddleName:  normalizeNameTitle(user.MiddleName),
		Email:       normalizeEmailLower(user.Email),
		PhoneNumber: normalizeRussianMobilePlus79(user.PhoneNumber),
		Birthdate:   user.BirthDate,
		Gender:      gender,
		Picture:     picture,
		City:        user.City,
		Address:     user.Address,
		School:      user.School,
		University:  user.University,
		RawClaims: map[string]interface{}{
			"sub":          user.Sub,
			"email":        user.Email,
			"phone_number": user.PhoneNumber,
			"given_name":   user.GivenName,
			"family_name":  user.FamilyName,
			"middle_name":  user.MiddleName,
			"birthdate":    user.BirthDate,
		},
	}

	g.reg.Logger().
		WithField("oidc_provider", g.config.ID).
		WithField("oidc_stage", "userinfo_claims").
		WithField("request_id", requestID).
		WithField("userinfo_url", userinfoURL).
		WithField("userinfo_given_name_all_upper", isAllUpperText(user.GivenName)).
		WithField("userinfo_family_name_all_upper", isAllUpperText(user.FamilyName)).
		WithField("userinfo_email_all_upper", isAllUpperText(user.Email)).
		WithSensitiveField("userinfo_claims_raw", user).
		WithSensitiveField("userinfo_claims_mapped", claims).
		WithField("latency_ms", time.Since(stageStart).Milliseconds()).
		Debug("OIDC userinfo claims parsed")

	return claims, nil
}

func (g *ProviderSber) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, sberAPITimeout)
	defer cancel()

	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, err
	}

	tokenURL, err := url.Parse(o.Endpoint.TokenURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReasonf("%s", err))
	}

	tokenEndpoint := fmt.Sprintf("%s://%s%s", tokenURL.Scheme, tokenURL.Host, tokenURL.Path)
	requestID := randx.MustString(32, []rune("0123456789ABCDEF"))
	stageStart := time.Now()
	exchangeTrace := &sberTokenExchangeTrace{}
	client := g.reg.HTTPClient(ctx).HTTPClient
	clientTimeout := client.Timeout
	if clientTimeout == 0 || clientTimeout > sberAPITimeout {
		clientTimeout = sberAPITimeout
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &sberTokenLoggingTransport{
			base:       client.Transport,
			logger:     g.reg.Logger(),
			providerID: g.config.ID,
			tokenURL:   tokenEndpoint,
			requestID:  requestID,
			startedAt:  stageStart,
			trace:      exchangeTrace,
		},
		Timeout: clientTimeout,
	})

	token, err := o.Exchange(ctx, code, opts...)
	if err != nil {
		fields := g.reg.Logger().
			WithError(err).
			WithField("oidc_provider", g.config.ID).
			WithField("oidc_stage", "token_exchange").
			WithField("request_id", requestID).
			WithField("token_url", tokenEndpoint).
			WithField("latency_ms", time.Since(stageStart).Milliseconds())

		if statusCode, bodyFragment, ok := extractOAuth2RetrieveError(err); ok {
			fields = fields.WithField("http_status", statusCode)
			if bodyFragment != "" {
				fields = fields.WithField("response_body_fragment", bodyFragment)
			}
			fields.Error("OIDC token exchange failed with upstream response")
			curlCmd := formatSberTokenExchangeCurl(tokenEndpoint, requestID, exchangeTrace.RequestForm, exchangeTrace.TLSClientCertPresent)
			return nil, errors.WithStack(
				herodot.ErrUpstreamError.
					WithWrap(err).
					WithReasonf("sber token exchange failed: debug_version=%s stage=token_exchange provider=%s token_endpoint=%s request_id=%s http_status=%d response=%q tls_client_cert_present=%t tls_client_config_absent=%t mtls_cert_path=%q mtls_key_path=%q mtls_attach_error=%q curl_request=%q",
						sberTokenDebugVersion,
						g.config.ID,
						tokenEndpoint,
						requestID,
						statusCode,
						bodyFragment,
						exchangeTrace.TLSClientCertPresent,
						exchangeTrace.TLSClientConfigAbsent,
						exchangeTrace.MTLSCertPath,
						exchangeTrace.MTLSKeyPath,
						exchangeTrace.MTLSAttachError,
						curlCmd,
					),
			)
		}

		fields.Error("OIDC token exchange failed")
		return nil, errors.WithStack(
			herodot.ErrUpstreamError.
				WithWrap(err).
				WithReasonf("Sber token exchange failed: stage=token_exchange provider=%s token_endpoint=%s request_id=%s", g.config.ID, tokenEndpoint, requestID),
		)
	}

	if err := validateSberIDToken(token, g.config.ID, g.config.ClientID, sberFlowIDFromContext(ctx)); err != nil {
		return nil, err
	}

	g.reg.Logger().
		WithField("oidc_provider", g.config.ID).
		WithField("oidc_stage", "token_exchange").
		WithField("request_id", requestID).
		WithField("token_url", tokenEndpoint).
		WithField("latency_ms", time.Since(stageStart).Milliseconds()).
		Debug("OIDC token exchange succeeded")

	return token, nil
}
