package configuration

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/x/viperx"

	"github.com/ory/hive/schema"
)

func TestViperProvider(t *testing.T) {
	t.Run("suite=loaders", func(t *testing.T) {
		viper.Reset()
		viperx.InitializeConfig(
			"hive",
			"./../../docs/",
			logrus.New(),
		)

		err := viperx.Validate(schema.MustNewWindowsCompatibleReferenceLoader("file://../../docs/config.schema.json"))
		if err != nil {
			viperx.LoggerWithValidationErrorFields(logrus.New(), err).Error(err.Error())
		}

		require.NoError(t, err, "%+v", errors.Cause(err))
		p := NewViperProvider(logrus.New())

		t.Run("group=urls", func(t *testing.T) {
			assert.Equal(t, "http://test.hive.ory.sh/login", p.LoginURL().String())
			assert.Equal(t, "http://test.hive.ory.sh/register", p.RegisterURL().String())
			assert.Equal(t, "http://test.hive.ory.sh/mfa", p.MultiFactorURL().String())
			assert.Equal(t, "http://test.hive.ory.sh/error", p.ErrorURL().String())

			assert.Equal(t, "http://admin.hive.ory.sh", p.SelfAdminURL().String())
			assert.Equal(t, "http://public.hive.ory.sh", p.SelfPublicURL().String())

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
			assert.Equal(t, "http://test.hive.ory.sh/default-identity.schema.json", p.DefaultIdentityTraitsSchemaURL().String())
		})

		t.Run("group=serve", func(t *testing.T) {
			assert.Equal(t, "admin.hive.ory.sh:1234", p.AdminListenOn())
			assert.Equal(t, "public.hive.ory.sh:1235", p.PublicListenOn())
		})

		t.Run("group=dsn", func(t *testing.T) {
			assert.Equal(t, "memory", p.DSN())
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
				{id: "oidc", enabled: true, config: `{"providers":[{"client_id":"a","client_secret":"b","id":"github","provider":"github","schema_url":"http://test.hive.ory.sh/default-identity.schema.json"}]}`},
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
				assert.EqualValues(t, "redirect", hook.Run)
				assert.JSONEq(t, `{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.hive.ory.sh:4000/"}`, string(hook.Config))
			})

			for _, tc := range []struct {
				strategy       string
				redirectConfig string
			}{
				{
					strategy:       "password",
					redirectConfig: `{"allow_user_defined_redirect":true,"default_redirect_url":"http://test.hive.ory.sh:4000/"}`,
				},
				{
					strategy:       "oidc",
					redirectConfig: `{"allow_user_defined_redirect":true,"default_redirect_url":"http://test.hive.ory.sh:4000/"}`,
				},
			} {
				t.Run("hook=after/strategy="+tc.strategy, func(t *testing.T) {
					hooks := p.SelfServiceLoginAfterHooks(tc.strategy)

					hook := hooks[0]
					assert.EqualValues(t, "session", hook.Run)

					hook = hooks[1]
					assert.EqualValues(t, "redirect", hook.Run)
					assert.JSONEq(t, tc.redirectConfig, string(hook.Config))
				})
			}

		})

		t.Run("method=login", func(t *testing.T) {
			assert.Equal(t, time.Minute*99, p.SelfServiceLoginRequestLifespan())

			t.Run("hook=before", func(t *testing.T) {
				hook := p.SelfServiceLoginBeforeHooks()[0]
				assert.EqualValues(t, "redirect", hook.Run)
				assert.JSONEq(t, `{"allow_user_defined_redirect":false,"default_redirect_url":"http://test.hive.ory.sh:4000/"}`, string(hook.Config))
			})

			for _, tc := range []struct {
				strategy       string
				redirectConfig string
			}{
				{
					strategy:       "password",
					redirectConfig: `{"allow_user_defined_redirect":true,"default_redirect_url":"http://test.hive.ory.sh:4000/"}`,
				},
				{
					strategy:       "oidc",
					redirectConfig: `{"allow_user_defined_redirect":true,"default_redirect_url":"http://test.hive.ory.sh:4000/"}`,
				},
			} {
				t.Run("hook=after/strategy="+tc.strategy, func(t *testing.T) {
					hooks := p.SelfServiceLoginAfterHooks(tc.strategy)

					hook := hooks[0]
					assert.EqualValues(t, "session", hook.Run)

					hook = hooks[1]
					assert.EqualValues(t, "redirect", hook.Run)
					assert.JSONEq(t, tc.redirectConfig, string(hook.Config))
				})
			}
		})

		t.Run("group=hashers", func(t *testing.T) {
			assert.Equal(t, &HasherArgon2Config{
				Memory:      1048576,
				Iterations:  2,
				Parallelism: 4,
				SaltLength:  16,
				KeyLength:   32,
			}, p.HashersArgon2())
		})
	})
}
