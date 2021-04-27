---
id: identity-data-model
title: Identity Data Model
---

import Mermaid from '@theme/Mermaid'

An identity ("user", "user account", "account", "subject") is the "who" of a
software system. It can be a customer, employee, user, contractor, or even a
programmatic identity such as an IoT device, application, or some other type of
"robot."

:::info

In Ory Kratos' terminology we call all of them "identities", and it is always
exposed as `identity` in the API endpoints, requests, and response payloads. In
the documentation however, we mix these words as "account recovery" or "account
activation" is a widely accepted and understood terminology and user flow, while
"identity recovery" or "identity activation" is not.

:::

The following examples use YAML for improved readability. However, the API
payload is usually in JSON format. An `identity` has the following properties:

```yaml title="$ curl http://kratos/admin-endpoint/identities/9f425a8d-7efc-4768-8f23-7647a74fdf13"
# A UUID that is generated when the identity is created and
# which cannot be changed or updated at a later stage.
id: '9f425a8d-7efc-4768-8f23-7647a74fdf13'

# This section represents all the credentials associated with this identity.
# It is further explained in the "Credentials" section.
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

# This is the JSON Schema ID used for validating the identities' traits.
# `default` is a special keyword to set this to the schema set by
# `default_schema_url`, but it can be any another schema as well.
# e.g. customer, employee, employee-v2
schema_id: default

# Traits represent information about the identity, such as the first or last name. The traits content is completely
# up to you and will be validated using the JSON Schema at `traits_schema_url`.
traits:
  # These are just examples
  email: office@ory.sh
  name:
    first: Aeneas
    last: Rekkas
  favorite_animal: Dog
  accepted_tos: true
```

## Identity State

Identities are

- `created` - via API or self-service registration;
- `updated` - via API or self-service settings, account recovery, etc.;
- `disabled` - not yet implemented, see
  [#598](https://github.com/ory/kratos/issues/598);
- `deleted` - via API or with a self-service flow (not yet implemented see
  [#596](https://github.com/ory/kratos/issues/596)).

The identity state is therefore `active` or `disabled` (not yet implemented see
[#598](https://github.com/ory/kratos/issues/598))

<Mermaid
chart={`stateDiagram-v2 [*] --> Active: create Active --> Active: update Active --> Disabled: disable Disabled --> [*]: delete Disabled --> Active: enable `}
/>

## Identity Traits and JSON Schemas

Traits are data associated with an identity. You have to define its schema
according to your application's needs. They are also supposed to be modified by
the identity itself e.g. as part of the registration or profile update process
as well as anyone having access to Ory Krato's Admin API.

Ory Kratos uses
[JSON Schema](https://json-schema.org/learn/getting-started-step-by-step.html)
to validate Identity Traits.

Ory Kratos defines JSON Schema extension "Vocabulary" that allows you to tell
Ory Kratos that a specific trait adds some specific meaning to the standard JSON
Schema (more on that later).

Each identity can, theoretically, have a different JSON Schema. This is useful
in the following situations:

- there is more than one type of identity in the system for instance customers,
  support or staff;
- the system includes both users and robots sometimes also known as named
  service accounts;
- the system needs to ingest another company's identity model, and
- the system's identity model changes or grows over time and requires
  versioning.

The following example illustrates a usage scenario with three types of
identities: Regular customers,
[grandfather accounts](https://en.wikipedia.org/wiki/Grandfather_clause), and
Service Accounts (e.g. Microsoft provides
[Service Accounts](https://docs.microsoft.com/en-us/windows/security/identity-protection/access-control/service-accounts)).
There would be one JSON Schema per type of identity:

- Customers: `http://mydomain.com/schemas/v2/customer.schema.json`
- Grandfather Accounts: `http://mydomain.com/schemas/v1/customer.schema.json`
- Service Accounts: `http://mydomain.com/schemas/service-account.schema.json`

Ory Kratos expects the JSON Schemas in its configuration file:

```yaml
identity:
  # This will be the default JSON Schema. If `schema_id` is empty when creating an identity using the
  # Admin API, or a user signs up using a selfservice flow, this schema will be used.
  #
  # This is a required configuration field!
  default_schema_url: http://foo.bar.com/person.schema.json

  # Optionally define additional schemas here:
  schemas:
    # When creating an identity that uses this schema, `traits_schema_id: customer` would be set for that identity.
    - id: customer
      url: http://foo.bar.com/customer.schema.json
```

Ory Kratos validates the Identity Traits against the corresponding schema on all
writing operations (create / update). The employed business logic must be able
to distinguish these three types of identities. You might use a switch statement
like in the following example:

```go
// This is an example program that can deal with all three types of identities
session, err := ory.SessionFromRequest(r)
// some error handling
switch (session.Identity.SchemaID) {
    case "customer":
        // ...
    case "employee":
        // ...
    case "default":
        fallthrough
    default:
        // ...
}
```

## JSON Schema Vocabulary Extensions

Because Ory Kratos does not know that a particular field has a system-relevant
meaning you have to specify these in the schema. This includes for example:

- The email address for recovering a lost password
- The identifier for logging in with a password e.g. username or email
- The phone number for enabling SMS 2FA
- ...

Ory Kratos' JSON Schema Vocabulary Extension can be used within a property:

```json
{
  "$id": "http://mydomain.com/schemas/v2/customer.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "A customer (v2)",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "email": {
          "title": "E-Mail",
          "type": "string",
          "format": "email",

          // This tells Ory Kratos that the field should be used as the "username" for the username+password flow.
          // It is an extension to the regular JSON Schema vocabulary.
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
  }
}
```

An overview of available configuration options follows in the next sections.

### Identifier for Username and Password Flows

You can configure Ory Kratos to use specific fields as the _identifier_ e.g.
username, email, phone number, etc., in the Username and Password Registration
and Login Flow:

```json
{
  "ory.sh/kratos": {
    "credentials": {
      "password": {
        "identifier": true
      }
    }
  }
}
```

Looking at the traits from above

```yaml
traits:
  # These are just examples
  email: office@ory.sh
  name:
    first: Aeneas
    last: Rekkas
  favorite_animal: Dog
  accepted_tos: true
```

and using a JSON Schema that uses the `email` field as the identifier for the
password flow

```json
{
  "$id": "http://mydomain.com/schemas/v2/customer.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "A customer (v2)",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "email": {
          "title": "E-Mail",
          "type": "string",
          "format": "email",

          // This tells Ory Kratos that the field should be used as the "username" for the Username and Password Flow.
          "ory.sh/kratos": {
            "credentials": {
              "password": {
                "identifier": true
              }
            }
          }
        },
        "name": {
          "type": "object",
          "properties": {
            "first": {
              "type": "string"
            },
            "last": {
              "type": "string"
            }
          }
        },
        "favorite_animal": {
          "type": "string"
        },
        "accepted_tos": {
          "type": "string"
        }
      },
      "required": ["email"],
      "additionalProperties": false
    }
  }
}
```

In this example, Ory Kratos understands that traits:email: `office@ory.sh` is
the identity's identifier. The system expects `office@ory.sh` plus a password to
sign in.

[Username and Password Credentials](credentials/username-email-password.mdx)
contains more information and examples.

There are currently no other extensions supported for Identity Traits. Further
fields will be added in future releases!
