// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"net/http"

	"github.com/gofrs/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
)

var organizationFilter = []StrategyFilter{func(s Strategy) bool {
	a, b := s.(interface {
		SupportsOrganizations() bool
	})
	return b && a.SupportsOrganizations()
}}

func PrepareOrganizations(r *http.Request, f *Flow, sess *session.Session) []StrategyFilter {
	if f.RequestedAAL != identity.AuthenticatorAssuranceLevel1 {
		return []StrategyFilter{}
	}

	if f.OrganizationID.Valid {
		return organizationFilter
	}

	var orgID uuid.NullUUID
	if rawOrg := r.URL.Query().Get("organization"); rawOrg != "" {
		orgIDFromURL, err := uuid.FromString(rawOrg)
		if err != nil {
			trace.SpanFromContext(r.Context()).RecordError(err, trace.WithAttributes(attribute.String("organization", rawOrg)))
		} else {
			orgID = uuid.NullUUID{UUID: orgIDFromURL, Valid: true}
		}
	}

	if sess != nil && sess.Identity != nil && sess.Identity.OrganizationID.Valid {
		orgID = sess.Identity.OrganizationID
	}

	if !orgID.Valid {
		return []StrategyFilter{}
	}

	f.OrganizationID = orgID
	// We only apply the filter on AAL1, because the OIDC strategy can only satsify
	// AAL1.
	return organizationFilter
}
