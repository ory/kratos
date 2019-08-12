package oidc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"
	"github.com/ory/x/urlx"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"

	"github.com/phayes/freeport"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
	"github.com/ory/hive/selfservice"
	. "github.com/ory/hive/selfservice/oidc"
	"github.com/ory/hive/selfservice/password"
	"github.com/ory/hive/x"
)

func hookConfig(u string) (m []map[string]interface{}) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, `[
	{
		"run": "session"
	},
	{
		"run": "redirect",
		"config": {
          "default_redirect_url": "%s",
          "allow_user_defined_redirect": true
		}
	}
]`, u); err != nil {
		panic(err)
	}

	if err := json.NewDecoder(&b).Decode(&m); err != nil {
		panic(err)
	}

	return m
}

func TestStrategy(t *testing.T) {

	var (
		subject      string
		scope        []string
		remoteAdmin  = os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_ADMIN")
		remotePublic = os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_PUBLIC")
	)

	hydraIntegrationTS, hydraIntegrationTSURL := newHydraIntegration(t, &remoteAdmin, &subject, &scope, os.Getenv("TEST_SELFSERVICE_OIDC_HYDRA_INTEGRATION_ADDR"))
	defer hydraIntegrationTS.Close()

	if testing.Short() {
		t.Skip()
	}

	if remotePublic == "" && remoteAdmin == "" {
		publicPort, err := freeport.GetFreePort()
		require.NoError(t, err)

		pool, err := dockertest.NewPool("")
		require.NoError(t, err)
		hydra, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "oryd/hydra",
			Tag:        "v1.0.0",
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
		defer hydra.Close()
		require.NoError(t, hydra.Expire(uint(60*15)))

		require.NotEmpty(t, hydra.GetPort("4444/tcp"), "%+v", hydra.Container.NetworkSettings.Ports)
		require.NotEmpty(t, hydra.GetPort("4445/tcp"), "%+v", hydra.Container)

		remotePublic = "http://127.0.0.1:" + hydra.GetPort("4444/tcp")
		remoteAdmin = "http://127.0.0.1:" + hydra.GetPort("4445/tcp")
	}

	_, reg := internal.NewMemoryRegistry(t)
	reg.SelfServiceStrategies()[0].(*password.Strategy).WithTokenGenerator(func(r *http.Request) string {
		return "nosurf"
	}) // We need to replace the password strategy token generator because it is being used by the error handler...

	s := reg.SelfServiceStrategies()[1].(*Strategy).WithTokenGenerator(func(r *http.Request) string {
		return "nosurf"
	})

	router := x.NewRouterPublic()
	s.SetRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	returnTS := newReturnTs(t, reg)
	defer returnTS.Close()

	uiTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var e interface{}
		var err error
		if r.URL.Path == "/login" {
			e, err = reg.LoginRequestManager().GetLoginRequest(r.Context(), r.URL.Query().Get("request"))
		} else if r.URL.Path == "/registration" {
			e, err = reg.RegistrationRequestManager().GetRegistrationRequest(r.Context(), r.URL.Query().Get("request"))
		}

		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	defer uiTS.Close()

	createClient(t, remoteAdmin, ts.URL+BasePath+"/callback/valid")

	cb := map[string]interface{}{
		"config": &ConfigurationCollection{
			Providers: []Configuration{
				{
					Provider:     "generic",
					ID:           "valid",
					ClientID:     "client",
					ClientSecret: "secret",
					IssuerURL:    remotePublic + "/",
					SchemaURL:    "file://./stub/hydra.schema.json",
				},
				{
					Provider:     "generic",
					ID:           "invalid-issuer",
					ClientID:     "client",
					ClientSecret: "secret",
					IssuerURL:    strings.Replace(remotePublic, "127.0.0.1", "localhost", 1) + "/",
					SchemaURL:    "file://./stub/hydra.schema.json",
				},
			},
		},
	}

	errTS := newErrTs(t, reg)
	defer errTS.Close()

	viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC), cb)
	viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
	viper.Set(configuration.ViperKeyURLsError, errTS.URL)
	viper.Set(configuration.ViperKeyURLsLogin, uiTS.URL+"/login")
	viper.Set(configuration.ViperKeyURLsRegistration, uiTS.URL+"/registration")
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")
	viper.Set(configuration.ViperKeySelfServiceRegistrationAfterConfig+"."+string(identity.CredentialsTypeOIDC), hookConfig(returnTS.URL))
	viper.Set(configuration.ViperKeySelfServiceLoginAfterConfig+"."+string(identity.CredentialsTypeOIDC), hookConfig(returnTS.URL))
	// viper.Set(configuration.ViperKeySignupDefaultReturnToURL, returnTS.URL)
	// viper.Set(configuration.ViperKeyAuthnDefaultReturnToURL, returnTS.URL)

	t.Logf("Hive Public URL: %s", ts.URL)
	t.Logf("Hive Error URL: %s", errTS.URL)
	t.Logf("Hydra Public URL: %s", remotePublic)
	t.Logf("Hydra Admin URL: %s", remoteAdmin)
	t.Logf("Hydra Integration URL: %s", hydraIntegrationTSURL)
	t.Logf("Return URL: %s", returnTS.URL)

	var newClient = func(t *testing.T) *http.Client {
		jar, err := cookiejar.New(nil)
		require.NoError(t, err)
		return &http.Client{
			Jar: jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				t.Logf("Redirect: %s", req.URL.String())
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

	subject = "foo@bar.com"
	scope = []string{}

	// make request
	var mr = func(t *testing.T, provider, request string, fv url.Values) (*http.Response, []byte) {
		fv.Set("provider", provider)
		res, err := newClient(t).PostForm(ts.URL+BasePath+"/auth/"+request, fv)
		require.NoError(t, err)

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)

		require.Equal(t, 200, res.StatusCode, "%s\n\t%s", res.Request.URL.String(), body)

		return res, body
	}

	// assert system error (redirect to error endpoint)
	var ase = func(t *testing.T, res *http.Response, body []byte, code int, reason string) {
		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

		assert.Equal(t, int64(code), gjson.GetBytes(body, "0.code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), reason, "%s", body)
	}

	// assert ui error (redirect to login/registration ui endpoint)
	var aue = func(t *testing.T, res *http.Response, body []byte, reason string) {
		require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.oidc.config.errors.0.message").String(), reason, "%s", body)
	}

	// assert identity (success)
	var ai = func(t *testing.T, res *http.Response, body []byte) {
		assert.Contains(t, res.Request.URL.String(), returnTS.URL)
		assert.Equal(t, subject, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
	}

	// new login request
	var nlr = func(t *testing.T, redirectTo string, exp time.Duration) *selfservice.LoginRequest {
		r := selfservice.NewLoginRequest(exp, &http.Request{
			URL:    urlx.ParseOrPanic(redirectTo),
			Header: map[string][]string{},
		})
		r.Methods[identity.CredentialsTypeOIDC] = &selfservice.DefaultRequestMethod{
			Method: identity.CredentialsTypeOIDC,
			Config: NewRequestMethodConfig(),
		}
		require.NoError(t, reg.LoginRequestManager().CreateLoginRequest(context.Background(), r))
		return r
	}

	// new registration request
	var nrr = func(t *testing.T, redirectTo string, exp time.Duration) *selfservice.RegistrationRequest {
		r := selfservice.NewRegistrationRequest(exp, &http.Request{
			URL:    urlx.ParseOrPanic(redirectTo),
			Header: map[string][]string{},
		})
		r.Methods[identity.CredentialsTypeOIDC] = &selfservice.DefaultRequestMethod{
			Method: identity.CredentialsTypeOIDC,
			Config: NewRequestMethodConfig(),
		}
		require.NoError(t, reg.RegistrationRequestManager().CreateRegistrationRequest(context.Background(), r))
		return r
	}

	t.Run("case=should fail because provider does not exist", func(t *testing.T) {
		res, body := mr(t, "provider-does-not-exist", "request-does-not-exist", url.Values{})
		ase(t, res, body, http.StatusNotFound, "is unknown or has not been configured")
	})

	t.Run("case=should fail because the issuer is mismatching", func(t *testing.T) {
		res, body := mr(t, "invalid-issuer", "request-does-not-exist", url.Values{})
		ase(t, res, body, http.StatusInternalServerError, "issuer did not match the issuer returned by provider")
	})

	t.Run("case=should fail because request does not exist", func(t *testing.T) {
		res, body := mr(t, "valid", "request-does-not-exist", url.Values{})
		ase(t, res, body, http.StatusNotFound, "Unable to find request")
	})

	t.Run("case=should fail because the login request is expired", func(t *testing.T) {
		r := nlr(t, returnTS.URL, -time.Minute)
		res, body := mr(t, "valid", r.ID, url.Values{})
		aue(t, res, body, "login request expired")
	})

	t.Run("case=should fail because the registration request is expired", func(t *testing.T) {
		r := nrr(t, returnTS.URL, -time.Minute)
		res, body := mr(t, "valid", r.ID, url.Values{})
		aue(t, res, body, "registration request expired")
	})

	t.Run("case=should fail registration because scope was not provided", func(t *testing.T) {
		subject = "foo@bar.com"
		scope = []string{}

		r := nrr(t, returnTS.URL, time.Minute)
		res, body := mr(t, "valid", r.ID, url.Values{})
		aue(t, res, body, "no id_token was returned")
	})

	t.Run("case=should fail login because scope was not provided", func(t *testing.T) {
		r := nlr(t, returnTS.URL, time.Minute)
		res, body := mr(t, "valid", r.ID, url.Values{})
		aue(t, res, body, "no id_token was returned")
	})

	t.Run("case=should fail registration request because subject is not an email", func(t *testing.T) {
		subject = "not-an-email"
		scope = []string{"openid"}

		r := nrr(t, returnTS.URL, time.Minute)
		res, body := mr(t, "valid", r.ID, url.Values{})

		require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.oidc.config.fields.traits\\.subject.error").String(), "not match format 'email'", "%s", body)
	})

	t.Run("case=register and then login", func(t *testing.T) {
		subject = "register-then-login@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should pass registration", func(t *testing.T) {
			r := nrr(t, returnTS.URL, time.Minute)
			res, body := mr(t, "valid", r.ID, url.Values{})
			ai(t, res, body)
		})

		t.Run("case=should pass login", func(t *testing.T) {
			r := nlr(t, returnTS.URL, time.Minute)
			res, body := mr(t, "valid", r.ID, url.Values{})
			ai(t, res, body)
		})
	})

	t.Run("case=login without registered account", func(t *testing.T) {
		subject = "login-without-register@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should pass login", func(t *testing.T) {
			r := nlr(t, returnTS.URL, time.Minute)
			res, body := mr(t, "valid", r.ID, url.Values{})
			ai(t, res, body)
		})
	})

	t.Run("case=register and register again but login", func(t *testing.T) {
		subject = "register-twice@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should pass registration", func(t *testing.T) {
			r := nrr(t, returnTS.URL, time.Minute)
			res, body := mr(t, "valid", r.ID, url.Values{})
			ai(t, res, body)
		})

		t.Run("case=should pass second time registration", func(t *testing.T) {
			r := nlr(t, returnTS.URL, time.Minute)
			res, body := mr(t, "valid", r.ID, url.Values{})
			ai(t, res, body)
		})
	})

	t.Run("case=register and complete data", func(t *testing.T) {
		subject = "incomplete-data@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should fail registration on first attempt", func(t *testing.T) {
			r := nrr(t, returnTS.URL, time.Minute)
			res, body := mr(t, "valid", r.ID, url.Values{"traits.name": {"i"}})
			aue(t, res, body, "String length must be greater than or equal to 2")
			assert.Equal(t, "traits.name", gjson.GetBytes(body, "methods.oidc.config.fields.traits\\.name.name").String(), "%s", body) // make sure the field is being echoed
			assert.Equal(t, "i", gjson.GetBytes(body, "methods.oidc.config.fields.traits\\.name.value").String(), "%s", body)          // make sure the field is being echoed
		})

		t.Run("case=should pass registration with valid data", func(t *testing.T) {
			r := nrr(t, returnTS.URL, time.Minute)
			res, body := mr(t, "valid", r.ID, url.Values{"traits.name": {"valid-name"}})
			ai(t, res, body)
		})
	})

	t.Run("case=should fail to register if email is already being used by password credentials", func(t *testing.T) {
		subject = "email-exist-with-password-strategy@ory.sh"
		scope = []string{"openid"}

		t.Run("case=create password identity", func(t *testing.T) {
			i := identity.NewIdentity("")
			i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
				Identifiers: []string{subject},
			})

			_, err := reg.IdentityPool().Create(context.Background(), i)
			require.NoError(t, err)
		})

		t.Run("case=should fail registration", func(t *testing.T) {
			r := nrr(t, returnTS.URL, time.Minute)
			res, body := mr(t, "valid", r.ID, url.Values{})
			aue(t, res, body, "idk man")
		})
	})
}
