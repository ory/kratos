# AdminIdentityImportCredentials

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Oidc** | Pointer to [**AdminCreateIdentityImportCredentialsOidc**](AdminCreateIdentityImportCredentialsOidc.md) |  | [optional] 
**Password** | Pointer to [**AdminCreateIdentityImportCredentialsPassword**](AdminCreateIdentityImportCredentialsPassword.md) |  | [optional] 

## Methods

### NewAdminIdentityImportCredentials

`func NewAdminIdentityImportCredentials() *AdminIdentityImportCredentials`

NewAdminIdentityImportCredentials instantiates a new AdminIdentityImportCredentials object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAdminIdentityImportCredentialsWithDefaults

`func NewAdminIdentityImportCredentialsWithDefaults() *AdminIdentityImportCredentials`

NewAdminIdentityImportCredentialsWithDefaults instantiates a new AdminIdentityImportCredentials object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOidc

`func (o *AdminIdentityImportCredentials) GetOidc() AdminCreateIdentityImportCredentialsOidc`

GetOidc returns the Oidc field if non-nil, zero value otherwise.

### GetOidcOk

`func (o *AdminIdentityImportCredentials) GetOidcOk() (*AdminCreateIdentityImportCredentialsOidc, bool)`

GetOidcOk returns a tuple with the Oidc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOidc

`func (o *AdminIdentityImportCredentials) SetOidc(v AdminCreateIdentityImportCredentialsOidc)`

SetOidc sets Oidc field to given value.

### HasOidc

`func (o *AdminIdentityImportCredentials) HasOidc() bool`

HasOidc returns a boolean if a field has been set.

### GetPassword

`func (o *AdminIdentityImportCredentials) GetPassword() AdminCreateIdentityImportCredentialsPassword`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *AdminIdentityImportCredentials) GetPasswordOk() (*AdminCreateIdentityImportCredentialsPassword, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *AdminIdentityImportCredentials) SetPassword(v AdminCreateIdentityImportCredentialsPassword)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *AdminIdentityImportCredentials) HasPassword() bool`

HasPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


