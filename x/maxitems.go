// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

// MaxItemsPerPage is used to prevent DoS attacks against large lists by limiting the items per page to 500.
func MaxItemsPerPage(is int) int {
	if is > 500 {
		return 500
	}
	return is
}
