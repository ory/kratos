package registration

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/decoderx"
)

func DecodeBody(p interface{}, r *http.Request, dec *decoderx.HTTP, conf *config.Config, schema []byte) error {
	ds, err := conf.DefaultIdentityTraitsSchemaURL()
	if err != nil {
		return err
	}
	raw, err := sjson.SetBytes(schema,
		"properties.traits.$ref", ds.String()+"#/properties/traits")
	if err != nil {
		return errors.WithStack(err)
	}

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return errors.WithStack(err)
	}

	return dec.Decode(r, p, compiler, decoderx.HTTPDecoderSetValidatePayloads(true), decoderx.HTTPDecoderJSONFollowsFormFormat())
}
