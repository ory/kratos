// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/phayes/freeport"
	"github.com/rakutentech/jwk-go/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"

	"github.com/ory/dockertest/v4"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
	"github.com/ory/x/httpx"
	"github.com/ory/x/ioutilx"
)

type idTokenClaims struct {
	traits struct {
		website string
		groups  []string
	}
	metadataPublic struct {
		picture string
	}
	metadataAdmin struct {
		phoneNumber string
	}
}

func (token idTokenClaims) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		IDToken struct {
			Website     string   `json:"website,omitempty"`
			Groups      []string `json:"groups,omitempty"`
			Picture     string   `json:"picture,omitempty"`
			PhoneNumber string   `json:"phone_number,omitempty"`
		} `json:"id_token"`
	}{
		IDToken: struct {
			Website     string   `json:"website,omitempty"`
			Groups      []string `json:"groups,omitempty"`
			Picture     string   `json:"picture,omitempty"`
			PhoneNumber string   `json:"phone_number,omitempty"`
		}{
			Website:     token.traits.website,
			Groups:      token.traits.groups,
			Picture:     token.metadataPublic.picture,
			PhoneNumber: token.metadataAdmin.phoneNumber,
		},
	})
}

func createClient(t *testing.T, remote string, redir []string) (id, secret string) {
	client := testhelpers.NewTestClient(t)
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&struct {
			Scope                   string   `json:"scope"`
			GrantTypes              []string `json:"grant_types"`
			ResponseTypes           []string `json:"response_types"`
			RedirectURIs            []string `json:"redirect_uris"`
			TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		}{
			GrantTypes:    []string{"authorization_code", "refresh_token"},
			ResponseTypes: []string{"code"},
			Scope:         "offline offline_access openid",
			RedirectURIs:  redir,

			// This is a workaround to prevent golang.org/x/oauth2 from
			// swallowing the actual error messages from failed token exchanges.
			//
			// The library first attempts to use the Authorization header to
			// pass Client ID+secret during token exchange (client_secret_basic
			// in Hydra terminology). If that fails (with any error), it tries
			// again with the Client ID+secret passed in the HTTP POST body
			// (client_secret_post in Hydra). If that also fails, this second
			// error is returned.
			//
			// Now, if the the client was indeed configured to use
			// client_secret_basic, but the token exchange fails for another
			// reason, the error message will be swallowed and replaced with
			// "invalid_client".
			//
			// Manually setting this to client_secret_post means that during
			// tests, all token exchanges will first fail with `invalid_client`
			// and then be retried with the correct method. This is the only way
			// to get the actual error message from the server, however.
			//
			// https://github.com/golang/oauth2/blob/5fd42413edb3b1699004a31b72e485e0e4ba1b13/pkg/token.go#L227-L242
			TokenEndpointAuthMethod: "client_secret_post",
		}))

		res, err := client.Post(remote+"/admin/clients", "application/json", &b)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()

		body := ioutilx.MustReadAll(res.Body)
		require.Equal(t, http.StatusCreated, res.StatusCode)

		id = gjson.GetBytes(body, "client_id").String()
		secret = gjson.GetBytes(body, "client_secret").String()
	}, time.Minute, time.Second)
	return
}

type hydraFlowParams struct {
	subject string
	claims  idTokenClaims
	scope   []string
	// acr, when non-nil, is returned as the upstream `acr`
	// claim on the login accept. The empty string is omitted from the
	// payload (the harness uses `omitempty` on the wire format).
	acr *string
	// amr, when non-nil, is returned as the upstream `amr`
	// claim on the login accept. An empty slice is omitted from the
	// payload (the harness uses `omitempty` on the wire format).
	amr []string
}

