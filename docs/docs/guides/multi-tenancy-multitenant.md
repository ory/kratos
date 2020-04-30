---
id: multi-tenancy-multitenant
title: Multitenancy
---

ORY Kratos does not implement multi-tenancy in its application logic, but it is possible to implement multi-tenancy
with ORY Kratos!

The recommended approach is to run one or more (depending on your scale) SQL databases and create one database schema
per tenant in these database instances. So one PostgreSQL database instance could, for example, host 15000 tenants
who each have access to one schema.

ORY Kratos itself should run as one instance per tenant with a configuration tailored for that specific tenant. The
minimum required change is using different secrets and the tenant's
DSN (`postgresql://user:pass@.../tenant-123`). Because ORY Kratos is very lightweight, the deployment overhead becomes negligible.

While deployment complexity increases - but is addressable with e.g. Kubernetes - this approach has several advantages:

- Absolute isolation of tenants which implies: better security, better privacy, more control.
- Easy sharding and partitioning because database schemas isolate tenants.
- No complexity in ORY Kratos' business logic and security defenses.
