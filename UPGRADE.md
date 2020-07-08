# Upgrading

The intent of this document is to make migration of breaking changes as easy as
possible. Please note that not all breaking changes might be included here.
Please check the [CHANGELOG.md](./CHANGELOG.md) for a full list of changes
before finalizing the upgrade process.

## unreleased

These changes have not yet been released and this area's purpose is to keep
track of future changes.

## v0.4.4-alpha.1

Please head over to the [CHANGELOG](https://github.com/ory/kratos/blob/master/CHANGELOG.md#040-alpha1-2020-07-08)

## v0.3.0-alpha.1

This release finalizes the OpenID Connect and OAuth2 login, registration, and settings strategy with JsonNet data transformation! From now on, "Sign in with Google, Github, ..." is officially supported! It's also possible to link and unlink these connections using the Self-Service Settings Flow! The documentation has been updated to reflect those changes and includes guides to setting up "Sign in with GitHub" in under 5 Minutes! Please be aware that existing OpenID Connect connections will stop working. Check out the "Breaking Changes" section for more info! Want to learn more? Check [out the docs](https://www.ory.sh/kratos/docs/concepts/credentials/openid-connect-oidc-oauth2)!

We changed the config validation output, making it easier than ever to find bugs in your config:

```
% kratos --config invalid-config.yml serve
INFO[0001] Config file loaded successfully.              path=invalid-config.yml
ERRO[0001] The provided configuration is invalid and could not be loaded. Check the output below to understand why.  config_file=invalid-config.yml

dsn: <nil>
     ^-- one or more required properties are missing

urls.whitelisted_return_to_urls: https://selfservice.office.example.com
                                 ^-- expected array, but got string

FATA[0001] The services failed to start because the configuration is invalid. Check the output above for more details.
```

This release concludes over 50 commits and 16.000 lines of code changed.

## v0.2.0-alpha.2

This is a heavy release with over hundreds of commits and files changed! Let's
take a look at some of the highlights!

### ORY Oathkeeper now optional

Using ORY Oathkeeper to protect your API is now optional. The basic quickstart
now uses a much simpler set up. Go
[check it out](https://www.ory.sh/kratos/docs/quickstart) now!

### PostgreSQL, MySQL, CockroachDB support now tested and official!

All three databases now pass acceptance tests and are thus officially supported!

### Self-Service Profile Flow

The self-service profile flow has been refactored into a more generic flow
allowing users to make modifications to their traits and credentials. Check out
the [docs to learn
more]([here](https://www.ory.sh/kratos/docs/self-service/flows/user-settings-profile-management)
about the flow and it's features.

Please keep in mind that the flow's APIs have changed. We recommend re-reading
the docs!

### Managing Privileged Profile Fields

Flows such as changing ones profile or primary email address should not be
possible unless the login session is fresh. This prevents your colleague or evil
friend to take over your account while you make yourself a coffee.

ORY Kratos now supports this by redirecting the user to the login screen if
changes to sensitive fields are made. The changes will only be applied after
successful reauthentication.

### Changes to Hooks

This patch focuses on refactoring how self-service flows terminate and changes
how hooks behave and when they are executed.

Before this patch, it was not clear whether hooks run before or after an
identity is persisted. This caused problems with multiple writes on the HTTP
ResponseWriter and other bugs.

This patch removes certain hooks from after login, registration, and profile
flows. Per default, these flows now respond with an appropriate payload (
redirect for browsers, JSON for API clients) and deprecate the `redirect` hook.
This patch includes documentation which explains how these hooks work now.

Additionally, the documentation was updated. Especially the sections about hooks
have been refactored. The login and user registration docs have been updated to
reflect the latest changes as well.

BREAKING CHANGE: Please remove the `redirect` hook from both login,
registration, and settings after configuration. Please remove the `session` hook
from your login after configuration. Hooks have moved down a level and are now
configured at `selfservice.<login|registration|settings>.<after|before>.hooks`
instead of `selfservice.<login|registration|settings>.<after|before>`.
Hooks are now identified by `hook:` instead of `job:`. Please rename those
sections accordingly.

We recommend re-reading the
[Hooks Documentation](https://www.ory.sh/kratos/docs/self-service/hooks/index).

### Changing Passwords

It's now possible to change your password using the Self-Service Settings Flow!
Lean more about this flow
[here](https://www.ory.sh/kratos/docs/self-service/flows/user-settings-profile-management)

### End-To-End Tests

We added tons of end-to-end and integration tests to find and fix pesky bugs.
