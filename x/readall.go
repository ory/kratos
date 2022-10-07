// Copyright © 2022 Ory Corp

package x

import (
	"io"
)

func MustReadAll(r io.Reader) []byte {
	all, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return all
}
