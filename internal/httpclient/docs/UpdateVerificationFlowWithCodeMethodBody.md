# UpdateVerificationFlowWithCodeMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **string** | The verification code | [optional] 
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Email** | Pointer to **string** | Email to Verify  Needs to be set when initiating the flow. If the email is a registered verification email, a verification link will be sent. If the email is not known, a email with details on what happened will be sent instead.  format: email | [optional] 
**Flow** | Pointer to **string** | The id of the flow | [optional] 
**Method** | Pointer to **string** | Method is the recovery method | [optional] 

## Methods

### NewUpdateVerificationFlowWithCodeMethodBody

`func NewUpdateVerificationFlowWithCodeMethodBody() *UpdateVerificationFlowWithCodeMethodBody`

NewUpdateVerificationFlowWithCodeMethodBody instantiates a new UpdateVerificationFlowWithCodeMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateVerificationFlowWithCodeMethodBodyWithDefaults

`func NewUpdateVerificationFlowWithCodeMethodBodyWithDefaults() *UpdateVerificationFlowWithCodeMethodBody`

NewUpdateVerificationFlowWithCodeMethodBodyWithDefaults instantiates a new UpdateVerificationFlowWithCodeMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *UpdateVerificationFlowWithCodeMethodBody) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *UpdateVerificationFlowWithCodeMethodBody) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetCsrfToken

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *UpdateVerificationFlowWithCodeMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *UpdateVerificationFlowWithCodeMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetEmail

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UpdateVerificationFlowWithCodeMethodBody) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *UpdateVerificationFlowWithCodeMethodBody) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetFlow

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetFlow() string`

GetFlow returns the Flow field if non-nil, zero value otherwise.

### GetFlowOk

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetFlowOk() (*string, bool)`

GetFlowOk returns a tuple with the Flow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlow

`func (o *UpdateVerificationFlowWithCodeMethodBody) SetFlow(v string)`

SetFlow sets Flow field to given value.

### HasFlow

`func (o *UpdateVerificationFlowWithCodeMethodBody) HasFlow() bool`

HasFlow returns a boolean if a field has been set.

### GetMethod

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *UpdateVerificationFlowWithCodeMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *UpdateVerificationFlowWithCodeMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.

### HasMethod

`func (o *UpdateVerificationFlowWithCodeMethodBody) HasMethod() bool`

HasMethod returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


