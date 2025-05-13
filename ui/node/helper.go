// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package node

func PasswordLoginOrder(in []string) []string {
	if len(in) == 0 {
		return []string{"csrf_token", "password"}
	}
	if len(in) == 1 {
		return append([]string{"csrf_token"}, in[0], "password")
	}
	return append([]string{"csrf_token", in[0], "password"}, in[1:]...)
}
