# SelfServiceRegistrationViaApiResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Identity** | [**Identity**](Identity.md) |  | 
**Session** | Pointer to [**Session**](Session.md) |  | [optional] 
**SessionToken** | Pointer to **string** | The Session Token  This field is only set when the session hook is configured as a post-registration hook.  A session token is equivalent to a session cookie, but it can be sent in the HTTP Authorization Header:  Authorization: bearer ${session-token}  The session token is only issued for API flows, not for Browser flows! | [optional] 

## Methods

### NewSelfServiceRegistrationViaApiResponse

`func NewSelfServiceRegistrationViaApiResponse(identity Identity, ) *SelfServiceRegistrationViaApiResponse`

NewSelfServiceRegistrationViaApiResponse instantiates a new SelfServiceRegistrationViaApiResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceRegistrationViaApiResponseWithDefaults

`func NewSelfServiceRegistrationViaApiResponseWithDefaults() *SelfServiceRegistrationViaApiResponse`

NewSelfServiceRegistrationViaApiResponseWithDefaults instantiates a new SelfServiceRegistrationViaApiResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIdentity

`func (o *SelfServiceRegistrationViaApiResponse) GetIdentity() Identity`

GetIdentity returns the Identity field if non-nil, zero value otherwise.

### GetIdentityOk

`func (o *SelfServiceRegistrationViaApiResponse) GetIdentityOk() (*Identity, bool)`

GetIdentityOk returns a tuple with the Identity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentity

`func (o *SelfServiceRegistrationViaApiResponse) SetIdentity(v Identity)`

SetIdentity sets Identity field to given value.


### GetSession

`func (o *SelfServiceRegistrationViaApiResponse) GetSession() Session`

GetSession returns the Session field if non-nil, zero value otherwise.

### GetSessionOk

`func (o *SelfServiceRegistrationViaApiResponse) GetSessionOk() (*Session, bool)`

GetSessionOk returns a tuple with the Session field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSession

`func (o *SelfServiceRegistrationViaApiResponse) SetSession(v Session)`

SetSession sets Session field to given value.

### HasSession

`func (o *SelfServiceRegistrationViaApiResponse) HasSession() bool`

HasSession returns a boolean if a field has been set.

### GetSessionToken

`func (o *SelfServiceRegistrationViaApiResponse) GetSessionToken() string`

GetSessionToken returns the SessionToken field if non-nil, zero value otherwise.

### GetSessionTokenOk

`func (o *SelfServiceRegistrationViaApiResponse) GetSessionTokenOk() (*string, bool)`

GetSessionTokenOk returns a tuple with the SessionToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSessionToken

`func (o *SelfServiceRegistrationViaApiResponse) SetSessionToken(v string)`

SetSessionToken sets SessionToken field to given value.

### HasSessionToken

`func (o *SelfServiceRegistrationViaApiResponse) HasSessionToken() bool`

HasSessionToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


