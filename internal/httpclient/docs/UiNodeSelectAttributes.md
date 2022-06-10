# UiNodeSelectAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Disabled** | **bool** | Sets the input&#39;s disabled field to true or false. | 
**Label** | Pointer to [**UiText**](UiText.md) |  | [optional] 
**Name** | **string** | The input&#39;s element name. | 
**NodeType** | **string** | NodeType represents this node&#39;s types. It is a mirror of &#x60;node.type&#x60; and is primarily used to allow compatibility with OpenAPI 3.0.  In this struct it technically always is \&quot;select\&quot;. | 
**Options** | [**[]UiNodeSelectOption**](UiNodeSelectOption.md) | Options represents the options for a select node. | 
**Required** | Pointer to **bool** | Mark this input field as required. | [optional] 
**Value** | Pointer to **map[string]interface{}** | The input&#39;s value. | [optional] 

## Methods

### NewUiNodeSelectAttributes

`func NewUiNodeSelectAttributes(disabled bool, name string, nodeType string, options []UiNodeSelectOption, ) *UiNodeSelectAttributes`

NewUiNodeSelectAttributes instantiates a new UiNodeSelectAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUiNodeSelectAttributesWithDefaults

`func NewUiNodeSelectAttributesWithDefaults() *UiNodeSelectAttributes`

NewUiNodeSelectAttributesWithDefaults instantiates a new UiNodeSelectAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDisabled

`func (o *UiNodeSelectAttributes) GetDisabled() bool`

GetDisabled returns the Disabled field if non-nil, zero value otherwise.

### GetDisabledOk

`func (o *UiNodeSelectAttributes) GetDisabledOk() (*bool, bool)`

GetDisabledOk returns a tuple with the Disabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisabled

`func (o *UiNodeSelectAttributes) SetDisabled(v bool)`

SetDisabled sets Disabled field to given value.


### GetLabel

`func (o *UiNodeSelectAttributes) GetLabel() UiText`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *UiNodeSelectAttributes) GetLabelOk() (*UiText, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *UiNodeSelectAttributes) SetLabel(v UiText)`

SetLabel sets Label field to given value.

### HasLabel

`func (o *UiNodeSelectAttributes) HasLabel() bool`

HasLabel returns a boolean if a field has been set.

### GetName

`func (o *UiNodeSelectAttributes) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UiNodeSelectAttributes) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UiNodeSelectAttributes) SetName(v string)`

SetName sets Name field to given value.


### GetNodeType

`func (o *UiNodeSelectAttributes) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *UiNodeSelectAttributes) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *UiNodeSelectAttributes) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.


### GetOptions

`func (o *UiNodeSelectAttributes) GetOptions() []UiNodeSelectOption`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *UiNodeSelectAttributes) GetOptionsOk() (*[]UiNodeSelectOption, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *UiNodeSelectAttributes) SetOptions(v []UiNodeSelectOption)`

SetOptions sets Options field to given value.


### GetRequired

`func (o *UiNodeSelectAttributes) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *UiNodeSelectAttributes) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *UiNodeSelectAttributes) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *UiNodeSelectAttributes) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### GetValue

`func (o *UiNodeSelectAttributes) GetValue() map[string]interface{}`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *UiNodeSelectAttributes) GetValueOk() (*map[string]interface{}, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *UiNodeSelectAttributes) SetValue(v map[string]interface{})`

SetValue sets Value field to given value.

### HasValue

`func (o *UiNodeSelectAttributes) HasValue() bool`

HasValue returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


