// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"strings"

	"github.com/knadh/koanf/maps"

	"github.com/knadh/koanf/v2"

	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
)

type (
	TestConfigProvider struct {
		contextx.Contextualizer
	}
	contextKey  int
	mapProvider map[string]any
)

func (t *TestConfigProvider) Config(ctx context.Context, config *configx.Provider) *configx.Provider {
	config = t.Contextualizer.Config(ctx, config)
	k := config.Copy()
	if values, ok := ctx.Value(contextConfigKey).(mapProvider); ok && values != nil {
		// our trusty provider never errors
		_ = k.Load(values, nil)
	}
	c := *config
	c.Koanf = k
	return &c
}

const contextConfigKey contextKey = 1

var (
	_ contextx.Contextualizer = (*TestConfigProvider)(nil)
	_ koanf.Provider          = (*mapProvider)(nil)
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

func (m mapProvider) ReadBytes() ([]byte, error) {
	return nil, nil
}

func (m mapProvider) Read() (map[string]any, error) {
	return m, nil
}
