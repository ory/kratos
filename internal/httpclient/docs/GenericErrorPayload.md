# GenericErrorPayload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **int64** | Code represents the error status code (404, 403, 401, ...). | [optional] 
**Debug** | Pointer to **string** | Debug contains debug information. This is usually not available and has to be enabled. | [optional] 
**Details** | Pointer to **map[string]map[string]interface{}** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**Reason** | Pointer to **string** |  | [optional] 
**Request** | Pointer to **string** |  | [optional] 
**Status** | Pointer to **string** |  | [optional] 

## Methods

### NewGenericErrorPayload

`func NewGenericErrorPayload() *GenericErrorPayload`

NewGenericErrorPayload instantiates a new GenericErrorPayload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGenericErrorPayloadWithDefaults

`func NewGenericErrorPayloadWithDefaults() *GenericErrorPayload`

NewGenericErrorPayloadWithDefaults instantiates a new GenericErrorPayload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *GenericErrorPayload) GetCode() int64`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *GenericErrorPayload) GetCodeOk() (*int64, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *GenericErrorPayload) SetCode(v int64)`

SetCode sets Code field to given value.

### HasCode

`func (o *GenericErrorPayload) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetDebug

`func (o *GenericErrorPayload) GetDebug() string`

GetDebug returns the Debug field if non-nil, zero value otherwise.

### GetDebugOk

`func (o *GenericErrorPayload) GetDebugOk() (*string, bool)`

GetDebugOk returns a tuple with the Debug field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDebug

`func (o *GenericErrorPayload) SetDebug(v string)`

SetDebug sets Debug field to given value.

### HasDebug

`func (o *GenericErrorPayload) HasDebug() bool`

HasDebug returns a boolean if a field has been set.

### GetDetails

`func (o *GenericErrorPayload) GetDetails() map[string]map[string]interface{}`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *GenericErrorPayload) GetDetailsOk() (*map[string]map[string]interface{}, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *GenericErrorPayload) SetDetails(v map[string]map[string]interface{})`

SetDetails sets Details field to given value.

### HasDetails

`func (o *GenericErrorPayload) HasDetails() bool`

HasDetails returns a boolean if a field has been set.

### GetMessage

`func (o *GenericErrorPayload) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *GenericErrorPayload) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *GenericErrorPayload) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *GenericErrorPayload) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetReason

`func (o *GenericErrorPayload) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *GenericErrorPayload) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *GenericErrorPayload) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *GenericErrorPayload) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetRequest

`func (o *GenericErrorPayload) GetRequest() string`

GetRequest returns the Request field if non-nil, zero value otherwise.

### GetRequestOk

`func (o *GenericErrorPayload) GetRequestOk() (*string, bool)`

GetRequestOk returns a tuple with the Request field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequest

`func (o *GenericErrorPayload) SetRequest(v string)`

SetRequest sets Request field to given value.

### HasRequest

`func (o *GenericErrorPayload) HasRequest() bool`

HasRequest returns a boolean if a field has been set.

### GetStatus

`func (o *GenericErrorPayload) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *GenericErrorPayload) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *GenericErrorPayload) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *GenericErrorPayload) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


