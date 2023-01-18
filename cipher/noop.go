// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"context"
	"encoding/hex"

	"github.com/ory/kratos/driver/config"
)

// Noop is default cipher implementation witch does not do encryption

type NoopConfiguration interface {
	config.Provider
}

type Noop struct {
	c NoopConfiguration
}

func NewNoop(c NoopConfiguration) *Noop {
	return &Noop{c: c}
}

// Encrypt encode message to hex
func (c *Noop) Encrypt(_ context.Context, message []byte) (string, error) {
	return hex.EncodeToString(message), nil
}

// Decrypt decode the hex message
func (c *Noop) Decrypt(_ context.Context, ciphertext string) ([]byte, error) {
	return hex.DecodeString(ciphertext)
}
