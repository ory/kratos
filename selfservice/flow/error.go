// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/swagger"
	"github.com/ory/x/urlx"

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

// Is sent when a flow is replaced by a different flow of the same class
//
// swagger:model errorFlowReplaced
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type errorFlowReplaced struct {
	Error swagger.GenericError `json:"error"`
	// The flow ID that should be used for the new flow as it contains the correct messages.
	FlowID uuid.UUID `json:"use_flow_id"`
}

// ReplacedError is sent when a flow is replaced by a different flow of the same class
type ReplacedError struct {
	*herodot.DefaultError `json:"error"`

	// The flow ID that should be used for the new flow as it contains the correct messages.
	FlowID uuid.UUID `json:"use_flow_id"`

	flow Flow

	// TODO: This error could be enhanced by providing a "flow class" (e.g. "Recovery", "Settings", "Verification", "Login", etc.)
}

func (e *ReplacedError) WithFlow(flow Flow) *ReplacedError {
	e.FlowID = flow.GetID()
	e.flow = flow
	return e
}

func (e *ReplacedError) GetFlow() Flow {
	return e.flow
}

func (e *ReplacedError) EnhanceJSONError() interface{} {
	return e
}

func NewFlowReplacedError(message *text.Message) *ReplacedError {
	return &ReplacedError{
		DefaultError: x.ErrGone.WithID(text.ErrIDSelfServiceFlowReplaced).
			WithError("self-service flow replaced").
			WithReasonf(message.Text),
	}
}

// Is sent when a flow is expired
//
// swagger:model selfServiceFlowExpiredError
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type selfServiceFlowExpiredError struct {
	Error swagger.GenericError `json:"error"`

	// When the flow has expired
	ExpiredAt time.Time `json:"expired_at"`

	// Please use the "expired_at" field instead to have a more accurate result.
	//
	// Deprecated: true
	Since time.Duration `json:"since"`

	// The flow ID that should be used for the new flow as it contains the correct messages.
	FlowID uuid.UUID `json:"use_flow_id"`
}

// ExpiredError is sent when a flow is expired
type ExpiredError struct {
	*herodot.DefaultError `json:"error"`

	// When the flow has expired
	ExpiredAt time.Time `json:"expired_at"`

	// DEPRECATED: Please use the "expired_at" field instead to have a more accurate result.
	Since time.Duration `json:"since"`

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

func NewFlowExpiredError(at time.Time) *ExpiredError {
	ago := time.Since(at)
	return &ExpiredError{
		ExpiredAt: at.UTC(),
		Since:     ago,
		DefaultError: x.ErrGone.WithID(text.ErrIDSelfServiceFlowExpired).
			WithError("self-service flow expired").
			WithReasonf("The self-service flow expired %.2f minutes ago, initialize a new one.", ago.Minutes()),
	}
}

// Is sent when a flow requires a browser to change its location.
//
// swagger:model errorBrowserLocationChangeRequired
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type errorBrowserLocationChangeRequired struct {
	Error swagger.ErrorGeneric `json:"error"`

	// Points to where to redirect the user to next.
	RedirectBrowserTo string `json:"redirect_browser_to"`
}

// BrowserLocationChangeRequiredError is sent when a flow requires a browser to change its location.
type BrowserLocationChangeRequiredError struct {
	*herodot.DefaultError `json:"error"`

	// Points to where to redirect the user to next.
	RedirectBrowserTo string `json:"redirect_browser_to"`
}

func (e *BrowserLocationChangeRequiredError) EnhanceJSONError() interface{} {
	return e
}

func NewBrowserLocationChangeRequiredError(redirectTo string) *BrowserLocationChangeRequiredError {
	return &BrowserLocationChangeRequiredError{
		RedirectBrowserTo: redirectTo,
		DefaultError: &herodot.DefaultError{
			IDField:     text.ErrIDSelfServiceBrowserLocationChangeRequiredError,
			CodeField:   http.StatusUnprocessableEntity,
			StatusField: http.StatusText(http.StatusUnprocessableEntity),
			ReasonField: fmt.Sprintf("In order to complete this flow please redirect the browser to: %s", redirectTo),
			DebugField:  "",
			ErrorField:  "browser location change required",
		},
	}
}

func HandleHookError(_ http.ResponseWriter, r *http.Request, f Flow, traits identity.Traits, group node.UiNodeGroup, flowError error, logger x.LoggingProvider, csrf x.CSRFTokenGeneratorProvider) error {
	if f != nil {
		if traits != nil {
			cont, err := container.NewFromStruct("", group, traits, "traits")
			if err != nil {
				logger.Logger().WithError(err).Error("could not update flow UI")
				return err
			}

			for _, n := range cont.Nodes {
				// we only set the value and not the whole field because we want to keep types from the initial form generation
				f.GetUI().Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
			}
		}

		if f.GetType() == TypeBrowser {
			f.GetUI().SetCSRF(csrf.GenerateCSRFToken(r))
		}
	}

	return flowError
}

func GetFlowExpiredRedirectURL(ctx context.Context, config *config.Config, route, returnTo string) *url.URL {
	redirectURL := urlx.AppendPaths(config.SelfPublicURL(ctx), route)
	if returnTo != "" {
		redirectURL = urlx.CopyWithQuery(redirectURL, url.Values{"return_to": {returnTo}})
	}

	return redirectURL
}
