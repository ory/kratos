# VerificationFlowMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**VerificationFlowMethodConfig**](verificationFlowMethodConfig.md) |  | 
**Method** | **string** | Method contains the request credentials type. | 

## Methods

### NewVerificationFlowMethod

`func NewVerificationFlowMethod(config VerificationFlowMethodConfig, method string, ) *VerificationFlowMethod`

NewVerificationFlowMethod instantiates a new VerificationFlowMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVerificationFlowMethodWithDefaults

`func NewVerificationFlowMethodWithDefaults() *VerificationFlowMethod`

NewVerificationFlowMethodWithDefaults instantiates a new VerificationFlowMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *VerificationFlowMethod) GetConfig() VerificationFlowMethodConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *VerificationFlowMethod) GetConfigOk() (*VerificationFlowMethodConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *VerificationFlowMethod) SetConfig(v VerificationFlowMethodConfig)`

SetConfig sets Config field to given value.


### GetMethod

`func (o *VerificationFlowMethod) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *VerificationFlowMethod) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *VerificationFlowMethod) SetMethod(v string)`

SetMethod sets Method field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


