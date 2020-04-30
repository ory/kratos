# Upgrading

The intent of this document is to make migration of breaking changes as easy as possible. Please note that not all
breaking changes might be included here. Please check the [CHANGELOG.md](./CHANGELOG.md) for a full list of changes
before finalizing the upgrade process.

## v0.2.0-alpha.1

This is a heavy release with over 150 commits and 430 files changed! Let's take a look
at some of the highlights!

### ORY Oathkeeper now optional

Using ORY Oathkeeper to protect your API is now optional. The basic quickstart now uses a much simpler set up.
Go [check it out](https://www.ory.sh/kratos/docs/quickstart) now!

### PostgreSQL, MySQL, CockroachDB support official!

All three databases now pass acceptance tests and are thus officially supported!

### Self-Service Profile Flow

The self-service profile flow has been refactored into a more generic flow which allows the user to make modification
to his/her traits and credentials. Check out the
[docs to learn more]([here](https://www.ory.sh/kratos/docs/self-service/flows/user-settings-profile-management)
about the flow and it's features.

Please keep in mind that the flow's APIs have changed!

### Managing Privileged Profile Fields

Flows such as changing ones profile or primary email address should not be possible unless the login session is fresh.
This prevents your colleague or evil friend to take over your account while you make yourself a coffee.

ORY Kratos now supports this by redirecting the user to the login screen if the changes are made to sensitive fields.
The changes will only be applied after successful reauthentication.

### Changing Passwords

It's now possible to change your password using the Self-Service Settings Flow! Lean more about this flow
[here](https://www.ory.sh/kratos/docs/self-service/flows/user-settings-profile-management)

### End-To-End Tests

We added tons of end-to-end and integration tests to find and fix pesky bugs.
