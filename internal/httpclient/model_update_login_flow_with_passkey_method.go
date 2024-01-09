/*
 * Ory Identities API
 *
 * This is the API specification for Ory Identities with features such as registration, login, recovery, account verification, profile settings, password reset, identity management, session management, email and sms delivery, and more.
 *
 * API version:
 * Contact: office@ory.sh
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
)

// UpdateLoginFlowWithPasskeyMethod Update Login Flow with Passkey Method
type UpdateLoginFlowWithPasskeyMethod struct {
	// Sending the anti-csrf token is only required for browser login flows.
	CsrfToken *string `json:"csrf_token,omitempty"`
	// Method should be set to \"passkey\" when logging in using the Passkey strategy.
	Method string `json:"method"`
	// Login a WebAuthn Security Key  This must contain the ID of the WebAuthN connection.
	PasskeyLogin *string `json:"passkey_login,omitempty"`
}

// NewUpdateLoginFlowWithPasskeyMethod instantiates a new UpdateLoginFlowWithPasskeyMethod object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateLoginFlowWithPasskeyMethod(method string) *UpdateLoginFlowWithPasskeyMethod {
	this := UpdateLoginFlowWithPasskeyMethod{}
	this.Method = method
	return &this
}

// NewUpdateLoginFlowWithPasskeyMethodWithDefaults instantiates a new UpdateLoginFlowWithPasskeyMethod object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateLoginFlowWithPasskeyMethodWithDefaults() *UpdateLoginFlowWithPasskeyMethod {
	this := UpdateLoginFlowWithPasskeyMethod{}
	return &this
}

// GetCsrfToken returns the CsrfToken field value if set, zero value otherwise.
func (o *UpdateLoginFlowWithPasskeyMethod) GetCsrfToken() string {
	if o == nil || o.CsrfToken == nil {
		var ret string
		return ret
	}
	return *o.CsrfToken
}

// GetCsrfTokenOk returns a tuple with the CsrfToken field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithPasskeyMethod) GetCsrfTokenOk() (*string, bool) {
	if o == nil || o.CsrfToken == nil {
		return nil, false
	}
	return o.CsrfToken, true
}

// HasCsrfToken returns a boolean if a field has been set.
func (o *UpdateLoginFlowWithPasskeyMethod) HasCsrfToken() bool {
	if o != nil && o.CsrfToken != nil {
		return true
	}

	return false
}

// SetCsrfToken gets a reference to the given string and assigns it to the CsrfToken field.
func (o *UpdateLoginFlowWithPasskeyMethod) SetCsrfToken(v string) {
	o.CsrfToken = &v
}

// GetMethod returns the Method field value
func (o *UpdateLoginFlowWithPasskeyMethod) GetMethod() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Method
}

// GetMethodOk returns a tuple with the Method field value
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithPasskeyMethod) GetMethodOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Method, true
}

// SetMethod sets field value
func (o *UpdateLoginFlowWithPasskeyMethod) SetMethod(v string) {
	o.Method = v
}

// GetPasskeyLogin returns the PasskeyLogin field value if set, zero value otherwise.
func (o *UpdateLoginFlowWithPasskeyMethod) GetPasskeyLogin() string {
	if o == nil || o.PasskeyLogin == nil {
		var ret string
		return ret
	}
	return *o.PasskeyLogin
}

// GetPasskeyLoginOk returns a tuple with the PasskeyLogin field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithPasskeyMethod) GetPasskeyLoginOk() (*string, bool) {
	if o == nil || o.PasskeyLogin == nil {
		return nil, false
	}
	return o.PasskeyLogin, true
}

// HasPasskeyLogin returns a boolean if a field has been set.
func (o *UpdateLoginFlowWithPasskeyMethod) HasPasskeyLogin() bool {
	if o != nil && o.PasskeyLogin != nil {
		return true
	}

	return false
}

// SetPasskeyLogin gets a reference to the given string and assigns it to the PasskeyLogin field.
func (o *UpdateLoginFlowWithPasskeyMethod) SetPasskeyLogin(v string) {
	o.PasskeyLogin = &v
}

func (o UpdateLoginFlowWithPasskeyMethod) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.CsrfToken != nil {
		toSerialize["csrf_token"] = o.CsrfToken
	}
	if true {
		toSerialize["method"] = o.Method
	}
	if o.PasskeyLogin != nil {
		toSerialize["passkey_login"] = o.PasskeyLogin
	}
	return json.Marshal(toSerialize)
}

type NullableUpdateLoginFlowWithPasskeyMethod struct {
	value *UpdateLoginFlowWithPasskeyMethod
	isSet bool
}

func (v NullableUpdateLoginFlowWithPasskeyMethod) Get() *UpdateLoginFlowWithPasskeyMethod {
	return v.value
}

func (v *NullableUpdateLoginFlowWithPasskeyMethod) Set(val *UpdateLoginFlowWithPasskeyMethod) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateLoginFlowWithPasskeyMethod) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateLoginFlowWithPasskeyMethod) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateLoginFlowWithPasskeyMethod(val *UpdateLoginFlowWithPasskeyMethod) *NullableUpdateLoginFlowWithPasskeyMethod {
	return &NullableUpdateLoginFlowWithPasskeyMethod{value: val, isSet: true}
}

func (v NullableUpdateLoginFlowWithPasskeyMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateLoginFlowWithPasskeyMethod) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
