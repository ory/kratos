// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"context"
	"net/url"

	"github.com/gofrs/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ParseOrganizationFromURLQuery is a helper function that parses the organization ID for a self-service flow from
// the URL query parameters. If the organization ID is not found in the URL query parameters, the function will return
// an NULL UUID.
func ParseOrganizationFromURLQuery(ctx context.Context, q url.Values) (orgID uuid.NullUUID) {
	if rawOrg := q.Get("organization"); rawOrg != "" {
		orgIDFromURL, err := uuid.FromString(rawOrg)
		if err != nil {
			trace.SpanFromContext(ctx).RecordError(err, trace.WithAttributes(attribute.String("organization", rawOrg)))
		} else {
			orgID = uuid.NullUUID{UUID: orgIDFromURL, Valid: true}
		}
	}
	return
}

type OrganizationImplementor interface {
	SupportsOrganizations() bool
}
