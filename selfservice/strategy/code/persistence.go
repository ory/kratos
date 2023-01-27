// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"

	"github.com/gofrs/uuid"
)

type (
	RecoveryCodePersister interface {
		CreateRecoveryCode(ctx context.Context, dto *CreateRecoveryCodeParams) (*RecoveryCode, error)
		UseRecoveryCode(ctx context.Context, fID uuid.UUID, code string) (*RecoveryCode, error)
		DeleteRecoveryCodesOfFlow(ctx context.Context, fID uuid.UUID) error
	}

	RecoveryCodePersistenceProvider interface {
		RecoveryCodePersister() RecoveryCodePersister
	}

	VerificationCodePersister interface {
		CreateVerificationCode(context.Context, *CreateVerificationCodeParams) (*VerificationCode, error)
		UseVerificationCode(context.Context, uuid.UUID, string) (*VerificationCode, error)
		DeleteVerificationCodesOfFlow(context.Context, uuid.UUID) error
	}

	VerificationCodePersistenceProvider interface {
		VerificationCodePersister() VerificationCodePersister
	}
)
