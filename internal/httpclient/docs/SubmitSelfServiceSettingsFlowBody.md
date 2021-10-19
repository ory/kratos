# SubmitSelfServiceSettingsFlowBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | CSRFToken is the anti-CSRF token | [optional] 
**Method** | **string** | Method  Should be set to \&quot;totp\&quot; when trying to add, update, or remove a totp pairing. | 
**Password** | **string** | Password is the updated password | 
**Traits** | **map[string]interface{}** | The identity&#39;s traits  in: body | 
**Flow** | Pointer to **string** | Flow ID is the flow&#39;s ID.  in: query | [optional] 
**Link** | Pointer to **string** | Link this provider  Either this or &#x60;unlink&#x60; must be set.  type: string in: body | [optional] 
**Unlink** | Pointer to **string** | Unlink this provider  Either this or &#x60;link&#x60; must be set.  type: string in: body | [optional] 
**TotpCode** | Pointer to **string** | ValidationTOTP must contain a valid TOTP based on the | [optional] 
**TotpUnlink** | Pointer to **bool** | UnlinkTOTP if true will remove the TOTP pairing, effectively removing the credential. This can be used to set up a new TOTP device. | [optional] 

## Methods

### NewSubmitSelfServiceSettingsFlowBody

`func NewSubmitSelfServiceSettingsFlowBody(method string, password string, traits map[string]interface{}, ) *SubmitSelfServiceSettingsFlowBody`

NewSubmitSelfServiceSettingsFlowBody instantiates a new SubmitSelfServiceSettingsFlowBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceSettingsFlowBodyWithDefaults

`func NewSubmitSelfServiceSettingsFlowBodyWithDefaults() *SubmitSelfServiceSettingsFlowBody`

NewSubmitSelfServiceSettingsFlowBodyWithDefaults instantiates a new SubmitSelfServiceSettingsFlowBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowBody) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceSettingsFlowBody) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceSettingsFlowBody) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceSettingsFlowBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceSettingsFlowBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetPassword

`func (o *SubmitSelfServiceSettingsFlowBody) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *SubmitSelfServiceSettingsFlowBody) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetTraits

`func (o *SubmitSelfServiceSettingsFlowBody) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *SubmitSelfServiceSettingsFlowBody) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.


### GetFlow

`func (o *SubmitSelfServiceSettingsFlowBody) GetFlow() string`

GetFlow returns the Flow field if non-nil, zero value otherwise.

### GetFlowOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetFlowOk() (*string, bool)`

GetFlowOk returns a tuple with the Flow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlow

`func (o *SubmitSelfServiceSettingsFlowBody) SetFlow(v string)`

SetFlow sets Flow field to given value.

### HasFlow

`func (o *SubmitSelfServiceSettingsFlowBody) HasFlow() bool`

HasFlow returns a boolean if a field has been set.

### GetLink

`func (o *SubmitSelfServiceSettingsFlowBody) GetLink() string`

GetLink returns the Link field if non-nil, zero value otherwise.

### GetLinkOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetLinkOk() (*string, bool)`

GetLinkOk returns a tuple with the Link field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLink

`func (o *SubmitSelfServiceSettingsFlowBody) SetLink(v string)`

SetLink sets Link field to given value.

### HasLink

`func (o *SubmitSelfServiceSettingsFlowBody) HasLink() bool`

HasLink returns a boolean if a field has been set.

### GetUnlink

`func (o *SubmitSelfServiceSettingsFlowBody) GetUnlink() string`

GetUnlink returns the Unlink field if non-nil, zero value otherwise.

### GetUnlinkOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetUnlinkOk() (*string, bool)`

GetUnlinkOk returns a tuple with the Unlink field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnlink

`func (o *SubmitSelfServiceSettingsFlowBody) SetUnlink(v string)`

SetUnlink sets Unlink field to given value.

### HasUnlink

`func (o *SubmitSelfServiceSettingsFlowBody) HasUnlink() bool`

HasUnlink returns a boolean if a field has been set.

### GetTotpCode

`func (o *SubmitSelfServiceSettingsFlowBody) GetTotpCode() string`

GetTotpCode returns the TotpCode field if non-nil, zero value otherwise.

### GetTotpCodeOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetTotpCodeOk() (*string, bool)`

GetTotpCodeOk returns a tuple with the TotpCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotpCode

`func (o *SubmitSelfServiceSettingsFlowBody) SetTotpCode(v string)`

SetTotpCode sets TotpCode field to given value.

### HasTotpCode

`func (o *SubmitSelfServiceSettingsFlowBody) HasTotpCode() bool`

HasTotpCode returns a boolean if a field has been set.

### GetTotpUnlink

`func (o *SubmitSelfServiceSettingsFlowBody) GetTotpUnlink() bool`

GetTotpUnlink returns the TotpUnlink field if non-nil, zero value otherwise.

### GetTotpUnlinkOk

`func (o *SubmitSelfServiceSettingsFlowBody) GetTotpUnlinkOk() (*bool, bool)`

GetTotpUnlinkOk returns a tuple with the TotpUnlink field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotpUnlink

`func (o *SubmitSelfServiceSettingsFlowBody) SetTotpUnlink(v bool)`

SetTotpUnlink sets TotpUnlink field to given value.

### HasTotpUnlink

`func (o *SubmitSelfServiceSettingsFlowBody) HasTotpUnlink() bool`

HasTotpUnlink returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


