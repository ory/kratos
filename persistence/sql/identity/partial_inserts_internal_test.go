// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/sqlxx"
)

func TestIdentitiesNeedPartialInserts(t *testing.T) {
	withExternalID := func(id string) *identity.Identity {
		return &identity.Identity{ExternalID: sqlxx.NullString(id)}
	}

	for _, tc := range []struct {
		name       string
		identities []*identity.Identity
		want       bool
	}{
		{
			name:       "empty batch",
			identities: nil,
			want:       false,
		},
		{
			name:       "single identity without external id",
			identities: []*identity.Identity{{}},
			want:       false,
		},
		{
			name:       "single identity with external id",
			identities: []*identity.Identity{withExternalID("ext-1")},
			want:       false,
		},
		{
			// The identities insert cannot conflict without an external id (the
			// only unique constraint), so partial inserts are unnecessary.
			name:       "many identities without external ids",
			identities: []*identity.Identity{{}, {}, {}},
			want:       false,
		},
		{
			name:       "many identities, one with external id",
			identities: []*identity.Identity{{}, withExternalID("ext-1"), {}},
			want:       true,
		},
		{
			name:       "many identities all with external ids",
			identities: []*identity.Identity{withExternalID("ext-1"), withExternalID("ext-2")},
			want:       true,
		},
	} {
		t.Run("case="+tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, identitiesNeedPartialInserts(tc.identities))
		})
	}
}
