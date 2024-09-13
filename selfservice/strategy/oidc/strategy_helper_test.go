// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/julienschmidt/httprouter"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"github.com/rakutentech/jwk-go/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/resilience"
	"github.com/ory/x/urlx"
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

func (token *idTokenClaims) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		IdToken struct {
			Website     string   `json:"website,omitempty"`
			Groups      []string `json:"groups,omitempty"`
			Picture     string   `json:"picture,omitempty"`
			PhoneNumber string   `json:"phone_number,omitempty"`
		} `json:"id_token"`
	}{
		IdToken: struct {
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
	require.NoError(t, resilience.Retry(logrusx.New("", ""), time.Second*10, time.Minute*2, func() error {
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
			// https://github.com/golang/oauth2/blob/5fd42413edb3b1699004a31b72e485e0e4ba1b13/internal/token.go#L227-L242
			TokenEndpointAuthMethod: "client_secret_post",
		}))

		res, err := http.Post(remote+"/admin/clients", "application/json", &b)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body := ioutilx.MustReadAll(res.Body)
		if http.StatusCreated != res.StatusCode {
			return errors.Errorf("got status code: %d\n%s", res.StatusCode, body)
		}

		id = gjson.GetBytes(body, "client_id").String()
		secret = gjson.GetBytes(body, "client_secret").String()
		return nil
	}))
	return
}

func newHydraIntegration(t *testing.T, remote *string, subject *string, claims *idTokenClaims, scope *[]string, addr string) (*http.Server, string) {
	router := httprouter.New()

	type p struct {
		Subject    string          `json:"subject,omitempty"`
		Session    json.RawMessage `json:"session,omitempty"`
		GrantScope []string        `json:"grant_scope,omitempty"`
	}

	do := func(w http.ResponseWriter, r *http.Request, href string, payload io.Reader) {
		req, err := http.NewRequest("PUT", href, payload)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		body := ioutilx.MustReadAll(res.Body)
		require.Equal(t, http.StatusOK, res.StatusCode, "%s", body)

		var response struct {
			RedirectTo string `json:"redirect_to"`
		}
		require.NoError(t, json.NewDecoder(bytes.NewBuffer(body)).Decode(&response))
		require.NotNil(t, response.RedirectTo, "%s", body)

		http.Redirect(w, r, response.RedirectTo, http.StatusSeeOther)
	}

	router.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NotEmpty(t, *remote)
		require.NotEmpty(t, *subject)

		challenge := r.URL.Query().Get("login_challenge")
		require.NotEmpty(t, challenge)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&p{
			Subject: *subject,
		}))
		href := urlx.MustJoin(*remote, "/admin/oauth2/auth/requests/login/accept") + "?login_challenge=" + challenge
		do(w, r, href, &b)
	})

	router.GET("/consent", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NotEmpty(t, *remote)
		require.NotNil(t, *scope)

		challenge := r.URL.Query().Get("consent_challenge")
		require.NotEmpty(t, challenge)

		var b bytes.Buffer
		msg, err := json.Marshal(claims)
		require.NoError(t, err)
		require.NoError(t, json.NewEncoder(&b).Encode(&p{GrantScope: *scope, Session: msg}))
		href := urlx.MustJoin(*remote, "/admin/oauth2/auth/requests/consent/accept") + "?consent_challenge=" + challenge
		do(w, r, href, &b)
	})

	if addr == "" {
		server := httptest.NewServer(router)
		t.Cleanup(server.Close)
		server.URL = strings.Replace(server.URL, "127.0.0.1", "localhost", 1)
		return server.Config, server.URL
	}

	parsed, err := url.ParseRequestURI(addr)
	require.NoError(t, err)

	listener, err := net.Listen("tcp", ":"+parsed.Port())
	require.NoError(t, err, "port busy?")
	server := &http.Server{Handler: router}
	go server.Serve(listener)
	t.Cleanup(func() {
		assert.NoError(t, server.Close())
	})
	return server, addr
}

func newReturnTs(t *testing.T, reg driver.Registry) *httptest.Server {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app_code" {
			reg.Writer().Write(w, r, "ok")
			return
		}
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		require.NoError(t, err)
		reg.Writer().Write(w, r, sess)
	}))
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL)
	t.Cleanup(ts.Close)
	return ts
}

func newUI(t *testing.T, reg driver.Registry) *httptest.Server {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var e interface{}
		var err error
		if r.URL.Path == "/login" {
			e, err = reg.LoginFlowPersister().GetLoginFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		} else if r.URL.Path == "/registration" {
			e, err = reg.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		} else if r.URL.Path == "/settings" {
			e, err = reg.SettingsFlowPersister().GetSettingsFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		}

		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	t.Cleanup(ts.Close)
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceLoginUI, ts.URL+"/login")
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceRegistrationUI, ts.URL+"/registration")
	reg.Config().MustSet(ctx, config.ViperKeySelfServiceSettingsURL, ts.URL+"/settings")
	return ts
}

