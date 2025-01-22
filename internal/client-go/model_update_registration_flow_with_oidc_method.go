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

// UpdateRegistrationFlowWithOidcMethod Update Registration Flow with OpenID Connect Method
type UpdateRegistrationFlowWithOidcMethod struct {
	// The Provider to register with
	Provider string `json:"Provider"`
	// The CSRF Token
	CsrfToken *string `json:"csrf_token,omitempty"`
	// IDToken is an optional id token provided by an OIDC Provider  If submitted, it is verified using the OIDC Provider's public key set and the claims are used to populate the OIDC credentials of the identity. If the OIDC Provider does not store additional claims (such as name, etc.) in the IDToken itself, you can use the `traits` field to populate the identity's traits. Note, that Apple only includes the users email in the IDToken.  Supported providers are Apple Google
	IdToken *string `json:"id_token,omitempty"`
	// IDTokenNonce is the nonce, used when generating the IDToken. If the Provider supports nonce validation, the nonce will be validated against this value and is required.
	IdTokenNonce *string `json:"id_token_nonce,omitempty"`
	// Method to use  This field must be set to `oidc` when using the oidc method.
	Method string `json:"method"`
	// The identity traits
	Traits map[string]interface{} `json:"traits,omitempty"`
	// Transient data to pass along to any webhooks
	TransientPayload map[string]interface{} `json:"transient_payload,omitempty"`
	// UpstreamParameters are the parameters that are passed to the upstream identity Provider.  These parameters are optional and depend on what the upstream identity Provider supports. Supported parameters are: `login_hint` (string): The `login_hint` parameter suppresses the account chooser and either pre-fills the email box on the sign-in form, or selects the proper session. `hd` (string): The `hd` parameter limits the login/registration process to a Google Organization, e.g. `mycollege.edu`. `prompt` (string): The `prompt` specifies whether the Authorization Server prompts the End-User for reauthentication and consent, e.g. `select_account`.
	UpstreamParameters map[string]interface{} `json:"upstream_parameters,omitempty"`
}

// NewUpdateRegistrationFlowWithOidcMethod instantiates a new UpdateRegistrationFlowWithOidcMethod object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateRegistrationFlowWithOidcMethod(provider string, method string) *UpdateRegistrationFlowWithOidcMethod {
	this := UpdateRegistrationFlowWithOidcMethod{}
	this.Provider = provider
	this.Method = method
	return &this
}

// NewUpdateRegistrationFlowWithOidcMethodWithDefaults instantiates a new UpdateRegistrationFlowWithOidcMethod object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateRegistrationFlowWithOidcMethodWithDefaults() *UpdateRegistrationFlowWithOidcMethod {
	this := UpdateRegistrationFlowWithOidcMethod{}
	return &this
}

// GetProvider returns the Provider field value
func (o *UpdateRegistrationFlowWithOidcMethod) GetProvider() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Provider
}

// GetProviderOk returns a tuple with the Provider field value
// and a boolean to check if the value has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) GetProviderOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Provider, true
}

// SetProvider sets field value
func (o *UpdateRegistrationFlowWithOidcMethod) SetProvider(v string) {
	o.Provider = v
}

// GetCsrfToken returns the CsrfToken field value if set, zero value otherwise.
func (o *UpdateRegistrationFlowWithOidcMethod) GetCsrfToken() string {
	if o == nil || o.CsrfToken == nil {
		var ret string
		return ret
	}
	return *o.CsrfToken
}

// GetCsrfTokenOk returns a tuple with the CsrfToken field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) GetCsrfTokenOk() (*string, bool) {
	if o == nil || o.CsrfToken == nil {
		return nil, false
	}
	return o.CsrfToken, true
}

// HasCsrfToken returns a boolean if a field has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) HasCsrfToken() bool {
	if o != nil && o.CsrfToken != nil {
		return true
	}

	return false
}

// SetCsrfToken gets a reference to the given string and assigns it to the CsrfToken field.
func (o *UpdateRegistrationFlowWithOidcMethod) SetCsrfToken(v string) {
	o.CsrfToken = &v
}

// GetIdToken returns the IdToken field value if set, zero value otherwise.
func (o *UpdateRegistrationFlowWithOidcMethod) GetIdToken() string {
	if o == nil || o.IdToken == nil {
		var ret string
		return ret
	}
	return *o.IdToken
}

// GetIdTokenOk returns a tuple with the IdToken field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) GetIdTokenOk() (*string, bool) {
	if o == nil || o.IdToken == nil {
		return nil, false
	}
	return o.IdToken, true
}

// HasIdToken returns a boolean if a field has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) HasIdToken() bool {
	if o != nil && o.IdToken != nil {
		return true
	}

	return false
}

// SetIdToken gets a reference to the given string and assigns it to the IdToken field.
func (o *UpdateRegistrationFlowWithOidcMethod) SetIdToken(v string) {
	o.IdToken = &v
}

// GetIdTokenNonce returns the IdTokenNonce field value if set, zero value otherwise.
func (o *UpdateRegistrationFlowWithOidcMethod) GetIdTokenNonce() string {
	if o == nil || o.IdTokenNonce == nil {
		var ret string
		return ret
	}
	return *o.IdTokenNonce
}

// GetIdTokenNonceOk returns a tuple with the IdTokenNonce field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) GetIdTokenNonceOk() (*string, bool) {
	if o == nil || o.IdTokenNonce == nil {
		return nil, false
	}
	return o.IdTokenNonce, true
}

