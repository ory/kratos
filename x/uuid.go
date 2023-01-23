// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"github.com/gofrs/uuid"
)

var EmptyUUID uuid.UUID

func NewUUID() uuid.UUID {
	return uuid.Must(uuid.NewV4())
}

func ParseUUID(in string) uuid.UUID {
	return uuid.FromStringOrNil(in)
}
