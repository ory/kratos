// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// GenericErrorPayload generic error payload
//
// swagger:model genericErrorPayload
type GenericErrorPayload struct {

	// Code represents the error status code (404, 403, 401, ...).
	Code int64 `json:"code,omitempty"`

	// Debug contains debug information. This is usually not available and has to be enabled.
	Debug string `json:"debug,omitempty"`

	// details
	Details map[string]interface{} `json:"details,omitempty"`

	// message
	Message string `json:"message,omitempty"`

	// reason
	Reason string `json:"reason,omitempty"`

	// request
	Request string `json:"request,omitempty"`

	// status
	Status string `json:"status,omitempty"`
}

// Validate validates this generic error payload
func (m *GenericErrorPayload) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *GenericErrorPayload) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GenericErrorPayload) UnmarshalBinary(b []byte) error {
	var res GenericErrorPayload
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
