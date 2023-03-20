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

// ContinueWithFlowView struct for ContinueWithFlowView
type ContinueWithFlowView struct {
	Class *string `json:"class,omitempty"`
	Id    *string `json:"id,omitempty"`
}

// NewContinueWithFlowView instantiates a new ContinueWithFlowView object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewContinueWithFlowView() *ContinueWithFlowView {
	this := ContinueWithFlowView{}
	return &this
}

// NewContinueWithFlowViewWithDefaults instantiates a new ContinueWithFlowView object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewContinueWithFlowViewWithDefaults() *ContinueWithFlowView {
	this := ContinueWithFlowView{}
	return &this
}

// GetClass returns the Class field value if set, zero value otherwise.
func (o *ContinueWithFlowView) GetClass() string {
	if o == nil || o.Class == nil {
		var ret string
		return ret
	}
	return *o.Class
}

// GetClassOk returns a tuple with the Class field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ContinueWithFlowView) GetClassOk() (*string, bool) {
	if o == nil || o.Class == nil {
		return nil, false
	}
	return o.Class, true
}

// HasClass returns a boolean if a field has been set.
func (o *ContinueWithFlowView) HasClass() bool {
	if o != nil && o.Class != nil {
		return true
	}

	return false
}

// SetClass gets a reference to the given string and assigns it to the Class field.
func (o *ContinueWithFlowView) SetClass(v string) {
	o.Class = &v
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *ContinueWithFlowView) GetId() string {
	if o == nil || o.Id == nil {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ContinueWithFlowView) GetIdOk() (*string, bool) {
	if o == nil || o.Id == nil {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *ContinueWithFlowView) HasId() bool {
	if o != nil && o.Id != nil {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *ContinueWithFlowView) SetId(v string) {
	o.Id = &v
}

func (o ContinueWithFlowView) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Class != nil {
		toSerialize["class"] = o.Class
	}
	if o.Id != nil {
		toSerialize["id"] = o.Id
	}
	return json.Marshal(toSerialize)
}

type NullableContinueWithFlowView struct {
	value *ContinueWithFlowView
	isSet bool
}

func (v NullableContinueWithFlowView) Get() *ContinueWithFlowView {
	return v.value
}

func (v *NullableContinueWithFlowView) Set(val *ContinueWithFlowView) {
	v.value = val
	v.isSet = true
}

func (v NullableContinueWithFlowView) IsSet() bool {
	return v.isSet
}

func (v *NullableContinueWithFlowView) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableContinueWithFlowView(val *ContinueWithFlowView) *NullableContinueWithFlowView {
	return &NullableContinueWithFlowView{value: val, isSet: true}
}

func (v NullableContinueWithFlowView) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableContinueWithFlowView) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
