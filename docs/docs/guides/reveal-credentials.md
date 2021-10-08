---
id: reveal-credentials title: Get Access, Refresh, ID Tokens from Social Sign In
---

This document describes how to retrieve OpenID Connect / OAuth 2.0 Access, Refresh, and ID Tokens from Social Sign In at
the `GET /identities` API. This guide assumes that you have the `oidc` method enabled.

```shell script
$ curl --request GET -sL \
    --header "Content-Type: application/json" \
    http://127.0.0.1:4434/identities/3ade335e-62e6-4abb-b190-6efd48e077fc?declassify_credential=oidc

{
  "id": "714a9ddc-9fde-42ad-be42-784dfeadd098",
  "credentials": {
    "oidc": {
      "type": "oidc",
      "identifiers": [
        "google:some-user"
        "github:another-user"
      ],
      "config": {
        "providers": [
          {
            "initial_id_token": "********************",
            "subject": "some-user",
            "provider": "google",
            "initial_access_token": "********************",
            "initial_refresh_token": "********************",
          },
          {
            "initial_id_token": "********************",
            "subject": "another-user",
            "provider": "github",
            "initial_access_token": "********************",
            "initial_refresh_token": "********************",
          }
        ]
      },
      "created_at": "2021-10-08T12:17:18.834351+02:00",
      "updated_at": "2021-10-08T12:17:18.834351+02:00"
    }
  },
  "schema_id": "default",
  "schema_url": "http://localhost:61342/schemas/default",
  "state": "active",
  "state_changed_at": "2021-10-08T12:17:18.83324+02:00",
  "traits": {
    "subject": "foo.oidc@bar.com"
  },
  "verifiable_addresses": [
    {
      "id": "88da96df-0457-4d69-832d-5e70ef25055c",
      "value": "foo.oidc@bar.com",
      "verified": false,
      "via": "",
      "status": "",
      "verified_at": null,
      "created_at": "2021-10-08T12:17:18.83324+02:00",
      "updated_at": "2021-10-08T12:17:18.834202+02:00"
    }
  ],
  "created_at": "2021-10-08T12:17:18.834043+02:00",
  "updated_at": "2021-10-08T12:17:18.834043+02:00"
}
```

## Encryption

By default Access Token and Refresh Token are plaintext recorded
[Noop Cipher](setting-up-noop-cipher-parameters.mdx)

For a tighter security aspect you could choose following cipher:

- AES by following this [setup](setting-up-aes-cipher-parameters.mdx)
- XChaCha20 Poly1305 by following this
  [setup](setting-up-xchacha-cipher-parameters.mdx)
