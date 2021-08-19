# UiNodeImageAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Height** | Pointer to **int64** | Height of the image | [optional] 
**Id** | **string** | A unique identifier | 
**Src** | **string** | The image&#39;s source URL.  format: uri | 
**Width** | Pointer to **int64** | Width of the image | [optional] 

## Methods

### NewUiNodeImageAttributes

`func NewUiNodeImageAttributes(id string, src string, ) *UiNodeImageAttributes`

NewUiNodeImageAttributes instantiates a new UiNodeImageAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUiNodeImageAttributesWithDefaults

`func NewUiNodeImageAttributesWithDefaults() *UiNodeImageAttributes`

NewUiNodeImageAttributesWithDefaults instantiates a new UiNodeImageAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHeight

`func (o *UiNodeImageAttributes) GetHeight() int64`

GetHeight returns the Height field if non-nil, zero value otherwise.

### GetHeightOk

`func (o *UiNodeImageAttributes) GetHeightOk() (*int64, bool)`

GetHeightOk returns a tuple with the Height field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHeight

`func (o *UiNodeImageAttributes) SetHeight(v int64)`

SetHeight sets Height field to given value.

### HasHeight

`func (o *UiNodeImageAttributes) HasHeight() bool`

HasHeight returns a boolean if a field has been set.

### GetId

`func (o *UiNodeImageAttributes) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *UiNodeImageAttributes) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *UiNodeImageAttributes) SetId(v string)`

SetId sets Id field to given value.


### GetSrc

`func (o *UiNodeImageAttributes) GetSrc() string`

GetSrc returns the Src field if non-nil, zero value otherwise.

### GetSrcOk

`func (o *UiNodeImageAttributes) GetSrcOk() (*string, bool)`

GetSrcOk returns a tuple with the Src field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSrc

`func (o *UiNodeImageAttributes) SetSrc(v string)`

SetSrc sets Src field to given value.


### GetWidth

`func (o *UiNodeImageAttributes) GetWidth() int64`

GetWidth returns the Width field if non-nil, zero value otherwise.

### GetWidthOk

`func (o *UiNodeImageAttributes) GetWidthOk() (*int64, bool)`

GetWidthOk returns a tuple with the Width field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWidth

`func (o *UiNodeImageAttributes) SetWidth(v int64)`

SetWidth sets Width field to given value.

### HasWidth

`func (o *UiNodeImageAttributes) HasWidth() bool`

HasWidth returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


