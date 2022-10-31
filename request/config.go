package request

import (
	"encoding/json"
	"net/http"

	"github.com/tidwall/gjson"
)

type (
	Auth struct {
		Type   string
		Config json.RawMessage
	}

	Config struct {
		Method      string      `json:"method"`
		URL         string      `json:"url"`
		TemplateURI string      `json:"body"`
		Header      http.Header `json:"header"`
		Auth        Auth        `json:"auth,omitempty"`
	}
)

func parseConfig(r json.RawMessage) (*Config, error) {
	type rawConfig struct {
		Method      string          `json:"method"`
		URL         string          `json:"url"`
		TemplateURI string          `json:"body"`
		Header      json.RawMessage `json:"header"`
		Auth        Auth            `json:"auth,omitempty"`
	}

	var rc rawConfig
	err := json.Unmarshal(r, &rc)
	if err != nil {
		return nil, err
	}

	rawHeader := gjson.ParseBytes(rc.Header).Map()
	hdr := http.Header{}

	_, ok := rawHeader["Content-Type"]
	if !ok {
		hdr.Set("Content-Type", ContentTypeJSON)
	}

	for key, value := range rawHeader {
		hdr.Set(key, value.String())
	}

	c := Config{
		Method:      rc.Method,
		URL:         rc.URL,
		TemplateURI: rc.TemplateURI,
		Header:      hdr,
		Auth:        rc.Auth,
	}

	return &c, nil
}
