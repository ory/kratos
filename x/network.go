// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"context"

	"github.com/gofrs/uuid"
)

type NetworkIDProvider interface {
	NetworkID(ctx context.Context) uuid.UUID
}
