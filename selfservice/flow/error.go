package flow

import (
	"fmt"
	"net/http"
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
	DefaultError *herodot.DefaultError `json:"error"`

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

func (e *ExpiredError) EnhanceJSONError() interface{} {
	return e
}

func (e *ExpiredError) Error() string {
	return e.DefaultError.Error()
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

// Is sent when a flow requires a browser to change its location.
//
// swagger:model selfServiceBrowserLocationChangeRequiredError
type BrowserLocationChangeRequiredError struct {
	DefaultError *herodot.DefaultError `json:"error"`

	// Since when the flow has expired
	RedirectBrowserTo string `json:"redirect_browser_to"`
}

func (e *BrowserLocationChangeRequiredError) EnhanceJSONError() interface{} {
	return e
}

func NewBrowserLocationChangeRequiredError(redirectTo string) *BrowserLocationChangeRequiredError {
	return &BrowserLocationChangeRequiredError{
		RedirectBrowserTo: redirectTo,
		DefaultError: &herodot.DefaultError{
			IDField: text.ErrIDSelfServiceBrowserLocationChangeRequiredError,
			CodeField:    http.StatusUnprocessableEntity,
			StatusField:  http.StatusText(http.StatusUnprocessableEntity),
			ReasonField:  fmt.Sprintf("In order to complete this flow please redirect the browser to: %s", redirectTo),
			DebugField:   "",
			ErrorField:   "browser location change required",
		},
	}
}

func (e *BrowserLocationChangeRequiredError) Error() string {
	return e.DefaultError.Error()
}
