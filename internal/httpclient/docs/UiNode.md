# UiNode

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | [**UiNodeAttributes**](uiNodeAttributes.md) |  | 
**Group** | Pointer to **string** |  | [optional] 
**Messages** | Pointer to [**[]UiText**](UiText.md) |  | [optional] 
**Type** | **string** |  | 

## Methods

### NewUiNode

`func NewUiNode(attributes UiNodeAttributes, type_ string, ) *UiNode`

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

### HasGroup

`func (o *UiNode) HasGroup() bool`

HasGroup returns a boolean if a field has been set.

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

### HasMessages

`func (o *UiNode) HasMessages() bool`

HasMessages returns a boolean if a field has been set.

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


