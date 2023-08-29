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

// SelfServiceRegistrationFlowState choose_method: ask the user to choose a method (e.g. registration with email) sent_email: the email has been sent to the user passed_challenge: the request was successful and the registration challenge was passed.
type SelfServiceRegistrationFlowState string

// List of SelfServiceRegistrationFlowState
const (
	SELFSERVICEREGISTRATIONFLOWSTATE_CHOOSE_METHOD    SelfServiceRegistrationFlowState = "choose_method"
	SELFSERVICEREGISTRATIONFLOWSTATE_SENT_EMAIL       SelfServiceRegistrationFlowState = "sent_email"
	SELFSERVICEREGISTRATIONFLOWSTATE_PASSED_CHALLENGE SelfServiceRegistrationFlowState = "passed_challenge"
)

func (v *SelfServiceRegistrationFlowState) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := SelfServiceRegistrationFlowState(value)
	for _, existing := range []SelfServiceRegistrationFlowState{"choose_method", "sent_email", "passed_challenge"} {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid SelfServiceRegistrationFlowState", value)
}

// Ptr returns reference to SelfServiceRegistrationFlowState value
func (v SelfServiceRegistrationFlowState) Ptr() *SelfServiceRegistrationFlowState {
	return &v
}

type NullableSelfServiceRegistrationFlowState struct {
	value *SelfServiceRegistrationFlowState
	isSet bool
}

func (v NullableSelfServiceRegistrationFlowState) Get() *SelfServiceRegistrationFlowState {
	return v.value
}

func (v *NullableSelfServiceRegistrationFlowState) Set(val *SelfServiceRegistrationFlowState) {
	v.value = val
	v.isSet = true
}

func (v NullableSelfServiceRegistrationFlowState) IsSet() bool {
	return v.isSet
}

func (v *NullableSelfServiceRegistrationFlowState) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSelfServiceRegistrationFlowState(val *SelfServiceRegistrationFlowState) *NullableSelfServiceRegistrationFlowState {
	return &NullableSelfServiceRegistrationFlowState{value: val, isSet: true}
}

func (v NullableSelfServiceRegistrationFlowState) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSelfServiceRegistrationFlowState) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
