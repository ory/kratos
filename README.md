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

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [What is ORY Kratos?](#what-is-ory-kratos)
  - [Who's using it?](#whos-using-it)
- [Getting Started](#getting-started)
  - [Quickstart](#quickstart)
  - [Installation](#installation)
- [Ecosystem](#ecosystem)
  - [ORY Security Console: Administrative User Interface](#ory-security-console-administrative-user-interface)
  - [ORY Oathkeeper: Identity & Access Proxy](#ory-oathkeeper-identity--access-proxy)
  - [ORY Keto: Access Control Policies as a Server](#ory-keto-access-control-policies-as-a-server)
- [Security](#security)
  - [Disclosing vulnerabilities](#disclosing-vulnerabilities)
- [Telemetry](#telemetry)
- [Documentation](#documentation)
  - [Guide](#guide)
  - [HTTP API documentation](#http-api-documentation)
  - [Upgrading and Changelog](#upgrading-and-changelog)
  - [Command line documentation](#command-line-documentation)
  - [Develop](#develop)
    - [Dependencies](#dependencies)
      - [Install Tools](#install-tools)
    - [Formatting Code](#formatting-code)
    - [Running Tests](#running-tests)
      - [Short Tests](#short-tests)
      - [Regular Tests](#regular-tests)
    - [Build Docker](#build-docker)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## What is ORY Kratos?

ORY Kratos is an API-first Identity and User Management system that is built
according to
[cloud architecture best practices](https://www.ory.sh/docs/next/ecosystem/software-architecture-philosophy).
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

We highly recommend reading the [ORY Kratos introduction docs](https://www.ory.sh/docs/next/kratos/)
to learn more about ORY Krato's background, feature set, and differentiation
from other products.

### Who's using it?

<!--BEGIN ADOPTERS-->

The ORY community stands on the shoulders of individuals, companies, and
maintainers. We thank everyone involved - from submitting bug reports and
feature requests, to contributing patches, to sponsoring our work. Our community
is 1000+ strong and growing rapidly. The ORY stack protects 1.200.000.000+ API
requests every month with over 15.000+ active service nodes. We would have never
been able to achieve this without each and everyone of you!

The following list represents companies that have accompanied us along the way
and that have made outstanding contributions to our ecosystem. _If you think
that your company deserves a spot here, reach out to
<a href="mailto:hi@ory.sh">hi@ory.sh</a> now_!

**Please consider giving back by becoming a sponsor of our open source work on
<a href="https://www.patreon.com/_ory">Patreon</a> or
<a href="https://opencollective.com/ory">Open Collective</a>.**

<table>
    <thead>
        <tr>
            <th>Type</th>
            <th>Name</th>
            <th>Logo</th>
            <th>Website</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>Sponsor</td>
            <td>Raspberry PI Foundation</td>
            <td align="center"><img height="32px" src="./docs/adopters/raspi.svg" alt="Raspberry PI Foundation"></td>
            <td><a href="https://www.raspberrypi.org/">raspberrypi.org</a></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Kyma Project</a>
            <td align="center"><img height="32px" src="./docs/adopters/kyma.svg" alt="Kyma Project"></td>
            <td><a href="https://kyma-project.io">kyma-project.io</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>ThoughtWorks</td>
            <td align="center"><img height="32px" src="./docs/adopters/tw.svg" alt="ThoughtWorks"></td>
            <td><a href="https://www.thoughtworks.com/">thoughtworks.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Tulip</td>
            <td align="center"><img height="32px" src="./docs/adopters/tulip.svg" alt="Tulip Retail"></td>
            <td><a href="https://tulip.com/">tulip.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Cashdeck / All My Funds</td>
            <td align="center"><img height="32px" src="./docs/adopters/allmyfunds.svg" alt="All My Funds"></td>
            <td><a href="https://cashdeck.com.au/">cashdeck.com.au</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>3Rein</td>
            <td align="center"><img height="32px" src="./docs/adopters/3R-horiz.svg" alt="3Rein"></td>
            <td><a href="https://3rein.com/">3rein.com</a></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Hootsuite</td>
            <td align="center"><img height="32px" src="./docs/adopters/hootsuite.svg" alt="Hootsuite"></td>
            <td><a href="https://hootsuite.com/">hootsuite.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Segment</td>
            <td align="center"><img height="32px" src="./docs/adopters/segment.svg" alt="Segment"></td>
            <td><a href="https://segment.com/">segment.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Arduino</td>
            <td align="center"><img height="32px" src="./docs/adopters/arduino.svg" alt="Arduino"></td>
            <td><a href="https://www.arduino.cc/">arduino.cc</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>OrderMyGear</td>
            <td align="center"><img height="32px" src="./docs/adopters/ordermygear.svg" alt="OrderMyGear"></td>
            <td><a href="https://www.ordermygear.com/">ordermygear.com</a></td>
        </tr>
    </tdbody>
</table>

We also want to thank all individual contributors

<img src="https://opencollective.com/ory/contributors.svg?width=890&button=false" /></a>

as well as all of our backers

<a href="https://opencollective.com/ory#backers" target="_blank"><img src="https://opencollective.com/ory/backers.svg?width=890"></a>

and past & current supporters (in alphabetical order) on
[Patreon](https://www.patreon.com/_ory): Alexander Alimovs, Billy, Chancy
Kennedy, Drozzy, Edwin Trejos, Howard Edidin, Ken Adler Oz Haven, Stefan Hans,
TheCrealm.

<em>\* Uses one of ORY's major projects in production.</em>

<!--END ADOPTERS-->




## Getting Started

To get started, head over to the [ORY Kratos Documentation](https://www.ory.sh/docs/next/kratos).

### Quickstart

The **[ORY Kratos Quickstart](https://www.ory.sh/docs/next/kratos/quickstart)** teaches you ORY Kratos basics
and sets up an example based on Docker Compose in less than five minutes.

### Installation

Head over to the [ORY Developer Documentation](https://www.ory.sh/docs/next/kratos/install.md) to learn how to install ORY Kratos on Linux, macOS, Windows, and Docker and how to build ORY Kratos from source.

## Ecosystem

<!--BEGIN ECOSYSTEM-->
We build Ory on several guiding principles when it comes to our architecture design:

- Minimal dependencies
- Runs everywhere
- Scales without effort
- Minimize room for human and network errors

ORY's architecture designed to run best on a Container Orchestration Systems such as Kubernetes, CloudFoundry, OpenShift, and similar projects.
Binaries are small (5-15MB) and available for all popular processor types (ARM, AMD64, i386) and operating
systems (FreeBSD, Linux, macOS, Windows) without system dependencies (Java, Node, Ruby, libxml, ...).

### ORY Kratos: Identity and User Infrastructure and Management

[ORY Kratos](https://github.com/ory/kratos) is an API-first Identity and User
Management system that is built according to
[cloud architecture best practices](https://www.ory.sh/docs/next/ecosystem/software-architecture-philosophy).
It implements core use cases that almost every software application needs to
deal with: Self-service Login and Registration, Multi-Factor Authentication
(MFA/2FA), Account Recovery and Verification, Profile and Account Management.

### ORY Hydra: OAuth2 & OpenID Connect Server

[ORY Hydra](https://github.com/ory/hydra) is an OpenID Certified™ OAuth2 and OpenID Connect
Provider can connect to any existing identity database (LDAP, AD, KeyCloak, PHP+MySQL, ...)
and user interface.

### ORY Oathkeeper: Identity & Access Proxy

[ORY Oathkeeper](https://github.com/ory/oathkeeper) is a BeyondCorp/Zero Trust
Identity & Access Proxy (IAP) with configurable authentication, authorization,
and request mutation rules for your web services: Authenticate JWT, Access Tokens,
API Keys, mTLS; Check if the contained subject is allowed to perform the request;
Encode resulting content into custom headers (`X-User-ID`), JSON Web Tokens
and more!

### ORY Keto: Access Control Policies as a Server

[ORY Keto](https://github.com/ory/keto) is a policy decision point. It uses a
set of access control policies, similar to AWS IAM Policies, in order to
determine whether a subject (user, application, service, car, ...) is authorized
to perform a certain action on a resource.
<!--END ECOSYSTEM-->



## Security

Running identity infrastructure requires [attention and knowledge of threat models](https://www.ory.sh/docs/next/kratos/concepts/security).

### Disclosing vulnerabilities

If you think you found a security vulnerability, please refrain from posting it publicly on the forums, the chat, or GitHub
and send us an email to [hi@ory.am](mailto:hi@ory.sh) instead.

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

### Develop

We encourage all contributions and encourage you to read our [contribution guidelines](./CONTRIBUTING.md)

#### Dependencies

You need Go 1.13+ with `GO111MODULE=on` and (for the test suites):

- Docker and Docker Compose
- Makefile
- NodeJS / npm

It is possible to develop ORY Kratos on Windows, but please be aware that all guides assume a Unix shell like bash or zsh.

##### Install Tools

When cloning ORY Kratos, run `make tools`. It will download several required dependencies. If you haven't run the command
in a while it's probably a good idea to run it again.

#### Formatting Code

You can format all code using `make format`. Our CI checks if your code is properly formatted.

#### Running Tests

There are three types of tests you can run:

- Short tests (do not require a SQL database like PostgreSQL)
- Regular tests (do require PostgreSQL, MySQL, CockroachDB)
- End to end tests (do require databases and will use a test browser)

##### Short Tests

Short tests run fairly quickly. You can either test all of the code at once

```shell script
go test -short -tags sqlite ./...
```

or test just a specific module:

```shell script
cd client; go test -tags sqlite -short .
```

##### Regular Tests

Regular tests require a database set up. Our test suite is able to work with docker directly (using [ory/dockertest](https://github.com/ory/dockertest))
but we encourage to use the Makefile instead. Using dockertest can bloat the number of Docker Images on your system
and are quite slow. Instead we recommend doing:

```shell script
make test
```

Please be aware that `make test` recreates the databases every time you run `make test`. This can be annoying if
you are trying to fix something very specific and need the database tests all the time. In that case we
suggest that you initialize the databases with:

```shell script
make resetdb
export TEST_DATABASE_MYSQL='mysql://root:secret@(127.0.0.1:3444)/mysql?parseTime=true'
export TEST_DATABASE_POSTGRESQL='postgres://postgres:secret@127.0.0.1:3445/kratos?sslmode=disable'
export TEST_DATABASE_COCKROACHDB='cockroach://root@127.0.0.1:3446/defaultdb?sslmode=disable'
```

Then you can run `go test` as often as you'd like:

```shell script
go test -tags sqlite ./...

# or in a module:
cd client; go test  -tags sqlite  .
```

#### Build Docker

You can build a development Docker Image using:

```shell script
make docker
```
