# SubmitSelfServiceVerificationFlowBody

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** | Email is the email to which to send the verification message to. | 
**Method** | **string** | Method supports &#x60;link&#x60; only right now. | 

## Methods

### NewSubmitSelfServiceVerificationFlowBody

`func NewSubmitSelfServiceVerificationFlowBody(email string, method string, ) *SubmitSelfServiceVerificationFlowBody`

NewSubmitSelfServiceVerificationFlowBody instantiates a new SubmitSelfServiceVerificationFlowBody object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitSelfServiceVerificationFlowBodyWithDefaults

`func NewSubmitSelfServiceVerificationFlowBodyWithDefaults() *SubmitSelfServiceVerificationFlowBody`

NewSubmitSelfServiceVerificationFlowBodyWithDefaults instantiates a new SubmitSelfServiceVerificationFlowBody object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *SubmitSelfServiceVerificationFlowBody) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *SubmitSelfServiceVerificationFlowBody) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *SubmitSelfServiceVerificationFlowBody) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetMethod

`func (o *SubmitSelfServiceVerificationFlowBody) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *SubmitSelfServiceVerificationFlowBody) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *SubmitSelfServiceVerificationFlowBody) SetMethod(v string)`

SetMethod sets Method field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


