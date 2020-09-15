package oidc

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/identity"
)

func decoderRegistration(ref string) (decoderx.HTTPDecoderOption, error) {
	raw, err := sjson.SetBytes([]byte(registrationFormPayloadSchema), "properties.traits.$ref", ref+"#/properties/traits")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	o, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return o, nil
}

type decodedForm struct {
	Traits   map[string]interface{}    `json:"traits"`
	Recovery recoverySecurityQuestions `json:"recovery"`
}

type recoverySecurityQuestions struct {
	SecurityQuestions map[string]string `json:"security_questions"`
}

// merge merges the userFormValues (extracted from the initial POST request) prefixed with `traits` (encoded) with the
// values coming from the OpenID Provider (openIDProviderValues).
func merge(userFormValues string, openIDProviderValues json.RawMessage, option decoderx.HTTPDecoderOption) (identity.Traits, error) {
	if userFormValues == "" {
		return identity.Traits(openIDProviderValues), nil
	}

	var df decodedForm

	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(userFormValues))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err := decoderx.NewHTTP().Decode(
		req, &df,
		decoderx.HTTPFormDecoder(),
		option,
		decoderx.HTTPDecoderSetValidatePayloads(false),
	); err != nil {
		return nil, err
	}

	var decodedTraits map[string]interface{}
	if err := json.NewDecoder(bytes.NewBuffer(openIDProviderValues)).Decode(&decodedTraits); err != nil {
		return nil, err
	}

	// decoderForm (coming from POST request) overrides decodedTraits (coming from OP)
	if err := mergo.Merge(&decodedTraits, df.Traits, mergo.WithOverride); err != nil {
		return nil, err
	}

	var result bytes.Buffer
	if err := json.NewEncoder(&result).Encode(decodedTraits); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}
