package webauthn

import (
	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
)

var ErrNotEnoughCredentials = &jsonschema.ValidationError{
	Message: "unable to remove this security key because it would lock you out of your account", InstancePtr: "#/webauthn_remove"}
var ErrNoCredentials = errors.New("required credentials not found")