func wrapClientForHydraLoginConsent(t *testing.T, wrapped *http.Client, hydraAdmin string, params hydraFlowParams) *http.Client {
	adminClient := testhelpers.NewTestClient(t)
	doPut := func(href, payload string) string {
		req, err := http.NewRequest("PUT", hydraAdmin+href, strings.NewReader(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		res, err := adminClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()

		body := ioutilx.MustReadAll(res.Body)
		require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)

		redirectTo := gjson.GetBytes(body, "redirect_to")
		require.Truef(t, redirectTo.Exists(), "%s", body)
		require.NotEmptyf(t, redirectTo.Str, "%s", body)

		return redirectTo.Str
	}

	c := *wrapped
	tsp := c.Transport
	if tsp == nil {
		tsp = http.DefaultTransport
	}
	c.Transport = httpx.TransportFunc(func(req *http.Request) (*http.Response, error) {
		if strings.HasPrefix(req.URL.String(), hydraLoginURL) {
			challenge := req.URL.Query().Get("login_challenge")
			require.NotEmptyf(t, challenge, "%s", req.URL)

			payload := map[string]any{
				"subject": params.subject,
			}
			if params.acr != nil {
				payload["acr"] = *params.acr
			}
			if params.amr != nil {
				payload["amr"] = params.amr
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			redirectTo := doPut("/admin/oauth2/auth/requests/login/accept?login_challenge="+challenge, string(body))
			req, err := http.NewRequest("GET", redirectTo, nil)
			require.NoError(t, err)
			return c.Do(req)
		}
		if strings.HasPrefix(req.URL.String(), hydraConsentURL) {
			challenge := req.URL.Query().Get("consent_challenge")
			require.NotEmptyf(t, challenge, "%s", req.URL)

			body, err := json.Marshal(map[string]any{
				"grant_scope": params.scope,
				"session":     params.claims,
			})
			require.NoError(t, err)
			redirectTo := doPut("/admin/oauth2/auth/requests/consent/accept?consent_challenge="+challenge, string(body))
			req, err := http.NewRequest("GET", redirectTo, nil)
			require.NoError(t, err)
			return c.Do(req)
		}
		return tsp.RoundTrip(req)
	})
	return &c
}

func newUI(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var e interface{}
		var err error
		switch r.URL.Path {
		case "/login":
			e, err = reg.LoginFlowPersister().GetLoginFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		case "/registration":
			e, err = reg.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		case "/settings":
			e, err = reg.SettingsFlowPersister().GetSettingsFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		}

		if err != nil {
			reg.Writer().Write(w, r, err)
			return
		}
		reg.Writer().Write(w, r, e)
	}))
	t.Cleanup(ts.Close)
	reg.Config().MustSet(t.Context(), config.ViperKeySelfServiceLoginUI, ts.URL+"/login")
	reg.Config().MustSet(t.Context(), config.ViperKeySelfServiceRegistrationUI, ts.URL+"/registration")
	reg.Config().MustSet(t.Context(), config.ViperKeySelfServiceSettingsURL, ts.URL+"/settings")
	return ts
}

// These URLs are not actually resolving, but instead the client is catching the redirect and handling it.
const (
	hydraLoginURL   = "https://hydra-placeholder-ui/login"
	hydraConsentURL = "https://hydra-placeholder-ui/consent"
)

