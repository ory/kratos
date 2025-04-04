// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	confighelpers "github.com/ory/kratos/driver/config/testhelpers"

	"github.com/ory/x/contextx"

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

	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	t.Run("type=verification", func(t *testing.T) {
		t.Parallel()
		// BEFORE hooks
		for _, tc := range []struct {
			uc     string
			config map[string]any
			expect func(reg *driver.RegistryDefault) []verification.PreHookExecutor
		}{
			{
				uc:     "No hooks configured",
				expect: func(reg *driver.RegistryDefault) []verification.PreHookExecutor { return nil },
			},
			{
				uc: "Two web_hooks are configured",
				config: map[string]any{
					config.ViperKeySelfServiceVerificationBeforeHooks: []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []verification.PreHookExecutor {
					return []verification.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

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
			config map[string]any
			expect func(reg *driver.RegistryDefault) []verification.PostHookExecutor
		}{
			{
				uc:     "No hooks configured",
				prep:   func(conf *config.Config) {},
				expect: func(reg *driver.RegistryDefault) []verification.PostHookExecutor { return nil },
			},
			{
				uc: "Multiple web_hooks configured",
				config: map[string]any{
					config.ViperKeySelfServiceVerificationAfter + ".hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []verification.PostHookExecutor {
					return []verification.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

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
			config map[string]any
			expect func(reg *driver.RegistryDefault) []recovery.PreHookExecutor
		}{
			{
				uc:     "No hooks configured",
				expect: func(reg *driver.RegistryDefault) []recovery.PreHookExecutor { return nil },
			},
			{
				uc: "Two web_hooks are configured",
				config: map[string]any{
					config.ViperKeySelfServiceRecoveryBeforeHooks: []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []recovery.PreHookExecutor {
					return []recovery.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

				h := reg.PreRecoveryHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}

		// AFTER hooks
		for _, tc := range []struct {
			uc     string
			config map[string]any
			expect func(reg *driver.RegistryDefault) []recovery.PostHookExecutor
		}{
			{
				uc:     "No hooks configured",
				expect: func(reg *driver.RegistryDefault) []recovery.PostHookExecutor { return nil },
			},
			{
				uc: "Multiple web_hooks configured",
				config: map[string]any{
					config.ViperKeySelfServiceRecoveryAfter + ".hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []recovery.PostHookExecutor {
					return []recovery.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

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
			config map[string]any
			expect func(reg *driver.RegistryDefault) []registration.PreHookExecutor
		}{
			{
				uc: "No hooks configured",
				expect: func(reg *driver.RegistryDefault) []registration.PreHookExecutor {
					return nil
				},
			},
			{
				uc: "Two web_hooks are configured",
				config: map[string]any{
					config.ViperKeySelfServiceRegistrationBeforeHooks: []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []registration.PreHookExecutor {
					return []registration.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

				h := reg.PreRegistrationHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.EqualValues(t, expectedExecutors, h)
			})
		}

		// AFTER hooks
		for _, tc := range []struct {
			uc     string
			config map[string]any
			expect func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor
		}{
			{
				uc:     "No hooks configured",
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor { return nil },
			},
			{
				uc: "Only session hook configured for password strategy",
				config: map[string]any{
					config.ViperKeySelfServiceVerificationEnabled: true,
					config.ViperKeySelfServiceRegistrationAfter + ".password.hooks": []map[string]any{
						{"hook": "session"},
					},
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
				config: map[string]any{
					config.ViperKeySelfServiceVerificationEnabled: true,
					config.ViperKeySelfServiceRegistrationAfter + ".password.hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"headers": map[string]string{"X-Custom-Header": "test"}, "url": "foo", "method": "POST", "body": "bar"}},
						{"hook": "session"},
					},
				},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor {
					return []registration.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
						hook.NewWebHook(reg, json.RawMessage(`{"body":"bar","headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewSessionIssuer(reg),
					}
				},
			},
			{
				uc: "Two web_hooks are configured on a global level",
				config: map[string]any{
					config.ViperKeySelfServiceRegistrationAfter + ".hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor {
					return []registration.PostHookPostPersistExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
			{
				uc: "Hooks are configured on a global level, as well as on a strategy level",
				config: map[string]any{
					config.ViperKeySelfServiceRegistrationAfter + ".password.hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "session"},
					},
					config.ViperKeySelfServiceRegistrationAfter + ".hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
					config.ViperKeySelfServiceVerificationEnabled: true,
				},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor {
					return []registration.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"foo"}`)),
						hook.NewSessionIssuer(reg),
					}
				},
			},
			{
				uc: "show_verification_ui is configured",
				config: map[string]any{
					config.ViperKeySelfServiceRegistrationAfter + ".hooks": []map[string]any{
						{"hook": "show_verification_ui"},
					},
				},
				expect: func(reg *driver.RegistryDefault) []registration.PostHookPostPersistExecutor {
					return []registration.PostHookPostPersistExecutor{
						hook.NewShowVerificationUIHook(reg),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

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
			config map[string]any
			expect func(reg *driver.RegistryDefault) []login.PreHookExecutor
		}{
			{
				uc:     "No hooks configured",
				expect: func(reg *driver.RegistryDefault) []login.PreHookExecutor { return nil },
			},
			{
				uc: "Two web_hooks are configured",
				config: map[string]any{
					config.ViperKeySelfServiceLoginBeforeHooks: []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []login.PreHookExecutor {
					return []login.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

				h := reg.PreLoginHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}

		// AFTER hooks
		for _, tc := range []struct {
			uc     string
			config map[string]any
			expect func(reg *driver.RegistryDefault) []login.PostHookExecutor
		}{
			{
				uc:     "No hooks configured",
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor { return nil },
			},
			{
				uc: "Only revoke_active_sessions hook configured for password strategy",
				config: map[string]any{
					config.ViperKeySelfServiceLoginAfter + ".password.hooks": []map[string]any{
						{"hook": "revoke_active_sessions"},
					},
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewSessionDestroyer(reg),
					}
				},
			},
			{
				uc: "Only require_verified_address hook configured for password strategy",
				config: map[string]any{
					config.ViperKeySelfServiceLoginAfter + ".password.hooks": []map[string]any{
						{"hook": "require_verified_address"},
					},
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewAddressVerifier(),
					}
				},
			},
			{
				uc: "A revoke_active_sessions hook, require_verified_address hook and a web_hook are configured for password strategy",
				config: map[string]any{
					config.ViperKeySelfServiceLoginAfter + ".password.hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"headers": map[string]string{"X-Custom-Header": "test"}, "url": "foo", "method": "POST", "body": "bar"}},
						{"hook": "require_verified_address"},
						{"hook": "revoke_active_sessions"},
					},
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"body":"bar","headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewAddressVerifier(),
						hook.NewSessionDestroyer(reg),
					}
				},
			},
			{
				uc: "Two web_hooks are configured on a global level",
				config: map[string]any{
					config.ViperKeySelfServiceLoginAfter + ".hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
			{
				uc: "Hooks are configured on a global level, as well as on a strategy level",
				config: map[string]any{
					config.ViperKeySelfServiceLoginAfter + ".password.hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "revoke_active_sessions"},
						{"hook": "require_verified_address"},
					},
					config.ViperKeySelfServiceLoginAfter + ".hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []login.PostHookExecutor {
					return []login.PostHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"foo"}`)),
						hook.NewSessionDestroyer(reg),
						hook.NewAddressVerifier(),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

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
			config map[string]any
			expect func(reg *driver.RegistryDefault) []settings.PreHookExecutor
		}{
			{
				uc:     "No hooks configured",
				expect: func(reg *driver.RegistryDefault) []settings.PreHookExecutor { return nil },
			},
			{
				uc: "Two web_hooks are configured",
				config: map[string]any{
					config.ViperKeySelfServiceSettingsBeforeHooks: []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []settings.PreHookExecutor {
					return []settings.PreHookExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("before/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

				h := reg.PreSettingsHooks(ctx)

				expectedExecutors := tc.expect(reg)
				require.Len(t, h, len(expectedExecutors))
				assert.Equal(t, expectedExecutors, h)
			})
		}

		// AFTER hooks
		for _, tc := range []struct {
			uc     string
			config map[string]any
			expect func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor
		}{
			{
				uc:     "No hooks configured",
				expect: func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor { return nil },
			},
			{
				uc: "Only verify hook configured for the strategy",
				config: map[string]any{
					config.ViperKeySelfServiceVerificationEnabled: true,
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
				config: map[string]any{
					config.ViperKeySelfServiceSettingsAfter + ".profile.hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"headers": []map[string]string{{"X-Custom-Header": "test"}}, "url": "foo", "method": "POST", "body": "bar"}},
					},
					config.ViperKeySelfServiceVerificationEnabled: true,
				},
				expect: func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor {
					return []settings.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
						hook.NewWebHook(reg, json.RawMessage(`{"body":"bar","headers":[{"X-Custom-Header":"test"}],"method":"POST","url":"foo"}`)),
					}
				},
			},
			{
				uc: "Two web_hooks are configured on a global level",
				config: map[string]any{
					config.ViperKeySelfServiceSettingsAfter + ".hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
						{"hook": "web_hook", "config": map[string]any{"url": "bar", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor {
					return []settings.PostHookPostPersistExecutor{
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"POST","url":"foo"}`)),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"bar"}`)),
					}
				},
			},
			{
				uc: "Hooks are configured on a global level, as well as on a strategy level",
				config: map[string]any{
					config.ViperKeySelfServiceVerificationEnabled: true,
					config.ViperKeySelfServiceSettingsAfter + ".profile.hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "GET", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
					config.ViperKeySelfServiceSettingsAfter + ".hooks": []map[string]any{
						{"hook": "web_hook", "config": map[string]any{"url": "foo", "method": "POST", "headers": map[string]string{"X-Custom-Header": "test"}}},
					},
				},
				expect: func(reg *driver.RegistryDefault) []settings.PostHookPostPersistExecutor {
					return []settings.PostHookPostPersistExecutor{
						hook.NewVerifier(reg),
						hook.NewWebHook(reg, json.RawMessage(`{"headers":{"X-Custom-Header":"test"},"method":"GET","url":"foo"}`)),
					}
				},
			},
		} {
			t.Run(fmt.Sprintf("after/uc=%s", tc.uc), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

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
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	t.Run("case=registration", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct {
			name   string
			config map[string]any
			expect []string
		}{
			{
				name: "no strategies",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     false,
				},
				expect: []string{"profile"},
			},
			{
				name: "only password",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": true,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     false,
				},
				expect: []string{"profile", "password"},
			},
			{
				name: "oidc and password",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".oidc.enabled":     true,
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": true,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     false,
				},
				expect: []string{"profile", "password", "oidc"},
			},
			{
				name: "oidc, password and totp",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".oidc.enabled":     true,
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": true,
					config.ViperKeySelfServiceStrategyConfig + ".totp.enabled":     true,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     false,
				},
				expect: []string{"profile", "password", "oidc"},
			},
			{
				name: "password and code",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": true,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     true,
				},
				expect: []string{"profile", "password", "code"},
			},
		} {
			t.Run(fmt.Sprintf("subcase=%s", tc.name), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)
				s := reg.RegistrationStrategies(ctx)
				require.Len(t, s, len(tc.expect))
				for k, e := range tc.expect {
					assert.Equal(t, e, s[k].ID().String())
				}
			})
		}
	})

	t.Run("case=login", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name   string
			config map[string]any
			expect []string
		}{
			{
				name: "no strategies",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     false,
				},
			},
			{
				name: "only password",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": true,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     false,
				},
				expect: []string{"password"},
			},
			{
				name: "oidc and password",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".oidc.enabled":     true,
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": true,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     false,
				},
				expect: []string{"password", "oidc"},
			},
			{
				name: "oidc, password and totp",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".oidc.enabled":     true,
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": true,
					config.ViperKeySelfServiceStrategyConfig + ".totp.enabled":     true,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     false,
				},
				expect: []string{"password", "oidc", "totp"},
			},
			{
				name: "password and code",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": true,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":     true,
				},
				expect: []string{"password", "code"},
			},
			{
				name: "code is enabled if passwordless_enabled is true",
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled":          false,
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled":              false,
					config.ViperKeySelfServiceStrategyConfig + ".code.passwordless_enabled": true,
				},
				expect: []string{"code"},
			},
		} {
			t.Run(fmt.Sprintf("run=%s", tc.name), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)
				s := reg.LoginStrategies(ctx)
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
			config map[string]any
			expect []string
		}{
			{
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled": false,
					config.ViperKeySelfServiceStrategyConfig + ".link.enabled": false,
				},
			},
			{
				config: map[string]any{
					config.ViperKeySelfServiceStrategyConfig + ".code.enabled": true,
					config.ViperKeySelfServiceStrategyConfig + ".link.enabled": true,
				}, expect: []string{"code", "link"},
			},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				t.Parallel()

				ctx := confighelpers.WithConfigValues(ctx, tc.config)

				s := reg.RecoveryStrategies(ctx)
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
			configOptions []configx.OptionModifier
			expect        []string
		}{
			{
				configOptions: []configx.OptionModifier{configx.WithValues(map[string]any{
					config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
					config.ViperKeySelfServiceStrategyConfig + ".oidc.enabled":     false,
					config.ViperKeySelfServiceStrategyConfig + ".profile.enabled":  false,
				})},
			},
			{
				configOptions: []configx.OptionModifier{configx.WithValues(map[string]any{
					config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
					config.ViperKeySelfServiceStrategyConfig + ".profile.enabled":  true,
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
				})},
				expect: []string{"profile"},
			},
			{
				configOptions: []configx.OptionModifier{configx.WithValues(map[string]any{
					config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
					config.ViperKeySelfServiceStrategyConfig + ".profile.enabled":  true,
					config.ViperKeySelfServiceStrategyConfig + ".password.enabled": false,
					config.ViperKeySelfServiceStrategyConfig + ".totp.enabled":     true,
				})},
				expect: []string{"profile", "totp"},
			},
			{
				configOptions: []configx.OptionModifier{configx.WithValues(map[string]any{
					config.ViperKeyDSN: config.DefaultSQLiteMemoryDSN,
				})},
				expect: []string{"profile", "password"},
			},
			{
				configOptions: []configx.OptionModifier{
					configx.WithConfigFiles("../test/e2e/profiles/verification/.kratos.yml"),
					configx.WithValue(config.ViperKeyDSN, config.DefaultSQLiteMemoryDSN),
				},
				expect: []string{"profile", "password"},
			},
		} {
			t.Run(fmt.Sprintf("run=%d", k), func(t *testing.T) {
				conf := config.MustNew(t, l, os.Stderr, &contextx.Default{}, append(tc.configOptions, configx.SkipValidation())...)

				reg, err := driver.NewRegistryFromDSN(ctx, conf, l)
				require.NoError(t, err)

				s := reg.SettingsStrategies(ctx)
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
		expects := []string{"password", "oidc", "code", "totp", "passkey", "webauthn", "lookup_secret", "identifier_first"}
		s := reg.AllLoginStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].ID().String())
		}
	})

	t.Run("case=all registration strategies", func(t *testing.T) {
		expects := []string{"profile", "password", "oidc", "code", "passkey", "webauthn"}
		s := reg.AllRegistrationStrategies()
		require.Len(t, s, len(expects))
		for k, e := range expects {
			assert.Equal(t, e, s[k].ID().String())
		}
	})

	t.Run("case=all settings strategies", func(t *testing.T) {
		expects := []string{"profile", "password", "oidc", "totp", "passkey", "webauthn", "lookup_secret"}
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
	ctx := context.Background()
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	t.Run("returns error if active strategy is disabled", func(t *testing.T) {
		ctx := confighelpers.WithConfigValues(ctx, map[string]any{
			"selfservice.methods.code.enabled":    false,
			config.ViperKeySelfServiceRecoveryUse: "code",
		})

		_, err := reg.GetActiveRecoveryStrategy(ctx)
		require.Error(t, err)
	})

	t.Run("returns active strategy", func(t *testing.T) {
		for _, sID := range []string{
			"code", "link",
		} {
			t.Run(fmt.Sprintf("strategy=%s", sID), func(t *testing.T) {
				ctx := confighelpers.WithConfigValues(ctx, map[string]any{
					fmt.Sprintf("selfservice.methods.%s.enabled", sID): true,
					config.ViperKeySelfServiceRecoveryUse:              sID,
				})

				s, err := reg.GetActiveRecoveryStrategy(ctx)
				require.NoError(t, err)
				require.Equal(t, sID, s.RecoveryStrategyID())
			})
		}
	})
}

func TestGetActiveVerificationStrategy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)
	t.Run("returns error if active strategy is disabled", func(t *testing.T) {
		ctx := confighelpers.WithConfigValues(ctx, map[string]any{
			"selfservice.methods.code.enabled":        false,
			config.ViperKeySelfServiceVerificationUse: "code",
		})
		_, err := reg.GetActiveVerificationStrategy(ctx)
		require.Error(t, err)
	})

	t.Run("returns active strategy", func(t *testing.T) {
		for _, sID := range []string{
			"code", "link",
		} {
			t.Run(fmt.Sprintf("strategy=%s", sID), func(t *testing.T) {
				ctx := confighelpers.WithConfigValues(ctx, map[string]any{
					fmt.Sprintf("selfservice.methods.%s.enabled", sID): true,
					config.ViperKeySelfServiceVerificationUse:          sID,
				})

				s, err := reg.GetActiveVerificationStrategy(ctx)
				require.NoError(t, err)
				require.Equal(t, sID, s.VerificationStrategyID())
			})
		}
	})
}
