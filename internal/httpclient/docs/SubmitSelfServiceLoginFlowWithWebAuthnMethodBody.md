# SubmitSelfServiceLoginFlowWithWebAuthnMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Identifier** | Pointer to **string** | Identifier is the email or username of the user trying to log in. This field is only required when using WebAuthn for passwordless login. When using WebAuthn for multi-factor authentication, it is not needed. | [optional] 
**Method** | **string** | Method should be set to \&quot;webAuthn\&quot; when logging in using the WebAuthn strategy. | 
**WebauthnLogin** | Pointer to **string** | Login a WebAuthn Security Key  This must contain the ID of the WebAuthN connection. | [optional] 

## Methods

### NewSubmitSelfServiceLoginFlowWithWebAuthnMethodBody

`func NewSubmitSelfServiceLoginFlowWithWebAuthnMethodBody(method string, ) *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody`

NewSubmitSelfServiceLoginFlowWithWebAuthnMethodBody instantiates a new SubmitSelfServiceLoginFlowWithWebAuthnMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceLoginFlowWithWebAuthnMethodBodyWithDefaults

`func NewSubmitSelfServiceLoginFlowWithWebAuthnMethodBodyWithDefaults() *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody`

NewSubmitSelfServiceLoginFlowWithWebAuthnMethodBodyWithDefaults instantiates a new SubmitSelfServiceLoginFlowWithWebAuthnMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetIdentifier

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) GetIdentifier() string`

GetIdentifier returns the Identifier field if non-nil, zero value otherwise.

### GetIdentifierOk

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) GetIdentifierOk() (*string, bool)`

GetIdentifierOk returns a tuple with the Identifier field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentifier

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) SetIdentifier(v string)`

SetIdentifier sets Identifier field to given value.

### HasIdentifier

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) HasIdentifier() bool`

HasIdentifier returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetWebauthnLogin

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) GetWebauthnLogin() string`

GetWebauthnLogin returns the WebauthnLogin field if non-nil, zero value otherwise.

### GetWebauthnLoginOk

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) GetWebauthnLoginOk() (*string, bool)`

GetWebauthnLoginOk returns a tuple with the WebauthnLogin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebauthnLogin

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) SetWebauthnLogin(v string)`

SetWebauthnLogin sets WebauthnLogin field to given value.

### HasWebauthnLogin

`func (o *SubmitSelfServiceLoginFlowWithWebAuthnMethodBody) HasWebauthnLogin() bool`

HasWebauthnLogin returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


