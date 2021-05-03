# LoginFlow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | Pointer to **string** | and so on. | [optional] 
**ExpiresAt** | **time.Time** | ExpiresAt is the time (UTC) when the flow expires. If the user still wishes to log in, a new flow has to be initiated. | 
**Forced** | Pointer to **bool** | Forced stores whether this login flow should enforce re-authentication. | [optional] 
**Id** | **string** |  | 
**IssuedAt** | **time.Time** | IssuedAt is the time (UTC) when the flow started. | 
**RequestUrl** | **string** | RequestURL is the initial URL that was requested from Ory Kratos. It can be used to forward information contained in the URL&#39;s path or query for example. | 
**Type** | **string** | The flow type can either be &#x60;api&#x60; or &#x60;browser&#x60;. | 
**Ui** | [**UiContainer**](UiContainer.md) |  | 

## Methods

### NewLoginFlow

`func NewLoginFlow(expiresAt time.Time, id string, issuedAt time.Time, requestUrl string, type_ string, ui UiContainer, ) *LoginFlow`

NewLoginFlow instantiates a new LoginFlow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLoginFlowWithDefaults

`func NewLoginFlowWithDefaults() *LoginFlow`

NewLoginFlowWithDefaults instantiates a new LoginFlow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActive

`func (o *LoginFlow) GetActive() string`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *LoginFlow) GetActiveOk() (*string, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *LoginFlow) SetActive(v string)`

SetActive sets Active field to given value.

### HasActive

`func (o *LoginFlow) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetExpiresAt

`func (o *LoginFlow) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *LoginFlow) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *LoginFlow) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.


### GetForced

`func (o *LoginFlow) GetForced() bool`

GetForced returns the Forced field if non-nil, zero value otherwise.

### GetForcedOk

`func (o *LoginFlow) GetForcedOk() (*bool, bool)`

GetForcedOk returns a tuple with the Forced field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForced

`func (o *LoginFlow) SetForced(v bool)`

SetForced sets Forced field to given value.

### HasForced

`func (o *LoginFlow) HasForced() bool`

HasForced returns a boolean if a field has been set.

### GetId

`func (o *LoginFlow) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *LoginFlow) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *LoginFlow) SetId(v string)`

SetId sets Id field to given value.


### GetIssuedAt

`func (o *LoginFlow) GetIssuedAt() time.Time`

GetIssuedAt returns the IssuedAt field if non-nil, zero value otherwise.

### GetIssuedAtOk

`func (o *LoginFlow) GetIssuedAtOk() (*time.Time, bool)`

GetIssuedAtOk returns a tuple with the IssuedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssuedAt

`func (o *LoginFlow) SetIssuedAt(v time.Time)`

SetIssuedAt sets IssuedAt field to given value.


### GetRequestUrl

`func (o *LoginFlow) GetRequestUrl() string`

GetRequestUrl returns the RequestUrl field if non-nil, zero value otherwise.

### GetRequestUrlOk

`func (o *LoginFlow) GetRequestUrlOk() (*string, bool)`

GetRequestUrlOk returns a tuple with the RequestUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestUrl

`func (o *LoginFlow) SetRequestUrl(v string)`

SetRequestUrl sets RequestUrl field to given value.


### GetType

`func (o *LoginFlow) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *LoginFlow) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *LoginFlow) SetType(v string)`

SetType sets Type field to given value.


### GetUi

`func (o *LoginFlow) GetUi() UiContainer`

GetUi returns the Ui field if non-nil, zero value otherwise.

### GetUiOk

`func (o *LoginFlow) GetUiOk() (*UiContainer, bool)`

GetUiOk returns a tuple with the Ui field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUi

`func (o *LoginFlow) SetUi(v UiContainer)`

SetUi sets Ui field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


