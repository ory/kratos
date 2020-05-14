package aliases

import (
	"database/sql/driver"
	"github.com/ory/x/sqlxx"
	"net/http"
)

type HTTPHeader http.Header

func (h HTTPHeader) Scan(value interface{}) error {
	return sqlxx.JSONScan(&h, value)
}

func (h HTTPHeader) Value() (driver.Value, error) {
	return sqlxx.JSONValue(&h)
}