func newHydra(t *testing.T) (remoteAdmin, remotePublic string) {
	remoteAdmin = os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_ADMIN")
	remotePublic = os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_PUBLIC")

	start := time.Now()
	if remotePublic == "" && remoteAdmin == "" {
		t.Logf("Environment did not provide Ory Hydra, starting fresh.")
		pool := dockertest.NewPoolT(t, "")

		// The public host port must be known before the container starts because
		// it is part of URLS_SELF_ISSUER. freeport releases the port before Docker
		// binds it, so another process can grab it in between — retry on that race.
		var hydra dockertest.ClosableResource
		ctx := t.Context()
		require.EventuallyWithT(t, func(t *assert.CollectT) {
			publicPort, err := freeport.GetFreePort()
			require.NoError(t, err)

			hydra, err = pool.Run(ctx, "oryd/hydra",
				// Keep tag in sync with the version in ci.yaml
				dockertest.WithTag("v2.2.0"),
				dockertest.WithoutReuse(),
				dockertest.WithEnv([]string{
					"DSN=memory",
					fmt.Sprintf("URLS_SELF_ISSUER=http://127.0.0.1:%d/", publicPort),
					"URLS_LOGIN=" + hydraLoginURL,
					"URLS_CONSENT=" + hydraConsentURL,
					"LOG_LEAK_SENSITIVE_VALUES=true",
					"SECRETS_SYSTEM=someverylongsecretthatis32byteslong",
				}),
				dockertest.WithCmd([]string{"serve", "all", "--dev"}),
				dockertest.WithContainerConfig(func(cc *container.Config) {
					cc.ExposedPorts = network.PortSet{
						network.MustParsePort("4444/tcp"): struct{}{},
						network.MustParsePort("4445/tcp"): struct{}{},
					}
				}),
				dockertest.WithHostConfig(func(hc *container.HostConfig) {
					hc.PortBindings = network.PortMap{
						network.MustParsePort("4444/tcp"): {{HostPort: strconv.Itoa(publicPort)}},
						network.MustParsePort("4445/tcp"): {{HostPort: ""}},
					}
				}),
			)
			require.NoError(t, err)
		}, 2*time.Minute, 100*time.Millisecond)
		hydra.Cleanup(t)
		require.NotEmpty(t, hydra.GetPort("4444/tcp"), "%+v", hydra.Container().NetworkSettings.Ports)
		require.NotEmpty(t, hydra.GetPort("4445/tcp"), "%+v", hydra.Container)

		remotePublic = "http://127.0.0.1:" + hydra.GetPort("4444/tcp")
		remoteAdmin = "http://127.0.0.1:" + hydra.GetPort("4445/tcp")

		client := testhelpers.NewTestClient(t)
		require.EventuallyWithT(t, func(t *assert.CollectT) {
			res, err := client.Get(remotePublic + "/health/ready")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			res, err = client.Get(remotePublic + "/.well-known/openid-configuration")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			res, err = client.Get(remoteAdmin + "/health/ready")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
		}, 30*time.Second, 100*time.Millisecond)
	}

	t.Logf("Ory Hydra started after %s at: %s %s", time.Since(start), remotePublic, remoteAdmin)

	return remoteAdmin, remotePublic
}

func newOIDCProvider(
	t *testing.T,
	kratos *httptest.Server,
	hydraPublic string,
	hydraAdmin string,
	id string,
	opts ...func(*oidc.Configuration),
) oidc.Configuration {
	clientID, secret := createClient(t, hydraAdmin, []string{kratos.URL + oidc.RouteBase + "/callback/" + id, kratos.URL + oidc.RouteCallbackGeneric})

	cfg := oidc.Configuration{
		Provider:     "generic",
		ID:           id,
		ClientID:     clientID,
		ClientSecret: secret,
		IssuerURL:    hydraPublic + "/",
		Mapper:       "file://./stub/oidc.hydra.jsonnet",
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

func setProviderConfig(t *testing.T, conf *config.Config, providers ...oidc.Configuration) {
	baseKey := fmt.Sprintf("%s.%s", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeOIDC)
	currentConfig := conf.GetProvider(t.Context()).Get(baseKey + ".config")
	currentEnabled := conf.GetProvider(t.Context()).Get(baseKey + ".enabled")

	conf.MustSet(t.Context(), baseKey+".config", &oidc.ConfigurationCollection{Providers: providers})
	conf.MustSet(t.Context(), baseKey+".enabled", true)

	t.Cleanup(func() {
		conf.MustSet(t.Context(), baseKey+".config", currentConfig)
		conf.MustSet(t.Context(), baseKey+".enabled", currentEnabled)
	})
}

// AssertSystemError asserts an error ui response
func AssertSystemError(t *testing.T, errTS *httptest.Server, res *http.Response, body []byte, code int, reason string) {
	require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

	assert.Equal(t, int64(code), gjson.GetBytes(body, "code").Int(), "%s", body)
	assert.Contains(t, gjson.GetBytes(body, "reason").String(), reason, "%s", body)
}

//go:embed stub/jwk.json
var rawKey []byte

//go:embed stub/jwks_public.json
var publicJWKS []byte

// Just a public key set, to be able to test what happens if an ID token was issued by a different private key.
//
//go:embed stub/jwks_public2.json
var publicJWKS2 []byte

func createIDToken(t *testing.T, cl jwt.RegisteredClaims) string {
	key := &jwk.KeySpec{}
	require.NoError(t, json.Unmarshal(rawKey, key))
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &struct {
		*jwt.RegisteredClaims
		Email string `json:"email"`
	}{
		RegisteredClaims: &cl,
		Email:            "acme@ory.sh",
	})
	token.Header["kid"] = key.KeyID
	s, err := token.SignedString(key.Key)
	require.NoError(t, err)
	return s
}
