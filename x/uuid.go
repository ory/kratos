// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	db "github.com/gofrs/uuid"
	"github.com/google/uuid"
)

var EmptyUUID db.UUID

func NewUUID() db.UUID {
	return db.UUID(uuid.New())
}

func ParseUUID(in string) db.UUID {
	id, _ := uuid.Parse(in)
	return db.UUID(id)
}

func IsZeroUUID(id db.UUID) bool {
	return id == db.UUID{}
}
