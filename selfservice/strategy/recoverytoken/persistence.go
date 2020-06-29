package recoverytoken

import (
	"context"
)

type (
	Persister interface {
		CreateRecoveryToken(ctx context.Context, token *Token) error
		UseRecoveryToken(ctx context.Context, token string) (*Token, error)
		DeleteRecoveryToken(ctx context.Context, token string) error
	}

	PersistenceProvider interface {
		RecoveryTokenPersister() Persister
	}
)
