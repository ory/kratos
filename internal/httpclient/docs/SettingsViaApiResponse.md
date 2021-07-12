# SettingsViaApiResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Flow** | [**SelfServiceSettingsFlow**](SelfServiceSettingsFlow.md) |  | 
**Identity** | [**Identity**](Identity.md) |  | 

## Methods

### NewSettingsViaApiResponse

`func NewSettingsViaApiResponse(flow SelfServiceSettingsFlow, identity Identity, ) *SettingsViaApiResponse`

NewSettingsViaApiResponse instantiates a new SettingsViaApiResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSettingsViaApiResponseWithDefaults

`func NewSettingsViaApiResponseWithDefaults() *SettingsViaApiResponse`

NewSettingsViaApiResponseWithDefaults instantiates a new SettingsViaApiResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFlow

`func (o *SettingsViaApiResponse) GetFlow() SelfServiceSettingsFlow`

GetFlow returns the Flow field if non-nil, zero value otherwise.

### GetFlowOk

`func (o *SettingsViaApiResponse) GetFlowOk() (*SelfServiceSettingsFlow, bool)`

GetFlowOk returns a tuple with the Flow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlow

`func (o *SettingsViaApiResponse) SetFlow(v SelfServiceSettingsFlow)`

SetFlow sets Flow field to given value.


### GetIdentity

`func (o *SettingsViaApiResponse) GetIdentity() Identity`

GetIdentity returns the Identity field if non-nil, zero value otherwise.

### GetIdentityOk

`func (o *SettingsViaApiResponse) GetIdentityOk() (*Identity, bool)`

GetIdentityOk returns a tuple with the Identity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentity

`func (o *SettingsViaApiResponse) SetIdentity(v Identity)`

SetIdentity sets Identity field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


