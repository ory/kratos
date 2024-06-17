// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"

	"github.com/ory/kratos/embedx"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
)

type (
	TestConfigProvider struct {
		contextx.Contextualizer
		Options []configx.OptionModifier
	}
	contextKey int
)

func (t *TestConfigProvider) NewProvider(ctx context.Context, opts ...configx.OptionModifier) (*configx.Provider, error) {
	return configx.New(ctx, []byte(embedx.ConfigSchema), append(t.Options, opts...)...)
}

func (t *TestConfigProvider) Config(ctx context.Context, config *configx.Provider) *configx.Provider {
	config = t.Contextualizer.Config(ctx, config)
	values, ok := ctx.Value(contextConfigKey).([]map[string]any)
	if !ok {
		return config
	}
	opts := make([]configx.OptionModifier, 0, len(values))
	for _, v := range values {
		opts = append(opts, configx.WithValues(v))
	}
	config, err := t.NewProvider(ctx, opts...)
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

func WithConfigValues(ctx context.Context, setValues map[string]any) context.Context {
	values, ok := ctx.Value(contextConfigKey).([]map[string]any)
	if !ok {
		values = make([]map[string]any, 0)
	}
	newValues := make([]map[string]any, len(values), len(values)+1)
	copy(newValues, values)
	newValues = append(newValues, setValues)

	return context.WithValue(ctx, contextConfigKey, newValues)
}
