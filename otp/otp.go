package otp

import (
	"github.com/pkg/errors"

	"github.com/ory/x/randx"
)

// Entropy sets the number of characters used for generating verification codes. This must not be
// changed to another value as we only have 32 characters available in the SQL schema.
const Entropy = 32

func New() (string, error) {
	code, err := randx.RuneSequence(Entropy, randx.AlphaNum)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(code), nil
}
