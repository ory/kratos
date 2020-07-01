package testhelpers

import (
	"fmt"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
)

func StrategyEnable(strategy string, enable bool) {
	viper.Set(fmt.Sprintf("%s.%s.enabled", configuration.ViperKeySelfServiceStrategyConfig, strategy), enable)
}

func RecoveryFlowEnable(enable bool) {
	viper.Set(configuration.ViperKeySelfServiceRecoveryEnabled, enable)
}

func VerificationFlowEnable(enable bool) {
	viper.Set(configuration.ViperKeySelfServiceVerificationEnabled, enable)
}
