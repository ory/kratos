package configuration_test

import (
	"testing"
	"time"

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
			logrus.New(),
		)

		require.NoError(t, viperx.ValidateFromURL("file://../../.schema/config.schema.json"))
		p := configuration.NewViperProvider(logrus.New(), true)

		t.Run("group=urls", func(t *testing.T) {
			assert.Equal(t, "http://test.kratos.ory.sh/login", p.LoginURL().String())
			assert.Equal(t, "http://test.kratos.ory.sh/settings", p.SettingsURL().String())
			assert.Equal(t, "http://test.kratos.ory.sh/register", p.RegisterURL().String())
			assert.Equal(t, "http://test.kratos.ory.sh/mfa", p.MultiFactorURL().String())
			assert.Equal(t, "http://test.kratos.ory.sh/error", p.ErrorURL().String())

			assert.Equal(t, "http://admin.kratos.ory.sh", p.SelfAdminURL().String())
			assert.Equal(t, "http://public.kratos.ory.sh", p.SelfPublicURL().String())

			var ds []string
			for _, v := range p.WhitelistedReturnToDomains() {
				ds = append(ds, v.String())
			}
			assert.Equal(t, []string{
				"http://return-to-1-test.ory.sh/",
				"http://return-to-2-test.ory.sh/",
			}, ds)
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
			}, p.SessionSecrets())
		})

		t.Run("group=strategies", func(t *testing.T) {
			for _, tc := range []struct {
				id      string
				config  string
				enabled bool
			}{
				{id: "password", enabled: true, config: "{}"},
				{id: "oidc", enabled: true, config: `{"providers":[{"client_id":"a","client_secret":"b","id":"github","provider":"github","schema_url":"http://test.kratos.ory.sh/default-identity.schema.json"}]}`},
			} {
				strategy := p.SelfServiceStrategy(tc.id)
				assert.Equal(t, tc.enabled, strategy.Enabled)
				assert.EqualValues(t, string(tc.config), string(strategy.Config))
			}
		})

		t.Run("method=registration", func(t *testing.T) {
			assert.Equal(t, time.Minute*98, p.SelfServiceRegistrationRequestLifespan())

			t.Run("hook=before", func(t *testing.T) {
				hook := p.SelfServiceRegistrationBeforeHooks()[0]
				assert.EqualValues(t, "redirect", hook.Job)
				assert.JSONEq(t, `{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`, string(hook.Config))
			})

			for _, tc := range []struct {
				strategy       string
				redirectConfig string
			}{
				{
					strategy:       "password",
					redirectConfig: `{"allow_user_defined_redirect":true,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`,
				},
				{
					strategy:       "oidc",
					redirectConfig: `{"allow_user_defined_redirect":true,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`,
				},
			} {
				t.Run("hook=after/strategy="+tc.strategy, func(t *testing.T) {
					hooks := p.SelfServiceRegistrationAfterHooks(tc.strategy)

					hook := hooks[0]
					assert.EqualValues(t, "session", hook.Job)

					hook = hooks[1]
					assert.EqualValues(t, "redirect", hook.Job)
					assert.JSONEq(t, tc.redirectConfig, string(hook.Config))
				})
			}

		})

		t.Run("method=login", func(t *testing.T) {
			assert.Equal(t, time.Minute*99, p.SelfServiceLoginRequestLifespan())

			t.Run("hook=before", func(t *testing.T) {
				hook := p.SelfServiceLoginBeforeHooks()[0]
				assert.EqualValues(t, "redirect", hook.Job)
				assert.JSONEq(t, `{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`, string(hook.Config))
			})

			for _, tc := range []struct {
				strategy       string
				redirectConfig string
			}{
				{
					strategy:       "password",
					redirectConfig: `{"allow_user_defined_redirect":true,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`,
				},
				{
					strategy:       "oidc",
					redirectConfig: `{"allow_user_defined_redirect":true,"default_redirect_url":"http://test.kratos.ory.sh:4000/"}`,
				},
			} {
				t.Run("hook=after/strategy="+tc.strategy, func(t *testing.T) {
					hooks := p.SelfServiceLoginAfterHooks(tc.strategy)

					hook := hooks[0]
					assert.EqualValues(t, "revoke_active_sessions", hook.Job)

					hook = hooks[1]
					assert.EqualValues(t, "session", hook.Job)

					hook = hooks[2]
					assert.EqualValues(t, "redirect", hook.Job)
					assert.JSONEq(t, tc.redirectConfig, string(hook.Config))
				})
			}
		})

		t.Run("group=hashers", func(t *testing.T) {
			assert.Equal(t, &configuration.HasherArgon2Config{
				Memory:      1048576,
				Iterations:  2,
				Parallelism: 4,
				SaltLength:  16,
				KeyLength:   32,
			}, p.HashersArgon2())
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

func TestViperProvider_DSN(t *testing.T) {
	t.Run("case=dsn: memory", func(t *testing.T) {
		viper.Reset()
		viper.Set(configuration.ViperKeyDSN, "memory")

		l := logrus.New()
		p := configuration.NewViperProvider(l, false)

		assert.Equal(t, "sqlite://mem.db?mode=memory&_fk=true&cache=shared", p.DSN())
	})

	t.Run("case=dsn: not memory", func(t *testing.T) {
		dsn := "sqlite://foo.db?_fk=true"
		viper.Reset()
		viper.Set(configuration.ViperKeyDSN, dsn)

		l := logrus.New()
		p := configuration.NewViperProvider(l, false)

		assert.Equal(t, dsn, p.DSN())
	})

	t.Run("case=dsn: not set", func(t *testing.T) {
		dsn := ""
		viper.Reset()
		viper.Set(configuration.ViperKeyDSN, dsn)

		l := logrus.New()
		p := configuration.NewViperProvider(l, false)

		var exitCode int
		l.ExitFunc = func(i int) {
			exitCode = i
		}
		h := InterceptHook{}
		l.AddHook(h)
		assert.Equal(t, dsn, p.DSN())
		assert.NotEqual(t, 0, exitCode)
	})
}
