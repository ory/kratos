# SelfServiceRecoveryFlow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | Pointer to **string** | Active, if set, contains the registration method that is being used. It is initially not set. | [optional] 
**ExpiresAt** | **time.Time** | ExpiresAt is the time (UTC) when the request expires. If the user still wishes to update the setting, a new request has to be initiated. | 
**Id** | **string** |  | 
**IssuedAt** | **time.Time** | IssuedAt is the time (UTC) when the request occurred. | 
**RequestUrl** | **string** | RequestURL is the initial URL that was requested from Ory Kratos. It can be used to forward information contained in the URL&#39;s path or query for example. | 
**State** | [**SelfServiceRecoveryFlowState**](SelfServiceRecoveryFlowState.md) |  | 
**Type** | Pointer to **string** | The flow type can either be &#x60;api&#x60; or &#x60;browser&#x60;. | [optional] 
**Ui** | [**UiContainer**](UiContainer.md) |  | 

## Methods

### NewSelfServiceRecoveryFlow

`func NewSelfServiceRecoveryFlow(expiresAt time.Time, id string, issuedAt time.Time, requestUrl string, state SelfServiceRecoveryFlowState, ui UiContainer, ) *SelfServiceRecoveryFlow`

NewSelfServiceRecoveryFlow instantiates a new SelfServiceRecoveryFlow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceRecoveryFlowWithDefaults

`func NewSelfServiceRecoveryFlowWithDefaults() *SelfServiceRecoveryFlow`

NewSelfServiceRecoveryFlowWithDefaults instantiates a new SelfServiceRecoveryFlow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActive

`func (o *SelfServiceRecoveryFlow) GetActive() string`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *SelfServiceRecoveryFlow) GetActiveOk() (*string, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *SelfServiceRecoveryFlow) SetActive(v string)`

SetActive sets Active field to given value.

### HasActive

`func (o *SelfServiceRecoveryFlow) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetExpiresAt

`func (o *SelfServiceRecoveryFlow) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *SelfServiceRecoveryFlow) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *SelfServiceRecoveryFlow) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.


### GetId

`func (o *SelfServiceRecoveryFlow) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SelfServiceRecoveryFlow) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SelfServiceRecoveryFlow) SetId(v string)`

SetId sets Id field to given value.


### GetIssuedAt

`func (o *SelfServiceRecoveryFlow) GetIssuedAt() time.Time`

GetIssuedAt returns the IssuedAt field if non-nil, zero value otherwise.

### GetIssuedAtOk

`func (o *SelfServiceRecoveryFlow) GetIssuedAtOk() (*time.Time, bool)`

GetIssuedAtOk returns a tuple with the IssuedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuedAt

`func (o *SelfServiceRecoveryFlow) SetIssuedAt(v time.Time)`

SetIssuedAt sets IssuedAt field to given value.


### GetRequestUrl

`func (o *SelfServiceRecoveryFlow) GetRequestUrl() string`

GetRequestUrl returns the RequestUrl field if non-nil, zero value otherwise.

### GetRequestUrlOk

`func (o *SelfServiceRecoveryFlow) GetRequestUrlOk() (*string, bool)`

GetRequestUrlOk returns a tuple with the RequestUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestUrl

`func (o *SelfServiceRecoveryFlow) SetRequestUrl(v string)`

SetRequestUrl sets RequestUrl field to given value.


### GetState

`func (o *SelfServiceRecoveryFlow) GetState() SelfServiceRecoveryFlowState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *SelfServiceRecoveryFlow) GetStateOk() (*SelfServiceRecoveryFlowState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *SelfServiceRecoveryFlow) SetState(v SelfServiceRecoveryFlowState)`

SetState sets State field to given value.


### GetType

`func (o *SelfServiceRecoveryFlow) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *SelfServiceRecoveryFlow) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *SelfServiceRecoveryFlow) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *SelfServiceRecoveryFlow) HasType() bool`

HasType returns a boolean if a field has been set.

### GetUi

`func (o *SelfServiceRecoveryFlow) GetUi() UiContainer`

GetUi returns the Ui field if non-nil, zero value otherwise.

### GetUiOk

`func (o *SelfServiceRecoveryFlow) GetUiOk() (*UiContainer, bool)`

GetUiOk returns a tuple with the Ui field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUi

`func (o *SelfServiceRecoveryFlow) SetUi(v UiContainer)`

SetUi sets Ui field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


