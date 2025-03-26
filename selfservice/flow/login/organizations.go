// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"net/http"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
)

var organizationFilter = []StrategyFilter{func(s Strategy) bool {
	a, b := s.(flow.OrganizationImplementor)
	return b && a.SupportsOrganizations()
}}

func PrepareOrganizations(r *http.Request, f *Flow, sess *session.Session) []StrategyFilter {
	if f.RequestedAAL != identity.AuthenticatorAssuranceLevel1 {
		return []StrategyFilter{}
	}

	if f.OrganizationID.Valid {
		return organizationFilter
	}

	orgID := flow.ParseOrganizationFromURLQuery(r.Context(), r.URL.Query())
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
