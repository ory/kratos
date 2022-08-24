# SubmitSelfServiceRecoveryFlowWithCodeMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **string** | Code from recovery email  Sent to the user once a recovery has been initiated and is used to prove that the user is in possession of the email | [optional] 
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Email** | Pointer to **string** | Email to Recover  Needs to be set when initiating the flow. If the email is a registered recovery email, a recovery link will be sent. If the email is not known, a email with details on what happened will be sent instead.  format: email | [optional] 
**Method** | **string** | Method supports &#x60;link&#x60; and &#x60;code&#x60; only right now. | 

## Methods

### NewSubmitSelfServiceRecoveryFlowWithCodeMethodBody

`func NewSubmitSelfServiceRecoveryFlowWithCodeMethodBody(method string, ) *SubmitSelfServiceRecoveryFlowWithCodeMethodBody`

NewSubmitSelfServiceRecoveryFlowWithCodeMethodBody instantiates a new SubmitSelfServiceRecoveryFlowWithCodeMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceRecoveryFlowWithCodeMethodBodyWithDefaults

`func NewSubmitSelfServiceRecoveryFlowWithCodeMethodBodyWithDefaults() *SubmitSelfServiceRecoveryFlowWithCodeMethodBody`

NewSubmitSelfServiceRecoveryFlowWithCodeMethodBodyWithDefaults instantiates a new SubmitSelfServiceRecoveryFlowWithCodeMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetCsrfToken

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetEmail

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceRecoveryFlowWithCodeMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


