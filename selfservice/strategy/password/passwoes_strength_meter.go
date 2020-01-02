package password

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/x/decoderx"
	_ "github.com/ory/x/jsonschemax/fileloader"
	_ "github.com/ory/x/jsonschemax/httploader"

	"github.com/ory/gojsonschema"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

const (
	PasswordStrengthMeterPath = "/password/strength/check"

	passwordStrengthMeterFormPayloadSchema = `{
	"$id": "https://schemas.ory.sh/kratos/selfservice/password/password_strength_meter/config.schema.json",
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"required": ["password"],
	"properties": {
			"password": {
					"type": "string",
					"minLength": 1
			}
	}
}`
)

type PasswordStrengthMeterFormPayload struct {
	Password string `json:"password"`
}

type PasswordStrengthMeter struct {
	Score int `json:"score"`
}

func (s *Strategy) RegisterPasswordStrengthMeterRoutes(r *x.RouterPublic) {
	r.POST(PasswordStrengthMeterPath, s.handlePasswordStrengthMeter)
}

func (s *Strategy) decoderPasswordStrength() (decoderx.HTTPDecoderOption, error) {
	raw, err := sjson.SetBytes([]byte(passwordStrengthMeterFormPayloadSchema), "properties.traits.$ref", s.c.DefaultIdentityTraitsSchemaURL().String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	o, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return o, nil
}

// TODO - write error handling
func (s *Strategy) handlePasswordStrengthMeterError(w http.ResponseWriter, r *http.Request, p *PasswordStrengthMeterFormPayload, err error) {

}

// TODO - error handling and testing
func (s *Strategy) handlePasswordStrengthMeter(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p PasswordStrengthMeterFormPayload
	option, err := s.decoderPasswordStrength()
	if err != nil {
		s.handlePasswordStrengthMeterError(w, r, nil, err)
		return
	}

	if err := decoderx.NewHTTP().Decode(r, &p,
		decoderx.HTTPFormDecoder(),
		option,
		decoderx.HTTPDecoderSetIgnoreParseErrorsStrategy(decoderx.ParseErrorIgnore),
		decoderx.HTTPDecoderSetValidatePayloads(false),
	); err != nil {
		s.handlePasswordStrengthMeterError(w, r, &p, err)
		return
	}

	if len(p.Password) == 0 {
		s.handlePasswordStrengthMeterError(w, r, &p, errors.WithStack(schema.NewRequiredError("", gojsonschema.NewJsonContext("password", nil))))
		return
	}

	score := zxcvbn.PasswordStrength(p.Password, nil).Score
	data, err := json.Marshal(PasswordStrengthMeter{
		Score: score,
	})
	if err != nil {
		s.handlePasswordStrengthMeterError(w, r, &p, errors.WithStack(schema.NewRequiredError("", gojsonschema.NewJsonContext("password", nil))))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	return
}
