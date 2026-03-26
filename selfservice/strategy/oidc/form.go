// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"encoding/json"

	"dario.cat/mergo"

	"github.com/ory/kratos/identity"
)

// merge merges the JSON messages in a and b. If a value is defined in both a and b, a wins.
func merge(a, b json.RawMessage) (identity.Traits, error) {
	if len(a) == 0 || string(a) == "{}" {
		return identity.Traits(b), nil
	}

	var aMap, bMap map[string]any

	if err := json.Unmarshal(a, &aMap); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &bMap); err != nil {
		return nil, err
	}

	// values in aMap (src) win.
	if err := mergo.Merge(&bMap, &aMap, mergo.WithOverride); err != nil {
		return nil, err
	}

	traits, err := json.Marshal(bMap)
	if err != nil {
		return nil, err
	}

	return traits, nil
}
