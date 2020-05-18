---
id: index
title: Introduction
---

ORY Kratos is an API-first Identity and User Management system that is built
according to
[cloud architecture best practices](https://www.ory.sh/docs/ecosystem/software-architecture-philosophy/).
It implements core use cases that almost every software application needs to
deal with:

- **Self-service Login and Registration**: Allow end-users to create and sign
  into accounts (we call them **identities**) using Username / Email and
  password combinations, Social Sign In ("Sign in with Google, GitHub"),
  Passwordless flows, and others.
- **Multi-Factor Authentication (MFA/2FA)**: Support protocols such as TOTP
  ([RFC 6238](https://tools.ietf.org/html/rfc6238) and
  [IETF RFC 4226](https://tools.ietf.org/html/rfc4226) - better known as
  [Google Authenticator](https://en.wikipedia.org/wiki/Google_Authenticator))
- **Account Verification**: Verify that an E-Mail address, phone number, or
  physical address actually belong to that identity.
- **Account Recovery**: Recover access using "Forgot Password" flows, Security
  Codes (in case of MKFA device loss), and others.
- **Profile and Account Management**: Update passwords, personal details, email
  addresses, linked social profiles using secure flows.
- **Admin APIs**: Import, update, delete identities.

Identity is a hard problem and ORY Kratos solves it in a unique way. We value
security, flexibility, and integration with cloud technology (such as
Kubernetes) most:

- ORY Kratos does not ship a HTML Rendering Engine. You can build your own UI
  (or use our example UIs) in the language and framework you feel most
  comfortable with!
- The workflow engine allows you to fully customize your use experience: Whether
  your users need to activate their account after registration, or a multi-step
  (progressive) registration - it's all possible!
- One Identity Data Model does not fit all - you might have customers that need
  a billing address, internal support staff that is assigned to a cost center,
  and that smart fridge on floor 4. You can express the different data models
  using [JSON Schema](https://json-schema.org/) and make the system work for
  you - not the other way around!

To learn more about what's different and how ORY Kratos compares to other (open
source) solutions, head over to [Concepts](./concepts/index.md) and
[Comparison](./further-reading/comparison.md).

But it's not just features - ORY Kratos stands out because it runs on any
operating system (Linux, macOS, Windows) and on most processors (i386, amd64,
arm, ...). The compiled binary has _no system or library or file dependencies_
and can be run as a single, static binary on top of a, for example, raw Linux
kernel. The binary and Docker Image is less than 20MB small.

ORY Kratos scales horizontally without effort. The only external dependency is a
RDBMS - we currently support SQLite, PostgreSQL, MySQL, CockroachDB. You will
not need memcached, etcd, or any other system to scale ORY Kratos.

We believe in strong separation of concerns, it is a guiding principle in the
design of each ORY project. As such, we build software that solves specific
problems very well and offloads adjacent issues (such as a user interface) to
other applications.
