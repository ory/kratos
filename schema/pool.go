package schema

import "github.com/gofrs/uuid"

type (
	JsonSchema struct {
		ID  uuid.UUID `json:"-" faker:"uuid" db:"id" rw:"r"`
		Url string    `json:"-" faker:"-" db:"url"`
	}

	Pool interface {
		GetSchema(uuid uuid.UUID) (*JsonSchema, error)
		CreateSchema(s JsonSchema) error
	}

	PoolProvider interface {
		SchemaPool() Pool
	}
)
