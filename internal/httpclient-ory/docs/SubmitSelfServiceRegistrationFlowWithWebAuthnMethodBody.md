# SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | CSRFToken is the anti-CSRF token | [optional] 
**Method** | **string** | Method  Should be set to \&quot;webauthn\&quot; when trying to add, update, or remove a webAuthn pairing. | 
**Traits** | **map[string]interface{}** | The identity&#39;s traits | 
**WebauthnRegister** | Pointer to **string** | Register a WebAuthn Security Key  It is expected that the JSON returned by the WebAuthn registration process is included here. | [optional] 
**WebauthnRegisterDisplayname** | Pointer to **string** | Name of the WebAuthn Security Key to be Added  A human-readable name for the security key which will be added. | [optional] 

## Methods

### NewSubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody

`func NewSubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody(method string, traits map[string]interface{}, ) *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody`

NewSubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody instantiates a new SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceRegistrationFlowWithWebAuthnMethodBodyWithDefaults

`func NewSubmitSelfServiceRegistrationFlowWithWebAuthnMethodBodyWithDefaults() *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody`

NewSubmitSelfServiceRegistrationFlowWithWebAuthnMethodBodyWithDefaults instantiates a new SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetTraits

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.


### GetWebauthnRegister

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetWebauthnRegister() string`

GetWebauthnRegister returns the WebauthnRegister field if non-nil, zero value otherwise.

### GetWebauthnRegisterOk

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetWebauthnRegisterOk() (*string, bool)`

GetWebauthnRegisterOk returns a tuple with the WebauthnRegister field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebauthnRegister

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) SetWebauthnRegister(v string)`

SetWebauthnRegister sets WebauthnRegister field to given value.

### HasWebauthnRegister

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) HasWebauthnRegister() bool`

HasWebauthnRegister returns a boolean if a field has been set.

### GetWebauthnRegisterDisplayname

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetWebauthnRegisterDisplayname() string`

GetWebauthnRegisterDisplayname returns the WebauthnRegisterDisplayname field if non-nil, zero value otherwise.

### GetWebauthnRegisterDisplaynameOk

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) GetWebauthnRegisterDisplaynameOk() (*string, bool)`

GetWebauthnRegisterDisplaynameOk returns a tuple with the WebauthnRegisterDisplayname field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebauthnRegisterDisplayname

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) SetWebauthnRegisterDisplayname(v string)`

SetWebauthnRegisterDisplayname sets WebauthnRegisterDisplayname field to given value.

### HasWebauthnRegisterDisplayname

`func (o *SubmitSelfServiceRegistrationFlowWithWebAuthnMethodBody) HasWebauthnRegisterDisplayname() bool`

HasWebauthnRegisterDisplayname returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


