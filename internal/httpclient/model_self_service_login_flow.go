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

// SelfServiceLoginFlow This object represents a login flow. A login flow is initiated at the \"Initiate Login API / Browser Flow\" endpoint by a client.  Once a login flow is completed successfully, a session cookie or session token will be issued.
type SelfServiceLoginFlow struct {
	// and so on.
	Active *string `json:"active,omitempty"`
	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// ExpiresAt is the time (UTC) when the flow expires. If the user still wishes to log in, a new flow has to be initiated.
	ExpiresAt time.Time `json:"expires_at"`
	// Forced stores whether this login flow should enforce re-authentication.
	Forced *bool  `json:"forced,omitempty"`
	Id     string `json:"id"`
	// IssuedAt is the time (UTC) when the flow started.
	IssuedAt time.Time `json:"issued_at"`
	// RequestURL is the initial URL that was requested from Ory Kratos. It can be used to forward information contained in the URL's path or query for example.
	RequestUrl string `json:"request_url"`
	// The flow type can either be `api` or `browser`.
	Type string      `json:"type"`
	Ui   UiContainer `json:"ui"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// NewSelfServiceLoginFlow instantiates a new SelfServiceLoginFlow object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSelfServiceLoginFlow(expiresAt time.Time, id string, issuedAt time.Time, requestUrl string, type_ string, ui UiContainer) *SelfServiceLoginFlow {
	this := SelfServiceLoginFlow{}
	this.ExpiresAt = expiresAt
	this.Id = id
	this.IssuedAt = issuedAt
	this.RequestUrl = requestUrl
	this.Type = type_
	this.Ui = ui
	return &this
}

// NewSelfServiceLoginFlowWithDefaults instantiates a new SelfServiceLoginFlow object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSelfServiceLoginFlowWithDefaults() *SelfServiceLoginFlow {
	this := SelfServiceLoginFlow{}
	return &this
}

// GetActive returns the Active field value if set, zero value otherwise.
func (o *SelfServiceLoginFlow) GetActive() string {
	if o == nil || o.Active == nil {
		var ret string
		return ret
	}
	return *o.Active
}

// GetActiveOk returns a tuple with the Active field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetActiveOk() (*string, bool) {
	if o == nil || o.Active == nil {
		return nil, false
	}
	return o.Active, true
}

// HasActive returns a boolean if a field has been set.
func (o *SelfServiceLoginFlow) HasActive() bool {
	if o != nil && o.Active != nil {
		return true
	}

	return false
}

// SetActive gets a reference to the given string and assigns it to the Active field.
func (o *SelfServiceLoginFlow) SetActive(v string) {
	o.Active = &v
}

// GetCreatedAt returns the CreatedAt field value if set, zero value otherwise.
func (o *SelfServiceLoginFlow) GetCreatedAt() time.Time {
	if o == nil || o.CreatedAt == nil {
		var ret time.Time
		return ret
	}
	return *o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetCreatedAtOk() (*time.Time, bool) {
	if o == nil || o.CreatedAt == nil {
		return nil, false
	}
	return o.CreatedAt, true
}

// HasCreatedAt returns a boolean if a field has been set.
func (o *SelfServiceLoginFlow) HasCreatedAt() bool {
	if o != nil && o.CreatedAt != nil {
		return true
	}

	return false
}

// SetCreatedAt gets a reference to the given time.Time and assigns it to the CreatedAt field.
func (o *SelfServiceLoginFlow) SetCreatedAt(v time.Time) {
	o.CreatedAt = &v
}

// GetExpiresAt returns the ExpiresAt field value
func (o *SelfServiceLoginFlow) GetExpiresAt() time.Time {
	if o == nil {
		var ret time.Time
		return ret
	}

	return o.ExpiresAt
}

// GetExpiresAtOk returns a tuple with the ExpiresAt field value
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetExpiresAtOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ExpiresAt, true
}

// SetExpiresAt sets field value
func (o *SelfServiceLoginFlow) SetExpiresAt(v time.Time) {
	o.ExpiresAt = v
}

// GetForced returns the Forced field value if set, zero value otherwise.
func (o *SelfServiceLoginFlow) GetForced() bool {
	if o == nil || o.Forced == nil {
		var ret bool
		return ret
	}
	return *o.Forced
}

// GetForcedOk returns a tuple with the Forced field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetForcedOk() (*bool, bool) {
	if o == nil || o.Forced == nil {
		return nil, false
	}
	return o.Forced, true
}

// HasForced returns a boolean if a field has been set.
func (o *SelfServiceLoginFlow) HasForced() bool {
	if o != nil && o.Forced != nil {
		return true
	}

	return false
}

// SetForced gets a reference to the given bool and assigns it to the Forced field.
func (o *SelfServiceLoginFlow) SetForced(v bool) {
	o.Forced = &v
}

// GetId returns the Id field value
func (o *SelfServiceLoginFlow) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *SelfServiceLoginFlow) SetId(v string) {
	o.Id = v
}

// GetIssuedAt returns the IssuedAt field value
func (o *SelfServiceLoginFlow) GetIssuedAt() time.Time {
	if o == nil {
		var ret time.Time
		return ret
	}

	return o.IssuedAt
}

// GetIssuedAtOk returns a tuple with the IssuedAt field value
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetIssuedAtOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return &o.IssuedAt, true
}

// SetIssuedAt sets field value
func (o *SelfServiceLoginFlow) SetIssuedAt(v time.Time) {
	o.IssuedAt = v
}

// GetRequestUrl returns the RequestUrl field value
func (o *SelfServiceLoginFlow) GetRequestUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RequestUrl
}

// GetRequestUrlOk returns a tuple with the RequestUrl field value
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetRequestUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.RequestUrl, true
}

// SetRequestUrl sets field value
func (o *SelfServiceLoginFlow) SetRequestUrl(v string) {
	o.RequestUrl = v
}

// GetType returns the Type field value
func (o *SelfServiceLoginFlow) GetType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *SelfServiceLoginFlow) SetType(v string) {
	o.Type = v
}

// GetUi returns the Ui field value
func (o *SelfServiceLoginFlow) GetUi() UiContainer {
	if o == nil {
		var ret UiContainer
		return ret
	}

	return o.Ui
}

// GetUiOk returns a tuple with the Ui field value
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetUiOk() (*UiContainer, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Ui, true
}

// SetUi sets field value
func (o *SelfServiceLoginFlow) SetUi(v UiContainer) {
	o.Ui = v
}

// GetUpdatedAt returns the UpdatedAt field value if set, zero value otherwise.
func (o *SelfServiceLoginFlow) GetUpdatedAt() time.Time {
	if o == nil || o.UpdatedAt == nil {
		var ret time.Time
		return ret
	}
	return *o.UpdatedAt
}

// GetUpdatedAtOk returns a tuple with the UpdatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SelfServiceLoginFlow) GetUpdatedAtOk() (*time.Time, bool) {
	if o == nil || o.UpdatedAt == nil {
		return nil, false
	}
	return o.UpdatedAt, true
}

// HasUpdatedAt returns a boolean if a field has been set.
func (o *SelfServiceLoginFlow) HasUpdatedAt() bool {
	if o != nil && o.UpdatedAt != nil {
		return true
	}

	return false
}

// SetUpdatedAt gets a reference to the given time.Time and assigns it to the UpdatedAt field.
func (o *SelfServiceLoginFlow) SetUpdatedAt(v time.Time) {
	o.UpdatedAt = &v
}

func (o SelfServiceLoginFlow) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Active != nil {
		toSerialize["active"] = o.Active
	}
	if o.CreatedAt != nil {
		toSerialize["created_at"] = o.CreatedAt
	}
	if true {
		toSerialize["expires_at"] = o.ExpiresAt
	}
	if o.Forced != nil {
		toSerialize["forced"] = o.Forced
	}
	if true {
		toSerialize["id"] = o.Id
	}
	if true {
		toSerialize["issued_at"] = o.IssuedAt
	}
	if true {
		toSerialize["request_url"] = o.RequestUrl
	}
	if true {
		toSerialize["type"] = o.Type
	}
	if true {
		toSerialize["ui"] = o.Ui
	}
	if o.UpdatedAt != nil {
		toSerialize["updated_at"] = o.UpdatedAt
	}
	return json.Marshal(toSerialize)
}

type NullableSelfServiceLoginFlow struct {
	value *SelfServiceLoginFlow
	isSet bool
}

func (v NullableSelfServiceLoginFlow) Get() *SelfServiceLoginFlow {
	return v.value
}

func (v *NullableSelfServiceLoginFlow) Set(val *SelfServiceLoginFlow) {
	v.value = val
	v.isSet = true
}

func (v NullableSelfServiceLoginFlow) IsSet() bool {
	return v.isSet
}

func (v *NullableSelfServiceLoginFlow) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSelfServiceLoginFlow(val *SelfServiceLoginFlow) *NullableSelfServiceLoginFlow {
	return &NullableSelfServiceLoginFlow{value: val, isSet: true}
}

func (v NullableSelfServiceLoginFlow) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSelfServiceLoginFlow) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
