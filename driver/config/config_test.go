package config_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/ory/x/watcherx"

	"github.com/ory/kratos/x"

	"github.com/ory/kratos/internal/testhelpers"

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
		p := config.MustNew(t, logrusx.New("", ""), os.Stderr,
			configx.WithConfigFiles("stub/.kratos.yaml"))

		t.Run("group=urls", func(t *testing.T) {
			assert.Equal(t, "http://test.kratos.ory.sh/login", p.SelfServiceFlowLoginUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/settings", p.SelfServiceFlowSettingsUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/register", p.SelfServiceFlowRegistrationUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/error", p.SelfServiceFlowErrorURL().String())

			assert.Equal(t, "http://admin.kratos.ory.sh", p.SelfAdminURL().String())
			assert.Equal(t, "http://public.kratos.ory.sh", p.SelfPublicURL().String())

			var ds []string
			for _, v := range p.SelfServiceBrowserWhitelistedReturnToDomains() {
				ds = append(ds, v.String())
			}

			assert.Equal(t, []string{
				"http://return-to-1-test.ory.sh/",
				"http://return-to-2-test.ory.sh/",
				"http://*.wildcards.ory.sh",
				"/return-to-relative-test/",
			}, ds)

			pWithFragments := config.MustNew(t, logrusx.New("", ""),
				os.Stderr,
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

			pWithRelativeFragments := config.MustNew(t, logrusx.New("", ""),
				os.Stderr,
				configx.WithValues(map[string]interface{}{
					config.ViperKeySelfServiceLoginUI:        "/login",
					config.ViperKeySelfServiceSettingsURL:    "/settings",
					config.ViperKeySelfServiceRegistrationUI: "/register",
					config.ViperKeySelfServiceErrorUI:        "/error",
				}),
				configx.SkipValidation(),
			)

			assert.Equal(t, "/login", pWithRelativeFragments.SelfServiceFlowLoginUI().String())
			assert.Equal(t, "/settings", pWithRelativeFragments.SelfServiceFlowSettingsUI().String())
			assert.Equal(t, "/register", pWithRelativeFragments.SelfServiceFlowRegistrationUI().String())
			assert.Equal(t, "/error", pWithRelativeFragments.SelfServiceFlowErrorURL().String())

			for _, v := range []string{
				"#/login",
				"test.kratos.ory.sh/login",
			} {

				logger := logrusx.New("", "")
				logger.Logger.ExitFunc = func(code int) { panic("") }
				hook := new(test.Hook)
				logger.Logger.Hooks.Add(hook)

				pWithIncorrectUrls := config.MustNew(t, logger,
					os.Stderr,
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
			c := config.MustNew(t, logrusx.New("", ""), os.Stderr,
				configx.WithConfigFiles("stub/.kratos.mock.identities.yaml"),
				configx.SkipValidation())

			assert.Equal(t, "http://test.kratos.ory.sh/default-identity.schema.json", c.DefaultIdentityTraitsSchemaURL().String())

			ss := c.IdentityTraitsSchemas()
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
			var cipherExpected [32]byte
			for k, v := range []byte("secret-thirty-two-character-long") {
				cipherExpected[k] = byte(v)
			}
			assert.Equal(t, [][32]byte{
				cipherExpected,
			}, p.SecretsCipher())
		})

		t.Run("group=methods", func(t *testing.T) {
			for _, tc := range []struct {
				id      string
				config  string
				enabled bool
			}{
				{id: "password", enabled: true, config: `{"haveibeenpwned_host":"api.pwnedpasswords.com","haveibeenpwned_enabled":true,"ignore_network_errors":true,"max_breaches":0,"min_password_length":8,"identifier_similarity_check_enabled":true}`},
				{id: "oidc", enabled: true, config: `{"providers":[{"client_id":"a","client_secret":"b","id":"github","provider":"github","mapper_url":"http://test.kratos.ory.sh/default-identity.schema.json"}]}`},
				{id: "totp", enabled: true, config: `{"issuer":"issuer.ory.sh"}`},
			} {
				strategy := p.SelfServiceStrategy(tc.id)
				assert.Equal(t, tc.enabled, strategy.Enabled)
				assert.JSONEq(t, tc.config, string(strategy.Config))
			}
		})

		t.Run("method=registration", func(t *testing.T) {
			assert.Equal(t, true, p.SelfServiceFlowRegistrationEnabled())
			assert.Equal(t, time.Minute*98, p.SelfServiceFlowRegistrationRequestLifespan())

			t.Run("hook=before", func(t *testing.T) {
				expHooks := []config.SelfServiceHook{
					{Name: "web_hook", Config: json.RawMessage(`{"method":"GET","url":"https://test.kratos.ory.sh/before_registration_hook"}`)},
				}

				hooks := p.SelfServiceFlowRegistrationBeforeHooks()

				require.Len(t, hooks, 1)
				assert.Equal(t, expHooks, hooks)
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
						{Name: "web_hook", Config: json.RawMessage(`{"body":"/path/to/template.jsonnet","method":"POST","url":"https://test.kratos.ory.sh/after_registration_password_hook"}`)},
						// {Name: "verify", Config: json.RawMessage(`{}`)},
						// {Name: "redirect", Config: json.RawMessage(`{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`)},
					},
				},
				{
					strategy: "oidc",
					hooks: []config.SelfServiceHook{
						// {Name: "verify", Config: json.RawMessage(`{}`)},
						{Name: "web_hook", Config: json.RawMessage(`{"body":"/path/to/template.jsonnet","method":"GET","url":"https://test.kratos.ory.sh/after_registration_oidc_hook"}`)},
						{Name: "session", Config: json.RawMessage(`{}`)},
						// {Name: "redirect", Config: json.RawMessage(`{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`)},
					},
				},
				{
					strategy: config.HookGlobal,
					hooks: []config.SelfServiceHook{
						{Name: "web_hook", Config: json.RawMessage(`{"auth":{"config":{"in":"header","name":"My-Key","value":"My-Key-Value"},"type":"api_key"},"body":"/path/to/template.jsonnet","method":"POST","url":"https://test.kratos.ory.sh/after_registration_global_hook"}`)},
					},
				},
			} {
				t.Run("hook=after/strategy="+tc.strategy, func(t *testing.T) {
					hooks := p.SelfServiceFlowRegistrationAfterHooks(tc.strategy)
					assert.Equal(t, tc.hooks, hooks)
				})
			}
		})

		t.Run("method=totp", func(t *testing.T) {
			assert.Equal(t, "issuer.ory.sh", p.TOTPIssuer())
		})

		t.Run("method=login", func(t *testing.T) {
			assert.Equal(t, time.Minute*99, p.SelfServiceFlowLoginRequestLifespan())

			t.Run("hook=before", func(t *testing.T) {
				expHooks := []config.SelfServiceHook{
					{Name: "web_hook", Config: json.RawMessage(`{"method":"POST","url":"https://test.kratos.ory.sh/before_login_hook"}`)},
				}

				hooks := p.SelfServiceFlowLoginBeforeHooks()

				require.Len(t, hooks, 1)
				assert.Equal(t, expHooks, hooks)
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
						{Name: "require_verified_address", Config: json.RawMessage(`{}`)},
						{Name: "web_hook", Config: json.RawMessage(`{"auth":{"config":{"password":"super-secret","user":"test-user"},"type":"basic_auth"},"body":"/path/to/template.jsonnet","method":"POST","url":"https://test.kratos.ory.sh/after_login_password_hook"}`)},
					},
				},
				{
					strategy: "oidc",
					hooks: []config.SelfServiceHook{
						{Name: "web_hook", Config: json.RawMessage(`{"body":"/path/to/template.jsonnet","method":"GET","url":"https://test.kratos.ory.sh/after_login_oidc_hook"}`)},
						{Name: "revoke_active_sessions", Config: json.RawMessage(`{}`)},
					},
				},
				{
					strategy: config.HookGlobal,
					hooks: []config.SelfServiceHook{
						{Name: "web_hook", Config: json.RawMessage(`{"body":"/path/to/template.jsonnet","method":"POST","url":"https://test.kratos.ory.sh/after_login_global_hook"}`)},
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

			for _, tc := range []struct {
				strategy string
				hooks    []config.SelfServiceHook
			}{
				{
					strategy: "password",
					hooks: []config.SelfServiceHook{
						{Name: "web_hook", Config: json.RawMessage(`{"body":"/path/to/template.jsonnet","method":"POST","url":"https://test.kratos.ory.sh/after_settings_password_hook"}`)},
					},
				},
				{
					strategy: "profile",
					hooks: []config.SelfServiceHook{
						{Name: "web_hook", Config: json.RawMessage(`{"body":"/path/to/template.jsonnet","method":"POST","url":"https://test.kratos.ory.sh/after_settings_profile_hook"}`)},
					},
				},
				{
					strategy: config.HookGlobal,
					hooks: []config.SelfServiceHook{
						{Name: "web_hook", Config: json.RawMessage(`{"body":"/path/to/template.jsonnet","method":"POST","url":"https://test.kratos.ory.sh/after_settings_global_hook"}`)},
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

			hooks := p.SelfServiceFlowRecoveryAfterHooks(config.HookGlobal)
			assert.Equal(t, []config.SelfServiceHook{{Name: "web_hook", Config: json.RawMessage(`{"body":"/path/to/template.jsonnet","method":"GET","url":"https://test.kratos.ory.sh/after_recovery_hook"}`)}}, hooks)
		})

		t.Run("method=verification", func(t *testing.T) {
			assert.Equal(t, time.Minute*97, p.SelfServiceFlowVerificationRequestLifespan())
			assert.Equal(t, "http://test.kratos.ory.sh/verification", p.SelfServiceFlowVerificationUI().String())

			hooks := p.SelfServiceFlowVerificationAfterHooks(config.HookGlobal)
			assert.Equal(t, []config.SelfServiceHook{{Name: "web_hook", Config: json.RawMessage(`{"body":"/path/to/template.jsonnet","method":"GET","url":"https://test.kratos.ory.sh/after_verification_hook"}`)}}, hooks)
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
	p := config.MustNew(t, logrusx.New("", ""), os.Stderr, configx.SkipValidation())

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

	p := config.MustNew(t, logrusx.New("", ""), os.Stderr, configx.SkipValidation())
	assert.Equal(t, "https://"+machineHostname+":4433/", p.SelfPublicURL().String())
	assert.Equal(t, "https://"+machineHostname+":4434/", p.SelfAdminURL().String())

	p.MustSet(config.ViperKeyPublicPort, 4444)
	p.MustSet(config.ViperKeyAdminPort, 4445)
	assert.Equal(t, "https://"+machineHostname+":4444/", p.SelfPublicURL().String())
	assert.Equal(t, "https://"+machineHostname+":4445/", p.SelfAdminURL().String())

	p.MustSet(config.ViperKeyPublicHost, "public.ory.sh")
	p.MustSet(config.ViperKeyAdminHost, "admin.ory.sh")
	assert.Equal(t, "https://public.ory.sh:4444/", p.SelfPublicURL().String())
	assert.Equal(t, "https://admin.ory.sh:4445/", p.SelfAdminURL().String())

	// Set to dev mode
	p.MustSet("dev", true)
	assert.Equal(t, "http://public.ory.sh:4444/", p.SelfPublicURL().String())
	assert.Equal(t, "http://admin.ory.sh:4445/", p.SelfAdminURL().String())
}

func TestProviderSelfServiceLinkMethodBaseURL(t *testing.T) {
	machineHostname, err := os.Hostname()
	if err != nil {
		machineHostname = "127.0.0.1"
	}

	p := config.MustNew(t, logrusx.New("", ""), os.Stderr, configx.SkipValidation())
	assert.Equal(t, "https://"+machineHostname+":4433/", p.SelfServiceLinkMethodBaseURL().String())

	p.MustSet(config.ViperKeyLinkBaseURL, "https://example.org/bar")
	assert.Equal(t, "https://example.org/bar", p.SelfServiceLinkMethodBaseURL().String())
}

func TestViperProvider_Secrets(t *testing.T) {
	p := config.MustNew(t, logrusx.New("", ""), os.Stderr, configx.SkipValidation())

	def := p.SecretsDefault()
	assert.NotEmpty(t, def)
	assert.Equal(t, def, p.SecretsSession())
	assert.Equal(t, def, p.SecretsDefault())
	assert.Empty(t, p.SecretsCipher())
	err := p.Set(config.ViperKeySecretsCipher, []string{"short-secret-key"})
	require.NoError(t, err)
	assert.Equal(t, [][32]byte{}, p.SecretsCipher())
}

func TestViperProvider_Defaults(t *testing.T) {
	l := logrusx.New("", "")

	for k, tc := range []struct {
		init   func() *config.Config
		expect func(t *testing.T, p *config.Config)
	}{
		{
			init: func() *config.Config {
				return config.MustNew(t, l, os.Stderr, configx.SkipValidation())
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l,
					os.Stderr,
					configx.WithConfigFiles("stub/.defaults.yml"), configx.SkipValidation())
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l, os.Stderr, configx.WithConfigFiles("stub/.defaults-password.yml"), configx.SkipValidation())
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l, os.Stderr, configx.WithConfigFiles("../../test/e2e/profiles/recovery/.kratos.yml"), configx.SkipValidation())
			},
			expect: func(t *testing.T, p *config.Config) {
				assert.True(t, p.SelfServiceFlowRecoveryEnabled())
				assert.False(t, p.SelfServiceFlowVerificationEnabled())
				assert.True(t, p.SelfServiceFlowRegistrationEnabled())
				assert.True(t, p.SelfServiceStrategy("password").Enabled)
				assert.True(t, p.SelfServiceStrategy("profile").Enabled)
				assert.True(t, p.SelfServiceStrategy("link").Enabled)
				assert.False(t, p.SelfServiceStrategy("oidc").Enabled)
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l, os.Stderr, configx.WithConfigFiles("../../test/e2e/profiles/verification/.kratos.yml"), configx.SkipValidation())
			},
			expect: func(t *testing.T, p *config.Config) {
				assert.False(t, p.SelfServiceFlowRecoveryEnabled())
				assert.True(t, p.SelfServiceFlowVerificationEnabled())
				assert.True(t, p.SelfServiceFlowRegistrationEnabled())
				assert.True(t, p.SelfServiceStrategy("password").Enabled)
				assert.True(t, p.SelfServiceStrategy("profile").Enabled)
				assert.True(t, p.SelfServiceStrategy("link").Enabled)
				assert.False(t, p.SelfServiceStrategy("oidc").Enabled)
			},
		},
		{
			init: func() *config.Config {
				return config.MustNew(t, l, os.Stderr, configx.WithConfigFiles("../../test/e2e/profiles/oidc/.kratos.yml"), configx.SkipValidation())
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
		p := config.MustNew(t, l, os.Stderr, configx.SkipValidation())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/login", p.SelfServiceFlowLoginUI().String())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/settings", p.SelfServiceFlowSettingsUI().String())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/registration", p.SelfServiceFlowRegistrationUI().String())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/recovery", p.SelfServiceFlowRecoveryUI().String())
		assert.Equal(t, "https://www.ory.sh/kratos/docs/fallback/verification", p.SelfServiceFlowVerificationUI().String())
	})
}

func TestViperProvider_ReturnTo(t *testing.T) {
	l := logrusx.New("", "")
	p := config.MustNew(t, l, os.Stderr, configx.SkipValidation())

	p.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	assert.Equal(t, "https://www.ory.sh/", p.SelfServiceFlowVerificationReturnTo(urlx.ParseOrPanic("https://www.ory.sh/")).String())
	assert.Equal(t, "https://www.ory.sh/", p.SelfServiceFlowRecoveryReturnTo().String())

	p.MustSet(config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, "https://www.ory.sh/recovery")
	assert.Equal(t, "https://www.ory.sh/recovery", p.SelfServiceFlowRecoveryReturnTo().String())

	p.MustSet(config.ViperKeySelfServiceVerificationBrowserDefaultReturnTo, "https://www.ory.sh/verification")
	assert.Equal(t, "https://www.ory.sh/verification", p.SelfServiceFlowVerificationReturnTo(urlx.ParseOrPanic("https://www.ory.sh/")).String())
}

func TestSession(t *testing.T) {
	l := logrusx.New("", "")
	p := config.MustNew(t, l, os.Stderr, configx.SkipValidation())

	assert.Equal(t, "ory_kratos_session", p.SessionName())
	p.MustSet(config.ViperKeySessionName, "ory_session")
	assert.Equal(t, "ory_session", p.SessionName())

	assert.Equal(t, time.Hour*24, p.SessionLifespan())
	p.MustSet(config.ViperKeySessionLifespan, "1m")
	assert.Equal(t, time.Minute, p.SessionLifespan())

	assert.Equal(t, true, p.SessionPersistentCookie())
	p.MustSet(config.ViperKeySessionPersistentCookie, false)
	assert.Equal(t, false, p.SessionPersistentCookie())
}

func TestCookies(t *testing.T) {
	l := logrusx.New("", "")
	p := config.MustNew(t, l, os.Stderr, configx.SkipValidation())

	t.Run("path", func(t *testing.T) {
		assert.Equal(t, "/", p.CookiePath())
		assert.Equal(t, "/", p.SessionPath())

		p.MustSet(config.ViperKeyCookiePath, "/cookie")
		assert.Equal(t, "/cookie", p.CookiePath())
		assert.Equal(t, "/cookie", p.SessionPath())

		p.MustSet(config.ViperKeySessionPath, "/session")
		assert.Equal(t, "/cookie", p.CookiePath())
		assert.Equal(t, "/session", p.SessionPath())
	})

	t.Run("SameSite", func(t *testing.T) {
		assert.Equal(t, http.SameSiteLaxMode, p.CookieSameSiteMode())
		assert.Equal(t, http.SameSiteLaxMode, p.SessionSameSiteMode())

		p.MustSet(config.ViperKeyCookieSameSite, "Strict")
		assert.Equal(t, http.SameSiteStrictMode, p.CookieSameSiteMode())
		assert.Equal(t, http.SameSiteStrictMode, p.SessionSameSiteMode())

		p.MustSet(config.ViperKeySessionSameSite, "None")
		assert.Equal(t, http.SameSiteStrictMode, p.CookieSameSiteMode())
		assert.Equal(t, http.SameSiteNoneMode, p.SessionSameSiteMode())
	})

	t.Run("domain", func(t *testing.T) {
		assert.Equal(t, "", p.CookieDomain())
		assert.Equal(t, "", p.SessionDomain())

		p.MustSet(config.ViperKeyCookieDomain, "www.cookie.com")
		assert.Equal(t, "www.cookie.com", p.CookieDomain())
		assert.Equal(t, "www.cookie.com", p.SessionDomain())

		p.MustSet(config.ViperKeySessionDomain, "www.session.com")
		assert.Equal(t, "www.cookie.com", p.CookieDomain())
		assert.Equal(t, "www.session.com", p.SessionDomain())
	})
}

func TestViperProvider_DSN(t *testing.T) {
	t.Run("case=dsn: memory", func(t *testing.T) {
		p := config.MustNew(t, logrusx.New("", ""), os.Stderr, configx.SkipValidation())
		p.MustSet(config.ViperKeyDSN, "memory")

		assert.Equal(t, config.DefaultSQLiteMemoryDSN, p.DSN())
	})

	t.Run("case=dsn: not memory", func(t *testing.T) {
		p := config.MustNew(t, logrusx.New("", ""), os.Stderr, configx.SkipValidation())

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
		p := config.MustNew(t, l, os.Stderr, configx.SkipValidation())

		assert.Equal(t, dsn, p.DSN())
		assert.NotEqual(t, 0, exitCode)
	})
}

func TestViperProvider_ParseURIOrFail(t *testing.T) {
	var exitCode int

	l := logrusx.New("", "", logrusx.WithExitFunc(func(i int) {
		exitCode = i
	}))
	p := config.MustNew(t, l, os.Stderr, configx.SkipValidation())
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

func TestViperProvider_HaveIBeenPwned(t *testing.T) {
	p := config.MustNew(t, logrusx.New("", ""), os.Stderr, configx.SkipValidation())
	t.Run("case=hipb: host", func(t *testing.T) {
		p.MustSet(config.ViperKeyPasswordHaveIBeenPwnedHost, "foo.bar")
		assert.Equal(t, "foo.bar", p.PasswordPolicyConfig().HaveIBeenPwnedHost)
	})

	t.Run("case=hibp: enabled", func(t *testing.T) {
		p.MustSet(config.ViperKeyPasswordHaveIBeenPwnedEnabled, true)
		assert.Equal(t, true, p.PasswordPolicyConfig().HaveIBeenPwnedEnabled)
	})

	t.Run("case=hibp: enabled", func(t *testing.T) {
		p.MustSet(config.ViperKeyPasswordHaveIBeenPwnedEnabled, false)
		assert.Equal(t, false, p.PasswordPolicyConfig().HaveIBeenPwnedEnabled)
	})

	t.Run("case=hibp: max_breaches", func(t *testing.T) {
		p.MustSet(config.ViperKeyPasswordMaxBreaches, 10)
		assert.Equal(t, uint(10), p.PasswordPolicyConfig().MaxBreaches)
	})

	t.Run("case=hibp: ignore_network_errors", func(t *testing.T) {
		p.MustSet(config.ViperKeyIgnoreNetworkErrors, true)
		assert.Equal(t, true, p.PasswordPolicyConfig().IgnoreNetworkErrors)
	})

	t.Run("case=hibp: ignore_network_errors", func(t *testing.T) {
		p.MustSet(config.ViperKeyIgnoreNetworkErrors, false)
		assert.Equal(t, false, p.PasswordPolicyConfig().IgnoreNetworkErrors)
	})
}

func TestLoadingTLSConfig(t *testing.T) {
	certPath := filepath.Join(os.TempDir(), "e2e_test_cert_"+x.NewUUID().String()+".pem")
	keyPath := filepath.Join(os.TempDir(), "e2e_test_key_"+x.NewUUID().String()+".pem")

	testhelpers.GenerateTLSCertificateFilesForTests(t, certPath, keyPath)

	certRaw, err := ioutil.ReadFile(certPath)
	assert.Nil(t, err)

	keyRaw, err := ioutil.ReadFile(keyPath)
	assert.Nil(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(certRaw)
	keyBase64 := base64.StdEncoding.EncodeToString(keyRaw)

	t.Run("case=public: loading inline base64 certificate", func(t *testing.T) {
		logger := logrusx.New("", "")
		logger.Logger.ExitFunc = func(code int) { panic("") }
		hook := new(test.Hook)
		logger.Logger.Hooks.Add(hook)

		p := config.MustNew(t, logger, os.Stderr, configx.SkipValidation())
		p.MustSet(config.ViperKeyPublicTLSKeyBase64, keyBase64)
		p.MustSet(config.ViperKeyPublicTLSCertBase64, certBase64)
		assert.NotNil(t, p.GetTSLCertificatesForPublic())
		assert.Equal(t, "Setting up HTTPS for public", hook.LastEntry().Message)
	})

	t.Run("case=public: loading certificate from a file", func(t *testing.T) {
		logger := logrusx.New("", "")
		logger.Logger.ExitFunc = func(code int) { panic("") }
		hook := new(test.Hook)
		logger.Logger.Hooks.Add(hook)

		p := config.MustNew(t, logger, os.Stderr, configx.SkipValidation())
		p.MustSet(config.ViperKeyPublicTLSKeyPath, keyPath)
		p.MustSet(config.ViperKeyPublicTLSCertPath, certPath)
		assert.NotNil(t, p.GetTSLCertificatesForPublic())
		assert.Equal(t, "Setting up HTTPS for public", hook.LastEntry().Message)
	})

	t.Run("case=public: failing to load inline base64 certificate", func(t *testing.T) {
		logger := logrusx.New("", "")
		logger.Logger.ExitFunc = func(code int) {}
		hook := new(test.Hook)
		logger.Logger.Hooks.Add(hook)

		p := config.MustNew(t, logger, os.Stderr, configx.SkipValidation())
		p.MustSet(config.ViperKeyPublicTLSKeyBase64, "empty")
		p.MustSet(config.ViperKeyPublicTLSCertBase64, certBase64)
		assert.Nil(t, p.GetTSLCertificatesForPublic())
		assert.Equal(t, "TLS has not been configured for public, skipping", hook.LastEntry().Message)
	})

	t.Run("case=public: failing to load certificate from a file", func(t *testing.T) {
		logger := logrusx.New("", "")
		logger.Logger.ExitFunc = func(code int) {}
		hook := new(test.Hook)
		logger.Logger.Hooks.Add(hook)

		p := config.MustNew(t, logger, os.Stderr, configx.SkipValidation())
		p.MustSet(config.ViperKeyPublicTLSKeyPath, "/dev/null")
		p.MustSet(config.ViperKeyPublicTLSCertPath, certPath)
		assert.Nil(t, p.GetTSLCertificatesForPublic())
		assert.Equal(t, "TLS has not been configured for public, skipping", hook.LastEntry().Message)
	})

	t.Run("case=admin: loading inline base64 certificate", func(t *testing.T) {
		logger := logrusx.New("", "")
		logger.Logger.ExitFunc = func(code int) { panic("") }
		hook := new(test.Hook)
		logger.Logger.Hooks.Add(hook)

		p := config.MustNew(t, logger, os.Stderr, configx.SkipValidation())
		p.MustSet(config.ViperKeyAdminTLSKeyBase64, keyBase64)
		p.MustSet(config.ViperKeyAdminTLSCertBase64, certBase64)
		assert.NotNil(t, p.GetTSLCertificatesForAdmin())
		assert.Equal(t, "Setting up HTTPS for admin", hook.LastEntry().Message)
	})

	t.Run("case=admin: loading certificate from a file", func(t *testing.T) {
		logger := logrusx.New("", "")
		logger.Logger.ExitFunc = func(code int) { panic("") }
		hook := new(test.Hook)
		logger.Logger.Hooks.Add(hook)

		p := config.MustNew(t, logger, os.Stderr, configx.SkipValidation())
		p.MustSet(config.ViperKeyAdminTLSKeyPath, keyPath)
		p.MustSet(config.ViperKeyAdminTLSCertPath, certPath)
		assert.NotNil(t, p.GetTSLCertificatesForAdmin())
		assert.Equal(t, "Setting up HTTPS for admin", hook.LastEntry().Message)
	})

	t.Run("case=admin: failing to load inline base64 certificate", func(t *testing.T) {
		logger := logrusx.New("", "")
		logger.Logger.ExitFunc = func(code int) {}
		hook := new(test.Hook)
		logger.Logger.Hooks.Add(hook)

		p := config.MustNew(t, logger, os.Stderr, configx.SkipValidation())
		p.MustSet(config.ViperKeyAdminTLSKeyBase64, "empty")
		p.MustSet(config.ViperKeyAdminTLSCertBase64, certBase64)
		assert.Nil(t, p.GetTSLCertificatesForAdmin())
		assert.Equal(t, "TLS has not been configured for admin, skipping", hook.LastEntry().Message)
	})

	t.Run("case=admin: failing to load certificate from a file", func(t *testing.T) {
		logger := logrusx.New("", "")
		logger.Logger.ExitFunc = func(code int) {}
		hook := new(test.Hook)
		logger.Logger.Hooks.Add(hook)

		p := config.MustNew(t, logger, os.Stderr, configx.SkipValidation())
		p.MustSet(config.ViperKeyAdminTLSKeyPath, "/dev/null")
		p.MustSet(config.ViperKeyAdminTLSCertPath, certPath)
		assert.Nil(t, p.GetTSLCertificatesForAdmin())
		assert.Equal(t, "TLS has not been configured for admin, skipping", hook.LastEntry().Message)
	})

}

func TestIdentitySchemaValidation(t *testing.T) {
	files := []string{"stub/.identity.test.json", "stub/.identity.other.json"}

	type identity struct {
		DefaultSchemaUrl string   `json:"default_schema_url"`
		Schemas          []string `json:"schemas"`
	}

	type configFile struct {
		identityFileName string
		SelfService      map[string]string            `json:"selfservice"`
		Courier          map[string]map[string]string `json:"courier"`
		DSN              string                       `json:"dsn"`
		Identity         *identity                    `json:"identity"`
	}

	setup := func(t *testing.T, file string) *configFile {
		identityTest, err := os.ReadFile(file)
		assert.NoError(t, err)
		return &configFile{
			identityFileName: file,
			SelfService: map[string]string{
				"default_browser_return_url": "https://some-return-url",
			},
			Courier: map[string]map[string]string{
				"smtp": {
					"connection_uri": "smtp://foo@bar",
				},
			},
			DSN: "memory",
			Identity: &identity{
				DefaultSchemaUrl: "base64://" + base64.StdEncoding.EncodeToString(identityTest),
				Schemas:          []string{},
			},
		}
	}

	marshalAndWrite := func(t *testing.T, ctx context.Context, tmpFile *os.File, identity *configFile) {
		j, err := yaml.Marshal(identity)
		assert.NoError(t, err)

		_, err = tmpFile.Seek(0, 0)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Truncate(0))
		_, err = io.WriteString(tmpFile, string(j))
		assert.NoError(t, err)
		assert.NoError(t, tmpFile.Sync())
	}

	testWatch := func(t *testing.T, ctx context.Context, cmd *cobra.Command, i *configFile) (*config.Config, *test.Hook, *os.File, *configFile, chan bool) {
		c := make(chan bool, 1)
		tdir := t.TempDir()
		assert.NoError(t,
			os.MkdirAll(tdir, // DO NOT CHANGE THIS: https://github.com/fsnotify/fsnotify/issues/340
				os.ModePerm))
		tmpConfig, err := os.Create(filepath.Join(tdir, "config.yaml"))
		assert.NoError(t, err)

		marshalAndWrite(t, ctx, tmpConfig, i)

		l := logrusx.New("kratos-"+tmpConfig.Name(), "test")
		hook := test.NewLocal(l.Logger)

		conf, err := config.New(ctx, l, os.Stderr,
			configx.WithConfigFiles(tmpConfig.Name()),
			configx.AttachWatcher(func(event watcherx.Event, err error) {
				c <- true
			}))
		assert.NoError(t, err)

		// clean the hooks since it will throw an event on first boot
		hook.Reset()

		return conf, hook, tmpConfig, i, c
	}

	t.Run("case=skip invalid schema validation", func(t *testing.T) {
		ctx := context.Background()
		_, err := config.New(ctx, logrusx.New("", ""), os.Stderr,
			configx.WithConfigFiles("stub/.kratos.invalid.identities.yaml"),
			configx.SkipValidation())
		assert.NoError(t, err)
	})

	t.Run("case=invalid schema should throw error", func(t *testing.T) {
		ctx := context.Background()
		var stdErr bytes.Buffer
		_, err := config.New(ctx, logrusx.New("", ""), &stdErr,
			configx.WithConfigFiles("stub/.kratos.invalid.identities.yaml"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minimum 1 properties allowed, but found 0")
		assert.Contains(t, stdErr.String(), "minimum 1 properties allowed, but found 0")
	})

	t.Run("case=must fail on loading unreachable schemas", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		t.Cleanup(cancel)

		err := make(chan error, 1)
		go func(err chan error) {
			_, e := config.New(ctx, logrusx.New("", ""), os.Stderr,
				configx.WithConfigFiles("stub/.kratos.mock.identities.yaml"))
			err <- e
		}(err)

		select {
		case <-ctx.Done():
			panic("the test could not complete as the context timed out before the identity schema loader timed out")
		case e := <-err:
			assert.Error(t, e)
			assert.Contains(t, e.Error(), "no such host")
		}

	})

	t.Run("case=validate schema is validated on file change", func(t *testing.T) {
		var identities []*configFile

		for _, f := range files {
			identities = append(identities, setup(t, f))
		}

		invalidIdentity := setup(t, "stub/.identity.invalid.json")

		for _, i := range identities {
			t.Run("test=identity file "+i.identityFileName, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

				_, hook, tmpConfig, i, c := testWatch(t, ctx, &cobra.Command{}, i)
				// Change the identity config to an invalid file
				i.Identity.DefaultSchemaUrl = invalidIdentity.Identity.DefaultSchemaUrl

				t.Cleanup(func() {
					cancel()
					tmpConfig.Close()
				})

				var wg sync.WaitGroup
				wg.Add(1)
				go func(t *testing.T, ctx context.Context, tmpFile *os.File, identity *configFile) {
					defer wg.Done()
					marshalAndWrite(t, ctx, tmpConfig, i)
				}(t, ctx, tmpConfig, i)

				select {
				case <-ctx.Done():
					panic("the test could not complete as the context timed out before the file watcher updated")
				case <-c:
					lastHook, err := hook.LastEntry().String()
					assert.NoError(t, err)

					assert.Contains(t, lastHook, "The changed identity schema configuration is invalid and could not be loaded.")
				}

				wg.Wait()
			})
		}
	})
}

func TestChangeMinPasswordLength(t *testing.T) {
	t.Run("case=must fail on minimum password length below enforced minimum", func(t *testing.T) {
		ctx := context.Background()

		_, err := config.New(ctx, logrusx.New("", ""), os.Stderr,
			configx.WithConfigFiles("stub/.kratos.yaml"),
			configx.WithValue(config.ViperKeyPasswordMinLength, 5))

		assert.Error(t, err)
	})

	t.Run("case=must not fail on minimum password length above enforced minimum", func(t *testing.T) {
		ctx := context.Background()

		_, err := config.New(ctx, logrusx.New("", ""), os.Stderr,
			configx.WithConfigFiles("stub/.kratos.yaml"),
			configx.WithValue(config.ViperKeyPasswordMinLength, 9))

		assert.NoError(t, err)
	})
}
