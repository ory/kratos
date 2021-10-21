# SubmitSelfServiceLoginFlowWithTotpMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Method** | **string** | Method should be set to \&quot;totp\&quot; when logging in using the TOTP strategy. | 
**TotpCode** | **string** | The TOTP code. | 

## Methods

### NewSubmitSelfServiceLoginFlowWithTotpMethodBody

`func NewSubmitSelfServiceLoginFlowWithTotpMethodBody(method string, totpCode string, ) *SubmitSelfServiceLoginFlowWithTotpMethodBody`

NewSubmitSelfServiceLoginFlowWithTotpMethodBody instantiates a new SubmitSelfServiceLoginFlowWithTotpMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceLoginFlowWithTotpMethodBodyWithDefaults

`func NewSubmitSelfServiceLoginFlowWithTotpMethodBodyWithDefaults() *SubmitSelfServiceLoginFlowWithTotpMethodBody`

NewSubmitSelfServiceLoginFlowWithTotpMethodBodyWithDefaults instantiates a new SubmitSelfServiceLoginFlowWithTotpMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetTotpCode

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) GetTotpCode() string`

GetTotpCode returns the TotpCode field if non-nil, zero value otherwise.

### GetTotpCodeOk

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) GetTotpCodeOk() (*string, bool)`

GetTotpCodeOk returns a tuple with the TotpCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotpCode

`func (o *SubmitSelfServiceLoginFlowWithTotpMethodBody) SetTotpCode(v string)`

SetTotpCode sets TotpCode field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