func newHydra(t *testing.T, subject *string, claims *idTokenClaims, scope *[]string) (remoteAdmin, remotePublic, hydraIntegrationTSURL string) {
	remoteAdmin = os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_ADMIN")
	remotePublic = os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_PUBLIC")

	hydraIntegrationTS, hydraIntegrationTSURL := newHydraIntegration(t, &remoteAdmin, subject, claims, scope, os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_INTEGRATION_ADDR"))
	t.Cleanup(func() {
		require.NoError(t, hydraIntegrationTS.Close())
	})

	if remotePublic == "" && remoteAdmin == "" {
		t.Logf("Environment did not provide Ory Hydra, starting fresh.")
		publicPort, err := freeport.GetFreePort()
		require.NoError(t, err)

		pool, err := dockertest.NewPool("")
		require.NoError(t, err)
		hydra, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "oryd/hydra",
			// Keep tag in sync with the version in ci.yaml
			Tag: "v2.2.0@sha256:6c0f9195fe04ae16b095417b323881f8c9008837361160502e11587663b37c09",
			Env: []string{
				"DSN=memory",
				fmt.Sprintf("URLS_SELF_ISSUER=http://localhost:%d/", publicPort),
				"URLS_LOGIN=" + hydraIntegrationTSURL + "/login",
				"URLS_CONSENT=" + hydraIntegrationTSURL + "/consent",
				"LOG_LEAK_SENSITIVE_VALUES=true",
				"SECRETS_SYSTEM=someverylongsecretthatis32byteslong",
			},
			Cmd:          []string{"serve", "all", "--dev"},
			ExposedPorts: []string{"4444/tcp", "4445/tcp"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"4444/tcp": {{HostPort: fmt.Sprintf("%d/tcp", publicPort)}},
			},
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, hydra.Close())
		})
		require.NoError(t, hydra.Expire(uint(60*5)))

		require.NotEmpty(t, hydra.GetPort("4444/tcp"), "%+v", hydra.Container.NetworkSettings.Ports)
		require.NotEmpty(t, hydra.GetPort("4445/tcp"), "%+v", hydra.Container)

		remotePublic = "http://localhost:" + hydra.GetPort("4444/tcp")
		remoteAdmin = "http://localhost:" + hydra.GetPort("4445/tcp")

		err = resilience.Retry(logrusx.New("", ""), time.Second*1, time.Second*5, func() error {
			pr := remotePublic + "/health/ready"
			res, err := http.DefaultClient.Get(pr)
			if err != nil || res.StatusCode != 200 {
				return errors.Errorf("Hydra public is not ready at " + pr)
			}

			wellKnown := remotePublic + "/.well-known/openid-configuration"
			res, err = http.DefaultClient.Get(wellKnown)
			if err != nil || res.StatusCode != 200 {
				return errors.Errorf("Hydra .well-known is not ready at " + wellKnown)
			}

			ar := remoteAdmin + "/health/ready"
			res, err = http.DefaultClient.Get(ar)
			if err != nil && res.StatusCode != 200 {
				return errors.Errorf("Hydra admin is not ready at " + ar)
			} else {
				return nil
			}
		})
		require.NoError(t, err)

	}

	t.Logf("Ory Hydra running at: %s %s", remotePublic, remoteAdmin)

	return remoteAdmin, remotePublic, hydraIntegrationTSURL
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

func viperSetProviderConfig(t *testing.T, conf *config.Config, providers ...oidc.Configuration) {
	ctx := context.Background()
	baseKey := fmt.Sprintf("%s.%s", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeOIDC)
	currentConfig := conf.GetProvider(ctx).Get(baseKey + ".config")
	currentEnabled := conf.GetProvider(ctx).Get(baseKey + ".enabled")

	conf.MustSet(ctx, baseKey+".config", &oidc.ConfigurationCollection{Providers: providers})
	conf.MustSet(ctx, baseKey+".enabled", true)

	t.Cleanup(func() {
		conf.MustSet(ctx, baseKey+".config", currentConfig)
		conf.MustSet(ctx, baseKey+".enabled", currentEnabled)
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

type claims struct {
	*jwt.RegisteredClaims
	Email string `json:"email"`
}

func createIdToken(t *testing.T, cl jwt.RegisteredClaims) string {
	key := &jwk.KeySpec{}
	require.NoError(t, json.Unmarshal(rawKey, key))
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &claims{
		RegisteredClaims: &cl,
		Email:            "acme@ory.sh",
	})
	token.Header["kid"] = key.KeyID
	s, err := token.SignedString(key.Key)
	require.NoError(t, err)
	return s
}
