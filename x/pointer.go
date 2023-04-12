// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import "github.com/gofrs/uuid"

func PointToUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
