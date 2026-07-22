//go:build !commercial

// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/x"
)

func TestEdition(t *testing.T) {
	assert.Equal(t, "oss", x.Edition())
}
