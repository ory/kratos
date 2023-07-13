// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ory/kratos/selfservice/flow/recovery"

	"github.com/ory/kratos/selfservice/flow/verification"

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
	t.Parallel()
	ctx := context.Background()

	t.Run("type=verification", func(t *testing.T) {
		t.Parallel()
		// BEFORE hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []verification.PreHookExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []verification.PreHookExecutor { return nil },
			},
			{
				uc: "Two web_hooks are configured",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationBeforeHooks, []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []verification.PreHookExecutor {
					return []verification.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PreVerificationHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}

		// AFTER hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []verification.PostHookExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []verification.PostHookExecutor { return nil },
			},
			{
				uc: "Multiple web_hooks configured",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationAfter+".hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []verification.PostHookExecutor {
					return []verification.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PostVerificationHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}
	})

	t.Run("type=recovery", func(t *testing.T) {
		t.Parallel()
		// BEFORE hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []recovery.PreHookExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []recovery.PreHookExecutor { return nil },
			},
			{
				uc: "Two web_hooks are configured",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBeforeHooks, []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []recovery.PreHookExecutor {
					return []recovery.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PreRecoveryHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}

		// AFTER hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []recovery.PostHookExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []recovery.PostHookExecutor { return nil },
			},
			{
				uc: "Multiple web_hooks configured",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryAfter+".hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []recovery.PostHookExecutor {
					return []recovery.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PostRecoveryHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}
	})

	t.Run("type=registration", func(t *testing.T) {
		t.Parallel()
		// BEFORE hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []registration.PreHookExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []registration.PreHookExecutor { return nil },
			},
			{
				uc: "Two web_hooks are configured",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationBeforeHooks, []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []registration.PreHookExecutor {
					return []registration.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PreRegistrationHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}

		// AFTER hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor { return nil },
			},
			{
				uc: "Only session hook configured for password strategy",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".password.hooks", []map[string]interface{}{
						{"hook": "session"},
					})
				},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor {
					return []registration.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
						hook.NewSessionIssuer(reg),
					}
				},
			},
			{
				uc: "A session hook and a web_hook are configured for password strategy",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".password.hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST", "body": "bar"}},
						{"hook": "session"},
					})
				},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor {
					return []registration.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
						hook.NewWebHook(reg, json.RawMessage(`{"body":"bar","method":"POST","url":"foo"}`)),
						hook.NewSessionIssuer(reg),
					}
				},
			},
			{
				uc: "Two web_hooks are configured on a global level",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor {
					return []registration.PostHookPostPersistExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
			{
				uc: "Hooks are configured on a global level, as well as on a strategy level",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".password.hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "GET"}},
						{"hook": "session"},
					})
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "POST"}},
					})
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
				},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor {
					return []registration.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"foo"}`)),
						hook.NewSessionIssuer(reg),
					}
				},
			},
			{
				uc: "show_verification_ui is configured",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".hooks", []map[string]interface{}{
						{"hook": "show_verification_ui"},
					})
				},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor {
					return []registration.PostHookPostPersistExecutor{
						hook.NewShowVerificationUIHook(reg),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PostRegistrationPostPersistHooks(ctx, identity.CredentialsTypePassword)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}
	})

	t.Run("type=login", func(t *testing.T) {
		t.Parallel()
		// BEFORE hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []login.PreHookExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []login.PreHookExecutor { return nil },
			},
			{
				uc: "Two web_hooks are configured",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceLoginBeforeHooks, []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []login.PreHookExecutor {
					return []login.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PreLoginHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}

		// AFTER hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []login.PostHookExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor { return nil },
			},
			{
				uc: "Only revoke_active_sessions hook configured for password strategy",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+".password.hooks", []map[string]interface{}{
						{"hook": "revoke_active_sessions"},
					})
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewSessionDestroyer(reg),
					}
				},
			},
			{
				uc: "Only require_verified_address hook configured for password strategy",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+".password.hooks", []map[string]interface{}{
						{"hook": "require_verified_address"},
					})
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewAddressVerifier(),
					}
				},
			},
			{
				uc: "A revoke_active_sessions hook, require_verified_address hook and a web_hook are configured for password strategy",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+".password.hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST", "body": "bar"}},
						{"hook": "require_verified_address"},
						{"hook": "revoke_active_sessions"},
					})
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"body":"bar","method":"POST","url":"foo"}`)),
						hook.NewAddressVerifier(),
						hook.NewSessionDestroyer(reg),
					}
				},
			},
			{
				uc: "Two web_hooks are configured on a global level",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+".hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
			{
				uc: "Hooks are configured on a global level, as well as on a strategy level",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+".password.hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "GET"}},
						{"hook": "revoke_active_sessions"},
						{"hook": "require_verified_address"},
					})
					conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+".hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"foo"}`)),
						hook.NewSessionDestroyer(reg),
						hook.NewAddressVerifier(),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PostLoginHooks(ctx, identity.CredentialsTypePassword)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}
	})

	t.Run("type=settings", func(t *testing.T) {
		t.Parallel()
		// BEFORE hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []settings.PreHookExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []settings.PreHookExecutor { return nil },
			},
			{
				uc: "Two web_hooks are configured",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceSettingsBeforeHooks, []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []settings.PreHookExecutor {
					return []settings.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PreSettingsHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}

		// AFTER hooks
		for _, tc := range []struct {
			uc     string
			prep   func(conf *config.Config)
			expect func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor { return nil },
			},
			{
				uc: "Only verify hook configured for the strategy",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
					// I think this is a bug as there is a hook named verify defined for both profile and password
					// strategies. Instead of using it, the code makes use of the property used above and which
					// is defined in an entirely different flow (verification).
				},
				expect: func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor {
					return []settings.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
					}
				},
			},
			{
				uc: "A verify hook and a web_hook are configured for profile strategy",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter+".profile.hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST", "body": "bar"}},
					})
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
				},
				expect: func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor {
					return []settings.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
						hook.NewWebHook(reg, json.RawMessage(`{"body":"bar","method":"POST","url":"foo"}`)),
					}
				},
			},
			{
				uc: "Two web_hooks are configured on a global level",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter+".hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
						{"hook": "web_hook", "config": map[string]interface{}{"url": "bar", "method": "GET"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor {
					return []settings.PostHookPostPersistExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"bar"}`)),
					}
				},
			},
			{
				uc: "Hooks are configured on a global level, as well as on a strategy level",
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
					conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter+".profile.hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "GET"}},
					})
					conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter+".hooks", []map[string]interface{}{
						{"hook": "web_hook", "config": map[string]interface{}{"url": "foo", "method": "POST"}},
					})
				},
				expect: func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor {
					return []settings.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
						hook.NewWebHook(reg, json.RawMessage(`{"method":"GET","url":"foo"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				h := reg.PostSettingsPostPersistHooks(ctx, "profile")

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}
	})
}

