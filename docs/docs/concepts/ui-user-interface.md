---
id: ui-user-interface
title: User Interface
---

ORY Kratos has no user interface included. Instead, it defines HTTP flows and
APIs that make it simple to write your own UI in a variety of languages and
frameworks.

The following two examples are typical UIs used in connection with ORY Kratos.

## Administrative User Interface (Admin UI)

The AUI might show all of the identities in the system and provide features to
administrators such as editing profiles, resetting passwords, and so on.

At present, there is no Open Source AUI for ORY Kratos.

## Self-service User Interface (SSUI)

The SSUI shows screens such as "login", "Registration", "Update your profile",
"Recover access to your account", and others. The following provides more reference for
SSUI at
[github.com/ory/kratos-selfservice-ui-node](https://github.com/ory/kratos-selfservice-ui-node).

The SSUI can be built in any programming language including Java, Node, or
Python and can be run both a server or a end-user device for example a browser,
or a mobile phone. Implementing a SSUI is simple and straight forward. There is
no complex authentication mechanism required and no need to worry about possible
attack vectors such as CSRF or Session Attacks since ORY Kratos provides the
preventive measures built in.

Chapter [Self-Service Flows](../self-service/flows/index) contains further
information on APIs and flows related to the SSUI, and build self service
applications.

## Messages

ORY Kratos helps users understand what is happening by providing messages that
explain what went wrong or what needs to be done. Examples are
"The provided credentials are invalid", "Missing property email" and similar.

Typically login, registration, settings, ... flows include such messages on
different levels:

1. At the root level, indicating that the message affects the whole request (e.g. request expired)
2. At the method (password, oidc, profile) level, indicating that the message affects a specific
method / form.
3. At the field level, indicating that the message affects a form field (e.g. validation errors).

Each message has a layout of:

```json5
{
  id: 1234, // This ID will not change and can be used to translate the message or use your own message content.
  text: "Some default text", // A default text in english that you can display if you do not want to customize the message.
  context: {} // A JSON object which may contain additional fields such as `expires_at`. This is helpful if you want to craft your own messages.
}
```

We will list all messages, their contents, their contexts, and their IDs at a later stage. For now
please check out the code in the [text module](https://github.com/ory/kratos/tree/master/text).

The message ID is a 7-digit number (`xyyzzzz`) where

- `x` is the message type which is either `1` for an info message (e.g. `1020000`),
`4` (e.g. `4020000`) for an input validation error message, and `5` (e.g. `5020000`) for a generic error message.
- `yy` is the module or flow this error references and can be:
  - `01` for login messages (e.g. `1010000`)
  - `02` for logout messages (e.g. `1020000`)
  - `03` for multi-factor authentication messages (e.g. `1030000`)
  - `04` for registration messages (e.g. `1030000`)
  - `05` for settings messages (e.g. `1050000`)
  - `06` for account recovery messages (e.g. `1060000`)
  - `07` for email/phone verification messages (e.g. `1070000`)
- `zzzz` is the message ID and typically starts at `0001`.
   For example, message ID `4070001` (`4` for input validation error, `07` for verification, `0001` for the concrete message) is: `The verification code has expired or was otherwise invalid. Please request another code.`.
