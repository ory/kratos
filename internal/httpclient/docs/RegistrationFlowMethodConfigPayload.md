# RegistrationFlowMethodConfigPayload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Action** | **string** | Action should be used as the form action URL &#x60;&lt;form action&#x3D;\&quot;{{ .Action }}\&quot; method&#x3D;\&quot;post\&quot;&gt;&#x60;. | 
**Messages** | Pointer to [**[]UiText**](UiText.md) |  | [optional] 
**Method** | **string** | Method is the form method (e.g. POST) | 
**Nodes** | [**[]UiNode**](UiNode.md) |  | 
**Providers** | Pointer to [**[][]UiNode**]([]UiNode.md) | Providers is set for the \&quot;oidc\&quot; registration method. | [optional] 

## Methods

### NewRegistrationFlowMethodConfigPayload

`func NewRegistrationFlowMethodConfigPayload(action string, method string, nodes []UiNode, ) *RegistrationFlowMethodConfigPayload`

NewRegistrationFlowMethodConfigPayload instantiates a new RegistrationFlowMethodConfigPayload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistrationFlowMethodConfigPayloadWithDefaults

`func NewRegistrationFlowMethodConfigPayloadWithDefaults() *RegistrationFlowMethodConfigPayload`

NewRegistrationFlowMethodConfigPayloadWithDefaults instantiates a new RegistrationFlowMethodConfigPayload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAction

`func (o *RegistrationFlowMethodConfigPayload) GetAction() string`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *RegistrationFlowMethodConfigPayload) GetActionOk() (*string, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *RegistrationFlowMethodConfigPayload) SetAction(v string)`

SetAction sets Action field to given value.


### GetMessages

`func (o *RegistrationFlowMethodConfigPayload) GetMessages() []UiText`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *RegistrationFlowMethodConfigPayload) GetMessagesOk() (*[]UiText, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *RegistrationFlowMethodConfigPayload) SetMessages(v []UiText)`

SetMessages sets Messages field to given value.

### HasMessages

`func (o *RegistrationFlowMethodConfigPayload) HasMessages() bool`

HasMessages returns a boolean if a field has been set.

### GetMethod

`func (o *RegistrationFlowMethodConfigPayload) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *RegistrationFlowMethodConfigPayload) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *RegistrationFlowMethodConfigPayload) SetMethod(v string)`

SetMethod sets Method field to given value.


### GetNodes

`func (o *RegistrationFlowMethodConfigPayload) GetNodes() []UiNode`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *RegistrationFlowMethodConfigPayload) GetNodesOk() (*[]UiNode, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *RegistrationFlowMethodConfigPayload) SetNodes(v []UiNode)`

SetNodes sets Nodes field to given value.


### GetProviders

`func (o *RegistrationFlowMethodConfigPayload) GetProviders() [][]UiNode`

GetProviders returns the Providers field if non-nil, zero value otherwise.

### GetProvidersOk

`func (o *RegistrationFlowMethodConfigPayload) GetProvidersOk() (*[][]UiNode, bool)`

GetProvidersOk returns a tuple with the Providers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviders

`func (o *RegistrationFlowMethodConfigPayload) SetProviders(v [][]UiNode)`

SetProviders sets Providers field to given value.

### HasProviders

`func (o *RegistrationFlowMethodConfigPayload) HasProviders() bool`

HasProviders returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


