# SuccessfulSelfServiceLoginWithoutBrowser

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Session** | [**Session**](Session.md) |  | 
**SessionToken** | Pointer to **string** | The Session Token  A session token is equivalent to a session cookie, but it can be sent in the HTTP Authorization Header:  Authorization: bearer ${session-token}  The session token is only issued for API flows, not for Browser flows! | [optional] 

## Methods

### NewSuccessfulSelfServiceLoginWithoutBrowser

`func NewSuccessfulSelfServiceLoginWithoutBrowser(session Session, ) *SuccessfulSelfServiceLoginWithoutBrowser`

NewSuccessfulSelfServiceLoginWithoutBrowser instantiates a new SuccessfulSelfServiceLoginWithoutBrowser object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSuccessfulSelfServiceLoginWithoutBrowserWithDefaults

`func NewSuccessfulSelfServiceLoginWithoutBrowserWithDefaults() *SuccessfulSelfServiceLoginWithoutBrowser`

NewSuccessfulSelfServiceLoginWithoutBrowserWithDefaults instantiates a new SuccessfulSelfServiceLoginWithoutBrowser object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSession

`func (o *SuccessfulSelfServiceLoginWithoutBrowser) GetSession() Session`

GetSession returns the Session field if non-nil, zero value otherwise.

### GetSessionOk

`func (o *SuccessfulSelfServiceLoginWithoutBrowser) GetSessionOk() (*Session, bool)`

GetSessionOk returns a tuple with the Session field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSession

`func (o *SuccessfulSelfServiceLoginWithoutBrowser) SetSession(v Session)`

SetSession sets Session field to given value.


### GetSessionToken

`func (o *SuccessfulSelfServiceLoginWithoutBrowser) GetSessionToken() string`

GetSessionToken returns the SessionToken field if non-nil, zero value otherwise.

### GetSessionTokenOk

`func (o *SuccessfulSelfServiceLoginWithoutBrowser) GetSessionTokenOk() (*string, bool)`

GetSessionTokenOk returns a tuple with the SessionToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSessionToken

`func (o *SuccessfulSelfServiceLoginWithoutBrowser) SetSessionToken(v string)`

SetSessionToken sets SessionToken field to given value.

### HasSessionToken

`func (o *SuccessfulSelfServiceLoginWithoutBrowser) HasSessionToken() bool`

HasSessionToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


