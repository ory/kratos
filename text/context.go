// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"encoding/json"
)

func context(ctx map[string]any) []byte {
	if len(ctx) == 0 {
		panic("context must not be empty")
	}
	res, err := json.Marshal(ctx)
	if err != nil {
		panic(err)
	}
	return res
}
