# CompleteSelfServiceLoginFlowWithPasswordMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Identifier** | Pointer to **string** | Identifier is the email or username of the user trying to log in. | [optional] 
**Password** | Pointer to **string** | The user&#39;s password. | [optional] 

## Methods

### NewCompleteSelfServiceLoginFlowWithPasswordMethod

`func NewCompleteSelfServiceLoginFlowWithPasswordMethod() *CompleteSelfServiceLoginFlowWithPasswordMethod`

NewCompleteSelfServiceLoginFlowWithPasswordMethod instantiates a new CompleteSelfServiceLoginFlowWithPasswordMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCompleteSelfServiceLoginFlowWithPasswordMethodWithDefaults

`func NewCompleteSelfServiceLoginFlowWithPasswordMethodWithDefaults() *CompleteSelfServiceLoginFlowWithPasswordMethod`

NewCompleteSelfServiceLoginFlowWithPasswordMethodWithDefaults instantiates a new CompleteSelfServiceLoginFlowWithPasswordMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetIdentifier

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) GetIdentifier() string`

GetIdentifier returns the Identifier field if non-nil, zero value otherwise.

### GetIdentifierOk

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) GetIdentifierOk() (*string, bool)`

GetIdentifierOk returns a tuple with the Identifier field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentifier

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) SetIdentifier(v string)`

SetIdentifier sets Identifier field to given value.

### HasIdentifier

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) HasIdentifier() bool`

HasIdentifier returns a boolean if a field has been set.

### GetPassword

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *CompleteSelfServiceLoginFlowWithPasswordMethod) HasPassword() bool`

HasPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


