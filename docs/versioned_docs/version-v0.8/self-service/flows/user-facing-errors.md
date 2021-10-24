---
id: user-facing-errors
title: User-Facing Errors
---

Because Ory Kratos does not render User Interfaces, we implemented a flow that
allows you to implement the error page in any way you want.

## User-Facing Errors in the Browser

When a user-facing error occurs (e.g. during Self Service User Login), Ory
Kratos will store error message and context and redirect the User's Browser to
the Error UI URL set by the `selfservice.flows.error.ui_url` configuration or
`SELFSERVICE_FLOWS_ERROR_UI_URL` environment variable.

Assuming `selfservice.flows.error.ui_url` is set to
`https://example.org/errors`, Ory Kratos will redirect the User's Browser to
`https://example.org/errors?id=abcde`.

The route matching `https://example.org/errors` uses the `error` URL Query
parameter value `abcde` to make a request to Ory Kratos' Public or Admin API
`https://kratos-<public|admin>/self-service/errors?id=abcde`. The JSON Response
contains a list of errors and their details, for example:

```json
[
  {
    "code": 500,
    "message": "no such file or directory"
  }
]
```

We are working on documenting possible error messages and homogenize error
layouts. In general, errors have the following keys defined:

```json
{
  "code": 500,
  "message": "some message",
  "reason": "some reason",
  "debug": "some debug info"
}
```

## User-Facing Errors when consuming APIs

When a user-facing error occurs and the HTTP client is an API Client (e.g.
Mobile App), the error will be returned as the HTTP Response. No additional
steps are required.

## Using Stub Errors

The error endpoint supports stub errors which can be used to implement your
Error UI:

- `?id=stub:500` - returns a stub 500 (Internal Server Error) error.

To call a stub error, simply do:

```shell script
$ curl -s \
  'https://playground.projects.oryapis.com/api/kratos/public/self-service/errors?id=stub:500' | jq

{
  "id": "9f900efa-a5ea-4dfd-8311-a8c7448ffeec",
  "error": {
    "code": 500,
    "status": "Internal Server Error",
    "reason": "This is a stub error.",
    "message": "An internal server error occurred, please contact the system administrator"
  },
  "created_at": "0001-01-01T00:00:00Z",
  "updated_at": "0001-01-01T00:00:00Z"
}
```
