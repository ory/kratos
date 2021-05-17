# JsonError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Error** | [**GenericError**](GenericError.md) |  | 

## Methods

### NewJsonError

`func NewJsonError(error_ GenericError, ) *JsonError`

NewJsonError instantiates a new JsonError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJsonErrorWithDefaults

`func NewJsonErrorWithDefaults() *JsonError`

NewJsonErrorWithDefaults instantiates a new JsonError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetError

`func (o *JsonError) GetError() GenericError`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *JsonError) GetErrorOk() (*GenericError, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *JsonError) SetError(v GenericError)`

SetError sets Error field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


