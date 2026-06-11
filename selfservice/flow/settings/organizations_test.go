// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/settings"
)

func TestPrepareOrganizations(t *testing.T) {
	ctx := context.Background()
	orgQuery := uuid.Must(uuid.NewV4())
	orgIdentity := uuid.Must(uuid.NewV4())
	configured := []config.Organization{
		{ID: orgQuery, Domains: []string{"corp.example"}},
		{ID: orgIdentity, Domains: []string{"other-corp.example"}},
	}

	// identityWithEmail returns an unbound identity carrying a single email
	// verifiable address, mirroring what the session loader hydrates via
	// identity.ExpandDefault.
	identityWithEmail := func(email string) *identity.Identity {
		return &identity.Identity{
			VerifiableAddresses: []identity.VerifiableAddress{
				{Value: email, Via: identity.AddressTypeEmail},
			},
		}
	}

	t.Run("returns empty filter when no org id is set", func(t *testing.T) {
		f := &settings.Flow{}
		r := httptest.NewRequest("GET", "/self-service/settings/browser", nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, identityWithEmail("user@corp.example"), configured)
		assert.Empty(t, filters)
		assert.False(t, f.OrganizationID.Valid)
	})

	t.Run("uses query param when the identity's email is under the org's domains", func(t *testing.T) {
		f := &settings.Flow{}
		r := httptest.NewRequest("GET", "/self-service/settings/browser?organization="+orgQuery.String(), nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, identityWithEmail("user@corp.example"), configured)
		assert.Len(t, filters, 1)
		assert.Equal(t, orgQuery, f.OrganizationID.UUID)
	})

	t.Run("domain match is case-insensitive", func(t *testing.T) {
		f := &settings.Flow{}
		r := httptest.NewRequest("GET", "/self-service/settings/browser?organization="+orgQuery.String(), nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, identityWithEmail("User@CORP.Example"), configured)
		assert.Len(t, filters, 1)
		assert.Equal(t, orgQuery, f.OrganizationID.UUID)
	})

	t.Run("subdomain of a configured domain is not a match", func(t *testing.T) {
		// Domain comparison is an exact match, not a suffix match: an email
		// under sub.corp.example must not satisfy an org that claims
		// corp.example. Suffix matching would let anyone controlling an
		// arbitrary subdomain-style address scope flows to the org.
		f := &settings.Flow{}
		r := httptest.NewRequest("GET", "/self-service/settings/browser?organization="+orgQuery.String(), nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, identityWithEmail("user@sub.corp.example"), configured)
		assert.Empty(t, filters)
		assert.False(t, f.OrganizationID.Valid)
	})

	t.Run("query param with a non-matching email domain falls back to unscoped", func(t *testing.T) {
		// The query parameter is free-form caller input. Without a domain
		// match it must not scope the flow — otherwise any authenticated
		// identity could enumerate the SSO providers of every org in the
		// project.
		f := &settings.Flow{}
		r := httptest.NewRequest("GET", "/self-service/settings/browser?organization="+orgQuery.String(), nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, identityWithEmail("user@elsewhere.example"), configured)
		assert.Empty(t, filters)
		assert.False(t, f.OrganizationID.Valid)
	})

	t.Run("query param with an identity without email addresses falls back to unscoped", func(t *testing.T) {
		f := &settings.Flow{}
		r := httptest.NewRequest("GET", "/self-service/settings/browser?organization="+orgQuery.String(), nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, &identity.Identity{}, configured)
		assert.Empty(t, filters)
		assert.False(t, f.OrganizationID.Valid)
	})

	t.Run("non-email addresses do not count as a domain match", func(t *testing.T) {
		f := &settings.Flow{}
		i := &identity.Identity{
			VerifiableAddresses: []identity.VerifiableAddress{
				{Value: "user@corp.example", Via: identity.AddressTypeSMS},
			},
		}
		r := httptest.NewRequest("GET", "/self-service/settings/browser?organization="+orgQuery.String(), nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, i, configured)
		assert.Empty(t, filters)
		assert.False(t, f.OrganizationID.Valid)
	})

	t.Run("nil identity cannot be scoped via the query param", func(t *testing.T) {
		f := &settings.Flow{}
		r := httptest.NewRequest("GET", "/self-service/settings/browser?organization="+orgQuery.String(), nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, nil, configured)
		assert.Empty(t, filters)
		assert.False(t, f.OrganizationID.Valid)
	})

	t.Run("identity org id overrides query param without a domain check", func(t *testing.T) {
		// An identity already bound to an org cannot be silently re-bound to
		// the org carried by the query string. Membership is proven by the
		// binding itself, so no email-domain match is required.
		f := &settings.Flow{}
		r := httptest.NewRequest("GET", "/self-service/settings/browser?organization="+orgQuery.String(), nil).WithContext(ctx)
		i := &identity.Identity{OrganizationID: uuid.NullUUID{UUID: orgIdentity, Valid: true}}
		filters := settings.PrepareOrganizations(r, f, i, configured)
		assert.Len(t, filters, 1)
		assert.Equal(t, orgIdentity, f.OrganizationID.UUID)
	})

	t.Run("preserves an already-set flow org id", func(t *testing.T) {
		f := &settings.Flow{OrganizationID: uuid.NullUUID{UUID: orgQuery, Valid: true}}
		r := httptest.NewRequest("GET", "/self-service/settings/browser", nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, &identity.Identity{}, configured)
		assert.Len(t, filters, 1)
		assert.Equal(t, orgQuery, f.OrganizationID.UUID)
	})

	t.Run("identity org id pointing at a deleted org falls back to unscoped", func(t *testing.T) {
		// An identity bound to an organization that has since been deleted
		// must not trap the user in an org-scoped flow with no nodes —
		// notably the recovery -> settings hand-off relies on the password
		// node being rendered for offboarded users.
		f := &settings.Flow{}
		r := httptest.NewRequest("GET", "/self-service/settings/browser", nil).WithContext(ctx)
		i := &identity.Identity{OrganizationID: uuid.NullUUID{UUID: uuid.Must(uuid.NewV4()), Valid: true}}
		filters := settings.PrepareOrganizations(r, f, i, configured)
		assert.Empty(t, filters)
		assert.False(t, f.OrganizationID.Valid)
	})

	t.Run("query param org id that is not configured falls back to unscoped", func(t *testing.T) {
		f := &settings.Flow{}
		unknown := uuid.Must(uuid.NewV4())
		r := httptest.NewRequest("GET", "/self-service/settings/browser?organization="+unknown.String(), nil).WithContext(ctx)
		filters := settings.PrepareOrganizations(r, f, identityWithEmail("user@corp.example"), configured)
		assert.Empty(t, filters)
		assert.False(t, f.OrganizationID.Valid)
	})
}
