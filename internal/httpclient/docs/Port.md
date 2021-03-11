# Port

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IP** | Pointer to **string** | IP | [optional] 
**PrivatePort** | **int32** | Port on the container | 
**PublicPort** | Pointer to **int32** | Port exposed on the host | [optional] 
**Type** | **string** | type | 

## Methods

### NewPort

`func NewPort(privatePort int32, type_ string, ) *Port`

NewPort instantiates a new Port object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPortWithDefaults

`func NewPortWithDefaults() *Port`

NewPortWithDefaults instantiates a new Port object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIP

`func (o *Port) GetIP() string`

GetIP returns the IP field if non-nil, zero value otherwise.

### GetIPOk

`func (o *Port) GetIPOk() (*string, bool)`

GetIPOk returns a tuple with the IP field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIP

`func (o *Port) SetIP(v string)`

SetIP sets IP field to given value.

### HasIP

`func (o *Port) HasIP() bool`

HasIP returns a boolean if a field has been set.

### GetPrivatePort

`func (o *Port) GetPrivatePort() int32`

GetPrivatePort returns the PrivatePort field if non-nil, zero value otherwise.

### GetPrivatePortOk

`func (o *Port) GetPrivatePortOk() (*int32, bool)`

GetPrivatePortOk returns a tuple with the PrivatePort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivatePort

`func (o *Port) SetPrivatePort(v int32)`

SetPrivatePort sets PrivatePort field to given value.


### GetPublicPort

`func (o *Port) GetPublicPort() int32`

GetPublicPort returns the PublicPort field if non-nil, zero value otherwise.

### GetPublicPortOk

`func (o *Port) GetPublicPortOk() (*int32, bool)`

GetPublicPortOk returns a tuple with the PublicPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicPort

`func (o *Port) SetPublicPort(v int32)`

SetPublicPort sets PublicPort field to given value.

### HasPublicPort

`func (o *Port) HasPublicPort() bool`

HasPublicPort returns a boolean if a field has been set.

### GetType

`func (o *Port) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Port) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Port) SetType(v string)`

SetType sets Type field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


