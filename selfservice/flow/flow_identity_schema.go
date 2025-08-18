// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net/url"

	"github.com/ory/kratos/driver/config"
)

// swagger:type string
type IdentitySchema string

// Scan implements the Scanner interface.
func (is *IdentitySchema) Scan(value any) error {
	var v sql.NullString
	if err := (&v).Scan(value); err != nil {
		return err
	}
	*is = IdentitySchema(v.String)
	return nil
}

// Value implements the driver Valuer interface.
func (is *IdentitySchema) Value() (driver.Value, error) {
	if is == nil || len(*is) == 0 {
		return sql.NullString{}.Value()
	}
	return sql.NullString{Valid: true, String: string(*is)}.Value()
}

// URL returns the URL of the identity schema, or the default identity traits
// schema URL if the schema is empty.
func (is *IdentitySchema) URL(ctx context.Context, config *config.Config) (*url.URL, error) {
	if is == nil || len(*is) == 0 {
		return config.DefaultIdentityTraitsSchemaURL(ctx)
	}
	schemas, err := config.IdentityTraitsSchemas(ctx)
	if err != nil {
		return nil, err
	}
	schema, err := schemas.FindSchemaByID(string(*is))
	if err != nil {
		return nil, err
	}

	return config.ParseURI(schema.URL)
}

// ID returns the ID of the identity schema, or the default identity schema ID.
func (is *IdentitySchema) ID(ctx context.Context, config *config.Config) string {
	if is == nil || len(*is) == 0 {
		return config.DefaultIdentityTraitsSchemaID(ctx)
	}
	return string(*is)
}
