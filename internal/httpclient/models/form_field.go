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

// FormField FormField FormField FormField Field represents a HTML Form Field
//
// swagger:model formField
type FormField struct {

	// Disabled is the equivalent of `<input {{if .Disabled}}disabled{{end}}">`
	Disabled bool `json:"disabled,omitempty"`

	// messages
	Messages Messages `json:"messages,omitempty"`

	// Name is the equivalent of `<input name="{{.Name}}">`
	// Required: true
	Name *string `json:"name"`

	// Pattern is the equivalent of `<input pattern="{{.Pattern}}">`
	Pattern string `json:"pattern,omitempty"`

	// Required is the equivalent of `<input required="{{.Required}}">`
	Required bool `json:"required,omitempty"`

	// Type is the equivalent of `<input type="{{.Type}}">`
	// Required: true
	Type *string `json:"type"`

	// Value is the equivalent of `<input value="{{.Value}}">`
	Value interface{} `json:"value,omitempty"`
}

// Validate validates this form field
func (m *FormField) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateMessages(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateType(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *FormField) validateMessages(formats strfmt.Registry) error {
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

func (m *FormField) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	return nil
}

func (m *FormField) validateType(formats strfmt.Registry) error {

	if err := validate.Required("type", "body", m.Type); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this form field based on the context it is used
func (m *FormField) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateMessages(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *FormField) contextValidateMessages(ctx context.Context, formats strfmt.Registry) error {

	if err := m.Messages.ContextValidate(ctx, formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("messages")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *FormField) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *FormField) UnmarshalBinary(b []byte) error {
	var res FormField
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
