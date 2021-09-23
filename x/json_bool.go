package x

import (
	"errors"
	"fmt"
)

// ConvertibleBoolean can unmarshal both booleans and strings.
type ConvertibleBoolean bool

func (bit *ConvertibleBoolean) UnmarshalJSON(data []byte) error {
	asString := string(data)
	if asString == "true" || asString == `"true"` {
		*bit = true
	} else if asString == "false" || asString == `"false"` {
		*bit = false
	} else {
		return errors.New(fmt.Sprintf("Boolean unmarshal error: invalid input %s", asString))
	}
	return nil
}
