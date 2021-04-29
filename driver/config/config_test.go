package config_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/ory/x/configx"

	"github.com/sirupsen/logrus/hooks/test"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/urlx"

	_ "github.com/ory/jsonschema/v3/fileloader"

	"github.com/ory/kratos/driver/config"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViperProvider(t *testing.T) {
	t.Run("suite=loaders", func(t *testing.T) {
		p := config.MustNew(t, logrusx.New("", ""),
			configx.WithConfigFiles("../../internal/.kratos.yaml"))

		t.Run("group=urls", func(t *testing.T) {
			assert.Equal(t, "http://test.kratos.ory.sh/login", p.SelfServiceFlowLoginUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/settings", p.SelfServiceFlowSettingsUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/register", p.SelfServiceFlowRegistrationUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/error", p.SelfServiceFlowErrorURL().String())

			assert.Equal(t, "http://admin.kratos.ory.sh", p.SelfAdminURL().String())
			assert.Equal(t, "http://public.kratos.ory.sh", p.SelfPublicURL(nil).String())

			var ds []string
			for _, v := range p.SelfServiceBrowserWhitelistedReturnToDomains() {
				ds = append(ds, v.String())
			}

			assert.Equal(t, []string{
				"http://return-to-1-test.ory.sh/",
				"http://return-to-2-test.ory.sh/",
			}, ds)

			pWithFragments := config.MustNew(t, logrusx.New("", ""),
				configx.WithValues(map[string]interface{}{
					config.ViperKeySelfServiceLoginUI:        "http://test.kratos.ory.sh/#/login",
					config.ViperKeySelfServiceSettingsURL:    "http://test.kratos.ory.sh/#/settings",
					config.ViperKeySelfServiceRegistrationUI: "http://test.kratos.ory.sh/#/register",
					config.ViperKeySelfServiceErrorUI:        "http://test.kratos.ory.sh/#/error",
				}),
				configx.SkipValidation(),
			)

			assert.Equal(t, "http://test.kratos.ory.sh/#/login", pWithFragments.SelfServiceFlowLoginUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/#/settings", pWithFragments.SelfServiceFlowSettingsUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/#/register", pWithFragments.SelfServiceFlowRegistrationUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/#/error", pWithFragments.SelfServiceFlowErrorURL().String())

			for _, v := range []string{
				"#/login",
				"/login",
				"/",
				"test.kratos.ory.sh/login",
			} {

				logger := logrusx.New("", "")
				logger.Logger.ExitFunc = func(code int) { panic("") }
				hook := new(test.Hook)
				logger.Logger.Hooks.Add(hook)

				pWithIncorrectUrls := config.MustNew(t, logger,
					configx.WithValues(map[string]interface{}{
						config.ViperKeySelfServiceLoginUI: v,
					}),
					configx.SkipValidation(),
				)

				assert.Panics(t, func() { pWithIncorrectUrls.SelfServiceFlowLoginUI() })

				assert.Equal(t, logrus.FatalLevel, hook.LastEntry().Level)
				assert.Equal(t, "Configuration value from key selfservice.flows.login.ui_url is not a valid URL: "+v, hook.LastEntry().Message)
				assert.Equal(t, 1, len(hook.Entries))
			}
		})

		t.Run("group=default_return_to", func(t *testing.T) {
			assert.Equal(t, "https://self-service/login/password/return_to", p.SelfServiceFlowLoginReturnTo("password").String())
			assert.Equal(t, "https://self-service/login/return_to", p.SelfServiceFlowLoginReturnTo("oidc").String())

			assert.Equal(t, "https://self-service/registration/return_to", p.SelfServiceFlowRegistrationReturnTo("password").String())
			assert.Equal(t, "https://self-service/registration/oidc/return_to", p.SelfServiceFlowRegistrationReturnTo("oidc").String())

			assert.Equal(t, "https://self-service/settings/password/return_to", p.SelfServiceFlowSettingsReturnTo("password", p.SelfServiceBrowserDefaultReturnTo()).String())
			assert.Equal(t, "https://self-service/settings/return_to", p.SelfServiceFlowSettingsReturnTo("profile", p.SelfServiceBrowserDefaultReturnTo()).String())

			assert.Equal(t, "http://test.kratos.ory.sh:4000/", p.SelfServiceFlowLogoutRedirectURL().String())
			p.MustSet(config.ViperKeySelfServiceLogoutBrowserDefaultReturnTo, "")
			assert.Equal(t, "http://return-to-3-test.ory.sh/", p.SelfServiceFlowLogoutRedirectURL().String())
		})

		t.Run("group=identity", func(t *testing.T) {
			assert.Equal(t, "http://test.kratos.ory.sh/default-identity.schema.json", p.DefaultIdentityTraitsSchemaURL().String())

			ss := p.IdentityTraitsSchemas()
			assert.Equal(t, 2, len(ss))

			assert.Contains(t, ss, config.Schema{
				ID:  "default",
				URL: "http://test.kratos.ory.sh/default-identity.schema.json",
			})
			assert.Contains(t, ss, config.Schema{
				ID:  "other",
				URL: "http://test.kratos.ory.sh/other-identity.schema.json",
			})
		})

		t.Run("group=serve", func(t *testing.T) {
			assert.Equal(t, "admin.kratos.ory.sh:1234", p.AdminListenOn())
			assert.Equal(t, "public.kratos.ory.sh:1235", p.PublicListenOn())
		})

		t.Run("group=dsn", func(t *testing.T) {
			assert.Equal(t, "sqlite://foo.db?mode=memory&_fk=true", p.DSN())
		})

		t.Run("group=secrets", func(t *testing.T) {
			assert.Equal(t, [][]byte{
				[]byte("session-key-7f8a9b77-1"),
				[]byte("session-key-7f8a9b77-2"),
			}, p.SecretsSession())
		})

		t.Run("group=methods", func(t *testing.T) {
			for _, tc := range []struct {
				id      string
				config  string
				enabled bool
			}{
				{id: "password", enabled: true, config: `{"ignore_network_errors":true,"max_breaches":0}`},
				{id: "oidc", enabled: true, config: `{"providers":[{"client_id":"a","client_secret":"b","id":"github","provider":"github","mapper_url":"http://test.kratos.ory.sh/default-identity.schema.json"}]}`},
			} {
				strategy := p.SelfServiceStrategy(tc.id)
				assert.Equal(t, tc.enabled, strategy.Enabled)
				assert.JSONEq(t, tc.config, string(strategy.Config))
			}
		})

		t.Run("method=registration", func(t *testing.T) {
			assert.Equal(t, time.Minute*98, p.SelfServiceFlowRegistrationRequestLifespan())

			t.Run("hook=before", func(t *testing.T) {
				hook := p.SelfServiceFlowRegistrationBeforeHooks()
				require.Len(t, hook, 0)
				// assert.EqualValues(t, "redirect", hook.Name)
				// assert.JSONEq(t, `{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`, string(hook.Config))
			})

			for _, tc := range []struct {
				strategy string
				hooks    []config.SelfServiceHook
			}{
				{
					strategy: "password",
					hooks: []config.SelfServiceHook{
						{Name: "session", Config: json.RawMessage(`{}`)},
						{Name: "web-hook", Config: json.RawMessage(`{"method":"POST","url":"https://test.kratos.ory.sh/after_registration_password_hook"}`)},
						// {Name: "verify", Config: json.RawMessage(`{}`)},
						// {Name: "redirect", Config: json.RawMessage(`{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`)},
					},
				},
				{
					strategy: "oidc",
					hooks: []config.SelfServiceHook{
						// {Name: "verify", Config: json.RawMessage(`{}`)},
						{Name: "web-hook", Config: json.RawMessage(`{"method":"GET","url":"https://test.kratos.ory.sh/after_registration_oidc_hook"}`)},
						{Name: "session", Config: json.RawMessage(`{}`)},
						// {Name: "redirect", Config: json.RawMessage(`{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`)},
					},
				},
				{
					strategy: "global",
					hooks: []config.SelfServiceHook{
						{Name: "web-hook", Config: json.RawMessage(`{"method":"POST","url":"https://test.kratos.ory.sh/after_registration_global_hook"}`)},
					},
				},
			} {
				t.Run("hook=after/strategy="+tc.strategy, func(t *testing.T) {
					hooks := p.SelfServiceFlowRegistrationAfterHooks(tc.strategy)
					assert.Equal(t, tc.hooks, hooks)
				})
			}
		})

		t.Run("method=login", func(t *testing.T) {
			assert.Equal(t, time.Minute*99, p.SelfServiceFlowLoginRequestLifespan())

			t.Run("hook=before", func(t *testing.T) {
				hook := p.SelfServiceFlowLoginBeforeHooks()
				require.Len(t, hook, 0)
				// assert.EqualValues(t, "redirect", hook.Name)
				// assert.JSONEq(t, `{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`, string(hook.Config))
			})

			for _, tc := range []struct {
				strategy string
				hooks    []config.SelfServiceHook
			}{
				{
					strategy: "password",
					hooks: []config.SelfServiceHook{
						{Name: "revoke_active_sessions", Config: json.RawMessage(`{}`)},
						{Name: "web-hook", Config: json.RawMessage(`{"method":"POST","url":"https://test.kratos.ory.sh/after_login_password_hook"}`)},
					},
				},
				{
					strategy: "oidc",
					hooks: []config.SelfServiceHook{
						{Name: "web-hook", Config: json.RawMessage(`{"method":"GET","url":"https://test.kratos.ory.sh/after_login_oidc_hook"}`)},
						{Name: "revoke_active_sessions", Config: json.RawMessage(`{}`)},
					},
				},
				{
					strategy: "global",
					hooks: []config.SelfServiceHook{
						{Name: "web-hook", Config: json.RawMessage(`{"method":"POST","url":"https://test.kratos.ory.sh/after_login_global_hook"}`)},
					},
				},
			} {
				t.Run("hook=after/strategy="+tc.strategy, func(t *testing.T) {
					hooks := p.SelfServiceFlowLoginAfterHooks(tc.strategy)
					assert.Equal(t, tc.hooks, hooks)
				})
			}
		})

		t.Run("method=settings", func(t *testing.T) {
			assert.Equal(t, time.Minute*99, p.SelfServiceFlowSettingsFlowLifespan())
			assert.Equal(t, time.Minute*5, p.SelfServiceFlowSettingsPrivilegedSessionMaxAge())

			t.Run("hook=before", func(t *testing.T) {
				hook := p.SelfServiceFlowLoginBeforeHooks()
				require.Len(t, hook, 0)
				// assert.EqualValues(t, "redirect", hook.Name)
				// assert.JSONEq(t, `{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`, string(hook.Config))
			})

			for _, tc := range []struct {
				strategy string
				hooks    []config.SelfServiceHook
			}{
				{
					strategy: "password",
					hooks: []config.SelfServiceHook{
						{Name: "web-hook", Config: json.RawMessage(`{"method":"POST","url":"https://test.kratos.ory.sh/after_settings_password_hook"}`)},
					},
				},
				{
					strategy: "profile",
					hooks: []config.SelfServiceHook{
						{Name: "web-hook", Config: json.RawMessage(`{"method":"POST","url":"https://test.kratos.ory.sh/after_settings_profile_hook"}`)},
					},
				},
				{
					strategy: "global",
					hooks: []config.SelfServiceHook{
						{Name: "web-hook", Config: json.RawMessage(`{"method":"POST","url":"https://test.kratos.ory.sh/after_settings_global_hook"}`)},
					},
				},
			} {
				t.Run("hook=after/strategy="+tc.strategy, func(t *testing.T) {
					hooks := p.SelfServiceFlowSettingsAfterHooks(tc.strategy)
					assert.Equal(t, tc.hooks, hooks)
				})
			}
		})

		t.Run("method=recovery", func(t *testing.T) {
			assert.Equal(t, true, p.SelfServiceFlowRecoveryEnabled())
			assert.Equal(t, time.Minute*98, p.SelfServiceFlowRecoveryRequestLifespan())
			assert.Equal(t, "http://test.kratos.ory.sh/recovery", p.SelfServiceFlowRecoveryUI().String())

			hooks := p.SelfServiceFlowRecoveryAfterHooks("global")
			assert.Equal(t, []config.SelfServiceHook{{Name: "web-hook", Config: json.RawMessage(`{"method":"GET","url":"https://test.kratos.ory.sh/after_recovery_hook"}`)}}, hooks)
		})

		t.Run("method=verification", func(t *testing.T) {
			assert.Equal(t, time.Minute*97, p.SelfServiceFlowVerificationRequestLifespan())
			assert.Equal(t, "http://test.kratos.ory.sh/verification", p.SelfServiceFlowVerificationUI().String())

			hooks := p.SelfServiceFlowVerificationAfterHooks("global")
			assert.Equal(t, []config.SelfServiceHook{{Name: "web-hook", Config: json.RawMessage(`{"method":"GET","url":"https://test.kratos.ory.sh/after_verification_hook"}`)}}, hooks)
		})

		t.Run("group=hashers", func(t *testing.T) {
			c := p.HasherArgon2()
			assert.Equal(t, &config.Argon2{Memory: 1048576, Iterations: 2, Parallelism: 4,
				SaltLength: 16, KeyLength: 32, DedicatedMemory: config.Argon2DefaultDedicatedMemory, ExpectedDeviation: config.Argon2DefaultDeviation, ExpectedDuration: config.Argon2DefaultDuration}, c)
		})

		t.Run("group=set_provider_by_json", func(t *testing.T) {
			providerConfigJSON := `{"providers": [{"id":"github-test","provider":"github","client_id":"set_json_test","client_secret":"secret","mapper_url":"http://mapper-url","scope":["user:email"]}]}`
			strategyConfigJSON := fmt.Sprintf(`{"enabled":true, "config": %s}`, providerConfigJSON)

			p.MustSet(config.ViperKeySelfServiceStrategyConfig+".oidc", strategyConfigJSON)
			strategy := p.SelfServiceStrategy("oidc")
			assert.JSONEq(t, providerConfigJSON, string(strategy.Config))
		})
	})
}

type InterceptHook struct {
	lastEntry *logrus.Entry
}

func (l InterceptHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.FatalLevel}
}

func (l InterceptHook) Fire(e *logrus.Entry) error {
	l.lastEntry = e
	return nil
}

func TestBcrypt(t *testing.T) {
	p := config.MustNew(t, logrusx.New("", ""), configx.SkipValidation())

	require.NoError(t, p.Set(config.ViperKeyHasherBcryptCost, 4))
	require.NoError(t, p.Set("dev", false))
	assert.EqualValues(t, uint32(12), p.HasherBcrypt().Cost)

	require.NoError(t, p.Set("dev", true))
	assert.EqualValues(t, uint32(4), p.HasherBcrypt().Cost)
}

func TestProviderBaseURLs(t *testing.T) {
	machineHostname, err := os.Hostname()
	if err != nil {
		machineHostname = "127.0.0.1"
	}

	p := config.MustNew(t, logrusx.New("", ""), configx.SkipValidation())
	assert.Equal(t, "https://"+machineHostname+":4433/", p.SelfPublicURL(nil).String())
	assert.Equal(t, "https://"+machineHostname+":4434/", p.SelfAdminURL().String())

	p.MustSet(config.ViperKeyPublicPort, 4444)
	p.MustSet(config.ViperKeyAdminPort, 4445)
	assert.Equal(t, "https://"+machineHostname+":4444/", p.SelfPublicURL(nil).String())
	assert.Equal(t, "https://"+machineHostname+":4445/", p.SelfAdminURL().String())

	p.MustSet(config.ViperKeyPublicHost, "public.ory.sh")
	p.MustSet(config.ViperKeyAdminHost, "admin.ory.sh")
	assert.Equal(t, "https://public.ory.sh:4444/", p.SelfPublicURL(nil).String())
	assert.Equal(t, "https://admin.ory.sh:4445/", p.SelfAdminURL().String())

	// Set to dev mode
	p.MustSet("dev", true)
	assert.Equal(t, "http://public.ory.sh:4444/", p.SelfPublicURL(nil).String())
	assert.Equal(t, "http://admin.ory.sh:4445/", p.SelfAdminURL().String())

	// Check domain aliases
	p.MustSet(config.ViperKeyPublicDomainAliases, []config.DomainAlias{
		{
			MatchDomain: "www.google.com",
			BasePath:    "/.ory/",
			Scheme:      "https",
		},
		{
			MatchDomain: "www.amazon.com",
			BasePath:    "/",
			Scheme:      "http",
		},
		{
			MatchDomain: "ory.sh:1234",
			BasePath:    "/",
			Scheme:      "https",
		},
	})
	assert.Equal(t, "http://public.ory.sh:4444/", p.SelfPublicURL(nil).String())
	assert.Equal(t, "http://public.ory.sh:4444/", p.SelfPublicURL(&http.Request{
		URL:  new(url.URL),
		Host: "www.not-google.com",
	}).String())
	assert.Equal(t, "https://www.GooGle.com:312/.ory/", p.SelfPublicURL(&http.Request{
		URL:  new(url.URL),
		Host: "www.GooGle.com:312",
	}).String())
	assert.Equal(t, "http://www.amazon.com/", p.SelfPublicURL(&http.Request{
		URL:  new(url.URL),
		Host: "www.amazon.com",
	}).String())

	// Check domain aliases with alias query param
	assert.Equal(t, "http://www.amazon.com/", p.SelfPublicURL(&http.Request{
		URL:  &url.URL{RawQuery: url.Values{"alias": {"www.amazon.com"}}.Encode()},
		Host: "www.GooGle.com:312",
	}).String())
	assert.Equal(t, "https://ory.sh:1234/", p.SelfPublicURL(&http.Request{
		URL:  &url.URL{RawQuery: url.Values{"alias": {"ory.sh:1234"}}.Encode()},
		Host: "www.amazon.com",
	}).String())
	assert.Equal(t, "http://www.amazon.com:8181/", p.SelfPublicURL(&http.Request{
		URL:  new(url.URL),
		Host: "www.amazon.com:8181",
	}).String())
	assert.Equal(t, "http://www.amazon.com:8181/", p.SelfPublicURL(&http.Request{
		URL:  &url.URL{RawQuery: url.Values{"alias": {"www.amazon.com:8181"}}.Encode()},
		Host: "www.GooGle.com:312",
	}).String())
}

func TestViperProvider_Secrets(t *testing.T) {
	p := config.MustNew(t, logrusx.New("", ""), configx.SkipValidation())

	def := p.SecretsDefault()
	assert.NotEmpty(t, def)
	assert.Equal(t, def, p.SecretsSession())
	assert.Equal(t, def, p.SecretsDefault())
}

func TestViperProvider_Defaults(t *testing.T) {
	l := logrusx.New("", "")

	for k, tc := range []struct {
		init   func() *config.Config
		expect func(t *testing.T, p *config.Config)
	}{
		{
			init: func() *config.Config {
				return config.MustNew(t, l, configx.SkipValidation())
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l, configx.WithConfigFiles("stub/.defaults.yml"), configx.SkipValidation())
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l, configx.WithConfigFiles("stub/.defaults-password.yml"), configx.SkipValidation())
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l, configx.WithConfigFiles("../../test/e2e/profiles/recovery/.kratos.yml"), configx.SkipValidation())
			},
			expect: func(t *testing.T, p *config.Config) {
				assert.True(t, p.SelfServiceFlowRecoveryEnabled())
				assert.False(t, p.SelfServiceFlowVerificationEnabled())
				assert.True(t, p.SelfServiceStrategy("password").Enabled)
				assert.True(t, p.SelfServiceStrategy("profile").Enabled)
				assert.True(t, p.SelfServiceStrategy("link").Enabled)
				assert.False(t, p.SelfServiceStrategy("oidc").Enabled)
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l, configx.WithConfigFiles("../../test/e2e/profiles/verification/.kratos.yml"), configx.SkipValidation())
			},
			expect: func(t *testing.T, p *config.Config) {
				assert.False(t, p.SelfServiceFlowRecoveryEnabled())
				assert.True(t, p.SelfServiceFlowVerificationEnabled())
				assert.True(t, p.SelfServiceStrategy("password").Enabled)
				assert.True(t, p.SelfServiceStrategy("profile").Enabled)
				assert.True(t, p.SelfServiceStrategy("link").Enabled)
				assert.False(t, p.SelfServiceStrategy("oidc").Enabled)
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l, configx.WithConfigFiles("../../test/e2e/profiles/oidc/.kratos.yml"), configx.SkipValidation())
			},
			expect: func(t *testing.T, p *config.Config) {
				assert.False(t, p.SelfServiceFlowRecoveryEnabled())
				assert.False(t, p.SelfServiceFlowVerificationEnabled())
				assert.True(t, p.SelfServiceStrategy("password").Enabled)
				assert.True(t, p.SelfServiceStrategy("profile").Enabled)
				assert.True(t, p.SelfServiceStrategy("link").Enabled)
				assert.True(t, p.SelfServiceStrategy("oidc").Enabled)
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			p := tc.init()

			if tc.expect != nil {
				tc.expect(t, p)
				return
			}
			assert.False(t, p.SelfServiceFlowRecoveryEnabled())
			assert.False(t, p.SelfServiceFlowVerificationEnabled())
			assert.True(t, p.SelfServiceStrategy("password").Enabled)
			assert.True(t, p.SelfServiceStrategy("profile").Enabled)
			assert.True(t, p.SelfServiceStrategy("link").Enabled)
			assert.False(t, p.SelfServiceStrategy("oidc").Enabled)
		})
	}

	t.Run("suite=ui_url", func(t *testing.T) {
		p := config.MustNew(t, l, configx.SkipValidation())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/login", p.SelfServiceFlowLoginUI().String())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/settings", p.SelfServiceFlowSettingsUI().String())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/registration", p.SelfServiceFlowRegistrationUI().String())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/recovery", p.SelfServiceFlowRecoveryUI().String())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/verification", p.SelfServiceFlowVerificationUI().String())
	})
}

