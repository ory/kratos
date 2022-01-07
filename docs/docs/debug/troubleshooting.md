---
id: troubleshooting
title: Tips & Troubleshooting
---

:::info

If your question is not covered here, please open a
[discussion](https://github.com/ory/kratos/discussions) and ping
`@ory/documenters`

:::

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

### Can I freeze certain identity fields after registration?

> Also I want read-only or hidden user-specific data. Can I use traits for that?
> Example: Some of our identity trait fields may be the subject to identity
> validation
> ([know your customer](https://en.wikipedia.org/wiki/Know_your_customer)). Once
> they are validated we do not allow the identity to be edited by the end-user.

Currently traits are visible to the user and also editable by them. It makes
sense to have read-only and/or hidden traits for users and there is an open
issue for this feature: [Issue](https://github.com/ory/kratos/issues/47).

### How can I verify in advance if a username is available during registration?

You can not right now. It would allow account enumeration attacks. See also the
[section in the documentation](https://www.ory.sh/kratos/docs/concepts/security/#account-enumeration).

### Do have plans to support automatic user migration scenarios?

> E.g. configure a callback to the legacy system when you cannot find the
> corresponding user, and store the identity on successful legacy system
> response.

No plans yet, but migration is an important piece of the puzzle. We will focus
on importing existing users first. An alternative to callback and custom code is
fronting the legacy system with Ory Hydra (OAuth2/OIDC Server) and then using
that as an upstream in Ory Kratos.

### Do you have protection for brute force attacks i.e. repetitive login attempts? Leaked database scans?

See the following document:
[Ory Kratos Security Measures](https://www.ory.sh/kratos/docs/concepts/security#passwords)

### How do I append a random suffix e.g. a number to OIDC usernames if the username returned by a provider is already taken?

You can use the md5 function built in the
[jsonnet standart library](https://jsonnet.org/ref/stdlib.html). See the
following issue as well:
[Add random function to the jsonnet library](https://github.com/ory/kratos/issues/988).

### I want to implement a single-page-app (SPA). Is this possible with Ory Kratos?

Yes, definitely. Check out our
[in-depth tutorial](https://www.ory.sh/login-spa-react-nextjs-authentication-example-api-open-source/)
and [example app](https://github.com/ory/kratos-nextjs-react-example).

### Is the code audited by an independent entity and is there a bug bounty program?

We will do an audit when the APIs and core are stable, so when 1.0 is released .
A bug bounty program is in the making but not done yet. If you are a security
researcher and interested in working on Kratos, please reach out to
[security@ory.sh](mailto:security@ory.sh).

### Is it possible to login through social providers if the identifiers match even if user did not signup through the provider?

> Right now there's an explicit extra step to link them as default
> configuration. Is there a way to login through them without linking if e.g.
> emails match. There are security considerations to think about, but it's a
> flow that's currently being used; e.g Atlassian.

We strongly discourage this practice, consider the following scenario:

> You have signed up using david@company.org to myapp.com. I, the hacker, know
> this. I create an google account with email david@company.org and sign in. Now
> I am in your account.

### Skip Social Login Link-Up

> I want users to register either through email-password or via Github social
> login. But in my configuration when the user registers through the
> email/password strategy then after the email verification, Kratos is
> redirecting the user to Github to link the Github account. I want this flow
> optional. How can I do that?

Add the session hook after password/social:

```
      after:
        password:
          hooks:
            - hook: session
```

### Social Login: Is the OIDC token saved in the Ory Kratos database?

The OIDC token is not being saved in the Ory Kratos database, Ory Kratos
exchanges it for a session. For example: Ory Kratos starts the OIDC flow with
Google. Google sends back the OIDC `id_token` (if successfully authenticated).
Ory Kratos gets the user information from the `id_token` (specified in claims,
default email or username). Then Ory Kratos checks the database if the identity
exists; if yes, Ory Kratos returns an session token to the user; if no, the
identity is created and it returns a session token. The `id_token` from Google
is discarded after completion of the flow.

### Why are both plain text and HTML templates required?

The courier uses them as
[alternatives](https://github.com/ory/kratos/blob/master/courier/courier.go#L205)
for fallback.

```
// AddAlternative adds an alternative part to the message.
//
// It is commonly used to send HTML emails that default to the plain text
// version for backward compatibility. AddAlternative appends the new part to
// the end of the message. So the plain text part should be added before the
// HTML part. See http://en.wikipedia.org/wiki/MIME#Alternative
```
