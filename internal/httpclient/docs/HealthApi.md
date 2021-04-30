# \HealthApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**IsInstanceAlive**](HealthApi.md#IsInstanceAlive) | **Get** /health/alive | Check alive status
[**IsInstanceReady**](HealthApi.md#IsInstanceReady) | **Get** /health/ready | Check readiness status



## IsInstanceAlive

> HealthStatus IsInstanceAlive(ctx).Execute()

Check alive status



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
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.HealthApi.IsInstanceAlive(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `HealthApi.IsInstanceAlive``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `IsInstanceAlive`: HealthStatus
    fmt.Fprintf(os.Stdout, "Response from `HealthApi.IsInstanceAlive`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiIsInstanceAliveRequest struct via the builder pattern


### Return type

[**HealthStatus**](healthStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## IsInstanceReady

> HealthStatus IsInstanceReady(ctx).Execute()

Check readiness status



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
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.HealthApi.IsInstanceReady(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `HealthApi.IsInstanceReady``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `IsInstanceReady`: HealthStatus
    fmt.Fprintf(os.Stdout, "Response from `HealthApi.IsInstanceReady`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiIsInstanceReadyRequest struct via the builder pattern


### Return type

[**HealthStatus**](healthStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

