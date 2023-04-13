package saml

import (
	"net/http"

	"github.com/ory/herodot"
	"google.golang.org/grpc/codes"
)

var (
	ErrScopeMissing = herodot.ErrBadRequest.
			WithError("authentication failed because a required scope was not granted").
			WithReason(`Unable to finish because one or more permissions were not granted. Please retry and accept all permissions.`)

	ErrIDTokenMissing = herodot.ErrBadRequest.
				WithError("authentication failed because id_token is missing").
				WithReason(`Authentication failed because no id_token was returned. Please accept the "openid" permission and try again.`)

	ErrAPIFlowNotSupported = herodot.ErrBadRequest.WithError("API-based flows are not supported for this method").
				WithReason("SAML SignIn and Registeration are only supported for flows initiated using the Browser endpoint.")

	ErrInvalidSAMLMetadataError = herodot.DefaultError{
		StatusField:   http.StatusText(http.StatusOK),
		ErrorField:    "Not valid SAML metadata file",
		CodeField:     http.StatusOK,
		GRPCCodeField: codes.InvalidArgument,
	}

	ErrInvalidCertificateError = herodot.DefaultError{
		StatusField:   http.StatusText(http.StatusOK),
		ErrorField:    "Not valid certificate",
		CodeField:     http.StatusOK,
		GRPCCodeField: codes.InvalidArgument,
	}

	ErrInvalidSAMLConfiguration = herodot.DefaultError{
		StatusField:   http.StatusText(http.StatusOK),
		ErrorField:    "Invalid SAML configuration in the configuration file",
		CodeField:     http.StatusOK,
		GRPCCodeField: codes.InvalidArgument,
	}
)
