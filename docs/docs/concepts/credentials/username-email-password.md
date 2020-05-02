---
id: username-email-password
title: Username / Email & Password
---

The `password` method is the most commonly used form of authentication, it
requires an `identifier` (username, email, phone number, ...) and a `password`
during registration and login.

ORY Kratos hashes the password after registration, password reset, and password
change using [Argon2](https://github.com/P-H-C/phc-winner-argon2), the winner of
the Password Hashing Competition (PHC).

## Configuration

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

## JSON Schema

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

## Example

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
