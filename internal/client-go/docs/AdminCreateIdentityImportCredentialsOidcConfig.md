# AdminCreateIdentityImportCredentialsOidcConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | Pointer to [**AdminCreateIdentityImportCredentialsPasswordConfig**](AdminCreateIdentityImportCredentialsPasswordConfig.md) |  | [optional] 
**Providers** | Pointer to [**[]AdminCreateIdentityImportCredentialsOidcProvider**](AdminCreateIdentityImportCredentialsOidcProvider.md) | A list of OpenID Connect Providers | [optional] 

## Methods

### NewAdminCreateIdentityImportCredentialsOidcConfig

`func NewAdminCreateIdentityImportCredentialsOidcConfig() *AdminCreateIdentityImportCredentialsOidcConfig`

NewAdminCreateIdentityImportCredentialsOidcConfig instantiates a new AdminCreateIdentityImportCredentialsOidcConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAdminCreateIdentityImportCredentialsOidcConfigWithDefaults

`func NewAdminCreateIdentityImportCredentialsOidcConfigWithDefaults() *AdminCreateIdentityImportCredentialsOidcConfig`

NewAdminCreateIdentityImportCredentialsOidcConfigWithDefaults instantiates a new AdminCreateIdentityImportCredentialsOidcConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *AdminCreateIdentityImportCredentialsOidcConfig) GetConfig() AdminCreateIdentityImportCredentialsPasswordConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *AdminCreateIdentityImportCredentialsOidcConfig) GetConfigOk() (*AdminCreateIdentityImportCredentialsPasswordConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *AdminCreateIdentityImportCredentialsOidcConfig) SetConfig(v AdminCreateIdentityImportCredentialsPasswordConfig)`

SetConfig sets Config field to given value.

### HasConfig

`func (o *AdminCreateIdentityImportCredentialsOidcConfig) HasConfig() bool`

HasConfig returns a boolean if a field has been set.

### GetProviders

`func (o *AdminCreateIdentityImportCredentialsOidcConfig) GetProviders() []AdminCreateIdentityImportCredentialsOidcProvider`

GetProviders returns the Providers field if non-nil, zero value otherwise.

### GetProvidersOk

`func (o *AdminCreateIdentityImportCredentialsOidcConfig) GetProvidersOk() (*[]AdminCreateIdentityImportCredentialsOidcProvider, bool)`

GetProvidersOk returns a tuple with the Providers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProviders

`func (o *AdminCreateIdentityImportCredentialsOidcConfig) SetProviders(v []AdminCreateIdentityImportCredentialsOidcProvider)`

SetProviders sets Providers field to given value.

### HasProviders

`func (o *AdminCreateIdentityImportCredentialsOidcConfig) HasProviders() bool`

HasProviders returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


