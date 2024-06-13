// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"strings"

	"github.com/knadh/koanf/maps"

	"github.com/ory/kratos/embedx"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
)

type (
	TestConfigProvider struct {
		contextx.Contextualizer
		Options []configx.OptionModifier
	}
	contextKey  int
	mapProvider map[string]any
)

func (t *TestConfigProvider) NewProvider(ctx context.Context, opts ...configx.OptionModifier) (*configx.Provider, error) {
	return configx.New(ctx, []byte(embedx.ConfigSchema), append(t.Options, opts...)...)
}

func (t *TestConfigProvider) Config(ctx context.Context, config *configx.Provider) *configx.Provider {
	config = t.Contextualizer.Config(ctx, config)
	values, ok := ctx.Value(contextConfigKey).(mapProvider)
	if !ok {
		return config
	}
	config, err := t.NewProvider(ctx, configx.WithValues(values))
	if err != nil {
		// This is not production code. The provider is only used in tests.
		panic(err)
	}
	return config
}

const contextConfigKey contextKey = 1

var (
	_ contextx.Contextualizer = (*TestConfigProvider)(nil)
)

func WithConfigValue(ctx context.Context, key string, value any) context.Context {
	return WithConfigValues(ctx, map[string]any{key: value})
}

func WithConfigValues(ctx context.Context, newValues map[string]any) context.Context {
	values, ok := ctx.Value(contextConfigKey).(mapProvider)
	if !ok {
		values = make(mapProvider)
	}
	expandedValues := make([]map[string]any, 0, len(newValues))
	for k, v := range newValues {
		parts := strings.Split(k, ".")
		val := map[string]any{parts[len(parts)-1]: v}
		if len(parts) > 1 {
			for i := len(parts) - 2; i >= 0; i-- {
				val = map[string]any{parts[i]: val}
			}
		}
		expandedValues = append(expandedValues, val)
	}
	for _, v := range expandedValues {
		maps.Merge(v, values)
	}

	return context.WithValue(ctx, contextConfigKey, values)
}
