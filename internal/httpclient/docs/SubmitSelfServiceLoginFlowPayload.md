# SubmitSelfServiceLoginFlowPayload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Method** | Pointer to **string** | Method should be set to \&quot;password\&quot; when logging in using the identifier and password strategy. | [optional] 
**Password** | Pointer to **string** | The user&#39;s password. | [optional] 
**PasswordIdentifier** | Pointer to **string** | Identifier is the email or username of the user trying to log in. | [optional] 

## Methods

### NewSubmitSelfServiceLoginFlowPayload

`func NewSubmitSelfServiceLoginFlowPayload() *SubmitSelfServiceLoginFlowPayload`

NewSubmitSelfServiceLoginFlowPayload instantiates a new SubmitSelfServiceLoginFlowPayload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceLoginFlowPayloadWithDefaults

`func NewSubmitSelfServiceLoginFlowPayloadWithDefaults() *SubmitSelfServiceLoginFlowPayload`

NewSubmitSelfServiceLoginFlowPayloadWithDefaults instantiates a new SubmitSelfServiceLoginFlowPayload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceLoginFlowPayload) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceLoginFlowPayload) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceLoginFlowPayload) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceLoginFlowPayload) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceLoginFlowPayload) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceLoginFlowPayload) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceLoginFlowPayload) SetMethod(v string)`

SetMethod sets Method field to given value.

### HasMethod

`func (o *SubmitSelfServiceLoginFlowPayload) HasMethod() bool`

HasMethod returns a boolean if a field has been set.

### GetPassword

`func (o *SubmitSelfServiceLoginFlowPayload) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *SubmitSelfServiceLoginFlowPayload) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *SubmitSelfServiceLoginFlowPayload) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *SubmitSelfServiceLoginFlowPayload) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetPasswordIdentifier

`func (o *SubmitSelfServiceLoginFlowPayload) GetPasswordIdentifier() string`

GetPasswordIdentifier returns the PasswordIdentifier field if non-nil, zero value otherwise.

### GetPasswordIdentifierOk

`func (o *SubmitSelfServiceLoginFlowPayload) GetPasswordIdentifierOk() (*string, bool)`

GetPasswordIdentifierOk returns a tuple with the PasswordIdentifier field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPasswordIdentifier

`func (o *SubmitSelfServiceLoginFlowPayload) SetPasswordIdentifier(v string)`

SetPasswordIdentifier sets PasswordIdentifier field to given value.

### HasPasswordIdentifier

`func (o *SubmitSelfServiceLoginFlowPayload) HasPasswordIdentifier() bool`

HasPasswordIdentifier returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


