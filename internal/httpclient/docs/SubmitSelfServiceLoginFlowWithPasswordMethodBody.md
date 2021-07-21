# SubmitSelfServiceLoginFlowWithPasswordMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Method** | **string** | Method should be set to \&quot;password\&quot; when logging in using the identifier and password strategy. | 
**Password** | **string** | The user&#39;s password. | 
**PasswordIdentifier** | **string** | Identifier is the email or username of the user trying to log in. | 

## Methods

### NewSubmitSelfServiceLoginFlowWithPasswordMethodBody

`func NewSubmitSelfServiceLoginFlowWithPasswordMethodBody(method string, password string, passwordIdentifier string, ) *SubmitSelfServiceLoginFlowWithPasswordMethodBody`

NewSubmitSelfServiceLoginFlowWithPasswordMethodBody instantiates a new SubmitSelfServiceLoginFlowWithPasswordMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceLoginFlowWithPasswordMethodBodyWithDefaults

`func NewSubmitSelfServiceLoginFlowWithPasswordMethodBodyWithDefaults() *SubmitSelfServiceLoginFlowWithPasswordMethodBody`

NewSubmitSelfServiceLoginFlowWithPasswordMethodBodyWithDefaults instantiates a new SubmitSelfServiceLoginFlowWithPasswordMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetPassword

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetPasswordIdentifier

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) GetPasswordIdentifier() string`

GetPasswordIdentifier returns the PasswordIdentifier field if non-nil, zero value otherwise.

### GetPasswordIdentifierOk

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) GetPasswordIdentifierOk() (*string, bool)`

GetPasswordIdentifierOk returns a tuple with the PasswordIdentifier field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPasswordIdentifier

`func (o *SubmitSelfServiceLoginFlowWithPasswordMethodBody) SetPasswordIdentifier(v string)`

SetPasswordIdentifier sets PasswordIdentifier field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


