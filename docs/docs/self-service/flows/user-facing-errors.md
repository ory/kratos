---
id: user-facing-errors
title: User-Facing Errors
---

Because ORY Kratos does not render User Interfaces, we implemented a flow that
allows you to implement the error page in any way you want.

## User-Facing Errors in the Browser

When a user-facing error occurs (e.g. during Self Service User Login), ORY
Kratos will store error message and context and redirect the User's Browser to
the Error UI URL set by the `selfservice.flows.error.ui_url` configuration or
`SELFSERVICE_FLOWS_ERROR_UI_URL` environment variable.

Assuming `selfservice.flows.error.ui_url` is set to
`https://example.org/errors`, ORY Kratos will redirect the User's Browser to
`https://example.org/errors?error=abcde`.

The route matching `https://example.org/errors` uses the `error` URL Query
parameter value `abcde` to make a request to ORY Kratos' Public or Admin API
`https://kratos-<public|admin/self-service/errors?error=abcde`. The JSON
Response contains a list of errors and their details, for example:

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
