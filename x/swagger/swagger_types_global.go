// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package swagger

import "github.com/ory/herodot"

// JSON API Error Response
//
// The standard Ory JSON API error format.
//
// swagger:model errorGeneric
type ErrorGeneric struct {
	// Contains error details
	//
	// required: true
	Error GenericError `json:"error"`
}

// swagger:model genericError
type GenericError struct{ herodot.DefaultError }

// Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is typically 204.
//
// swagger:response emptyResponse
type _ struct{}
