// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/logrusx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequiredAALGetters(t *testing.T) {
	t.Parallel()

	for _, endpoint := range []struct {
		name    string
		key     string
		aal     func(*config.Config, context.Context) string
		methods func(*config.Config, context.Context) []string
	}{
		{"whoami", config.ViperKeySessionWhoAmIAAL, (*config.Config).SessionWhoAmIAAL, (*config.Config).SessionWhoAmIConsideredMethods},
		{"settings", config.ViperKeySelfServiceSettingsRequiredAAL, (*config.Config).SelfServiceSettingsRequiredAAL, (*config.Config).SelfServiceSettingsConsideredMethods},
	} {
		t.Run(endpoint.name, func(t *testing.T) {
			for _, tc := range []struct {
				name    string
				value   interface{}
				aal     string
				methods []string
			}{
				{name: "default", aal: "highest_available"},
				{name: "legacy aal1", value: "aal1", aal: "aal1"},
				{name: "legacy highest_available", value: "highest_available", aal: "highest_available"},
				{name: "object aal1", value: map[string]interface{}{"value": "aal1"}, aal: "aal1"},
				{name: "object omitted methods", value: map[string]interface{}{"value": "highest_available"}, aal: "highest_available"},
				{name: "object filtered methods", value: map[string]interface{}{"value": "highest_available", "considered_methods": []string{"totp", "code", "lookup_secret", "webauthn"}}, aal: "highest_available", methods: []string{"totp", "code", "lookup_secret", "webauthn"}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					values := map[string]interface{}{}
					if tc.value != nil {
						values[endpoint.key] = tc.value
					}
					p := config.MustNew(t, logrusx.New("", ""), &contextx.Default{}, configx.WithValues(values), configx.SkipValidation())
					assert.Equal(t, tc.aal, endpoint.aal(p, t.Context()))
					assert.Equal(t, tc.methods, endpoint.methods(p, t.Context()))
				})
			}
		})
	}
}

func TestRequiredAALSchema(t *testing.T) {
	t.Parallel()

	for _, path := range []string{"../../embedx/config.schema.json", "../../.schemastore/config.schema.json"} {
		t.Run(path, func(t *testing.T) {
			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			compiler := jsonschema.NewCompiler()
			const id = "https://github.com/ory/kratos/embedx/config.schema.json"
			require.NoError(t, compiler.AddResource(id, bytes.NewReader(raw)))
			for _, pointer := range []string{
				"/properties/session/properties/whoami/properties/required_aal",
				"/properties/selfservice/properties/flows/properties/settings/properties/required_aal",
			} {
				t.Run(pointer, func(t *testing.T) {
					schema, err := compiler.Compile(t.Context(), id+"#"+pointer)
					require.NoError(t, err)
					for _, tc := range []struct {
						name  string
						value string
						valid bool
					}{
						{"legacy aal1", `"aal1"`, true},
						{"legacy highest_available", `"highest_available"`, true},
						{"object aal1", `{"value":"aal1"}`, true},
						{"object omitted methods", `{"value":"highest_available"}`, true},
						{"single method", `{"value":"highest_available","considered_methods":["code"]}`, true},
						{"all methods", `{"value":"highest_available","considered_methods":["totp","code","lookup_secret","webauthn"]}`, true},
						{"unknown legacy aal", `"aal2"`, false},
						{"null", `null`, false},
						{"number", `1`, false},
						{"array", `[]`, false},
						{"missing value", `{}`, false},
						{"filter without value", `{"considered_methods":["totp"]}`, false},
						{"unknown value", `{"value":"aal2"}`, false},
						{"malformed value", `{"value":true}`, false},
						{"unknown property", `{"value":"highest_available","unknown":true}`, false},
						{"empty filter", `{"value":"highest_available","considered_methods":[]}`, false},
						{"null filter", `{"value":"highest_available","considered_methods":null}`, false},
						{"string filter", `{"value":"highest_available","considered_methods":"totp"}`, false},
						{"malformed method", `{"value":"highest_available","considered_methods":[1]}`, false},
						{"unknown method", `{"value":"highest_available","considered_methods":["password"]}`, false},
						{"duplicate method", `{"value":"highest_available","considered_methods":["totp","totp"]}`, false},
						{"aal1 with filter", `{"value":"aal1","considered_methods":["totp"]}`, false},
					} {
						t.Run(tc.name, func(t *testing.T) {
							err := schema.Validate(strings.NewReader(tc.value))
							if tc.valid {
								require.NoError(t, err)
							} else {
								require.Error(t, err)
							}
						})
					}
				})
			}
		})
	}
}