func TestDriverDefault_Strategies(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Run("case=registration", func(t *testing.T) {
		t.Parallel()
		for k, tc := range []struct {
			prep   func(conf *config.Config)
			expect []string
		}{
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", false)
				},
			},
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
				},
				expect: []string{"password"},
			},
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".oidc.enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
				},
				expect: []string{"password", "oidc"},
			},
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".oidc.enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".totp.enabled", true)
				},
				expect: []string{"password", "oidc"},
			},
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.registration_enabled", true)
				},
				expect: []string{"password", "code"},
			},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				s := reg.RegistrationStrategies(context.Background())
				require.Len(t, s, len(tc.expect))
				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].ID().String())
				}
			})
		}
	})

	t.Run("case=login", func(t *testing.T) {
		t.Parallel()
		for k, tc := range []struct {
			prep   func(conf *config.Config)
			expect []string
		}{
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", false)
				},
			},
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
				},
				expect: []string{"password"},
			},
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".oidc.enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
				},
				expect: []string{"password", "oidc"},
			},
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".oidc.enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".totp.enabled", true)
				},
				expect: []string{"password", "oidc", "totp"},
			},
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.login_enabled", true)
				},
				expect: []string{"password", "code"},
			},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
				tc.prep(conf)

				s := reg.LoginStrategies(context.Background())
				require.Len(t, s, len(tc.expect))
				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].ID().String())
				}
			})
		}
	})

	t.Run("case=recovery", func(t *testing.T) {
		t.Parallel()
		for k, tc := range []struct {
			prep   func(conf *config.Config)
			expect []string
		}{
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.enabled", false)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".link.enabled", false)
				},
			},
			{
				prep: func(conf *config.Config) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".link.enabled", true)
				}, expect: []string{"code", "link"},
			},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
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
		t.Parallel()
		l := logrusx.New("", "")

		for k, tc := range []struct {
			prep   func(t *testing.T) *config.Config
			expect []string
		}{
			{
				prep: func(t *testing.T) *config.Config {
					c := config.MustNew(t, l,
						os.Stderr,
						configx.WithValues(map[string]interface{}{
							config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
							config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
							config.ViperKeySelfServiceStrategyConfig + ".oidc.enabled":     false,
							config.ViperKeySelfServiceStrategyConfig + ".profile.enabled":  false,
						}),
						configx.SkipValidation())
					return c
				},
			},
			{
				prep: func(t *testing.T) *config.Config {
					c := config.MustNew(t, l,
						os.Stderr,
						configx.WithValues(map[string]interface{}{
							config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
							config.ViperKeySelfServiceStrategyConfig + ".profile.enabled":  true,
							config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
						}),
						configx.SkipValidation())
					return c
				},
				expect: []string{"profile"},
			},
			{
				prep: func(t *testing.T) *config.Config {
					c := config.MustNew(t, l,
						os.Stderr,
						configx.WithValues(map[string]interface{}{
							config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
							config.ViperKeySelfServiceStrategyConfig + ".profile.enabled":  true,
							config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
							config.ViperKeySelfServiceStrategyConfig + ".totp.enabled":     true,
						}),
						configx.SkipValidation())
					return c
				},
				expect: []string{"profile", "totp"},
			},
			{
				prep: func(t *testing.T) *config.Config {
					return config.MustNew(t, l,
						os.Stderr,
						configx.WithValues(map[string]interface{}{
							config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
						}),
						configx.SkipValidation())
				},
				expect: []string{"password", "profile"},
			},
			{
				prep: func(t *testing.T) *config.Config {
					return config.MustNew(t, l,
						os.Stderr,
						configx.WithConfigFiles("../test/e2e/profiles/verification/.kratos.yml"),
						configx.WithValue(config.ViperKeyDSN, config.DefaultSQLiteMemoryDSN),
						configx.SkipValidation())
				},
				expect: []string{"password", "profile"},
			},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				conf := tc.prep(t)

				reg, err := driver.NewRegistryFromDSN(ctx, conf, logrusx.New("", ""))
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
	t.Parallel()
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	t.Run("case=all login strategies", func(t *testing.T) {
		expects := []string{"password", "oidc", "code", "totp", "webauthn", "lookup_secret"}
		s := reg.AllLoginStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].ID().String())
		}
	})

	t.Run("case=all registration strategies", func(t *testing.T) {
		expects := []string{"password", "oidc", "code", "webauthn"}
		s := reg.AllRegistrationStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].ID().String())
		}
	})

	t.Run("case=all settings strategies", func(t *testing.T) {
		expects := []string{"password", "oidc", "profile", "totp", "webauthn", "lookup_secret"}
		s := reg.AllSettingsStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].SettingsStrategyID())
		}
	})

	t.Run("case=all recovery strategies", func(t *testing.T) {
		expects := []string{"code", "link"}
		s := reg.AllRecoveryStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].RecoveryStrategyID())
		}
	})
}

