// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
)

type Channel interface {
	ID() string
	Dispatch(ctx context.Context, msg Message) error
}
