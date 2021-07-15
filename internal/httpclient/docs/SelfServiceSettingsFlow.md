# SelfServiceSettingsFlow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | Pointer to **string** | Active, if set, contains the registration method that is being used. It is initially not set. | [optional] 
**ExpiresAt** | **time.Time** | ExpiresAt is the time (UTC) when the flow expires. If the user still wishes to update the setting, a new flow has to be initiated. | 
**Id** | **string** |  | 
**Identity** | [**Identity**](Identity.md) |  | 
**IssuedAt** | **time.Time** | IssuedAt is the time (UTC) when the flow occurred. | 
**RequestUrl** | **string** | RequestURL is the initial URL that was requested from Ory Kratos. It can be used to forward information contained in the URL&#39;s path or query for example. | 
**State** | [**SelfServiceSettingsFlowState**](SelfServiceSettingsFlowState.md) |  | 
**Type** | Pointer to **string** | The flow type can either be &#x60;api&#x60; or &#x60;browser&#x60;. | [optional] 
**Ui** | [**UiContainer**](UiContainer.md) |  | 

## Methods

### NewSelfServiceSettingsFlow

`func NewSelfServiceSettingsFlow(expiresAt time.Time, id string, identity Identity, issuedAt time.Time, requestUrl string, state SelfServiceSettingsFlowState, ui UiContainer, ) *SelfServiceSettingsFlow`

NewSelfServiceSettingsFlow instantiates a new SelfServiceSettingsFlow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceSettingsFlowWithDefaults

`func NewSelfServiceSettingsFlowWithDefaults() *SelfServiceSettingsFlow`

NewSelfServiceSettingsFlowWithDefaults instantiates a new SelfServiceSettingsFlow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActive

`func (o *SelfServiceSettingsFlow) GetActive() string`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *SelfServiceSettingsFlow) GetActiveOk() (*string, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *SelfServiceSettingsFlow) SetActive(v string)`

SetActive sets Active field to given value.

### HasActive

`func (o *SelfServiceSettingsFlow) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetExpiresAt

`func (o *SelfServiceSettingsFlow) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *SelfServiceSettingsFlow) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *SelfServiceSettingsFlow) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.


### GetId

`func (o *SelfServiceSettingsFlow) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SelfServiceSettingsFlow) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SelfServiceSettingsFlow) SetId(v string)`

SetId sets Id field to given value.


### GetIdentity

`func (o *SelfServiceSettingsFlow) GetIdentity() Identity`

GetIdentity returns the Identity field if non-nil, zero value otherwise.

### GetIdentityOk

`func (o *SelfServiceSettingsFlow) GetIdentityOk() (*Identity, bool)`

GetIdentityOk returns a tuple with the Identity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentity

`func (o *SelfServiceSettingsFlow) SetIdentity(v Identity)`

SetIdentity sets Identity field to given value.


### GetIssuedAt

`func (o *SelfServiceSettingsFlow) GetIssuedAt() time.Time`

GetIssuedAt returns the IssuedAt field if non-nil, zero value otherwise.

### GetIssuedAtOk

`func (o *SelfServiceSettingsFlow) GetIssuedAtOk() (*time.Time, bool)`

GetIssuedAtOk returns a tuple with the IssuedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuedAt

`func (o *SelfServiceSettingsFlow) SetIssuedAt(v time.Time)`

SetIssuedAt sets IssuedAt field to given value.


### GetRequestUrl

`func (o *SelfServiceSettingsFlow) GetRequestUrl() string`

GetRequestUrl returns the RequestUrl field if non-nil, zero value otherwise.

### GetRequestUrlOk

`func (o *SelfServiceSettingsFlow) GetRequestUrlOk() (*string, bool)`

GetRequestUrlOk returns a tuple with the RequestUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestUrl

`func (o *SelfServiceSettingsFlow) SetRequestUrl(v string)`

SetRequestUrl sets RequestUrl field to given value.


### GetState

`func (o *SelfServiceSettingsFlow) GetState() SelfServiceSettingsFlowState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *SelfServiceSettingsFlow) GetStateOk() (*SelfServiceSettingsFlowState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *SelfServiceSettingsFlow) SetState(v SelfServiceSettingsFlowState)`

SetState sets State field to given value.


### GetType

`func (o *SelfServiceSettingsFlow) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *SelfServiceSettingsFlow) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *SelfServiceSettingsFlow) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *SelfServiceSettingsFlow) HasType() bool`

HasType returns a boolean if a field has been set.

### GetUi

`func (o *SelfServiceSettingsFlow) GetUi() UiContainer`

GetUi returns the Ui field if non-nil, zero value otherwise.

### GetUiOk

`func (o *SelfServiceSettingsFlow) GetUiOk() (*UiContainer, bool)`

GetUiOk returns a tuple with the Ui field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUi

`func (o *SelfServiceSettingsFlow) SetUi(v UiContainer)`

SetUi sets Ui field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


