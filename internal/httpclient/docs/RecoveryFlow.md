# RecoveryFlow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | Pointer to **string** | Active, if set, contains the registration method that is being used. It is initially not set. | [optional] 
**ExpiresAt** | **time.Time** | ExpiresAt is the time (UTC) when the request expires. If the user still wishes to update the setting, a new request has to be initiated. | 
**Id** | **string** |  | 
**IssuedAt** | **time.Time** | IssuedAt is the time (UTC) when the request occurred. | 
**RequestUrl** | **string** | RequestURL is the initial URL that was requested from Ory Kratos. It can be used to forward information contained in the URL&#39;s path or query for example. | 
**State** | [**RecoveryFlowState**](RecoveryFlowState.md) |  | 
**Type** | Pointer to **string** | The flow type can either be &#x60;api&#x60; or &#x60;browser&#x60;. | [optional] 
**Ui** | [**UiContainer**](UiContainer.md) |  | 

## Methods

### NewRecoveryFlow

`func NewRecoveryFlow(expiresAt time.Time, id string, issuedAt time.Time, requestUrl string, state RecoveryFlowState, ui UiContainer, ) *RecoveryFlow`

NewRecoveryFlow instantiates a new RecoveryFlow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRecoveryFlowWithDefaults

`func NewRecoveryFlowWithDefaults() *RecoveryFlow`

NewRecoveryFlowWithDefaults instantiates a new RecoveryFlow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActive

`func (o *RecoveryFlow) GetActive() string`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *RecoveryFlow) GetActiveOk() (*string, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *RecoveryFlow) SetActive(v string)`

SetActive sets Active field to given value.

### HasActive

`func (o *RecoveryFlow) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetExpiresAt

`func (o *RecoveryFlow) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *RecoveryFlow) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *RecoveryFlow) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.


### GetId

`func (o *RecoveryFlow) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RecoveryFlow) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RecoveryFlow) SetId(v string)`

SetId sets Id field to given value.


### GetIssuedAt

`func (o *RecoveryFlow) GetIssuedAt() time.Time`

GetIssuedAt returns the IssuedAt field if non-nil, zero value otherwise.

### GetIssuedAtOk

`func (o *RecoveryFlow) GetIssuedAtOk() (*time.Time, bool)`

GetIssuedAtOk returns a tuple with the IssuedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuedAt

`func (o *RecoveryFlow) SetIssuedAt(v time.Time)`

SetIssuedAt sets IssuedAt field to given value.


### GetRequestUrl

`func (o *RecoveryFlow) GetRequestUrl() string`

GetRequestUrl returns the RequestUrl field if non-nil, zero value otherwise.

### GetRequestUrlOk

`func (o *RecoveryFlow) GetRequestUrlOk() (*string, bool)`

GetRequestUrlOk returns a tuple with the RequestUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestUrl

`func (o *RecoveryFlow) SetRequestUrl(v string)`

SetRequestUrl sets RequestUrl field to given value.


### GetState

`func (o *RecoveryFlow) GetState() RecoveryFlowState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *RecoveryFlow) GetStateOk() (*RecoveryFlowState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *RecoveryFlow) SetState(v RecoveryFlowState)`

SetState sets State field to given value.


### GetType

`func (o *RecoveryFlow) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *RecoveryFlow) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *RecoveryFlow) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *RecoveryFlow) HasType() bool`

HasType returns a boolean if a field has been set.

### GetUi

`func (o *RecoveryFlow) GetUi() UiContainer`

GetUi returns the Ui field if non-nil, zero value otherwise.

### GetUiOk

`func (o *RecoveryFlow) GetUiOk() (*UiContainer, bool)`

GetUiOk returns a tuple with the Ui field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUi

`func (o *RecoveryFlow) SetUi(v UiContainer)`

SetUi sets Ui field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


