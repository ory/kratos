package flow

import (
	"github.com/pkg/errors"
)

var (
	ErrStrategyNotResponsible = errors.New("strategy is not responsible for this request")
)
