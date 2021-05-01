---
id: integration-with-other-systems-using-web-hooks
title: Integration using Web-Hooks
---

Ory Kratos supports integration with other systems (e.g. ERP, Marketing, etc)
using Web-Hooks.

If you for example would like to send a marketing email using
[Mailchimp](https://mailchimp.com) upon a user registration, you can do so
easily:

```yaml title="path/to/my/kratos.config.yml"
selfservice:
  flows:
    registration:
      after:
        password:
          hooks:
            - hook: web-hook
              config:
                url: https://mandrillapp.com/api/1.0/messages/send
                method: POST
                body: /path/to/my/mandrillapp.jsonnet
```

```jsonnet title="/path/to/my/mandrillapp.jsonnet"
function(ctx) {
  key: "<Your-Api-Key>",
  message: {
    from_email: "hello@example.com",
    subject: "Hello from Ory Kratos",
    text: "Welcome to Ory Kratos",
    to: [
      {
        email: ctx.session.identity.verifiable_addresses[0].value,
        type: "to"
      }
    ]
  }
}
```
