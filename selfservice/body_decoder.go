package selfservice

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
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

func (d *BodyDecoder) form(r *http.Request) (json.RawMessage, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error()))
	}

	payload := []byte("{}")
	for k := range r.PostForm {
		v := r.PostFormValue(k)

		var typed interface{} = v
		var err error
		if x.IsValidNumber(v) {
			typed, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				typed, err = strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse number: %s", err.Error()))
				}
			}
		} else if strings.ToLower(v) == "true" || strings.ToLower(v) == "false" {
			typed, err = strconv.ParseBool(v)
			if err != nil {
				return nil, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse bool: %s", err.Error()))
			}
		} else if strings.ToLower(v) == "__object__" {
			if len(gjson.GetBytes(payload, k).Raw) == 0 {
				typed = map[string]interface{}{}
			} else {
				continue
			}
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
		p, err = d.form(r)
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
