# SubmitSelfServiceSettingsFlowWithTOTPMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | CSRFToken is the anti-CSRF token | [optional] 
**Method** | **string** | Method  Should be set to \&quot;totp\&quot; when trying to add, update, or remove a totp pairing. | 
**TotpCode** | Pointer to **string** | ValidationTOTP must contain a valid TOTP based on the | [optional] 
**TotpUnlink** | Pointer to **bool** | UnlinkTOTP if true will remove the TOTP pairing, effectively removing the credential. This can be used to set up a new TOTP device. | [optional] 

## Methods

### NewSubmitSelfServiceSettingsFlowWithTOTPMethodBody

`func NewSubmitSelfServiceSettingsFlowWithTOTPMethodBody(method string, ) *SubmitSelfServiceSettingsFlowWithTOTPMethodBody`

NewSubmitSelfServiceSettingsFlowWithTOTPMethodBody instantiates a new SubmitSelfServiceSettingsFlowWithTOTPMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceSettingsFlowWithTOTPMethodBodyWithDefaults

`func NewSubmitSelfServiceSettingsFlowWithTOTPMethodBodyWithDefaults() *SubmitSelfServiceSettingsFlowWithTOTPMethodBody`

NewSubmitSelfServiceSettingsFlowWithTOTPMethodBodyWithDefaults instantiates a new SubmitSelfServiceSettingsFlowWithTOTPMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetTotpCode

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) GetTotpCode() string`

GetTotpCode returns the TotpCode field if non-nil, zero value otherwise.

### GetTotpCodeOk

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) GetTotpCodeOk() (*string, bool)`

GetTotpCodeOk returns a tuple with the TotpCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotpCode

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) SetTotpCode(v string)`

SetTotpCode sets TotpCode field to given value.

### HasTotpCode

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) HasTotpCode() bool`

HasTotpCode returns a boolean if a field has been set.

### GetTotpUnlink

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) GetTotpUnlink() bool`

GetTotpUnlink returns the TotpUnlink field if non-nil, zero value otherwise.

### GetTotpUnlinkOk

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) GetTotpUnlinkOk() (*bool, bool)`

GetTotpUnlinkOk returns a tuple with the TotpUnlink field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotpUnlink

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) SetTotpUnlink(v bool)`

SetTotpUnlink sets TotpUnlink field to given value.

### HasTotpUnlink

`func (o *SubmitSelfServiceSettingsFlowWithTOTPMethodBody) HasTotpUnlink() bool`

HasTotpUnlink returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


