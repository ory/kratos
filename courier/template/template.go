// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
		CourierTemplatesLoginValid() *config.CourierEmailTemplate
		CourierTemplatesRegistrationValid() *config.CourierEmailTemplate
	}

	Dependencies interface {
		CourierConfig() config.CourierConfigs
		HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client
	}
)
