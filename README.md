<h1 align="center"><img src="https://raw.githubusercontent.com/ory/meta/master/static/banners/kratos.svg" alt="Ory Kratos - Cloud native Identity and User Management"></h1>

<h4 align="center">
    <a href="https://www.ory.sh/chat">Chat</a> |
    <a href="https://github.com/ory/kratos/discussions">Discussions</a> |
    <a href="http://eepurl.com/di390P">Newsletter</a><br/><br/>
    <a href="https://www.ory.sh/kratos/docs/">Guide</a> |
    <a href="https://www.ory.sh/kratos/docs/sdk/api">API Docs</a> |
    <a href="https://godoc.org/github.com/ory/kratos">Code Docs</a><br/><br/>
    <a href="https://opencollective.com/ory">Support this project!</a><br/><br/>
    <a href="https://www.ory.sh/jobs/">Work in Open Source, Ory is hiring!</a>
</h4>

---

<p align="left">
    <a href="https://github.com/ory/kratos/actions/workflows/ci.yaml"><img src="https://github.com/ory/kratos/actions/workflows/ci.yaml/badge.svg?branch=master&event=push" alt="CI Tasks for Ory Kratos"></a>
    <a href="https://codecov.io/gh/ory/kratos"><img src="https://codecov.io/gh/ory/kratos/branch/master/graph/badge.svg?token=6t0QqOdurR"/></a>
    <a href="https://bestpractices.coreinfrastructure.org/projects/4979"><img src="https://bestpractices.coreinfrastructure.org/projects/4979/badge" alt="CII Best Practices"></a>
    <a href="https://opencollective.com/ory" alt="sponsors on Open Collective"><img src="https://opencollective.com/ory/backers/badge.svg" /></a>
    <a href="https://opencollective.com/ory" alt="Sponsors on Open Collective"><img src="https://opencollective.com/ory/sponsors/badge.svg" /></a>
    <a href="https://github.com/ory/kratos/blob/master/CODE_OF_CONDUCT.md" alt="Ory Code of Conduct"><img src="https://img.shields.io/badge/ory-code%20of%20conduct-green" /></a>
</>

Ory Kratos is _the_ developer-friendly, security-hardened and battle-tested
Identity, User Management and Authentication system for the Cloud. Finally, it
is no longer necessary to implement User Login for the umpteenth time!

## Ory Kratos on the Ory Network

