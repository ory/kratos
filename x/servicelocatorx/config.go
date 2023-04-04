// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package servicelocatorx

import (
	"context"

	"github.com/ory/kratos/driver/config"
)

type key int

const (
	keyConfig key = iota + 1
)

// ContextWithConfig returns a new context with the provided config.
func ContextWithConfig(ctx context.Context, c *config.Config) context.Context {
	return context.WithValue(ctx, keyConfig, c)
}

// ConfigFromContext returns the config from the context.
func ConfigFromContext(ctx context.Context, fallback *config.Config) *config.Config {
	if c, ok := ctx.Value(keyConfig).(*config.Config); ok {
		return c
	}
	return fallback
}
