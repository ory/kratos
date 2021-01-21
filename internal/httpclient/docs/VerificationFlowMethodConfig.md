# VerificationFlowMethodConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Action** | **string** | Action should be used as the form action URL &#x60;&lt;form action&#x3D;\&quot;{{ .Action }}\&quot; method&#x3D;\&quot;post\&quot;&gt;&#x60;. | 
**Messages** | Pointer to [**[]UiText**](UiText.md) |  | [optional] 
**Method** | **string** | Method is the form method (e.g. POST) | 
**Nodes** | [**[]UiNode**](UiNode.md) |  | 

## Methods

### NewVerificationFlowMethodConfig

`func NewVerificationFlowMethodConfig(action string, method string, nodes []UiNode, ) *VerificationFlowMethodConfig`

NewVerificationFlowMethodConfig instantiates a new VerificationFlowMethodConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVerificationFlowMethodConfigWithDefaults

`func NewVerificationFlowMethodConfigWithDefaults() *VerificationFlowMethodConfig`

NewVerificationFlowMethodConfigWithDefaults instantiates a new VerificationFlowMethodConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAction

`func (o *VerificationFlowMethodConfig) GetAction() string`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *VerificationFlowMethodConfig) GetActionOk() (*string, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *VerificationFlowMethodConfig) SetAction(v string)`

SetAction sets Action field to given value.


### GetMessages

`func (o *VerificationFlowMethodConfig) GetMessages() []UiText`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *VerificationFlowMethodConfig) GetMessagesOk() (*[]UiText, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *VerificationFlowMethodConfig) SetMessages(v []UiText)`

SetMessages sets Messages field to given value.

### HasMessages

`func (o *VerificationFlowMethodConfig) HasMessages() bool`

HasMessages returns a boolean if a field has been set.

### GetMethod

`func (o *VerificationFlowMethodConfig) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *VerificationFlowMethodConfig) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *VerificationFlowMethodConfig) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetNodes

`func (o *VerificationFlowMethodConfig) GetNodes() []UiNode`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *VerificationFlowMethodConfig) GetNodesOk() (*[]UiNode, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *VerificationFlowMethodConfig) SetNodes(v []UiNode)`

SetNodes sets Nodes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


