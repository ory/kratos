package oidc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/logrusx"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/ory/viper"
	"github.com/ory/x/resilience"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
)

func createClient(t *testing.T, remote string, redir, id string) {
	require.NoError(t, resilience.Retry(logrusx.New("", ""), time.Second*10, time.Minute*2, func() error {
		if req, err := http.NewRequest("DELETE", remote+"/clients/"+id, nil); err != nil {
			return err
		} else if _, err := http.DefaultClient.Do(req); err != nil {
			return err
		}

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&struct {
			ClientID      string   `json:"client_id"`
			ClientSecret  string   `json:"client_secret"`
			Scope         string   `json:"scope"`
			GrantTypes    []string `json:"grant_types"`
			ResponseTypes []string `json:"response_types"`
			RedirectURIs  []string `json:"redirect_uris"`
		}{
			ClientID:      id,
			ClientSecret:  "secret",
			GrantTypes:    []string{"authorization_code", "refresh_token"},
			ResponseTypes: []string{"code"},
			Scope:         "offline offline_access openid",
			RedirectURIs:  []string{redir},
		}))

		res, err := http.Post(remote+"/clients", "application/json", &b)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if http.StatusCreated != res.StatusCode {
			return errors.Errorf("got status code: %d", http.StatusCreated)
		}
		return nil
	}))
}

func newHydraIntegration(t *testing.T, remote *string, subject *string, scope *[]string, addr string) (*http.Server, string) {
	router := httprouter.New()

	type p struct {
		Subject    string   `json:"subject,omitempty"`
		GrantScope []string `json:"grant_scope,omitempty"`
	}

	var do = func(w http.ResponseWriter, r *http.Request, href string, payload io.Reader) {
		req, err := http.NewRequest("PUT", href, payload)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		body := x.MustReadAll(res.Body)
		require.Equal(t, http.StatusOK, res.StatusCode, "%s", body)

		var response struct {
			RedirectTo string `json:"redirect_to"`
		}
		require.NoError(t, json.NewDecoder(bytes.NewBuffer(body)).Decode(&response))
		require.NotNil(t, response.RedirectTo, "%s", body)

		http.Redirect(w, r, response.RedirectTo, http.StatusFound)
	}

	router.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NotEmpty(t, *remote)
		require.NotEmpty(t, *subject)

		challenge := r.URL.Query().Get("login_challenge")
		require.NotEmpty(t, challenge)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&p{Subject: *subject}))
		href := urlx.MustJoin(*remote, "/oauth2/auth/requests/login/accept") + "?login_challenge=" + challenge
		do(w, r, href, &b)
	})

	router.GET("/consent", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NotEmpty(t, *remote)
		require.NotNil(t, *scope)

		challenge := r.URL.Query().Get("consent_challenge")
		require.NotEmpty(t, challenge)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&p{GrantScope: *scope}))
		href := urlx.MustJoin(*remote, "/oauth2/auth/requests/consent/accept") + "?consent_challenge=" + challenge
		do(w, r, href, &b)
	})

	if addr == "" {
		server := httptest.NewServer(router)
		t.Cleanup(server.Close)
		return server.Config, server.URL
	}

	parsed, err := url.ParseRequestURI(addr)
	require.NoError(t, err)

	server := &http.Server{Addr: ":" + parsed.Port(), Handler: router}
	go func(t *testing.T) {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			require.NoError(t, err)
		} else if err == nil {
			require.NoError(t, server.Close())
		}
	}(t)
	t.Cleanup(func() {
		require.NoError(t, server.Close())
	})
	return server, addr
}

func newReturnTs(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		require.NoError(t, err)
		require.Empty(t, sess.Identity.Credentials)
		reg.Writer().Write(w, r, sess)
	}))
	viper.Set(configuration.ViperKeyURLsDefaultReturnTo, ts.URL)
	t.Cleanup(ts.Close)
	return ts
}

