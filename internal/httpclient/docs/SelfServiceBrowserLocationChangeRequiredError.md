# SelfServiceBrowserLocationChangeRequiredError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Error** | Pointer to [**GenericError**](GenericError.md) |  | [optional] 
**RedirectBrowserTo** | Pointer to **string** | Since when the flow has expired | [optional] 

## Methods

### NewSelfServiceBrowserLocationChangeRequiredError

`func NewSelfServiceBrowserLocationChangeRequiredError() *SelfServiceBrowserLocationChangeRequiredError`

NewSelfServiceBrowserLocationChangeRequiredError instantiates a new SelfServiceBrowserLocationChangeRequiredError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceBrowserLocationChangeRequiredErrorWithDefaults

`func NewSelfServiceBrowserLocationChangeRequiredErrorWithDefaults() *SelfServiceBrowserLocationChangeRequiredError`

NewSelfServiceBrowserLocationChangeRequiredErrorWithDefaults instantiates a new SelfServiceBrowserLocationChangeRequiredError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetError

`func (o *SelfServiceBrowserLocationChangeRequiredError) GetError() GenericError`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *SelfServiceBrowserLocationChangeRequiredError) GetErrorOk() (*GenericError, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *SelfServiceBrowserLocationChangeRequiredError) SetError(v GenericError)`

SetError sets Error field to given value.

### HasError

`func (o *SelfServiceBrowserLocationChangeRequiredError) HasError() bool`

HasError returns a boolean if a field has been set.

### GetRedirectBrowserTo

`func (o *SelfServiceBrowserLocationChangeRequiredError) GetRedirectBrowserTo() string`

GetRedirectBrowserTo returns the RedirectBrowserTo field if non-nil, zero value otherwise.

### GetRedirectBrowserToOk

`func (o *SelfServiceBrowserLocationChangeRequiredError) GetRedirectBrowserToOk() (*string, bool)`

GetRedirectBrowserToOk returns a tuple with the RedirectBrowserTo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRedirectBrowserTo

`func (o *SelfServiceBrowserLocationChangeRequiredError) SetRedirectBrowserTo(v string)`

SetRedirectBrowserTo sets RedirectBrowserTo field to given value.

### HasRedirectBrowserTo

`func (o *SelfServiceBrowserLocationChangeRequiredError) HasRedirectBrowserTo() bool`

HasRedirectBrowserTo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


