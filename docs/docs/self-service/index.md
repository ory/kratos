---
id: index
title: Before you start reading
---

ORY Kratos implements flows that users perform themselves as opposed to
administrative intervention. Facebook and Google both provide self-service
registration and profile management features as you are able to make changes to
your profile and sign up yourself.

Most believe that user management systems are easy to implement because picking
the right password hashing algorithm and sending an account verification code is
a solvable challenge. The real complexity however hides in the details and
attack vectors of self-service flows. Most data leaks happen because someone is
able to exploit

- registration: with attack vectors such as account enumeration (), ...;
- login: phishing, account enumeration, leaked password databases, brute-force,
  ...;
- user settings: account enumeration, account takeover, ...;
- account recovery: social engineering attacks, account takeover, spoofing, and
  so on.

ORY Kratos applies best practices established by experts (National Institute of
Sciences NIST, Internet Engineering Task Force IETF, Microsoft Research, Google
Research, Troy Hunt, ...) and implements the following flows:

- [Login and Registration](flows/user-login-user-registration.mdx)
- [Logout](flows/user-logout.md)
- [User Settings](flows/user-settings-profile-management.mdx)
- [Account Recovery](flows/password-reset-account-recovery.md)
- [Address Verification](flows/verify-email-account-activation.mdx)
- [User-Facing Error](flows/user-facing-errors.md)
- [2FA / MFA](flows/2fa-mfa-multi-factor-authentication.md)

Some flows break down into strategies which implement some of the flow's
business logic:

- The [Password Strategy](strategies/username-email-password.md) implements
  login and registration flows (with email/username and password), account
  recovery flows ("reset your password"), and user settings flows ("change your
  password").
- The
  [OpenID Connect Strategy](strategies/openid-connect-social-sign-in-oauth2.md)
  implements login and registration flows (with email/username and password),
  and user settings flows ("un/link another social account").
- The [Profile Strategy](strategies/profile.md) implement the settings flow
  ("change your first/last name, ...").

Some flows additionally implement the ability [to run hooks]() which allow users
to be immediately signed in after registration, notify another system on
successful registration (e.g. Mailchimp).
