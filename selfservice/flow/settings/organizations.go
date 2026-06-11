// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
)

var organizationFilter = []StrategyFilter{func(s Strategy) bool {
	a, b := s.(flow.OrganizationImplementor)
	return b && a.SupportsOrganizations()
}}

// PrepareOrganizations resolves the org id from the identity (winner) or the
// `?organization=` query param, sets it on the flow, and returns a filter
// that keeps only org-implementing strategies. Returns an empty filter when
// no org id applies — the flow then proceeds as a normal (non-org) settings
// flow.
//
// The identity is the authoritative source: an identity already bound to an
// org cannot be silently re-bound, and the session cache must not be relied
// on because it can lag behind a just-persisted org binding (PostSettingsHook
// rebuilds the flow with the freshly linked identity).
//
// The query param is free-form caller input, so it is only honored when one
// of the identity's email addresses is under the organization's configured
// domains. Without that gate, any authenticated identity could render an
// org-scoped flow for every org in the project and enumerate its SSO
// providers. The org hooks block self-service registration and profile
// updates under a claimed domain, so a matching address implies the domain
// owner provisioned the account.
//
// When the resolved org id does not match any configured organization, the
// flow falls back to the unscoped path. An identity may carry an
// organization_id for an organization that has since been deleted; scoping
// to a dangling org would render no nodes and trap chained flows such as
// recovery -> settings.
func PrepareOrganizations(r *http.Request, f *Flow, i *identity.Identity, configuredOrgs []config.Organization) []StrategyFilter {
	if f.OrganizationID.Valid {
		return organizationFilter
	}

	if i != nil && i.OrganizationID.Valid {
		if _, ok := findConfiguredOrganization(i.OrganizationID.UUID, configuredOrgs); !ok {
			return []StrategyFilter{}
		}
		f.OrganizationID = i.OrganizationID
		return organizationFilter
	}

	orgID := flow.ParseOrganizationFromURLQuery(r.Context(), r.URL.Query())
	if !orgID.Valid {
		return []StrategyFilter{}
	}

	org, ok := findConfiguredOrganization(orgID.UUID, configuredOrgs)
	if !ok {
		return []StrategyFilter{}
	}

	if !identityEmailMatchesOrganizationDomains(i, org) {
		return []StrategyFilter{}
	}

	f.OrganizationID = orgID
	return organizationFilter
}

func findConfiguredOrganization(id uuid.UUID, configured []config.Organization) (config.Organization, bool) {
	idx := slices.IndexFunc(configured, func(o config.Organization) bool {
		return o.ID == id
	})
	if idx < 0 {
		return config.Organization{}, false
	}
	return configured[idx], true
}

// identityEmailMatchesOrganizationDomains reports whether one of the
// identity's email addresses is under one of the organization's domains. The
// normalization mirrors the org email-domain matching in the cloud module's
// organization hook.
func identityEmailMatchesOrganizationDomains(i *identity.Identity, org config.Organization) bool {
	if i == nil {
		return false
	}
	// TODO(jonas): This should actually use the `organization.via` field of the identity schema.
	for _, addr := range i.VerifiableAddresses {
		if addr.Via != identity.AddressTypeEmail {
			continue
		}
		_, domain, ok := strings.Cut(x.NormalizeEmailIdentifier(addr.Value), "@")
		if !ok || domain == "" {
			continue
		}
		if slices.ContainsFunc(org.Domains, func(orgDomain string) bool {
			return strings.EqualFold(domain, x.NormalizeOtherIdentifier(orgDomain))
		}) {
			return true
		}
	}
	return false
}
