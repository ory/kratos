---
id: troubleshooting
title: Tips & Troubleshooting
---

:::info

Please add your troubleshooting tricks and other tips to this document, You can
either open a [discussion](https://github.com/ory/kratos/discussions) and ping
`@ory/documenters` or
[edit the page directly](https://github.com/ory/kratos/edit/master/docs/docs/debug/troubleshooting.md).

:::

### `400: Bad Request` on self-service flows

Make sure you are starting and finishing the request in one browser.
Self-service browser flows need to be executed in the same browser from start to
finish!  
Starting the flow in e.g. Safari and completing it in Chrome won't work. API
Clients like Electron, Postman or Insomnia are browsers themselves, which can
cause requests to fail. For testing purposes [cURL](https://curl.se/) is a good
choice.

### How can I separate customers/employee data, but have them use the same login dialog?

> We want to separate our customers and employees, so we store them in different
> databases. But we would like to have them use the same login dialog for our
> portal.

You can deploy Ory Kratos two times, and use the same login UI pointing to two
different Kratos login endpoints - `/login/customer` or `/login/employee`,
either by having two different login routes, or by adding some logic to your
login UI that reroutes customers to `/login/customer` and employees to
`/login/employee`. So you define the same login or registration UI URLs in both
of the Kratos configurations. You may need to tell your login/registration UI
which Kratos it is supposed to talk to. The instances are cheap to deploy and
the databases are completely isolated from each other. For example something
like `/login/customer` and `/login/employee`.

### How can I verify in advance if a username is available during registration?

You can not right now. It would allow account enumeration attacks. See also the
[section in the documentation](https://www.ory.sh/kratos/docs/concepts/security/#account-enumeration).

### Do have plans to support automatic user migration scenarios?

> E.g. configure a callback to the legacy system when you cannot find the
> corresponding user, and store the identity on successful legacy system
> response.

An alternative to callback and custom code is fronting the legacy system with
Ory Hydra (OAuth2/OIDC Server) and then using that as an upstream in Ory Kratos.
