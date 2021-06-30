# AdminUpdateIdentityBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SchemaId** | Pointer to **string** | SchemaID is the ID of the JSON Schema to be used for validating the identity&#39;s traits. If set will update the Identity&#39;s SchemaID. | [optional] 
**State** | **interface{}** | State is the identity&#39;s state. | 
**Traits** | **map[string]interface{}** | Traits represent an identity&#39;s traits. The identity is able to create, modify, and delete traits in a self-service manner. The input will always be validated against the JSON Schema defined in &#x60;schema_id&#x60;. | 

## Methods

### NewAdminUpdateIdentityBody

`func NewAdminUpdateIdentityBody(state interface{}, traits map[string]interface{}, ) *AdminUpdateIdentityBody`

NewAdminUpdateIdentityBody instantiates a new AdminUpdateIdentityBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAdminUpdateIdentityBodyWithDefaults

`func NewAdminUpdateIdentityBodyWithDefaults() *AdminUpdateIdentityBody`

NewAdminUpdateIdentityBodyWithDefaults instantiates a new AdminUpdateIdentityBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSchemaId

`func (o *AdminUpdateIdentityBody) GetSchemaId() string`

GetSchemaId returns the SchemaId field if non-nil, zero value otherwise.

### GetSchemaIdOk

`func (o *AdminUpdateIdentityBody) GetSchemaIdOk() (*string, bool)`

GetSchemaIdOk returns a tuple with the SchemaId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchemaId

`func (o *AdminUpdateIdentityBody) SetSchemaId(v string)`

SetSchemaId sets SchemaId field to given value.

### HasSchemaId

`func (o *AdminUpdateIdentityBody) HasSchemaId() bool`

HasSchemaId returns a boolean if a field has been set.

### GetState

`func (o *AdminUpdateIdentityBody) GetState() interface{}`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *AdminUpdateIdentityBody) GetStateOk() (*interface{}, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *AdminUpdateIdentityBody) SetState(v interface{})`

SetState sets State field to given value.


### SetStateNil

`func (o *AdminUpdateIdentityBody) SetStateNil(b bool)`

 SetStateNil sets the value for State to be an explicit nil

### UnsetState
`func (o *AdminUpdateIdentityBody) UnsetState()`

UnsetState ensures that no value is present for State, not even an explicit nil
### GetTraits

`func (o *AdminUpdateIdentityBody) GetTraits() map[string]interface{}`

GetTraits returns the Traits field if non-nil, zero value otherwise.

### GetTraitsOk

`func (o *AdminUpdateIdentityBody) GetTraitsOk() (*map[string]interface{}, bool)`

GetTraitsOk returns a tuple with the Traits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraits

`func (o *AdminUpdateIdentityBody) SetTraits(v map[string]interface{})`

SetTraits sets Traits field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


