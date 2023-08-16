// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrDuplicateCredentials(t *testing.T) {
	inner := errors.New("inner error")
	err := &ErrDuplicateCredentials{inner, nil, nil, ""}
	assert.ErrorIs(t, err, inner)
}
