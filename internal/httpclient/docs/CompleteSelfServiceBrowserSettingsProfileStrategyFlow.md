# CompleteSelfServiceBrowserSettingsProfileStrategyFlow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | The Anti-CSRF Token  This token is only required when performing browser flows. | [optional] 
**Method** | Pointer to **string** | Method  Should be set to profile when trying to update a profile.  type: string | [optional] 
**Traits** | **map[string]interface{}** | Traits contains all of the identity&#39;s traits. | 

## Methods

### NewCompleteSelfServiceBrowserSettingsProfileStrategyFlow

`func NewCompleteSelfServiceBrowserSettingsProfileStrategyFlow(traits map[string]interface{}, ) *CompleteSelfServiceBrowserSettingsProfileStrategyFlow`

NewCompleteSelfServiceBrowserSettingsProfileStrategyFlow instantiates a new CompleteSelfServiceBrowserSettingsProfileStrategyFlow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCompleteSelfServiceBrowserSettingsProfileStrategyFlowWithDefaults

`func NewCompleteSelfServiceBrowserSettingsProfileStrategyFlowWithDefaults() *CompleteSelfServiceBrowserSettingsProfileStrategyFlow`

NewCompleteSelfServiceBrowserSettingsProfileStrategyFlowWithDefaults instantiates a new CompleteSelfServiceBrowserSettingsProfileStrategyFlow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) SetMethod(v string)`

SetMethod sets Method field to given value.

### HasMethod

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) HasMethod() bool`

HasMethod returns a boolean if a field has been set.

### GetTraits

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


