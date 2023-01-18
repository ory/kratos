// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hash

type Scrypt struct {
	Cost           uint32
	Block          uint32
	Parrellization uint32
	SaltLength     uint32
	KeyLength      uint32
}
