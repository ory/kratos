// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// RecoveryFlowMethod RecoveryFlowMethod recovery flow method
//
// swagger:model recoveryFlowMethod
type RecoveryFlowMethod struct {

	// config
	// Required: true
	Config *RecoveryFlowMethodConfig `json:"config"`

	// Method contains the request credentials type.
	// Required: true
	Method *string `json:"method"`
}

// Validate validates this recovery flow method
func (m *RecoveryFlowMethod) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateConfig(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMethod(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *RecoveryFlowMethod) validateConfig(formats strfmt.Registry) error {

	if err := validate.Required("config", "body", m.Config); err != nil {
		return err
	}

	if m.Config != nil {
		if err := m.Config.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("config")
			}
			return err
		}
	}

	return nil
}

func (m *RecoveryFlowMethod) validateMethod(formats strfmt.Registry) error {

	if err := validate.Required("method", "body", m.Method); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this recovery flow method based on the context it is used
func (m *RecoveryFlowMethod) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateConfig(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *RecoveryFlowMethod) contextValidateConfig(ctx context.Context, formats strfmt.Registry) error {

	if m.Config != nil {
		if err := m.Config.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("config")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *RecoveryFlowMethod) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RecoveryFlowMethod) UnmarshalBinary(b []byte) error {
	var res RecoveryFlowMethod
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
