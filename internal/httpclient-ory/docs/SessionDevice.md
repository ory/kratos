# SessionDevice

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Device record ID | 
**IpAddress** | Pointer to **string** | IPAddress of the client | [optional] 
**Location** | Pointer to **string** | Geo Location corresponding to the IP Address | [optional] 
**UserAgent** | Pointer to **string** | UserAgent of the client | [optional] 

## Methods

### NewSessionDevice

`func NewSessionDevice(id string, ) *SessionDevice`

NewSessionDevice instantiates a new SessionDevice object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionDeviceWithDefaults

`func NewSessionDeviceWithDefaults() *SessionDevice`

NewSessionDeviceWithDefaults instantiates a new SessionDevice object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *SessionDevice) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SessionDevice) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SessionDevice) SetId(v string)`

SetId sets Id field to given value.


### GetIpAddress

`func (o *SessionDevice) GetIpAddress() string`

GetIpAddress returns the IpAddress field if non-nil, zero value otherwise.

### GetIpAddressOk

`func (o *SessionDevice) GetIpAddressOk() (*string, bool)`

GetIpAddressOk returns a tuple with the IpAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIpAddress

`func (o *SessionDevice) SetIpAddress(v string)`

SetIpAddress sets IpAddress field to given value.

### HasIpAddress

`func (o *SessionDevice) HasIpAddress() bool`

HasIpAddress returns a boolean if a field has been set.

### GetLocation

`func (o *SessionDevice) GetLocation() string`

GetLocation returns the Location field if non-nil, zero value otherwise.

### GetLocationOk

`func (o *SessionDevice) GetLocationOk() (*string, bool)`

GetLocationOk returns a tuple with the Location field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocation

`func (o *SessionDevice) SetLocation(v string)`

SetLocation sets Location field to given value.

### HasLocation

`func (o *SessionDevice) HasLocation() bool`

HasLocation returns a boolean if a field has been set.

### GetUserAgent

`func (o *SessionDevice) GetUserAgent() string`

GetUserAgent returns the UserAgent field if non-nil, zero value otherwise.

### GetUserAgentOk

`func (o *SessionDevice) GetUserAgentOk() (*string, bool)`

GetUserAgentOk returns a tuple with the UserAgent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserAgent

`func (o *SessionDevice) SetUserAgent(v string)`

SetUserAgent sets UserAgent field to given value.

### HasUserAgent

`func (o *SessionDevice) HasUserAgent() bool`

HasUserAgent returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


