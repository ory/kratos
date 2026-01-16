// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/x/decoderx"
)

func DecodeBody(p interface{}, r *http.Request, schema []byte, ds *url.URL) error {
	raw, err := sjson.SetBytes(schema,
		"properties.traits.$ref", ds.String()+"#/properties/traits")
	if err != nil {
		return errors.WithStack(err)
	}

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return errors.WithStack(err)
	}

	return decoderx.Decode(r, p, compiler, decoderx.HTTPDecoderSetValidatePayloads(true), decoderx.HTTPDecoderJSONFollowsFormFormat())
}
