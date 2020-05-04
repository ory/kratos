---
id: openid-connect-social-sign-in-oauth2
title: Social Sign In with OpenID Connect and OAuth2
---

:::tip

Have you read the [OpenID Connect / OAuth2 Documentation](../../concepts/credentials/openid-connect-oidc-oauth2.mdx)
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
