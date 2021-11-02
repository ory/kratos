package totp

import (
	"bytes"
	"context"
	"encoding/base64"
	"image/png"

	"github.com/pkg/errors"
	"github.com/pquerna/otp"
	stdtotp "github.com/pquerna/otp/totp"

	"github.com/ory/kratos/driver/config"
)

// rfc4226 recommends:
//
// The algorithm MUST use a strong shared secret. The length of
// the shared secret MUST be at least 128 bits. This document
// RECOMMENDs a shared secret length of 160 bits.
//
// So we need 160/8 = 20 key length. stdtotp.Generate uses the key
// length for reading from crypto.Rand.
const secretSize = 160 / 8

func NewKey(ctx context.Context, accountName string, d interface {
	config.Provider
}) (*otp.Key, error) {
	key, err := stdtotp.Generate(stdtotp.GenerateOpts{
		Issuer:      d.Config(ctx).TOTPIssuer(),
		AccountName: accountName,
		SecretSize:  secretSize,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return key, err
}

func KeyToHTMLImage(key *otp.Key) (string, error) {
	var buf bytes.Buffer
	img, err := key.Image(256, 256)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if err := png.Encode(&buf, img); err != nil {
		return "", errors.WithStack(err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
