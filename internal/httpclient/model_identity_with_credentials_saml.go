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
)

// checks if the IdentityWithCredentialsSaml type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IdentityWithCredentialsSaml{}

// IdentityWithCredentialsSaml Payload to import SAML credentials
type IdentityWithCredentialsSaml struct {
	Config               *IdentityWithCredentialsSamlConfig `json:"config,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _IdentityWithCredentialsSaml IdentityWithCredentialsSaml

// NewIdentityWithCredentialsSaml instantiates a new IdentityWithCredentialsSaml object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIdentityWithCredentialsSaml() *IdentityWithCredentialsSaml {
	this := IdentityWithCredentialsSaml{}
	return &this
}

// NewIdentityWithCredentialsSamlWithDefaults instantiates a new IdentityWithCredentialsSaml object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIdentityWithCredentialsSamlWithDefaults() *IdentityWithCredentialsSaml {
	this := IdentityWithCredentialsSaml{}
	return &this
}

// GetConfig returns the Config field value if set, zero value otherwise.
func (o *IdentityWithCredentialsSaml) GetConfig() IdentityWithCredentialsSamlConfig {
	if o == nil || IsNil(o.Config) {
		var ret IdentityWithCredentialsSamlConfig
		return ret
	}
	return *o.Config
}

// GetConfigOk returns a tuple with the Config field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IdentityWithCredentialsSaml) GetConfigOk() (*IdentityWithCredentialsSamlConfig, bool) {
	if o == nil || IsNil(o.Config) {
		return nil, false
	}
	return o.Config, true
}

// HasConfig returns a boolean if a field has been set.
func (o *IdentityWithCredentialsSaml) HasConfig() bool {
	if o != nil && !IsNil(o.Config) {
		return true
	}

	return false
}

// SetConfig gets a reference to the given IdentityWithCredentialsSamlConfig and assigns it to the Config field.
func (o *IdentityWithCredentialsSaml) SetConfig(v IdentityWithCredentialsSamlConfig) {
	o.Config = &v
}

func (o IdentityWithCredentialsSaml) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IdentityWithCredentialsSaml) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Config) {
		toSerialize["config"] = o.Config
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *IdentityWithCredentialsSaml) UnmarshalJSON(data []byte) (err error) {
	varIdentityWithCredentialsSaml := _IdentityWithCredentialsSaml{}

	err = json.Unmarshal(data, &varIdentityWithCredentialsSaml)

	if err != nil {
		return err
	}

	*o = IdentityWithCredentialsSaml(varIdentityWithCredentialsSaml)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "config")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableIdentityWithCredentialsSaml struct {
	value *IdentityWithCredentialsSaml
	isSet bool
}

func (v NullableIdentityWithCredentialsSaml) Get() *IdentityWithCredentialsSaml {
	return v.value
}

func (v *NullableIdentityWithCredentialsSaml) Set(val *IdentityWithCredentialsSaml) {
	v.value = val
	v.isSet = true
}

func (v NullableIdentityWithCredentialsSaml) IsSet() bool {
	return v.isSet
}

func (v *NullableIdentityWithCredentialsSaml) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIdentityWithCredentialsSaml(val *IdentityWithCredentialsSaml) *NullableIdentityWithCredentialsSaml {
	return &NullableIdentityWithCredentialsSaml{value: val, isSet: true}
}

func (v NullableIdentityWithCredentialsSaml) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIdentityWithCredentialsSaml) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
