# UiText

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Context** | Pointer to **map[string]interface{}** | The message&#39;s context. Useful when customizing messages. | [optional] 
**Id** | **int64** |  | 
**Text** | **string** | The message text. Written in american english. | 
**Type** | **string** |  | 

## Methods

### NewUiText

`func NewUiText(id int64, text string, type_ string, ) *UiText`

NewUiText instantiates a new UiText object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUiTextWithDefaults

`func NewUiTextWithDefaults() *UiText`

NewUiTextWithDefaults instantiates a new UiText object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContext

`func (o *UiText) GetContext() map[string]interface{}`

GetContext returns the Context field if non-nil, zero value otherwise.

### GetContextOk

`func (o *UiText) GetContextOk() (*map[string]interface{}, bool)`

GetContextOk returns a tuple with the Context field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContext

`func (o *UiText) SetContext(v map[string]interface{})`

SetContext sets Context field to given value.

### HasContext

`func (o *UiText) HasContext() bool`

HasContext returns a boolean if a field has been set.

### GetId

`func (o *UiText) GetId() int64`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *UiText) GetIdOk() (*int64, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *UiText) SetId(v int64)`

SetId sets Id field to given value.


### GetText

`func (o *UiText) GetText() string`

GetText returns the Text field if non-nil, zero value otherwise.

### GetTextOk

`func (o *UiText) GetTextOk() (*string, bool)`

GetTextOk returns a tuple with the Text field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetText

`func (o *UiText) SetText(v string)`

SetText sets Text field to given value.


### GetType

`func (o *UiText) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *UiText) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *UiText) SetType(v string)`

SetType sets Type field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


