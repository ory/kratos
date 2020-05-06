---
id: openid-connect-social-sign-in-oauth2
title: Social Sign In with OpenID Connect and OAuth2
---

:::tip

Have you read the
[OpenID Connect / OAuth2 Documentation](../../concepts/credentials/openid-connect-oidc-oauth2.mdx)
yet? If not, this would be a perfect opportunity!

:::

The Social Sign In Strategy enables you to use

- [GitHub](http://github.com/);
- [Apple](https://developer.apple.com/sign-in-with-apple/);
- [Google](https://developers.google.com/identity/sign-in/web/sign-in);
- [Facebook](https://developers.facebook.com/docs/facebook-login/);
- [ORY Hydra](https://www.ory.sh/hydra);
- [Keycloak](https://www.keycloak.org);
- and every other OpenID Connect Certified Provider

as the Identity Provider.

Because of the nature of this flow (a browser is required) it does not work
API-only flows.

## Browser Clients

### Login

Sign In only works when an identity exists for that profile already. If it does
not exist, a registration flow will be performed instead.

### Registration

Sign Up on conflict with existing primary identifiers like email:

- Sign Up is dis-allowed and the user is asked to instead log in and then link
  his/her account instead.

### Settings

A user may link and unlink social profiles. Unlinking is only allowed if at
least one other sign in method is enabled.

## API Clients

API-based login and registration using this strategy will be addressed in a
future release of ORY Kratos.

Please be aware that OpenID Connect providers always require a Browser, with the
exception of "Sign in with Apple" on recent iOS versions.

## Security and Defenses

### Account Enumeration

:::warn

This feature is a work in progress and is tracked as
[kratos#133](https://github.com/ory/kratos/issues/133).

:::

Scenario: In some cases you might want to use the email address returned by e.g.
the GitHub profile to be added to your user's account. You might also want to use
it as an email + password identifier so that the user is able to log in
with a password as well. An attacker is able to exploit that by creating a social
profile on another site (e.g. Google) and use the victims email address to set it up
(this only works when the victim does not yet have an account with that email at Google).
The attacker can then use that "spoofed" social profile to try and sign up with your ORY Kratos
installation. Because the victim's email address is already known to ORY Kratos, the
attacker might be able to observe the behavior (e.g. seeing an error page) and come
to the conclusion that the victim already has an account at the website.

Mitigation: This attack surface is rather small and requires a lot of effort, including
forging an e.g. Google profile, and can fail at several stages. To completely mitigate
any chance of that happening, only accept email addresses that are marked as verified
in your Jsonnet code:

```json title="contrib/quickstart/kratos/email-password/oidc.github.jsonnet"
local claims = {
  email_verified: false
} + std.extVar('claims');

{
  identity: {
    traits: {
      // Allowing unverified email addresses enables account
      // enumeration attacks, especially if the value is used for
      // e.g. verification or as a password login identifier.
      //
      // Therefore we only return the email if it (a) exists and (b) is marked verified
      // by GitHub.
      [if "email_primary" in claims && claims.email_verified then "email" else null]: claims.email_primary,
    },
  },
}
```
