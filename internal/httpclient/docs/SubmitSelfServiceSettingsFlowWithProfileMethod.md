# SubmitSelfServiceSettingsFlowWithProfileMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | The Anti-CSRF Token  This token is only required when performing browser flows. | [optional] 
**Method** | Pointer to **string** | Method  Should be set to profile when trying to update a profile.  type: string | [optional] 
**Traits** | **map[string]interface{}** | Traits contains all of the identity&#39;s traits. | 

## Methods

### NewSubmitSelfServiceSettingsFlowWithProfileMethod

`func NewSubmitSelfServiceSettingsFlowWithProfileMethod(traits map[string]interface{}, ) *SubmitSelfServiceSettingsFlowWithProfileMethod`

NewSubmitSelfServiceSettingsFlowWithProfileMethod instantiates a new SubmitSelfServiceSettingsFlowWithProfileMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceSettingsFlowWithProfileMethodWithDefaults

`func NewSubmitSelfServiceSettingsFlowWithProfileMethodWithDefaults() *SubmitSelfServiceSettingsFlowWithProfileMethod`

NewSubmitSelfServiceSettingsFlowWithProfileMethodWithDefaults instantiates a new SubmitSelfServiceSettingsFlowWithProfileMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) SetMethod(v string)`

SetMethod sets Method field to given value.

### HasMethod

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) HasMethod() bool`

HasMethod returns a boolean if a field has been set.

### GetTraits

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *SubmitSelfServiceSettingsFlowWithProfileMethod) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


