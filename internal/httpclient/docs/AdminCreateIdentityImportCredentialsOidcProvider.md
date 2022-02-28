# AdminCreateIdentityImportCredentialsOidcProvider

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Provider** | **string** | The OpenID Connect provider to link the subject to. Usually something like &#x60;google&#x60; or &#x60;github&#x60;. | 
**Subject** | **string** | The subject (&#x60;sub&#x60;) of the OpenID Connect connection. Usually the &#x60;sub&#x60; field of the ID Token. | 

## Methods

### NewAdminCreateIdentityImportCredentialsOidcProvider

`func NewAdminCreateIdentityImportCredentialsOidcProvider(provider string, subject string, ) *AdminCreateIdentityImportCredentialsOidcProvider`

NewAdminCreateIdentityImportCredentialsOidcProvider instantiates a new AdminCreateIdentityImportCredentialsOidcProvider object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAdminCreateIdentityImportCredentialsOidcProviderWithDefaults

`func NewAdminCreateIdentityImportCredentialsOidcProviderWithDefaults() *AdminCreateIdentityImportCredentialsOidcProvider`

NewAdminCreateIdentityImportCredentialsOidcProviderWithDefaults instantiates a new AdminCreateIdentityImportCredentialsOidcProvider object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProvider

`func (o *AdminCreateIdentityImportCredentialsOidcProvider) GetProvider() string`

GetProvider returns the Provider field if non-nil, zero value otherwise.

### GetProviderOk

`func (o *AdminCreateIdentityImportCredentialsOidcProvider) GetProviderOk() (*string, bool)`

GetProviderOk returns a tuple with the Provider field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProvider

`func (o *AdminCreateIdentityImportCredentialsOidcProvider) SetProvider(v string)`

SetProvider sets Provider field to given value.


### GetSubject

`func (o *AdminCreateIdentityImportCredentialsOidcProvider) GetSubject() string`

GetSubject returns the Subject field if non-nil, zero value otherwise.

### GetSubjectOk

`func (o *AdminCreateIdentityImportCredentialsOidcProvider) GetSubjectOk() (*string, bool)`

GetSubjectOk returns a tuple with the Subject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubject

`func (o *AdminCreateIdentityImportCredentialsOidcProvider) SetSubject(v string)`

SetSubject sets Subject field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


