// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
)

func TestPrepareOrganizations(t *testing.T) {
	t.Run("should return empty filter if AAL is not AAL1", func(t *testing.T) {
		f := &Flow{RequestedAAL: identity.AuthenticatorAssuranceLevel2}
		r := &http.Request{URL: new(url.URL)}
		sess := &session.Session{}

		filters := PrepareOrganizations(r, f, sess)
		assert.Empty(t, filters)
	})

	t.Run("should return empty filter if OrganizationID is not set", func(t *testing.T) {
		f := &Flow{RequestedAAL: identity.AuthenticatorAssuranceLevel1}
		r := &http.Request{URL: new(url.URL)}
		sess := &session.Session{}

		filters := PrepareOrganizations(r, f, sess)
		assert.Empty(t, filters)
	})

	t.Run("should return organization filter if OrganizationID is valid", func(t *testing.T) {
		f := &Flow{RequestedAAL: identity.AuthenticatorAssuranceLevel1}
		r := &http.Request{URL: new(url.URL)}
		sess := &session.Session{
			Identity: &identity.Identity{OrganizationID: uuid.NullUUID{Valid: true}},
		}

		filters := PrepareOrganizations(r, f, sess)
		require.NotEmpty(t, filters)
		assert.Equal(t, organizationFilter, filters)
	})

	t.Run("should parse OrganizationID from URL query if not in session", func(t *testing.T) {
		f := &Flow{RequestedAAL: identity.AuthenticatorAssuranceLevel1}
		r := &http.Request{
			URL: &url.URL{
				RawQuery: "organization=123e4567-e89b-12d3-a456-426614174000",
			},
		}
		sess := &session.Session{}
		filters := PrepareOrganizations(r, f, sess)
		require.NotEmpty(t, filters)
		assert.Equal(t, organizationFilter, filters)
	})

	t.Run("should use organization ID already in flow", func(t *testing.T) {
		orgID := uuid.NullUUID{UUID: uuid.Must(uuid.NewV4()), Valid: true}
		f := &Flow{
			RequestedAAL:   identity.AuthenticatorAssuranceLevel1,
			OrganizationID: orgID,
		}
		r := &http.Request{URL: new(url.URL)}
		sess := &session.Session{}

		filters := PrepareOrganizations(r, f, sess)
		require.NotEmpty(t, filters)
		assert.Equal(t, organizationFilter, filters)
		assert.Equal(t, orgID, f.OrganizationID)
	})

	t.Run("should prioritize session org ID over query param when both present", func(t *testing.T) {
		sessionOrgID := uuid.Must(uuid.NewV4())
		queryOrgID := uuid.Must(uuid.NewV4())

		f := &Flow{RequestedAAL: identity.AuthenticatorAssuranceLevel1}
		r := &http.Request{
			URL: &url.URL{
				RawQuery: "organization=" + queryOrgID.String(),
			},
		}
		sess := &session.Session{
			Identity: &identity.Identity{
				OrganizationID: uuid.NullUUID{UUID: sessionOrgID, Valid: true},
			},
		}

		filters := PrepareOrganizations(r, f, sess)
		require.NotEmpty(t, filters)
		assert.Equal(t, organizationFilter, filters)
		assert.Equal(t, sessionOrgID, f.OrganizationID.UUID)
	})

	t.Run("should not return filter when organization is set but AAL2 requested", func(t *testing.T) {
		orgID := uuid.NullUUID{UUID: uuid.Must(uuid.NewV4()), Valid: true}
		f := &Flow{
			RequestedAAL:   identity.AuthenticatorAssuranceLevel2,
			OrganizationID: orgID,
		}
		r := &http.Request{URL: new(url.URL)}
		sess := &session.Session{}

		filters := PrepareOrganizations(r, f, sess)
		assert.Empty(t, filters)
		// Organization ID should remain set in the flow
		assert.Equal(t, orgID, f.OrganizationID)
	})

	t.Run("should use organization ID from query when session has no org ID", func(t *testing.T) {
		queryOrgID := uuid.Must(uuid.NewV4())

		f := &Flow{RequestedAAL: identity.AuthenticatorAssuranceLevel1}
		r := &http.Request{
			URL: &url.URL{
				RawQuery: "organization=" + queryOrgID.String(),
			},
		}
		sess := &session.Session{
			Identity: &identity.Identity{
				// No organization ID set
				OrganizationID: uuid.NullUUID{Valid: false},
			},
		}

		filters := PrepareOrganizations(r, f, sess)
		require.NotEmpty(t, filters)
		assert.Equal(t, organizationFilter, filters)
		assert.Equal(t, queryOrgID, f.OrganizationID.UUID)
	})
}
