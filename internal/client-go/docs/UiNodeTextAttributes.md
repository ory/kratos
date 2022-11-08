# UiNodeTextAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | A unique identifier | 
**NodeType** | **string** | NodeType represents this node&#39;s types. It is a mirror of &#x60;node.type&#x60; and is primarily used to allow compatibility with OpenAPI 3.0.  In this struct it technically always is \&quot;text\&quot;. | 
**Text** | [**UiText**](UiText.md) |  | 

## Methods

### NewUiNodeTextAttributes

`func NewUiNodeTextAttributes(id string, nodeType string, text UiText, ) *UiNodeTextAttributes`

NewUiNodeTextAttributes instantiates a new UiNodeTextAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUiNodeTextAttributesWithDefaults

`func NewUiNodeTextAttributesWithDefaults() *UiNodeTextAttributes`

NewUiNodeTextAttributesWithDefaults instantiates a new UiNodeTextAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *UiNodeTextAttributes) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *UiNodeTextAttributes) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *UiNodeTextAttributes) SetId(v string)`

SetId sets Id field to given value.


### GetNodeType

`func (o *UiNodeTextAttributes) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *UiNodeTextAttributes) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *UiNodeTextAttributes) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.


### GetText

`func (o *UiNodeTextAttributes) GetText() UiText`

GetText returns the Text field if non-nil, zero value otherwise.

### GetTextOk

`func (o *UiNodeTextAttributes) GetTextOk() (*UiText, bool)`

GetTextOk returns a tuple with the Text field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetText

`func (o *UiNodeTextAttributes) SetText(v UiText)`

SetText sets Text field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


