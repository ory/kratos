package template

import (
	"context"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/ory/kratos/driver/config"
)

type (
	TemplateConfig interface {
		CourierTemplatesRoot() string
		CourierTemplatesVerificationInvalid() *config.CourierEmailTemplate
		CourierTemplatesVerificationValid() *config.CourierEmailTemplate
		CourierTemplatesRecoveryInvalid() *config.CourierEmailTemplate
		CourierTemplatesRecoveryValid() *config.CourierEmailTemplate
	}
	TemplateDependencies interface {
		CourierConfig(ctx context.Context) config.CourierConfigs
		HTTPClient(ctx context.Context) *retryablehttp.Client
	}
)
