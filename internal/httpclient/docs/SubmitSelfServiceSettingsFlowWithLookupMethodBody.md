# SubmitSelfServiceSettingsFlowWithLookupMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | CSRFToken is the anti-CSRF token | [optional] 
**LookupSecretConfirm** | Pointer to **bool** | If set to true will save the regenerated lookup secrets | [optional] 
**LookupSecretRegenerate** | Pointer to **bool** | If set to true will regenerate the lookup secrets | [optional] 
**LookupSecretReveal** | Pointer to **bool** | If set to true will reveal the lookup secrets | [optional] 
**Method** | **string** | Method  Should be set to \&quot;lookup\&quot; when trying to add, update, or remove a lookup pairing. | 

## Methods

### NewSubmitSelfServiceSettingsFlowWithLookupMethodBody

`func NewSubmitSelfServiceSettingsFlowWithLookupMethodBody(method string, ) *SubmitSelfServiceSettingsFlowWithLookupMethodBody`

NewSubmitSelfServiceSettingsFlowWithLookupMethodBody instantiates a new SubmitSelfServiceSettingsFlowWithLookupMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceSettingsFlowWithLookupMethodBodyWithDefaults

`func NewSubmitSelfServiceSettingsFlowWithLookupMethodBodyWithDefaults() *SubmitSelfServiceSettingsFlowWithLookupMethodBody`

NewSubmitSelfServiceSettingsFlowWithLookupMethodBodyWithDefaults instantiates a new SubmitSelfServiceSettingsFlowWithLookupMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetLookupSecretConfirm

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetLookupSecretConfirm() bool`

GetLookupSecretConfirm returns the LookupSecretConfirm field if non-nil, zero value otherwise.

### GetLookupSecretConfirmOk

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetLookupSecretConfirmOk() (*bool, bool)`

GetLookupSecretConfirmOk returns a tuple with the LookupSecretConfirm field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLookupSecretConfirm

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) SetLookupSecretConfirm(v bool)`

SetLookupSecretConfirm sets LookupSecretConfirm field to given value.

### HasLookupSecretConfirm

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) HasLookupSecretConfirm() bool`

HasLookupSecretConfirm returns a boolean if a field has been set.

### GetLookupSecretRegenerate

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetLookupSecretRegenerate() bool`

GetLookupSecretRegenerate returns the LookupSecretRegenerate field if non-nil, zero value otherwise.

### GetLookupSecretRegenerateOk

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetLookupSecretRegenerateOk() (*bool, bool)`

GetLookupSecretRegenerateOk returns a tuple with the LookupSecretRegenerate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLookupSecretRegenerate

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) SetLookupSecretRegenerate(v bool)`

SetLookupSecretRegenerate sets LookupSecretRegenerate field to given value.

### HasLookupSecretRegenerate

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) HasLookupSecretRegenerate() bool`

HasLookupSecretRegenerate returns a boolean if a field has been set.

### GetLookupSecretReveal

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetLookupSecretReveal() bool`

GetLookupSecretReveal returns the LookupSecretReveal field if non-nil, zero value otherwise.

### GetLookupSecretRevealOk

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetLookupSecretRevealOk() (*bool, bool)`

GetLookupSecretRevealOk returns a tuple with the LookupSecretReveal field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLookupSecretReveal

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) SetLookupSecretReveal(v bool)`

SetLookupSecretReveal sets LookupSecretReveal field to given value.

### HasLookupSecretReveal

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) HasLookupSecretReveal() bool`

HasLookupSecretReveal returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceSettingsFlowWithLookupMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


