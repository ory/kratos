package x

import (
	"context"

	"github.com/gofrs/uuid"
)

type NetworkIDProvider interface {
	NetworkID(ctx context.Context) uuid.UUID
}
