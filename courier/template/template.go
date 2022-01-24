package template

import "github.com/ory/kratos/driver/config"

type (
	TemplateConfig interface {
		CourierTemplatesRoot() string
		CourierTemplatesVerificationInvalid() *config.CourierEmailTemplate
		CourierTemplatesVerificationValid() *config.CourierEmailTemplate
		CourierTemplatesRecoveryInvalid() *config.CourierEmailTemplate
		CourierTemplatesRecoveryValid() *config.CourierEmailTemplate
	}
)
