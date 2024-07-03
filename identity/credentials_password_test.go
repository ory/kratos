// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialsPassword_ShouldUsePasswordMigrationHook(t *testing.T) {
	tests := []struct {
		name string
		cp   *CredentialsPassword
		want bool
	}{{
		name: "pw set",
		cp: &CredentialsPassword{
			HashedPassword:           "pw",
			UsePasswordMigrationHook: true,
		},
		want: false,
	}, {
		name: "pw not set",
		cp: &CredentialsPassword{
			HashedPassword:           "",
			UsePasswordMigrationHook: true,
		},
		want: true,
	}, {
		name: "nil",
		want: false,
	}, {
		name: "pw not set, hook not set",
		cp: &CredentialsPassword{
			HashedPassword: "",
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.cp.ShouldUsePasswordMigrationHook(), "ShouldUsePasswordMigrationHook()")
		})
	}
}
