# AdminCreateIdentityBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SchemaId** | **string** | SchemaID is the ID of the JSON Schema to be used for validating the identity&#39;s traits. | 
**Traits** | **map[string]interface{}** | Traits represent an identity&#39;s traits. The identity is able to create, modify, and delete traits in a self-service manner. The input will always be validated against the JSON Schema defined in &#x60;schema_url&#x60;. | 

## Methods

### NewAdminCreateIdentityBody

`func NewAdminCreateIdentityBody(schemaId string, traits map[string]interface{}, ) *AdminCreateIdentityBody`

NewAdminCreateIdentityBody instantiates a new AdminCreateIdentityBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAdminCreateIdentityBodyWithDefaults

`func NewAdminCreateIdentityBodyWithDefaults() *AdminCreateIdentityBody`

NewAdminCreateIdentityBodyWithDefaults instantiates a new AdminCreateIdentityBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSchemaId

`func (o *AdminCreateIdentityBody) GetSchemaId() string`

GetSchemaId returns the SchemaId field if non-nil, zero value otherwise.

### GetSchemaIdOk

`func (o *AdminCreateIdentityBody) GetSchemaIdOk() (*string, bool)`

GetSchemaIdOk returns a tuple with the SchemaId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchemaId

`func (o *AdminCreateIdentityBody) SetSchemaId(v string)`

SetSchemaId sets SchemaId field to given value.


### GetTraits

`func (o *AdminCreateIdentityBody) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *AdminCreateIdentityBody) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *AdminCreateIdentityBody) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


