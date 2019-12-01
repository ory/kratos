package aliases

import (
	"database/sql/driver"
	"net/http"
)

type HTTPHeader http.Header

func (h HTTPHeader) Scan(value interface{}) error {
	return JSONScan(&h, value)
}

func (h HTTPHeader) Value() (driver.Value, error) {
	return JSONValue(&h)
}
