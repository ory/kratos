// Ory Hive
//
// Welcome to the ORY Hive HTTP API documentation!
//
//     Schemes: http, https
//     Host:
//     BasePath: /
//     Version: latest
//
//     Consumes:
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Produces:
//     - application/json
//
//     Extensions:
//     ---
//     x-request-id: string
//     x-forwarded-proto: string
//     ---
//
// swagger:meta
package main

// Error response
//
// Error responses are sent when an error (e.g. unauthorized, bad request, ...) occurred.
//
// swagger:model genericError
type genericError struct {
	// Name is the error name.
	//
	// required: true
	// example: The requested resource could not be found
	Name string `json:"error"`

	// Hint contains further information on the nature of the error.
	//
	// example: Object with RequestID 12345 does not exist
	Hint string `json:"error_hint"`

	// Code represents the error status code (404, 403, 401, ...).
	//
	// example: 404
	Code int `json:"error_code"`

	// Debug contains debug information. This is usually not available and has to be enabled.
	//
	// example: The database adapter was unable to find the element
	Debug string `json:"error_debug"`
}

// Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is
// typically 201.
//
// swagger:response emptyResponse
type emptyResponse struct{}
