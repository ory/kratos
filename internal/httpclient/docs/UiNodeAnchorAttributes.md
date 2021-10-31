# UiNodeAnchorAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Href** | **string** | The link&#39;s href (destination) URL.  format: uri | 
**Id** | **string** | A unique identifier | 
**NodeType** | **string** |  | 
**Title** | [**UiText**](UiText.md) |  | 

## Methods

### NewUiNodeAnchorAttributes

`func NewUiNodeAnchorAttributes(href string, id string, nodeType string, title UiText, ) *UiNodeAnchorAttributes`

NewUiNodeAnchorAttributes instantiates a new UiNodeAnchorAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUiNodeAnchorAttributesWithDefaults

`func NewUiNodeAnchorAttributesWithDefaults() *UiNodeAnchorAttributes`

NewUiNodeAnchorAttributesWithDefaults instantiates a new UiNodeAnchorAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHref

`func (o *UiNodeAnchorAttributes) GetHref() string`

GetHref returns the Href field if non-nil, zero value otherwise.

### GetHrefOk

`func (o *UiNodeAnchorAttributes) GetHrefOk() (*string, bool)`

GetHrefOk returns a tuple with the Href field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHref

`func (o *UiNodeAnchorAttributes) SetHref(v string)`

SetHref sets Href field to given value.


### GetId

`func (o *UiNodeAnchorAttributes) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *UiNodeAnchorAttributes) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *UiNodeAnchorAttributes) SetId(v string)`

SetId sets Id field to given value.


### GetNodeType

`func (o *UiNodeAnchorAttributes) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *UiNodeAnchorAttributes) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *UiNodeAnchorAttributes) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.


### GetTitle

`func (o *UiNodeAnchorAttributes) GetTitle() UiText`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *UiNodeAnchorAttributes) GetTitleOk() (*UiText, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *UiNodeAnchorAttributes) SetTitle(v UiText)`

SetTitle sets Title field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


