// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"net/http"

	"github.com/gofrs/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var organizationFilter = []StrategyFilter{func(s Strategy) bool {
	a, b := s.(interface {
		SupportsOrganizations() bool
	})
	return b && a.SupportsOrganizations()
}}

func PrepareOrganizations(r *http.Request, f *Flow) []StrategyFilter {
	if f.OrganizationID.Valid {
		return organizationFilter
	}

	if rawOrg := r.URL.Query().Get("organization"); rawOrg != "" {
		orgID, err := uuid.FromString(rawOrg)
		if err != nil {
			trace.SpanFromContext(r.Context()).RecordError(err, trace.WithAttributes(attribute.String("organization", rawOrg)))
		} else {
			f.OrganizationID = uuid.NullUUID{UUID: orgID, Valid: true}
			return organizationFilter
		}
	}

	return []StrategyFilter{}
}
