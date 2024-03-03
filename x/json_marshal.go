// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import "encoding/json"

// ParseRawMessageOrEmpty parses a json.RawMessage and returns an empty map if the input is empty.
func ParseRawMessageOrEmpty(input json.RawMessage) (map[string]interface{}, error) {
	if len(input) == 0 {
		return map[string]interface{}{}, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(input, &m); err != nil {
		return nil, err
	}
	return m, nil
}
