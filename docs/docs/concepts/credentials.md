---
id: credentials
title: Credentials
---

Each identity has one or more credentials associated with it:

```yaml
credentials:
  password:
    id: password
    identifiers:
      - john.doe@acme.com
      - johnd@ory.sh
    config:
      hashed_password: ...
  oidc:
    id: oidc
    identifiers:
      - google:j8kf7a3...
      - facebook:83475891...
    config:
      - provider: google
        identifier: j8kf7a3
      - provider: facebook
        identifier: 83475891
```

ORY Kratos supports several credential types:

- `password`: The most common _identifier (username, email, ...) + password_
  credential.
- `oidc`: The "Log in with Google/Facebook/GitHub/..." credential.
- Other credentials - support other credential types (X509 Certificates,
  Biometrics, ...) at will be added a later stage.

Each credential - regardless of its type - has one or more identifiers attached
to it. Each identifier is universally unique. Assuming we had one identity with
credentials

```yaml
credentials:
  password:
    id: password
    identifiers:
      - john.doe@acme.com
```

and tried to create (or update) another identity with the same identifier
(`john.doe@acme.com`), the system would reject the request with a 409 Conflict
state.

While credentials must be unique per type, the can be duplicates amongst
multiple types:

```yaml
# This is ok:
credentials:
  password:
    id: password
    identifiers:
      - john.doe@acme.com
  oidc:
    id: oidc
    identifiers:
      - john.doe@acme.com
```

The same would apply if those were two separate identities:

```yaml
# Identity 1
credentials:
  password:
    id: password
    identifiers:
      - john.doe@acme.com
---
# Identity 2
credentials:
  oidc:
    id: oidc
    identifiers:
      - john.doe@acme.com
```

## Username and Password

The `password` method is the most commonly used form of authentication, it
requires an `identifier` (username, email, phone number, ...) and a `password`
during registration and login.

ORY Kratos hashes the password after registration, password reset, and password
change using [Argon2](https://github.com/P-H-C/phc-winner-argon2), the winner of
the Password Hashing Competition (PHC).

### Configuration

Enabling this method is as easy as setting

```yaml
selfservice:
  strategies:
    password:
      enabled: true
```

in your ORY Kratos configuration. You can configure the Argon2 hasher using the
following options:

```yaml
hashers:
  argon2:
    parallelism: 1
    memory: 131072 # 128MB
    iterations: 3
    salt_length: 16
    key_length: 32
```

For a complete reference, defaults, and description please check the
[Configuration Reference](../reference/configuration.md).

For a better understanding of security implications imposed by Argon2
Configuration, head over to [Argon2 Security](./security.md#argon2).

### JSON Schema

When processing an identity and its traits, the method will use the JSON Schema
to extract one or more identifiers. Assuming you want your identities to sign up
with an email address, and use that email address as a valid identifier during
login, you can use a schema along the lines of:

```json
{
  "$id": "https://example.com/example.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "email": {
      "type": "string",
      "format": "email",
      "title": "E-Mail",
      "ory.sh/kratos": {
        "credentials": {
          "password": {
            "identifier": true
          }
        }
      }
    }
  }
}
```

If you want a unique username instead, you could write the schema as follows:

```json
{
  "$id": "https://example.com/example.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "username": {
      "type": "string",
      "title": "Username",
      "ory.sh/kratos": {
        "credentials": {
          "password": {
            "identifier": true
          }
        }
      }
    }
  }
}
```

You are not limited to one identifier per identity. You could also combine both
fields and support a use case of "username" and "email" as an identifier for
login:

```json
{
  "$id": "https://example.com/example.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "email": {
      "type": "string",
      "format": "email",
      "title": "E-Mail",
      "ory.sh/kratos": {
        "credentials": {
          "password": {
            "identifier": true
          }
        }
      }
    },
    "username": {
      "type": "string",
      "title": "Username",
      "ory.sh/kratos": {
        "credentials": {
          "password": {
            "identifier": true
          }
        }
      }
    }
  }
}
```

### Example

Assuming your traits schema is as follows:

```json
{
  "$id": "https://example.com/example.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "first_name": {
      "type": "string"
    },
    "email": {
      "type": "string",
      "format": "email",
      "ory.sh/kratos": {
        "credentials": {
          "password": {
            "identifier": true
          }
        }
      }
    },
    "username": {
      "type": "string",
      "ory.sh/kratos": {
        "credentials": {
          "password": {
            "identifier": true
          }
        }
      }
    }
  },
  "additionalProperties": false
}
```

And an identity registers with the following JSON payload (more on registration
in
[Selfservice Registration](../self-service/flows/user-login-user-registration.md)):

```json
{
  "traits": {
    "first_name": "John Doe",
    "email": "john.doe@example.org",
    "username": "johndoe123"
  },
  "password": "my-secret-password"
}
```

The `password` method would generate a credentials block as follows:

```yaml
credentials:
  password:
    id: password
    identifiers:
      - john.doe@example.org
      - johndoe123
    config:
      hashed_password: ... # this would be `argon2(my-secret-password)`
```

Because credential identifiers need to be unique, no other identity can be
created that has `johndoe123` or `john.doe@example.org` as their `email` or
`username`.

## Social Sign In / OpenID Connect / OAuth2

The `oidc` method uses OpenID Connect, or OAuth2 where OpenID Connect is not
supported, to authenticate identities using a third-party identity provider,
such as Google, Microsoft, GitHub - or any other OAuth2 / OpenID Connect
provider (for example [ORY Hydra](https://www.ory.sh/hydra)).

### Configuration

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

### JSON Schema

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
([learn more about the syntax](../reference/json-schema-json-paths.md)) in your
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

### Example: Sign in with GitHub

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
