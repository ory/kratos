// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"github.com/ory/kratos/selfservice/strategy/code"
)

var CodeRegex = `(?m)(?:^|code is: |Use code |#| )(` + code.CodePattern + `)$`
