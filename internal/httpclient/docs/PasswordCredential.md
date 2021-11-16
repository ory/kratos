# PasswordCredential

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsHashed** | **bool** | This field must be true if the value represents a hashed password. Otherwise the value is assumed to be a cleartext password and it will be hashed before storing. | 
**Value** | **string** | Value represents the value of the password in clear text or hashed | 

## Methods

### NewPasswordCredential

`func NewPasswordCredential(isHashed bool, value string, ) *PasswordCredential`

NewPasswordCredential instantiates a new PasswordCredential object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPasswordCredentialWithDefaults

`func NewPasswordCredentialWithDefaults() *PasswordCredential`

NewPasswordCredentialWithDefaults instantiates a new PasswordCredential object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIsHashed

`func (o *PasswordCredential) GetIsHashed() bool`

GetIsHashed returns the IsHashed field if non-nil, zero value otherwise.

### GetIsHashedOk

`func (o *PasswordCredential) GetIsHashedOk() (*bool, bool)`

GetIsHashedOk returns a tuple with the IsHashed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsHashed

`func (o *PasswordCredential) SetIsHashed(v bool)`

SetIsHashed sets IsHashed field to given value.


### GetValue

`func (o *PasswordCredential) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *PasswordCredential) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *PasswordCredential) SetValue(v string)`

SetValue sets Value field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


