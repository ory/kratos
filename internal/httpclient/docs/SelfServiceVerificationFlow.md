# SelfServiceVerificationFlow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | Pointer to **string** | Active, if set, contains the registration method that is being used. It is initially not set. | [optional] 
**ExpiresAt** | Pointer to **time.Time** | ExpiresAt is the time (UTC) when the request expires. If the user still wishes to verify the address, a new request has to be initiated. | [optional] 
**Id** | **string** |  | 
**IssuedAt** | Pointer to **time.Time** | IssuedAt is the time (UTC) when the request occurred. | [optional] 
**RequestUrl** | Pointer to **string** | RequestURL is the initial URL that was requested from Ory Kratos. It can be used to forward information contained in the URL&#39;s path or query for example. | [optional] 
**State** | [**SelfServiceVerificationFlowState**](SelfServiceVerificationFlowState.md) |  | 
**Type** | **string** | The flow type can either be &#x60;api&#x60; or &#x60;browser&#x60;. | 
**Ui** | [**UiContainer**](UiContainer.md) |  | 

## Methods

### NewSelfServiceVerificationFlow

`func NewSelfServiceVerificationFlow(id string, state SelfServiceVerificationFlowState, type_ string, ui UiContainer, ) *SelfServiceVerificationFlow`

NewSelfServiceVerificationFlow instantiates a new SelfServiceVerificationFlow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelfServiceVerificationFlowWithDefaults

`func NewSelfServiceVerificationFlowWithDefaults() *SelfServiceVerificationFlow`

NewSelfServiceVerificationFlowWithDefaults instantiates a new SelfServiceVerificationFlow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActive

`func (o *SelfServiceVerificationFlow) GetActive() string`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *SelfServiceVerificationFlow) GetActiveOk() (*string, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *SelfServiceVerificationFlow) SetActive(v string)`

SetActive sets Active field to given value.

### HasActive

`func (o *SelfServiceVerificationFlow) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetExpiresAt

`func (o *SelfServiceVerificationFlow) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *SelfServiceVerificationFlow) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *SelfServiceVerificationFlow) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *SelfServiceVerificationFlow) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.

### GetId

`func (o *SelfServiceVerificationFlow) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SelfServiceVerificationFlow) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SelfServiceVerificationFlow) SetId(v string)`

SetId sets Id field to given value.


### GetIssuedAt

`func (o *SelfServiceVerificationFlow) GetIssuedAt() time.Time`

GetIssuedAt returns the IssuedAt field if non-nil, zero value otherwise.

### GetIssuedAtOk

`func (o *SelfServiceVerificationFlow) GetIssuedAtOk() (*time.Time, bool)`

GetIssuedAtOk returns a tuple with the IssuedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuedAt

`func (o *SelfServiceVerificationFlow) SetIssuedAt(v time.Time)`

SetIssuedAt sets IssuedAt field to given value.

### HasIssuedAt

`func (o *SelfServiceVerificationFlow) HasIssuedAt() bool`

HasIssuedAt returns a boolean if a field has been set.

### GetRequestUrl

`func (o *SelfServiceVerificationFlow) GetRequestUrl() string`

GetRequestUrl returns the RequestUrl field if non-nil, zero value otherwise.

### GetRequestUrlOk

`func (o *SelfServiceVerificationFlow) GetRequestUrlOk() (*string, bool)`

GetRequestUrlOk returns a tuple with the RequestUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestUrl

`func (o *SelfServiceVerificationFlow) SetRequestUrl(v string)`

SetRequestUrl sets RequestUrl field to given value.

### HasRequestUrl

`func (o *SelfServiceVerificationFlow) HasRequestUrl() bool`

HasRequestUrl returns a boolean if a field has been set.

### GetState

`func (o *SelfServiceVerificationFlow) GetState() SelfServiceVerificationFlowState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *SelfServiceVerificationFlow) GetStateOk() (*SelfServiceVerificationFlowState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *SelfServiceVerificationFlow) SetState(v SelfServiceVerificationFlowState)`

SetState sets State field to given value.


### GetType

`func (o *SelfServiceVerificationFlow) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *SelfServiceVerificationFlow) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *SelfServiceVerificationFlow) SetType(v string)`

SetType sets Type field to given value.


### GetUi

`func (o *SelfServiceVerificationFlow) GetUi() UiContainer`

GetUi returns the Ui field if non-nil, zero value otherwise.

### GetUiOk

`func (o *SelfServiceVerificationFlow) GetUiOk() (*UiContainer, bool)`

GetUiOk returns a tuple with the Ui field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUi

`func (o *SelfServiceVerificationFlow) SetUi(v UiContainer)`

SetUi sets Ui field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


