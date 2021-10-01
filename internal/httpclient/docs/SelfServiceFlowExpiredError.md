# SelfServiceFlowExpiredError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **int64** | The status code | [optional] 
**Debug** | Pointer to **string** | Debug information  This field is often not exposed to protect against leaking sensitive information. | [optional] 
**Details** | Pointer to **map[string]map[string]interface{}** | Further error details | [optional] 
**Id** | Pointer to **string** | The error ID  Useful when trying to identify various errors in application logic. | [optional] 
**Message** | **string** | Error message  The error&#39;s message. | 
**Reason** | Pointer to **string** | A human-readable reason for the error | [optional] 
**Request** | Pointer to **string** | The request ID  The request ID is often exposed internally in order to trace errors across service architectures. This is often a UUID. | [optional] 
**Status** | Pointer to **string** | The status description | [optional] 
**UseFlowId** | Pointer to **string** |  | [optional] 

## Methods

### NewSelfServiceFlowExpiredError

`func NewSelfServiceFlowExpiredError(message string, ) *SelfServiceFlowExpiredError`

NewSelfServiceFlowExpiredError instantiates a new SelfServiceFlowExpiredError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceFlowExpiredErrorWithDefaults

`func NewSelfServiceFlowExpiredErrorWithDefaults() *SelfServiceFlowExpiredError`

NewSelfServiceFlowExpiredErrorWithDefaults instantiates a new SelfServiceFlowExpiredError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *SelfServiceFlowExpiredError) GetCode() int64`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *SelfServiceFlowExpiredError) GetCodeOk() (*int64, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *SelfServiceFlowExpiredError) SetCode(v int64)`

SetCode sets Code field to given value.

### HasCode

`func (o *SelfServiceFlowExpiredError) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetDebug

`func (o *SelfServiceFlowExpiredError) GetDebug() string`

GetDebug returns the Debug field if non-nil, zero value otherwise.

### GetDebugOk

`func (o *SelfServiceFlowExpiredError) GetDebugOk() (*string, bool)`

GetDebugOk returns a tuple with the Debug field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDebug

`func (o *SelfServiceFlowExpiredError) SetDebug(v string)`

SetDebug sets Debug field to given value.

### HasDebug

`func (o *SelfServiceFlowExpiredError) HasDebug() bool`

HasDebug returns a boolean if a field has been set.

### GetDetails

`func (o *SelfServiceFlowExpiredError) GetDetails() map[string]map[string]interface{}`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *SelfServiceFlowExpiredError) GetDetailsOk() (*map[string]map[string]interface{}, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *SelfServiceFlowExpiredError) SetDetails(v map[string]map[string]interface{})`

SetDetails sets Details field to given value.

### HasDetails

`func (o *SelfServiceFlowExpiredError) HasDetails() bool`

HasDetails returns a boolean if a field has been set.

### GetId

`func (o *SelfServiceFlowExpiredError) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SelfServiceFlowExpiredError) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SelfServiceFlowExpiredError) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *SelfServiceFlowExpiredError) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMessage

`func (o *SelfServiceFlowExpiredError) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *SelfServiceFlowExpiredError) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *SelfServiceFlowExpiredError) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetReason

`func (o *SelfServiceFlowExpiredError) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *SelfServiceFlowExpiredError) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *SelfServiceFlowExpiredError) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *SelfServiceFlowExpiredError) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetRequest

`func (o *SelfServiceFlowExpiredError) GetRequest() string`

GetRequest returns the Request field if non-nil, zero value otherwise.

### GetRequestOk

`func (o *SelfServiceFlowExpiredError) GetRequestOk() (*string, bool)`

GetRequestOk returns a tuple with the Request field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequest

`func (o *SelfServiceFlowExpiredError) SetRequest(v string)`

SetRequest sets Request field to given value.

### HasRequest

`func (o *SelfServiceFlowExpiredError) HasRequest() bool`

HasRequest returns a boolean if a field has been set.

### GetStatus

`func (o *SelfServiceFlowExpiredError) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *SelfServiceFlowExpiredError) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *SelfServiceFlowExpiredError) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *SelfServiceFlowExpiredError) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

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


