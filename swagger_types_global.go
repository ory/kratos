// nolint:deadcode,unused
package main

import "github.com/ory/herodot"

// JSON API Error Response
//
// The standard Ory JSON API error format.
//
// swagger:model jsonError
type jsonError struct {
	// Contains error details
	//
	// required: true
	Error herodot.DefaultError `json:"error"`
}

// Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is typically 201.
//
// swagger:response emptyResponse
type emptyResponse struct{}
