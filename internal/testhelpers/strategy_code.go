package testhelpers

import (
	"fmt"

	"github.com/ory/kratos/selfservice/strategy/code"
)

var CodeRegex = fmt.Sprintf(`(\d{%d})`, code.CodeLength)
