// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package claims

import (
	"encoding/json"
	"strings"
)

type Locale string

func (l *Locale) UnmarshalJSON(data []byte) error {
	var linkedInLocale struct {
		Language string `json:"language"`
		Country  string `json:"country"`
	}
	if err := json.Unmarshal(data, &linkedInLocale); err == nil {
		switch {
		case linkedInLocale.Language == "":
			*l = Locale(linkedInLocale.Country)
		case linkedInLocale.Country == "":
			*l = Locale(linkedInLocale.Language)
		default:
			*l = Locale(strings.Join([]string{linkedInLocale.Language, linkedInLocale.Country}, "-"))
		}

		return nil
	}

	return json.Unmarshal(data, (*string)(l))
}
