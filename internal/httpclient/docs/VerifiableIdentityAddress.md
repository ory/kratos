# VerifiableIdentityAddress

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreatedAt** | Pointer to **time.Time** | When this entry was created | [optional] 
**Id** | **string** |  | 
**Status** | **string** | VerifiableAddressStatus must not exceed 16 characters as that is the limitation in the SQL Schema | 
**UpdatedAt** | Pointer to **time.Time** | When this entry was last updated | [optional] 
**Value** | **string** | The address value  example foo@user.com | 
**Verified** | **bool** | Indicates if the address has already been verified | 
**VerifiedAt** | Pointer to **time.Time** |  | [optional] 
**Via** | **string** | VerifiableAddressType must not exceed 16 characters as that is the limitation in the SQL Schema | 

## Methods

### NewVerifiableIdentityAddress

`func NewVerifiableIdentityAddress(id string, status string, value string, verified bool, via string, ) *VerifiableIdentityAddress`

NewVerifiableIdentityAddress instantiates a new VerifiableIdentityAddress object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVerifiableIdentityAddressWithDefaults

`func NewVerifiableIdentityAddressWithDefaults() *VerifiableIdentityAddress`

NewVerifiableIdentityAddressWithDefaults instantiates a new VerifiableIdentityAddress object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreatedAt

`func (o *VerifiableIdentityAddress) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *VerifiableIdentityAddress) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *VerifiableIdentityAddress) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *VerifiableIdentityAddress) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetId

`func (o *VerifiableIdentityAddress) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *VerifiableIdentityAddress) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *VerifiableIdentityAddress) SetId(v string)`

SetId sets Id field to given value.


### GetStatus

`func (o *VerifiableIdentityAddress) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *VerifiableIdentityAddress) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *VerifiableIdentityAddress) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetUpdatedAt

`func (o *VerifiableIdentityAddress) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *VerifiableIdentityAddress) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *VerifiableIdentityAddress) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *VerifiableIdentityAddress) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetValue

`func (o *VerifiableIdentityAddress) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *VerifiableIdentityAddress) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *VerifiableIdentityAddress) SetValue(v string)`

SetValue sets Value field to given value.


### GetVerified

`func (o *VerifiableIdentityAddress) GetVerified() bool`

GetVerified returns the Verified field if non-nil, zero value otherwise.

### GetVerifiedOk

`func (o *VerifiableIdentityAddress) GetVerifiedOk() (*bool, bool)`

GetVerifiedOk returns a tuple with the Verified field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVerified

`func (o *VerifiableIdentityAddress) SetVerified(v bool)`

SetVerified sets Verified field to given value.


### GetVerifiedAt

`func (o *VerifiableIdentityAddress) GetVerifiedAt() time.Time`

GetVerifiedAt returns the VerifiedAt field if non-nil, zero value otherwise.

### GetVerifiedAtOk

`func (o *VerifiableIdentityAddress) GetVerifiedAtOk() (*time.Time, bool)`

GetVerifiedAtOk returns a tuple with the VerifiedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVerifiedAt

`func (o *VerifiableIdentityAddress) SetVerifiedAt(v time.Time)`

SetVerifiedAt sets VerifiedAt field to given value.

### HasVerifiedAt

`func (o *VerifiableIdentityAddress) HasVerifiedAt() bool`

HasVerifiedAt returns a boolean if a field has been set.

### GetVia

`func (o *VerifiableIdentityAddress) GetVia() string`

GetVia returns the Via field if non-nil, zero value otherwise.

### GetViaOk

`func (o *VerifiableIdentityAddress) GetViaOk() (*string, bool)`

GetViaOk returns a tuple with the Via field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVia

`func (o *VerifiableIdentityAddress) SetVia(v string)`

SetVia sets Via field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


