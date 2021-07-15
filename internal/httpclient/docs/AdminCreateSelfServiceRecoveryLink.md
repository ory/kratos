# AdminCreateSelfServiceRecoveryLink

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExpiresIn** | Pointer to **string** | Link Expires In  The recovery link will expire at that point in time. Defaults to the configuration value of &#x60;selfservice.flows.recovery.request_lifespan&#x60;. | [optional] 
**IdentityId** | **string** |  | 

## Methods

### NewAdminCreateSelfServiceRecoveryLink

`func NewAdminCreateSelfServiceRecoveryLink(identityId string, ) *AdminCreateSelfServiceRecoveryLink`

NewAdminCreateSelfServiceRecoveryLink instantiates a new AdminCreateSelfServiceRecoveryLink object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAdminCreateSelfServiceRecoveryLinkWithDefaults

`func NewAdminCreateSelfServiceRecoveryLinkWithDefaults() *AdminCreateSelfServiceRecoveryLink`

NewAdminCreateSelfServiceRecoveryLinkWithDefaults instantiates a new AdminCreateSelfServiceRecoveryLink object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExpiresIn

`func (o *AdminCreateSelfServiceRecoveryLink) GetExpiresIn() string`

GetExpiresIn returns the ExpiresIn field if non-nil, zero value otherwise.

### GetExpiresInOk

`func (o *AdminCreateSelfServiceRecoveryLink) GetExpiresInOk() (*string, bool)`

GetExpiresInOk returns a tuple with the ExpiresIn field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresIn

`func (o *AdminCreateSelfServiceRecoveryLink) SetExpiresIn(v string)`

SetExpiresIn sets ExpiresIn field to given value.

### HasExpiresIn

`func (o *AdminCreateSelfServiceRecoveryLink) HasExpiresIn() bool`

HasExpiresIn returns a boolean if a field has been set.

### GetIdentityId

`func (o *AdminCreateSelfServiceRecoveryLink) GetIdentityId() string`

GetIdentityId returns the IdentityId field if non-nil, zero value otherwise.

### GetIdentityIdOk

`func (o *AdminCreateSelfServiceRecoveryLink) GetIdentityIdOk() (*string, bool)`

GetIdentityIdOk returns a tuple with the IdentityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentityId

`func (o *AdminCreateSelfServiceRecoveryLink) SetIdentityId(v string)`

SetIdentityId sets IdentityId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


