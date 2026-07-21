// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthnx

import (
	"github.com/pkg/errors"

	"github.com/ory/jsonschema/v3"
)

var ErrNoCredentials = errors.New("required credentials not found")

func ErrNotEnoughCredentials() *jsonschema.ValidationError {
	return &jsonschema.ValidationError{Message: "unable to remove this security key because it would lock you out of your account", InstancePtr: "#/webauthn_remove"}
}

// ErrCredentialAlreadyRegistered is returned when a settings flow submits a credential whose ID
// is already registered on the identity. Compliant browsers prevent this via excludeCredentials,
// so this guards against clients that ignore the exclusion list.
func ErrCredentialAlreadyRegistered(instancePtr string) *jsonschema.ValidationError {
	return &jsonschema.ValidationError{Message: "this security key or passkey is already registered with your account", InstancePtr: instancePtr}
}
