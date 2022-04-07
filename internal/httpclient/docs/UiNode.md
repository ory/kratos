# UiNode

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | [**UiNodeAttributes**](UiNodeAttributes.md) |  | 
**Group** | **string** | Group specifies which group (e.g. password authenticator) this node belongs to. | 
**Messages** | [**[]UiText**](UiText.md) |  | 
**Meta** | [**UiNodeMeta**](UiNodeMeta.md) |  | 
**Type** | **string** | The node&#39;s type | 

## Methods

### NewUiNode

`func NewUiNode(attributes UiNodeAttributes, group string, messages []UiText, meta UiNodeMeta, type_ string, ) *UiNode`

NewUiNode instantiates a new UiNode object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUiNodeWithDefaults

`func NewUiNodeWithDefaults() *UiNode`

NewUiNodeWithDefaults instantiates a new UiNode object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *UiNode) GetAttributes() UiNodeAttributes`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *UiNode) GetAttributesOk() (*UiNodeAttributes, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *UiNode) SetAttributes(v UiNodeAttributes)`

SetAttributes sets Attributes field to given value.


### GetGroup

`func (o *UiNode) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *UiNode) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *UiNode) SetGroup(v string)`

SetGroup sets Group field to given value.


### GetMessages

`func (o *UiNode) GetMessages() []UiText`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *UiNode) GetMessagesOk() (*[]UiText, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *UiNode) SetMessages(v []UiText)`

SetMessages sets Messages field to given value.


### GetMeta

`func (o *UiNode) GetMeta() UiNodeMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *UiNode) GetMetaOk() (*UiNodeMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *UiNode) SetMeta(v UiNodeMeta)`

SetMeta sets Meta field to given value.


### GetType

`func (o *UiNode) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *UiNode) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *UiNode) SetType(v string)`

SetType sets Type field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


