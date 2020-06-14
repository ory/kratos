---
id: username-email-password
title: Username or Email and Password
---

The `password` strategy implements the most-common used form of login and
registration: An identifier (username, email, phone number, ...) and a password.

It implements several flows, specifically
[User Login and User Registration](../flows/user-login-user-registration.mdx)
and [User Settings](../flows/user-settings-profile-management.mdx).

To enable the `password` strategy, set `selfservice.strategies.password.enabled`
to true in your ORY Kratos configuration:

```yaml
selfservice:
  strategies:
    password:
      enabled: true
```

Passwords are hashed using the
[Argon2 Hashing Algorithm](../../concepts/security.md#Argon2) which can be
configured in the
[ORY Kratos Argon2 Configuration](../../concepts/security.md#Argon2).

When a user signs up using this strategy, the Default Identity Traits Schema
(set using `identity.traits.default_schema_url`) is used:

```yaml
identity:
  traits:
    # also supports http(s) of course
    default_schema_url: file:///path/to/default-identity.schema.json
```

If you don't know what that means, please read the
["Concepts: Identity" Chapter](../../concepts/identity-user-model.md).

## Choosing between Username, Email, Phone Number

Before you start, you need to decide what data you want to collect from your
users and why! It is hard to change this decision afterwards, so make sure
you've taken everything into account!

When logging in, the user will use a login identifier and a password to sign up
and in. The identifier can be

- a username - e.g. "john.doe" or "johndoe123" or "oryuser",
- an email address - e.g. `john.doe@gmail.com`,
- a phone number - e.g. `+49-1234-4321-1234-4321`.

All of these approaches have up- and downsides.

Using the email address as the login identifier is easy to remember, does not
require additional fields (because the email address is already being
collected), and is usually unique. It's usually unique because sometimes
companies use a "shared" email account (e.g. office@acme.org) to access
services. In that case, multiple real identities are using the same email
identifier to log in.

The email address however represents a unique identifier and personally
identifiable information (PII). An attacker could for example check if an email
address (e.g. `john.doe@gmail.com`) is registered at a site (e.g. an adult
website) and use that information for blackmail (see
[Account Enumeration Attacks](../../concepts/security.md#account-enumeration-attacks)).

The same considerations apply to using a phone number as the primary
registration & login identifier.

Using a free text username reduces the privacy risk because it is much harder to
make a connection between the username and a real world identity. It's still
possible in cases where users choose a username such as
"john.doe.from.mineapolis.1970", but finding the right username identifier is
still difficult and there is plausible deniability because anyone could use that
username.

A free text username however requires capturing additional fields (e.g. an email
address for password resets / account recovery) and is hard to remember. It is
often very difficult to find unique usernames as people tend to use a
combination of their names and initials (e.g. `john.doe`) which has a high
chance of collision. Therefore, one ends up with usernames such as
`john.doe1234432`.

It is important to understand that ORY Kratos lowercases all `password`
identifiers and therefore E-Mail addresses. Characters `+` or `.` which have
special meaning for some E-Mail Providers (e.g. GMail) are not normalized:

- `userNAME` is equal to `username`
- `foo@BaR.com` is equal to `foo@bar.com`
- `foo+baz@bar.com` is NOT equal to `foo@bar.com`
- `foo.baz@bar.com` is NOT equal to `foobar@bar.com`

You need to decide which route you want to take.

### Email and Password

To use the email address as the login identifier, define the following Identity
Traits Schema:

```json
{
  "$id": "https://example.com/registration.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
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
    }
  }
}
```

### Multiple E-Mails and Password

You can allow users to sign up with multiple E-Mail Addresses and use any of
those for log in:

```json
{
  "$id": "https://example.com/registration.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "emails": {
      "type": "array",
      "items": {
        "type": "string",
        "format": "email",
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
```

### Username and Password

To use a username as the login identifier, define the following Identity Traits
Schema:

```json
{
  "$id": "https://example.com/registration.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
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
  }
}
```

### Username and Email and Password

You may also mix usernames and passwords:

```json
{
  "$id": "https://example.com/registration.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
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
  }
}
```

### Phone Number And Password

> This will be addressed in a future release and is tracked as
> [kratos#137](https://github.com/ory/kratos/issues/137).

## Browser Clients

### Registration

This strategy uses the high-level registration flow defined in chapter
[Self-Service Registration User Flow](../flows/user-login-user-registration.mdx).

Once the user is redirected to the Registration UI URL, the endpoint responsible
for that URL makes a request to ORY Kratos' Public / Admin API and appends the
`request` query parameter.

ORY Kratos uses the JSON Schema defined in `identity.traits.default_schema_url`
to generate a list of form fields and add it to the Registration Request.

Using a JSON Schema like

```json title="my/identity.schema.json"
{
  "$id": "https://schemas.ory.sh/presets/kratos/quickstart/email-password/identity.schema.json",
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
          "type": "string"
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

will result in the following Registration Request

```json5
{
  id: '713df601-d6c8-4331-8195-c29b92db459f',
  expires_at: '2020-01-27T16:31:00.3507956Z',
  issued_at: '2020-01-27T16:21:00.3508076Z',
  request_url: 'http://127.0.0.1:4455/auth/browser/registration',
  methods: {
    password: {
      method: 'password',
      config: {
        action: 'http://127.0.0.1:4455/.ory/kratos/public/self-service/browser/flows/registration/strategies/password?request=713df601-d6c8-4331-8195-c29b92db459f',
        method: 'POST',
        fields: [
          {
            name: 'csrf_token',
            type: 'hidden',
            required: true,
            value: '0klCuilgIO2k0Ev3J3IEsMOlmxg0RhjiiiWXVKm3Pd7HxZLVkDHWoOSfiT+/BJn69Dg2fmq6MHv8HkEx6MrVlw==',
          },
          {
            name: 'traits.email',
            type: 'email',
            required: true,
          },
          {
            name: 'password',
            type: 'password',
            required: true,
          },
          {
            name: 'traits.name.first',
            type: 'text',
          },
          {
            name: 'traits.name.last',
            type: 'text',
          },
        ],
      },
    },
  },
}
```

which in turn is easily to render by filling out a HTML Form template:

```html
<form method="{{ method }}" action="{{ action }}">
  <!-- repeat this for every field -->
  <input type="{{ field.type }}" name="{{ field.name }} required="{{
  field.required }}" value="{{ field.value }}"
  <!-- ... -->>

  <input type="submit" value="Create account" />
</form>
```

Once the user clicks "Create Account", the payload will be sent to ORY Kratos'
Public API. The data will be validated against the JSON Schema (e.g. checking if
a required field is missing, if some condition like `minLength` is not
fulfilled, ...). If the data is invalid or incomplete, the browser will be
redirected to the same login endpoint with the same request ID. When fetching
that request ID again, error details will be included in the JSON Response, such
as:

```json5
{
  id: '713df601-d6c8-4331-8195-c29b92db459f',
  expires_at: '2020-01-27T16:31:00.3507956Z',
  issued_at: '2020-01-27T16:21:00.3508076Z',
  request_url: 'http://127.0.0.1:4433/self-service/browser/flows/registration',
  methods: {
    password: {
      method: 'password',
      config: {
        errors: [
          {
            message: 'Please update the Form Fields to proceed.',
          },
        ],
        action: 'http://127.0.0.1:4455/.ory/kratos/public/self-service/browser/flows/registration/strategies/password?request=713df601-d6c8-4331-8195-c29b92db459f',
        method: 'POST',
        fields: [
          /* ... */
          {
            name: 'password',
            type: 'text',
            value: 't4aegbydfv5234',
            errors: [
              {
                message: "traits.email: Does not match format 'email'",
              },
            ],
          },
          /* ... */
        ],
      },
    },
  },
}
```

> Validation error messages and context will be improved in future releases.
> This is tracked as [kratos#185](https://github.com/ory/kratos/issues/185).

### Login

The Login flow is almost identical to the registration flow. The only difference
is that only three fields will be requested:

```json5
{
  id: '0cfb0f7e-3866-453c-9c23-28cc2cb7fead',
  expires_at: '2020-01-27T16:48:53.8826084Z',
  issued_at: '2020-01-27T16:38:53.8826392Z',
  request_url: 'http://127.0.0.1:4433/self-service/browser/flows/login',
  methods: {
    password: {
      method: 'password',
      config: {
        action: 'http://127.0.0.1:4455/.ory/kratos/public/self-service/browser/flows/login/strategies/password?request=0cfb0f7e-3866-453c-9c23-28cc2cb7fead',
        method: 'POST',
        fields: [
          {
            name: 'csrf_token',
            type: 'hidden',
            required: true,
            value: 'F0LABRxm/os+18VBUcbmz98LkJid1sEj++4X41rcdbcCzhBqpTcIxn6YB4nJsHuF6JY9/sMq6bqN1cGGG6Gd/g==',
          },
          {
            name: 'identifier',
            type: 'text',
            required: true,
          },
          {
            name: 'password',
            type: 'password',
            required: true,
          },
        ],
      },
    },
  },
}
```

If the login form is filled out incorrectly, errors are included in the
response:

```json5
{
  id: '0cfb0f7e-3866-453c-9c23-28cc2cb7fead',
  expires_at: '2020-01-27T16:48:53.8826084Z',
  issued_at: '2020-01-27T16:38:53.8826392Z',
  request_url: 'http://127.0.0.1:4455/auth/browser/login',
  methods: {
    password: {
      method: 'password',
      config: {
        errors: [
          {
            message: 'Please check the data you provided.',
          },
        ],
        action: 'http://127.0.0.1:4455/.ory/kratos/public/self-service/browser/flows/login/strategies/password?request=0cfb0f7e-3866-453c-9c23-28cc2cb7fead',
        method: 'POST',
        fields: [
          /* ... */
          {
            name: 'password',
            type: 'password',
            required: true,
            errors: [
              {
                message: 'password: password is required',
              },
            ],
          },
        ],
      },
    },
  },
}
```

### Settings

The Settings flow allows a user to change his/her password. This action will
require the user to sign in again, unless the session is younger than the
configured:

```yaml title="path/to/kratos/config.yml"
selfservice:
  settings:
    # Sessions older than a minute requires the user to sign in again before
    # the password is changed.
    privileged_session_max_age: 1m
```

The Settings Request payload for this strategy looks as follows:

```json
{
  "id": "71da1753-e135-441c-b4df-e7b7cd90ad88",
  "expires_at": "2020-05-02T15:52:09.67209Z",
  "issued_at": "2020-05-02T14:52:09.67209Z",
  "request_url": "http://127.0.0.1:4433/self-service/browser/flows/settings",
  "active": "password",
  "methods": {
    "password": {
      "method": "password",
      "config": {
        "action": "http://127.0.0.1:4455/.ory/kratos/public/self-service/browser/flows/settings/strategies/password?request=71da1753-e135-441c-b4df-e7b7cd90ad88",
        "method": "POST",
        "fields": [
          {
            "name": "password",
            "type": "password",
            "required": true
          },
          {
            "name": "csrf_token",
            "type": "hidden",
            "required": true,
            "value": "UjEPiUMubRiAl0NG7yUzsww8XjpJvW+HBrh6JirjLxPqhlW2ql+0kYknjd8gdsx0v08vQSmqUEcZhNPsvkr2Kw=="
          }
        ]
      }
    }
  },
  "identity": {
    "id": "f48c43bb-50ea-4520-9280-37a891175aba",
    "traits": {
      "email": "h71x9a@j6r8c"
    }
  },
  "update_successful": false
}
```

If the form validation fails, an error will bei included:

```json5
{
  id: '71da1753-e135-441c-b4df-e7b7cd90ad88',
  // expires_at, ...
  active: 'password',
  methods: {
    config: {
      // action, method ...
      errors: [
        {
          message: 'Please check the data you provided.',
        },
      ],
      fields: [
        // ...
        {
          name: 'password',
          type: 'password',
          required: true,
          errors: [
            {
              message: 'password: password is required',
            },
          ],
        },
      ],
    },
  },
  // identity, ...
  update_successful: false,
}
```

A successful flow will be marked with:

```json5
{
  id: '71da1753-e135-441c-b4df-e7b7cd90ad88',
  // expires_at, ...
  update_successful: true,
}
```

Apart from that, there is nothing else to configure. Just render the HTML Form
which includes the update password field!

### Settings

FIXME - TO BE DONE

## API Clients

API-based login and registration using this strategy will be addressed in a
future release of ORY Kratos.

## Security and Defenses

Password-based authentication flows are subject to frequent abuse through

- Social Engineering Attacks;
- Password Guessing Attacks;
- Phishing Attacks.

### Anti-automation

> This feature is a work in progress and is tracked as
> [kratos#133](https://github.com/ory/kratos/issues/138).

Actions that cause out-of-band communications, such as sending an activation
link via email or an activation code via SMS, can be abused by automated
systems. The goal of such an attack is to send out so many emails or SMS, that
your reputation worsens (spam filters) or you're faced with massive costs
(carrier fees).

CAPTCHA renders these attacks either very difficult or impossible. ORY Kratos
has CAPTCHA support built-in. ORY Kratos will prompt the user to complete a
CAPTCHA in the following scenarios:

- The user tries to register more than one account within 72 hours.
- The user failed provide valid credentials for the third time within 12 hours.
- The user tries to recover their account for the second time within 72 hours.

For integration guidelines, please check the individual flow's (registration,
login, account recovery) integration documentation.

### Account Takeover Defenses

The Settings flow implements account takeover defenses as it is not possible to
change the password without knowing the existing password. A good example of
this flow is the
[GitHub sudo mode](https://help.github.com/en/github/authenticating-to-github/sudo-mode).

### Password Validation

> Further improvements are work in progress and are tracked on
> [GitHub](https://github.com/ory/kratos/issues?q=is%3Aopen+label%3Amodule%3Ass%2Fpassword+)

To prevent weak passwords ORY Kratos implements different measures. Users often
choose passwords similar to their traits. To prevent this ORY Kratos ensures
there is a sufficient
[Levenshtein-Distance](https://en.wikipedia.org/wiki/Levenshtein_distance) (aka
"Edit-Distance") between the identifier and the password. It also makes sure
that the identifier and password have a small enough longest common substring.

Furthermore the `password` strategy comes with a build-in check against the
["Have I been pwned"](https://haveibeenpwned.com) breach database. This way ORY
Kratos makes sure your users cannot use passwords like "password", "123456" or
any other commonly used one. To protect the value of the password the
[range API](https://haveibeenpwned.com/API/v3#SearchingPwnedPasswordsByRange) is
being used.

### Account Enumeration Defenses (work in progress)

:::warning

This feature is a work in progress and is tracked as
[kratos#133](https://github.com/ory/kratos/issues/133).

It does not yet work as documented!

:::

Account enumeration attacks allow a attacker to find out who is signed up. This
compromises the privacy of your users and can hurt reputation depending on the
service (e.g. "adult content").

This attack usually makes only sense if an email address or a phone number is
collected during registration. For chosen usernames, this attack is much more
difficult, as the attacker has to know what usernames the victim is using.

There are three common ways an attacker can figure out if a user is signed up at
a service:

- During login: "No user with this email address was found"
- During registration: "A user with this email address exists already"
- During password reset: "No user with this email address was found"

To mitigate this attack, the following strategies need to be deployed:

- The login form should return the same message regardless of whether the
  password is wrong or the email/username does not exist: "The provided
  credentials are invalid."
- The password reset form should always return a success message and send out an
  email. If the email address is registered, a normal password reset email is
  sent. If the email address is not registered, an email is sent to the address
  indicating that no account is set up with that email address. This is helpful
  to users that have multiple email addresses and are using the wrong email
  address for the password reset.
- The registration form should also always return a success message and send out
  an email. If the email address is not yet registered, a regular "account
  activation" email is sent out. If the email address is registered already, a
  email is sent out telling the user that the account is already set up, and
  link to the login screen.

If you wish to mitigate account enumeration attacks, it is important to note
that you can not sign in users directly after sign up! Depending on the type of
service you provide, you might not care about this specific attack in which case
direct login after sign up would be ok.

#### Enabling Account Enumeration Defenses

Assuming you wish to enable account enumeration defenses, you need to configure
ORY Kratos as follows:

- Collect one or more email addresses during sign up and enable email
  verification.
- **Do not** enable the `session` post-registration workflow. Use only the
  `redirect` post-registration workflow.

```yaml
selfservice:
  registration:
    after:
      password:
        # !! DO NOT enable `session` or all registration processes will fail!!
        # - hook: session

        # You **must** enable identifier verification or no email will be sent and the registration is thus just a blank
        # entry in the database with no way of logging in.
        - hook: verify
```

#### Disable Account Enumeration Defenses

Enforcing email verification, which requires an email round trip and disrupts
the sign up process, is not always feasible. In these cases, you might want to
disable account enumeration defenses.

You can disable the defense mechanism on a per-field basis in your Identity
Traits Schema:

```json title="path/to/my/identity.traits.schema.json"
{
  $id': 'https://example.com/identity.traits.schema.json',
  $schema': 'http://json-schema.org/draft-07/schema#',
  title': 'Person',
  type': 'object',
  properties':
    {
      'username':
        {
          'type': 'string',
          'ory.sh/kratos':
            {
              'credentials':
                {
                  'password':
                    {
                      'identifier': true,
                      'disable_account_enumeration_defenses': true,
                    },
                },
            },
        },
    },
}
```

This will tell ORY Kratos to display messages such as "a user with this email
address exists already" and "no user with this email address is registered on
this site". You can then enable the `session` post-registration workflow:

```yaml
selfservice:
  registration:
    after:
      password:
        - hook: session
        # You can optionally enable verification of the provided email address(es) or phone number(s)
        # - hook: verify
```
