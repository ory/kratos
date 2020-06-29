package configuration_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/driver/configuration"

	_ "github.com/ory/jsonschema/v3/fileloader"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/x/viperx"
)

func TestViperProvider(t *testing.T) {
	t.Run("suite=loaders", func(t *testing.T) {
		viper.Reset()
		viperx.InitializeConfig(
			"kratos",
			"./../../internal/",
			logrusx.New("", ""),
		)

		require.NoError(t, viperx.ValidateFromURL("file://../../.schema/config.schema.json"))
		p := configuration.NewViperProvider(logrusx.New("", ""), true)

		t.Run("group=urls", func(t *testing.T) {
			assert.Equal(t, "http://test.kratos.ory.sh/login", p.SelfServiceFlowLoginUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/settings", p.SelfServiceFlowSettingsUI().String())
			assert.Equal(t, "http://test.kratos.ory.sh/register", p.SelfServiceFlowRegisterUI().String())
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
			}, ds)
		})

		t.Run("group=default_return_to", func(t *testing.T) {
			assert.Equal(t, "https://self-service/login/password/return_to", p.SelfServiceFlowLoginReturnTo("password").String())
			assert.Equal(t, "https://self-service/login/return_to", p.SelfServiceFlowLoginReturnTo("oidc").String())

			assert.Equal(t, "https://self-service/registration/return_to", p.SelfServiceFlowRegistrationReturnTo("password").String())
			assert.Equal(t, "https://self-service/registration/oidc/return_to", p.SelfServiceFlowRegistrationReturnTo("oidc").String())

			assert.Equal(t, "https://self-service/settings/password/return_to", p.SelfServiceFlowSettingsReturnTo("password", p.SelfServiceBrowserDefaultReturnTo()).String())
			assert.Equal(t, "https://self-service/settings/return_to", p.SelfServiceFlowSettingsReturnTo("profile", p.SelfServiceBrowserDefaultReturnTo()).String())
		})

		t.Run("group=identity", func(t *testing.T) {
			assert.Equal(t, "http://test.kratos.ory.sh/default-identity.schema.json", p.DefaultIdentityTraitsSchemaURL().String())

			ss := p.IdentityTraitsSchemas()
			assert.Equal(t, 2, len(ss))

			assert.Contains(t, ss, configuration.SchemaConfig{
				ID:  "default",
				URL: "http://test.kratos.ory.sh/default-identity.schema.json",
			})
			assert.Contains(t, ss, configuration.SchemaConfig{
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

		t.Run("group=strategies", func(t *testing.T) {
			for _, tc := range []struct {
				id      string
				config  string
				enabled bool
			}{
				{id: "password", enabled: true, config: `{}`},
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
				hooks    []configuration.SelfServiceHook
			}{
				{
					strategy: "password",
					hooks: []configuration.SelfServiceHook{
						{Name: "session", Config: json.RawMessage(`{}`)},
						// {Name: "verify", Config: json.RawMessage(`{}`)},
						// {Name: "redirect", Config: json.RawMessage(`{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`)},
					},
				},
				{
					strategy: "oidc",
					hooks: []configuration.SelfServiceHook{
						// {Name: "verify", Config: json.RawMessage(`{}`)},
						{Name: "session", Config: json.RawMessage(`{}`)},
						// {Name: "redirect", Config: json.RawMessage(`{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`)},
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
				hooks    []configuration.SelfServiceHook
			}{
				{
					strategy: "password",
					hooks: []configuration.SelfServiceHook{
						{Name: "revoke_active_sessions", Config: json.RawMessage(`{}`)},
					},
				},
				{
					strategy: "oidc",
					hooks: []configuration.SelfServiceHook{
						{Name: "revoke_active_sessions", Config: json.RawMessage(`{}`)},
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
			assert.Equal(t, time.Minute*99, p.SelfServiceFlowSettingsRequestLifespan())
			assert.Equal(t, time.Minute*5, p.SelfServiceFlowSettingsPrivilegedSessionMaxAge())

			t.Run("hook=before", func(t *testing.T) {
				hook := p.SelfServiceFlowLoginBeforeHooks()
				require.Len(t, hook, 0)
				// assert.EqualValues(t, "redirect", hook.Name)
				// assert.JSONEq(t, `{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`, string(hook.Config))
			})

			for _, tc := range []struct {
				strategy string
				hooks    []configuration.SelfServiceHook
			}{
				{
					strategy: "password",
					hooks:    []configuration.SelfServiceHook{},
				},
				{
					strategy: "profile",
					hooks: []configuration.SelfServiceHook{
						// {Name: "verify", Config: json.RawMessage(`{}`)},
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
			assert.Equal(t, time.Minute*98, p.SelfServiceFlowRecoveryRequestLifespan())
			assert.Equal(t, "http://test.kratos.ory.sh/recovery", p.SelfServiceFlowRecoveryUI().String())
		})

		t.Run("method=verification", func(t *testing.T) {
			assert.Equal(t, time.Minute*97, p.SelfServiceFlowVerificationRequestLifespan())
			assert.Equal(t, "http://test.kratos.ory.sh/verification", p.SelfServiceFlowVerificationUI().String())
		})

		t.Run("group=hashers", func(t *testing.T) {
			assert.Equal(t, &configuration.HasherArgon2Config{Memory: 1048576, Iterations: 2, Parallelism: 4,
				SaltLength: 16, KeyLength: 32}, p.HasherArgon2())
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

func TestViperProvider_BaseURLs(t *testing.T) {
	viper.Reset()

	machineHostname, err := os.Hostname()
	if err != nil {
		machineHostname = "127.0.0.1"
	}

	l := logrusx.New("", "")
	p := configuration.NewViperProvider(l, false)
	assert.Equal(t, "https://"+machineHostname+":4433/", p.SelfPublicURL().String())
	assert.Equal(t, "https://"+machineHostname+":4434/", p.SelfAdminURL().String())

	viper.Set(configuration.ViperKeyPublicPort, 4444)
	viper.Set(configuration.ViperKeyAdminPort, 4445)

	assert.Equal(t, "https://"+machineHostname+":4444/", p.SelfPublicURL().String())
	assert.Equal(t, "https://"+machineHostname+":4445/", p.SelfAdminURL().String())

	viper.Set(configuration.ViperKeyPublicHost, "public.ory.sh")
	viper.Set(configuration.ViperKeyAdminHost, "admin.ory.sh")

	assert.Equal(t, "https://public.ory.sh:4444/", p.SelfPublicURL().String())
	assert.Equal(t, "https://admin.ory.sh:4445/", p.SelfAdminURL().String())

	p = configuration.NewViperProvider(l, true)
	assert.Equal(t, "http://public.ory.sh:4444/", p.SelfPublicURL().String())
	assert.Equal(t, "http://admin.ory.sh:4445/", p.SelfAdminURL().String())
}

func TestViperProvider_Secrets(t *testing.T) {
	viper.Reset()
	l := logrusx.New("", "")
	p := configuration.NewViperProvider(l, false)

	def := p.SecretsDefault()
	assert.NotEmpty(t, def)
	assert.Equal(t, def, p.SecretsSession())
	assert.Equal(t, def, p.SecretsDefault())
}

func TestViperProvider_Defaults(t *testing.T) {
	viper.Reset()
	l := logrusx.New("", "")
	p := configuration.NewViperProvider(l, false)

	assert.False(t, p.SelfServiceFlowRecoveryEnabled())
	assert.False(t, p.SelfServiceFlowVerificationEnabled())
	assert.True(t, p.SelfServiceStrategy("password").Enabled)
	assert.True(t, p.SelfServiceStrategy("profile").Enabled)
	assert.True(t, p.SelfServiceStrategy("recovery_token").Enabled)
	assert.False(t, p.SelfServiceStrategy("oidc").Enabled)
}

func TestViperProvider_DSN(t *testing.T) {
	t.Run("case=dsn: memory", func(t *testing.T) {
		viper.Reset()
		viper.Set(configuration.ViperKeyDSN, "memory")

		l := logrusx.New("", "")
		p := configuration.NewViperProvider(l, false)

		assert.Equal(t, "sqlite://mem.db?mode=memory&_fk=true&cache=shared", p.DSN())
	})

	t.Run("case=dsn: not memory", func(t *testing.T) {
		dsn := "sqlite://foo.db?_fk=true"
		viper.Reset()
		viper.Set(configuration.ViperKeyDSN, dsn)

		l := logrusx.New("", "")
		p := configuration.NewViperProvider(l, false)

		assert.Equal(t, dsn, p.DSN())
	})

	t.Run("case=dsn: not set", func(t *testing.T) {
		dsn := ""
		viper.Reset()
		viper.Set(configuration.ViperKeyDSN, dsn)

		var exitCode int
		l := logrusx.New("", "", logrusx.WithExitFunc(func(i int) {
			exitCode = i
		}), logrusx.WithHook(InterceptHook{}))
		p := configuration.NewViperProvider(l, false)

		assert.Equal(t, dsn, p.DSN())
		assert.NotEqual(t, 0, exitCode)
	})
}
