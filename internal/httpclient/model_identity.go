/*
 * Ory Kratos API
 *
 * Documentation for all public and administrative Ory Kratos APIs. Public and administrative APIs are exposed on different ports. Public APIs can face the public internet without any protection while administrative APIs should never be exposed without prior authorization. To protect the administative API port you should use something like Nginx, Ory Oathkeeper, or any other technology capable of authorizing incoming requests.
 *
 * API version: 1.0.0
 * Contact: hi@ory.sh
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
	"time"
)

// Identity An identity can be a real human, a service, an IoT device - everything that can be described as an \"actor\" in a system.
type Identity struct {
	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// Credentials represents all credentials that can be used for authenticating this identity.
	Credentials *map[string]IdentityCredentials `json:"credentials,omitempty"`
	// ID is the identity's unique identifier.  The Identity ID cannot be changed and cannot be chosen. This ensures future compatibility and optimization for distributed stores such as CockroachDB.
	Id string `json:"id"`
	// NullJSONRawMessage represents a json.RawMessage that works well with JSON, SQL, and Swagger and is NULLable-
	MetadataAdmin interface{} `json:"metadata_admin,omitempty"`
	// NullJSONRawMessage represents a json.RawMessage that works well with JSON, SQL, and Swagger and is NULLable-
	MetadataPublic interface{} `json:"metadata_public,omitempty"`
	// RecoveryAddresses contains all the addresses that can be used to recover an identity.
	RecoveryAddresses []RecoveryIdentityAddress `json:"recovery_addresses,omitempty"`
	// SchemaID is the ID of the JSON Schema to be used for validating the identity's traits.
	SchemaId string `json:"schema_id"`
	// SchemaURL is the URL of the endpoint where the identity's traits schema can be fetched from.  format: url
	SchemaUrl      string         `json:"schema_url"`
	State          *IdentityState `json:"state,omitempty"`
	StateChangedAt *time.Time     `json:"state_changed_at,omitempty"`
	// Traits represent an identity's traits. The identity is able to create, modify, and delete traits in a self-service manner. The input will always be validated against the JSON Schema defined in `schema_url`.
	Traits interface{} `json:"traits"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// VerifiableAddresses contains all the addresses that can be verified by the user.
	VerifiableAddresses []VerifiableIdentityAddress `json:"verifiable_addresses,omitempty"`
}

// NewIdentity instantiates a new Identity object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIdentity(id string, schemaId string, schemaUrl string, traits interface{}) *Identity {
	this := Identity{}
	this.Id = id
	this.SchemaId = schemaId
	this.SchemaUrl = schemaUrl
	this.Traits = traits
	return &this
}

// NewIdentityWithDefaults instantiates a new Identity object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIdentityWithDefaults() *Identity {
	this := Identity{}
	return &this
}

// GetCreatedAt returns the CreatedAt field value if set, zero value otherwise.
func (o *Identity) GetCreatedAt() time.Time {
	if o == nil || o.CreatedAt == nil {
		var ret time.Time
		return ret
	}
	return *o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Identity) GetCreatedAtOk() (*time.Time, bool) {
	if o == nil || o.CreatedAt == nil {
		return nil, false
	}
	return o.CreatedAt, true
}

// HasCreatedAt returns a boolean if a field has been set.
func (o *Identity) HasCreatedAt() bool {
	if o != nil && o.CreatedAt != nil {
		return true
	}

	return false
}

// SetCreatedAt gets a reference to the given time.Time and assigns it to the CreatedAt field.
func (o *Identity) SetCreatedAt(v time.Time) {
	o.CreatedAt = &v
}

// GetCredentials returns the Credentials field value if set, zero value otherwise.
func (o *Identity) GetCredentials() map[string]IdentityCredentials {
	if o == nil || o.Credentials == nil {
		var ret map[string]IdentityCredentials
		return ret
	}
	return *o.Credentials
}

// GetCredentialsOk returns a tuple with the Credentials field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Identity) GetCredentialsOk() (*map[string]IdentityCredentials, bool) {
	if o == nil || o.Credentials == nil {
		return nil, false
	}
	return o.Credentials, true
}

// HasCredentials returns a boolean if a field has been set.
func (o *Identity) HasCredentials() bool {
	if o != nil && o.Credentials != nil {
		return true
	}

	return false
}

// SetCredentials gets a reference to the given map[string]IdentityCredentials and assigns it to the Credentials field.
func (o *Identity) SetCredentials(v map[string]IdentityCredentials) {
	o.Credentials = &v
}

// GetId returns the Id field value
func (o *Identity) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *Identity) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *Identity) SetId(v string) {
	o.Id = v
}

// GetMetadataAdmin returns the MetadataAdmin field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *Identity) GetMetadataAdmin() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.MetadataAdmin
}

// GetMetadataAdminOk returns a tuple with the MetadataAdmin field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Identity) GetMetadataAdminOk() (*interface{}, bool) {
	if o == nil || o.MetadataAdmin == nil {
		return nil, false
	}
	return &o.MetadataAdmin, true
}

// HasMetadataAdmin returns a boolean if a field has been set.
func (o *Identity) HasMetadataAdmin() bool {
	if o != nil && o.MetadataAdmin != nil {
		return true
	}

	return false
}

// SetMetadataAdmin gets a reference to the given interface{} and assigns it to the MetadataAdmin field.
func (o *Identity) SetMetadataAdmin(v interface{}) {
	o.MetadataAdmin = v
}

// GetMetadataPublic returns the MetadataPublic field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *Identity) GetMetadataPublic() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.MetadataPublic
}

// GetMetadataPublicOk returns a tuple with the MetadataPublic field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Identity) GetMetadataPublicOk() (*interface{}, bool) {
	if o == nil || o.MetadataPublic == nil {
		return nil, false
	}
	return &o.MetadataPublic, true
}

// HasMetadataPublic returns a boolean if a field has been set.
func (o *Identity) HasMetadataPublic() bool {
	if o != nil && o.MetadataPublic != nil {
		return true
	}

	return false
}

// SetMetadataPublic gets a reference to the given interface{} and assigns it to the MetadataPublic field.
func (o *Identity) SetMetadataPublic(v interface{}) {
	o.MetadataPublic = v
}

// GetRecoveryAddresses returns the RecoveryAddresses field value if set, zero value otherwise.
func (o *Identity) GetRecoveryAddresses() []RecoveryIdentityAddress {
	if o == nil || o.RecoveryAddresses == nil {
		var ret []RecoveryIdentityAddress
		return ret
	}
	return o.RecoveryAddresses
}

// GetRecoveryAddressesOk returns a tuple with the RecoveryAddresses field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Identity) GetRecoveryAddressesOk() ([]RecoveryIdentityAddress, bool) {
	if o == nil || o.RecoveryAddresses == nil {
		return nil, false
	}
	return o.RecoveryAddresses, true
}

// HasRecoveryAddresses returns a boolean if a field has been set.
func (o *Identity) HasRecoveryAddresses() bool {
	if o != nil && o.RecoveryAddresses != nil {
		return true
	}

	return false
}

// SetRecoveryAddresses gets a reference to the given []RecoveryIdentityAddress and assigns it to the RecoveryAddresses field.
func (o *Identity) SetRecoveryAddresses(v []RecoveryIdentityAddress) {
	o.RecoveryAddresses = v
}

// GetSchemaId returns the SchemaId field value
func (o *Identity) GetSchemaId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.SchemaId
}

// GetSchemaIdOk returns a tuple with the SchemaId field value
// and a boolean to check if the value has been set.
func (o *Identity) GetSchemaIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.SchemaId, true
}

// SetSchemaId sets field value
func (o *Identity) SetSchemaId(v string) {
	o.SchemaId = v
}

// GetSchemaUrl returns the SchemaUrl field value
func (o *Identity) GetSchemaUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.SchemaUrl
}

// GetSchemaUrlOk returns a tuple with the SchemaUrl field value
// and a boolean to check if the value has been set.
func (o *Identity) GetSchemaUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.SchemaUrl, true
}

// SetSchemaUrl sets field value
func (o *Identity) SetSchemaUrl(v string) {
	o.SchemaUrl = v
}

// GetState returns the State field value if set, zero value otherwise.
func (o *Identity) GetState() IdentityState {
	if o == nil || o.State == nil {
		var ret IdentityState
		return ret
	}
	return *o.State
}

// GetStateOk returns a tuple with the State field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Identity) GetStateOk() (*IdentityState, bool) {
	if o == nil || o.State == nil {
		return nil, false
	}
	return o.State, true
}

// HasState returns a boolean if a field has been set.
func (o *Identity) HasState() bool {
	if o != nil && o.State != nil {
		return true
	}

	return false
}

// SetState gets a reference to the given IdentityState and assigns it to the State field.
func (o *Identity) SetState(v IdentityState) {
	o.State = &v
}

// GetStateChangedAt returns the StateChangedAt field value if set, zero value otherwise.
func (o *Identity) GetStateChangedAt() time.Time {
	if o == nil || o.StateChangedAt == nil {
		var ret time.Time
		return ret
	}
	return *o.StateChangedAt
}

// GetStateChangedAtOk returns a tuple with the StateChangedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Identity) GetStateChangedAtOk() (*time.Time, bool) {
	if o == nil || o.StateChangedAt == nil {
		return nil, false
	}
	return o.StateChangedAt, true
}

// HasStateChangedAt returns a boolean if a field has been set.
func (o *Identity) HasStateChangedAt() bool {
	if o != nil && o.StateChangedAt != nil {
		return true
	}

	return false
}

// SetStateChangedAt gets a reference to the given time.Time and assigns it to the StateChangedAt field.
func (o *Identity) SetStateChangedAt(v time.Time) {
	o.StateChangedAt = &v
}

// GetTraits returns the Traits field value
// If the value is explicit nil, the zero value for interface{} will be returned
func (o *Identity) GetTraits() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}

	return o.Traits
}

// GetTraitsOk returns a tuple with the Traits field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Identity) GetTraitsOk() (*interface{}, bool) {
	if o == nil || o.Traits == nil {
		return nil, false
	}
	return &o.Traits, true
}

// SetTraits sets field value
func (o *Identity) SetTraits(v interface{}) {
	o.Traits = v
}

// GetUpdatedAt returns the UpdatedAt field value if set, zero value otherwise.
func (o *Identity) GetUpdatedAt() time.Time {
	if o == nil || o.UpdatedAt == nil {
		var ret time.Time
		return ret
	}
	return *o.UpdatedAt
}

// GetUpdatedAtOk returns a tuple with the UpdatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Identity) GetUpdatedAtOk() (*time.Time, bool) {
	if o == nil || o.UpdatedAt == nil {
		return nil, false
	}
	return o.UpdatedAt, true
}

// HasUpdatedAt returns a boolean if a field has been set.
func (o *Identity) HasUpdatedAt() bool {
	if o != nil && o.UpdatedAt != nil {
		return true
	}

	return false
}

// SetUpdatedAt gets a reference to the given time.Time and assigns it to the UpdatedAt field.
func (o *Identity) SetUpdatedAt(v time.Time) {
	o.UpdatedAt = &v
}

// GetVerifiableAddresses returns the VerifiableAddresses field value if set, zero value otherwise.
func (o *Identity) GetVerifiableAddresses() []VerifiableIdentityAddress {
	if o == nil || o.VerifiableAddresses == nil {
		var ret []VerifiableIdentityAddress
		return ret
	}
	return o.VerifiableAddresses
}

// GetVerifiableAddressesOk returns a tuple with the VerifiableAddresses field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Identity) GetVerifiableAddressesOk() ([]VerifiableIdentityAddress, bool) {
	if o == nil || o.VerifiableAddresses == nil {
		return nil, false
	}
	return o.VerifiableAddresses, true
}

// HasVerifiableAddresses returns a boolean if a field has been set.
func (o *Identity) HasVerifiableAddresses() bool {
	if o != nil && o.VerifiableAddresses != nil {
		return true
	}

	return false
}

// SetVerifiableAddresses gets a reference to the given []VerifiableIdentityAddress and assigns it to the VerifiableAddresses field.
func (o *Identity) SetVerifiableAddresses(v []VerifiableIdentityAddress) {
	o.VerifiableAddresses = v
}

func (o Identity) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.CreatedAt != nil {
		toSerialize["created_at"] = o.CreatedAt
	}
	if o.Credentials != nil {
		toSerialize["credentials"] = o.Credentials
	}
	if true {
		toSerialize["id"] = o.Id
	}
	if o.MetadataAdmin != nil {
		toSerialize["metadata_admin"] = o.MetadataAdmin
	}
	if o.MetadataPublic != nil {
		toSerialize["metadata_public"] = o.MetadataPublic
	}
	if o.RecoveryAddresses != nil {
		toSerialize["recovery_addresses"] = o.RecoveryAddresses
	}
	if true {
		toSerialize["schema_id"] = o.SchemaId
	}
	if true {
		toSerialize["schema_url"] = o.SchemaUrl
	}
	if o.State != nil {
		toSerialize["state"] = o.State
	}
	if o.StateChangedAt != nil {
		toSerialize["state_changed_at"] = o.StateChangedAt
	}
	if o.Traits != nil {
		toSerialize["traits"] = o.Traits
	}
	if o.UpdatedAt != nil {
		toSerialize["updated_at"] = o.UpdatedAt
	}
	if o.VerifiableAddresses != nil {
		toSerialize["verifiable_addresses"] = o.VerifiableAddresses
	}
	return json.Marshal(toSerialize)
}

type NullableIdentity struct {
	value *Identity
	isSet bool
}

func (v NullableIdentity) Get() *Identity {
	return v.value
}

func (v *NullableIdentity) Set(val *Identity) {
	v.value = val
	v.isSet = true
}

func (v NullableIdentity) IsSet() bool {
	return v.isSet
}

func (v *NullableIdentity) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIdentity(val *Identity) *NullableIdentity {
	return &NullableIdentity{value: val, isSet: true}
}

func (v NullableIdentity) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIdentity) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