// HasIdTokenNonce returns a boolean if a field has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) HasIdTokenNonce() bool {
	if o != nil && o.IdTokenNonce != nil {
		return true
	}

	return false
}

// SetIdTokenNonce gets a reference to the given string and assigns it to the IdTokenNonce field.
func (o *UpdateRegistrationFlowWithOidcMethod) SetIdTokenNonce(v string) {
	o.IdTokenNonce = &v
}

// GetMethod returns the Method field value
func (o *UpdateRegistrationFlowWithOidcMethod) GetMethod() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Method
}

// GetMethodOk returns a tuple with the Method field value
// and a boolean to check if the value has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) GetMethodOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Method, true
}

// SetMethod sets field value
func (o *UpdateRegistrationFlowWithOidcMethod) SetMethod(v string) {
	o.Method = v
}

// GetTraits returns the Traits field value if set, zero value otherwise.
func (o *UpdateRegistrationFlowWithOidcMethod) GetTraits() map[string]interface{} {
	if o == nil || o.Traits == nil {
		var ret map[string]interface{}
		return ret
	}
	return o.Traits
}

// GetTraitsOk returns a tuple with the Traits field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) GetTraitsOk() (map[string]interface{}, bool) {
	if o == nil || o.Traits == nil {
		return nil, false
	}
	return o.Traits, true
}

// HasTraits returns a boolean if a field has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) HasTraits() bool {
	if o != nil && o.Traits != nil {
		return true
	}

	return false
}

// SetTraits gets a reference to the given map[string]interface{} and assigns it to the Traits field.
func (o *UpdateRegistrationFlowWithOidcMethod) SetTraits(v map[string]interface{}) {
	o.Traits = v
}

// GetTransientPayload returns the TransientPayload field value if set, zero value otherwise.
func (o *UpdateRegistrationFlowWithOidcMethod) GetTransientPayload() map[string]interface{} {
	if o == nil || o.TransientPayload == nil {
		var ret map[string]interface{}
		return ret
	}
	return o.TransientPayload
}

// GetTransientPayloadOk returns a tuple with the TransientPayload field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) GetTransientPayloadOk() (map[string]interface{}, bool) {
	if o == nil || o.TransientPayload == nil {
		return nil, false
	}
	return o.TransientPayload, true
}

// HasTransientPayload returns a boolean if a field has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) HasTransientPayload() bool {
	if o != nil && o.TransientPayload != nil {
		return true
	}

	return false
}

// SetTransientPayload gets a reference to the given map[string]interface{} and assigns it to the TransientPayload field.
func (o *UpdateRegistrationFlowWithOidcMethod) SetTransientPayload(v map[string]interface{}) {
	o.TransientPayload = v
}

// GetUpstreamParameters returns the UpstreamParameters field value if set, zero value otherwise.
func (o *UpdateRegistrationFlowWithOidcMethod) GetUpstreamParameters() map[string]interface{} {
	if o == nil || o.UpstreamParameters == nil {
		var ret map[string]interface{}
		return ret
	}
	return o.UpstreamParameters
}

// GetUpstreamParametersOk returns a tuple with the UpstreamParameters field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) GetUpstreamParametersOk() (map[string]interface{}, bool) {
	if o == nil || o.UpstreamParameters == nil {
		return nil, false
	}
	return o.UpstreamParameters, true
}

// HasUpstreamParameters returns a boolean if a field has been set.
func (o *UpdateRegistrationFlowWithOidcMethod) HasUpstreamParameters() bool {
	if o != nil && o.UpstreamParameters != nil {
		return true
	}

	return false
}

// SetUpstreamParameters gets a reference to the given map[string]interface{} and assigns it to the UpstreamParameters field.
func (o *UpdateRegistrationFlowWithOidcMethod) SetUpstreamParameters(v map[string]interface{}) {
	o.UpstreamParameters = v
}

func (o UpdateRegistrationFlowWithOidcMethod) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["Provider"] = o.Provider
	}
	if o.CsrfToken != nil {
		toSerialize["csrf_token"] = o.CsrfToken
	}
	if o.IdToken != nil {
		toSerialize["id_token"] = o.IdToken
	}
	if o.IdTokenNonce != nil {
		toSerialize["id_token_nonce"] = o.IdTokenNonce
	}
	if true {
		toSerialize["method"] = o.Method
	}
	if o.Traits != nil {
		toSerialize["traits"] = o.Traits
	}
	if o.TransientPayload != nil {
		toSerialize["transient_payload"] = o.TransientPayload
	}
	if o.UpstreamParameters != nil {
		toSerialize["upstream_parameters"] = o.UpstreamParameters
	}
	return json.Marshal(toSerialize)
}

type NullableUpdateRegistrationFlowWithOidcMethod struct {
	value *UpdateRegistrationFlowWithOidcMethod
	isSet bool
}

func (v NullableUpdateRegistrationFlowWithOidcMethod) Get() *UpdateRegistrationFlowWithOidcMethod {
	return v.value
}

func (v *NullableUpdateRegistrationFlowWithOidcMethod) Set(val *UpdateRegistrationFlowWithOidcMethod) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateRegistrationFlowWithOidcMethod) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateRegistrationFlowWithOidcMethod) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateRegistrationFlowWithOidcMethod(val *UpdateRegistrationFlowWithOidcMethod) *NullableUpdateRegistrationFlowWithOidcMethod {
	return &NullableUpdateRegistrationFlowWithOidcMethod{value: val, isSet: true}
}

func (v NullableUpdateRegistrationFlowWithOidcMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateRegistrationFlowWithOidcMethod) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
