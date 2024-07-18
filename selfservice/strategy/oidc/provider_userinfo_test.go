// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jarcoal/httpmock"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/x/httpx"
	"github.com/ory/x/otelx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRegistry struct {
	*driver.RegistryDefault
	cl *retryablehttp.Client
}

func (s *mockRegistry) HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client {
	return s.cl
}

func TestProviderClaimsRespectsErrorCodes(t *testing.T) {
	conf, base := internal.NewFastRegistryWithMocks(t)
	require.NoError(t, conf.Set(context.Background(), config.ViperKeyClientHTTPNoPrivateIPRanges, true))
	base.SetTracer(otelx.NewNoop(nil, nil))
	reg := &mockRegistry{base, retryablehttp.NewClient()}

	ctx := context.Background()
	token := &oauth2.Token{AccessToken: "foo", Expiry: time.Now().Add(time.Hour)}

	expectedClaims := &oidc.Claims{
		Issuer:            "ignore-me",
		Subject:           "123456789012345",
		Name:              "John Doe",
		GivenName:         "John",
		FamilyName:        "Doe",
		Nickname:          "John Doe",
		PreferredUsername: "John Doe",
		Email:             "john.doe@example.com",
		EmailVerified:     true,
		Birthdate:         "01/01/1990",
	}

	defaultUserinfoHandler := func(req *http.Request) (*http.Response, error) {
		if head := req.Header.Get("Authorization"); len(head) == 0 {
			resp, err := httpmock.NewJsonResponse(401, map[string]interface{}{"error": ""})
			return resp, err
		}

		resp, err := httpmock.NewJsonResponse(200, expectedClaims)
		return resp, err
	}

	for _, tc := range []struct {
		name             string
		issuer           string
		userInfoEndpoint string
		config           *oidc.Configuration
		provider         oidc.Provider
		userInfoHandler  func(req *http.Request) (*http.Response, error)
		expectedClaims   *oidc.Claims
		useToken         *oauth2.Token
		hook             func(t *testing.T)
	}{
		{
			name:             "auth0",
			userInfoHandler:  defaultUserinfoHandler,
			userInfoEndpoint: "https://www.auth0.com/userinfo",
			provider: oidc.NewProviderAuth0(&oidc.Configuration{
				IssuerURL: "https://www.auth0.com",
				ID:        "auth0",
				Provider:  "auth0",
			}, reg),
		},
		{
			name:             "salesforce",
			userInfoHandler:  defaultUserinfoHandler,
			userInfoEndpoint: "https://www.salesforce.com/userinfo",
			provider: oidc.NewProviderSalesforce(&oidc.Configuration{
				IssuerURL: "https://www.salesforce.com",
				ID:        "salesforce",
				Provider:  "salesforce",
			}, reg),
		},
		{
			name:             "netid",
			userInfoHandler:  defaultUserinfoHandler,
			userInfoEndpoint: "https://broker.netid.de/userinfo",
			provider: oidc.NewProviderNetID(&oidc.Configuration{
				IssuerURL: "https://broker.netid.de",
				ID:        "netid",
				Provider:  "netid",
				ClientID:  "foo",
			}, reg),
			useToken: token.WithExtra(map[string]interface{}{"id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZW1haWwiOiJqb2huLmRvZUBleGFtcGxlLmNvbSIsImlhdCI6MTUxNjIzOTAyMiwidGlkIjoiYTliODYzODUtZjMyYy00ODAzLWFmYzgtNGIyMzEyZmJkZjI0IiwiaXNzIjoiaHR0cHM6Ly9icm9rZXIubmV0aWQuZGUvIiwiYXVkIjpbImZvbyJdLCJleHAiOjQwNzE3Mjg1MDR9.Zt_-9jULoENQ7pq6rKrevhecBlWKR2rzNti42EJti_OelCrGbl5zyHYRfIg264VzEY0Zp9BAZTWmcF6Z-cpuD05RddTXZDrrC_47bJeYDL-bjDfKjoSZUt_1JnNFgqeyMA1Ji6HCIsEf_g8QnuSELAO0c-qb2ANmDLVL8_dY6oUmCN5oWLJS2q6xO-Iaz-AuKaGTbLZBcjCe2NB_--kIx4IrgLm78U90X3ePOV0CcYZLNvDgzEsVJ2M4ixKp87bYUaZZ3JJEjuxgnHqKRMExDKron3mvfQtC1L-dZQyeDo3-mCFJL_JhEw9zLmoWFBD7aARrVL_yAe31o26FB3S-Dg"}),
			hook: func(t *testing.T) {
				httpmock.RegisterResponder("GET", "https://broker.netid.de/.well-known/openid-configuration",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, map[string]interface{}{
							"issuer":   "https://broker.netid.de/",
							"jwks_uri": "https://broker.netid.de/.well-known/jwks.json",
						})
					},
				)
				httpmock.RegisterResponder("GET", "https://broker.netid.de/.well-known/jwks.json",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, json.RawMessage(`{
  "keys": [
    {
      "kty": "RSA",
      "e": "AQAB",
      "alg": "RS256",
      "use": "sig",
      "n": "uw0hrmZpA5CH5lDnjbCK6ZQD-I0BGxde-ChDzUz5eK4r7VtyMGvAxwD-k-mWw4FJ2NgYmT_T89sAtE6NQqT5u9HAe-CI22lf5LMvmqvMzekcmBAvXNw8VeTV_N6CbS9INJrxf20cObf-kpTxVxlYtYxYwIhYdOw3DwX8y31vI38qHQ4_OzTRo4KFVLCr68MzIHHRI4d5EHrFv1VFjiaa_ATwuwCMUfg0RMnK09FpMCgvp0E6bQeptXhBBNQVQkoC5whT1GzikfSxyeugjQ_XuTt1MKoyYsN2pmfrBdcfWrPYvV_JPgO1MkEtqErvtCByairINfXrHTMOxNWe3sYlXQ"
    }
  ]
}`))
					},
				)
			},
			expectedClaims: &oidc.Claims{Issuer: "https://broker.netid.de/", Subject: "1234567890", Name: "John Doe", GivenName: "John", FamilyName: "Doe", LastName: "", MiddleName: "", Nickname: "John Doe", PreferredUsername: "John Doe", Profile: "", Picture: "", Website: "", Email: "john.doe@example.com", EmailVerified: true, Gender: "", Birthdate: "01/01/1990", Zoneinfo: "", Locale: "", PhoneNumber: "", PhoneNumberVerified: false, UpdatedAt: 0, HD: "", Team: ""},
		},
		{
			name:             "vk",
			userInfoEndpoint: "https://api.vk.com/method/users.get",
			provider: oidc.NewProviderVK(&oidc.Configuration{
				IssuerURL: "https://oauth.vk.com",
				ID:        "vk",
				Provider:  "vk",
			}, reg),
			useToken: token.WithExtra(map[string]interface{}{"email": "john.doe@example.com"}),
			userInfoHandler: func(req *http.Request) (*http.Response, error) {
				if head := req.URL.Query().Get("access_token"); len(head) == 0 {
					resp, err := httpmock.NewJsonResponse(401, map[string]interface{}{"error": ""})
					return resp, err
				}

				resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
					"response": []map[string]interface{}{{"id": 123456789012345}},
				})
				return resp, err
			},

			expectedClaims: &oidc.Claims{
				Issuer:  "https://api.vk.com/method/users.get",
				Subject: "123456789012345",
				Email:   "john.doe@example.com",
			},
		},
		{
			name:             "yandex",
			userInfoEndpoint: "https://login.yandex.ru/info",
			provider: oidc.NewProviderYandex(&oidc.Configuration{
				IssuerURL: "https://oauth.yandex.com",
				ID:        "vk",
				Provider:  "vk",
			}, reg),
			useToken: token.WithExtra(map[string]interface{}{"email": "john.doe@example.com"}),
			userInfoHandler: func(req *http.Request) (*http.Response, error) {
				if head := req.URL.Query().Get("oauth_token"); len(head) == 0 {
					resp, err := httpmock.NewJsonResponse(401, map[string]interface{}{"error": ""})
					return resp, err
				}

				resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
					"id":            "123456789012345",
					"default_email": "john.doe@example.com",
				})
				return resp, err
			},

			expectedClaims: &oidc.Claims{
				Issuer:  "https://login.yandex.ru/info",
				Subject: "123456789012345",
				Email:   "john.doe@example.com",
				Picture: "https://avatars.yandex.net/get-yapic//islands-200",
			},
		},
		{
			name: "facebook",
			hook: func(t *testing.T) {
				httpmock.RegisterResponder("GET", "https://www.facebook.com/.well-known/openid-configuration",
					func(req *http.Request) (*http.Response, error) {
						resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
							"issuer": "https://www.facebook.com",
						})
						return resp, err
					},
				)
			},
			userInfoEndpoint: "https://graph.facebook.com/me",
			provider: oidc.NewProviderFacebook(&oidc.Configuration{
				ID:       "facebook",
				Provider: "facebook",
			}, reg),
			useToken: token,
			userInfoHandler: func(req *http.Request) (*http.Response, error) {
				if head := req.Header.Get("Authorization"); len(head) == 0 {
					resp, err := httpmock.NewJsonResponse(401, map[string]interface{}{"error": ""})
					return resp, err
				}
				if _, ok := req.URL.Query()["appsecret_proof"]; !ok {
					resp, err := httpmock.NewJsonResponse(400, map[string]interface{}{"error": ""})
					return resp, err
				}

				resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
					"id":         "123456789012345",
					"name":       "John Doe",
					"first_name": "John",
					"last_name":  "Doe",
					"email":      "john.doe@example.com",
					"birthday":   "01/01/1990",
				})
				return resp, err
			},
			expectedClaims: &oidc.Claims{
				Issuer:            "https://graph.facebook.com/me?fields=id,name,first_name,last_name,middle_name,email,picture,birthday,gender&appsecret_proof=0c0d98f7e3d9d45e72e8877bc1b104327efb9c07b18f2ffeced76d81307f1fff",
				Subject:           "123456789012345",
				Name:              "John Doe",
				GivenName:         "John",
				FamilyName:        "Doe",
				Nickname:          "John Doe",
				PreferredUsername: "John Doe",
				Email:             "john.doe@example.com",
				EmailVerified:     true,
				Birthdate:         "01/01/1990",
			},
		},
		{
			name:             "gitlab",
			userInfoHandler:  defaultUserinfoHandler,
			userInfoEndpoint: "https://www.gitlab.com/oauth/userinfo",
			provider: oidc.NewProviderGitLab(&oidc.Configuration{
				IssuerURL: "https://www.gitlab.com",
				ID:        "gitlab",
				Provider:  "gitlab",
			}, reg),
		},
		// Microsoft is a more complicated set up because it actually verifies the ID Token before using the userinfo endpoint.
		{
			name: "microsoft",
			userInfoHandler: func(req *http.Request) (*http.Response, error) {
				if head := req.Header.Get("Authorization"); len(head) == 0 {
					resp, err := httpmock.NewJsonResponse(401, map[string]interface{}{"error": ""})
					return resp, err
				}

				resp, err := httpmock.NewJsonResponse(200, json.RawMessage(`{"id":"new-id"}`))
				return resp, err
			},
			userInfoEndpoint: "https://graph.microsoft.com/v1.0/me",
			provider: oidc.NewProviderMicrosoft(&oidc.Configuration{
				IssuerURL:     "https://login.microsoftonline.com/",
				ID:            "microsoft",
				Provider:      "microsoft",
				Tenant:        "a9b86385-f32c-4803-afc8-4b2312fbdf24",
				ClientID:      "foo",
				ClientSecret:  "bar",
				SubjectSource: "me",
			}, reg),
			useToken: token.WithExtra(map[string]interface{}{"id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZW1haWwiOiJqb2huLmRvZUBleGFtcGxlLmNvbSIsImlhdCI6MTUxNjIzOTAyMiwidGlkIjoiYTliODYzODUtZjMyYy00ODAzLWFmYzgtNGIyMzEyZmJkZjI0IiwiaXNzIjoiaHR0cHM6Ly9sb2dpbi5taWNyb3NvZnRvbmxpbmUuY29tL2E5Yjg2Mzg1LWYzMmMtNDgwMy1hZmM4LTRiMjMxMmZiZGYyNC92Mi4wIiwiYXVkIjpbImZvbyJdLCJleHAiOjQwNzE3Mjg1MDR9.LRgolO5_-26uMrBo89NCzUfi87a8jf7rXlWgZVYnpfowzqn-U0JhNGVzEOXOACoPdX9HsEtYj4hZxYgYcd6z7yqgmOZXE-y58L5BHYZU1kk37O1Dl_VnN-BmeCWs_JZiXF2KEnu7BW8btYnb26qCnc3_8RGbyJwI4UU6ynbJzAmLCPUPgoMZ2Jpahx_K8vqLe-4rveLyUvVVHMoUAV16I-Wg08GcuA5cY0_91-QpA5Kq0AA58wbrUbAOEAOYOpa_63QqYcCcZLbnX_w09Z1YCXGGVSbbpzjr3cJ8EFA6doloBKRLWtNtEnGk4hyjHyp_ls89ZYqJ1ngy95AEswzwJQ"}),
			hook: func(t *testing.T) {
				httpmock.RegisterResponder("GET", "https://login.microsoftonline.com/a9b86385-f32c-4803-afc8-4b2312fbdf24/v2.0/.well-known/openid-configuration",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, map[string]interface{}{
							"issuer":   "https://login.microsoftonline.com/a9b86385-f32c-4803-afc8-4b2312fbdf24/v2.0",
							"jwks_uri": "https://login.microsoftonline.com/a9b86385-f32c-4803-afc8-4b2312fbdf24/v2.0/.well-known/jwks.json",
						})
					},
				)
				httpmock.RegisterResponder("GET", "https://login.microsoftonline.com/a9b86385-f32c-4803-afc8-4b2312fbdf24/v2.0/.well-known/jwks.json",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, json.RawMessage(`{
  "keys": [
    {
      "kty": "RSA",
      "e": "AQAB",
      "use": "sig",
      "alg": "RS256",
      "n": "xGnYx5u9S7LCrtINTI1Kav6pqzFj72JhYunqrDhnG18Iu-MJdCSmd26xiiZFY6UsJoG-te_DcUE_YHvhx6vwC0tJO-Z2uxwKRxWsMiyW8v3V0Bfbtu7jlf0kpQYeAeAIAHih7GU_v5jtxaGLNR9JIbmENUGOoIfydfTWLKHtwI0MB8tNVwjqP8e6ZSr6DNIjsKKuz98BeRlZ576jyH2GcmM9DguLZv73qU1m1OLoHFW5rAPH70-nhr9V67TTR1A1is9v85uLlKBnZgKp-EdYtBEvKRWSpw5vxCnKggJjEYATOB2H5m_c8F1e1tsJLAehaXKMKZYI-VlWWxj7KixDKw"
    }
  ]
}`))
					},
				)
			},
			expectedClaims: &oidc.Claims{
				Issuer: "https://login.microsoftonline.com/a9b86385-f32c-4803-afc8-4b2312fbdf24/v2.0", Subject: "new-id", Name: "John Doe", Email: "john.doe@example.com",
				RawClaims: map[string]interface{}{"aud": []interface{}{"foo"}, "exp": 4.071728504e+09, "iat": 1.516239022e+09, "iss": "https://login.microsoftonline.com/a9b86385-f32c-4803-afc8-4b2312fbdf24/v2.0", "email": "john.doe@example.com", "name": "John Doe", "sub": "1234567890", "tid": "a9b86385-f32c-4803-afc8-4b2312fbdf24"},
			},
		},
		{
			name: "dingtalk",
			userInfoHandler: func(req *http.Request) (*http.Response, error) {
				if head := req.Header.Get("x-acs-dingtalk-access-token"); len(head) == 0 {
					resp, err := httpmock.NewJsonResponse(401, map[string]interface{}{"error": ""})
					return resp, err
				}

				resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
					"openId": "123456789012345",
					"email":  "john.doe@example.com",
				})
				return resp, err
			},
			userInfoEndpoint: "https://api.dingtalk.com/v1.0/contact/users/me",
			provider: oidc.NewProviderDingTalk(&oidc.Configuration{
				IssuerURL: "https://www.dingtalk.com",
				ID:        "dingtalk",
				Provider:  "dingtalk",
			}, reg),
			expectedClaims: &oidc.Claims{
				Issuer:  "https://api.dingtalk.com/v1.0/contact/users/me",
				Subject: "123456789012345",
				Email:   "john.doe@example.com",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token := token
			if tc.useToken != nil {
				token = tc.useToken
			}

			t.Run("http error is respected", func(t *testing.T) {
				httpmock.ActivateNonDefault(reg.cl.HTTPClient)
				t.Cleanup(httpmock.DeactivateAndReset)

				if tc.hook != nil {
					tc.hook(t)
				}

				httpmock.RegisterResponder("GET", tc.userInfoEndpoint, func(req *http.Request) (*http.Response, error) {
					resp, err := httpmock.NewJsonResponse(455, map[string]interface{}{})
					return resp, err
				})

				_, err := tc.provider.(oidc.OAuth2Provider).Claims(ctx, token, url.Values{})
				var he *herodot.DefaultError
				require.ErrorAs(t, err, &he)
				assert.Equal(t, "OpenID Connect provider returned a 455 status code but 200 is expected.", he.Reason())
			})

			t.Run("call is successful", func(t *testing.T) {
				httpmock.ActivateNonDefault(reg.cl.HTTPClient)
				t.Cleanup(httpmock.DeactivateAndReset)

				if tc.hook != nil {
					tc.hook(t)
				}

				httpmock.RegisterResponder("GET", tc.userInfoEndpoint, tc.userInfoHandler)

				claims, err := tc.provider.(oidc.OAuth2Provider).Claims(ctx, token, url.Values{})
				require.NoError(t, err)
				if tc.expectedClaims == nil {
					assert.Equal(t, expectedClaims, claims)
				} else {
					assert.Equal(t, tc.expectedClaims, claims)
				}
			})
		})
	}
}