func TestViperProvider_ReturnTo(t *testing.T) {
	l := logrusx.New("", "")
	p := config.MustNew(t, l, configx.SkipValidation())

	p.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	assert.Equal(t, "https://www.ory.sh/", p.SelfServiceFlowVerificationReturnTo(urlx.ParseOrPanic("https://www.ory.sh/")).String())
	assert.Equal(t, "https://www.ory.sh/", p.SelfServiceFlowRecoveryReturnTo().String())

	p.MustSet(config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, "https://www.ory.sh/recovery")
	assert.Equal(t, "https://www.ory.sh/recovery", p.SelfServiceFlowRecoveryReturnTo().String())

	p.MustSet(config.ViperKeySelfServiceVerificationBrowserDefaultReturnTo, "https://www.ory.sh/verification")
	assert.Equal(t, "https://www.ory.sh/verification", p.SelfServiceFlowVerificationReturnTo(urlx.ParseOrPanic("https://www.ory.sh/")).String())
}

func TestViperProvider_DSN(t *testing.T) {
	t.Run("case=dsn: memory", func(t *testing.T) {
		p := config.MustNew(t, logrusx.New("", ""), configx.SkipValidation())
		p.MustSet(config.ViperKeyDSN, "memory")

		assert.Equal(t, config.DefaultSQLiteMemoryDSN, p.DSN())
	})

	t.Run("case=dsn: not memory", func(t *testing.T) {
		p := config.MustNew(t, logrusx.New("", ""), configx.SkipValidation())

		dsn := "sqlite://foo.db?_fk=true"
		p.MustSet(config.ViperKeyDSN, dsn)

		assert.Equal(t, dsn, p.DSN())
	})

	t.Run("case=dsn: not set", func(t *testing.T) {
		dsn := ""

		var exitCode int
		l := logrusx.New("", "", logrusx.WithExitFunc(func(i int) {
			exitCode = i
		}), logrusx.WithHook(InterceptHook{}))
		p := config.MustNew(t, l, configx.SkipValidation())

		assert.Equal(t, dsn, p.DSN())
		assert.NotEqual(t, 0, exitCode)
	})
}

