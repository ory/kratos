package main

// Error response
//
// Error responses are sent when an error (e.g. unauthorized, bad request, ...) occurred.
//
// swagger:model genericError
// nolint:deadcode,unused
type genericError struct {
	Error genericErrorPayload `json:"error"`
}

type genericErrorPayload struct {
	// Code represents the error status code (404, 403, 401, ...).
	//
	// example: 404
	Code int `json:"code,omitempty"`

	Status string `json:"status,omitempty"`

	Request string `json:"request,omitempty"`

	Reason string `json:"reason,omitempty"`

	Details []map[string]interface{} `json:"details,omitempty"`

	Message string `json:"message"`

	// Debug contains debug information. This is usually not available and has to be enabled.
	//
	// example: The database adapter was unable to find the element
	Debug string `json:"debug"`
}

// Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is
// typically 201.
//
// swagger:response emptyResponse
// nolint:deadcode,unused
type emptyResponse struct{}
