# SubmitSelfServiceRegistrationFlowBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | The CSRF Token | [optional] 
**Method** | **string** | Method to use  This field must be set to &#x60;password&#x60; when using the password method. | 
**Password** | **string** | Password to sign the user up with | 
**Traits** | **map[string]interface{}** | The identity&#39;s traits | 

## Methods

### NewSubmitSelfServiceRegistrationFlowBody

`func NewSubmitSelfServiceRegistrationFlowBody(method string, password string, traits map[string]interface{}, ) *SubmitSelfServiceRegistrationFlowBody`

NewSubmitSelfServiceRegistrationFlowBody instantiates a new SubmitSelfServiceRegistrationFlowBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceRegistrationFlowBodyWithDefaults

`func NewSubmitSelfServiceRegistrationFlowBodyWithDefaults() *SubmitSelfServiceRegistrationFlowBody`

NewSubmitSelfServiceRegistrationFlowBodyWithDefaults instantiates a new SubmitSelfServiceRegistrationFlowBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceRegistrationFlowBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceRegistrationFlowBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceRegistrationFlowBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceRegistrationFlowBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceRegistrationFlowBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceRegistrationFlowBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceRegistrationFlowBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetPassword

`func (o *SubmitSelfServiceRegistrationFlowBody) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *SubmitSelfServiceRegistrationFlowBody) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *SubmitSelfServiceRegistrationFlowBody) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetTraits

`func (o *SubmitSelfServiceRegistrationFlowBody) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *SubmitSelfServiceRegistrationFlowBody) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *SubmitSelfServiceRegistrationFlowBody) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


