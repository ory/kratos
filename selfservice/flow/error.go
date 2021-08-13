package flow

import (
	"github.com/pkg/errors"
)

var (
	ErrStrategyNotResponsible   = errors.New("strategy is not responsible for this request")
	ErrCompletedByStrategy      = errors.New("flow response completed by strategy")
	ErrStrategyAsksToReturnToUI = errors.New("flow strategy is redirecting to the ui")
)
