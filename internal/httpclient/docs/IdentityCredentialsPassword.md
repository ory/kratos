# IdentityCredentialsPassword

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**HashedPassword** | Pointer to **string** | HashedPassword is a hash-representation of the password. | [optional] 

## Methods

### NewIdentityCredentialsPassword

`func NewIdentityCredentialsPassword() *IdentityCredentialsPassword`

NewIdentityCredentialsPassword instantiates a new IdentityCredentialsPassword object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIdentityCredentialsPasswordWithDefaults

`func NewIdentityCredentialsPasswordWithDefaults() *IdentityCredentialsPassword`

NewIdentityCredentialsPasswordWithDefaults instantiates a new IdentityCredentialsPassword object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHashedPassword

`func (o *IdentityCredentialsPassword) GetHashedPassword() string`

GetHashedPassword returns the HashedPassword field if non-nil, zero value otherwise.

### GetHashedPasswordOk

`func (o *IdentityCredentialsPassword) GetHashedPasswordOk() (*string, bool)`

GetHashedPasswordOk returns a tuple with the HashedPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHashedPassword

`func (o *IdentityCredentialsPassword) SetHashedPassword(v string)`

SetHashedPassword sets HashedPassword field to given value.

### HasHashedPassword

`func (o *IdentityCredentialsPassword) HasHashedPassword() bool`

HasHashedPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


