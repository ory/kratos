// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

// CredentialsPassword is contains the configuration for credentials of the type password.
//
// swagger:model identityCredentialsPassword
type CredentialsPassword struct {
	// HashedPassword is a hash-representation of the password.
	HashedPassword string `json:"hashed_password"`

	// UsePasswordMigrationHook is set to true if the password should be migrated
	// using the password migration hook. If set, and the HashedPassword is empty, a
	// webhook will be called during login to migrate the password.
	UsePasswordMigrationHook bool `json:"use_password_migration_hook,omitempty"`
}

func (cp *CredentialsPassword) ShouldUsePasswordMigrationHook() bool {
	return cp != nil && cp.HashedPassword == "" && cp.UsePasswordMigrationHook
}
