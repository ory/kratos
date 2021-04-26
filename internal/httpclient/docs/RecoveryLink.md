# RecoveryLink

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExpiresAt** | Pointer to **time.Time** | Recovery Link Expires At  The timestamp when the recovery link expires. | [optional] 
**RecoveryLink** | **string** | Recovery Link  This link can be used to recover the account. | 

## Methods

### NewRecoveryLink

`func NewRecoveryLink(recoveryLink string, ) *RecoveryLink`

NewRecoveryLink instantiates a new RecoveryLink object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRecoveryLinkWithDefaults

`func NewRecoveryLinkWithDefaults() *RecoveryLink`

NewRecoveryLinkWithDefaults instantiates a new RecoveryLink object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExpiresAt

`func (o *RecoveryLink) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *RecoveryLink) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *RecoveryLink) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *RecoveryLink) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.

### GetRecoveryLink

`func (o *RecoveryLink) GetRecoveryLink() string`

GetRecoveryLink returns the RecoveryLink field if non-nil, zero value otherwise.

### GetRecoveryLinkOk

`func (o *RecoveryLink) GetRecoveryLinkOk() (*string, bool)`

GetRecoveryLinkOk returns a tuple with the RecoveryLink field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecoveryLink

`func (o *RecoveryLink) SetRecoveryLink(v string)`

SetRecoveryLink sets RecoveryLink field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


