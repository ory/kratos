// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// VerifiableAddress verifiable address
//
// swagger:model VerifiableAddress
type VerifiableAddress struct {

	// expires at
	// Required: true
	// Format: date-time
	ExpiresAt *strfmt.DateTime `json:"expires_at"`

	// id
	// Required: true
	// Format: uuid4
	ID UUID `json:"id"`

	// value
	// Required: true
	Value *string `json:"value"`

	// verified
	// Required: true
	Verified *bool `json:"verified"`

	// verified at
	// Format: date-time
	VerifiedAt strfmt.DateTime `json:"verified_at,omitempty"`

	// via
	// Required: true
	Via VerifiableAddressType `json:"via"`
}

// Validate validates this verifiable address
func (m *VerifiableAddress) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateExpiresAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateValue(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVerified(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVerifiedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVia(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *VerifiableAddress) validateExpiresAt(formats strfmt.Registry) error {

	if err := validate.Required("expires_at", "body", m.ExpiresAt); err != nil {
		return err
	}

	if err := validate.FormatOf("expires_at", "body", "date-time", m.ExpiresAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *VerifiableAddress) validateID(formats strfmt.Registry) error {

	if err := m.ID.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("id")
		}
		return err
	}

	return nil
}

func (m *VerifiableAddress) validateValue(formats strfmt.Registry) error {

	if err := validate.Required("value", "body", m.Value); err != nil {
		return err
	}

	return nil
}

func (m *VerifiableAddress) validateVerified(formats strfmt.Registry) error {

	if err := validate.Required("verified", "body", m.Verified); err != nil {
		return err
	}

	return nil
}

func (m *VerifiableAddress) validateVerifiedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.VerifiedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("verified_at", "body", "date-time", m.VerifiedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *VerifiableAddress) validateVia(formats strfmt.Registry) error {

	if err := m.Via.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("via")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *VerifiableAddress) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VerifiableAddress) UnmarshalBinary(b []byte) error {
	var res VerifiableAddress
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
