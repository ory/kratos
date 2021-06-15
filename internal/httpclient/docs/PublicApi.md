# \PublicApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetSchema**](PublicApi.md#GetSchema) | **Get** /schemas/{id} | 
[**GetSelfServiceError**](PublicApi.md#GetSelfServiceError) | **Get** /self-service/errors | Get User-Facing Self-Service Errors
[**GetSelfServiceLoginFlow**](PublicApi.md#GetSelfServiceLoginFlow) | **Get** /self-service/login/flows | Get Login Flow
[**GetSelfServiceRecoveryFlow**](PublicApi.md#GetSelfServiceRecoveryFlow) | **Get** /self-service/recovery/flows | Get Recovery Flow
[**GetSelfServiceRegistrationFlow**](PublicApi.md#GetSelfServiceRegistrationFlow) | **Get** /self-service/registration/flows | Get Registration Flow
[**GetSelfServiceSettingsFlow**](PublicApi.md#GetSelfServiceSettingsFlow) | **Get** /self-service/settings/flows | Get Settings Flow
[**GetSelfServiceVerificationFlow**](PublicApi.md#GetSelfServiceVerificationFlow) | **Get** /self-service/verification/flows | Get Verification Flow
[**InitializeSelfServiceBrowserLogoutFlow**](PublicApi.md#InitializeSelfServiceBrowserLogoutFlow) | **Get** /self-service/browser/flows/logout | Initialize Browser-Based Logout User Flow
[**InitializeSelfServiceLoginForBrowsers**](PublicApi.md#InitializeSelfServiceLoginForBrowsers) | **Get** /self-service/login/browser | Initialize Login Flow for Browsers
[**InitializeSelfServiceLoginWithoutBrowser**](PublicApi.md#InitializeSelfServiceLoginWithoutBrowser) | **Get** /self-service/login/api | Initialize Login Flow for APIs, Services, Apps, ...
[**InitializeSelfServiceRecoveryForBrowsers**](PublicApi.md#InitializeSelfServiceRecoveryForBrowsers) | **Get** /self-service/recovery/browser | Initialize Recovery Flow for Browsers
[**InitializeSelfServiceRecoveryWithoutBrowser**](PublicApi.md#InitializeSelfServiceRecoveryWithoutBrowser) | **Get** /self-service/recovery/api | Initialize Recovery Flow for APIs, Services, Apps, ...
[**InitializeSelfServiceRegistrationForBrowsers**](PublicApi.md#InitializeSelfServiceRegistrationForBrowsers) | **Get** /self-service/registration/browser | Initialize Registration Flow for Browsers
[**InitializeSelfServiceRegistrationWithoutBrowser**](PublicApi.md#InitializeSelfServiceRegistrationWithoutBrowser) | **Get** /self-service/registration/api | Initialize Registration Flow for APIs, Services, Apps, ...
[**InitializeSelfServiceSettingsForBrowsers**](PublicApi.md#InitializeSelfServiceSettingsForBrowsers) | **Get** /self-service/settings/browser | Initialize Settings Flow for Browsers
[**InitializeSelfServiceSettingsWithoutBrowser**](PublicApi.md#InitializeSelfServiceSettingsWithoutBrowser) | **Get** /self-service/settings/api | Initialize Settings Flow for APIs, Services, Apps, ...
[**InitializeSelfServiceVerificationForBrowsers**](PublicApi.md#InitializeSelfServiceVerificationForBrowsers) | **Get** /self-service/verification/browser | Initialize Verification Flow for Browser Clients
[**InitializeSelfServiceVerificationWithoutBrowser**](PublicApi.md#InitializeSelfServiceVerificationWithoutBrowser) | **Get** /self-service/verification/api | Initialize Verification Flow for APIs, Services, Apps, ...
[**RevokeSession**](PublicApi.md#RevokeSession) | **Delete** /sessions | Initialize Logout Flow for API Clients - Revoke a Session
[**SubmitSelfServiceLoginFlow**](PublicApi.md#SubmitSelfServiceLoginFlow) | **Post** /self-service/login | Submit a Login Flow
[**SubmitSelfServiceRecoveryFlow**](PublicApi.md#SubmitSelfServiceRecoveryFlow) | **Post** /self-service/recovery | Complete Recovery Flow
[**SubmitSelfServiceRecoveryFlowWithLinkMethod**](PublicApi.md#SubmitSelfServiceRecoveryFlowWithLinkMethod) | **Post** /self-service/recovery/methods/link | Complete Recovery Flow with Link Method
[**SubmitSelfServiceRegistrationFlow**](PublicApi.md#SubmitSelfServiceRegistrationFlow) | **Post** /self-service/registration | Submit a Registration Flow
[**SubmitSelfServiceSettingsFlow**](PublicApi.md#SubmitSelfServiceSettingsFlow) | **Post** /self-service/settings | Complete Settings Flow
[**SubmitSelfServiceVerificationFlow**](PublicApi.md#SubmitSelfServiceVerificationFlow) | **Post** /self-service/verification/flows | Complete Verification Flow
[**ToSession**](PublicApi.md#ToSession) | **Get** /sessions/whoami | Check Who the Current HTTP Session Belongs To



## GetSchema

> map[string]interface{} GetSchema(ctx, id).Execute()





### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | ID must be set to the ID of schema you want to get

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.GetSchema(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.GetSchema``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSchema`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.GetSchema`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID must be set to the ID of schema you want to get | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSchemaRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceError

> SelfServiceErrorContainer GetSelfServiceError(ctx).Error_(error_).Execute()

Get User-Facing Self-Service Errors



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    error_ := "error__example" // string | Error is the container's ID

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.GetSelfServiceError(context.Background()).Error_(error_).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.GetSelfServiceError``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceError`: SelfServiceErrorContainer
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.GetSelfServiceError`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceErrorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **error_** | **string** | Error is the container&#39;s ID | 

### Return type

[**SelfServiceErrorContainer**](SelfServiceErrorContainer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceLoginFlow

> LoginFlow GetSelfServiceLoginFlow(ctx).Id(id).Execute()

Get Login Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | The Login Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/login?flow=abcde`).

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.GetSelfServiceLoginFlow(context.Background()).Id(id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.GetSelfServiceLoginFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceLoginFlow`: LoginFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.GetSelfServiceLoginFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceLoginFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | The Login Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/login?flow&#x3D;abcde&#x60;). | 

### Return type

[**LoginFlow**](LoginFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceRecoveryFlow

> RecoveryFlow GetSelfServiceRecoveryFlow(ctx).Id(id).Execute()

Get Recovery Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | The Flow ID  The value for this parameter comes from `request` URL Query parameter sent to your application (e.g. `/recovery?flow=abcde`).

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.GetSelfServiceRecoveryFlow(context.Background()).Id(id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.GetSelfServiceRecoveryFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceRecoveryFlow`: RecoveryFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.GetSelfServiceRecoveryFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceRecoveryFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | The Flow ID  The value for this parameter comes from &#x60;request&#x60; URL Query parameter sent to your application (e.g. &#x60;/recovery?flow&#x3D;abcde&#x60;). | 

### Return type

[**RecoveryFlow**](RecoveryFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceRegistrationFlow

> RegistrationFlow GetSelfServiceRegistrationFlow(ctx).Id(id).Execute()

Get Registration Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | The Registration Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/registration?flow=abcde`).

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.GetSelfServiceRegistrationFlow(context.Background()).Id(id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.GetSelfServiceRegistrationFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceRegistrationFlow`: RegistrationFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.GetSelfServiceRegistrationFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceRegistrationFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | The Registration Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/registration?flow&#x3D;abcde&#x60;). | 

### Return type

[**RegistrationFlow**](RegistrationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceSettingsFlow

> SettingsFlow GetSelfServiceSettingsFlow(ctx).Id(id).XSessionToken(xSessionToken).Execute()

Get Settings Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | ID is the Settings Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/settings?flow=abcde`).
    xSessionToken := "xSessionToken_example" // string | The Session Token of the Identity performing the settings flow. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.GetSelfServiceSettingsFlow(context.Background()).Id(id).XSessionToken(xSessionToken).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.GetSelfServiceSettingsFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceSettingsFlow`: SettingsFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.GetSelfServiceSettingsFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceSettingsFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | ID is the Settings Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/settings?flow&#x3D;abcde&#x60;). | 
 **xSessionToken** | **string** | The Session Token of the Identity performing the settings flow. | 

### Return type

[**SettingsFlow**](SettingsFlow.md)

### Authorization

[sessionToken](../README.md#sessionToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceVerificationFlow

> VerificationFlow GetSelfServiceVerificationFlow(ctx).Id(id).Execute()

Get Verification Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | The Flow ID  The value for this parameter comes from `request` URL Query parameter sent to your application (e.g. `/verification?flow=abcde`).

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.GetSelfServiceVerificationFlow(context.Background()).Id(id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.GetSelfServiceVerificationFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceVerificationFlow`: VerificationFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.GetSelfServiceVerificationFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceVerificationFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | The Flow ID  The value for this parameter comes from &#x60;request&#x60; URL Query parameter sent to your application (e.g. &#x60;/verification?flow&#x3D;abcde&#x60;). | 

### Return type

[**VerificationFlow**](VerificationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceBrowserLogoutFlow

> InitializeSelfServiceBrowserLogoutFlow(ctx).Execute()

Initialize Browser-Based Logout User Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceBrowserLogoutFlow(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceBrowserLogoutFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceBrowserLogoutFlowRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceLoginForBrowsers

> LoginFlow InitializeSelfServiceLoginForBrowsers(ctx).Refresh(refresh).Execute()

Initialize Login Flow for Browsers



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    refresh := true // bool | Refresh a login session  If set to true, this will refresh an existing login session by asking the user to sign in again. This will reset the authenticated_at time of the session. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceLoginForBrowsers(context.Background()).Refresh(refresh).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceLoginForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceLoginForBrowsers`: LoginFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceLoginForBrowsers`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceLoginForBrowsersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **refresh** | **bool** | Refresh a login session  If set to true, this will refresh an existing login session by asking the user to sign in again. This will reset the authenticated_at time of the session. | 

### Return type

[**LoginFlow**](LoginFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceLoginWithoutBrowser

> LoginFlow InitializeSelfServiceLoginWithoutBrowser(ctx).Refresh(refresh).Execute()

Initialize Login Flow for APIs, Services, Apps, ...



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    refresh := true // bool | Refresh a login session  If set to true, this will refresh an existing login session by asking the user to sign in again. This will reset the authenticated_at time of the session. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceLoginWithoutBrowser(context.Background()).Refresh(refresh).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceLoginWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceLoginWithoutBrowser`: LoginFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceLoginWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceLoginWithoutBrowserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **refresh** | **bool** | Refresh a login session  If set to true, this will refresh an existing login session by asking the user to sign in again. This will reset the authenticated_at time of the session. | 

### Return type

[**LoginFlow**](LoginFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceRecoveryForBrowsers

> RecoveryFlow InitializeSelfServiceRecoveryForBrowsers(ctx).Execute()

Initialize Recovery Flow for Browsers



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceRecoveryForBrowsers(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceRecoveryForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceRecoveryForBrowsers`: RecoveryFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceRecoveryForBrowsers`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceRecoveryForBrowsersRequest struct via the builder pattern


### Return type

[**RecoveryFlow**](RecoveryFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceRecoveryWithoutBrowser

> RecoveryFlow InitializeSelfServiceRecoveryWithoutBrowser(ctx).Execute()

Initialize Recovery Flow for APIs, Services, Apps, ...



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceRecoveryWithoutBrowser(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceRecoveryWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceRecoveryWithoutBrowser`: RecoveryFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceRecoveryWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceRecoveryWithoutBrowserRequest struct via the builder pattern


### Return type

[**RecoveryFlow**](RecoveryFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceRegistrationForBrowsers

> RegistrationFlow InitializeSelfServiceRegistrationForBrowsers(ctx).Execute()

Initialize Registration Flow for Browsers



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceRegistrationForBrowsers(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceRegistrationForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceRegistrationForBrowsers`: RegistrationFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceRegistrationForBrowsers`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceRegistrationForBrowsersRequest struct via the builder pattern


### Return type

[**RegistrationFlow**](RegistrationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceRegistrationWithoutBrowser

> RegistrationFlow InitializeSelfServiceRegistrationWithoutBrowser(ctx).Execute()

Initialize Registration Flow for APIs, Services, Apps, ...



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceRegistrationWithoutBrowser(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceRegistrationWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceRegistrationWithoutBrowser`: RegistrationFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceRegistrationWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceRegistrationWithoutBrowserRequest struct via the builder pattern


### Return type

[**RegistrationFlow**](RegistrationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceSettingsForBrowsers

> SettingsFlow InitializeSelfServiceSettingsForBrowsers(ctx).Execute()

Initialize Settings Flow for Browsers



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceSettingsForBrowsers(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceSettingsForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceSettingsForBrowsers`: SettingsFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceSettingsForBrowsers`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceSettingsForBrowsersRequest struct via the builder pattern


### Return type

[**SettingsFlow**](SettingsFlow.md)

### Authorization

[sessionToken](../README.md#sessionToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceSettingsWithoutBrowser

> SettingsFlow InitializeSelfServiceSettingsWithoutBrowser(ctx).XSessionToken(xSessionToken).XSessionCookie(xSessionCookie).Execute()

Initialize Settings Flow for APIs, Services, Apps, ...



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    xSessionToken := "xSessionToken_example" // string | The Session Token of the Identity performing the settings flow. (optional)
    xSessionCookie := "xSessionCookie_example" // string | The Session Cookie of the Identity performing the settings flow. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceSettingsWithoutBrowser(context.Background()).XSessionToken(xSessionToken).XSessionCookie(xSessionCookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceSettingsWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceSettingsWithoutBrowser`: SettingsFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceSettingsWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceSettingsWithoutBrowserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xSessionToken** | **string** | The Session Token of the Identity performing the settings flow. | 
 **xSessionCookie** | **string** | The Session Cookie of the Identity performing the settings flow. | 

### Return type

[**SettingsFlow**](SettingsFlow.md)

### Authorization

[sessionToken](../README.md#sessionToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceVerificationForBrowsers

> VerificationFlow InitializeSelfServiceVerificationForBrowsers(ctx).Execute()

Initialize Verification Flow for Browser Clients



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceVerificationForBrowsers(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceVerificationForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceVerificationForBrowsers`: VerificationFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceVerificationForBrowsers`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceVerificationForBrowsersRequest struct via the builder pattern


### Return type

[**VerificationFlow**](VerificationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceVerificationWithoutBrowser

> VerificationFlow InitializeSelfServiceVerificationWithoutBrowser(ctx).Execute()

Initialize Verification Flow for APIs, Services, Apps, ...



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.InitializeSelfServiceVerificationWithoutBrowser(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.InitializeSelfServiceVerificationWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceVerificationWithoutBrowser`: VerificationFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.InitializeSelfServiceVerificationWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceVerificationWithoutBrowserRequest struct via the builder pattern


### Return type

[**VerificationFlow**](VerificationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RevokeSession

> RevokeSession(ctx).RevokeSession(revokeSession).Execute()

Initialize Logout Flow for API Clients - Revoke a Session



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    revokeSession := *openapiclient.NewRevokeSession("SessionToken_example") // RevokeSession | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.RevokeSession(context.Background()).RevokeSession(revokeSession).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.RevokeSession``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRevokeSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **revokeSession** | [**RevokeSession**](RevokeSession.md) |  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceLoginFlow

> LoginViaApiResponse SubmitSelfServiceLoginFlow(ctx).Flow(flow).SubmitSelfServiceLoginFlow(submitSelfServiceLoginFlow).Execute()

Submit a Login Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    flow := "flow_example" // string | The Login Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/login?flow=abcde`).
    submitSelfServiceLoginFlow := openapiclient.submitSelfServiceLoginFlow{SubmitSelfServiceLoginFlowWithPasswordMethod: openapiclient.NewSubmitSelfServiceLoginFlowWithPasswordMethod()} // SubmitSelfServiceLoginFlow |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.SubmitSelfServiceLoginFlow(context.Background()).Flow(flow).SubmitSelfServiceLoginFlow(submitSelfServiceLoginFlow).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.SubmitSelfServiceLoginFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceLoginFlow`: LoginViaApiResponse
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.SubmitSelfServiceLoginFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceLoginFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Login Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/login?flow&#x3D;abcde&#x60;). | 
 **submitSelfServiceLoginFlow** | [**SubmitSelfServiceLoginFlow**](SubmitSelfServiceLoginFlow.md) |  | 

### Return type

[**LoginViaApiResponse**](LoginViaApiResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceRecoveryFlow

> RecoveryFlow SubmitSelfServiceRecoveryFlow(ctx).Flow(flow).Body(body).Execute()

Complete Recovery Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    flow := "flow_example" // string | The Registration Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/registration?flow=abcde`).
    body := map[string]interface{}(Object) // map[string]interface{} |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.SubmitSelfServiceRecoveryFlow(context.Background()).Flow(flow).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.SubmitSelfServiceRecoveryFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceRecoveryFlow`: RecoveryFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.SubmitSelfServiceRecoveryFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceRecoveryFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Registration Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/registration?flow&#x3D;abcde&#x60;). | 
 **body** | **map[string]interface{}** |  | 

### Return type

[**RecoveryFlow**](RecoveryFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceRecoveryFlowWithLinkMethod

> SubmitSelfServiceRecoveryFlowWithLinkMethod(ctx).Token(token).Flow(flow).SubmitSelfServiceRecoveryFlowWithLinkMethod(submitSelfServiceRecoveryFlowWithLinkMethod).Execute()

Complete Recovery Flow with Link Method



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    token := "token_example" // string | Recovery Token  The recovery token which completes the recovery request. If the token is invalid (e.g. expired) an error will be shown to the end-user. (optional)
    flow := "flow_example" // string | The Flow ID  format: uuid (optional)
    submitSelfServiceRecoveryFlowWithLinkMethod := *openapiclient.NewSubmitSelfServiceRecoveryFlowWithLinkMethod() // SubmitSelfServiceRecoveryFlowWithLinkMethod |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.SubmitSelfServiceRecoveryFlowWithLinkMethod(context.Background()).Token(token).Flow(flow).SubmitSelfServiceRecoveryFlowWithLinkMethod(submitSelfServiceRecoveryFlowWithLinkMethod).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.SubmitSelfServiceRecoveryFlowWithLinkMethod``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceRecoveryFlowWithLinkMethodRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **token** | **string** | Recovery Token  The recovery token which completes the recovery request. If the token is invalid (e.g. expired) an error will be shown to the end-user. | 
 **flow** | **string** | The Flow ID  format: uuid | 
 **submitSelfServiceRecoveryFlowWithLinkMethod** | [**SubmitSelfServiceRecoveryFlowWithLinkMethod**](SubmitSelfServiceRecoveryFlowWithLinkMethod.md) |  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceRegistrationFlow

> RegistrationViaApiResponse SubmitSelfServiceRegistrationFlow(ctx).Flow(flow).SubmitSelfServiceRegistrationFlow(submitSelfServiceRegistrationFlow).Execute()

Submit a Registration Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    flow := "flow_example" // string | The Registration Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/registration?flow=abcde`).
    submitSelfServiceRegistrationFlow := openapiclient.submitSelfServiceRegistrationFlow{SubmitSelfServiceRegistrationFlowWithPasswordMethod: openapiclient.NewSubmitSelfServiceRegistrationFlowWithPasswordMethod("Method_example")} // SubmitSelfServiceRegistrationFlow |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.SubmitSelfServiceRegistrationFlow(context.Background()).Flow(flow).SubmitSelfServiceRegistrationFlow(submitSelfServiceRegistrationFlow).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.SubmitSelfServiceRegistrationFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceRegistrationFlow`: RegistrationViaApiResponse
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.SubmitSelfServiceRegistrationFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceRegistrationFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Registration Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/registration?flow&#x3D;abcde&#x60;). | 
 **submitSelfServiceRegistrationFlow** | [**SubmitSelfServiceRegistrationFlow**](SubmitSelfServiceRegistrationFlow.md) |  | 

### Return type

[**RegistrationViaApiResponse**](RegistrationViaApiResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceSettingsFlow

> SettingsViaApiResponse SubmitSelfServiceSettingsFlow(ctx).Flow(flow).XSessionToken(xSessionToken).SubmitSelfServiceSettingsFlow(submitSelfServiceSettingsFlow).Execute()

Complete Settings Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    flow := "flow_example" // string | The Settings Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/settings?flow=abcde`).
    xSessionToken := "xSessionToken_example" // string | The Session Token of the Identity performing the settings flow. (optional)
    submitSelfServiceSettingsFlow := openapiclient.submitSelfServiceSettingsFlow{SubmitSelfServiceSettingsFlowWithPasswordMethod: openapiclient.NewSubmitSelfServiceSettingsFlowWithPasswordMethod("Password_example")} // SubmitSelfServiceSettingsFlow |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.SubmitSelfServiceSettingsFlow(context.Background()).Flow(flow).XSessionToken(xSessionToken).SubmitSelfServiceSettingsFlow(submitSelfServiceSettingsFlow).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.SubmitSelfServiceSettingsFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceSettingsFlow`: SettingsViaApiResponse
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.SubmitSelfServiceSettingsFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceSettingsFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Settings Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/settings?flow&#x3D;abcde&#x60;). | 
 **xSessionToken** | **string** | The Session Token of the Identity performing the settings flow. | 
 **submitSelfServiceSettingsFlow** | [**SubmitSelfServiceSettingsFlow**](SubmitSelfServiceSettingsFlow.md) |  | 

### Return type

[**SettingsViaApiResponse**](SettingsViaApiResponse.md)

### Authorization

[sessionToken](../README.md#sessionToken)

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceVerificationFlow

> VerificationFlow SubmitSelfServiceVerificationFlow(ctx).Flow(flow).Body(body).Execute()

Complete Verification Flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    flow := "flow_example" // string | The Registration Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/registration?flow=abcde`).
    body := map[string]interface{}(Object) // map[string]interface{} |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.SubmitSelfServiceVerificationFlow(context.Background()).Flow(flow).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.SubmitSelfServiceVerificationFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceVerificationFlow`: VerificationFlow
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.SubmitSelfServiceVerificationFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceVerificationFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Registration Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/registration?flow&#x3D;abcde&#x60;). | 
 **body** | **map[string]interface{}** |  | 

### Return type

[**VerificationFlow**](VerificationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ToSession

> Session ToSession(ctx).XSessionToken(xSessionToken).XSessionCookie(xSessionCookie).Execute()

Check Who the Current HTTP Session Belongs To



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    xSessionToken := "xSessionToken_example" // string |  (optional)
    xSessionCookie := "xSessionCookie_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PublicApi.ToSession(context.Background()).XSessionToken(xSessionToken).XSessionCookie(xSessionCookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PublicApi.ToSession``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ToSession`: Session
    fmt.Fprintf(os.Stdout, "Response from `PublicApi.ToSession`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiToSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xSessionToken** | **string** |  | 
 **xSessionCookie** | **string** |  | 

### Return type

[**Session**](Session.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

