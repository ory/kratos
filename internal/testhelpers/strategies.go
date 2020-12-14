package testhelpers

import (
	"fmt"
	"testing"

	"github.com/ory/kratos/driver/config"
)

func StrategyEnable(t *testing.T, c *config.Provider, strategy string, enable bool) {
	c.MustSet(fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, strategy), enable)
}
