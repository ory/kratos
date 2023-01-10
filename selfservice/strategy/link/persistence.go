// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link

import (
	"context"

	"github.com/gofrs/uuid"
)

type (
	RecoveryTokenPersister interface {
		CreateRecoveryToken(ctx context.Context, token *RecoveryToken) error
		UseRecoveryToken(ctx context.Context, fID uuid.UUID, token string) (*RecoveryToken, error)
		DeleteRecoveryToken(ctx context.Context, token string) error
	}

	RecoveryTokenPersistenceProvider interface {
		RecoveryTokenPersister() RecoveryTokenPersister
	}

	VerificationTokenPersister interface {
		CreateVerificationToken(ctx context.Context, token *VerificationToken) error
		UseVerificationToken(ctx context.Context, fID uuid.UUID, token string) (*VerificationToken, error)
		DeleteVerificationToken(ctx context.Context, token string) error
	}

	VerificationTokenPersistenceProvider interface {
		VerificationTokenPersister() VerificationTokenPersister
	}
)
