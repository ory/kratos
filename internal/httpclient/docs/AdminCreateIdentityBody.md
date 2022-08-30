# AdminCreateIdentityBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Credentials** | Pointer to [**AdminIdentityImportCredentials**](AdminIdentityImportCredentials.md) |  | [optional] 
**MetadataAdmin** | Pointer to **interface{}** | Store metadata about the user which is only accessible through admin APIs such as &#x60;GET /admin/identities/&lt;id&gt;&#x60;. | [optional] 
**MetadataPublic** | Pointer to **interface{}** | Store metadata about the identity which the identity itself can see when calling for example the session endpoint. Do not store sensitive information (e.g. credit score) about the identity in this field. | [optional] 
**RecoveryAddresses** | Pointer to [**[]RecoveryIdentityAddress**](RecoveryIdentityAddress.md) | RecoveryAddresses contains all the addresses that can be used to recover an identity.  Use this structure to import recovery addresses for an identity. Please keep in mind that the address needs to be represented in the Identity Schema or this field will be overwritten on the next identity update. | [optional] 
**SchemaId** | **string** | SchemaID is the ID of the JSON Schema to be used for validating the identity&#39;s traits. | 
**State** | Pointer to [**IdentityState**](IdentityState.md) |  | [optional] 
**Traits** | **map[string]interface{}** | Traits represent an identity&#39;s traits. The identity is able to create, modify, and delete traits in a self-service manner. The input will always be validated against the JSON Schema defined in &#x60;schema_url&#x60;. | 
**VerifiableAddresses** | Pointer to [**[]VerifiableIdentityAddress**](VerifiableIdentityAddress.md) | VerifiableAddresses contains all the addresses that can be verified by the user.  Use this structure to import verified addresses for an identity. Please keep in mind that the address needs to be represented in the Identity Schema or this field will be overwritten on the next identity update. | [optional] 

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

### GetCredentials

`func (o *AdminCreateIdentityBody) GetCredentials() AdminIdentityImportCredentials`

GetCredentials returns the Credentials field if non-nil, zero value otherwise.

### GetCredentialsOk

`func (o *AdminCreateIdentityBody) GetCredentialsOk() (*AdminIdentityImportCredentials, bool)`

GetCredentialsOk returns a tuple with the Credentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentials

`func (o *AdminCreateIdentityBody) SetCredentials(v AdminIdentityImportCredentials)`

SetCredentials sets Credentials field to given value.

### HasCredentials

`func (o *AdminCreateIdentityBody) HasCredentials() bool`

HasCredentials returns a boolean if a field has been set.

### GetMetadataAdmin

`func (o *AdminCreateIdentityBody) GetMetadataAdmin() interface{}`

GetMetadataAdmin returns the MetadataAdmin field if non-nil, zero value otherwise.

### GetMetadataAdminOk

`func (o *AdminCreateIdentityBody) GetMetadataAdminOk() (*interface{}, bool)`

GetMetadataAdminOk returns a tuple with the MetadataAdmin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadataAdmin

`func (o *AdminCreateIdentityBody) SetMetadataAdmin(v interface{})`

SetMetadataAdmin sets MetadataAdmin field to given value.

### HasMetadataAdmin

`func (o *AdminCreateIdentityBody) HasMetadataAdmin() bool`

HasMetadataAdmin returns a boolean if a field has been set.

### SetMetadataAdminNil

`func (o *AdminCreateIdentityBody) SetMetadataAdminNil(b bool)`

 SetMetadataAdminNil sets the value for MetadataAdmin to be an explicit nil

### UnsetMetadataAdmin
`func (o *AdminCreateIdentityBody) UnsetMetadataAdmin()`

UnsetMetadataAdmin ensures that no value is present for MetadataAdmin, not even an explicit nil
### GetMetadataPublic

`func (o *AdminCreateIdentityBody) GetMetadataPublic() interface{}`

GetMetadataPublic returns the MetadataPublic field if non-nil, zero value otherwise.

### GetMetadataPublicOk

`func (o *AdminCreateIdentityBody) GetMetadataPublicOk() (*interface{}, bool)`

GetMetadataPublicOk returns a tuple with the MetadataPublic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadataPublic

`func (o *AdminCreateIdentityBody) SetMetadataPublic(v interface{})`

SetMetadataPublic sets MetadataPublic field to given value.

### HasMetadataPublic

`func (o *AdminCreateIdentityBody) HasMetadataPublic() bool`

HasMetadataPublic returns a boolean if a field has been set.

### SetMetadataPublicNil

`func (o *AdminCreateIdentityBody) SetMetadataPublicNil(b bool)`

 SetMetadataPublicNil sets the value for MetadataPublic to be an explicit nil

### UnsetMetadataPublic
`func (o *AdminCreateIdentityBody) UnsetMetadataPublic()`

UnsetMetadataPublic ensures that no value is present for MetadataPublic, not even an explicit nil
### GetRecoveryAddresses

`func (o *AdminCreateIdentityBody) GetRecoveryAddresses() []RecoveryIdentityAddress`

GetRecoveryAddresses returns the RecoveryAddresses field if non-nil, zero value otherwise.

### GetRecoveryAddressesOk

`func (o *AdminCreateIdentityBody) GetRecoveryAddressesOk() (*[]RecoveryIdentityAddress, bool)`

GetRecoveryAddressesOk returns a tuple with the RecoveryAddresses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecoveryAddresses

`func (o *AdminCreateIdentityBody) SetRecoveryAddresses(v []RecoveryIdentityAddress)`

SetRecoveryAddresses sets RecoveryAddresses field to given value.

### HasRecoveryAddresses

`func (o *AdminCreateIdentityBody) HasRecoveryAddresses() bool`

HasRecoveryAddresses returns a boolean if a field has been set.

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