The [Ory Network](https://www.ory.sh/cloud) is the fastest, most secure and
worry-free way to use Ory's Services. **Ory Identities** is powered by the Ory
Kratos open source identity server, and it's fully API-compatible.

The Ory Network provides the infrastructure for modern end-to-end security:

- **Identity & credential management scaling to billions of users and devices**
- **Registration, Login and Account management flows for passkey, biometric,
  social, SSO and multi-factor authentication**
- **Pre-built login, registration and account management pages and components**
- OAuth2 and OpenID provider for single sign on, API access and
  machine-to-machine authorization
- Low-latency permission checks based on Google's Zanzibar model and with
  built-in support for the Ory Permission Language

It's fully managed, highly available, developer & compliance-friendly!

- GDPR-friendly secure storage with data locality
- Cloud-native APIs, compatible with Ory's Open Source servers
- Comprehensive admin tools with the web-based Ory Console and the Ory Command
  Line Interface (CLI)
- Extensive documentation, straightforward examples and easy-to-follow guides
- Fair, usage-based [pricing](https://www.ory.sh/pricing)

Sign up for a
[**free developer account**](https://console.ory.sh/registration?utm_source=github&utm_medium=banner&utm_campaign=kratos-readme)
today!

## Ory Network Hybrid Support Plan

Ory offers a support plan for Ory Network Hybrid, including Ory on private cloud deployments. If you have a self-hosted solution and would like help, consider a support plan!  
The team at Ory has years of experience in cloud computing. Ory's offering is the only official program for qualified support from the maintainers.  
For more information see the **[website](https://www.ory.sh/support/)** or **[book a meeting](https://www.ory.sh/contact/)**!

### Quickstart

Install the [Ory CLI](https://www.ory.sh/docs/guides/cli/installation) and
create a new project to get started with Ory Identities right away:

```
# If you don't have Ory CLI installed yet:
bash <(curl https://raw.githubusercontent.com/ory/meta/master/install.sh) -b . ory
sudo mv ./ory /usr/local/bin/

# Sign up
ory auth

# Create project
ory create project
```

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

**Table of Contents**

- [Ory Kratos on the Ory Network](#ory-kratos-on-the-ory-network)
  - [Quickstart](#quickstart)
- [What is Ory Kratos?](#what-is-ory-kratos)
  - [Who is using it?](#who-is-using-it)
- [Getting Started](#getting-started)
  - [Installation](#installation)
- [Ecosystem](#ecosystem)
  - [Ory Kratos: Identity and User Infrastructure and Management](#ory-kratos-identity-and-user-infrastructure-and-management)
  - [Ory Hydra: OAuth2 & OpenID Connect Server](#ory-hydra-oauth2--openid-connect-server)
  - [Ory Oathkeeper: Identity & Access Proxy](#ory-oathkeeper-identity--access-proxy)
  - [Ory Keto: Access Control Policies as a Server](#ory-keto-access-control-policies-as-a-server)
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
    - [Install from source](#install-from-source)
    - [Formatting Code](#formatting-code)
    - [Running Tests](#running-tests)
      - [Short Tests](#short-tests)
      - [Regular Tests](#regular-tests)
      - [Updating Test Fixtures](#updating-test-fixtures)
      - [End-to-End Tests](#end-to-end-tests)
    - [Build Docker](#build-docker)
    - [Documentation Tests](#documentation-tests)
    - [Preview API documentation](#preview-api-documentation)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## What is Ory Kratos?

Ory Kratos is an API-first Identity and User Management system that is built
according to
[cloud architecture best practices](https://www.ory.sh/docs/ecosystem/software-architecture-philosophy).
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
  Codes (in case of MFA device loss), and others.
- **Profile and Account Management**: Update passwords, personal details, email
  addresses, linked social profiles using secure flows.
- **Admin APIs**: Import, update, delete identities.

We highly recommend reading the
[Ory Kratos introduction docs](https://www.ory.sh/kratos/docs/) to learn more
about Ory Krato's background, feature set, and differentiation from other
products.

### Who is using it?

<!--BEGIN ADOPTERS-->

The Ory community stands on the shoulders of individuals, companies, and
maintainers. We thank everyone involved - from submitting bug reports and
feature requests, to contributing patches, to sponsoring our work. Our community
is 1000+ strong and growing rapidly. The Ory stack protects 16.000.000.000+ API
requests every month with over 250.000+ active service nodes. We would have
never been able to achieve this without each and everyone of you!

The following list represents companies that have accompanied us along the way
and that have made outstanding contributions to our ecosystem. _If you think
that your company deserves a spot here, reach out to
<a href="mailto:office-muc@ory.sh">office-muc@ory.sh</a> now_!

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
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/raspi.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/raspi.svg" alt="Raspberry PI Foundation">
                </picture>
            </td>
            <td><a href="https://www.raspberrypi.org/">raspberrypi.org</a></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Kyma Project</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/kyma.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/kyma.svg" alt="Kyma Project">
                </picture>
            </td>
            <td><a href="https://kyma-project.io">kyma-project.io</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Tulip</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/tulip.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/tulip.svg" alt="Tulip Retail">
                </picture>
            </td>
            <td><a href="https://tulip.com/">tulip.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Cashdeck / All My Funds</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/allmyfunds.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/allmyfunds.svg" alt="All My Funds">
                </picture>
            </td>
            <td><a href="https://cashdeck.com.au/">cashdeck.com.au</a></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Hootsuite</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/hootsuite.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/hootsuite.svg" alt="Hootsuite">
                </picture>
            </td>
            <td><a href="https://hootsuite.com/">hootsuite.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Segment</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/segment.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/segment.svg" alt="Segment">
                </picture>
            </td>
            <td><a href="https://segment.com/">segment.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Arduino</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/arduino.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/arduino.svg" alt="Arduino">
                </picture>
            </td>
            <td><a href="https://www.arduino.cc/">arduino.cc</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>DataDetect</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/datadetect.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/datadetect.svg" alt="Datadetect">
                </picture>
            </td>
            <td><a href="https://unifiedglobalarchiving.com/data-detect/">unifiedglobalarchiving.com/data-detect/</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Sainsbury's</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/sainsburys.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/sainsburys.svg" alt="Sainsbury's">
                </picture>
            </td>
            <td><a href="https://www.sainsburys.co.uk/">sainsburys.co.uk</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Contraste</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/contraste.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/contraste.svg" alt="Contraste">
                </picture>
            </td>
            <td><a href="https://www.contraste.com/en">contraste.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Reyah</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/reyah.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/reyah.svg" alt="Reyah">
                </picture>
            </td>
            <td><a href="https://reyah.eu/">reyah.eu</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Zero</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/commitzero.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/commitzero.svg" alt="Project Zero by Commit">
                </picture>
            </td>
            <td><a href="https://getzero.dev/">getzero.dev</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Padis</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/padis.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/padis.svg" alt="Padis">
                </picture>
            </td>
            <td><a href="https://padis.io/">padis.io</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Cloudbear</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/cloudbear.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/cloudbear.svg" alt="Cloudbear">
                </picture>
            </td>
            <td><a href="https://cloudbear.eu/">cloudbear.eu</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Security Onion Solutions</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/securityonion.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/securityonion.svg" alt="Security Onion Solutions">
                </picture>
            </td>
            <td><a href="https://securityonionsolutions.com/">securityonionsolutions.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Factly</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/factly.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/factly.svg" alt="Factly">
                </picture>
            </td>
            <td><a href="https://factlylabs.com/">factlylabs.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Nortal</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/nortal.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/nortal.svg" alt="Nortal">
                </picture>
            </td>
            <td><a href="https://nortal.com/">nortal.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>OrderMyGear</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/ordermygear.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/ordermygear.svg" alt="OrderMyGear">
                </picture>
            </td>
            <td><a href="https://www.ordermygear.com/">ordermygear.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Spiri.bo</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/spiribo.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/spiribo.svg" alt="Spiri.bo">
                </picture>
            </td>
            <td><a href="https://spiri.bo/">spiri.bo</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Strivacity</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/strivacity.svg" />
                    <img height="16px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/strivacity.svg" alt="Spiri.bo">
                </picture>
            </td>
            <td><a href="https://strivacity.com/">strivacity.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Hanko</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/hanko.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/hanko.svg" alt="Hanko">
                </picture>
            </td>
            <td><a href="https://hanko.io/">hanko.io</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Rabbit</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/rabbit.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/rabbit.svg" alt="Rabbit">
                </picture>
            </td>
            <td><a href="https://rabbit.co.th/">rabbit.co.th</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>inMusic</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/inmusic.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/inmusic.svg" alt="InMusic">
                </picture>
            </td>
            <td><a href="https://inmusicbrands.com/">inmusicbrands.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Buhta</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/buhta.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/buhta.svg" alt="Buhta">
                </picture>
            </td>
            <td><a href="https://buhta.com/">buhta.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Connctd</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/connctd.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/connctd.svg" alt="Connctd">
                </picture>
            </td>
            <td><a href="https://connctd.com/">connctd.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Paralus</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/paralus.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/paralus.svg" alt="Paralus">
                </picture>
            </td>
            <td><a href="https://www.paralus.io/">paralus.io</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>TIER IV</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/tieriv.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/tieriv.svg" alt="TIER IV">
                </picture>
            </td>
            <td><a href="https://tier4.jp/en/">tier4.jp</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>R2Devops</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/r2devops.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/r2devops.svg" alt="R2Devops">
                </picture>
            </td>
            <td><a href="https://r2devops.io/">r2devops.io</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>LunaSec</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/lunasec.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/lunasec.svg" alt="LunaSec">
                </picture>
            </td>
            <td><a href="https://www.lunasec.io/">lunasec.io</a></td>
        </tr>
            <tr>
            <td>Adopter *</td>
            <td>Serlo</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/serlo.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/serlo.svg" alt="Serlo">
                </picture>
            </td>
            <td><a href="https://serlo.org/">serlo.org</a></td>
        </tr>
        </tr>
            <tr>
            <td>Adopter *</td>
            <td>dyrector.io</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/dyrector_io.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/dyrector_io.svg" alt="dyrector.io">
                </picture>
            </td>
            <td><a href="https://dyrector.io/">dyrector.io</a></td>
        </tr>
        </tr>
            <tr>
            <td>Adopter *</td>
            <td>Stackspin</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/stackspin.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/stackspin.svg" alt="stackspin.net">
                </picture>
            </td>
            <td><a href="https://www.stackspin.net/">stackspin.net</a></td>
        </tr>
        </tr>
            <tr>
            <td>Adopter *</td>
            <td>Amplitude</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/amplitude.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/amplitude.svg" alt="amplitude.com">
                </picture>
            </td>
            <td><a href="https://amplitude.com/">amplitude.com</a></td>
        </tr>
    </tbody>
</table>

We also want to thank all individual contributors

<a href="https://opencollective.com/ory" target="_blank"><img src="https://opencollective.com/ory/contributors.svg?width=890&limit=714&button=false" /></a>

as well as all of our backers

<a href="https://opencollective.com/ory#backers" target="_blank"><img src="https://opencollective.com/ory/backers.svg?width=890"></a>

and past & current supporters (in alphabetical order) on
[Patreon](https://www.patreon.com/_ory): Alexander Alimovs, Billy, Chancy
Kennedy, Drozzy, Edwin Trejos, Howard Edidin, Ken Adler Oz Haven, Stefan Hans,
TheCrealm.

<em>\* Uses one of Ory's major projects in production.</em>

<!--END ADOPTERS-->

## Getting Started

To get started with some easy examples, head over to the
[Get Started Documentation](https://www.ory.sh/docs/guides/protect-page-login/).

### Installation

Head over to the
[Ory Developer Documentation](https://www.ory.sh/kratos/docs/install) to learn
how to install Ory Kratos on Linux, macOS, Windows, and Docker and how to build
Ory Kratos from source.

## Ecosystem

<!--BEGIN ECOSYSTEM-->

We build Ory on several guiding principles when it comes to our architecture
design:

- Minimal dependencies
- Runs everywhere
- Scales without effort
- Minimize room for human and network errors

Ory's architecture is designed to run best on a Container Orchestration system
such as Kubernetes, CloudFoundry, OpenShift, and similar projects. Binaries are
small (5-15MB) and available for all popular processor types (ARM, AMD64, i386)
and operating systems (FreeBSD, Linux, macOS, Windows) without system
dependencies (Java, Node, Ruby, libxml, ...).

### Ory Kratos: Identity and User Infrastructure and Management

[Ory Kratos](https://github.com/ory/kratos) is an API-first Identity and User
Management system that is built according to
[cloud architecture best practices](https://www.ory.sh/docs/next/ecosystem/software-architecture-philosophy).
It implements core use cases that almost every software application needs to
deal with: Self-service Login and Registration, Multi-Factor Authentication
(MFA/2FA), Account Recovery and Verification, Profile, and Account Management.

### Ory Hydra: OAuth2 & OpenID Connect Server

[Ory Hydra](https://github.com/ory/hydra) is an OpenID Certified™ OAuth2 and
OpenID Connect Provider which easily connects to any existing identity system by
writing a tiny "bridge" application. It gives absolute control over the user
interface and user experience flows.

### Ory Oathkeeper: Identity & Access Proxy

[Ory Oathkeeper](https://github.com/ory/oathkeeper) is a BeyondCorp/Zero Trust
Identity & Access Proxy (IAP) with configurable authentication, authorization,
and request mutation rules for your web services: Authenticate JWT, Access
Tokens, API Keys, mTLS; Check if the contained subject is allowed to perform the
request; Encode resulting content into custom headers (`X-User-ID`), JSON Web
Tokens and more!

### Ory Keto: Access Control Policies as a Server

[Ory Keto](https://github.com/ory/keto) is a policy decision point. It uses a
set of access control policies, similar to AWS IAM Policies, in order to
determine whether a subject (user, application, service, car, ...) is authorized
to perform a certain action on a resource.

<!--END ECOSYSTEM-->

## Security

Running identity infrastructure requires
[attention and knowledge of threat models](https://www.ory.sh/kratos/docs/concepts/security).

### Disclosing vulnerabilities

If you think you found a security vulnerability, please refrain from posting it
publicly on the forums, the chat, or GitHub. You can find all info for
responsible disclosure in our
[security.txt](https://www.ory.sh/.well-known/security.txt).

## Telemetry

Ory's services collect summarized, anonymized data that can optionally be turned
off. Click [here](https://www.ory.sh/docs/ecosystem/sqa) to learn more.

## Documentation

### Guide

The Guide is available [here](https://www.ory.sh/kratos/docs).

### HTTP API documentation

The HTTP API is documented [here](https://www.ory.sh/kratos/docs/sdk/api).

### Upgrading and Changelog

New releases might introduce breaking changes. To help you identify and
incorporate those changes, we document these changes in the
[CHANGELOG.md](./CHANGELOG.md). For upgrading, please visit the
[upgrade guide](https://www.ory.sh/kratos/docs/guides/upgrade).

### Command line documentation

Run <code type="shell/command">kratos -h</code> or
<code type="shell/command">kratos help</code>.

### Develop

We encourage all contributions and encourage you to read our
[contribution guidelines](./CONTRIBUTING.md)

#### Dependencies

You need Go 1.16+ and (for the test suites):

- Docker and Docker Compose
- Makefile
- NodeJS / npm

It is possible to develop Ory Kratos on Windows, but please be aware that all
guides assume a Unix shell like bash or zsh.

#### Install from source

<pre type="make/command">
make install
</pre>

#### Formatting Code

You can format all code using <code type="make/command">make format</code>. Our
CI checks if your code is properly formatted.

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

Regular tests require a database set up. Our test suite is able to work with
docker directly (using [ory/dockertest](https://github.com/ory/dockertest)) but
we encourage to use the Makefile instead. Using dockertest can bloat the number
of Docker Images on your system and are quite slow. Instead we recommend doing:

<pre type="make/command">
make test
</pre>

Please be aware that <code type="make/command">make test</code> recreates the
databases every time you run <code type="make/command">make test</code>. This
can be annoying if you are trying to fix something very specific and need the
database tests all the time. In that case we suggest that you initialize the
databases with:

<a type="make/command">

```shell script
make test-resetdb
export TEST_DATABASE_MYSQL='mysql://root:secret@(127.0.0.1:3444)/mysql?parseTime=true'
export TEST_DATABASE_POSTGRESQL='postgres://postgres:secret@127.0.0.1:3445/kratos?sslmode=disable'
export TEST_DATABASE_COCKROACHDB='cockroach://root@127.0.0.1:3446/defaultdb?sslmode=disable'
```

</a>

Then you can run `go test` as often as you'd like:

```shell script
go test -tags sqlite ./...

# or in a module:
cd client; go test  -tags sqlite  .
```

##### Updating Test Fixtures

Some tests use fixtures. If payloads change, you can update them with:

```
make test-update-snapshots
```

This will only update the snapshots of the short tests. To update all snapshots,
run:

```bash
UPDATE_SNAPSHOTS=true go test -p 4 -tags sqlite ./...
```

You can also run this command from a sub folder.

##### End-to-End Tests

We use [Cypress](https://www.cypress.io) to run our e2e tests.

⚠️ To run Cypress on ARM based Mac's, at the moment it is
[necessary to install Rosetta 2](https://www.cypress.io/blog/2021/01/20/running-cypress-on-the-apple-m1-silicon-arm-architecture-using-rosetta-2/).
To install, use the command -
`softwareupdate --install-rosetta --agree-to-license`

The simplest way to develop e2e tests is:

<pre type="repo/executable">
./test/e2e/run.sh --dev sqlite
</pre>

You can run all tests (with databases) using:

<pre type="make/command">
make test-e2e
</pre>

For more details, run:

<pre type="repo/executable">
./test/e2e/run.sh
</pre>

**Run only a singular test**

Add `.only` to the test you would like to run.

For example:

```ts
it.only('invalid remote recovery email template', () => {
    ...
})
```

**Run a subset of tests**

This will require editing the `cypress.json` file located in the `test/e2e/`
folder.

Add the `testFiles` option and specify the test to run inside the
`cypress/integration` folder. As an example we will add only the `network`
tests.

```json
"testFiles": ["profiles/network/*"],
```

Now start the tests again using the run script or makefile.

#### Build Docker

You can build a development Docker Image using:

<pre type="make/command">
make docker
</pre>

#### Documentation Tests

To prepare documentation tests, run `npm i` to install
[Text-Runner](https://github.com/kevgo/text-runner).

- test all documentation: <code type="make/command">make test-docs</code>
- test an individual file: <code type="npm/installed-executable">text-run</code>

#### Preview API documentation

- update the SDK including the OpenAPI specification:
  <code type="make/command">make sdk</code>
- run preview server for API documentation: <code type="make/command">make
  docs/api</code>
- run preview server for swagger documentation: <code type="make/command">make
  docs/swagger</code>
