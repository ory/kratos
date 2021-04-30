# VerifiableAddress

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Status** | **string** |  | 
**Value** | **string** |  | 
**Verified** | **bool** |  | 
**VerifiedAt** | Pointer to **time.Time** |  | [optional] 
**Via** | **string** |  | 

## Methods

### NewVerifiableAddress

`func NewVerifiableAddress(id string, status string, value string, verified bool, via string, ) *VerifiableAddress`

NewVerifiableAddress instantiates a new VerifiableAddress object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVerifiableAddressWithDefaults

`func NewVerifiableAddressWithDefaults() *VerifiableAddress`

NewVerifiableAddressWithDefaults instantiates a new VerifiableAddress object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *VerifiableAddress) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *VerifiableAddress) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *VerifiableAddress) SetId(v string)`

SetId sets Id field to given value.


### GetStatus

`func (o *VerifiableAddress) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *VerifiableAddress) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *VerifiableAddress) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetValue

`func (o *VerifiableAddress) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *VerifiableAddress) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *VerifiableAddress) SetValue(v string)`

SetValue sets Value field to given value.


### GetVerified

`func (o *VerifiableAddress) GetVerified() bool`

GetVerified returns the Verified field if non-nil, zero value otherwise.

### GetVerifiedOk

`func (o *VerifiableAddress) GetVerifiedOk() (*bool, bool)`

GetVerifiedOk returns a tuple with the Verified field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVerified

`func (o *VerifiableAddress) SetVerified(v bool)`

SetVerified sets Verified field to given value.


### GetVerifiedAt

`func (o *VerifiableAddress) GetVerifiedAt() time.Time`

GetVerifiedAt returns the VerifiedAt field if non-nil, zero value otherwise.

### GetVerifiedAtOk

`func (o *VerifiableAddress) GetVerifiedAtOk() (*time.Time, bool)`

GetVerifiedAtOk returns a tuple with the VerifiedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVerifiedAt

`func (o *VerifiableAddress) SetVerifiedAt(v time.Time)`

SetVerifiedAt sets VerifiedAt field to given value.

### HasVerifiedAt

`func (o *VerifiableAddress) HasVerifiedAt() bool`

HasVerifiedAt returns a boolean if a field has been set.

### GetVia

`func (o *VerifiableAddress) GetVia() string`

GetVia returns the Via field if non-nil, zero value otherwise.

### GetViaOk

`func (o *VerifiableAddress) GetViaOk() (*string, bool)`

GetViaOk returns a tuple with the Via field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVia

`func (o *VerifiableAddress) SetVia(v string)`

SetVia sets Via field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


