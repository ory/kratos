---
id: ui-user-interface
title: User Interface
---

Ory Kratos has no user interface included. Instead, it defines HTTP flows and
APIs that make it simple to write your own UI in a variety of languages and
frameworks.

The following two examples are typical UIs used in connection with Ory Kratos.

## Administrative User Interface (Admin UI)

The AUI might show all of the identities in the system and provide features to
administrators such as editing profiles, resetting passwords, and so on.

At present, there is no Open Source AUI for Ory Kratos.

## Self-service User Interface (SSUI)

The SSUI renders forms such as "Login", "Registration", "Update your profile",
"Recover access to your account", and others. The following provides more
reference for SSUI at
[github.com/ory/kratos-selfservice-ui-node](https://github.com/ory/kratos-selfservice-ui-node).

The SSUI can be built in any programming language including Java, Node, or
Python and can be run both a server or a end-user device for example a browser,
or a mobile phone. Implementing a SSUI is simple and straight forward. There is
no complex authentication mechanism required and no need to worry about possible
attack vectors such as CSRF or Session Attacks since Ory Kratos provides the
preventive measures built in.

Chapter [Self-Service Flows](../self-service.mdx) contains further information
on APIs and flows related to the SSUI, and build self-service applications.

### UI Data Models

To make UI customization easy, Ory Kratos prepares all the necessary data for
forms that need to be shown during e.g. login, registration:

```json
{
  "id": "9b527900-2199-4221-9252-7971b3362282",
  "type": "browser",
  "expires_at": "2021-04-28T13:55:36.046466067Z",
  "issued_at": "2021-04-28T12:55:36.046466067Z",
  "ui": {
    "action": "http://127.0.0.1:4433/self-service/settings?flow=9b527900-2199-4221-9252-7971b3362282",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "csrf_token",
          "type": "hidden",
          "value": "U3r/lgEfT8rA1Lg0Eeo06oGO8mX6T4TKoe/z7rbInhvYeacbRg0IW9zrqnpU1wmQJXKiekNzdLnypx5naHXoPg==",
          "required": true,
          "disabled": false
        },
        "messages": null,
        "meta": {}
      }
    ]
  }
}
```

#### Node Groups

Nodes are grouped (using the `group` key) based on the source that generated the
node. Sources are the different methods such as "password" ("Sign in/up with ID
& password"), "oidc" (Social Sign In), "link" (Password reset and email
verification), "profile" ("Update your profile") and the "default" group which
typically contains the CSRF token.

You can use the node group to filter out items, re-arrange them, render them
differently - up to you!

#### Node Types

The first (and for now only) node type is the `input` type:

```json
{
  "type": "input",
  "group": "default",
  "attributes": {
    "name": "csrf_token",
    "type": "hidden",
    "value": "U3r/lgEfT8rA1Lg0Eeo06oGO8mX6T4TKoe/z7rbInhvYeacbRg0IW9zrqnpU1wmQJXKiekNzdLnypx5naHXoPg==",
    "required": true,
    "disabled": false
  },
  "messages": null,
  "meta": {}
}
```

It contains different attributes which you can map 1:1 to an HTML form:

```html
<input
  type="hidden"
  name="csrf_token"
  value="U3r/lgEfT8rA1Lg0Eeo06oGO8mX6T4TKoe/z7rbInhvYeacbRg0IW9zrqnpU1wmQJXKiekNzdLnypx5naHXoPg=="
  required
/>
```

Similarly, other form input types can be sent:

```json
[
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "traits.email",
      "type": "email",
      "value": "foo@ory.sh",
      "disabled": false
    },
    "messages": null,
    "meta": {
      "label": {
        "id": 1070002,
        "text": "E-Mail",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "method",
      "type": "submit",
      "value": "profile",
      "disabled": false
    },
    "messages": null,
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    }
  }
]
```

As you can see, some fields even include `meta.label` information which you can
use for the labels:

```html
<fieldset>
  <input type="email" name="traits.email" value="foo@ory.sh" />
  <span>E-Mail</span>
</fieldset>
<fieldset>
  <input type="submit" name="method" value="profile" />
  <span>Save</span>
</fieldset>
```

#### Node Order and Labels

For all traits, the labels and orders are taken from the Identity JSON Schema. A
JSON Schema such as

```json
{
  "$id": "https://schemas.ory.sh/presets/kratos/quickstart/email-password/identity.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "traits": {
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
        "website": {
          "title": "Website",
          "type": "string",
          "format": "uri",
          "minLength": 10
        },
        "consent": {
          "title": "Consent",
          "const": true
        },
        "newsletter": {
          "title": "Newsletter",
          "type": "boolean"
        }
      },
      "required": ["email", "website"],
      "additionalProperties": false
    }
  }
}
```

will generate the following fields - take note that the order of the JSON Schema
affects the order of the nodes:

```json
[
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "traits.email",
      "type": "email",
      "value": "foo@ory.sh",
      "disabled": false
    },
    "messages": null,
    "meta": {
      "label": {
        "id": 1070002,
        "text": "E-Mail",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "traits.name.first",
      "type": "text",
      "value": "Foo",
      "disabled": false
    },
    "messages": null,
    "meta": {
      "label": {
        "id": 1070002,
        "text": "First Name",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "traits.name.last",
      "type": "text",
      "value": "Bar",
      "disabled": false
    },
    "messages": null,
    "meta": {
      "label": {
        "id": 1070002,
        "text": "Last Name",
        "type": "info"
      }
    }
  },
  {
    "type": "input",
    "group": "profile",
    "attributes": {
      "name": "method",
      "type": "submit",
      "value": "profile",
      "disabled": false
    },
    "messages": null,
    "meta": {
      "label": {
        "id": 1070003,
        "text": "Save",
        "type": "info"
      }
    }
  }
]
```

Generally, submit buttons are the last node in a group. If you wish to have more
flexibility with regards to order or labeling the best option is to implement
this in your UI using map, filter, and other JSON transformation functions.

#### Messages

Ory Kratos helps users understand what is happening by providing messages that
explain what went wrong or what needs to be done. Examples are "The provided
credentials are invalid", "Missing property email" and similar.

Typically login, registration, settings, ... flows include such messages on
different levels:

1. At the root level, indicating that the message affects the whole request
   (e.g. request expired)
2. At the method (password, oidc, profile) level, indicating that the message
   affects a specific method / form.
3. At the field level, indicating that the message affects a form field (e.g.
   validation errors).

Each message has a layout of:

```json5
{
  id: 1234,
  // This ID will not change and can be used to translate the message or use your own message content.
  text: 'Some default text',
  // A default text in english that you can display if you do not want to customize the message.
  context: {}
  // A JSON object which may contain additional fields such as `expires_at`. This is helpful if you want to craft your own messages.
}
```

We will list all messages, their contents, their contexts, and their IDs at a
later stage. For now please check out the code in the
[text module](https://github.com/ory/kratos/tree/master/text).

The message ID is a 7-digit number (`xyyzzzz`) where

- `x` is the message type which is either `1` for an info message (e.g.
  `1020000`), `4` (e.g. `4020000`) for an input validation error message, and
  `5` (e.g. `5020000`) for a generic error message.
- `yy` is the module or flow this error references and can be:
  - `01` for login messages (e.g. `1010000`)
  - `02` for logout messages (e.g. `1020000`)
  - `03` for multi-factor authentication messages (e.g. `1030000`)
  - `04` for registration messages (e.g. `1040000`)
  - `05` for settings messages (e.g. `1050000`)
  - `06` for account recovery messages (e.g. `1060000`)
  - `07` for email/phone verification messages (e.g. `1070000`)
- `zzzz` is the message ID and typically starts at `0001`. For example, message
  ID `4070001` (`4` for input validation error, `07` for verification, `0001`
  for the concrete message) is:
  `The verification code has expired or was otherwise invalid. Please request another code.`.
