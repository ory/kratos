// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/ory/x/configx"

	"github.com/ory/kratos/driver/config"
)

// Deprecated: use EnableStrategy instead.
func StrategyEnable(t *testing.T, c *config.Config, strategy string, enable bool) {
	ctx := context.Background()
	c.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, strategy), enable)
}

func EnableStrategy(strategy string, enable bool) configx.OptionModifier {
	return configx.WithValue(fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, strategy), enable)
}
