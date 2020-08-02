---
id: production
title: Going to Production
---

:::warn

This document is still in development.

:::

## Database

ORY Kratos requires a production-grade database such as PostgreSQL, MySQL,
CockroachDB. Do not use SQLite in production!

## Security

When preparing for production it is paramount to omit the `--dev` flag from
`kratos serve`.

### Admin API

Never expose the ORY Kratos Admin API to the internet unsecured. Always require
authorization. A good practice is to not expose the Admin API at all to the
public internet and use a Zero Trust Networking Architecture within your
intranet.

## Scaling

There are no additional requirements for scaling ORY Kratos, just spin up
another container!
