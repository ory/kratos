# SubmitSelfServiceRecoveryFlowWithLinkMethod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CsrfToken** | Pointer to **string** | Sending the anti-csrf token is only required for browser login flows. | [optional] 
**Email** | Pointer to **string** | Email to Recover  Needs to be set when initiating the flow. If the email is a registered recovery email, a recovery link will be sent. If the email is not known, a email with details on what happened will be sent instead.  format: email in: body | [optional] 

## Methods

### NewSubmitSelfServiceRecoveryFlowWithLinkMethod

`func NewSubmitSelfServiceRecoveryFlowWithLinkMethod() *SubmitSelfServiceRecoveryFlowWithLinkMethod`

NewSubmitSelfServiceRecoveryFlowWithLinkMethod instantiates a new SubmitSelfServiceRecoveryFlowWithLinkMethod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceRecoveryFlowWithLinkMethodWithDefaults

`func NewSubmitSelfServiceRecoveryFlowWithLinkMethodWithDefaults() *SubmitSelfServiceRecoveryFlowWithLinkMethod`

NewSubmitSelfServiceRecoveryFlowWithLinkMethodWithDefaults instantiates a new SubmitSelfServiceRecoveryFlowWithLinkMethod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCsrfToken

`func (o *SubmitSelfServiceRecoveryFlowWithLinkMethod) GetCsrfToken() string`

GetCsrfToken returns the CsrfToken field if non-nil, zero value otherwise.

### GetCsrfTokenOk

`func (o *SubmitSelfServiceRecoveryFlowWithLinkMethod) GetCsrfTokenOk() (*string, bool)`

GetCsrfTokenOk returns a tuple with the CsrfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsrfToken

`func (o *SubmitSelfServiceRecoveryFlowWithLinkMethod) SetCsrfToken(v string)`

SetCsrfToken sets CsrfToken field to given value.

### HasCsrfToken

`func (o *SubmitSelfServiceRecoveryFlowWithLinkMethod) HasCsrfToken() bool`

HasCsrfToken returns a boolean if a field has been set.

### GetEmail

`func (o *SubmitSelfServiceRecoveryFlowWithLinkMethod) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *SubmitSelfServiceRecoveryFlowWithLinkMethod) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *SubmitSelfServiceRecoveryFlowWithLinkMethod) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *SubmitSelfServiceRecoveryFlowWithLinkMethod) HasEmail() bool`

HasEmail returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


