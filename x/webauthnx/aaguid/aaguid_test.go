// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package aaguid

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	goodUUIDBytes := uuid.Must(uuid.FromString("adce0002-35bc-c60a-648b-0b25f1f05503")).Bytes()
	badUUIDBytes := uuid.Must(uuid.NewV4()).Bytes()

	tests := []struct {
		name string
		id   []byte
		want string
	}{
		{
			name: "GoodUUID",
			id:   goodUUIDBytes,
			want: "Chrome on Mac",
		},
		{
			name: "BadUUID",
			id:   badUUIDBytes,
		},
		{
			name: "NilUUID",
			id:   nil,
		},
		{
			name: "EmptyUUID",
			id:   []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Lookup(tt.id)
			if tt.want == "" {
				assert.Nil(t, res)
			} else {
				assert.Equal(t, tt.want, res.Name)
			}
		})
	}
}
