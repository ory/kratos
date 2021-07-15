# SubmitSelfServiceSettingsFlowBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | The Anti-CSRF Token  This token is only required when performing browser flows. | [optional] 
**Method** | **string** | Method  Should be set to profile when trying to update a profile. | 
**Password** | **string** | Password is the updated password | 
**Traits** | **map[string]interface{}** | Traits contains all of the identity&#39;s traits. | 

## Methods

### NewSubmitSelfServiceSettingsFlowBody

`func NewSubmitSelfServiceSettingsFlowBody(method string, password string, traits map[string]interface{}, ) *SubmitSelfServiceSettingsFlowBody`

NewSubmitSelfServiceSettingsFlowBody instantiates a new SubmitSelfServiceSettingsFlowBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceSettingsFlowBodyWithDefaults

`func NewSubmitSelfServiceSettingsFlowBodyWithDefaults() *SubmitSelfServiceSettingsFlowBody`

NewSubmitSelfServiceSettingsFlowBodyWithDefaults instantiates a new SubmitSelfServiceSettingsFlowBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceSettingsFlowBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceSettingsFlowBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceSettingsFlowBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetPassword

`func (o *SubmitSelfServiceSettingsFlowBody) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *SubmitSelfServiceSettingsFlowBody) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetTraits

`func (o *SubmitSelfServiceSettingsFlowBody) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *SubmitSelfServiceSettingsFlowBody) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


