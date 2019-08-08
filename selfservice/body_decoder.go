package selfservice

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/x/jsonx"

	"github.com/ory/hive/x"
)

type BodyDecoder struct{}

func NewBodyDecoder() *BodyDecoder {
	return &BodyDecoder{}
}

func (d BodyDecoder) json(r *http.Request) (json.RawMessage, error) {
	var p json.RawMessage
	if err := jsonx.NewStrictDecoder(r.Body).Decode(&p); err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP json body: %s", err.Error()))
	}

	return p, nil
}

func (d *BodyDecoder) DecodeForm(form url.Values, o interface{}) (err error) {
	payload, err := d.form(form)
	if err != nil {
		return err
	}

	// This must not be a strict decoder
	return errors.WithStack(json.NewDecoder(bytes.NewBuffer(payload)).Decode(o))
}

func (d *BodyDecoder) ParseFormFieldOr(values []string, fallback interface{}) (typed interface{}) {
	out, err := d.ParseFormField(values)
	if err != nil {
		return fallback
	}
	return out
}

func (d *BodyDecoder) ParseFormField(values []string) (typed interface{}, err error) {
	if len(values) == 0 {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Values must have at least one element but got none."))
	}

	value := values[0]
	// This handles the special case of checkboxes:
	//
	//  <input type="hidden" name="signup_for_newsletter" value="false" />
	//  <input type="checkbox" name="signup_for_newsletter" value="true" />
	if len(values) > 1 {
		value = values[len(values)-1]
	}
	typed = value
	if x.IsValidNumber(value) {
		typed, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			typed, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse number: %s", err.Error()))
			}
		}
	} else if strings.ToLower(value) == "true" || strings.ToLower(value) == "false" {
		typed, err = strconv.ParseBool(value)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse bool: %s", err.Error()))
		}
	}

	return typed, err
}

func (d *BodyDecoder) form(form url.Values) (json.RawMessage, error) {
	payload := []byte("{}")
	for k := range form {
		typed, err := d.ParseFormField(form[k])
		if err != nil {
			return nil, err
		}

		payload, err = sjson.SetBytes(payload, k, typed)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return payload, nil
}

func (d *BodyDecoder) Decode(r *http.Request, o interface{}) (err error) {
	ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP request content type: %s", err.Error()))
	}

	var p json.RawMessage
	if ct == "application/json" {
		p, err = d.json(r)
	} else {
		if err := r.ParseForm(); err != nil {
			return errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error()))
		}
		p, err = d.form(r.PostForm)
	}

	if err != nil {
		return err
	}

	// This must not be a strict decoder
	if err := json.NewDecoder(bytes.NewBuffer(p)).Decode(o); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
