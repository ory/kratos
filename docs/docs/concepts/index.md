---
id: index
title: Overview
---

ORY Kratos is a new software archetype **Identity Infrastructure Service**.
Traditional identity systems - sometimes referred to as Identity and Access
Management (IAM), Identity Management (IdM), Identity Provider (IP/IdP), or
Identity as a Service (IDaaS) - have shortcomings that highlight the main
differences between ORY Kratos and other systems.

ORY Kratos solves identity on the network. It is not an on-device, for instance
mobile phones, user database. In ORY Kratos there is always an exchange of
credentials. In the case of web applications and identity:

- Username + Password -> Cookie, Token, ...
- Email + Password -> Cookie, Token, ...
- Passwordless login -> Cookie, Token, ...

Even for alternative use cases for example mobile, browser, or native
application there is either a cookie, which accesses the application directly
through the browser, or a token that accesses the application using a
programmatic client via an API. While it is nowadays common - but bad practice -
to use tokens for "Single Page Apps" or apps running on the client-side browser,
there is no real difference between these two approaches as both represent a set
of credentials valid for a certain domain or a number of domains.

Still, many identity systems primarily rely on OAuth2 and OpenID Connect. The
reasons for this would perhaps include the following:

- Assumption - it is certifiable;
- Assumption - it offloads complexity to developers who need to interact with
  and figure out e.g. `AppAuth`, `PassportJS`, and similar OAuth2 and OpenID
  Connect SDKs developed by open source communities;
- Assumption - selling complexity as security;
- Assumption - the nature of closed source does not allow for new, open and
  de-facto standards to emerge and instead uses a consenus driven feature set,
  even if it doesn't fit the use case 100%.

While ORY Kratos is currently not certifiable, it tackles these topics as
follows:

- With inspiration from the approach taken in the Kubernetes Project, ORY
  provides an open source project that hopefully becomes an open standard in the
  future.
- Prioritise simplicity and ease of use for developers. ORY Kratos integrates
  critical security components without relying on complex flows and protocols.

