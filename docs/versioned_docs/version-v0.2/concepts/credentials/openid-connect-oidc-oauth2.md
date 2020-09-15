---
id: openid-connect-oidc-oauth2
title: Social Sign In, OpenID Connect, and OAuth2
---

:::info

OpenID Connect is undergoing active refactoring and these docs will change. See
[#381](https://github.com/ory/kratos/pull/381).

:::

The `oidc` method uses OpenID Connect, or OAuth2 where OpenID Connect is not
supported, to authenticate identities using a third-party identity provider,
such as Google, Microsoft, GitHub - or any other OAuth2 / OpenID Connect
provider (for example [ORY Hydra](https://www.ory.sh/hydra)).

## Configuration

You can configure multiple OAuth2 / OpenID Connect providers. First, enable the
`oidc` method:

```yaml
selfservice:
  strategies:
    oidc:
      enabled: true
```

Next, you need to configure the providers you want to use (e.g. GitHub). Each
provider requires:

```yaml
id: github # The ID of the provider. DO NOT change this once this is in use.

# The provider you would like to use. ORY Kratos comes with some predefined providers to make
# life easier for you, but you can always opt for the "generic" provider, which works
# with any Certified OpenID Connect Provider (Google, ORY Hydra, ...):
provider: generic

# Other supported providers are (more to come):
#
# provider: github
# provider: google

# The OAuth2 / OpenID Connect provider will provide you with a OAuth2 Client ID and Client Secret. You need
# to set them here:
client_id: ...
client_secret: ...

schema_url: http://mydomain.com/github.schema.json # See section "Schema"

# What scope to request. Usually, this would be something like "profile" or "email".
# Please check the documentation of the OAuth2 / OpenID Connect provider to see what's allowed here.
scope:
  - email

# issuer_url is the OpenID Connect Server URL. You can leave this empty if `provider` is not set to `generic`.
# If set, neither `auth_url` nor `token_url` are required.
issuer_url: http://openid-connect-provider/

# auth_url is the authorize url, typically something like: https://example.org/oauth2/auth
# Should only be used when the OAuth2 / OpenID Connect server is not supporting OpenID Connect Discovery and when
# `provider` is set to `generic`.
auth_url: http://openid-connect-provider/oauth2/auth

# token_url is the token url, typically something like: https://example.org/oauth2/token
# Should only be used when the OAuth2 / OpenID Connect server is not supporting OpenID Connect Discovery and when
# `provider` is set to `generic`.
token_url: http://openid-connect-provider/oauth2/token
```

## JSON Schema

This strategy expects that you've set up your default JSON Schema for identity
traits. There are no extra settings for that.

You do however need to set up an additional JSON Schema for your provider. This
is required because we need to transform profile data coming from, for example
GitHub, to your identity model.

Defining that JSON Schema also allows you to require certain information. If you
ask the user to authorize the `photos` scope for example, you can configure the
JSON Schema in such a way that `photos` must be part of the identity data or the
flow will fail.

You will also need to project data coming from the provider onto your own data
model. You can express this using a JSON Path
([learn more about the syntax](../../reference/json-schema-json-paths.md)) in your
JSON Schema. Let's assume you want to map field `username` from the provider to
field `traits.name` in your identity:

```yaml
{
  '$id': 'https://example.com/social.schema.json',
  '$schema': 'http://json-schema.org/draft-07/schema#',
  'type': 'object',
  'properties':
    {
      'username':
        {
          'type': 'string',
          'ory.sh/kratos':
            { 'mappings': { 'identity': { 'traits': [{ 'path': 'name' }] } } },
        },
    },
  'required': ['username'],
}
```

If the OpenID Connect provider returns

```json
{
  "sub": "123123123",
  "username": "john.doe"
}
```

for example (`sub` is the OpenID Connect field for the identity's ID), that
would be transformed to identity:

```yaml
id: '9f425a8d-7efc-4768-8f23-7647a74fdf13'

credentials:
  oidc:
    id: oidc
    identifiers:
      - example:123123123
    config:
      - provider: example
        identifier: 123123123

traits_schema_url: http://foo.bar.com/person.schema.json # This come from the default identity schema url.

traits:
  name: john.doe # This is extracted from `username` using
```

## Example: Sign in with GitHub

Let's say you want to enable "Sign in with GitHub". All you have to do is:

- Create a
  [GitHub OAuth2 Client](https://developer.github.com/apps/building-oauth-apps/creating-an-oauth-app/)
- Set the "Authorization callback URL" to:
  `http://<domain-of-ory-kratos>:<public-port>/auth/browser/methods/oidc/callback/<provider-id>`

```yaml
selfservice:
  strategies:
    oidc:
      enabled: true
      config:
        providers:
          - id: github # this is `<provider-id>` in the Authorization callback URL
            provider: github
            client_id: .... # Replace this with the OAuth2 Client ID provided by GitHub
            client_secret: .... # Replace this with the OAuth2 Client Secret provided by GitHub
            schema_url: http://mydomain.com/github.schema.json # See section "Schema"
            scope:
              - user:email
```

The following schema would take `email_primary` and `username` and map them into
your identity model to `traits.email` and `traits.name`:

```json
{
  "$id": "http://mydomain.com/github.schema.json ",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "email_primary": {
      "type": "string",
      "ory.sh/kratos": {
        "mappings": {
          "identity": {
            "traits": [
              {
                "path": "email"
              }
            ]
          }
        }
      }
    },
    "username": {
      "type": "string",
      "ory.sh/kratos": {
        "mappings": {
          "identity": {
            "traits": [
              {
                "path": "name"
              }
            ]
          }
        }
      }
    }
  }
}
```

More examples will soon follow.
