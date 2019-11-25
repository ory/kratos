package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type LoginRequestMethods map[CredentialsType]*LoginRequestMethod

func (ns LoginRequestMethods) Scan(value interface{}) error {
	var b bytes.Buffer
	switch v := value.(type) {
	case []byte:
		b.Write(v)
	case string:
		b.WriteString(v)
	default:
		return errors.Errorf("unable to decode value of type: %T %v", value, value)
	}

	if err := json.NewDecoder(&b).Decode(&ns); err != nil {
		return fmt.Errorf("unable to decode payload to LoginRequestMethods: %s", err)
	}

	return nil
}

func (ns LoginRequestMethods) Value() (driver.Value, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(&ns); err != nil {
		return nil, err
	}
	return b.String(), nil
}

type LoginRequestMethod struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Method contains the request credentials type.
	Method CredentialsType `json:"credentials_type" db:"credentials_type"`

	// Config is the credential type's config.
	Config LoginRequestMethodConfig `json:"config" db:"config"`
}
