# CourierMessageDispatch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreatedAt** | Pointer to **time.Time** | CreatedAt is a helper struct field for gobuffalo.pop. | [optional] 
**Error** | Pointer to **string** | An optional error | [optional] 
**Id** | Pointer to **string** | The ID of this message dispatch | [optional] 
**MessageId** | Pointer to **string** | The ID of th message being dispatched | [optional] 
**Status** | Pointer to **string** | The status of this dispatch Either \&quot;failed\&quot; or \&quot;success\&quot; failed CourierMessageDispatchStatusFailed success CourierMessageDispatchStatusSuccess | [optional] 
**UpdatedAt** | Pointer to **time.Time** | UpdatedAt is a helper struct field for gobuffalo.pop. | [optional] 

## Methods

### NewCourierMessageDispatch

`func NewCourierMessageDispatch() *CourierMessageDispatch`

NewCourierMessageDispatch instantiates a new CourierMessageDispatch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCourierMessageDispatchWithDefaults

`func NewCourierMessageDispatchWithDefaults() *CourierMessageDispatch`

NewCourierMessageDispatchWithDefaults instantiates a new CourierMessageDispatch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreatedAt

`func (o *CourierMessageDispatch) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *CourierMessageDispatch) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *CourierMessageDispatch) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *CourierMessageDispatch) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetError

`func (o *CourierMessageDispatch) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *CourierMessageDispatch) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *CourierMessageDispatch) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *CourierMessageDispatch) HasError() bool`

HasError returns a boolean if a field has been set.

### GetId

`func (o *CourierMessageDispatch) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *CourierMessageDispatch) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *CourierMessageDispatch) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *CourierMessageDispatch) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMessageId

`func (o *CourierMessageDispatch) GetMessageId() string`

GetMessageId returns the MessageId field if non-nil, zero value otherwise.

### GetMessageIdOk

`func (o *CourierMessageDispatch) GetMessageIdOk() (*string, bool)`

GetMessageIdOk returns a tuple with the MessageId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessageId

`func (o *CourierMessageDispatch) SetMessageId(v string)`

SetMessageId sets MessageId field to given value.

### HasMessageId

`func (o *CourierMessageDispatch) HasMessageId() bool`

HasMessageId returns a boolean if a field has been set.

### GetStatus

`func (o *CourierMessageDispatch) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *CourierMessageDispatch) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *CourierMessageDispatch) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *CourierMessageDispatch) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *CourierMessageDispatch) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *CourierMessageDispatch) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *CourierMessageDispatch) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *CourierMessageDispatch) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


