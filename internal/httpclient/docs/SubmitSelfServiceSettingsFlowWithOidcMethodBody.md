# SubmitSelfServiceSettingsFlowWithOidcMethodBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Flow** | Pointer to **string** | Flow ID is the flow&#39;s ID.  in: query | [optional] 
**Link** | Pointer to **string** | Link this provider  Either this or &#x60;unlink&#x60; must be set.  type: string in: body | [optional] 
**Method** | **string** | Method  Should be set to profile when trying to update a profile. | 
**Unlink** | Pointer to **string** | Unlink this provider  Either this or &#x60;link&#x60; must be set.  type: string in: body | [optional] 

## Methods

### NewSubmitSelfServiceSettingsFlowWithOidcMethodBody

`func NewSubmitSelfServiceSettingsFlowWithOidcMethodBody(method string, ) *SubmitSelfServiceSettingsFlowWithOidcMethodBody`

NewSubmitSelfServiceSettingsFlowWithOidcMethodBody instantiates a new SubmitSelfServiceSettingsFlowWithOidcMethodBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceSettingsFlowWithOidcMethodBodyWithDefaults

`func NewSubmitSelfServiceSettingsFlowWithOidcMethodBodyWithDefaults() *SubmitSelfServiceSettingsFlowWithOidcMethodBody`

NewSubmitSelfServiceSettingsFlowWithOidcMethodBodyWithDefaults instantiates a new SubmitSelfServiceSettingsFlowWithOidcMethodBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFlow

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) GetFlow() string`

GetFlow returns the Flow field if non-nil, zero value otherwise.

### GetFlowOk

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) GetFlowOk() (*string, bool)`

GetFlowOk returns a tuple with the Flow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlow

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) SetFlow(v string)`

SetFlow sets Flow field to given value.

### HasFlow

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) HasFlow() bool`

HasFlow returns a boolean if a field has been set.

### GetLink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) GetLink() string`

GetLink returns the Link field if non-nil, zero value otherwise.

### GetLinkOk

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) GetLinkOk() (*string, bool)`

GetLinkOk returns a tuple with the Link field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) SetLink(v string)`

SetLink sets Link field to given value.

### HasLink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) HasLink() bool`

HasLink returns a boolean if a field has been set.

### GetMethod

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetUnlink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) GetUnlink() string`

GetUnlink returns the Unlink field if non-nil, zero value otherwise.

### GetUnlinkOk

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) GetUnlinkOk() (*string, bool)`

GetUnlinkOk returns a tuple with the Unlink field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnlink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) SetUnlink(v string)`

SetUnlink sets Unlink field to given value.

### HasUnlink

`func (o *SubmitSelfServiceSettingsFlowWithOidcMethodBody) HasUnlink() bool`

HasUnlink returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