Using ORY Kratos it is possible to consume OAuth2 and OpenID Connect, and/or
create an OAuth2 and OpenID Connect Provider by combining ORY Kratos with
[ORY Hydra](http://github.com/ory/hydra) .

With a primary developer audience, ORY designs, secures, and tests critical
network flows, system architectures, user flows, protocols, and business logic.

## Today's Landscape

Let's take a look at different approaches and software systems today.

### Full-stack Identity and Access Management (IAM)

_Disclaimer: There are neither product nor project names in this section. This
section's goal is to describe circumstances and problems that mostly stem from
the community's experience. For information purposes, there is a preliminary
comparision of ORY Kratos and other projects and products available at
[Comparison](../further-reading/comparison.md)_.

Full-stack IAM is usually sold as a one-size-fits-all solution. Due to size and
complexity, these full-stack solutions are typically written in an enterprise
class programming language such as Java EE. The full-stack products have rich
feature sets that include:

- theming to customize the user experience, and to constrain the anticipated
  theming use case;
- HTML Template Engines specific to the language used, such as Java Server Pages
  or
  [Apache FreeMarker™](https://www.keycloak.org/docs/latest/server_development/#html-templates);
- plugin loaders and APIs to add custom logic or even custom API endpoints,
  specific to the language used by the project; and
- features such as integrated Load Balancers, Service Discovery, and other
  features designed prior to today's mature cloud architectures.

Full-stack software projects come with some overhead:

- the software has a large disk, CPU, and memory footprint;
- while scaling and clustering for High Availability is possible, it is complex
  since inter-process-communication for caches and other features is required
  for example using protocols such as [JGroups](http://jgroups.org/); and
- starting off with pre-defined use cases is easy, yet customization and
  application specific features require much more work.

Most full-stack projects we've seen are in-house solutions for IAM problems.
Imagine Google releasing their IAM as an open source product. It's certainly
great, and it covers a lot of ground, but it also comes with drawbacks:

- Strict data models specific to the company developing the product:
- Inflexible login process with either a username or an email for login, but not
  both or unable to change it later;
- Any application specific additional attributes are stored as unstructured
  data, sometimes even as plain key/value pairs;
- Complex build pipelines when using modern frontend frameworks like React or
  Angular in the HTML Rendering engine;
- The user model stays the same, even when differentiating between customers and
  employees in your system; and
- API consumption is usually an after-thought because most flows are built
  around the user doing something in the browser. All of the above leads to
  added complexity in application development and deployment due to session
  management, cookie management, CSRF protection, and other mechanisms related
  to identity and security.

### Identity as a Service (IDaaS)

In today's market, with many proprietary SAAS companies offering Identity as a
Service, it seems easy to make sign-on cumbersome for both developers and users.
Even with delegated third party login processes such as "Login with Google,"
where OAuth2 and OpenID Connect are often the primary protocols, the challenge
is making a secure and simple login without any extra overhead, for instance,
with Oauth2 and OpenID.

ORY's focus is on simplicity, user experience, and above all, using the right
tools and technologies for the target application. Feedback from ORY's user
community as well as the open source development efforts involved in ORY Hydra-
OAuth2 and OpenID Connect server [ORY Hydra](https://github.com/ory/hydra), show
that implementing OAuth2 or OpenID Connect is often frustrating and too complex.
These technologies are not one size fits all, and not designed for every
implementation scenario.

The main point is that OAuth2 and OpenID Connect protocols solve identity
federation. For instance, when the target application authenticates an identity
or authorizes access by using a system outside of the application's control, for
example within an enterprise, company domain or another third party service.
These protocols don't solve processes like updating a user's profile, adding a
secondary recovery email, solving 2FA, storing and managing sessions, or solving
global logout. These processes are the developer's responsibility and while the
OAuth2 and OpenID Connect protocols offer a way to securely solve identity
federation over the browser, they do not solve:

- Storage and management of all these tokens. These would often end up in the
  localStore making them vulnerable to XSS attacks. Or in a cookie issued by an
  HTTP server; here an extra function would need to be developed.
- Managing and storing the user session. This would require the developer to
  create the cookie for the session, delete the cookie at log out, and to make
  sure that the cookie implements best security practices.
- Refreshing expiring tokens. While this function is straightforward for one
  request, synchronising for example fifteen concurrent requests can be
  difficult.

The point is that OAuth2 is hard to use because the intended use cases (for
instance "Facebook Photo Backups") are very specific, and the proper security
mechanisms need to be in place to deal with malicious third parties. ORY Kratos
simplifies user login.

### Use a library

Due to the multitude of programming languages across front end, backend and core
infrastructure, ORY Kratos embraced a polyglot design. In other words ORY Kratos
provides Software Development Kits (SDK) and libraries for the main programming
languages.

As software development teams grow, product requirements change. During a
development lifecycle different parts, for instance humans, servers, and code,
need to scale to size. Over time, the original product splits into smaller, more
manageable chunks. The libraries chosen at the outset will need to run on the
newest version of the programming language. Sometimes even the programming
language or implementation framework are subject to change as a whole. Still,
the identity system is often so interlinked with all of the source code,
middlewares, annotations, shared state, etc., that it becomes an absolute
nightmare to decouple.

### Roll your own

Starting from scratch is sometimes the only option to fulfill the product's
architectural requirements. In this case, the following are some of the main
considerations and challenges in ORY's efforts:

- To manage changing user models;
- To choose and use some encryption algorithms such as BCrypt, PBKDF2, Scrypt,
  Argon2, Argon2i, Argon2id, Argon2d. There are many alternatives and most of
  the algorithms have numerous versions and parametrization options for example
  "Salt length parameter." The multitude of options makes it hard to configure
  the most secure setting;
- To consider and implement a wide range of processes and process variants. For
  instance, the user signs up via email, and later uses "Sign up with Google"
  that has the same email address. Or, the user previously signed up using the
  email/password process or flow, and then signs up using "Sign up with Google"
  and vice versa. Even these fairly simple use cases warrant considerable
  development effort to correctly implement with some degree of user
  friendliness;
- To prevent security threats for example
  [account enumeration attacks](<https://wiki.owasp.org/index.php/Testing_for_User_Enumeration_and_Guessable_User_Account_(OWASP-AT-002)>);
- To implement two factor authentication (2FA). When the user loses access to a
  registered and approved device, there should be a fallback phone number for
  SMS or offline security codes;
- To understand and use all of the important rules such as activation, welcome,
  etc., around sending emails that don't alert spam filters;
- To support a broad ecosystem of products and services. For example in the
  event that a target application needs to notify Stripe when a new customer
  signs up.

The list above is purposely kept short. There are very many things to consider
when building Kratos in concert with the other products ORY Keto, ORY Hydra, and
ORY Oathkeeper. The list is really, really long.

## Introducing ORY Kratos

Considering all of the information above, there would be enough context to
understand why and how ORY Kratos started, and why it's different from other
more conventional approaches. ORY Kratos uses a new stack, is open source, and
peer reviewed and developed in a broad community of experts.

### Solving a specific problem domain

ORY Kratos addresses a clearly defined problem domain:

- managing credentials such as passwords, recovery email addresses, and security
  questions;
- authentication including secure login, keeping track of sessions and devices;
- selfservice account management for example update profile, add/update email
  addresses, and changing passwords;
- account/identity administration such as create, read, delete, update, import,
  and get; and
- managing identity data for example first name, last name, profile picture, and
  birthday, etc.

ORY has numerous products that support the protocols OAuth2 or OpenID Connect in
[ORY Hydra](http://github.com/ory/hydra), a permission system in
[ORY Keto](http://github.com/ory/keto), and a Reverse Proxy in
[ORY Oathkeeper](http://github.com/ory/oathkeeper).

### Software Architecture

ORY's
[Software Architecture and Philosophy](https://www.ory.sh/docs/ecosystem/software-architecture-philosophy)
document, explains the architectural beliefs and framework behind the ORY
Products in particular:

- Small runtime footprint with an about five (5) MB binary running on all
  operating systems without any system, library, or VM dependencies;
- Fully virtualised in a fifteen (15) MB Docker image;
- Easy to manage with exactly one binary for the server and the cli;
- Run-time orchestration using the latest Kubernetes providing fast and easy to
  use [Helm charts](https://github.com/ory/k8s);
- Horizontal scaling with no etcd key value store or memcached or adjacent tool
  required.

### Bring your own User Interface (Framework)

ORY's approach to user interface and user experience is to provide for an
interaction concept with maximum flexibility and creativity. Some companies need
[progressive profiling](https://blog.hubspot.com/blog/tabid/6307/bid/34155/how-to-capture-more-and-better-lead-intel-with-progressive-profiling.aspx)
and build a NodeJS app. Other companies desire to capture everything in one go,
using Client-Side JavaScript library such as Angular or React. Some companies
want an iOS-native registration and login experience. While ORY's cloud native
headless API approach address many integration and UI issues, with ORY Kratos,
predefined flows make it easy to implement a custom user interface for login,
registration, profile management, account reset, etc. Furthermore, to make it
very easy to get started there is a reference implementation
[github.com/ory/kratos-selfservice-ui-node](https://github.com/ory/kratos-selfservice-ui-node).

For more details about each individual flow, consult the
[Self-Service Flows Chapter](../self-service.mdx).

### Bring your own Identity Model(s)

Sometimes it is necessary to store more than one type of identity in your
system:

- A customer that uses email + password to login, and needs to set a birthdate;
  or
- An employee that uses a unique username + password to login with a cost center
  attached to the profile.

ORY Kratos implements both scenarios by using
[JSON Schemas for Identities](./identity-data-model.mdx)

### Forget passport-js, oidc-client, ...

While proprietary and bespoke middleware can protect APIs and Web endpoints, ORY
Open Source provides a base solution for many use cases. For example, ORY Kratos
integrates with ORY Oathkeeper, a Reverse Proxy solution. Defining Access Rules
is as easy as writing a few lines of JSON / JSON5 / YAML!

Please consult the [Quickstart documentation](../quickstart.mdx), for further
information.
