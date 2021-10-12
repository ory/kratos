# SubmitSelfServiceRegistrationFlowWithOidcMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | The CSRF Token | [optional] 
**Method** | **string** | Method to use  This field must be set to &#x60;oidc&#x60; when using the oidc method. | 
**Provider** | **string** | The provider to register with | 
**Traits** | Pointer to **map[string]interface{}** | The identity traits | [optional] 

## Methods

### NewSubmitSelfServiceRegistrationFlowWithOidcMethodBody

`func NewSubmitSelfServiceRegistrationFlowWithOidcMethodBody(method string, provider string, ) *SubmitSelfServiceRegistrationFlowWithOidcMethodBody`

NewSubmitSelfServiceRegistrationFlowWithOidcMethodBody instantiates a new SubmitSelfServiceRegistrationFlowWithOidcMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceRegistrationFlowWithOidcMethodBodyWithDefaults

`func NewSubmitSelfServiceRegistrationFlowWithOidcMethodBodyWithDefaults() *SubmitSelfServiceRegistrationFlowWithOidcMethodBody`

NewSubmitSelfServiceRegistrationFlowWithOidcMethodBodyWithDefaults instantiates a new SubmitSelfServiceRegistrationFlowWithOidcMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetProvider

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) GetProvider() string`

GetProvider returns the Provider field if non-nil, zero value otherwise.

### GetProviderOk

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) GetProviderOk() (*string, bool)`

GetProviderOk returns a tuple with the Provider field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProvider

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) SetProvider(v string)`

SetProvider sets Provider field to given value.


### GetTraits

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.

### HasTraits

`func (o *SubmitSelfServiceRegistrationFlowWithOidcMethodBody) HasTraits() bool`

HasTraits returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


