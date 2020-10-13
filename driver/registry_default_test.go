package driver_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/viperx"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/hook"
)

func TestDriverDefault_Hooks(t *testing.T) {
	t.Run("case=verification", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		viper.Set(configuration.ViperKeySelfServiceVerificationEnabled, true)

		t.Run("type=registration", func(t *testing.T) {
			h := reg.PostRegistrationPostPersistHooks(identity.CredentialsTypePassword)
			require.Len(t, h, 1)
			assert.Equal(t, []registration.PostHookPostPersistExecutor{hook.NewVerifier(reg, conf)}, h)

			viper.Set(configuration.ViperKeySelfServiceRegistrationAfter+".password.hooks",
				[]map[string]interface{}{{"hook": "session"}})

			h = reg.PostRegistrationPostPersistHooks(identity.CredentialsTypePassword)
			require.Len(t, h, 2)
			assert.Equal(t, []registration.PostHookPostPersistExecutor{
				hook.NewVerifier(reg, conf),
				hook.NewSessionIssuer(reg),
			}, h)
		})

		t.Run("type=login", func(t *testing.T) {
			h := reg.PostLoginHooks(identity.CredentialsTypePassword)
			require.Len(t, h, 0)

			viper.Set(configuration.ViperKeySelfServiceLoginAfter+".password.hooks",
				[]map[string]interface{}{{"hook": "revoke_active_sessions"}})

			h = reg.PostLoginHooks(identity.CredentialsTypePassword)
			require.Len(t, h, 1)
			assert.Equal(t, []login.PostHookExecutor{hook.NewSessionDestroyer(reg)}, h)
		})

		t.Run("type=settings", func(t *testing.T) {
			h := reg.PostSettingsPostPersistHooks("profile")
			require.Len(t, h, 1)
			assert.Equal(t, []settings.PostHookPostPersistExecutor{hook.NewVerifier(reg, conf)}, h)
		})
	})
}

func TestDriverDefault_Strategies(t *testing.T) {
	for k, tc := range []struct {
		prep   func()
		expect []string
	}{
		{prep: func() {
			viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".password.enabled", false)
		}},
		{prep: func() {
			viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
		}, expect: []string{"password"}},
		{prep: func() {
			viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".oidc.enabled", true)
			viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
		}, expect: []string{"password", "oidc"}},
	} {
		t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
			_, reg := internal.NewFastRegistryWithMocks(t)
			tc.prep()

			t.Run("case=registration", func(t *testing.T) {
				s := reg.RegistrationStrategies()
				require.Len(t, s, len(tc.expect))
				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].ID().String())
				}
			})

			t.Run("case=login", func(t *testing.T) {
				s := reg.LoginStrategies()
				require.Len(t, s, len(tc.expect))
				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].ID().String())
				}
			})
		})
	}

	t.Run("case=recovery", func(t *testing.T) {
		for k, tc := range []struct {
			prep   func()
			expect []string
		}{
			{prep: func() {
				viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".link.enabled", false)
			}},
			{prep: func() {
				viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".link.enabled", true)
			}, expect: []string{"link"}},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				_, reg := internal.NewFastRegistryWithMocks(t)
				tc.prep()

				s := reg.RecoveryStrategies()
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
			prep   func()
			expect []string
		}{
			{prep: func() {
				viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".password.enabled", false)
				viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".oidc.enabled", false)
				viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".profile.enabled", false)
			}},
			{prep: func() {
				viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".profile.enabled", true)
				viper.Set(configuration.ViperKeySelfServiceStrategyConfig+".password.enabled", false)
			}, expect: []string{"profile"}},
			{prep: func() {}, expect: []string{"password", "profile"}},
			{prep: func() {
				viperx.InitializeConfig("kratos", "../test/e2e/profiles/verification", l)
			}, expect: []string{"password", "profile"}},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				_, reg := internal.NewFastRegistryWithMocks(t)
				tc.prep()

				s := reg.SettingsStrategies()
				require.Len(t, s, len(tc.expect))
				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].SettingsStrategyID())
				}
			})
		}
	})
}
