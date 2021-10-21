package x

import (
	"fmt"
)

// ConvertibleBoolean can unmarshal both booleans and strings.
type ConvertibleBoolean bool

func (bit *ConvertibleBoolean) UnmarshalJSON(data []byte) error {
	asString := string(data)
	switch asString {
	case "true", `"true"`:
		*bit = true
	case "false", `"false"`:
		*bit = false
	default:
		return fmt.Errorf("boolean unmarshal error: invalid input %s", asString)
	}
	return nil
}
