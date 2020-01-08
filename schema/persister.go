package schema

import (
	"github.com/gofrs/uuid"
	"time"
)

type (
	Schema struct {
		ID  uuid.UUID `json:"-" faker:"uuid" db:"id" rw:"r"`
		URL string    `json:"url" faker:"-" db:"url"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" db:"updated_at"`
	}

	Persister interface {
		GetSchema(uuid uuid.UUID) (*Schema, error)
		GetSchemaByUrl(url string) (*Schema, error)
		RegisterSchema(s *Schema) error
	}

	PersistenceProvider interface {
		SchemaPersister() Persister
	}
)
