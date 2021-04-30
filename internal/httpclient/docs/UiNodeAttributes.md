# UiNodeAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Disabled** | **bool** | Sets the input&#39;s disabled field to true or false. | 
**Label** | Pointer to [**UiText**](UiText.md) |  | [optional] 
**Name** | **string** | The input&#39;s element name. | 
**Pattern** | Pointer to **string** | The input&#39;s pattern. | [optional] 
**Required** | Pointer to **bool** | Mark this input field as required. | [optional] 
**Type** | **string** |  | 
**Value** | Pointer to [**UiNodeInputAttributesValue**](UiNodeInputAttributesValue.md) |  | [optional] 
**Text** | [**UiText**](UiText.md) |  | 
**Src** | **string** | The image&#39;s source URL.  format: uri | 
**Href** | **string** | The link&#39;s href (destination) URL.  format: uri | 
**Title** | [**UiText**](UiText.md) |  | 

## Methods

### NewUiNodeAttributes

`func NewUiNodeAttributes(disabled bool, name string, type_ string, text UiText, src string, href string, title UiText, ) *UiNodeAttributes`

NewUiNodeAttributes instantiates a new UiNodeAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUiNodeAttributesWithDefaults

`func NewUiNodeAttributesWithDefaults() *UiNodeAttributes`

NewUiNodeAttributesWithDefaults instantiates a new UiNodeAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDisabled

`func (o *UiNodeAttributes) GetDisabled() bool`

GetDisabled returns the Disabled field if non-nil, zero value otherwise.

### GetDisabledOk

`func (o *UiNodeAttributes) GetDisabledOk() (*bool, bool)`

GetDisabledOk returns a tuple with the Disabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisabled

`func (o *UiNodeAttributes) SetDisabled(v bool)`

SetDisabled sets Disabled field to given value.


### GetLabel

`func (o *UiNodeAttributes) GetLabel() UiText`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *UiNodeAttributes) GetLabelOk() (*UiText, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *UiNodeAttributes) SetLabel(v UiText)`

SetLabel sets Label field to given value.

### HasLabel

`func (o *UiNodeAttributes) HasLabel() bool`

HasLabel returns a boolean if a field has been set.

### GetName

`func (o *UiNodeAttributes) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UiNodeAttributes) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UiNodeAttributes) SetName(v string)`

SetName sets Name field to given value.


### GetPattern

`func (o *UiNodeAttributes) GetPattern() string`

GetPattern returns the Pattern field if non-nil, zero value otherwise.

### GetPatternOk

`func (o *UiNodeAttributes) GetPatternOk() (*string, bool)`

GetPatternOk returns a tuple with the Pattern field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPattern

`func (o *UiNodeAttributes) SetPattern(v string)`

SetPattern sets Pattern field to given value.

### HasPattern

`func (o *UiNodeAttributes) HasPattern() bool`

HasPattern returns a boolean if a field has been set.

### GetRequired

`func (o *UiNodeAttributes) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *UiNodeAttributes) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *UiNodeAttributes) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *UiNodeAttributes) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### GetType

`func (o *UiNodeAttributes) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *UiNodeAttributes) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *UiNodeAttributes) SetType(v string)`

SetType sets Type field to given value.


### GetValue

`func (o *UiNodeAttributes) GetValue() UiNodeInputAttributesValue`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *UiNodeAttributes) GetValueOk() (*UiNodeInputAttributesValue, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *UiNodeAttributes) SetValue(v UiNodeInputAttributesValue)`

SetValue sets Value field to given value.

### HasValue

`func (o *UiNodeAttributes) HasValue() bool`

HasValue returns a boolean if a field has been set.

### GetText

`func (o *UiNodeAttributes) GetText() UiText`

GetText returns the Text field if non-nil, zero value otherwise.

### GetTextOk

`func (o *UiNodeAttributes) GetTextOk() (*UiText, bool)`

GetTextOk returns a tuple with the Text field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetText

`func (o *UiNodeAttributes) SetText(v UiText)`

SetText sets Text field to given value.


### GetSrc

`func (o *UiNodeAttributes) GetSrc() string`

GetSrc returns the Src field if non-nil, zero value otherwise.

### GetSrcOk

`func (o *UiNodeAttributes) GetSrcOk() (*string, bool)`

GetSrcOk returns a tuple with the Src field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSrc

`func (o *UiNodeAttributes) SetSrc(v string)`

SetSrc sets Src field to given value.


### GetHref

`func (o *UiNodeAttributes) GetHref() string`

GetHref returns the Href field if non-nil, zero value otherwise.

### GetHrefOk

`func (o *UiNodeAttributes) GetHrefOk() (*string, bool)`

GetHrefOk returns a tuple with the Href field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHref

`func (o *UiNodeAttributes) SetHref(v string)`

SetHref sets Href field to given value.


### GetTitle

`func (o *UiNodeAttributes) GetTitle() UiText`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *UiNodeAttributes) GetTitleOk() (*UiText, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *UiNodeAttributes) SetTitle(v UiText)`

SetTitle sets Title field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


