# CompleteSelfServiceBrowserSettingsOIDCFlowPayload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Flow** | Pointer to **string** | Flow ID is the flow&#39;s ID.  in: query | [optional] 
**Link** | Pointer to **string** | Link this provider  Either this or &#x60;unlink&#x60; must be set.  type: string in: body | [optional] 
**Unlink** | Pointer to **string** | Unlink this provider  Either this or &#x60;link&#x60; must be set.  type: string in: body | [optional] 

## Methods

### NewCompleteSelfServiceBrowserSettingsOIDCFlowPayload

`func NewCompleteSelfServiceBrowserSettingsOIDCFlowPayload() *CompleteSelfServiceBrowserSettingsOIDCFlowPayload`

NewCompleteSelfServiceBrowserSettingsOIDCFlowPayload instantiates a new CompleteSelfServiceBrowserSettingsOIDCFlowPayload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCompleteSelfServiceBrowserSettingsOIDCFlowPayloadWithDefaults

`func NewCompleteSelfServiceBrowserSettingsOIDCFlowPayloadWithDefaults() *CompleteSelfServiceBrowserSettingsOIDCFlowPayload`

NewCompleteSelfServiceBrowserSettingsOIDCFlowPayloadWithDefaults instantiates a new CompleteSelfServiceBrowserSettingsOIDCFlowPayload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFlow

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) GetFlow() string`

GetFlow returns the Flow field if non-nil, zero value otherwise.

### GetFlowOk

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) GetFlowOk() (*string, bool)`

GetFlowOk returns a tuple with the Flow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlow

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) SetFlow(v string)`

SetFlow sets Flow field to given value.

### HasFlow

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) HasFlow() bool`

HasFlow returns a boolean if a field has been set.

### GetLink

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) GetLink() string`

GetLink returns the Link field if non-nil, zero value otherwise.

### GetLinkOk

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) GetLinkOk() (*string, bool)`

GetLinkOk returns a tuple with the Link field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLink

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) SetLink(v string)`

SetLink sets Link field to given value.

### HasLink

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) HasLink() bool`

HasLink returns a boolean if a field has been set.

### GetUnlink

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) GetUnlink() string`

GetUnlink returns the Unlink field if non-nil, zero value otherwise.

### GetUnlinkOk

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) GetUnlinkOk() (*string, bool)`

GetUnlinkOk returns a tuple with the Unlink field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnlink

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) SetUnlink(v string)`

SetUnlink sets Unlink field to given value.

### HasUnlink

`func (o *CompleteSelfServiceBrowserSettingsOIDCFlowPayload) HasUnlink() bool`

HasUnlink returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


