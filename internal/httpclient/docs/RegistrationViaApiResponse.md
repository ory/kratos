# RegistrationViaApiResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Identity** | [**Identity**](Identity.md) |  | 
**Session** | Pointer to [**Session**](session.md) |  | [optional] 
**SessionToken** | **string** | The Session Token  This field is only set when the session hook is configured as a post-registration hook.  A session token is equivalent to a session cookie, but it can be sent in the HTTP Authorization Header:  Authorization: bearer ${session-token}  The session token is only issued for API flows, not for Browser flows! | 

## Methods

### NewRegistrationViaApiResponse

`func NewRegistrationViaApiResponse(identity Identity, sessionToken string, ) *RegistrationViaApiResponse`

NewRegistrationViaApiResponse instantiates a new RegistrationViaApiResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistrationViaApiResponseWithDefaults

`func NewRegistrationViaApiResponseWithDefaults() *RegistrationViaApiResponse`

NewRegistrationViaApiResponseWithDefaults instantiates a new RegistrationViaApiResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIdentity

`func (o *RegistrationViaApiResponse) GetIdentity() Identity`

GetIdentity returns the Identity field if non-nil, zero value otherwise.

### GetIdentityOk

`func (o *RegistrationViaApiResponse) GetIdentityOk() (*Identity, bool)`

GetIdentityOk returns a tuple with the Identity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentity

`func (o *RegistrationViaApiResponse) SetIdentity(v Identity)`

SetIdentity sets Identity field to given value.


### GetSession

`func (o *RegistrationViaApiResponse) GetSession() Session`

GetSession returns the Session field if non-nil, zero value otherwise.

### GetSessionOk

`func (o *RegistrationViaApiResponse) GetSessionOk() (*Session, bool)`

GetSessionOk returns a tuple with the Session field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSession

`func (o *RegistrationViaApiResponse) SetSession(v Session)`

SetSession sets Session field to given value.

### HasSession

`func (o *RegistrationViaApiResponse) HasSession() bool`

HasSession returns a boolean if a field has been set.

### GetSessionToken

`func (o *RegistrationViaApiResponse) GetSessionToken() string`

GetSessionToken returns the SessionToken field if non-nil, zero value otherwise.

### GetSessionTokenOk

`func (o *RegistrationViaApiResponse) GetSessionTokenOk() (*string, bool)`

GetSessionTokenOk returns a tuple with the SessionToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSessionToken

`func (o *RegistrationViaApiResponse) SetSessionToken(v string)`

SetSessionToken sets SessionToken field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


