// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"net/http"
	"strings"
)

func bearerTokenFromRequest(r *http.Request) (string, bool) {
	parts := strings.Split(r.Header.Get("Authorization"), " ")

	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1], true
	}

	return "", false
}
