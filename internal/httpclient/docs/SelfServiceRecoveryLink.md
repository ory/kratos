# SelfServiceRecoveryLink

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExpiresAt** | Pointer to **time.Time** | Recovery Link Expires At  The timestamp when the recovery link expires. | [optional] 
**RecoveryLink** | **string** | Recovery Link  This link can be used to recover the account. | 

## Methods

### NewSelfServiceRecoveryLink

`func NewSelfServiceRecoveryLink(recoveryLink string, ) *SelfServiceRecoveryLink`

NewSelfServiceRecoveryLink instantiates a new SelfServiceRecoveryLink object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceRecoveryLinkWithDefaults

`func NewSelfServiceRecoveryLinkWithDefaults() *SelfServiceRecoveryLink`

NewSelfServiceRecoveryLinkWithDefaults instantiates a new SelfServiceRecoveryLink object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExpiresAt

`func (o *SelfServiceRecoveryLink) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *SelfServiceRecoveryLink) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *SelfServiceRecoveryLink) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *SelfServiceRecoveryLink) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.

### GetRecoveryLink

`func (o *SelfServiceRecoveryLink) GetRecoveryLink() string`

GetRecoveryLink returns the RecoveryLink field if non-nil, zero value otherwise.

### GetRecoveryLinkOk

`func (o *SelfServiceRecoveryLink) GetRecoveryLinkOk() (*string, bool)`

GetRecoveryLinkOk returns a tuple with the RecoveryLink field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecoveryLink

`func (o *SelfServiceRecoveryLink) SetRecoveryLink(v string)`

SetRecoveryLink sets RecoveryLink field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


