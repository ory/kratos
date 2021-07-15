# SubmitSelfServiceSettingsFlowWithOidcMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Flow** | Pointer to **string** | Flow ID is the flow&#39;s ID.  in: query | [optional] 
**Link** | Pointer to **string** | Link this provider  Either this or &#x60;unlink&#x60; must be set.  type: string in: body | [optional] 
**Unlink** | Pointer to **string** | Unlink this provider  Either this or &#x60;link&#x60; must be set.  type: string in: body | [optional] 

## Methods

### NewSubmitSelfServiceSettingsFlowWithOidcMethod

`func NewSubmitSelfServiceSettingsFlowWithOidcMethod() *SubmitSelfServiceSettingsFlowWithOidcMethod`

NewSubmitSelfServiceSettingsFlowWithOidcMethod instantiates a new SubmitSelfServiceSettingsFlowWithOidcMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceSettingsFlowWithOidcMethodWithDefaults

`func NewSubmitSelfServiceSettingsFlowWithOidcMethodWithDefaults() *SubmitSelfServiceSettingsFlowWithOidcMethod`

NewSubmitSelfServiceSettingsFlowWithOidcMethodWithDefaults instantiates a new SubmitSelfServiceSettingsFlowWithOidcMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFlow

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) GetFlow() string`

GetFlow returns the Flow field if non-nil, zero value otherwise.

### GetFlowOk

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) GetFlowOk() (*string, bool)`

GetFlowOk returns a tuple with the Flow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlow

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) SetFlow(v string)`

SetFlow sets Flow field to given value.

### HasFlow

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) HasFlow() bool`

HasFlow returns a boolean if a field has been set.

### GetLink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) GetLink() string`

GetLink returns the Link field if non-nil, zero value otherwise.

### GetLinkOk

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) GetLinkOk() (*string, bool)`

GetLinkOk returns a tuple with the Link field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) SetLink(v string)`

SetLink sets Link field to given value.

### HasLink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) HasLink() bool`

HasLink returns a boolean if a field has been set.

### GetUnlink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) GetUnlink() string`

GetUnlink returns the Unlink field if non-nil, zero value otherwise.

### GetUnlinkOk

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) GetUnlinkOk() (*string, bool)`

GetUnlinkOk returns a tuple with the Unlink field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnlink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) SetUnlink(v string)`

SetUnlink sets Unlink field to given value.

### HasUnlink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethod) HasUnlink() bool`

HasUnlink returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


