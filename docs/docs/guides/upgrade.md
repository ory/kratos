---
id: upgrade
title: Applying Upgrades
---

This guide covers basics to consider when upgrading Ory Kratos, please also
visit [CHANGELOG.md](https://github.com/ory/kratos/blob/master/CHANGELOG.md) for
more technical details.

Generally upgrading Ory Kratos can be split into three main steps:

- Make a backup.
- [Install](/install.md) the new version (depending on how you run Ory Kratos).
- Run [`kratos migrate sql`](../cli/kratos-migrate-sql.md) to run the
  appropriate SQL queries.

Ory Kratos will run the `migrate sql` command for all versions. For example when
upgrading from v0.6 to v0.8, the migrations will be run for v0.6 to v0.7 and
then for v0.7 to v0.8. As such upgrading from any version to the latest version
directly is possible. Should you run into problems with a direct upgrade,
consider a stepped upgrade and please visit the community
[chat](https://slack.ory.sh/) or
[discussions](https://github.com/ory/kratos/discussions).

:::warning

Back up your data! Applying upgrades can lead to data loss if handled
incorrectly.

:::

### Upgrading Tips

We recommend taking the following steps to ensure that no data is lost:

> Please note: These are recommendations and should be used in accordance with
> other Ops best practices. The steps required for a smooth and secure upgrade
> process may vary with different setups, tech stacks, and environments.

1. Write down a high-level upgrade plan
   - Who will perform the upgrade?
   - How will the upgrade be performed?
   - What components are affected?
2. Devise roll-out plan
   - When will the upgrade be performed?
   - Will there be an outage?
   - How long will it be?
   - What is your rollback plan?
3. Back up everything!
4. Run a trial upgrade on a local environment.
5. Run an upgrade on a staging environment.
6. Perform tests on staging & prepare production environment.
7. Run the upgrade on production.

### Breaking changes overview

- **[Ory Kratos v0.8 Breaking changes](https://github.com/ory/kratos/blob/v0.8.0-alpha.1/CHANGELOG.md#breaking-changes)**
- **[Ory Kratos v0.7 Breaking changes](https://github.com/ory/kratos/blob/v0.7.0-alpha.1/CHANGELOG.md#breaking-changes)**
- **[Ory Kratos v0.6 Breaking changes](https://github.com/ory/kratos/blob/v0.6.0-alpha.1/CHANGELOG.md#breaking-changes)**
- **[Ory Kratos v0.5 Breaking changes](https://github.com/ory/kratos/blob/v0.5.0-alpha.1/CHANGELOG.md#breaking-changes)**

:::note

Skip the hassle of applying upgrades to Ory Kratos? Take a look at
[Ory Cloud](https://www.ory.sh/docs).

:::
