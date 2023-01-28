// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
