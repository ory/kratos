# UiNodeAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Disabled** | **bool** | Sets the input&#39;s disabled field to true or false. | 
**Label** | Pointer to [**UiText**](UiText.md) |  | [optional] 
**Name** | **string** | The input&#39;s element name. | 
**Onclick** | Pointer to **string** | OnClick may contain javascript which should be executed on click. This is primarily used for WebAuthn. | [optional] 
**Onload** | Pointer to **string** | OnLoad may contain javascript which should be executed on load. This is primarily used for WebAuthn. Using this value makes most sense when used on the server-side. For JavaScript apps running in the browser please load the WebAuthn JavaScript:  &lt;script src&#x3D;\&quot;https://public-kratos.example.org/.well-known/ory/webauthn.js\&quot; type&#x3D;\&quot;script\&quot; async /&gt; | [optional] 
**Pattern** | Pointer to **string** | The input&#39;s pattern. | [optional] 
**Required** | Pointer to **bool** | Mark this input field as required. | [optional] 
**Type** | **string** |  | 
**Value** | Pointer to **interface{}** | The input&#39;s value. | [optional] 
**Id** | **string** | A unique identifier | 
**Text** | [**UiText**](UiText.md) |  | 
**Height** | Pointer to **int64** | Height of the image | [optional] 
**Src** | **string** | The image&#39;s source URL.  format: uri | 
**Width** | Pointer to **int64** | Width of the image | [optional] 
**Href** | **string** | The link&#39;s href (destination) URL.  format: uri | 
**Title** | [**UiText**](UiText.md) |  | 

## Methods

### NewUiNodeAttributes

`func NewUiNodeAttributes(disabled bool, name string, type_ string, id string, text UiText, src string, href string, title UiText, ) *UiNodeAttributes`

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


### GetOnclick

`func (o *UiNodeAttributes) GetOnclick() string`

GetOnclick returns the Onclick field if non-nil, zero value otherwise.

### GetOnclickOk

`func (o *UiNodeAttributes) GetOnclickOk() (*string, bool)`

GetOnclickOk returns a tuple with the Onclick field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnclick

`func (o *UiNodeAttributes) SetOnclick(v string)`

SetOnclick sets Onclick field to given value.

### HasOnclick

`func (o *UiNodeAttributes) HasOnclick() bool`

HasOnclick returns a boolean if a field has been set.

### GetOnload

`func (o *UiNodeAttributes) GetOnload() string`

GetOnload returns the Onload field if non-nil, zero value otherwise.

### GetOnloadOk

`func (o *UiNodeAttributes) GetOnloadOk() (*string, bool)`

GetOnloadOk returns a tuple with the Onload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnload

`func (o *UiNodeAttributes) SetOnload(v string)`

SetOnload sets Onload field to given value.

### HasOnload

`func (o *UiNodeAttributes) HasOnload() bool`

HasOnload returns a boolean if a field has been set.

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

`func (o *UiNodeAttributes) GetValue() interface{}`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *UiNodeAttributes) GetValueOk() (*interface{}, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *UiNodeAttributes) SetValue(v interface{})`

SetValue sets Value field to given value.

### HasValue

`func (o *UiNodeAttributes) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *UiNodeAttributes) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *UiNodeAttributes) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil
### GetId

`func (o *UiNodeAttributes) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *UiNodeAttributes) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *UiNodeAttributes) SetId(v string)`

SetId sets Id field to given value.


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


### GetHeight

`func (o *UiNodeAttributes) GetHeight() int64`

GetHeight returns the Height field if non-nil, zero value otherwise.

### GetHeightOk

`func (o *UiNodeAttributes) GetHeightOk() (*int64, bool)`

GetHeightOk returns a tuple with the Height field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHeight

`func (o *UiNodeAttributes) SetHeight(v int64)`

SetHeight sets Height field to given value.

### HasHeight

`func (o *UiNodeAttributes) HasHeight() bool`

HasHeight returns a boolean if a field has been set.

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


### GetWidth

`func (o *UiNodeAttributes) GetWidth() int64`

GetWidth returns the Width field if non-nil, zero value otherwise.

### GetWidthOk

`func (o *UiNodeAttributes) GetWidthOk() (*int64, bool)`

GetWidthOk returns a tuple with the Width field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWidth

`func (o *UiNodeAttributes) SetWidth(v int64)`

SetWidth sets Width field to given value.

### HasWidth

`func (o *UiNodeAttributes) HasWidth() bool`

HasWidth returns a boolean if a field has been set.

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


