// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package container

type (
	richError interface {
		StatusCode() int
		Reason() string
	}
)
