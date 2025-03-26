// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"net/http"

	"github.com/ory/kratos/selfservice/flow"
)

var organizationFilter = []StrategyFilter{func(s Strategy) bool {
	a, b := s.(flow.OrganizationImplementor)
	return b && a.SupportsOrganizations()
}}

func PrepareOrganizations(r *http.Request, f *Flow) []StrategyFilter {
	if f.OrganizationID.Valid {
		return organizationFilter
	}

	orgID := flow.ParseOrganizationFromURLQuery(r.Context(), r.URL.Query())
	if !orgID.Valid {
		return []StrategyFilter{}
	}

	f.OrganizationID = orgID
	return organizationFilter
}
