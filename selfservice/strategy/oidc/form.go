package oidc

import (
	"bytes"
	"encoding/json"
	"github.com/imdario/mergo"
	"github.com/ory/kratos/identity"
)

// merge merges the userFormValues (extracted from the initial POST request) prefixed with `traits` (encoded) with the
// values coming from the OpenID Provider (openIDProviderValues).
func merge(containerTraits json.RawMessage, openIDProviderValues json.RawMessage) (identity.Traits, error) {
	if len(containerTraits) == 0 || string(containerTraits) == "{}" {
		return identity.Traits(openIDProviderValues), nil
	}

	var pt map[string]interface{}
	if err := json.NewDecoder(bytes.NewBuffer(openIDProviderValues)).Decode(&pt); err != nil {
		return nil, err
	}

	var ct map[string]interface{}
	if err := json.NewDecoder(bytes.NewBuffer(containerTraits)).Decode(&ct); err != nil {
		return nil, err
	}

	// decoderForm (coming from POST request) overrides decodedTraits (coming from OP)
	if err := mergo.Merge(&pt, &ct, mergo.WithOverride); err != nil {
		return nil, err
	}

	var result bytes.Buffer
	if err := json.NewEncoder(&result).Encode(pt); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}
