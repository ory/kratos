---
id: integration-with-other-systems-using-web-hooks
title: Integration using Web-Hooks
---

Ory Kratos supports integration with other systems (e.g. ERP, Marketing, etc)
using Web-Hooks.

### Mailchimp

To send marketing email using [Mailchimp](https://mailchimp.com) upon registration,
add the following to your Ory Kratos config:

```yaml title="path/to/my/kratos.config.yml"
selfservice:
  flows:
    registration:
      after:
        password:
          hooks:
            - hook: web_hook
              config:
                url: https://mandrillapp.com/api/1.0/messages/send
                method: POST
                body: /path/to/my/mandrillapp.jsonnet
```

And create a JsonNet file. Please be aware that Mailchimps' authentication
mechanism currently requires to save the Mailchimp key in the JsonNet. For
other systems you would be using the `web_hook`'s `auth` config.

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
