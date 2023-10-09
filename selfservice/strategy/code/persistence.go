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

	RegistrationCodePersistenceProvider interface {
		RegistrationCodePersister() RegistrationCodePersister
	}

	RegistrationCodePersister interface {
		CreateRegistrationCode(context.Context, *CreateRegistrationCodeParams) (*RegistrationCode, error)
		UseRegistrationCode(ctx context.Context, flowID uuid.UUID, code string, addresses ...string) (*RegistrationCode, error)
		DeleteRegistrationCodesOfFlow(ctx context.Context, flowID uuid.UUID) error
		GetUsedRegistrationCode(ctx context.Context, flowID uuid.UUID) (*RegistrationCode, error)
	}

	LoginCodePersistenceProvider interface {
		LoginCodePersister() LoginCodePersister
	}

	LoginCodePersister interface {
		CreateLoginCode(context.Context, *CreateLoginCodeParams) (*LoginCode, error)
		UseLoginCode(ctx context.Context, flowID uuid.UUID, identityID uuid.UUID, code string) (*LoginCode, error)
		DeleteLoginCodesOfFlow(ctx context.Context, flowID uuid.UUID) error
		GetUsedLoginCode(ctx context.Context, flowID uuid.UUID) (*LoginCode, error)
	}
)
