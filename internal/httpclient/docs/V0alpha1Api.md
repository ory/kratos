# \V0alpha1Api

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AdminCreateIdentity**](V0alpha1Api.md#AdminCreateIdentity) | **Post** /identities | Create an Identity
[**AdminCreateSelfServiceRecoveryLink**](V0alpha1Api.md#AdminCreateSelfServiceRecoveryLink) | **Post** /recovery/link | Create a Recovery Link
[**AdminDeleteIdentity**](V0alpha1Api.md#AdminDeleteIdentity) | **Delete** /identities/{id} | Delete an Identity
[**AdminGetIdentity**](V0alpha1Api.md#AdminGetIdentity) | **Get** /identities/{id} | Get an Identity
[**AdminListIdentities**](V0alpha1Api.md#AdminListIdentities) | **Get** /identities | List Identities
[**AdminUpdateIdentity**](V0alpha1Api.md#AdminUpdateIdentity) | **Put** /identities/{id} | Update an Identity
[**CreateSelfServiceLogoutFlowUrlForBrowsers**](V0alpha1Api.md#CreateSelfServiceLogoutFlowUrlForBrowsers) | **Get** /self-service/logout/browser | Create a Logout URL for Browsers
[**GetJsonSchema**](V0alpha1Api.md#GetJsonSchema) | **Get** /schemas/{id} | 
[**GetSelfServiceError**](V0alpha1Api.md#GetSelfServiceError) | **Get** /self-service/errors | Get Self-Service Errors
[**GetSelfServiceLoginFlow**](V0alpha1Api.md#GetSelfServiceLoginFlow) | **Get** /self-service/login/flows | Get Login Flow
[**GetSelfServiceRecoveryFlow**](V0alpha1Api.md#GetSelfServiceRecoveryFlow) | **Get** /self-service/recovery/flows | Get Recovery Flow
[**GetSelfServiceRegistrationFlow**](V0alpha1Api.md#GetSelfServiceRegistrationFlow) | **Get** /self-service/registration/flows | Get Registration Flow
[**GetSelfServiceSettingsFlow**](V0alpha1Api.md#GetSelfServiceSettingsFlow) | **Get** /self-service/settings/flows | Get Settings Flow
[**GetSelfServiceVerificationFlow**](V0alpha1Api.md#GetSelfServiceVerificationFlow) | **Get** /self-service/verification/flows | Get Verification Flow
[**InitializeSelfServiceLoginFlowForBrowsers**](V0alpha1Api.md#InitializeSelfServiceLoginFlowForBrowsers) | **Get** /self-service/login/browser | Initialize Login Flow for Browsers
[**InitializeSelfServiceLoginFlowWithoutBrowser**](V0alpha1Api.md#InitializeSelfServiceLoginFlowWithoutBrowser) | **Get** /self-service/login/api | Initialize Login Flow for APIs, Services, Apps, ...
[**InitializeSelfServiceRecoveryFlowForBrowsers**](V0alpha1Api.md#InitializeSelfServiceRecoveryFlowForBrowsers) | **Get** /self-service/recovery/browser | Initialize Recovery Flow for Browsers
[**InitializeSelfServiceRecoveryFlowWithoutBrowser**](V0alpha1Api.md#InitializeSelfServiceRecoveryFlowWithoutBrowser) | **Get** /self-service/recovery/api | Initialize Recovery Flow for APIs, Services, Apps, ...
[**InitializeSelfServiceRegistrationFlowForBrowsers**](V0alpha1Api.md#InitializeSelfServiceRegistrationFlowForBrowsers) | **Get** /self-service/registration/browser | Initialize Registration Flow for Browsers
[**InitializeSelfServiceRegistrationFlowWithoutBrowser**](V0alpha1Api.md#InitializeSelfServiceRegistrationFlowWithoutBrowser) | **Get** /self-service/registration/api | Initialize Registration Flow for APIs, Services, Apps, ...
[**InitializeSelfServiceSettingsFlowForBrowsers**](V0alpha1Api.md#InitializeSelfServiceSettingsFlowForBrowsers) | **Get** /self-service/settings/browser | Initialize Settings Flow for Browsers
[**InitializeSelfServiceSettingsFlowWithoutBrowser**](V0alpha1Api.md#InitializeSelfServiceSettingsFlowWithoutBrowser) | **Get** /self-service/settings/api | Initialize Settings Flow for APIs, Services, Apps, ...
[**InitializeSelfServiceVerificationFlowForBrowsers**](V0alpha1Api.md#InitializeSelfServiceVerificationFlowForBrowsers) | **Get** /self-service/verification/browser | Initialize Verification Flow for Browser Clients
[**InitializeSelfServiceVerificationFlowWithoutBrowser**](V0alpha1Api.md#InitializeSelfServiceVerificationFlowWithoutBrowser) | **Get** /self-service/verification/api | Initialize Verification Flow for APIs, Services, Apps, ...
[**SubmitSelfServiceLoginFlow**](V0alpha1Api.md#SubmitSelfServiceLoginFlow) | **Post** /self-service/login | Submit a Login Flow
[**SubmitSelfServiceLogoutFlow**](V0alpha1Api.md#SubmitSelfServiceLogoutFlow) | **Get** /self-service/logout | Complete Self-Service Logout
[**SubmitSelfServiceLogoutFlowWithoutBrowser**](V0alpha1Api.md#SubmitSelfServiceLogoutFlowWithoutBrowser) | **Delete** /self-service/logout/api | Perform Logout for APIs, Services, Apps, ...
[**SubmitSelfServiceRecoveryFlow**](V0alpha1Api.md#SubmitSelfServiceRecoveryFlow) | **Post** /self-service/recovery | Complete Recovery Flow
[**SubmitSelfServiceRegistrationFlow**](V0alpha1Api.md#SubmitSelfServiceRegistrationFlow) | **Post** /self-service/registration | Submit a Registration Flow
[**SubmitSelfServiceSettingsFlow**](V0alpha1Api.md#SubmitSelfServiceSettingsFlow) | **Post** /self-service/settings | Complete Settings Flow
[**SubmitSelfServiceVerificationFlow**](V0alpha1Api.md#SubmitSelfServiceVerificationFlow) | **Post** /self-service/verification | Complete Verification Flow
[**ToSession**](V0alpha1Api.md#ToSession) | **Get** /sessions/whoami | Check Who the Current HTTP Session Belongs To



## AdminCreateIdentity

> Identity AdminCreateIdentity(ctx).AdminCreateIdentityBody(adminCreateIdentityBody).Execute()

Create an Identity



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
    adminCreateIdentityBody := *openapiclient.NewAdminCreateIdentityBody("SchemaId_example", map[string]interface{}(123)) // AdminCreateIdentityBody |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.AdminCreateIdentity(context.Background()).AdminCreateIdentityBody(adminCreateIdentityBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.AdminCreateIdentity``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdminCreateIdentity`: Identity
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.AdminCreateIdentity`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAdminCreateIdentityRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **adminCreateIdentityBody** | [**AdminCreateIdentityBody**](AdminCreateIdentityBody.md) |  | 

### Return type

[**Identity**](Identity.md)

### Authorization

[oryAccessToken](../README.md#oryAccessToken)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdminCreateSelfServiceRecoveryLink

> SelfServiceRecoveryLink AdminCreateSelfServiceRecoveryLink(ctx).AdminCreateSelfServiceRecoveryLinkBody(adminCreateSelfServiceRecoveryLinkBody).Execute()

Create a Recovery Link



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
    adminCreateSelfServiceRecoveryLinkBody := *openapiclient.NewAdminCreateSelfServiceRecoveryLinkBody("IdentityId_example") // AdminCreateSelfServiceRecoveryLinkBody |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.AdminCreateSelfServiceRecoveryLink(context.Background()).AdminCreateSelfServiceRecoveryLinkBody(adminCreateSelfServiceRecoveryLinkBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.AdminCreateSelfServiceRecoveryLink``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdminCreateSelfServiceRecoveryLink`: SelfServiceRecoveryLink
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.AdminCreateSelfServiceRecoveryLink`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAdminCreateSelfServiceRecoveryLinkRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **adminCreateSelfServiceRecoveryLinkBody** | [**AdminCreateSelfServiceRecoveryLinkBody**](AdminCreateSelfServiceRecoveryLinkBody.md) |  | 

### Return type

[**SelfServiceRecoveryLink**](SelfServiceRecoveryLink.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdminDeleteIdentity

> AdminDeleteIdentity(ctx, id).Execute()

Delete an Identity



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
    id := "id_example" // string | ID is the identity's ID.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.AdminDeleteIdentity(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.AdminDeleteIdentity``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID is the identity&#39;s ID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdminDeleteIdentityRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[oryAccessToken](../README.md#oryAccessToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdminGetIdentity

> Identity AdminGetIdentity(ctx, id).Execute()

Get an Identity



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
    id := "id_example" // string | ID must be set to the ID of identity you want to get

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.AdminGetIdentity(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.AdminGetIdentity``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdminGetIdentity`: Identity
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.AdminGetIdentity`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID must be set to the ID of identity you want to get | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdminGetIdentityRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Identity**](Identity.md)

### Authorization

[oryAccessToken](../README.md#oryAccessToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdminListIdentities

> []Identity AdminListIdentities(ctx).PerPage(perPage).Page(page).Execute()

List Identities



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
    perPage := int64(789) // int64 | Items per Page  This is the number of items per page. (optional) (default to 100)
    page := int64(789) // int64 | Pagination Page (optional) (default to 0)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.AdminListIdentities(context.Background()).PerPage(perPage).Page(page).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.AdminListIdentities``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdminListIdentities`: []Identity
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.AdminListIdentities`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAdminListIdentitiesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **perPage** | **int64** | Items per Page  This is the number of items per page. | [default to 100]
 **page** | **int64** | Pagination Page | [default to 0]

### Return type

[**[]Identity**](Identity.md)

### Authorization

[oryAccessToken](../README.md#oryAccessToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AdminUpdateIdentity

> Identity AdminUpdateIdentity(ctx, id).AdminUpdateIdentityBody(adminUpdateIdentityBody).Execute()

Update an Identity



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
    id := "id_example" // string | ID must be set to the ID of identity you want to update
    adminUpdateIdentityBody := *openapiclient.NewAdminUpdateIdentityBody(interface{}(123), map[string]interface{}(123)) // AdminUpdateIdentityBody |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.AdminUpdateIdentity(context.Background(), id).AdminUpdateIdentityBody(adminUpdateIdentityBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.AdminUpdateIdentity``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AdminUpdateIdentity`: Identity
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.AdminUpdateIdentity`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID must be set to the ID of identity you want to update | 

### Other Parameters

Other parameters are passed through a pointer to a apiAdminUpdateIdentityRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **adminUpdateIdentityBody** | [**AdminUpdateIdentityBody**](AdminUpdateIdentityBody.md) |  | 

### Return type

[**Identity**](Identity.md)

### Authorization

[oryAccessToken](../README.md#oryAccessToken)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateSelfServiceLogoutFlowUrlForBrowsers

> SelfServiceLogoutUrl CreateSelfServiceLogoutFlowUrlForBrowsers(ctx).Cookie(cookie).Execute()

Create a Logout URL for Browsers



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
    cookie := "cookie_example" // string | HTTP Cookies  If you call this endpoint from a backend, please include the original Cookie header in the request. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.CreateSelfServiceLogoutFlowUrlForBrowsers(context.Background()).Cookie(cookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.CreateSelfServiceLogoutFlowUrlForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateSelfServiceLogoutFlowUrlForBrowsers`: SelfServiceLogoutUrl
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.CreateSelfServiceLogoutFlowUrlForBrowsers`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateSelfServiceLogoutFlowUrlForBrowsersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **cookie** | **string** | HTTP Cookies  If you call this endpoint from a backend, please include the original Cookie header in the request. | 

### Return type

[**SelfServiceLogoutUrl**](SelfServiceLogoutUrl.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetJsonSchema

> map[string]interface{} GetJsonSchema(ctx, id).Execute()





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
    resp, r, err := apiClient.V0alpha1Api.GetJsonSchema(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.GetJsonSchema``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetJsonSchema`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.GetJsonSchema`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID must be set to the ID of schema you want to get | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetJsonSchemaRequest struct via the builder pattern


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

> SelfServiceError GetSelfServiceError(ctx).Id(id).Execute()

Get Self-Service Errors



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
    id := "id_example" // string | Error is the error's ID

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.GetSelfServiceError(context.Background()).Id(id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.GetSelfServiceError``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceError`: SelfServiceError
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.GetSelfServiceError`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceErrorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Error is the error&#39;s ID | 

### Return type

[**SelfServiceError**](SelfServiceError.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceLoginFlow

> SelfServiceLoginFlow GetSelfServiceLoginFlow(ctx).Id(id).Cookie(cookie).Execute()

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
    cookie := "cookie_example" // string | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.GetSelfServiceLoginFlow(context.Background()).Id(id).Cookie(cookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.GetSelfServiceLoginFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceLoginFlow`: SelfServiceLoginFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.GetSelfServiceLoginFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceLoginFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | The Login Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/login?flow&#x3D;abcde&#x60;). | 
 **cookie** | **string** | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. | 

### Return type

[**SelfServiceLoginFlow**](SelfServiceLoginFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceRecoveryFlow

> SelfServiceRecoveryFlow GetSelfServiceRecoveryFlow(ctx).Id(id).Cookie(cookie).Execute()

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
    cookie := "cookie_example" // string | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.GetSelfServiceRecoveryFlow(context.Background()).Id(id).Cookie(cookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.GetSelfServiceRecoveryFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceRecoveryFlow`: SelfServiceRecoveryFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.GetSelfServiceRecoveryFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceRecoveryFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | The Flow ID  The value for this parameter comes from &#x60;request&#x60; URL Query parameter sent to your application (e.g. &#x60;/recovery?flow&#x3D;abcde&#x60;). | 
 **cookie** | **string** | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. | 

### Return type

[**SelfServiceRecoveryFlow**](SelfServiceRecoveryFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceRegistrationFlow

> SelfServiceRegistrationFlow GetSelfServiceRegistrationFlow(ctx).Id(id).Cookie(cookie).Execute()

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
    cookie := "cookie_example" // string | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.GetSelfServiceRegistrationFlow(context.Background()).Id(id).Cookie(cookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.GetSelfServiceRegistrationFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceRegistrationFlow`: SelfServiceRegistrationFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.GetSelfServiceRegistrationFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceRegistrationFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | The Registration Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/registration?flow&#x3D;abcde&#x60;). | 
 **cookie** | **string** | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. | 

### Return type

[**SelfServiceRegistrationFlow**](SelfServiceRegistrationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceSettingsFlow

> SelfServiceSettingsFlow GetSelfServiceSettingsFlow(ctx).Id(id).XSessionToken(xSessionToken).Cookie(cookie).Execute()

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
    xSessionToken := "xSessionToken_example" // string | The Session Token  When using the SDK in an app without a browser, please include the session token here. (optional)
    cookie := "cookie_example" // string | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. You only need to do this for browser- based flows. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.GetSelfServiceSettingsFlow(context.Background()).Id(id).XSessionToken(xSessionToken).Cookie(cookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.GetSelfServiceSettingsFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceSettingsFlow`: SelfServiceSettingsFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.GetSelfServiceSettingsFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceSettingsFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | ID is the Settings Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/settings?flow&#x3D;abcde&#x60;). | 
 **xSessionToken** | **string** | The Session Token  When using the SDK in an app without a browser, please include the session token here. | 
 **cookie** | **string** | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. You only need to do this for browser- based flows. | 

### Return type

[**SelfServiceSettingsFlow**](SelfServiceSettingsFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSelfServiceVerificationFlow

> SelfServiceVerificationFlow GetSelfServiceVerificationFlow(ctx).Id(id).Cookie(cookie).Execute()

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
    cookie := "cookie_example" // string | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.GetSelfServiceVerificationFlow(context.Background()).Id(id).Cookie(cookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.GetSelfServiceVerificationFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSelfServiceVerificationFlow`: SelfServiceVerificationFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.GetSelfServiceVerificationFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSelfServiceVerificationFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | The Flow ID  The value for this parameter comes from &#x60;request&#x60; URL Query parameter sent to your application (e.g. &#x60;/verification?flow&#x3D;abcde&#x60;). | 
 **cookie** | **string** | HTTP Cookies  When using the SDK on the server side you must include the HTTP Cookie Header originally sent to your HTTP handler here. | 

### Return type

[**SelfServiceVerificationFlow**](SelfServiceVerificationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceLoginFlowForBrowsers

> SelfServiceLoginFlow InitializeSelfServiceLoginFlowForBrowsers(ctx).Refresh(refresh).Execute()

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
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceLoginFlowForBrowsers(context.Background()).Refresh(refresh).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceLoginFlowForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceLoginFlowForBrowsers`: SelfServiceLoginFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceLoginFlowForBrowsers`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceLoginFlowForBrowsersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **refresh** | **bool** | Refresh a login session  If set to true, this will refresh an existing login session by asking the user to sign in again. This will reset the authenticated_at time of the session. | 

### Return type

[**SelfServiceLoginFlow**](SelfServiceLoginFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceLoginFlowWithoutBrowser

> SelfServiceLoginFlow InitializeSelfServiceLoginFlowWithoutBrowser(ctx).Refresh(refresh).Execute()

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
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceLoginFlowWithoutBrowser(context.Background()).Refresh(refresh).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceLoginFlowWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceLoginFlowWithoutBrowser`: SelfServiceLoginFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceLoginFlowWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceLoginFlowWithoutBrowserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **refresh** | **bool** | Refresh a login session  If set to true, this will refresh an existing login session by asking the user to sign in again. This will reset the authenticated_at time of the session. | 

### Return type

[**SelfServiceLoginFlow**](SelfServiceLoginFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceRecoveryFlowForBrowsers

> SelfServiceRecoveryFlow InitializeSelfServiceRecoveryFlowForBrowsers(ctx).Execute()

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
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceRecoveryFlowForBrowsers(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceRecoveryFlowForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceRecoveryFlowForBrowsers`: SelfServiceRecoveryFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceRecoveryFlowForBrowsers`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceRecoveryFlowForBrowsersRequest struct via the builder pattern


### Return type

[**SelfServiceRecoveryFlow**](SelfServiceRecoveryFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceRecoveryFlowWithoutBrowser

> SelfServiceRecoveryFlow InitializeSelfServiceRecoveryFlowWithoutBrowser(ctx).Execute()

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
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceRecoveryFlowWithoutBrowser(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceRecoveryFlowWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceRecoveryFlowWithoutBrowser`: SelfServiceRecoveryFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceRecoveryFlowWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceRecoveryFlowWithoutBrowserRequest struct via the builder pattern


### Return type

[**SelfServiceRecoveryFlow**](SelfServiceRecoveryFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceRegistrationFlowForBrowsers

> SelfServiceRegistrationFlow InitializeSelfServiceRegistrationFlowForBrowsers(ctx).Execute()

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
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceRegistrationFlowForBrowsers(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceRegistrationFlowForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceRegistrationFlowForBrowsers`: SelfServiceRegistrationFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceRegistrationFlowForBrowsers`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceRegistrationFlowForBrowsersRequest struct via the builder pattern


### Return type

[**SelfServiceRegistrationFlow**](SelfServiceRegistrationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceRegistrationFlowWithoutBrowser

> SelfServiceRegistrationFlow InitializeSelfServiceRegistrationFlowWithoutBrowser(ctx).Execute()

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
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceRegistrationFlowWithoutBrowser(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceRegistrationFlowWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceRegistrationFlowWithoutBrowser`: SelfServiceRegistrationFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceRegistrationFlowWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceRegistrationFlowWithoutBrowserRequest struct via the builder pattern


### Return type

[**SelfServiceRegistrationFlow**](SelfServiceRegistrationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceSettingsFlowForBrowsers

> SelfServiceSettingsFlow InitializeSelfServiceSettingsFlowForBrowsers(ctx).Cookie(cookie).Execute()

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
    cookie := "cookie_example" // string | The Session Cookie of the Identity performing the settings flow. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceSettingsFlowForBrowsers(context.Background()).Cookie(cookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceSettingsFlowForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceSettingsFlowForBrowsers`: SelfServiceSettingsFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceSettingsFlowForBrowsers`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceSettingsFlowForBrowsersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **cookie** | **string** | The Session Cookie of the Identity performing the settings flow. | 

### Return type

[**SelfServiceSettingsFlow**](SelfServiceSettingsFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceSettingsFlowWithoutBrowser

> SelfServiceSettingsFlow InitializeSelfServiceSettingsFlowWithoutBrowser(ctx).XSessionToken(xSessionToken).Execute()

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceSettingsFlowWithoutBrowser(context.Background()).XSessionToken(xSessionToken).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceSettingsFlowWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceSettingsFlowWithoutBrowser`: SelfServiceSettingsFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceSettingsFlowWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceSettingsFlowWithoutBrowserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xSessionToken** | **string** | The Session Token of the Identity performing the settings flow. | 

### Return type

[**SelfServiceSettingsFlow**](SelfServiceSettingsFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceVerificationFlowForBrowsers

> SelfServiceVerificationFlow InitializeSelfServiceVerificationFlowForBrowsers(ctx).Execute()

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
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceVerificationFlowForBrowsers(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceVerificationFlowForBrowsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceVerificationFlowForBrowsers`: SelfServiceVerificationFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceVerificationFlowForBrowsers`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceVerificationFlowForBrowsersRequest struct via the builder pattern


### Return type

[**SelfServiceVerificationFlow**](SelfServiceVerificationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeSelfServiceVerificationFlowWithoutBrowser

> SelfServiceVerificationFlow InitializeSelfServiceVerificationFlowWithoutBrowser(ctx).Execute()

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
    resp, r, err := apiClient.V0alpha1Api.InitializeSelfServiceVerificationFlowWithoutBrowser(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.InitializeSelfServiceVerificationFlowWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InitializeSelfServiceVerificationFlowWithoutBrowser`: SelfServiceVerificationFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.InitializeSelfServiceVerificationFlowWithoutBrowser`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeSelfServiceVerificationFlowWithoutBrowserRequest struct via the builder pattern


### Return type

[**SelfServiceVerificationFlow**](SelfServiceVerificationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceLoginFlow

> SuccessfulSelfServiceLoginWithoutBrowser SubmitSelfServiceLoginFlow(ctx).Flow(flow).SubmitSelfServiceLoginFlowBody(submitSelfServiceLoginFlowBody).Execute()

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
    submitSelfServiceLoginFlowBody := openapiclient.submitSelfServiceLoginFlowBody{SubmitSelfServiceLoginFlowWithPasswordMethodBody: openapiclient.NewSubmitSelfServiceLoginFlowWithPasswordMethodBody("Method_example", "Password_example", "PasswordIdentifier_example")} // SubmitSelfServiceLoginFlowBody |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.SubmitSelfServiceLoginFlow(context.Background()).Flow(flow).SubmitSelfServiceLoginFlowBody(submitSelfServiceLoginFlowBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.SubmitSelfServiceLoginFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceLoginFlow`: SuccessfulSelfServiceLoginWithoutBrowser
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.SubmitSelfServiceLoginFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceLoginFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Login Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/login?flow&#x3D;abcde&#x60;). | 
 **submitSelfServiceLoginFlowBody** | [**SubmitSelfServiceLoginFlowBody**](SubmitSelfServiceLoginFlowBody.md) |  | 

### Return type

[**SuccessfulSelfServiceLoginWithoutBrowser**](SuccessfulSelfServiceLoginWithoutBrowser.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceLogoutFlow

> SubmitSelfServiceLogoutFlow(ctx).Token(token).Execute()

Complete Self-Service Logout



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
    token := "token_example" // string | A Valid Logout Token  If you do not have a logout token because you only have a session cookie, call `/self-service/logout/urls` to generate a URL for this endpoint. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.SubmitSelfServiceLogoutFlow(context.Background()).Token(token).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.SubmitSelfServiceLogoutFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceLogoutFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **token** | **string** | A Valid Logout Token  If you do not have a logout token because you only have a session cookie, call &#x60;/self-service/logout/urls&#x60; to generate a URL for this endpoint. | 

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


## SubmitSelfServiceLogoutFlowWithoutBrowser

> SubmitSelfServiceLogoutFlowWithoutBrowser(ctx).SubmitSelfServiceLogoutFlowWithoutBrowserBody(submitSelfServiceLogoutFlowWithoutBrowserBody).Execute()

Perform Logout for APIs, Services, Apps, ...



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
    submitSelfServiceLogoutFlowWithoutBrowserBody := *openapiclient.NewSubmitSelfServiceLogoutFlowWithoutBrowserBody("SessionToken_example") // SubmitSelfServiceLogoutFlowWithoutBrowserBody | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.SubmitSelfServiceLogoutFlowWithoutBrowser(context.Background()).SubmitSelfServiceLogoutFlowWithoutBrowserBody(submitSelfServiceLogoutFlowWithoutBrowserBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.SubmitSelfServiceLogoutFlowWithoutBrowser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceLogoutFlowWithoutBrowserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **submitSelfServiceLogoutFlowWithoutBrowserBody** | [**SubmitSelfServiceLogoutFlowWithoutBrowserBody**](SubmitSelfServiceLogoutFlowWithoutBrowserBody.md) |  | 

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


## SubmitSelfServiceRecoveryFlow

> SelfServiceRecoveryFlow SubmitSelfServiceRecoveryFlow(ctx).Flow(flow).Token(token).SubmitSelfServiceRecoveryFlowBody(submitSelfServiceRecoveryFlowBody).Execute()

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
    flow := "flow_example" // string | The Recovery Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/recovery?flow=abcde`).
    token := "token_example" // string | Recovery Token  The recovery token which completes the recovery request. If the token is invalid (e.g. expired) an error will be shown to the end-user.  This parameter is usually set in a link and not used by any direct API call. (optional)
    submitSelfServiceRecoveryFlowBody := openapiclient.submitSelfServiceRecoveryFlowBody{SubmitSelfServiceRecoveryFlowWithLinkMethodBody: openapiclient.NewSubmitSelfServiceRecoveryFlowWithLinkMethodBody("Email_example", "Method_example")} // SubmitSelfServiceRecoveryFlowBody |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.SubmitSelfServiceRecoveryFlow(context.Background()).Flow(flow).Token(token).SubmitSelfServiceRecoveryFlowBody(submitSelfServiceRecoveryFlowBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.SubmitSelfServiceRecoveryFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceRecoveryFlow`: SelfServiceRecoveryFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.SubmitSelfServiceRecoveryFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceRecoveryFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Recovery Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/recovery?flow&#x3D;abcde&#x60;). | 
 **token** | **string** | Recovery Token  The recovery token which completes the recovery request. If the token is invalid (e.g. expired) an error will be shown to the end-user.  This parameter is usually set in a link and not used by any direct API call. | 
 **submitSelfServiceRecoveryFlowBody** | [**SubmitSelfServiceRecoveryFlowBody**](SubmitSelfServiceRecoveryFlowBody.md) |  | 

### Return type

[**SelfServiceRecoveryFlow**](SelfServiceRecoveryFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceRegistrationFlow

> SuccessfulSelfServiceRegistrationWithoutBrowser SubmitSelfServiceRegistrationFlow(ctx).Flow(flow).SubmitSelfServiceRegistrationFlowBody(submitSelfServiceRegistrationFlowBody).Execute()

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
    submitSelfServiceRegistrationFlowBody := openapiclient.submitSelfServiceRegistrationFlowBody{SubmitSelfServiceRegistrationFlowWithPasswordMethodBody: openapiclient.NewSubmitSelfServiceRegistrationFlowWithPasswordMethodBody("Method_example", "Password_example", map[string]interface{}(123))} // SubmitSelfServiceRegistrationFlowBody |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.SubmitSelfServiceRegistrationFlow(context.Background()).Flow(flow).SubmitSelfServiceRegistrationFlowBody(submitSelfServiceRegistrationFlowBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.SubmitSelfServiceRegistrationFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceRegistrationFlow`: SuccessfulSelfServiceRegistrationWithoutBrowser
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.SubmitSelfServiceRegistrationFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceRegistrationFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Registration Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/registration?flow&#x3D;abcde&#x60;). | 
 **submitSelfServiceRegistrationFlowBody** | [**SubmitSelfServiceRegistrationFlowBody**](SubmitSelfServiceRegistrationFlowBody.md) |  | 

### Return type

[**SuccessfulSelfServiceRegistrationWithoutBrowser**](SuccessfulSelfServiceRegistrationWithoutBrowser.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceSettingsFlow

> SuccessfulSelfServiceSettingsWithoutBrowser SubmitSelfServiceSettingsFlow(ctx).Flow(flow).XSessionToken(xSessionToken).SubmitSelfServiceSettingsFlowBody(submitSelfServiceSettingsFlowBody).Execute()

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
    submitSelfServiceSettingsFlowBody := openapiclient.submitSelfServiceSettingsFlowBody{SubmitSelfServiceSettingsFlowWithPasswordMethodBody: openapiclient.NewSubmitSelfServiceSettingsFlowWithPasswordMethodBody("Method_example", "Password_example")} // SubmitSelfServiceSettingsFlowBody |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.SubmitSelfServiceSettingsFlow(context.Background()).Flow(flow).XSessionToken(xSessionToken).SubmitSelfServiceSettingsFlowBody(submitSelfServiceSettingsFlowBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.SubmitSelfServiceSettingsFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceSettingsFlow`: SuccessfulSelfServiceSettingsWithoutBrowser
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.SubmitSelfServiceSettingsFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceSettingsFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Settings Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/settings?flow&#x3D;abcde&#x60;). | 
 **xSessionToken** | **string** | The Session Token of the Identity performing the settings flow. | 
 **submitSelfServiceSettingsFlowBody** | [**SubmitSelfServiceSettingsFlowBody**](SubmitSelfServiceSettingsFlowBody.md) |  | 

### Return type

[**SuccessfulSelfServiceSettingsWithoutBrowser**](SuccessfulSelfServiceSettingsWithoutBrowser.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SubmitSelfServiceVerificationFlow

> SelfServiceVerificationFlow SubmitSelfServiceVerificationFlow(ctx).Flow(flow).Token(token).SubmitSelfServiceVerificationFlowBody(submitSelfServiceVerificationFlowBody).Execute()

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
    flow := "flow_example" // string | The Verification Flow ID  The value for this parameter comes from `flow` URL Query parameter sent to your application (e.g. `/verification?flow=abcde`).
    token := "token_example" // string | Verification Token  The verification token which completes the verification request. If the token is invalid (e.g. expired) an error will be shown to the end-user.  This parameter is usually set in a link and not used by any direct API call. (optional)
    submitSelfServiceVerificationFlowBody := openapiclient.submitSelfServiceVerificationFlowBody{SubmitSelfServiceVerificationFlowWithLinkMethodBody: openapiclient.NewSubmitSelfServiceVerificationFlowWithLinkMethodBody("Email_example", "Method_example")} // SubmitSelfServiceVerificationFlowBody |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.SubmitSelfServiceVerificationFlow(context.Background()).Flow(flow).Token(token).SubmitSelfServiceVerificationFlowBody(submitSelfServiceVerificationFlowBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.SubmitSelfServiceVerificationFlow``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SubmitSelfServiceVerificationFlow`: SelfServiceVerificationFlow
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.SubmitSelfServiceVerificationFlow`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSubmitSelfServiceVerificationFlowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flow** | **string** | The Verification Flow ID  The value for this parameter comes from &#x60;flow&#x60; URL Query parameter sent to your application (e.g. &#x60;/verification?flow&#x3D;abcde&#x60;). | 
 **token** | **string** | Verification Token  The verification token which completes the verification request. If the token is invalid (e.g. expired) an error will be shown to the end-user.  This parameter is usually set in a link and not used by any direct API call. | 
 **submitSelfServiceVerificationFlowBody** | [**SubmitSelfServiceVerificationFlowBody**](SubmitSelfServiceVerificationFlowBody.md) |  | 

### Return type

[**SelfServiceVerificationFlow**](SelfServiceVerificationFlow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json, application/x-www-form-urlencoded
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ToSession

> Session ToSession(ctx).XSessionToken(xSessionToken).Cookie(cookie).Execute()

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
    xSessionToken := "MP2YWEMeM8MxjkGKpH4dqOQ4Q4DlSPaj" // string | Set the Session Token when calling from non-browser clients. A session token has a format of `MP2YWEMeM8MxjkGKpH4dqOQ4Q4DlSPaj`. (optional)
    cookie := "ory_kratos_session=a19iOVAbdzdgl70Rq1QZmrKmcjDtdsviCTZx7m9a9yHIUS8Wa9T7hvqyGTsLHi6Qifn2WUfpAKx9DWp0SJGleIn9vh2YF4A16id93kXFTgIgmwIOvbVAScyrx7yVl6bPZnCx27ec4WQDtaTewC1CpgudeDV2jQQnSaCP6ny3xa8qLH-QUgYqdQuoA_LF1phxgRCUfIrCLQOkolX5nv3ze_f==" // string | Set the Cookie Header. This is especially useful when calling this endpoint from a server-side application. In that scenario you must include the HTTP Cookie Header which originally was included in the request to your server. An example of a session in the HTTP Cookie Header is: `ory_kratos_session=a19iOVAbdzdgl70Rq1QZmrKmcjDtdsviCTZx7m9a9yHIUS8Wa9T7hvqyGTsLHi6Qifn2WUfpAKx9DWp0SJGleIn9vh2YF4A16id93kXFTgIgmwIOvbVAScyrx7yVl6bPZnCx27ec4WQDtaTewC1CpgudeDV2jQQnSaCP6ny3xa8qLH-QUgYqdQuoA_LF1phxgRCUfIrCLQOkolX5nv3ze_f==`.  It is ok if more than one cookie are included here as all other cookies will be ignored. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.V0alpha1Api.ToSession(context.Background()).XSessionToken(xSessionToken).Cookie(cookie).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `V0alpha1Api.ToSession``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ToSession`: Session
    fmt.Fprintf(os.Stdout, "Response from `V0alpha1Api.ToSession`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiToSessionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xSessionToken** | **string** | Set the Session Token when calling from non-browser clients. A session token has a format of &#x60;MP2YWEMeM8MxjkGKpH4dqOQ4Q4DlSPaj&#x60;. | 
 **cookie** | **string** | Set the Cookie Header. This is especially useful when calling this endpoint from a server-side application. In that scenario you must include the HTTP Cookie Header which originally was included in the request to your server. An example of a session in the HTTP Cookie Header is: &#x60;ory_kratos_session&#x3D;a19iOVAbdzdgl70Rq1QZmrKmcjDtdsviCTZx7m9a9yHIUS8Wa9T7hvqyGTsLHi6Qifn2WUfpAKx9DWp0SJGleIn9vh2YF4A16id93kXFTgIgmwIOvbVAScyrx7yVl6bPZnCx27ec4WQDtaTewC1CpgudeDV2jQQnSaCP6ny3xa8qLH-QUgYqdQuoA_LF1phxgRCUfIrCLQOkolX5nv3ze_f&#x3D;&#x3D;&#x60;.  It is ok if more than one cookie are included here as all other cookies will be ignored. | 

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

