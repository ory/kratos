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

// ContinueWithRecoveryUi Indicates, that the UI flow could be continued by showing a recovery ui
type ContinueWithRecoveryUi struct {
	// Action will always be `show_recovery_ui` show_recovery_ui ContinueWithActionShowRecoveryUIString
	Action string                     `json:"action"`
	Flow   ContinueWithRecoveryUiFlow `json:"flow"`
}

// NewContinueWithRecoveryUi instantiates a new ContinueWithRecoveryUi object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewContinueWithRecoveryUi(action string, flow ContinueWithRecoveryUiFlow) *ContinueWithRecoveryUi {
	this := ContinueWithRecoveryUi{}
	this.Action = action
	this.Flow = flow
	return &this
}

// NewContinueWithRecoveryUiWithDefaults instantiates a new ContinueWithRecoveryUi object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewContinueWithRecoveryUiWithDefaults() *ContinueWithRecoveryUi {
	this := ContinueWithRecoveryUi{}
	return &this
}

// GetAction returns the Action field value
func (o *ContinueWithRecoveryUi) GetAction() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Action
}

// GetActionOk returns a tuple with the Action field value
// and a boolean to check if the value has been set.
func (o *ContinueWithRecoveryUi) GetActionOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Action, true
}

// SetAction sets field value
func (o *ContinueWithRecoveryUi) SetAction(v string) {
	o.Action = v
}

// GetFlow returns the Flow field value
func (o *ContinueWithRecoveryUi) GetFlow() ContinueWithRecoveryUiFlow {
	if o == nil {
		var ret ContinueWithRecoveryUiFlow
		return ret
	}

	return o.Flow
}

// GetFlowOk returns a tuple with the Flow field value
// and a boolean to check if the value has been set.
func (o *ContinueWithRecoveryUi) GetFlowOk() (*ContinueWithRecoveryUiFlow, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Flow, true
}

// SetFlow sets field value
func (o *ContinueWithRecoveryUi) SetFlow(v ContinueWithRecoveryUiFlow) {
	o.Flow = v
}

func (o ContinueWithRecoveryUi) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["action"] = o.Action
	}
	if true {
		toSerialize["flow"] = o.Flow
	}
	return json.Marshal(toSerialize)
}

type NullableContinueWithRecoveryUi struct {
	value *ContinueWithRecoveryUi
	isSet bool
}

func (v NullableContinueWithRecoveryUi) Get() *ContinueWithRecoveryUi {
	return v.value
}

func (v *NullableContinueWithRecoveryUi) Set(val *ContinueWithRecoveryUi) {
	v.value = val
	v.isSet = true
}

func (v NullableContinueWithRecoveryUi) IsSet() bool {
	return v.isSet
}

func (v *NullableContinueWithRecoveryUi) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableContinueWithRecoveryUi(val *ContinueWithRecoveryUi) *NullableContinueWithRecoveryUi {
	return &NullableContinueWithRecoveryUi{value: val, isSet: true}
}

func (v NullableContinueWithRecoveryUi) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableContinueWithRecoveryUi) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
