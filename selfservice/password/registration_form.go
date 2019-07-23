package password

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/form"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/gojsonschema"
	"github.com/ory/herodot"

	"github.com/ory/hive/schema"
	"github.com/ory/hive/x"
)

type RegistrationFormPayload struct {
	Password string          `json:"password"`
	Traits   json.RawMessage `json:"traits"`
}

func NewRegistrationFormDecoder() *RegistrationFormDecoder {
	return &RegistrationFormDecoder{}
}

type RegistrationFormDecoder struct{}

func (d *RegistrationFormDecoder) json(r *http.Request, p *RegistrationFormPayload) error {
	var fp struct {
		Request  string                 `json:"request"`
		Password string                 `json:"password"`
		Traits   map[string]interface{} `json:"traits"`
	}

	jd := json.NewDecoder(r.Body)
	jd.DisallowUnknownFields()

	if err := jd.Decode(&fp); err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP json body: %s", err.Error()))
	}

	if fp.Traits == nil {
		fp.Traits = make(map[string]interface{})
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(&fp); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithDebug(err.Error()).WithReasonf("Unable to transform JSON: %s", err.Error()))
	}

	if err := json.NewDecoder(&b).Decode(p); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithDebug(err.Error()).WithReasonf("Unable to transform JSON: %s", err.Error()))
	}

	return nil
}

func (d *RegistrationFormDecoder) form(r *http.Request, p *RegistrationFormPayload) error {
	var fp struct {
		Request  string            `form:"request"`
		Password string            `form:"password"`
		Traits   map[string]string `form:"traits"`
	}

	dec := form.NewDecoder()
	if err := r.ParseForm(); err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error()))
	} else if err := dec.Decode(&fp, r.PostForm); err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form payload: %s", err.Error()))
	}

	traits := json.RawMessage("{}")
	for path, value := range fp.Traits {
		var v interface{} = value
		var err error
		if x.IsValidNumber(value) {
			v, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				v, err = strconv.ParseFloat(value, 64)
			}
			if err != nil {
				return errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse number: %s", err.Error()))
			}
		} else if strings.ToLower(value) == "true" || strings.ToLower(value) == "false" {
			v, err = strconv.ParseBool(value)
			if err != nil {
				return errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse bool: %s", err.Error()))
			}
		}

		traits, err = sjson.SetBytes(traits, path, v)
		if err != nil {
			return errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse traits: %s", err.Error()))
		}
	}

	p.Traits = traits
	p.Password = fp.Password

	return nil
}

func (d *RegistrationFormDecoder) Decode(r *http.Request) (*RegistrationFormPayload, error) {
	ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP request content type: %s", err.Error()))
	}

	var payload RegistrationFormPayload
	if ct == "application/json" {
		err = d.json(r, &payload)
	} else {
		err = d.form(r, &payload)
	}

	if err != nil {
		return nil, err
	}

	if len(payload.Password) == 0 {
		return nil, errors.WithStack(schema.NewRequiredError("", gojsonschema.NewJsonContext("password", nil)))
	}

	if len(payload.Traits) == 0 {
		payload.Traits = json.RawMessage("{}")
	}

	return &payload, nil
}
