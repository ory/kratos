// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"errors"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
)

type WithIdentityIDError struct {
	err        error
	identityID uuid.UUID
}

func (e *WithIdentityIDError) Error() string {
	return e.err.Error()
}

func (e *WithIdentityIDError) Unwrap() error {
	return e.err
}

func (e *WithIdentityIDError) IdentityID() uuid.UUID {
	return e.identityID
}

func WrapWithIdentityIDError(err error, identityID uuid.UUID) error {
	if err == nil {
		return nil
	}

	return &WithIdentityIDError{
		err:        err,
		identityID: identityID,
	}
}

var (
	PseudoPanic = herodot.DefaultError{
		StatusField: http.StatusText(http.StatusInternalServerError),
		ErrorField:  "Code Bug Detected",
		ReasonField: "The code ended up at a place where it should not have. Please report this as an issue at https://github.com/ory/kratos",
		CodeField:   http.StatusInternalServerError,
	}
	PageTokenInvalid = herodot.ErrBadRequest.WithReason("The page token is invalid, do not craft your own page tokens")
)

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
