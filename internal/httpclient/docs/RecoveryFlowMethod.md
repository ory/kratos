# RecoveryFlowMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**RecoveryFlowMethodConfig**](recoveryFlowMethodConfig.md) |  | 
**Method** | **string** | Method contains the request credentials type. | 

## Methods

### NewRecoveryFlowMethod

`func NewRecoveryFlowMethod(config RecoveryFlowMethodConfig, method string, ) *RecoveryFlowMethod`

NewRecoveryFlowMethod instantiates a new RecoveryFlowMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRecoveryFlowMethodWithDefaults

`func NewRecoveryFlowMethodWithDefaults() *RecoveryFlowMethod`

NewRecoveryFlowMethodWithDefaults instantiates a new RecoveryFlowMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *RecoveryFlowMethod) GetConfig() RecoveryFlowMethodConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *RecoveryFlowMethod) GetConfigOk() (*RecoveryFlowMethodConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *RecoveryFlowMethod) SetConfig(v RecoveryFlowMethodConfig)`

SetConfig sets Config field to given value.


### GetMethod

`func (o *RecoveryFlowMethod) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *RecoveryFlowMethod) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *RecoveryFlowMethod) SetMethod(v string)`

SetMethod sets Method field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


