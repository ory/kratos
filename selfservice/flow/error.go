package flow

import (
	"time"

	"github.com/ory/kratos/x"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/text"
)

var (
	ErrStrategyNotResponsible   = errors.New("strategy is not responsible for this request")
	ErrCompletedByStrategy      = errors.New("flow response completed by strategy")
	ErrStrategyAsksToReturnToUI = errors.New("flow strategy is redirecting to the ui")
)

// Is sent when a flow is expired
//
// swagger:model selfServiceFlowExpiredError
type ExpiredError struct {
	*herodot.DefaultError

	// Since when the flow has expired
	Ago time.Duration `json:"since"`

	// The flow ID that should be used for the new flow as it contains the correct messages.
	FlowID uuid.UUID `json:"use_flow_id"`

	flow Flow
}

func (e *ExpiredError) WithFlow(flow Flow) *ExpiredError {
	e.FlowID = flow.GetID()
	e.flow = flow
	return e
}

func (e *ExpiredError) GetFlow() Flow {
	return e.flow
}

func NewFlowExpiredError(at time.Time) *ExpiredError {
	ago := time.Since(at)
	return &ExpiredError{
		Ago: ago,
		DefaultError: x.ErrGone.WithID(text.ErrIDSelfServiceFlowExpired).
			WithError("self-service flow expired").
			WithReasonf("The self-service flow expired %.2f minutes ago, initialize a new one.", ago.Minutes()),
	}
}