func newUI(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var e interface{}
		var err error
		if r.URL.Path == "/login" {
			e, err = reg.LoginRequestPersister().GetLoginRequest(r.Context(), x.ParseUUID(r.URL.Query().Get("request")))
		} else if r.URL.Path == "/registration" {
			e, err = reg.RegistrationRequestPersister().GetRegistrationRequest(r.Context(), x.ParseUUID(r.URL.Query().Get("request")))
		} else if r.URL.Path == "/settings" {
			e, err = reg.SettingsRequestPersister().GetSettingsRequest(r.Context(), x.ParseUUID(r.URL.Query().Get("request")))
		}

		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	t.Cleanup(ts.Close)
	viper.Set(configuration.ViperKeyURLsLogin, ts.URL+"/login")
	viper.Set(configuration.ViperKeyURLsRegistration, ts.URL+"/registration")
	viper.Set(configuration.ViperKeyURLsSettings, ts.URL+"/settings")
	return ts
}

func newHydra(t *testing.T, subject *string, scope *[]string) (remoteAdmin, remotePublic, hydraIntegrationTSURL string) {
	remoteAdmin = os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_ADMIN")
	remotePublic = os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_PUBLIC")

	hydraIntegrationTS, hydraIntegrationTSURL := newHydraIntegration(t, &remoteAdmin, subject, scope, os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_INTEGRATION_ADDR"))
	t.Cleanup(func() {
		require.NoError(t, hydraIntegrationTS.Close())
	})

	if remotePublic == "" && remoteAdmin == "" {
		t.Logf("Environment did not provide ORY Hydra, starting fresh.")
		publicPort, err := freeport.GetFreePort()
		require.NoError(t, err)

		pool, err := dockertest.NewPool("")
		require.NoError(t, err)
		hydra, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "oryd/hydra",
			Tag:        "v1.4.10",
			Env: []string{
				"DSN=memory",
				fmt.Sprintf("URLS_SELF_ISSUER=http://127.0.0.1:%d/", publicPort),
				"URLS_LOGIN=" + hydraIntegrationTSURL + "/login",
				"URLS_CONSENT=" + hydraIntegrationTSURL + "/consent",
			},
			Cmd:          []string{"serve", "all", "--dangerous-force-http"},
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

		remotePublic = "http://127.0.0.1:" + hydra.GetPort("4444/tcp")
		remoteAdmin = "http://127.0.0.1:" + hydra.GetPort("4445/tcp")
	}

	t.Logf("ORY Hydra running at: %s %s", remotePublic, remoteAdmin)

	return remoteAdmin, remotePublic, hydraIntegrationTSURL
}

func newOIDCProvider(
	t *testing.T,
	kratos *httptest.Server,
	hydraPublic string,
	hydraAdmin string,
	id, clientID string,
) oidc.Configuration {
	createClient(t, hydraAdmin, kratos.URL+oidc.BasePath+"/callback/"+id, clientID)

	return oidc.Configuration{
		Provider:     "generic",
		ID:           id,
		ClientID:     clientID,
		ClientSecret: "secret",
		IssuerURL:    hydraPublic + "/",
		Mapper:       "file://./stub/oidc.hydra.jsonnet",
	}
}

func viperSetProviderConfig(providers ...oidc.Configuration) {
	viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC),
		map[string]interface{}{"config": &oidc.ConfigurationCollection{Providers: providers}})
}

func newClient(t *testing.T, jar *cookiejar.Jar) *http.Client {
	if jar == nil {
		j, err := cookiejar.New(nil)
		jar = j
		require.NoError(t, err)
	}
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if debugRedirects {
				t.Logf("Redirect: %s", req.URL.String())
			}
			if len(via) >= 20 {
				for k, v := range via {
					t.Logf("Failed with redirect (%d): %s", k, v.URL.String())
				}
				return errors.New("stopped after 20 redirects")
			}
			return nil
		},
	}
}

// AssertSystemError asserts an error ui response
func AssertSystemError(t *testing.T, errTS *httptest.Server, res *http.Response, body []byte, code int, reason string) {
	require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

	assert.Equal(t, int64(code), gjson.GetBytes(body, "0.code").Int(), "%s", body)
	assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), reason, "%s", body)
}
