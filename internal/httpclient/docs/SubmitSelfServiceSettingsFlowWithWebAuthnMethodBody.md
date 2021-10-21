# SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | CSRFToken is the anti-CSRF token | [optional] 
**Method** | **string** | Method  Should be set to \&quot;webauthn\&quot; when trying to add, update, or remove a webAuthn pairing. | 
**WebauthnRegister** | Pointer to **string** | Register a WebAuthn Security Key  It is expected that the JSON returned by the WebAuthn registration process is included here. | [optional] 
**WebauthnRegisterDisplayname** | Pointer to **string** | Name of the WebAuthn Security Key to be Added  A human-readable name for the security key which will be added. | [optional] 
**WebauthnRemove** | Pointer to **string** | Remove a WebAuthn Security Key  This must contain the ID of the WebAuthN connection. | [optional] 

## Methods

### NewSubmitSelfServiceSettingsFlowWithWebAuthnMethodBody

`func NewSubmitSelfServiceSettingsFlowWithWebAuthnMethodBody(method string, ) *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody`

NewSubmitSelfServiceSettingsFlowWithWebAuthnMethodBody instantiates a new SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceSettingsFlowWithWebAuthnMethodBodyWithDefaults

`func NewSubmitSelfServiceSettingsFlowWithWebAuthnMethodBodyWithDefaults() *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody`

NewSubmitSelfServiceSettingsFlowWithWebAuthnMethodBodyWithDefaults instantiates a new SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetWebauthnRegister

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetWebauthnRegister() string`

GetWebauthnRegister returns the WebauthnRegister field if non-nil, zero value otherwise.

### GetWebauthnRegisterOk

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetWebauthnRegisterOk() (*string, bool)`

GetWebauthnRegisterOk returns a tuple with the WebauthnRegister field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebauthnRegister

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) SetWebauthnRegister(v string)`

SetWebauthnRegister sets WebauthnRegister field to given value.

### HasWebauthnRegister

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) HasWebauthnRegister() bool`

HasWebauthnRegister returns a boolean if a field has been set.

### GetWebauthnRegisterDisplayname

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetWebauthnRegisterDisplayname() string`

GetWebauthnRegisterDisplayname returns the WebauthnRegisterDisplayname field if non-nil, zero value otherwise.

### GetWebauthnRegisterDisplaynameOk

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetWebauthnRegisterDisplaynameOk() (*string, bool)`

GetWebauthnRegisterDisplaynameOk returns a tuple with the WebauthnRegisterDisplayname field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebauthnRegisterDisplayname

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) SetWebauthnRegisterDisplayname(v string)`

SetWebauthnRegisterDisplayname sets WebauthnRegisterDisplayname field to given value.

### HasWebauthnRegisterDisplayname

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) HasWebauthnRegisterDisplayname() bool`

HasWebauthnRegisterDisplayname returns a boolean if a field has been set.

### GetWebauthnRemove

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetWebauthnRemove() string`

GetWebauthnRemove returns the WebauthnRemove field if non-nil, zero value otherwise.

### GetWebauthnRemoveOk

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) GetWebauthnRemoveOk() (*string, bool)`

GetWebauthnRemoveOk returns a tuple with the WebauthnRemove field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebauthnRemove

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) SetWebauthnRemove(v string)`

SetWebauthnRemove sets WebauthnRemove field to given value.

### HasWebauthnRemove

`func (o *SubmitSelfServiceSettingsFlowWithWebAuthnMethodBody) HasWebauthnRemove() bool`

HasWebauthnRemove returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


