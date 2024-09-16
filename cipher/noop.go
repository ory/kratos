// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"context"
	"encoding/hex"
)

// Noop is default cipher implementation witch does not do encryption
type Noop struct{}

func NewNoop() *Noop {
	return &Noop{}
}

// Encrypt encode message to hex
func (*Noop) Encrypt(_ context.Context, message []byte) (string, error) {
	return hex.EncodeToString(message), nil
}

// Decrypt decode the hex message
func (*Noop) Decrypt(_ context.Context, ciphertext string) ([]byte, error) {
	return hex.DecodeString(ciphertext)
}
