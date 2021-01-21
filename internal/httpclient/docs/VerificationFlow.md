# VerificationFlow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | Pointer to **string** | Active, if set, contains the registration method that is being used. It is initially not set. | [optional] 
**ExpiresAt** | Pointer to **time.Time** | ExpiresAt is the time (UTC) when the request expires. If the user still wishes to verify the address, a new request has to be initiated. | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**IssuedAt** | Pointer to **time.Time** | IssuedAt is the time (UTC) when the request occurred. | [optional] 
**Messages** | Pointer to [**[]UiText**](UiText.md) |  | [optional] 
**Methods** | [**map[string]VerificationFlowMethod**](verificationFlowMethod.md) | Methods contains context for all account verification methods. If a registration request has been processed, but for example the password is incorrect, this will contain error messages. | 
**RequestUrl** | Pointer to **string** | RequestURL is the initial URL that was requested from ORY Kratos. It can be used to forward information contained in the URL&#39;s path or query for example. | [optional] 
**State** | **string** |  | 
**Type** | Pointer to **string** | The flow type can either be &#x60;api&#x60; or &#x60;browser&#x60;. | [optional] 

## Methods

### NewVerificationFlow

`func NewVerificationFlow(methods map[string]VerificationFlowMethod, state string, ) *VerificationFlow`

NewVerificationFlow instantiates a new VerificationFlow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVerificationFlowWithDefaults

`func NewVerificationFlowWithDefaults() *VerificationFlow`

NewVerificationFlowWithDefaults instantiates a new VerificationFlow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActive

`func (o *VerificationFlow) GetActive() string`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *VerificationFlow) GetActiveOk() (*string, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *VerificationFlow) SetActive(v string)`

SetActive sets Active field to given value.

### HasActive

`func (o *VerificationFlow) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetExpiresAt

`func (o *VerificationFlow) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *VerificationFlow) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *VerificationFlow) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *VerificationFlow) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.

### GetId

`func (o *VerificationFlow) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *VerificationFlow) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *VerificationFlow) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *VerificationFlow) HasId() bool`

HasId returns a boolean if a field has been set.

### GetIssuedAt

`func (o *VerificationFlow) GetIssuedAt() time.Time`

GetIssuedAt returns the IssuedAt field if non-nil, zero value otherwise.

### GetIssuedAtOk

`func (o *VerificationFlow) GetIssuedAtOk() (*time.Time, bool)`

GetIssuedAtOk returns a tuple with the IssuedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuedAt

`func (o *VerificationFlow) SetIssuedAt(v time.Time)`

SetIssuedAt sets IssuedAt field to given value.

### HasIssuedAt

`func (o *VerificationFlow) HasIssuedAt() bool`

HasIssuedAt returns a boolean if a field has been set.

### GetMessages

`func (o *VerificationFlow) GetMessages() []UiText`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *VerificationFlow) GetMessagesOk() (*[]UiText, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *VerificationFlow) SetMessages(v []UiText)`

SetMessages sets Messages field to given value.

### HasMessages

`func (o *VerificationFlow) HasMessages() bool`

HasMessages returns a boolean if a field has been set.

### GetMethods

`func (o *VerificationFlow) GetMethods() map[string]VerificationFlowMethod`

GetMethods returns the Methods field if non-nil, zero value otherwise.

### GetMethodsOk

`func (o *VerificationFlow) GetMethodsOk() (*map[string]VerificationFlowMethod, bool)`

GetMethodsOk returns a tuple with the Methods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethods

`func (o *VerificationFlow) SetMethods(v map[string]VerificationFlowMethod)`

SetMethods sets Methods field to given value.


### GetRequestUrl

`func (o *VerificationFlow) GetRequestUrl() string`

GetRequestUrl returns the RequestUrl field if non-nil, zero value otherwise.

### GetRequestUrlOk

`func (o *VerificationFlow) GetRequestUrlOk() (*string, bool)`

GetRequestUrlOk returns a tuple with the RequestUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestUrl

`func (o *VerificationFlow) SetRequestUrl(v string)`

SetRequestUrl sets RequestUrl field to given value.

### HasRequestUrl

`func (o *VerificationFlow) HasRequestUrl() bool`

HasRequestUrl returns a boolean if a field has been set.

### GetState

`func (o *VerificationFlow) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *VerificationFlow) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *VerificationFlow) SetState(v string)`

SetState sets State field to given value.


### GetType

`func (o *VerificationFlow) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *VerificationFlow) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *VerificationFlow) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *VerificationFlow) HasType() bool`

HasType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


