// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"errors"
	"net/http"

	"github.com/ory/herodot"
)

var PseudoPanic = herodot.DefaultError{
	StatusField: http.StatusText(http.StatusInternalServerError),
	ErrorField:  "Code Bug Detected",
	ReasonField: "The code ended up at a place where it should not have. Please report this as an issue at https://github.com/ory/kratos",
	CodeField:   http.StatusInternalServerError,
}

func RecoverStatusCode(err error, fallback int) int {
	var sc herodot.StatusCodeCarrier
	if errors.As(err, &sc) {
		return sc.StatusCode()
	}
	return fallback
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
