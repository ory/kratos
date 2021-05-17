# SelfServiceErrorContainer

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreatedAt** | Pointer to **time.Time** | CreatedAt is a helper struct field for gobuffalo.pop. | [optional] 
**Errors** | **[]map[string]interface{}** | Errors in the container | 
**Id** | **string** |  | 
**UpdatedAt** | Pointer to **time.Time** | UpdatedAt is a helper struct field for gobuffalo.pop. | [optional] 

## Methods

### NewSelfServiceErrorContainer

`func NewSelfServiceErrorContainer(errors []map[string]interface{}, id string, ) *SelfServiceErrorContainer`

NewSelfServiceErrorContainer instantiates a new SelfServiceErrorContainer object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceErrorContainerWithDefaults

`func NewSelfServiceErrorContainerWithDefaults() *SelfServiceErrorContainer`

NewSelfServiceErrorContainerWithDefaults instantiates a new SelfServiceErrorContainer object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreatedAt

`func (o *SelfServiceErrorContainer) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *SelfServiceErrorContainer) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *SelfServiceErrorContainer) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *SelfServiceErrorContainer) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetErrors

`func (o *SelfServiceErrorContainer) GetErrors() []map[string]interface{}`

GetErrors returns the Errors field if non-nil, zero value otherwise.

### GetErrorsOk

`func (o *SelfServiceErrorContainer) GetErrorsOk() (*[]map[string]interface{}, bool)`

GetErrorsOk returns a tuple with the Errors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrors

`func (o *SelfServiceErrorContainer) SetErrors(v []map[string]interface{})`

SetErrors sets Errors field to given value.


### GetId

`func (o *SelfServiceErrorContainer) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SelfServiceErrorContainer) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SelfServiceErrorContainer) SetId(v string)`

SetId sets Id field to given value.


### GetUpdatedAt

`func (o *SelfServiceErrorContainer) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *SelfServiceErrorContainer) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *SelfServiceErrorContainer) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *SelfServiceErrorContainer) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