func TestViperProvider_ParseURIOrFail(t *testing.T) {
	var exitCode int

	l := logrusx.New("", "", logrusx.WithExitFunc(func(i int) {
		exitCode = i
	}))
	p := config.MustNew(t, l, configx.SkipValidation())
	require.Zero(t, exitCode)

	const testKey = "testKeyNotUsedInTheRealSchema"

	for _, tc := range []struct {
		u        string
		expected url.URL
	}{
		{
			u: "file:///etc/config/kratos/identity.schema.json",
			expected: url.URL{
				Scheme: "file",
				Path:   "/etc/config/kratos/identity.schema.json",
			},
		},
		{
			u: "file://./identity.schema.json",
			expected: url.URL{
				Scheme: "file",
				Host:   ".",
				Path:   "/identity.schema.json",
			},
		},
		{
			u: "base64://bG9jYWwgc3ViamVjdCA9I",
			expected: url.URL{
				Scheme: "base64",
				Host:   "bG9jYWwgc3ViamVjdCA9I",
			},
		},
		{
			u: "https://foo.bar/schema.json",
			expected: url.URL{
				Scheme: "https",
				Host:   "foo.bar",
				Path:   "/schema.json",
			},
		},
	} {
		t.Run("case=parse "+tc.u, func(t *testing.T) {
			require.NoError(t, p.Set(testKey, tc.u))

			u := p.ParseURIOrFail(testKey)
			require.Zero(t, exitCode)
			assert.Equal(t, tc.expected, *u)
		})
	}
}
