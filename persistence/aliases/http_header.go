// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package aliases

import (
	"database/sql/driver"
	"net/http"

	"github.com/ory/x/sqlxx"
)

type HTTPHeader http.Header

func (h HTTPHeader) Scan(value interface{}) error {
	return sqlxx.JSONScan(&h, value)
}

func (h HTTPHeader) Value() (driver.Value, error) {
	return sqlxx.JSONValue(&h)
}