func TestGetActiveRecoveryStrategy(t *testing.T) {
	t.Parallel()
	conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
	t.Run("returns error if active strategy is disabled", func(t *testing.T) {
		conf.Set(context.Background(), "selfservice.methods.code.enabled", false)
		conf.Set(context.Background(), config.ViperKeySelfServiceRecoveryUse, "code")

		_, err := reg.GetActiveRecoveryStrategy(context.Background())
		require.Error(t, err)
	})

	t.Run("returns active strategy", func(t *testing.T) {
		for _, sID := range []string{
			"code", "link",
		} {
			t.Run(fmt.Sprintf("strategy=%s", sID), func(t *testing.T) {
				conf.Set(context.Background(), fmt.Sprintf("selfservice.methods.%s.enabled", sID), true)
				conf.Set(context.Background(), config.ViperKeySelfServiceRecoveryUse, sID)

				s, err := reg.GetActiveRecoveryStrategy(context.Background())
				require.NoError(t, err)
				require.Equal(t, sID, s.RecoveryStrategyID())
			})
		}
	})
}

func TestGetActiveVerificationStrategy(t *testing.T) {
	t.Parallel()
	conf, reg := internal.NewVeryFastRegistryWithoutDB(t)
	t.Run("returns error if active strategy is disabled", func(t *testing.T) {
		conf.Set(context.Background(), "selfservice.methods.code.enabled", false)
		conf.Set(context.Background(), config.ViperKeySelfServiceVerificationUse, "code")

		_, err := reg.GetActiveVerificationStrategy(context.Background())
		require.Error(t, err)
	})

	t.Run("returns active strategy", func(t *testing.T) {
		for _, sID := range []string{
			"code", "link",
		} {
			t.Run(fmt.Sprintf("strategy=%s", sID), func(t *testing.T) {
				conf.Set(context.Background(), fmt.Sprintf("selfservice.methods.%s.enabled", sID), true)
				conf.Set(context.Background(), config.ViperKeySelfServiceVerificationUse, sID)

				s, err := reg.GetActiveVerificationStrategy(context.Background())
				require.NoError(t, err)
				require.Equal(t, sID, s.VerificationStrategyID())
			})
		}
	})
}
