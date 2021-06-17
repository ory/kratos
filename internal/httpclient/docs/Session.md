# Session

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | Pointer to **bool** | Whether or not the session is active. | [optional] 
**AuthenticatedAt** | Pointer to **time.Time** | The Session Authentication Timestamp  When this session was authenticated at. | [optional] 
**ExpiresAt** | Pointer to **time.Time** | The Session Expiry  When this session expires at. | [optional] 
**Id** | **string** |  | 
**Identity** | [**Identity**](Identity.md) |  | 
**IssuedAt** | Pointer to **time.Time** | The Session Issuance Timestamp  When this session was authenticated at. | [optional] 

## Methods

### NewSession

`func NewSession(id string, identity Identity, ) *Session`

NewSession instantiates a new Session object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionWithDefaults

`func NewSessionWithDefaults() *Session`

NewSessionWithDefaults instantiates a new Session object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActive

`func (o *Session) GetActive() bool`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *Session) GetActiveOk() (*bool, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *Session) SetActive(v bool)`

SetActive sets Active field to given value.

### HasActive

`func (o *Session) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetAuthenticatedAt

`func (o *Session) GetAuthenticatedAt() time.Time`

GetAuthenticatedAt returns the AuthenticatedAt field if non-nil, zero value otherwise.

### GetAuthenticatedAtOk

`func (o *Session) GetAuthenticatedAtOk() (*time.Time, bool)`

GetAuthenticatedAtOk returns a tuple with the AuthenticatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthenticatedAt

`func (o *Session) SetAuthenticatedAt(v time.Time)`

SetAuthenticatedAt sets AuthenticatedAt field to given value.

### HasAuthenticatedAt

`func (o *Session) HasAuthenticatedAt() bool`

HasAuthenticatedAt returns a boolean if a field has been set.

### GetExpiresAt

`func (o *Session) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *Session) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *Session) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *Session) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.

### GetId

`func (o *Session) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Session) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Session) SetId(v string)`

SetId sets Id field to given value.


### GetIdentity

`func (o *Session) GetIdentity() Identity`

GetIdentity returns the Identity field if non-nil, zero value otherwise.

### GetIdentityOk

`func (o *Session) GetIdentityOk() (*Identity, bool)`

GetIdentityOk returns a tuple with the Identity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentity

`func (o *Session) SetIdentity(v Identity)`

SetIdentity sets Identity field to given value.


### GetIssuedAt

`func (o *Session) GetIssuedAt() time.Time`

GetIssuedAt returns the IssuedAt field if non-nil, zero value otherwise.

### GetIssuedAtOk

`func (o *Session) GetIssuedAtOk() (*time.Time, bool)`

GetIssuedAtOk returns a tuple with the IssuedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuedAt

`func (o *Session) SetIssuedAt(v time.Time)`

SetIssuedAt sets IssuedAt field to given value.

### HasIssuedAt

`func (o *Session) HasIssuedAt() bool`

HasIssuedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


