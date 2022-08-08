package code

import (
	"context"
)

type (
	RecoveryCodePersister interface {
		CreateRecoveryCode(ctx context.Context, code *RecoveryCode) error
		UseRecoveryCode(ctx context.Context, code string) (*RecoveryCode, error)
		DeleteRecoveryCode(ctx context.Context, code string) error
	}

	RecoveryCodePersistenceProvider interface {
		RecoveryCodePersister() RecoveryCodePersister
	}
)
