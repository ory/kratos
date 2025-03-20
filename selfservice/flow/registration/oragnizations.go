package registration

import (
	"github.com/gofrs/uuid"
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

func CreateOrganizationsFilter(r *http.Request, f *Flow) []StrategyFilter {
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
