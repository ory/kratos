# AdminCreateSelfServiceRecoveryLinkBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExpiresIn** | Pointer to **string** | Link Expires In  The recovery link will expire at that point in time. Defaults to the configuration value of &#x60;selfservice.flows.recovery.request_lifespan&#x60;. | [optional] 
**IdentityId** | **string** |  | 

## Methods

### NewAdminCreateSelfServiceRecoveryLinkBody

`func NewAdminCreateSelfServiceRecoveryLinkBody(identityId string, ) *AdminCreateSelfServiceRecoveryLinkBody`

NewAdminCreateSelfServiceRecoveryLinkBody instantiates a new AdminCreateSelfServiceRecoveryLinkBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAdminCreateSelfServiceRecoveryLinkBodyWithDefaults

`func NewAdminCreateSelfServiceRecoveryLinkBodyWithDefaults() *AdminCreateSelfServiceRecoveryLinkBody`

NewAdminCreateSelfServiceRecoveryLinkBodyWithDefaults instantiates a new AdminCreateSelfServiceRecoveryLinkBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExpiresIn

`func (o *AdminCreateSelfServiceRecoveryLinkBody) GetExpiresIn() string`

GetExpiresIn returns the ExpiresIn field if non-nil, zero value otherwise.

### GetExpiresInOk

`func (o *AdminCreateSelfServiceRecoveryLinkBody) GetExpiresInOk() (*string, bool)`

GetExpiresInOk returns a tuple with the ExpiresIn field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresIn

`func (o *AdminCreateSelfServiceRecoveryLinkBody) SetExpiresIn(v string)`

SetExpiresIn sets ExpiresIn field to given value.

### HasExpiresIn

`func (o *AdminCreateSelfServiceRecoveryLinkBody) HasExpiresIn() bool`

HasExpiresIn returns a boolean if a field has been set.

### GetIdentityId

`func (o *AdminCreateSelfServiceRecoveryLinkBody) GetIdentityId() string`

GetIdentityId returns the IdentityId field if non-nil, zero value otherwise.

### GetIdentityIdOk

`func (o *AdminCreateSelfServiceRecoveryLinkBody) GetIdentityIdOk() (*string, bool)`

GetIdentityIdOk returns a tuple with the IdentityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentityId

`func (o *AdminCreateSelfServiceRecoveryLinkBody) SetIdentityId(v string)`

SetIdentityId sets IdentityId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


