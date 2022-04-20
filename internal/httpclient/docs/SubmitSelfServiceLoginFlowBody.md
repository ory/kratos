# SubmitSelfServiceLoginFlowBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Identifier** | **string** | Identifier is the email or username of the user trying to log in. This field is only required when using WebAuthn for passwordless login. When using WebAuthn for multi-factor authentication, it is not needed. | 
**Method** | **string** | Method should be set to \&quot;lookup_secret\&quot; when logging in using the lookup_secret strategy. | 
**Password** | **string** | The user&#39;s password. | 
**PasswordIdentifier** | Pointer to **string** | Identifier is the email or username of the user trying to log in. This field is deprecated! | [optional] 
**Provider** | **string** | The provider to register with | 
**Traits** | Pointer to **map[string]interface{}** | The identity traits. This is a placeholder for the registration flow. | [optional] 
**TotpCode** | **string** | The TOTP code. | 
**WebauthnLogin** | Pointer to **string** | Login a WebAuthn Security Key  This must contain the ID of the WebAuthN connection. | [optional] 
**LookupSecret** | **string** | The lookup secret. | 

## Methods

### NewSubmitSelfServiceLoginFlowBody

`func NewSubmitSelfServiceLoginFlowBody(identifier string, method string, password string, provider string, totpCode string, lookupSecret string, ) *SubmitSelfServiceLoginFlowBody`

NewSubmitSelfServiceLoginFlowBody instantiates a new SubmitSelfServiceLoginFlowBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceLoginFlowBodyWithDefaults

`func NewSubmitSelfServiceLoginFlowBodyWithDefaults() *SubmitSelfServiceLoginFlowBody`

NewSubmitSelfServiceLoginFlowBodyWithDefaults instantiates a new SubmitSelfServiceLoginFlowBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceLoginFlowBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceLoginFlowBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceLoginFlowBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceLoginFlowBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetIdentifier

`func (o *SubmitSelfServiceLoginFlowBody) GetIdentifier() string`

GetIdentifier returns the Identifier field if non-nil, zero value otherwise.

### GetIdentifierOk

`func (o *SubmitSelfServiceLoginFlowBody) GetIdentifierOk() (*string, bool)`

GetIdentifierOk returns a tuple with the Identifier field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentifier

`func (o *SubmitSelfServiceLoginFlowBody) SetIdentifier(v string)`

SetIdentifier sets Identifier field to given value.


### GetMethod

`func (o *SubmitSelfServiceLoginFlowBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceLoginFlowBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceLoginFlowBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetPassword

`func (o *SubmitSelfServiceLoginFlowBody) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *SubmitSelfServiceLoginFlowBody) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *SubmitSelfServiceLoginFlowBody) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetPasswordIdentifier

`func (o *SubmitSelfServiceLoginFlowBody) GetPasswordIdentifier() string`

GetPasswordIdentifier returns the PasswordIdentifier field if non-nil, zero value otherwise.

### GetPasswordIdentifierOk

`func (o *SubmitSelfServiceLoginFlowBody) GetPasswordIdentifierOk() (*string, bool)`

GetPasswordIdentifierOk returns a tuple with the PasswordIdentifier field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPasswordIdentifier

`func (o *SubmitSelfServiceLoginFlowBody) SetPasswordIdentifier(v string)`

SetPasswordIdentifier sets PasswordIdentifier field to given value.

### HasPasswordIdentifier

`func (o *SubmitSelfServiceLoginFlowBody) HasPasswordIdentifier() bool`

HasPasswordIdentifier returns a boolean if a field has been set.

### GetProvider

`func (o *SubmitSelfServiceLoginFlowBody) GetProvider() string`

GetProvider returns the Provider field if non-nil, zero value otherwise.

### GetProviderOk

`func (o *SubmitSelfServiceLoginFlowBody) GetProviderOk() (*string, bool)`

GetProviderOk returns a tuple with the Provider field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProvider

`func (o *SubmitSelfServiceLoginFlowBody) SetProvider(v string)`

SetProvider sets Provider field to given value.


### GetTraits

`func (o *SubmitSelfServiceLoginFlowBody) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *SubmitSelfServiceLoginFlowBody) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *SubmitSelfServiceLoginFlowBody) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.

### HasTraits

`func (o *SubmitSelfServiceLoginFlowBody) HasTraits() bool`

HasTraits returns a boolean if a field has been set.

### GetTotpCode

`func (o *SubmitSelfServiceLoginFlowBody) GetTotpCode() string`

GetTotpCode returns the TotpCode field if non-nil, zero value otherwise.

### GetTotpCodeOk

`func (o *SubmitSelfServiceLoginFlowBody) GetTotpCodeOk() (*string, bool)`

GetTotpCodeOk returns a tuple with the TotpCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotpCode

`func (o *SubmitSelfServiceLoginFlowBody) SetTotpCode(v string)`

SetTotpCode sets TotpCode field to given value.


### GetWebauthnLogin

`func (o *SubmitSelfServiceLoginFlowBody) GetWebauthnLogin() string`

GetWebauthnLogin returns the WebauthnLogin field if non-nil, zero value otherwise.

### GetWebauthnLoginOk

`func (o *SubmitSelfServiceLoginFlowBody) GetWebauthnLoginOk() (*string, bool)`

GetWebauthnLoginOk returns a tuple with the WebauthnLogin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWebauthnLogin

`func (o *SubmitSelfServiceLoginFlowBody) SetWebauthnLogin(v string)`

SetWebauthnLogin sets WebauthnLogin field to given value.

### HasWebauthnLogin

`func (o *SubmitSelfServiceLoginFlowBody) HasWebauthnLogin() bool`

HasWebauthnLogin returns a boolean if a field has been set.

### GetLookupSecret

`func (o *SubmitSelfServiceLoginFlowBody) GetLookupSecret() string`

GetLookupSecret returns the LookupSecret field if non-nil, zero value otherwise.

### GetLookupSecretOk

`func (o *SubmitSelfServiceLoginFlowBody) GetLookupSecretOk() (*string, bool)`

GetLookupSecretOk returns a tuple with the LookupSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLookupSecret

`func (o *SubmitSelfServiceLoginFlowBody) SetLookupSecret(v string)`

SetLookupSecret sets LookupSecret field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


