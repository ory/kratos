// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import "encoding/json"

// TransientPayloadContainer is a container for transient data to pass along to any webhooks
type TransientPayloadContainer struct {
	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload TransientPayload `json:"transient_payload,omitempty" form:"transient_payload"`
}

type TransientPayload json.RawMessage

func (t TransientPayload) Unmarshal() (map[string]interface{}, error) {
	if len(t) == 0 {
		return map[string]interface{}{}, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(t, &m); err != nil {
		return nil, err
	}
	return m, nil
}
