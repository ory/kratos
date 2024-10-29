// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"

	"github.com/ory/kratos/session"
)

func CheckAALForTest(ctx context.Context, e *HookExecutor, s *session.Session, flow *Flow) error {
	return e.checkAAL(ctx, s, flow)
}
