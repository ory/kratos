# AdminCreateIdentityBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Password** | Pointer to [**PasswordCredential**](PasswordCredential.md) |  | [optional] 
**SchemaId** | **string** | SchemaID is the ID of the JSON Schema to be used for validating the identity&#39;s traits. | 
**State** | Pointer to [**IdentityState**](IdentityState.md) |  | [optional] 
**Traits** | **map[string]interface{}** | Traits represent an identity&#39;s traits. The identity is able to create, modify, and delete traits in a self-service manner. The input will always be validated against the JSON Schema defined in &#x60;schema_url&#x60;. | 
**VerifiableAddresses** | Pointer to [**[]VerifiableIdentityAddress**](VerifiableIdentityAddress.md) | VerifiableAddresses contains all the addresses that can be verified by the user. | [optional] 

## Methods

### NewAdminCreateIdentityBody

`func NewAdminCreateIdentityBody(schemaId string, traits map[string]interface{}, ) *AdminCreateIdentityBody`

NewAdminCreateIdentityBody instantiates a new AdminCreateIdentityBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAdminCreateIdentityBodyWithDefaults

`func NewAdminCreateIdentityBodyWithDefaults() *AdminCreateIdentityBody`

NewAdminCreateIdentityBodyWithDefaults instantiates a new AdminCreateIdentityBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPassword

`func (o *AdminCreateIdentityBody) GetPassword() PasswordCredential`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *AdminCreateIdentityBody) GetPasswordOk() (*PasswordCredential, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *AdminCreateIdentityBody) SetPassword(v PasswordCredential)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *AdminCreateIdentityBody) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetSchemaId

`func (o *AdminCreateIdentityBody) GetSchemaId() string`

GetSchemaId returns the SchemaId field if non-nil, zero value otherwise.

### GetSchemaIdOk

`func (o *AdminCreateIdentityBody) GetSchemaIdOk() (*string, bool)`

GetSchemaIdOk returns a tuple with the SchemaId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchemaId

`func (o *AdminCreateIdentityBody) SetSchemaId(v string)`

SetSchemaId sets SchemaId field to given value.


### GetState

`func (o *AdminCreateIdentityBody) GetState() IdentityState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *AdminCreateIdentityBody) GetStateOk() (*IdentityState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *AdminCreateIdentityBody) SetState(v IdentityState)`

SetState sets State field to given value.

### HasState

`func (o *AdminCreateIdentityBody) HasState() bool`

HasState returns a boolean if a field has been set.

### GetTraits

`func (o *AdminCreateIdentityBody) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *AdminCreateIdentityBody) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *AdminCreateIdentityBody) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.


### GetVerifiableAddresses

`func (o *AdminCreateIdentityBody) GetVerifiableAddresses() []VerifiableIdentityAddress`

GetVerifiableAddresses returns the VerifiableAddresses field if non-nil, zero value otherwise.

### GetVerifiableAddressesOk

`func (o *AdminCreateIdentityBody) GetVerifiableAddressesOk() (*[]VerifiableIdentityAddress, bool)`

GetVerifiableAddressesOk returns a tuple with the VerifiableAddresses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVerifiableAddresses

`func (o *AdminCreateIdentityBody) SetVerifiableAddresses(v []VerifiableIdentityAddress)`

SetVerifiableAddresses sets VerifiableAddresses field to given value.

### HasVerifiableAddresses

`func (o *AdminCreateIdentityBody) HasVerifiableAddresses() bool`

HasVerifiableAddresses returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


