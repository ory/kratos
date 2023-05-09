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

// ContinueWithVerificationUiAllOf struct for ContinueWithVerificationUiAllOf
type ContinueWithVerificationUiAllOf struct {
	Flow *ContinueWithVerificationUiAllOfFlow `json:"flow,omitempty"`
}

// NewContinueWithVerificationUiAllOf instantiates a new ContinueWithVerificationUiAllOf object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewContinueWithVerificationUiAllOf() *ContinueWithVerificationUiAllOf {
	this := ContinueWithVerificationUiAllOf{}
	return &this
}

// NewContinueWithVerificationUiAllOfWithDefaults instantiates a new ContinueWithVerificationUiAllOf object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewContinueWithVerificationUiAllOfWithDefaults() *ContinueWithVerificationUiAllOf {
	this := ContinueWithVerificationUiAllOf{}
	return &this
}

// GetFlow returns the Flow field value if set, zero value otherwise.
func (o *ContinueWithVerificationUiAllOf) GetFlow() ContinueWithVerificationUiAllOfFlow {
	if o == nil || o.Flow == nil {
		var ret ContinueWithVerificationUiAllOfFlow
		return ret
	}
	return *o.Flow
}

// GetFlowOk returns a tuple with the Flow field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ContinueWithVerificationUiAllOf) GetFlowOk() (*ContinueWithVerificationUiAllOfFlow, bool) {
	if o == nil || o.Flow == nil {
		return nil, false
	}
	return o.Flow, true
}

// HasFlow returns a boolean if a field has been set.
func (o *ContinueWithVerificationUiAllOf) HasFlow() bool {
	if o != nil && o.Flow != nil {
		return true
	}

	return false
}

// SetFlow gets a reference to the given ContinueWithVerificationUiAllOfFlow and assigns it to the Flow field.
func (o *ContinueWithVerificationUiAllOf) SetFlow(v ContinueWithVerificationUiAllOfFlow) {
	o.Flow = &v
}

func (o ContinueWithVerificationUiAllOf) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Flow != nil {
		toSerialize["flow"] = o.Flow
	}
	return json.Marshal(toSerialize)
}

type NullableContinueWithVerificationUiAllOf struct {
	value *ContinueWithVerificationUiAllOf
	isSet bool
}

func (v NullableContinueWithVerificationUiAllOf) Get() *ContinueWithVerificationUiAllOf {
	return v.value
}

func (v *NullableContinueWithVerificationUiAllOf) Set(val *ContinueWithVerificationUiAllOf) {
	v.value = val
	v.isSet = true
}

func (v NullableContinueWithVerificationUiAllOf) IsSet() bool {
	return v.isSet
}

func (v *NullableContinueWithVerificationUiAllOf) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableContinueWithVerificationUiAllOf(val *ContinueWithVerificationUiAllOf) *NullableContinueWithVerificationUiAllOf {
	return &NullableContinueWithVerificationUiAllOf{value: val, isSet: true}
}

func (v NullableContinueWithVerificationUiAllOf) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableContinueWithVerificationUiAllOf) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
