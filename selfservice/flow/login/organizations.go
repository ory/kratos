package login

import (
	"github.com/gofrs/uuid"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

var organizationFilter = []StrategyFilter{func(s Strategy) bool {
	_, ok := s.(interface {
		SupportsOrganizations() bool
	})
	return ok
}}

func CreateOrganizationsFilter(r *http.Request, f *Flow, sess *session.Session) []StrategyFilter {
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
