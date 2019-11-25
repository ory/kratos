package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"net/http"

	"github.com/go-errors/errors"
)

type HTTPHeader http.Header

func (ns HTTPHeader) Scan(value interface{}) error {
	var b bytes.Buffer
	switch v := value.(type) {
	case []byte:
		b.Write(v)
	case string:
		b.WriteString(v)
	default:
		return errors.Errorf("unable to decode value of type: %T %v", value,value)
	}

	if err := json.NewDecoder(&b).Decode(&ns); err != nil {
		return errors.Errorf("unable to decode payload to LoginRequestMethods: %s", err)
	}

	return nil
}

func (ns HTTPHeader) Value() (driver.Value, error) {
	var b bytes.Buffer
	if 	err := json.NewEncoder(&b).Encode(&ns); err != nil {
		return nil, err
	}
	return b.String(),nil
}

