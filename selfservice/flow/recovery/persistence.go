// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type (
	FlowPersister interface {
		CreateRecoveryFlow(context.Context, *Flow) error
		GetRecoveryFlow(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateRecoveryFlow(context.Context, *Flow) error
		DeleteExpiredRecoveryFlows(context.Context, time.Time, int) error
	}
	FlowPersistenceProvider interface {
		RecoveryFlowPersister() FlowPersister
	}
)
