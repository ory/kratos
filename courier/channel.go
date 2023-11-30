package courier

import (
	"context"
)

type Channel interface {
	ID() string
	Dispatch(ctx context.Context, msg Message) error
}
