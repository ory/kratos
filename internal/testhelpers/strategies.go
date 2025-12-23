// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/ory/kratos/driver/config"
)

// Deprecated: use MethodEnableConfig instead.
func StrategyEnable(t *testing.T, c *config.Config, strategy string, enable bool) {
	ctx := context.Background()
	c.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, strategy), enable)
}

func MethodEnableConfig[S ~string](method S, enable bool) map[string]any {
	return map[string]any{
		fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, method): enable,
	}
}
