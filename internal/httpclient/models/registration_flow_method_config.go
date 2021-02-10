// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// RegistrationFlowMethodConfig registration flow method config
//
// swagger:model registrationFlowMethodConfig
type RegistrationFlowMethodConfig struct {

	// Action should be used as the form action URL `<form action="{{ .Action }}" method="post">`.
	// Required: true
	Action *string `json:"action"`

	// fields
	// Required: true
	Fields FormFields `json:"fields"`

	// messages
	Messages Messages `json:"messages,omitempty"`

	// Method is the form method (e.g. POST)
	// Required: true
	Method *string `json:"method"`

	// Providers is set for the "oidc" registration method.
	Providers []*FormField `json:"providers"`
}

// Validate validates this registration flow method config
func (m *RegistrationFlowMethodConfig) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAction(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateFields(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMessages(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMethod(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateProviders(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *RegistrationFlowMethodConfig) validateAction(formats strfmt.Registry) error {

	if err := validate.Required("action", "body", m.Action); err != nil {
		return err
	}

	return nil
}

func (m *RegistrationFlowMethodConfig) validateFields(formats strfmt.Registry) error {

	if err := validate.Required("fields", "body", m.Fields); err != nil {
		return err
	}

	if err := m.Fields.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("fields")
		}
		return err
	}

	return nil
}

func (m *RegistrationFlowMethodConfig) validateMessages(formats strfmt.Registry) error {

	if swag.IsZero(m.Messages) { // not required
		return nil
	}

	if err := m.Messages.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("messages")
		}
		return err
	}

	return nil
}

func (m *RegistrationFlowMethodConfig) validateMethod(formats strfmt.Registry) error {

	if err := validate.Required("method", "body", m.Method); err != nil {
		return err
	}

	return nil
}

func (m *RegistrationFlowMethodConfig) validateProviders(formats strfmt.Registry) error {

	if swag.IsZero(m.Providers) { // not required
		return nil
	}

	for i := 0; i < len(m.Providers); i++ {
		if swag.IsZero(m.Providers[i]) { // not required
			continue
		}

		if m.Providers[i] != nil {
			if err := m.Providers[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("providers" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *RegistrationFlowMethodConfig) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RegistrationFlowMethodConfig) UnmarshalBinary(b []byte) error {
	var res RegistrationFlowMethodConfig
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
