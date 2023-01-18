// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package node

func PasswordLoginOrder(in []string) []string {
	if len(in) == 0 {
		return []string{"password"}
	}
	if len(in) == 1 {
		return append(in, "password")
	}
	return append([]string{in[0], "password"}, in[1:]...)
}
