# SelfServiceFlowExpiredError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Error** | Pointer to [**GenericError**](GenericError.md) |  | [optional] 
**Since** | Pointer to **int64** | A Duration represents the elapsed time between two instants as an int64 nanosecond count. The representation limits the largest representable duration to approximately 290 years. | [optional] 
**UseFlowId** | Pointer to **string** |  | [optional] 

## Methods

### NewSelfServiceFlowExpiredError

`func NewSelfServiceFlowExpiredError() *SelfServiceFlowExpiredError`

NewSelfServiceFlowExpiredError instantiates a new SelfServiceFlowExpiredError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceFlowExpiredErrorWithDefaults

`func NewSelfServiceFlowExpiredErrorWithDefaults() *SelfServiceFlowExpiredError`

NewSelfServiceFlowExpiredErrorWithDefaults instantiates a new SelfServiceFlowExpiredError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetError

`func (o *SelfServiceFlowExpiredError) GetError() GenericError`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *SelfServiceFlowExpiredError) GetErrorOk() (*GenericError, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *SelfServiceFlowExpiredError) SetError(v GenericError)`

SetError sets Error field to given value.

### HasError

`func (o *SelfServiceFlowExpiredError) HasError() bool`

HasError returns a boolean if a field has been set.

### GetSince

`func (o *SelfServiceFlowExpiredError) GetSince() int64`

GetSince returns the Since field if non-nil, zero value otherwise.

### GetSinceOk

`func (o *SelfServiceFlowExpiredError) GetSinceOk() (*int64, bool)`

GetSinceOk returns a tuple with the Since field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSince

`func (o *SelfServiceFlowExpiredError) SetSince(v int64)`

SetSince sets Since field to given value.

### HasSince

`func (o *SelfServiceFlowExpiredError) HasSince() bool`

HasSince returns a boolean if a field has been set.

### GetUseFlowId

`func (o *SelfServiceFlowExpiredError) GetUseFlowId() string`

GetUseFlowId returns the UseFlowId field if non-nil, zero value otherwise.

### GetUseFlowIdOk

`func (o *SelfServiceFlowExpiredError) GetUseFlowIdOk() (*string, bool)`

GetUseFlowIdOk returns a tuple with the UseFlowId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseFlowId

`func (o *SelfServiceFlowExpiredError) SetUseFlowId(v string)`

SetUseFlowId sets UseFlowId field to given value.

### HasUseFlowId

`func (o *SelfServiceFlowExpiredError) HasUseFlowId() bool`

HasUseFlowId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


