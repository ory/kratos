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
	"fmt"
)

// SelfServiceLoginFlowState The state represents the state of the login flow.  choose_method: ask the user to choose a method (e.g. login account via email) sent_email: the email has been sent to the user passed_challenge: the request was successful and the login challenge was passed.
type SelfServiceLoginFlowState string

// List of SelfServiceLoginFlowState
const (
	SELFSERVICELOGINFLOWSTATE_CHOOSE_METHOD    SelfServiceLoginFlowState = "choose_method"
	SELFSERVICELOGINFLOWSTATE_SENT_EMAIL       SelfServiceLoginFlowState = "sent_email"
	SELFSERVICELOGINFLOWSTATE_PASSED_CHALLENGE SelfServiceLoginFlowState = "passed_challenge"
)

func (v *SelfServiceLoginFlowState) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := SelfServiceLoginFlowState(value)
	for _, existing := range []SelfServiceLoginFlowState{"choose_method", "sent_email", "passed_challenge"} {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid SelfServiceLoginFlowState", value)
}

// Ptr returns reference to SelfServiceLoginFlowState value
func (v SelfServiceLoginFlowState) Ptr() *SelfServiceLoginFlowState {
	return &v
}

type NullableSelfServiceLoginFlowState struct {
	value *SelfServiceLoginFlowState
	isSet bool
}

func (v NullableSelfServiceLoginFlowState) Get() *SelfServiceLoginFlowState {
	return v.value
}

func (v *NullableSelfServiceLoginFlowState) Set(val *SelfServiceLoginFlowState) {
	v.value = val
	v.isSet = true
}

func (v NullableSelfServiceLoginFlowState) IsSet() bool {
	return v.isSet
}

func (v *NullableSelfServiceLoginFlowState) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSelfServiceLoginFlowState(val *SelfServiceLoginFlowState) *NullableSelfServiceLoginFlowState {
	return &NullableSelfServiceLoginFlowState{value: val, isSet: true}
}

func (v NullableSelfServiceLoginFlowState) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSelfServiceLoginFlowState) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
