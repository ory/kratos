package selfservice

import "github.com/ory/herodot"

var ErrIDTokenMissing = herodot.ErrBadRequest.
	WithError("authentication failed because id_token is missing").
	WithReasonf(`Authentication failed because no id_token was returned. Please accept the "openid" permission and try again.`)

var ErrScopeMissing = herodot.ErrBadRequest.
	WithError("authentication failed because a required scope was not granted").
	WithReasonf(`Unable to finish because one or more permissions were not granted. Please retry and accept all permissions.`)

var ErrLoginRequestExpired = herodot.ErrBadRequest.
	WithError("login request expired")

var ErrRegistrationRequestExpired = herodot.ErrBadRequest.
	WithError("registration request expired")
