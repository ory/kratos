# ErrorAuthenticatorAssuranceLevelNotSatisfied

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **int64** | The status code | [optional] 
**Debug** | Pointer to **string** | Debug information  This field is often not exposed to protect against leaking sensitive information. | [optional] 
**Details** | Pointer to **map[string]interface{}** | Further error details | [optional] 
**Id** | Pointer to **string** | The error ID  Useful when trying to identify various errors in application logic. | [optional] 
**Message** | **string** | Error message  The error&#39;s message. | 
**Reason** | Pointer to **string** | A human-readable reason for the error | [optional] 
**RedirectBrowserTo** | Pointer to **string** |  | [optional] 
**Request** | Pointer to **string** | The request ID  The request ID is often exposed internally in order to trace errors across service architectures. This is often a UUID. | [optional] 
**Status** | Pointer to **string** | The status description | [optional] 

## Methods

### NewErrorAuthenticatorAssuranceLevelNotSatisfied

`func NewErrorAuthenticatorAssuranceLevelNotSatisfied(message string, ) *ErrorAuthenticatorAssuranceLevelNotSatisfied`

NewErrorAuthenticatorAssuranceLevelNotSatisfied instantiates a new ErrorAuthenticatorAssuranceLevelNotSatisfied object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewErrorAuthenticatorAssuranceLevelNotSatisfiedWithDefaults

`func NewErrorAuthenticatorAssuranceLevelNotSatisfiedWithDefaults() *ErrorAuthenticatorAssuranceLevelNotSatisfied`

NewErrorAuthenticatorAssuranceLevelNotSatisfiedWithDefaults instantiates a new ErrorAuthenticatorAssuranceLevelNotSatisfied object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetCode() int64`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetCodeOk() (*int64, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) SetCode(v int64)`

SetCode sets Code field to given value.

### HasCode

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetDebug

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetDebug() string`

GetDebug returns the Debug field if non-nil, zero value otherwise.

### GetDebugOk

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetDebugOk() (*string, bool)`

GetDebugOk returns a tuple with the Debug field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDebug

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) SetDebug(v string)`

SetDebug sets Debug field to given value.

### HasDebug

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) HasDebug() bool`

HasDebug returns a boolean if a field has been set.

### GetDetails

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetDetails() map[string]interface{}`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetDetailsOk() (*map[string]interface{}, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) SetDetails(v map[string]interface{})`

SetDetails sets Details field to given value.

### HasDetails

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) HasDetails() bool`

HasDetails returns a boolean if a field has been set.

### GetId

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMessage

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetReason

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetRedirectBrowserTo

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetRedirectBrowserTo() string`

GetRedirectBrowserTo returns the RedirectBrowserTo field if non-nil, zero value otherwise.

### GetRedirectBrowserToOk

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetRedirectBrowserToOk() (*string, bool)`

GetRedirectBrowserToOk returns a tuple with the RedirectBrowserTo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRedirectBrowserTo

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) SetRedirectBrowserTo(v string)`

SetRedirectBrowserTo sets RedirectBrowserTo field to given value.

### HasRedirectBrowserTo

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) HasRedirectBrowserTo() bool`

HasRedirectBrowserTo returns a boolean if a field has been set.

### GetRequest

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetRequest() string`

GetRequest returns the Request field if non-nil, zero value otherwise.

### GetRequestOk

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetRequestOk() (*string, bool)`

GetRequestOk returns a tuple with the Request field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequest

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) SetRequest(v string)`

SetRequest sets Request field to given value.

### HasRequest

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) HasRequest() bool`

HasRequest returns a boolean if a field has been set.

### GetStatus

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ErrorAuthenticatorAssuranceLevelNotSatisfied) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


