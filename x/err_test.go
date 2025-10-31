// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x_test

import (
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestWrapWithIdentityIDError(t *testing.T) {
	t.Run("case=wraps error with identity ID", func(t *testing.T) {
		baseErr := errors.New("test error")
		identityID := uuid.Must(uuid.NewV4())

		wrappedErr := x.WrapWithIdentityIDError(baseErr, identityID)

		require.NotNil(t, wrappedErr)
		assert.Equal(t, "test error", wrappedErr.Error())

		var withIDErr *x.WithIdentityIDError
		require.True(t, errors.As(wrappedErr, &withIDErr))
		assert.Equal(t, identityID, withIDErr.IdentityID())
	})

	t.Run("case=unwraps to original error", func(t *testing.T) {
		baseErr := errors.New("original error")
		identityID := uuid.Must(uuid.NewV4())

		wrappedErr := x.WrapWithIdentityIDError(baseErr, identityID)

		unwrappedErr := errors.Unwrap(wrappedErr)
		assert.Equal(t, baseErr, unwrappedErr)
	})

	t.Run("case=returns nil when wrapping nil error", func(t *testing.T) {
		identityID := uuid.Must(uuid.NewV4())

		wrappedErr := x.WrapWithIdentityIDError(nil, identityID)

		assert.Nil(t, wrappedErr)
	})

	t.Run("case=preserves identity ID with nil UUID", func(t *testing.T) {
		baseErr := errors.New("test error")
		var identityID uuid.UUID // nil UUID

		wrappedErr := x.WrapWithIdentityIDError(baseErr, identityID)

		var withIDErr *x.WithIdentityIDError
		require.True(t, errors.As(wrappedErr, &withIDErr))
		assert.Equal(t, uuid.Nil, withIDErr.IdentityID())
	})

	t.Run("case=can wrap already wrapped error", func(t *testing.T) {
		baseErr := errors.New("base error")
		firstID := uuid.Must(uuid.NewV4())
		secondID := uuid.Must(uuid.NewV4())

		firstWrap := x.WrapWithIdentityIDError(baseErr, firstID)
		secondWrap := x.WrapWithIdentityIDError(firstWrap, secondID)

		var withIDErr *x.WithIdentityIDError
		require.True(t, errors.As(secondWrap, &withIDErr))
		// Should get the outermost identity ID
		assert.Equal(t, secondID, withIDErr.IdentityID())
	})

	t.Run("case=works with errors.Is", func(t *testing.T) {
		baseErr := errors.New("base error")
		identityID := uuid.Must(uuid.NewV4())

		wrappedErr := x.WrapWithIdentityIDError(baseErr, identityID)

		assert.True(t, errors.Is(wrappedErr, baseErr))
	})
}
