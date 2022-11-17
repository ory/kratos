// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"fmt"

	"github.com/ory/kratos/selfservice/strategy/code"
)

var CodeRegex = fmt.Sprintf(`(\d{%d})`, code.CodeLength)
