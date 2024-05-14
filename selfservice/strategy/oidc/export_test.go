// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"

	oidcv1 "github.com/ory/kratos/gen/oidc/v1"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/oidc/claims"
)

// ValidateFlowForTest exposes the unexported flow resolver so tests in the
// external oidc_test package can assert that OIDC callback state dispatches
// to the correct flow kind (login, registration, settings).
func (s *Strategy) ValidateFlowForTest(ctx context.Context, r *http.Request, rid uuid.UUID, kind oidcv1.FlowKind) (flow.Flow, error) {
	return s.validateFlow(ctx, r, rid, kind)
}

// ForwardErrorForTest exposes forwardError so tests can assert that
// pre-dispatch callback failures finish as a ready flow with a populated
// DebugPayload instead of panicking.
func (s *Strategy) ForwardErrorForTest(ctx context.Context, w http.ResponseWriter, r *http.Request, f flow.Flow, err error) {
	s.forwardError(ctx, w, r, f, err)
}

// ProcessTestLoginForTest exposes processTestLogin so external tests can
// exercise the dry-run OIDC callback path without routing through the full
// callback HTTP pipeline.
func (s *Strategy) ProcessTestLoginForTest(ctx context.Context, w http.ResponseWriter, r *http.Request, f *login.Flow, claims *claims.Claims, provider Provider) error {
	return s.processTestLogin(ctx, w, r, f, claims, provider)
}

// FinishTestLoginForTest exposes finishTestLogin so external tests can
// assert the idempotency guard against an already-captured DebugPayload.
func (s *Strategy) FinishTestLoginForTest(ctx context.Context, w http.ResponseWriter, r *http.Request, f *login.Flow, dp *login.DebugPayload) error {
	return s.finishTestLogin(ctx, w, r, f, dp)
}
