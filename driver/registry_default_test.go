package driver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ory/kratos/driver"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/hook"
)

func TestDriverDefault_Hooks(t *testing.T) {
	ctx := context.Background()
	t.Run("case=verification", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		conf.MustSet(config.ViperKeySelfServiceVerificationEnabled, true)

		t.Run("type=registration", func(t *testing.T) {
			h := reg.PostRegistrationPostPersistHooks(ctx, identity.CredentialsTypePassword)
			require.Len(t, h, 1)
			assert.Equal(t, []registration.PostHookPostPersistExecutor{hook.NewVerifier(reg)}, h)

			conf.MustSet(config.ViperKeySelfServiceRegistrationAfter+".password.hooks",
				[]map[string]interface{}{{"hook": "session"}})

			h = reg.PostRegistrationPostPersistHooks(ctx, identity.CredentialsTypePassword)
			require.Len(t, h, 2)
			assert.Equal(t, []registration.PostHookPostPersistExecutor{
				hook.NewVerifier(reg),
				hook.NewSessionIssuer(reg),
			}, h)
		})

		t.Run("type=login", func(t *testing.T) {
			h := reg.PostLoginHooks(ctx, identity.CredentialsTypePassword)
			require.Len(t, h, 0)

			conf.MustSet(config.ViperKeySelfServiceLoginAfter+".password.hooks",
				[]map[string]interface{}{{"hook": "revoke_active_sessions"}})

			h = reg.PostLoginHooks(ctx, identity.CredentialsTypePassword)
			require.Len(t, h, 1)
			assert.Equal(t, []login.PostHookExecutor{hook.NewSessionDestroyer(reg)}, h)
		})

		t.Run("type=settings", func(t *testing.T) {
			h := reg.PostSettingsPostPersistHooks(ctx, "profile")
			require.Len(t, h, 1)
			assert.Equal(t, []settings.PostHookPostPersistExecutor{hook.NewVerifier(reg)}, h)
		})
	})
}

func TestDriverDefault_Strategies(t *testing.T) {
	for k, tc := range []struct {
		prep   func(conf *config.Config)
		expect []string
	}{
		{prep: func(conf *config.Config) {
			conf.MustSet(config.ViperKeySelfServiceStrategyConfig+".password.enabled", false)
		}},
		{prep: func(conf *config.Config) {
			conf.MustSet(config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
		}, expect: []string{"password"}},
		{prep: func(conf *config.Config) {
			conf.MustSet(config.ViperKeySelfServiceStrategyConfig+".oidc.enabled", true)
			conf.MustSet(config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
		}, expect: []string{"password", "oidc"}},
	} {
		t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
			conf, reg := internal.NewFastRegistryWithMocks(t)
			tc.prep(conf)

			t.Run("case=registration", func(t *testing.T) {
				s := reg.RegistrationStrategies(context.Background())
				require.Len(t, s, len(tc.expect))
				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].ID().String())
				}
			})

			t.Run("case=login", func(t *testing.T) {
				s := reg.LoginStrategies(context.Background())
				require.Len(t, s, len(tc.expect))
				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].ID().String())
				}
			})
		})
	}

	t.Run("case=recovery", func(t *testing.T) {
		for k, tc := range []struct {
			prep   func(conf *config.Config)
			expect []string
		}{
			{prep: func(conf *config.Config) {
				conf.MustSet(config.ViperKeySelfServiceStrategyConfig+".link.enabled", false)
			}},
			{prep: func(conf *config.Config) {
				conf.MustSet(config.ViperKeySelfServiceStrategyConfig+".link.enabled", true)
			}, expect: []string{"link"}},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				conf, reg := internal.NewFastRegistryWithMocks(t)
				tc.prep(conf)

				s := reg.RecoveryStrategies(context.Background())
				require.Len(t, s, len(tc.expect))
				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].RecoveryStrategyID())
				}
			})
		}
	})

	t.Run("case=settings", func(t *testing.T) {
		l := logrusx.New("", "")

		for k, tc := range []struct {
			prep   func(t *testing.T) *config.Config
			expect []string
		}{
			{
				prep: func(t *testing.T) *config.Config {
					c := config.MustNew(l,
						configx.WithValues(map[string]interface{}{
							config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
							config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
							config.ViperKeySelfServiceStrategyConfig + ".oidc.enabled":     false,
							config.ViperKeySelfServiceStrategyConfig + ".profile.enabled":  false,
						}),
						configx.SkipValidation())
					return c
				}},
			{
				prep: func(t *testing.T) *config.Config {
					c := config.MustNew(l,
						configx.WithValues(map[string]interface{}{
							config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
							config.ViperKeySelfServiceStrategyConfig + ".profile.enabled":  true,
							config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
						}),
						configx.SkipValidation())
					return c
				},
				expect: []string{"profile"}},
			{
				prep: func(t *testing.T) *config.Config {
					return config.MustNew(l,
						configx.WithValues(map[string]interface{}{
							config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
						}),
						configx.SkipValidation())
				},
				expect: []string{"password", "profile"}},
			{
				prep: func(t *testing.T) *config.Config {
					return config.MustNew(l,
						configx.WithConfigFiles("../test/e2e/profiles/verification/.kratos.yml"),
						configx.WithValue(config.ViperKeyDSN, config.DefaultSQLiteMemoryDSN),
						configx.SkipValidation())
				},
				expect: []string{"password", "profile"}},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				conf := tc.prep(t)

				reg, err := driver.NewRegistryFromDSN(conf, logrusx.New("", ""))
				require.NoError(t, err)

				s := reg.SettingsStrategies(context.Background())
				require.Len(t, s, len(tc.expect))

				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].SettingsStrategyID())
				}
			})
		}
	})
}

func TestDefaultRegistry_AllStrategies(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)

	t.Run("case=all login strategies", func(t *testing.T) {
		expects := []string{"password", "oidc"}
		s := reg.AllLoginStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].ID().String())
		}
	})

	t.Run("case=all registration strategies", func(t *testing.T) {
		expects := []string{"password", "oidc"}
		s := reg.AllRegistrationStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].ID().String())
		}
	})

	t.Run("case=all settings strategies", func(t *testing.T) {
		expects := []string{"password", "oidc", "profile"}
		s := reg.AllSettingsStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].SettingsStrategyID())
		}
	})

	t.Run("case=all recovery strategies", func(t *testing.T) {
		expects := []string{"link"}
		s := reg.AllRecoveryStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].RecoveryStrategyID())
		}
	})
}
