package link

import (
	"context"
)

type (
	RecoveryTokenPersister interface {
		CreateRecoveryToken(ctx context.Context, token *RecoveryToken) error
		UseRecoveryToken(ctx context.Context, token string) (*RecoveryToken, error)
		DeleteRecoveryToken(ctx context.Context, token string) error
	}

	RecoveryTokenPersistenceProvider interface {
		RecoveryTokenPersister() RecoveryTokenPersister
	}

	VerificationTokenPersister interface {
		CreateVerificationToken(ctx context.Context, token *VerificationToken) error
		UseVerificationToken(ctx context.Context, token string) (*VerificationToken, error)
		DeleteVerificationToken(ctx context.Context, token string) error
	}

	VerificationTokenPersistenceProvider interface {
		VerificationTokenPersister() VerificationTokenPersister
	}
)
