// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

type Dependencies interface {
	CourierConfig() config.CourierConfigs
	x.HTTPClientProvider
}
