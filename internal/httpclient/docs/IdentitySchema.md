# IdentitySchema

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | The ID of the Identity JSON Schema | [optional] 
**Schema** | Pointer to **map[string]interface{}** | The actual Identity JSON Schema | [optional] 

## Methods

### NewIdentitySchema

`func NewIdentitySchema() *IdentitySchema`

NewIdentitySchema instantiates a new IdentitySchema object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIdentitySchemaWithDefaults

`func NewIdentitySchemaWithDefaults() *IdentitySchema`

NewIdentitySchemaWithDefaults instantiates a new IdentitySchema object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *IdentitySchema) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *IdentitySchema) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *IdentitySchema) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *IdentitySchema) HasId() bool`

HasId returns a boolean if a field has been set.

### GetSchema

`func (o *IdentitySchema) GetSchema() map[string]interface{}`

GetSchema returns the Schema field if non-nil, zero value otherwise.

### GetSchemaOk

`func (o *IdentitySchema) GetSchemaOk() (*map[string]interface{}, bool)`

GetSchemaOk returns a tuple with the Schema field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchema

`func (o *IdentitySchema) SetSchema(v map[string]interface{})`

SetSchema sets Schema field to given value.

### HasSchema

`func (o *IdentitySchema) HasSchema() bool`

HasSchema returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


