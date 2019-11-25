package models

import "github.com/gobuffalo/uuid"

type LoginRequestMethodConfig interface {
	GetID() uuid.UUID
}
