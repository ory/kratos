# SessionAuthenticationMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CompletedAt** | Pointer to **time.Time** | When the authentication challenge was completed. | [optional] 
**Method** | Pointer to **string** |  | [optional] 

## Methods

### NewSessionAuthenticationMethod

`func NewSessionAuthenticationMethod() *SessionAuthenticationMethod`

NewSessionAuthenticationMethod instantiates a new SessionAuthenticationMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSessionAuthenticationMethodWithDefaults

`func NewSessionAuthenticationMethodWithDefaults() *SessionAuthenticationMethod`

NewSessionAuthenticationMethodWithDefaults instantiates a new SessionAuthenticationMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCompletedAt

`func (o *SessionAuthenticationMethod) GetCompletedAt() time.Time`

GetCompletedAt returns the CompletedAt field if non-nil, zero value otherwise.

### GetCompletedAtOk

`func (o *SessionAuthenticationMethod) GetCompletedAtOk() (*time.Time, bool)`

GetCompletedAtOk returns a tuple with the CompletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedAt

`func (o *SessionAuthenticationMethod) SetCompletedAt(v time.Time)`

SetCompletedAt sets CompletedAt field to given value.

### HasCompletedAt

`func (o *SessionAuthenticationMethod) HasCompletedAt() bool`

HasCompletedAt returns a boolean if a field has been set.

### GetMethod

`func (o *SessionAuthenticationMethod) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SessionAuthenticationMethod) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SessionAuthenticationMethod) SetMethod(v string)`

SetMethod sets Method field to given value.

### HasMethod

`func (o *SessionAuthenticationMethod) HasMethod() bool`

HasMethod returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


