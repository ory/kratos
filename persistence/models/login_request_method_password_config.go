package models

import "github.com/gobuffalo/uuid"

type LoginRequestMethodPasswordConfig struct {
	ID  uuid.UUID `db:"id"`
	Bla string    `db:"bla"`
}

func (c *LoginRequestMethodPasswordConfig) GetID() uuid.UUID {
	return c.ID
}
