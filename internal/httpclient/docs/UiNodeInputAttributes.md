# UiNodeInputAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Disabled** | **bool** | Sets the input&#39;s disabled field to true or false. | 
**Label** | Pointer to [**UiText**](UiText.md) |  | [optional] 
**Name** | **string** | The input&#39;s element name. | 
**NodeType** | **string** | NodeType represents this node&#39;s types. It is a mirror of &#x60;node.type&#x60; and is primarily used to allow compatibility with OpenAPI 3.0.  In this struct it technically always is \&quot;input\&quot;. | 
**Onclick** | Pointer to **string** | OnClick may contain javascript which should be executed on click. This is primarily used for WebAuthn. | [optional] 
**Pattern** | Pointer to **string** | The input&#39;s pattern. | [optional] 
**Required** | Pointer to **bool** | Mark this input field as required. | [optional] 
**Type** | **string** |  | 
**Value** | Pointer to **interface{}** | The input&#39;s value. | [optional] 

## Methods

### NewUiNodeInputAttributes

`func NewUiNodeInputAttributes(disabled bool, name string, nodeType string, type_ string, ) *UiNodeInputAttributes`

NewUiNodeInputAttributes instantiates a new UiNodeInputAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUiNodeInputAttributesWithDefaults

`func NewUiNodeInputAttributesWithDefaults() *UiNodeInputAttributes`

NewUiNodeInputAttributesWithDefaults instantiates a new UiNodeInputAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDisabled

`func (o *UiNodeInputAttributes) GetDisabled() bool`

GetDisabled returns the Disabled field if non-nil, zero value otherwise.

### GetDisabledOk

`func (o *UiNodeInputAttributes) GetDisabledOk() (*bool, bool)`

GetDisabledOk returns a tuple with the Disabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisabled

`func (o *UiNodeInputAttributes) SetDisabled(v bool)`

SetDisabled sets Disabled field to given value.


### GetLabel

`func (o *UiNodeInputAttributes) GetLabel() UiText`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *UiNodeInputAttributes) GetLabelOk() (*UiText, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *UiNodeInputAttributes) SetLabel(v UiText)`

SetLabel sets Label field to given value.

### HasLabel

`func (o *UiNodeInputAttributes) HasLabel() bool`

HasLabel returns a boolean if a field has been set.

### GetName

`func (o *UiNodeInputAttributes) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UiNodeInputAttributes) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UiNodeInputAttributes) SetName(v string)`

SetName sets Name field to given value.


### GetNodeType

`func (o *UiNodeInputAttributes) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *UiNodeInputAttributes) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *UiNodeInputAttributes) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.


### GetOnclick

`func (o *UiNodeInputAttributes) GetOnclick() string`

GetOnclick returns the Onclick field if non-nil, zero value otherwise.

### GetOnclickOk

`func (o *UiNodeInputAttributes) GetOnclickOk() (*string, bool)`

GetOnclickOk returns a tuple with the Onclick field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnclick

`func (o *UiNodeInputAttributes) SetOnclick(v string)`

SetOnclick sets Onclick field to given value.

### HasOnclick

`func (o *UiNodeInputAttributes) HasOnclick() bool`

HasOnclick returns a boolean if a field has been set.

### GetPattern

`func (o *UiNodeInputAttributes) GetPattern() string`

GetPattern returns the Pattern field if non-nil, zero value otherwise.

### GetPatternOk

`func (o *UiNodeInputAttributes) GetPatternOk() (*string, bool)`

GetPatternOk returns a tuple with the Pattern field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPattern

`func (o *UiNodeInputAttributes) SetPattern(v string)`

SetPattern sets Pattern field to given value.

### HasPattern

`func (o *UiNodeInputAttributes) HasPattern() bool`

HasPattern returns a boolean if a field has been set.

### GetRequired

`func (o *UiNodeInputAttributes) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *UiNodeInputAttributes) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *UiNodeInputAttributes) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *UiNodeInputAttributes) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### GetType

`func (o *UiNodeInputAttributes) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *UiNodeInputAttributes) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *UiNodeInputAttributes) SetType(v string)`

SetType sets Type field to given value.


### GetValue

`func (o *UiNodeInputAttributes) GetValue() interface{}`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *UiNodeInputAttributes) GetValueOk() (*interface{}, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *UiNodeInputAttributes) SetValue(v interface{})`

SetValue sets Value field to given value.

### HasValue

`func (o *UiNodeInputAttributes) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *UiNodeInputAttributes) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *UiNodeInputAttributes) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


