package x

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var (
	regexMatchInt   = regexp.MustCompile("^[0-9]+$")
	regexMatchFloat = regexp.MustCompile("^[0-9]+\\.[0-9]+$")
	regexMatchBool  = regexp.MustCompile("^(?i)false|true|on$")
)

func TypeMap(m map[string]string) (map[string]interface{}, error) {
	jm := make(map[string]interface{})
	for k, v := range m {
		if regexMatchInt.MatchString(v) {
			vv, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, err
			}
			jm[k] = vv
		} else if regexMatchFloat.MatchString(v) {
			vv, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, err
			}
			jm[k] = vv
		} else if regexMatchBool.MatchString(v) {
			// Checkboxes have default values of `on` when checked, so set this to true.
			if strings.ToLower(v) == "on" {
				v = "true"
			}
			vv, err := strconv.ParseBool(strings.ToLower(v))
			if err != nil {
				return nil, err
			}
			jm[k] = vv
		} else {
			jm[k] = v
		}
	}

	return jm, nil
}

func UntypedMapToJSON(m map[string]string) (json.RawMessage, error) {
	jm, err := TypeMap(m)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(jm); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
