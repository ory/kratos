# SelfServiceLoginFlow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | Pointer to **string** | and so on. | [optional] 
**CreatedAt** | Pointer to **time.Time** | CreatedAt is a helper struct field for gobuffalo.pop. | [optional] 
**ExpiresAt** | **time.Time** | ExpiresAt is the time (UTC) when the flow expires. If the user still wishes to log in, a new flow has to be initiated. | 
**Forced** | Pointer to **bool** | Forced stores whether this login flow should enforce re-authentication. | [optional] 
**Id** | **string** |  | 
**IssuedAt** | **time.Time** | IssuedAt is the time (UTC) when the flow started. | 
**RequestUrl** | **string** | RequestURL is the initial URL that was requested from Ory Kratos. It can be used to forward information contained in the URL&#39;s path or query for example. | 
**Type** | **string** | The flow type can either be &#x60;api&#x60; or &#x60;browser&#x60;. | 
**Ui** | [**UiContainer**](UiContainer.md) |  | 
**UpdatedAt** | Pointer to **time.Time** | UpdatedAt is a helper struct field for gobuffalo.pop. | [optional] 

## Methods

### NewSelfServiceLoginFlow

`func NewSelfServiceLoginFlow(expiresAt time.Time, id string, issuedAt time.Time, requestUrl string, type_ string, ui UiContainer, ) *SelfServiceLoginFlow`

NewSelfServiceLoginFlow instantiates a new SelfServiceLoginFlow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceLoginFlowWithDefaults

`func NewSelfServiceLoginFlowWithDefaults() *SelfServiceLoginFlow`

NewSelfServiceLoginFlowWithDefaults instantiates a new SelfServiceLoginFlow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActive

`func (o *SelfServiceLoginFlow) GetActive() string`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *SelfServiceLoginFlow) GetActiveOk() (*string, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *SelfServiceLoginFlow) SetActive(v string)`

SetActive sets Active field to given value.

### HasActive

`func (o *SelfServiceLoginFlow) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetCreatedAt

`func (o *SelfServiceLoginFlow) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *SelfServiceLoginFlow) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *SelfServiceLoginFlow) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *SelfServiceLoginFlow) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetExpiresAt

`func (o *SelfServiceLoginFlow) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *SelfServiceLoginFlow) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *SelfServiceLoginFlow) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.


### GetForced

`func (o *SelfServiceLoginFlow) GetForced() bool`

GetForced returns the Forced field if non-nil, zero value otherwise.

### GetForcedOk

`func (o *SelfServiceLoginFlow) GetForcedOk() (*bool, bool)`

GetForcedOk returns a tuple with the Forced field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForced

`func (o *SelfServiceLoginFlow) SetForced(v bool)`

SetForced sets Forced field to given value.

### HasForced

`func (o *SelfServiceLoginFlow) HasForced() bool`

HasForced returns a boolean if a field has been set.

### GetId

`func (o *SelfServiceLoginFlow) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SelfServiceLoginFlow) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SelfServiceLoginFlow) SetId(v string)`

SetId sets Id field to given value.


### GetIssuedAt

`func (o *SelfServiceLoginFlow) GetIssuedAt() time.Time`

GetIssuedAt returns the IssuedAt field if non-nil, zero value otherwise.

### GetIssuedAtOk

`func (o *SelfServiceLoginFlow) GetIssuedAtOk() (*time.Time, bool)`

GetIssuedAtOk returns a tuple with the IssuedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuedAt

`func (o *SelfServiceLoginFlow) SetIssuedAt(v time.Time)`

SetIssuedAt sets IssuedAt field to given value.


### GetRequestUrl

`func (o *SelfServiceLoginFlow) GetRequestUrl() string`

GetRequestUrl returns the RequestUrl field if non-nil, zero value otherwise.

### GetRequestUrlOk

`func (o *SelfServiceLoginFlow) GetRequestUrlOk() (*string, bool)`

GetRequestUrlOk returns a tuple with the RequestUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestUrl

`func (o *SelfServiceLoginFlow) SetRequestUrl(v string)`

SetRequestUrl sets RequestUrl field to given value.


### GetType

`func (o *SelfServiceLoginFlow) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *SelfServiceLoginFlow) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *SelfServiceLoginFlow) SetType(v string)`

SetType sets Type field to given value.


### GetUi

`func (o *SelfServiceLoginFlow) GetUi() UiContainer`

GetUi returns the Ui field if non-nil, zero value otherwise.

### GetUiOk

`func (o *SelfServiceLoginFlow) GetUiOk() (*UiContainer, bool)`

GetUiOk returns a tuple with the Ui field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUi

`func (o *SelfServiceLoginFlow) SetUi(v UiContainer)`

SetUi sets Ui field to given value.


### GetUpdatedAt

`func (o *SelfServiceLoginFlow) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *SelfServiceLoginFlow) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *SelfServiceLoginFlow) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *SelfServiceLoginFlow) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


