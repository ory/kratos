# Upgrading

The intent of this document is to make migration of breaking changes as easy as possible. Please note that not all
breaking changes might be included here. Please check the [CHANGELOG.md](./CHANGELOG.md) for a full list of changes
before finalizing the upgrade process.

## v0.2.0-alpha.1

This is a heavy release with over hundreds of commits and files changed! Let's take a look
at some of the highlights!

### ORY Oathkeeper now optional

Using ORY Oathkeeper to protect your API is now optional. The basic quickstart now uses a much simpler set up.
Go [check it out](https://www.ory.sh/kratos/docs/quickstart) now!

### PostgreSQL, MySQL, CockroachDB support now tested and official!

All three databases now pass acceptance tests and are thus officially supported!

### OpenID Connect and OAuth2 now official!

Using social profiles to sign up and log in is now as easy as setting two config entries! Check out

- [The OpenID Connect & OAuth2 Credential Documentation](https://www.ory.sh/kratos/docs/concepts/credentials/openid-connect-oidc-oauth2);
- [The Flow Description](https://www.ory.sh/kratos/docs/concepts/credentials/openid-connect-oidc-oauth2);
- [The "Sign in with GitHub" Guide](https://www.ory.sh/kratos/docs/guides/sign-in-with-github)!

### Self-Service Profile Flow

The self-service profile flow has been refactored into a more generic flow allowing users to make modifications
to their traits and credentials. Check out the
[docs to learn more]([here](https://www.ory.sh/kratos/docs/self-service/flows/user-settings-profile-management)
about the flow and it's features.

Please keep in mind that the flow's APIs have changed. We recommend re-reading the docs!

### Managing Privileged Profile Fields

Flows such as changing ones profile or primary email address should not be possible unless the login session is fresh.
This prevents your colleague or evil friend to take over your account while you make yourself a coffee.

ORY Kratos now supports this by redirecting the user to the login screen if changes to sensitive fields are made.
The changes will only be applied after successful reauthentication.

### Changes to Hooks

This patch refactors how self-service flows terminate and
changes how hooks behave and when they are executed.

Before this patch, it was not clear whether hooks run before or
after an identity is persisted. This caused problems with multiple
writes on the HTTP ResponseWriter and other bugs.

This patch removes certain hooks from after login, registration, and profile flows.
Per default, these flows now respond with an appropriate payload (
redirect for browsers, JSON for API clients) and deprecate
the `redirect` hook. This patch includes documentation which explains
how these hooks work now.

Additionally, the documentation was updated. Especially the sections
about hooks have been refactored. The login and user registration docs
have been updated to reflect the latest changes as well.

BREAKING CHANGE: Please remove the `redirect` hook from both login,
registration, and settings after configuration. Please remove
the `session` hook from your login after configuration. Hooks
have moved down a level and are now configured at
`selfservice.<login|registration|settings>.<after|before>.hooks`
instead of
`selfservice.<login|registration|settings>.<after|before>.hooks`.
Hooks are now identified by `hook:` instead of `job:`. Please
rename those sections accordingly.

We recommend re-reading the [Hooks Documentation](https://www.ory.sh/kratos/docs/self-service/hooks/index).

### Changing Passwords

It's now possible to change your password using the Self-Service Settings Flow! Lean more about this flow
[here](https://www.ory.sh/kratos/docs/self-service/flows/user-settings-profile-management)

### End-To-End Tests

We added tons of end-to-end and integration tests to find and fix pesky bugs.
