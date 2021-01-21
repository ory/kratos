# LoginFlowMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**LoginFlowMethodConfig**](loginFlowMethodConfig.md) |  | 
**Method** | **string** | and so on. | 

## Methods

### NewLoginFlowMethod

`func NewLoginFlowMethod(config LoginFlowMethodConfig, method string, ) *LoginFlowMethod`

NewLoginFlowMethod instantiates a new LoginFlowMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLoginFlowMethodWithDefaults

`func NewLoginFlowMethodWithDefaults() *LoginFlowMethod`

NewLoginFlowMethodWithDefaults instantiates a new LoginFlowMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *LoginFlowMethod) GetConfig() LoginFlowMethodConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *LoginFlowMethod) GetConfigOk() (*LoginFlowMethodConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *LoginFlowMethod) SetConfig(v LoginFlowMethodConfig)`

SetConfig sets Config field to given value.


### GetMethod

`func (o *LoginFlowMethod) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *LoginFlowMethod) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *LoginFlowMethod) SetMethod(v string)`

SetMethod sets Method field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


