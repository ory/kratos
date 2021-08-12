# SubmitSelfServiceLoginFlowWithTOTPMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Method** | **string** | Method should be set to \&quot;totp\&quot; when logging in using the TOTP strategy. | 
**TotpCode** | **string** | The TOTP code. | 

## Methods

### NewSubmitSelfServiceLoginFlowWithTOTPMethodBody

`func NewSubmitSelfServiceLoginFlowWithTOTPMethodBody(method string, totpCode string, ) *SubmitSelfServiceLoginFlowWithTOTPMethodBody`

NewSubmitSelfServiceLoginFlowWithTOTPMethodBody instantiates a new SubmitSelfServiceLoginFlowWithTOTPMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceLoginFlowWithTOTPMethodBodyWithDefaults

`func NewSubmitSelfServiceLoginFlowWithTOTPMethodBodyWithDefaults() *SubmitSelfServiceLoginFlowWithTOTPMethodBody`

NewSubmitSelfServiceLoginFlowWithTOTPMethodBodyWithDefaults instantiates a new SubmitSelfServiceLoginFlowWithTOTPMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetTotpCode

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) GetTotpCode() string`

GetTotpCode returns the TotpCode field if non-nil, zero value otherwise.

### GetTotpCodeOk

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) GetTotpCodeOk() (*string, bool)`

GetTotpCodeOk returns a tuple with the TotpCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotpCode

`func (o *SubmitSelfServiceLoginFlowWithTOTPMethodBody) SetTotpCode(v string)`

SetTotpCode sets TotpCode field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


