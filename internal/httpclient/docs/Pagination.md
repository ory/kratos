# Pagination

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Page** | Pointer to **int64** | Pagination Page | [optional] [default to 0]
**PerPage** | Pointer to **int64** | Items per Page  This is the number of items per page. | [optional] [default to 250]

## Methods

### NewPagination

`func NewPagination() *Pagination`

NewPagination instantiates a new Pagination object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPaginationWithDefaults

`func NewPaginationWithDefaults() *Pagination`

NewPaginationWithDefaults instantiates a new Pagination object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPage

`func (o *Pagination) GetPage() int64`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *Pagination) GetPageOk() (*int64, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *Pagination) SetPage(v int64)`

SetPage sets Page field to given value.

### HasPage

`func (o *Pagination) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetPerPage

`func (o *Pagination) GetPerPage() int64`

GetPerPage returns the PerPage field if non-nil, zero value otherwise.

### GetPerPageOk

`func (o *Pagination) GetPerPageOk() (*int64, bool)`

GetPerPageOk returns a tuple with the PerPage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPerPage

`func (o *Pagination) SetPerPage(v int64)`

SetPerPage sets PerPage field to given value.

### HasPerPage

`func (o *Pagination) HasPerPage() bool`

HasPerPage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


