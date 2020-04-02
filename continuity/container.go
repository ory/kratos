package continuity

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/x"
)

type Container struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	IdentityID *uuid.UUID `json:"identity_id" db:"identity_id"`

	// ExpiresAt defines when this container expires.
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`

	// Payload is the container's payload.
	Payload sqlxx.NullJSONRawMessage `json:"payload" db:"payload"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

func (c *Container) UTC() *Container {
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	c.ExpiresAt = c.ExpiresAt.UTC()
	return c
}

func (c *Container) TableName() string {
	return "continuity_containers"
}

func NewContainer(name string, o managerOptions) *Container {
	return &Container{
		ID:         x.NewUUID(),
		Name:       name,
		IdentityID: x.PointToUUID(o.iid),
		ExpiresAt:  time.Now().Add(o.ttl).UTC().Truncate(time.Second),
		Payload:    sqlxx.NullJSONRawMessage(o.payload),
	}
}

func (c *Container) Valid(identity uuid.UUID) error {
	if c.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("You mast restart the flow because the resumable session has expired."))
	}

	if identity != uuid.Nil && x.DerefUUID(c.IdentityID) != identity {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("You mast restart the flow because the resumable session was initiated by another person."))
	}

	return nil
}
