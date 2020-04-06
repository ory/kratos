---
id: user-settings-profile-management
title: User Settings and Profile Management
---

ORY Kratos allows users to update their own settings and profile information
using two principal flows:

- Browser-based (easy): This flow works for all applications running on top of a
  browser. Websites, single-page apps, Cordova/Ionic, and so on.
- API-based (advanced): This flow works for native applications like iOS
  (Swift), Android (Java), Microsoft (.NET), React Native, Electron, and others.

This flow does not allow updates of security-sensitive information such as the
password, fields associated with login (e.g. email), fields associated with
account recovery (e.g. recovery email address). These fields must be updated
using a separate flow which requires prior security checks. using a separate
flow which requires prior security checks.

The updated profile must be valid against the JSON Schema defined for its
[Identity Traits](../../concepts/identity-user-model.md). If one or more fields
do not validate (e.g. "Not an email"), the profile will not be updated.

The only required configuration is setting the Profile UI URL in the
[ORY Kratos configuration file](../../reference/configuration.md):

```yaml
urls:
  settings_ui: https://.../..
```

## Self-Service User Profile Management for Browser Applications

This flow is similar to
[User Login and User Registration](user-login-user-registration.md) but does not
support before/after work flows or individual strategies. It uses the already
established [Network Flows for Browsers](index.md#network-flows-for-browsers).

### Server-Side Browser Applications

The [Network Flows for Browsers](index.md#network-flows-for-browsers) works as
follows for Profile Management:

1. An initial HTTP Request is made to
   `https://example.org/.ory/kratos/public/self-service/browser/flows/settings`.
2. ORY Kratos redirects the browser to the URL set in `urls.settings_ui` and
   appends the `request` URL Query Parameter (e.g.
   `https://example.org/settings?request=abcde`).
3. The Endpoint at `/settings` makes a HTTP GET Request to
   `https://ory-kratos-admin.example-org.vpc/self-service/browser/flows/requests/settings?request=abcde`
   and fetches Profile Management Request JSON Payload that represent the
   individual fields that can be updated.
4. The User updates the profile data and sends a HTTP POST request to, e.g.,
   `https://example.org/.ory/kratos/public/self-service/browser/flows/settings/strategies/password?request=abcde`.
   - If the profile data is invalid, all validation errors will be collected and
     added to the Profile Management JSON Payload. The Browser is redirected to
     the `urls.settings_ui` URL (e.g.
     `https://example.org/profile?request=abcde`).
   - If the profile data is valid, the identity's traits are updated and the
     process is complete.

Assuming the Identity's Traits JSON Schema is defined as

```json
{
  "$id": "https://example.org/identity.traits.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "email": {
      "type": "string",
      "format": "email",
      "title": "E-Mail",
      "minLength": 3,
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
          "type": "string",
          "minLength": 5
        },
        "last": {
          "type": "string"
        }
      }
    }
  },
  "required": ["email"],
  "additionalProperties": false
}
```

the resulting JSON Payload coming from
`https://ory-kratos-admin.example-org.vpc/self-service/browser/flows/requests/settings?request=abcde`
would look something along the lines of:

```json
{
  "id": "48068b5d-3438-4d6f-9955-331b96c41762",
  "expires_at": "2020-01-27T23:03:58.5986947Z",
  "issued_at": "2020-01-27T22:03:58.5987099Z",
  "request_url": "http://127.0.0.1:4455/settings",
  "form": {
    "action": "https://example.org/.ory/kratos/public/settings?48068b5d-3438-4d6f-9955-331b96c41762",
    "method": "POST",
    "fields": {
      "csrf_token": {
        "name": "csrf_token",
        "type": "hidden",
        "required": true,
        "value": "+5+WxP7+leOpfjHHkWWe99APD7845i82p4wGtfdWKHBK5fFg4BS7JjzdeI7kdsOUElyrG10ZR5vIqi7asNpqAA=="
      },
      "traits.email": {
        "name": "traits.email",
        "type": "text",
        "value": "foobar@ory.sh"
      },
      "traits.name.first": {
        "name": "traits.name.first",
        "type": "text",
        "value": "Foobar"
      },
      "traits.name.last": {
        "name": "traits.name.last",
        "type": "text",
        "value": "Barbaz"
      }
    }
  },
  "identity": {
    "id": "c631e58a-445a-4844-ae80-f0b426a1e11e",
    "traits_schema_id": "default",
    "traits_schema_url": "https://example.org/identity.traits.schema.json",
    "traits": {
      "email": "foobar@ory.sh",
      "name": {
        "last": "Foobar",
        "first": "Barbaz"
      }
    }
  }
}
```

If the user tries to save profile data that does not validate against the
provided JSON Schema, error payloads will be added to the fields affected:

```json5
{
  id: '48068b5d-3438-4d6f-9955-331b96c41762',
  // ...
  form: {
    // ...
    fields: {
      // ...
      'traits.name.first': {
        name: 'traits.name.first',
        type: 'text',
        value: 'abc',
        errors: [
          {
            message: 'traits.name.first: Must be at least 5 characters long',
          },
        ],
      },
      // ...
    },
  },
  // ...
}
```

Keep in mind that it is not possible to update the `traits.email` field as
updating that field requires prior authentication.

> Updating these "protected" fields will be implemented in a future release of
> ORY Kratos.

### Client-Side Browser Applications

Because Client-Side Browser Applications do not have access to ORY Kratos' Admin
API, they must use the ORY Kratos Public API instead. The flow for a Client-Side
Browser Application is almost the exact same as the one for Server-Side
Applications, with the small difference that
`https://example.org/.ory/kratos/public/self-service/browser/flows/requests/settings?request=abcde`
would be called via AJAX instead of making a request to
`https://ory-kratos-admin.example-org.vpc/self-service/browser/flows/requests/settings?request=abcde`.

> To prevent brute force, guessing, session injection, and other attacks, it is
> required that cookies are working for this endpoint. The cookie set in the
> initial HTTP request made to `https://example.org/.ory/kratos/public/settings`
> MUST be set and available when calling this endpoint!

## Self-Service User Profile Management for API Clients

Will be addressed in a future release.
