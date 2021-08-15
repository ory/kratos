---
id: reveal-credentials
title: Reveal Credentials
---

This document describe how to reveal credentials from `/identities` URI.


## Retrieve OIDC Access Token and Refresh Token

The access token and refresh token from oidc
provider. This assumes that you have oidc method configured.

```
/identities/{id}/reveal_credentials
```

```shell script
$ curl --request GET -sL \
    --header "Content-Type: application/json" \
    http://127.0.0.1:4434/identities/3ade335e-62e6-4abb-b190-6efd48e077fc?reveal_credentials=oidc_token

{
  "id": "3ade335e-62e6-4abb-b190-6efd48e077fc",
  "credentials": {
    "oidc": {
      "type": "oidc",
      "identifiers": [
        "google:j8kf7a3..."
      ],
      "created_at": "2021-08-14T11:09:32.460548Z",
      "updated_at": "2021-08-14T11:09:32.460548Z"
    }
  },
  "identifier_credentials": [
    {
      "subject": "j8kf7a3...",
      "provider": "google-kratos-test",
      "access_token": "*****************",
      "refresh_token": "**********************************"
    }
  ],
  "schema_id": "default",
  "schema_url": "http://127.0.0.1:4433/schemas/default",
  "state": "active",
  "state_changed_at": "2021-08-14T11:08:42.5200046Z",
  "traits": {
    "name": {
      "first": "Foo",
      "last": "Bar"
    },
    "email": "foo@ory.sh"
  },
  "verifiable_addresses": [
    {
      "id": "8db5996b-f76b-4f4b-83aa-6745b2edb6a3",
      "value": "foo@ory.sh",
      "verified": false,
      "via": "email",
      "status": "sent",
      "verified_at": null,
      "created_at": "2021-08-14T11:08:42.52204Z",
      "updated_at": "2021-08-14T11:09:32.457381Z"
    }
  ],
  "recovery_addresses": [
    {
      "id": "f9ecd4a2-3e41-4384-9614-cf97f60acbf9",
      "value": "foo@ory.sh",
      "via": "email",
      "created_at": "2021-08-14T11:08:42.522253Z",
      "updated_at": "2021-08-14T11:09:32.457889Z"
    }
  ],
  "created_at": "2021-08-14T11:08:42.521706Z",
  "updated_at": "2021-08-14T11:08:42.521706Z"
}
```
