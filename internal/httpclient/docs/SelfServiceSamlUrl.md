# SelfServiceSamlUrl

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SamlAcsUrl** | **string** | SamlAcsURL is a post endpoint to handle SAML Response  format: uri | 
**SamlMetadataUrl** | **string** | SamlMetadataURL is a get endpoint to get the metadata  format: uri | 

## Methods

### NewSelfServiceSamlUrl

`func NewSelfServiceSamlUrl(samlAcsUrl string, samlMetadataUrl string, ) *SelfServiceSamlUrl`

NewSelfServiceSamlUrl instantiates a new SelfServiceSamlUrl object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceSamlUrlWithDefaults

`func NewSelfServiceSamlUrlWithDefaults() *SelfServiceSamlUrl`

NewSelfServiceSamlUrlWithDefaults instantiates a new SelfServiceSamlUrl object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSamlAcsUrl

`func (o *SelfServiceSamlUrl) GetSamlAcsUrl() string`

GetSamlAcsUrl returns the SamlAcsUrl field if non-nil, zero value otherwise.

### GetSamlAcsUrlOk

`func (o *SelfServiceSamlUrl) GetSamlAcsUrlOk() (*string, bool)`

GetSamlAcsUrlOk returns a tuple with the SamlAcsUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSamlAcsUrl

`func (o *SelfServiceSamlUrl) SetSamlAcsUrl(v string)`

SetSamlAcsUrl sets SamlAcsUrl field to given value.


### GetSamlMetadataUrl

`func (o *SelfServiceSamlUrl) GetSamlMetadataUrl() string`

GetSamlMetadataUrl returns the SamlMetadataUrl field if non-nil, zero value otherwise.

### GetSamlMetadataUrlOk

`func (o *SelfServiceSamlUrl) GetSamlMetadataUrlOk() (*string, bool)`

GetSamlMetadataUrlOk returns a tuple with the SamlMetadataUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSamlMetadataUrl

`func (o *SelfServiceSamlUrl) SetSamlMetadataUrl(v string)`

SetSamlMetadataUrl sets SamlMetadataUrl field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


