/*
Ory Identities API

This is the API specification for Ory Identities with features such as registration, login, recovery, account verification, profile settings, password reset, identity management, session management, email and sms delivery, and more.

API version:
Contact: office@ory.sh
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
	"fmt"
)

// checks if the UpdateLoginFlowWithCodeMethod type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateLoginFlowWithCodeMethod{}

// UpdateLoginFlowWithCodeMethod Update Login flow using the code method
type UpdateLoginFlowWithCodeMethod struct {
	// Address is the address to send the code to, in case that there are multiple addresses. This field is only used in two-factor flows and is ineffective for passwordless flows.
	Address *string `json:"address,omitempty"`
	// Code is the 6 digits code sent to the user
	Code *string `json:"code,omitempty"`
	// CSRFToken is the anti-CSRF token
	CsrfToken string `json:"csrf_token"`
	// Identifier is the code identifier The identifier requires that the user has already completed the registration or settings with code flow.
	Identifier *string `json:"identifier,omitempty"`
	// Method should be set to \"code\" when logging in using the code strategy.
	Method string `json:"method"`
	// Resend is set when the user wants to resend the code
	Resend *string `json:"resend,omitempty"`
	// Transient data to pass along to any webhooks
	TransientPayload     map[string]interface{} `json:"transient_payload,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _UpdateLoginFlowWithCodeMethod UpdateLoginFlowWithCodeMethod

// NewUpdateLoginFlowWithCodeMethod instantiates a new UpdateLoginFlowWithCodeMethod object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateLoginFlowWithCodeMethod(csrfToken string, method string) *UpdateLoginFlowWithCodeMethod {
	this := UpdateLoginFlowWithCodeMethod{}
	this.CsrfToken = csrfToken
	this.Method = method
	return &this
}

// NewUpdateLoginFlowWithCodeMethodWithDefaults instantiates a new UpdateLoginFlowWithCodeMethod object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateLoginFlowWithCodeMethodWithDefaults() *UpdateLoginFlowWithCodeMethod {
	this := UpdateLoginFlowWithCodeMethod{}
	return &this
}

// GetAddress returns the Address field value if set, zero value otherwise.
func (o *UpdateLoginFlowWithCodeMethod) GetAddress() string {
	if o == nil || IsNil(o.Address) {
		var ret string
		return ret
	}
	return *o.Address
}

// GetAddressOk returns a tuple with the Address field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithCodeMethod) GetAddressOk() (*string, bool) {
	if o == nil || IsNil(o.Address) {
		return nil, false
	}
	return o.Address, true
}

// HasAddress returns a boolean if a field has been set.
func (o *UpdateLoginFlowWithCodeMethod) HasAddress() bool {
	if o != nil && !IsNil(o.Address) {
		return true
	}

	return false
}

// SetAddress gets a reference to the given string and assigns it to the Address field.
func (o *UpdateLoginFlowWithCodeMethod) SetAddress(v string) {
	o.Address = &v
}

// GetCode returns the Code field value if set, zero value otherwise.
func (o *UpdateLoginFlowWithCodeMethod) GetCode() string {
	if o == nil || IsNil(o.Code) {
		var ret string
		return ret
	}
	return *o.Code
}

// GetCodeOk returns a tuple with the Code field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithCodeMethod) GetCodeOk() (*string, bool) {
	if o == nil || IsNil(o.Code) {
		return nil, false
	}
	return o.Code, true
}

// HasCode returns a boolean if a field has been set.
func (o *UpdateLoginFlowWithCodeMethod) HasCode() bool {
	if o != nil && !IsNil(o.Code) {
		return true
	}

	return false
}

// SetCode gets a reference to the given string and assigns it to the Code field.
func (o *UpdateLoginFlowWithCodeMethod) SetCode(v string) {
	o.Code = &v
}

// GetCsrfToken returns the CsrfToken field value
func (o *UpdateLoginFlowWithCodeMethod) GetCsrfToken() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.CsrfToken
}

// GetCsrfTokenOk returns a tuple with the CsrfToken field value
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithCodeMethod) GetCsrfTokenOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CsrfToken, true
}

// SetCsrfToken sets field value
func (o *UpdateLoginFlowWithCodeMethod) SetCsrfToken(v string) {
	o.CsrfToken = v
}

// GetIdentifier returns the Identifier field value if set, zero value otherwise.
func (o *UpdateLoginFlowWithCodeMethod) GetIdentifier() string {
	if o == nil || IsNil(o.Identifier) {
		var ret string
		return ret
	}
	return *o.Identifier
}

// GetIdentifierOk returns a tuple with the Identifier field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithCodeMethod) GetIdentifierOk() (*string, bool) {
	if o == nil || IsNil(o.Identifier) {
		return nil, false
	}
	return o.Identifier, true
}

// HasIdentifier returns a boolean if a field has been set.
func (o *UpdateLoginFlowWithCodeMethod) HasIdentifier() bool {
	if o != nil && !IsNil(o.Identifier) {
		return true
	}

	return false
}

// SetIdentifier gets a reference to the given string and assigns it to the Identifier field.
func (o *UpdateLoginFlowWithCodeMethod) SetIdentifier(v string) {
	o.Identifier = &v
}

// GetMethod returns the Method field value
func (o *UpdateLoginFlowWithCodeMethod) GetMethod() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Method
}

// GetMethodOk returns a tuple with the Method field value
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithCodeMethod) GetMethodOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Method, true
}

// SetMethod sets field value
func (o *UpdateLoginFlowWithCodeMethod) SetMethod(v string) {
	o.Method = v
}

// GetResend returns the Resend field value if set, zero value otherwise.
func (o *UpdateLoginFlowWithCodeMethod) GetResend() string {
	if o == nil || IsNil(o.Resend) {
		var ret string
		return ret
	}
	return *o.Resend
}

// GetResendOk returns a tuple with the Resend field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithCodeMethod) GetResendOk() (*string, bool) {
	if o == nil || IsNil(o.Resend) {
		return nil, false
	}
	return o.Resend, true
}

// HasResend returns a boolean if a field has been set.
func (o *UpdateLoginFlowWithCodeMethod) HasResend() bool {
	if o != nil && !IsNil(o.Resend) {
		return true
	}

	return false
}

// SetResend gets a reference to the given string and assigns it to the Resend field.
func (o *UpdateLoginFlowWithCodeMethod) SetResend(v string) {
	o.Resend = &v
}

// GetTransientPayload returns the TransientPayload field value if set, zero value otherwise.
func (o *UpdateLoginFlowWithCodeMethod) GetTransientPayload() map[string]interface{} {
	if o == nil || IsNil(o.TransientPayload) {
		var ret map[string]interface{}
		return ret
	}
	return o.TransientPayload
}

// GetTransientPayloadOk returns a tuple with the TransientPayload field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateLoginFlowWithCodeMethod) GetTransientPayloadOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.TransientPayload) {
		return map[string]interface{}{}, false
	}
	return o.TransientPayload, true
}

// HasTransientPayload returns a boolean if a field has been set.
func (o *UpdateLoginFlowWithCodeMethod) HasTransientPayload() bool {
	if o != nil && !IsNil(o.TransientPayload) {
		return true
	}

	return false
}

// SetTransientPayload gets a reference to the given map[string]interface{} and assigns it to the TransientPayload field.
func (o *UpdateLoginFlowWithCodeMethod) SetTransientPayload(v map[string]interface{}) {
	o.TransientPayload = v
}

func (o UpdateLoginFlowWithCodeMethod) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateLoginFlowWithCodeMethod) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Address) {
		toSerialize["address"] = o.Address
	}
	if !IsNil(o.Code) {
		toSerialize["code"] = o.Code
	}
	toSerialize["csrf_token"] = o.CsrfToken
	if !IsNil(o.Identifier) {
		toSerialize["identifier"] = o.Identifier
	}
	toSerialize["method"] = o.Method
	if !IsNil(o.Resend) {
		toSerialize["resend"] = o.Resend
	}
	if !IsNil(o.TransientPayload) {
		toSerialize["transient_payload"] = o.TransientPayload
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *UpdateLoginFlowWithCodeMethod) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"csrf_token",
		"method",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err
	}

	for _, requiredProperty := range requiredProperties {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varUpdateLoginFlowWithCodeMethod := _UpdateLoginFlowWithCodeMethod{}

	err = json.Unmarshal(data, &varUpdateLoginFlowWithCodeMethod)

	if err != nil {
		return err
	}

	*o = UpdateLoginFlowWithCodeMethod(varUpdateLoginFlowWithCodeMethod)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "address")
		delete(additionalProperties, "code")
		delete(additionalProperties, "csrf_token")
		delete(additionalProperties, "identifier")
		delete(additionalProperties, "method")
		delete(additionalProperties, "resend")
		delete(additionalProperties, "transient_payload")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableUpdateLoginFlowWithCodeMethod struct {
	value *UpdateLoginFlowWithCodeMethod
	isSet bool
}

func (v NullableUpdateLoginFlowWithCodeMethod) Get() *UpdateLoginFlowWithCodeMethod {
	return v.value
}

func (v *NullableUpdateLoginFlowWithCodeMethod) Set(val *UpdateLoginFlowWithCodeMethod) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateLoginFlowWithCodeMethod) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateLoginFlowWithCodeMethod) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateLoginFlowWithCodeMethod(val *UpdateLoginFlowWithCodeMethod) *NullableUpdateLoginFlowWithCodeMethod {
	return &NullableUpdateLoginFlowWithCodeMethod{value: val, isSet: true}
}

func (v NullableUpdateLoginFlowWithCodeMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateLoginFlowWithCodeMethod) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
