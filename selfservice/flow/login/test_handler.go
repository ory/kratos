// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
)

const (
	// RouteAdminCreateTestFlow is the admin route for creating a dry-run OIDC
	// test login flow. The admin router prepends /admin automatically, so the
	// full path is POST /admin/test-login-flows.
	RouteAdminCreateTestFlow = "/test-login-flows"

	// RouteDeleteTestFlow is the public route for deleting a test login flow.
	// It is authorized by the flow's CSRF bearer token, not by a session.
	RouteDeleteTestFlow = "/self-service/login/test"

	// maxProviderIDLength bounds the provider_id payload field. Provider IDs
	// are short identifiers; anything longer indicates user error or abuse.
	maxProviderIDLength = 255
)

// TestStrategy populates a login flow's UI with a single submit button
// for the given OIDC provider. The OIDC strategy implements this interface.
// Defined in the login package to avoid an import cycle.
type TestStrategy interface {
	PopulateTestLoginFlow(r *http.Request, f *Flow, providerID string) error
}

// TestStrategyProvider exposes the enabled TestStrategy for the
// current request context.
type TestStrategyProvider interface {
	TestStrategy(ctx context.Context) TestStrategy
}

// Parameters for creating a test OIDC login flow.
//
// swagger:parameters createTestLoginFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createTestLoginFlow struct {
	// in: body
	// required: true
	Body createTestLoginFlowBody
}

// Request body for creating a test OIDC login flow.
//
// swagger:model createTestLoginFlowBody
type createTestLoginFlowBody struct {
	// ID of the OIDC provider to test. Must match a provider configured on
	// the project that serves this request.
	//
	// required: true
	// maxLength: 255
	ProviderID string `json:"provider_id"`
}

// swagger:route POST /admin/test-login-flows identity createTestLoginFlow
//
// # Create a test OIDC login flow
//
// Creates a dry-run OIDC test login flow pre-scoped to one provider. The
// returned flow carries a single-submit UI and a CSRF bearer token. No
// identity is persisted and no session is issued when the flow completes;
// the captured debug data is returned in the flow's test_context.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  201: loginFlow
//	  400: errorGeneric
//	  404: errorGeneric
//	  default: errorGeneric
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-admin-low
func (h *Handler) adminCreateTestLoginFlow(w http.ResponseWriter, r *http.Request) {
	// The strategy annotates this span with the provider lookup outcome
	// (test_login_flow.* attributes); without it, a failed lookup is
	// invisible in traces because the whole path is in-memory.
	ctx, span := h.d.Tracer(r.Context()).Tracer().Start(r.Context(), "login.Handler.adminCreateTestLoginFlow")
	defer span.End()
	r = r.WithContext(ctx)

	var body createTestLoginFlowBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest().WithReason("could not parse request body")))
		return
	}
	if body.ProviderID == "" {
		h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest().WithReason("provider_id is required")))
		return
	}
	if len(body.ProviderID) > maxProviderIDLength {
		h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest().WithReasonf("provider_id must be at most %d characters", maxProviderIDLength)))
		return
	}

	f, err := NewFlow(h.d, r, flow.TypeBrowser)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
	// Admin-created flows have no browser cookie to bind against; mint a
	// random UUID as the opaque CSRF bearer and validate it on submit.
	f.CSRFToken = x.NewUUID().String()

	if err := f.SetTestContext(&TestContext{ProviderID: body.ProviderID}); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
	// Mark as "refresh" so the OIDC strategy's alreadyAuthenticated guard does
	// not short-circuit the submit if the admin's browser happens to carry an
	// unrelated Ory session cookie. Test flows always run the OIDC round-trip
	// regardless of session state.
	f.Refresh = true

	strategy := h.d.TestStrategy(r.Context())
	if strategy == nil {
		h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError().WithReason("OIDC strategy is not registered; test login flows cannot be created")))
		return
	}
	if err := strategy.PopulateTestLoginFlow(r, f, body.ProviderID); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if err := h.d.LoginFlowPersister().CreateLoginFlow(r.Context(), f); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().WriteCode(w, r, http.StatusCreated, f)
}

// Delete Test Login Flow Parameters
//
// swagger:parameters deleteTestLoginFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type deleteTestLoginFlow struct {
	// ID of the test login flow to delete.
	//
	// required: true
	// in: query
	ID string `json:"id"`

	// HTTP Cookies. A captured test flow requires the ory_kratos_test_flow
	// cookie set by the OIDC callback; a flow still in the initial
	// choose-method state does not.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// swagger:route DELETE /self-service/login/test frontend deleteTestLoginFlow
//
// # Delete a test OIDC login flow
//
// Deletes a dry-run OIDC test login flow. A flow whose debug payload has
// been captured requires the HMAC cookie set by the OIDC callback; a flow
// still in the initial choose-method state is deletable with just the flow
// ID (it carries no PII, and the admin may want to abandon it).
//
//	Schemes: http, https
//
//	Responses:
//	  204: emptyResponse
//	  400: errorGeneric
//	  403: errorGeneric
//	  404: errorGeneric
//	  default: errorGeneric
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-public-high
func (h *Handler) deleteTestLoginFlow(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.FromString(r.URL.Query().Get("id"))
	if err != nil {
		h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest().WithReason("invalid flow id")))
		return
	}

	// Load the flow and discriminate by whether a debug payload has been
	// captured. A return of sqlcon.ErrNoRows maps to 404 — we don't leak
	// whether the ID belongs to a non-test login flow vs. a missing row.
	f, err := h.d.LoginFlowPersister().GetLoginFlow(r.Context(), id)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
	if !f.IsTest() {
		h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrNotFound()))
		return
	}
	if err := f.LoadTestContext(); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
	if f.TestContext != nil && f.TestContext.DebugPayload != nil &&
		!nosurf.VerifyToken(h.d.GenerateCSRFToken(r), f.CSRFToken) {
		h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrForbidden().WithReason("missing or invalid CSRF token")))
		return
	}

	if err := h.d.LoginFlowPersister().DeleteTestLoginFlow(r.Context(), id); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
