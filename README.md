<h1 align="center"><img src="./docs/images/banner_kratos.png" alt="ORY Kratos - Cloud native Identity and User Management"></h1>

<h4 align="center">
    <a href="https://discord.gg/PAMQWkr">Chat</a> |
    <a href="https://community.ory.am/">Forums</a> |
    <a href="http://eepurl.com/di390P">Newsletter</a><br/><br/>
    <a href="https://www.ory.sh/docs/next/kratos/">Guide</a> |
    <a href="https://www.ory.sh/docs/next/kratos/sdk/api">API Docs</a> |
    <a href="https://godoc.org/github.com/ory/kratos">Code Docs</a><br/><br/>
    <a href="https://opencollective.com/ory">Support this project!</a>
</h4>

---

<p align="left">
    <a href="https://circleci.com/gh/ory/kratos/tree/master"><img src="https://circleci.com/gh/ory/hydra/tree/master.svg?style=shield" alt="Build Status"></a>
    <a href="https://coveralls.io/github/ory/kratos?branch=master"> <img src="https://coveralls.io/repos/ory/hydra/badge.svg?branch=master&service=github" alt="Coverage Status"></a>
    <a href="https://goreportcard.com/report/github.com/ory/kratos"><img src="https://goreportcard.com/badge/github.com/ory/hydra" alt="Go Report Card"></a>
    <a href="https://bestpractices.coreinfrastructure.org/projects/364"><img src="https://bestpractices.coreinfrastructure.org/projects/364/badge" alt="CII Best Practices"></a>
    <a href="#backers" alt="sponsors on Open Collective"><img src="https://opencollective.com/ory/backers/badge.svg" /></a> <a href="#sponsors" alt="Sponsors on Open Collective"><img src="https://opencollective.com/ory/sponsors/badge.svg" /></a>
</p>

ORY Kratos is the first and only cloud native Identity and User Management system in the world. The days where you would implement a User Login for the 10th time are finally over! ORY Kratos includes

- **user login and registration** using a variety of configurable authentication mechanisms: **Username/Email + Password**, **Social Sign In** ("Sign in with GitHub, Google, ..."), Passwordless and others.
- **multi-factor authentication** supporting a wide range of protocols such as [Google Authenticator](https://en.wikipedia.org/wiki/Google_Authenticator) (formalized as [RFC 6238](https://tools.ietf.org/html/rfc6238) and [IETF RFC 4226](https://tools.ietf.org/html/rfc4226)).
- **account verification** and **account recovery** by several means: E-Mail, Recovery Codes, ...
- **storing user information** in a way that does not enforce *our* data model on *you*, but allows you to define what data certain users may store using [JSON Schema](https://json-schema.org/). If you have more than one identity type no problem - every identity can have its own JSON Schema - even versioned!
- **headless UI** - instead of learning our custom (and probably not that great) template engine, just bring your own! ORY Kratos is all APIs and you can write your UI in the language (JavaScript, Node, Java, PHP, ...) and framework (React, Vue, Angular, ...) you like! Check out our [reference UI implementation](https://github.com/ory/kratos-selfservice-ui-node) - it's below 100 lines of code!
- **a workflow engine** to decide what happens after, for example, a user signs up (redirect somewhere? require activation before login? issue session right away?) as well as to notify other systems on certain actions (create a Stripe account after sign up, synchronize with newsletter, ...).
- ... and of course many more features that would blow the scope of this introduction.

## Timeline

ORY Kratos is **not yet released** and is **undergoing continuous and active development**. The core featureset is done but several more steps are required before version 0.0.1 can be released. To find out the current progress, planned features for each milestone, and more information please head over to [milestones](https://github.com/ory/kratos/milestones).

## What's different?

> This section is a work in progress.

- There is no templating as with other full-stack solutions. You implement a "login, registration, ... ui" which interacts with ORY Kratos. Want Progressive Registration? No problem. Just need a username on sign up? Sure! How about your favorite pet name as a required sign up field? Of course!
- While other solutions support an API-driven approach, they leave you with the burden of making things secure (e.g. CSRF Tokens), storing state, and so on. In ORY Kratos, all of this is done for you using - among others - HTTP Redirection.
- ORY Kratos does not need OAuth2 and OpenID Connect. We know that big players in the market have tried selling you OAuth2 and OpenID Connect for years as "the most secure" and "very easy to use" protocol. Fact is, OAuth2 and OpenID Connect are not designed for first-party use ("I just want people to be able to log into my mobile app"). ORY Kratos makes integration a one-minute process using a HTTP Reverse Proxy. **Include links to docs here.** If you want OAuth2 (you want to become the new "Sign in with Google" provider), we have ORY Hydra that integrates natively with ORY Kratos!
- You decide what happens after sign up and login (each customizable on its own): Redirect the user to a certain page? Create a Stripe account? Require account activation via email before being allowed to sign in?

## Telemetry

Our services collect summarized, anonymized data that can optionally be turned off. Click
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
