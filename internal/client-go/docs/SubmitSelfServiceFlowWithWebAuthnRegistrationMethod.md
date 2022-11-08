# SubmitSelfServiceFlowWithWebAuthnRegistrationMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WebauthnRegister** | Pointer to **string** | Register a WebAuthn Security Key  It is expected that the JSON returned by the WebAuthn registration process is included here. | [optional] 
**WebauthnRegisterDisplayname** | Pointer to **string** | Name of the WebAuthn Security Key to be Added  A human-readable name for the security key which will be added. | [optional] 

## Methods

### NewSubmitSelfServiceFlowWithWebAuthnRegistrationMethod

`func NewSubmitSelfServiceFlowWithWebAuthnRegistrationMethod() *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod`

NewSubmitSelfServiceFlowWithWebAuthnRegistrationMethod instantiates a new SubmitSelfServiceFlowWithWebAuthnRegistrationMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceFlowWithWebAuthnRegistrationMethodWithDefaults

`func NewSubmitSelfServiceFlowWithWebAuthnRegistrationMethodWithDefaults() *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod`

NewSubmitSelfServiceFlowWithWebAuthnRegistrationMethodWithDefaults instantiates a new SubmitSelfServiceFlowWithWebAuthnRegistrationMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWebauthnRegister

`func (o *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod) GetWebauthnRegister() string`

GetWebauthnRegister returns the WebauthnRegister field if non-nil, zero value otherwise.

### GetWebauthnRegisterOk

`func (o *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod) GetWebauthnRegisterOk() (*string, bool)`

GetWebauthnRegisterOk returns a tuple with the WebauthnRegister field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebauthnRegister

`func (o *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod) SetWebauthnRegister(v string)`

SetWebauthnRegister sets WebauthnRegister field to given value.

### HasWebauthnRegister

`func (o *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod) HasWebauthnRegister() bool`

HasWebauthnRegister returns a boolean if a field has been set.

### GetWebauthnRegisterDisplayname

`func (o *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod) GetWebauthnRegisterDisplayname() string`

GetWebauthnRegisterDisplayname returns the WebauthnRegisterDisplayname field if non-nil, zero value otherwise.

### GetWebauthnRegisterDisplaynameOk

`func (o *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod) GetWebauthnRegisterDisplaynameOk() (*string, bool)`

GetWebauthnRegisterDisplaynameOk returns a tuple with the WebauthnRegisterDisplayname field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebauthnRegisterDisplayname

`func (o *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod) SetWebauthnRegisterDisplayname(v string)`

SetWebauthnRegisterDisplayname sets WebauthnRegisterDisplayname field to given value.

### HasWebauthnRegisterDisplayname

`func (o *SubmitSelfServiceFlowWithWebAuthnRegistrationMethod) HasWebauthnRegisterDisplayname() bool`

HasWebauthnRegisterDisplayname returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


