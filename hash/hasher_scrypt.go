// Copyright © 2022 Ory Corp

package hash

type Scrypt struct {
	Cost           uint32
	Block          uint32
	Parrellization uint32
	SaltLength     uint32
	KeyLength      uint32
}
