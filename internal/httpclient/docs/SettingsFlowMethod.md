# SettingsFlowMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**SettingsFlowMethodConfig**](settingsFlowMethodConfig.md) |  | 
**Method** | **string** | Method is the name of this flow method. | 

## Methods

### NewSettingsFlowMethod

`func NewSettingsFlowMethod(config SettingsFlowMethodConfig, method string, ) *SettingsFlowMethod`

NewSettingsFlowMethod instantiates a new SettingsFlowMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSettingsFlowMethodWithDefaults

`func NewSettingsFlowMethodWithDefaults() *SettingsFlowMethod`

NewSettingsFlowMethodWithDefaults instantiates a new SettingsFlowMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *SettingsFlowMethod) GetConfig() SettingsFlowMethodConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *SettingsFlowMethod) GetConfigOk() (*SettingsFlowMethodConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *SettingsFlowMethod) SetConfig(v SettingsFlowMethodConfig)`

SetConfig sets Config field to given value.


### GetMethod

`func (o *SettingsFlowMethod) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SettingsFlowMethod) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SettingsFlowMethod) SetMethod(v string)`

SetMethod sets Method field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


