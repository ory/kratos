<h1 align="center"><img src="./docs/images/banner_kratos.png" alt="ORY Kratos - Cloud native Identity and User Management"></h1>

<h4 align="center">
    <a href="https://discord.gg/PAMQWkr">Chat</a> |
    <a href="https://community.ory.sh/">Forums</a> |
    <a href="http://eepurl.com/di390P">Newsletter</a><br/><br/>
    <a href="https://www.ory.sh/docs/next/kratos/">Guide</a> |
    <a href="https://www.ory.sh/docs/next/kratos/sdk/api">API Docs</a> |
    <a href="https://godoc.org/github.com/ory/kratos">Code Docs</a><br/><br/>
    <a href="https://opencollective.com/ory">Support this project!</a>
</h4>

---

<p align="left">
    <a href="https://circleci.com/gh/ory/kratos/tree/master"><img src="https://circleci.com/gh/ory/kratos/tree/master.svg?style=shield" alt="Build Status"></a>
    <a href="https://coveralls.io/github/ory/kratos?branch=master"> <img src="https://coveralls.io/repos/ory/kratos/badge.svg?branch=master&service=github" alt="Coverage Status"></a>
    <a href="https://goreportcard.com/report/github.com/ory/kratos"><img src="https://goreportcard.com/badge/github.com/ory/kratos" alt="Go Report Card"></a>
    <a href="https://bestpractices.coreinfrastructure.org/projects/364"><img src="https://bestpractices.coreinfrastructure.org/projects/364/badge" alt="CII Best Practices"></a>
    <a href="#backers" alt="sponsors on Open Collective"><img src="https://opencollective.com/ory/backers/badge.svg" /></a> <a href="#sponsors" alt="Sponsors on Open Collective"><img src="https://opencollective.com/ory/sponsors/badge.svg" /></a>
</p>

ORY Kratos is the first and only cloud native Identity and User Management System in the world. Finally, it is no longer necessary to implement a User Login process for the umpteenth time! 

ORY Kratos includes:

- **User login and registration** using a variety of configurable authentication mechanisms: **Username/Email + Password**, **Social Sign In** ("Sign in with GitHub, Google, ..."), with or without password and others.
- **Multi-factor authentication** supporting a wide range of protocols such as [Google Authenticator](https://en.wikipedia.org/wiki/Google_Authenticator) (formalized as [RFC 6238](https://tools.ietf.org/html/rfc6238) and [IETF RFC 4226](https://tools.ietf.org/html/rfc4226)).
- **Account verification** and **account recovery** by several methods: E-Mail, recovery codes, ...
- **Storing user information** in a way that does not enforce *our* normative data model on *you*. With ORY Kratos developers define the dataset users may store using [JSON Schema](https://json-schema.org/). ORY Kratos supports multiple identity types - every identity can have its own JSON Schema - even versioned!
- **Headless UI** - instead of learning some custom, and probably inferior, template engine, developers can choose their own! ORY Kratos is all APIs. UI's can be programmed in various languages (JavaScript, Node, Java, PHP, ...) and numerous framework (React, Vue, Angular, ...)! Check out the [reference UI implementation](https://github.com/ory/kratos-selfservice-ui-node) - it's less than 100 lines of code!
- **Workflow engine** to decide what happens after, for example, a user signs up (redirect somewhere? require activation before login? issue session right away?) as well as to notify other systems on certain actions ("create a Stripe account after sign up", "synchronize with newsletter", ...).
- ... and of course many more features that will be discussed outside of the Introduction in the Ory Kratos documentation [here](https://www.ory.sh/docs/next/kratos).

## Timeline

ORY Kratos is **not yet released** and is **undergoing continuous and active development**. The core featureset is done but several more steps are required before version 0.0.1 can be released. To find out the current progress, planned features for each milestone, and more information please refer to [milestones](https://github.com/ory/kratos/milestones).

## What's different?

> This section is a work in progress.

- There is no templating as with other full-stack solutions. You implement a "login, registration, ... UI" that interacts with ORY Kratos. For instance:

   * _Want Progressive Registration?_       _No problem._ 

   * _Just need a username on sign up?_      _Sure!_ 

   * _How about your favorite pet name as a required sign up field?_      _Of course!_

- While other solutions support an API-driven approach, they leave you with the burden of making things secure, e.g. CSRF Tokens, storing state, and so on. In ORY Kratos, all of this is done using - among others - HTTP Redirection.
- ORY Kratos does not need OAuth2 and OpenID Connect. We know that big players in the market have tried selling you OAuth2 and OpenID Connect for years as "the most secure" and "a very easy to use" protocol. Fact is, OAuth2 and OpenID Connect are not designed for first-party use ("I just want people to be able to log into my mobile app"). ORY Kratos makes integration a one-minute process using a HTTP Reverse Proxy. [Reverse proxy] (https://en.wikipedia.org/wiki/Reverse_proxy) (https://httpd.apache.org/docs/2.4/howto/reverse_proxy.html) **Include links to docs here.** Ory Hydra is a OAuth2 Server and OpenID Certified™ OpenID Connect Provider written in Go - cloud native, security-first, open source API security for your infrastructure. If the goal is to become the new "Sign in with Google" provider, use ORY Hydra and natively integrate with ORY Kratos!
- The "Sign Up" and "Login" process is customisable. The developer determines next step after sign up and login, for instance:
"Redirect the user to a certain page?" 
"Create a Stripe account?" or
"Require account activation via email before being allowed to sign in?"

## Telemetry

Ory's services collect summarized, anonymized data that can optionally be turned off. Click
[here](https://www.ory.sh/docs/next/ecosystem/sqa) to learn more.

## Documentation

### Guide

The Guide is available [here](https://www.ory.sh/docs/next/kratos).

### HTTP API documentation

The HTTP API is documented [here](https://www.ory.sh/docs/next/kratos/sdk/api).

### Upgrading and Changelog

New releases might introduce breaking changes. To help you identify and incorporate those changes, we document these
changes in [UPGRADE.md](./UPGRADE.md) and [CHANGELOG.md](./CHANGELOG.md).

### Command line documentation

Run `kratos -h` or `kratos help`.
