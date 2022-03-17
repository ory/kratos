package template

import (
	"context"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/httpx"
)

type (
	Config interface {
		CourierTemplatesRoot() string
		CourierTemplatesVerificationInvalid() *config.CourierEmailTemplate
		CourierTemplatesVerificationValid() *config.CourierEmailTemplate
		CourierTemplatesRecoveryInvalid() *config.CourierEmailTemplate
		CourierTemplatesRecoveryValid() *config.CourierEmailTemplate
	}

	Dependencies interface {
		CourierConfig(ctx context.Context) config.CourierConfigs
		HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client
	}
)
