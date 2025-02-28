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

// checks if the UpdateRecoveryFlowWithCodeMethod type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateRecoveryFlowWithCodeMethod{}

// UpdateRecoveryFlowWithCodeMethod Update Recovery Flow with Code Method
type UpdateRecoveryFlowWithCodeMethod struct {
	// Code from the recovery email  If you want to submit a code, use this field, but make sure to _not_ include the email field, as well.
	Code *string `json:"code,omitempty"`
	// Sending the anti-csrf token is only required for browser login flows.
	CsrfToken *string `json:"csrf_token,omitempty"`
	// The email address of the account to recover  If the email belongs to a valid account, a recovery email will be sent.  If you want to notify the email address if the account does not exist, see the [notify_unknown_recipients flag](https://www.ory.sh/docs/kratos/self-service/flows/account-recovery-password-reset#attempted-recovery-notifications)  If a code was already sent, including this field in the payload will invalidate the sent code and re-send a new code.  format: email
	Email *string `json:"email,omitempty"`
	// Method is the method that should be used for this recovery flow  Allowed values are `link` and `code`. link RecoveryStrategyLink code RecoveryStrategyCode
	Method string `json:"method"`
	// Transient data to pass along to any webhooks
	TransientPayload     map[string]interface{} `json:"transient_payload,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _UpdateRecoveryFlowWithCodeMethod UpdateRecoveryFlowWithCodeMethod

// NewUpdateRecoveryFlowWithCodeMethod instantiates a new UpdateRecoveryFlowWithCodeMethod object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateRecoveryFlowWithCodeMethod(method string) *UpdateRecoveryFlowWithCodeMethod {
	this := UpdateRecoveryFlowWithCodeMethod{}
	this.Method = method
	return &this
}

// NewUpdateRecoveryFlowWithCodeMethodWithDefaults instantiates a new UpdateRecoveryFlowWithCodeMethod object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateRecoveryFlowWithCodeMethodWithDefaults() *UpdateRecoveryFlowWithCodeMethod {
	this := UpdateRecoveryFlowWithCodeMethod{}
	return &this
}

// GetCode returns the Code field value if set, zero value otherwise.
func (o *UpdateRecoveryFlowWithCodeMethod) GetCode() string {
	if o == nil || IsNil(o.Code) {
		var ret string
		return ret
	}
	return *o.Code
}

// GetCodeOk returns a tuple with the Code field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRecoveryFlowWithCodeMethod) GetCodeOk() (*string, bool) {
	if o == nil || IsNil(o.Code) {
		return nil, false
	}
	return o.Code, true
}

// HasCode returns a boolean if a field has been set.
func (o *UpdateRecoveryFlowWithCodeMethod) HasCode() bool {
	if o != nil && !IsNil(o.Code) {
		return true
	}

	return false
}

// SetCode gets a reference to the given string and assigns it to the Code field.
func (o *UpdateRecoveryFlowWithCodeMethod) SetCode(v string) {
	o.Code = &v
}

// GetCsrfToken returns the CsrfToken field value if set, zero value otherwise.
func (o *UpdateRecoveryFlowWithCodeMethod) GetCsrfToken() string {
	if o == nil || IsNil(o.CsrfToken) {
		var ret string
		return ret
	}
	return *o.CsrfToken
}

// GetCsrfTokenOk returns a tuple with the CsrfToken field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRecoveryFlowWithCodeMethod) GetCsrfTokenOk() (*string, bool) {
	if o == nil || IsNil(o.CsrfToken) {
		return nil, false
	}
	return o.CsrfToken, true
}

// HasCsrfToken returns a boolean if a field has been set.
func (o *UpdateRecoveryFlowWithCodeMethod) HasCsrfToken() bool {
	if o != nil && !IsNil(o.CsrfToken) {
		return true
	}

	return false
}

// SetCsrfToken gets a reference to the given string and assigns it to the CsrfToken field.
func (o *UpdateRecoveryFlowWithCodeMethod) SetCsrfToken(v string) {
	o.CsrfToken = &v
}

// GetEmail returns the Email field value if set, zero value otherwise.
func (o *UpdateRecoveryFlowWithCodeMethod) GetEmail() string {
	if o == nil || IsNil(o.Email) {
		var ret string
		return ret
	}
	return *o.Email
}

// GetEmailOk returns a tuple with the Email field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRecoveryFlowWithCodeMethod) GetEmailOk() (*string, bool) {
	if o == nil || IsNil(o.Email) {
		return nil, false
	}
	return o.Email, true
}

// HasEmail returns a boolean if a field has been set.
func (o *UpdateRecoveryFlowWithCodeMethod) HasEmail() bool {
	if o != nil && !IsNil(o.Email) {
		return true
	}

	return false
}

// SetEmail gets a reference to the given string and assigns it to the Email field.
func (o *UpdateRecoveryFlowWithCodeMethod) SetEmail(v string) {
	o.Email = &v
}

// GetMethod returns the Method field value
func (o *UpdateRecoveryFlowWithCodeMethod) GetMethod() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Method
}

// GetMethodOk returns a tuple with the Method field value
// and a boolean to check if the value has been set.
func (o *UpdateRecoveryFlowWithCodeMethod) GetMethodOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Method, true
}

// SetMethod sets field value
func (o *UpdateRecoveryFlowWithCodeMethod) SetMethod(v string) {
	o.Method = v
}

// GetTransientPayload returns the TransientPayload field value if set, zero value otherwise.
func (o *UpdateRecoveryFlowWithCodeMethod) GetTransientPayload() map[string]interface{} {
	if o == nil || IsNil(o.TransientPayload) {
		var ret map[string]interface{}
		return ret
	}
	return o.TransientPayload
}

// GetTransientPayloadOk returns a tuple with the TransientPayload field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRecoveryFlowWithCodeMethod) GetTransientPayloadOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.TransientPayload) {
		return map[string]interface{}{}, false
	}
	return o.TransientPayload, true
}

// HasTransientPayload returns a boolean if a field has been set.
func (o *UpdateRecoveryFlowWithCodeMethod) HasTransientPayload() bool {
	if o != nil && !IsNil(o.TransientPayload) {
		return true
	}

	return false
}

// SetTransientPayload gets a reference to the given map[string]interface{} and assigns it to the TransientPayload field.
func (o *UpdateRecoveryFlowWithCodeMethod) SetTransientPayload(v map[string]interface{}) {
	o.TransientPayload = v
}

func (o UpdateRecoveryFlowWithCodeMethod) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateRecoveryFlowWithCodeMethod) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Code) {
		toSerialize["code"] = o.Code
	}
	if !IsNil(o.CsrfToken) {
		toSerialize["csrf_token"] = o.CsrfToken
	}
	if !IsNil(o.Email) {
		toSerialize["email"] = o.Email
	}
	toSerialize["method"] = o.Method
	if !IsNil(o.TransientPayload) {
		toSerialize["transient_payload"] = o.TransientPayload
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *UpdateRecoveryFlowWithCodeMethod) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
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

	varUpdateRecoveryFlowWithCodeMethod := _UpdateRecoveryFlowWithCodeMethod{}

	err = json.Unmarshal(data, &varUpdateRecoveryFlowWithCodeMethod)

	if err != nil {
		return err
	}

	*o = UpdateRecoveryFlowWithCodeMethod(varUpdateRecoveryFlowWithCodeMethod)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "code")
		delete(additionalProperties, "csrf_token")
		delete(additionalProperties, "email")
		delete(additionalProperties, "method")
		delete(additionalProperties, "transient_payload")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableUpdateRecoveryFlowWithCodeMethod struct {
	value *UpdateRecoveryFlowWithCodeMethod
	isSet bool
}

func (v NullableUpdateRecoveryFlowWithCodeMethod) Get() *UpdateRecoveryFlowWithCodeMethod {
	return v.value
}

func (v *NullableUpdateRecoveryFlowWithCodeMethod) Set(val *UpdateRecoveryFlowWithCodeMethod) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateRecoveryFlowWithCodeMethod) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateRecoveryFlowWithCodeMethod) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateRecoveryFlowWithCodeMethod(val *UpdateRecoveryFlowWithCodeMethod) *NullableUpdateRecoveryFlowWithCodeMethod {
	return &NullableUpdateRecoveryFlowWithCodeMethod{value: val, isSet: true}
}

func (v NullableUpdateRecoveryFlowWithCodeMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateRecoveryFlowWithCodeMethod) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
