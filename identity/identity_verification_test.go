// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/x"
)

func TestNewVerifiableEmailAddress(t *testing.T) {
	iid := x.NewUUID()
	a := NewVerifiableEmailAddress("foo@ory.sh", iid)
	var nullTime *sqlxx.NullTime

	assert.Equal(t, a.Value, "foo@ory.sh")
	assert.Equal(t, a.Via, VerifiableAddressTypeEmail)
	assert.Equal(t, a.Status, VerifiableAddressStatusPending)
	assert.Equal(t, a.Verified, false)
	assert.EqualValues(t, nullTime, a.VerifiedAt)
	assert.Equal(t, iid, a.IdentityID)
	assert.Equal(t, uuid.Nil, a.ID)
}

var tagsIgnoredForHashing = map[string]struct{}{
	"id":          {},
	"created_at":  {},
	"updated_at":  {},
	"verified_at": {},
}

func reflectiveHash(record any) string {
	var (
		val    = reflect.ValueOf(record)
		typ    = reflect.TypeOf(record)
		values = []string{}
	)
	for i := 0; i < val.NumField(); i++ {
		dbTag, ok := typ.Field(i).Tag.Lookup("db")
		if !ok {
			continue
		}
		if _, ignore := tagsIgnoredForHashing[dbTag]; ignore {
			continue
		}
		if !val.Field(i).CanInterface() {
			continue
		}
		values = append(values, fmt.Sprintf("%v", val.Field(i).Interface()))
	}
	return strings.Join(values, "|")
}

// TestVerifiableAddress_Hash tests that the hash considers all fields that are
// written to the database (ignoring some well-known fields like the ID or
// timestamps).
func TestVerifiableAddress_Hash(t *testing.T) {
	now := sqlxx.NullTime(time.Now())
	cases := []struct {
		name string
		a    VerifiableAddress
	}{
		{
			name: "full fields",
			a: VerifiableAddress{
				ID:         x.NewUUID(),
				Value:      "foo@bar.me",
				Verified:   false,
				Via:        AddressTypeEmail,
				Status:     VerifiableAddressStatusPending,
				VerifiedAt: &now,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				IdentityID: x.NewUUID(),
				NID:        x.NewUUID(),
			},
		}, {
			name: "empty fields",
			a:    VerifiableAddress{},
		}, {
			name: "constructor",
			a:    *NewVerifiableEmailAddress("foo@ory.sh", x.NewUUID()),
		},
	}

	for _, tc := range cases {
		t.Run("case="+tc.name, func(t *testing.T) {
			assert.Equal(t,
				reflectiveHash(tc.a),
				tc.a.Hash(),
			)
		})
	}

}
