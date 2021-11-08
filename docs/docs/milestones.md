---
id: milestones
title: Milestones and Roadmap
---

## [Stable Release](https://github.com/ory/kratos/milestone/15)

All issues which need to be resolved before we release Ory Kratos as stable.

### [Feat](https://github.com/ory/kratos/labels/feat)

New feature or request.

#### Issues

- [ ] Support prefixed env vars
      ([kratos#1855](https://github.com/ory/kratos/issues/1855))

## [Incubating Release](https://github.com/ory/kratos/milestone/14)

This milestone tracks all issues we want to resolve before Ory Kratos goes into
incubating phase.

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [ ] Refresh Sessions Without Having to Log In Again
      ([kratos#615](https://github.com/ory/kratos/issues/615)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Fetching a settings request after error is missing identity data
      ([kratos#689](https://github.com/ory/kratos/issues/689)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Implement email TTL for non-working/non-existant emails
      ([kratos#944](https://github.com/ory/kratos/issues/944)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Courier Watcher should start a (tracing-) span
      ([kratos#1886](https://github.com/ory/kratos/issues/1886))
- [x] Feature Request: Have access to username in email templates
      ([kratos#925](https://github.com/ory/kratos/issues/925))
- [x] panic: a handle is already registered for path
      '/self-service/recovery/methods/link'
      ([kratos#1068](https://github.com/ory/kratos/issues/1068))
- [x] Relative path in ui_url or default_browser_return_url cause runtime crash
      ([kratos#1446](https://github.com/ory/kratos/issues/1446))
- [x] /schemas/default does not work with base64 identity schemas
      ([kratos#1529](https://github.com/ory/kratos/issues/1529))
- [x] Auto-Login on Registration flow does not return `session`, only `identity`
      for SPA requests
      ([kratos#1604](https://github.com/ory/kratos/issues/1604))

### [Feat](https://github.com/ory/kratos/labels/feat)

New feature or request.

#### Issues

- [ ] Do not send credentials to hooks
      ([kratos#77](https://github.com/ory/kratos/issues/77)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Implement immutable keyword in JSON Schema for Identity Traits
      ([kratos#117](https://github.com/ory/kratos/issues/117))
- [ ] Implement Hydra integration
      ([kratos#273](https://github.com/ory/kratos/issues/273))
- [ ] SMTP Error spams the server logs
      ([kratos#402](https://github.com/ory/kratos/issues/402))
- [ ] How to sign in with Twitter
      ([kratos#517](https://github.com/ory/kratos/issues/517))
- [ ] Throttling repeated login requests
      ([kratos#654](https://github.com/ory/kratos/issues/654))
- [ ] Require identity deactivation before administrative deletion
      ([kratos#657](https://github.com/ory/kratos/issues/657))
- [ ] Self-service GDPR identity export
      ([kratos#658](https://github.com/ory/kratos/issues/658))
- [ ] Rename strategy to method in internal APIs and Documentation
      ([kratos#683](https://github.com/ory/kratos/issues/683)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Support OAuth2 and OpenID Connect in API-based flows
      ([kratos#707](https://github.com/ory/kratos/issues/707))
- [ ] improve multi schema handling in different auth flows
      ([kratos#765](https://github.com/ory/kratos/issues/765))
- [ ] Add option for disabling registration
      ([kratos#882](https://github.com/ory/kratos/issues/882)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Password policies: make api.pwnedpasswords.com configurable
      ([kratos#969](https://github.com/ory/kratos/issues/969))
- [ ] Access identity schema information (e.g. usernames) within the jsonnet
      mapper ([kratos#987](https://github.com/ory/kratos/issues/987))
- [ ] login ui returned without flowid
      ([kratos#1055](https://github.com/ory/kratos/issues/1055))
- [ ] [Feature] LOG_LEVEL to allow stacktrace in error for /api endpoint
      ([kratos#1589](https://github.com/ory/kratos/issues/1589))
- [ ] Deprecate webhook loader without URI scheme
      ([kratos#1610](https://github.com/ory/kratos/issues/1610)) -
      [@Patrik](https://github.com/zepatrik)
- [ ] EMail on registration with existing identifier
      ([kratos#1625](https://github.com/ory/kratos/issues/1625))
- [ ] Require second factor only if device is unknown
      ([kratos#1643](https://github.com/ory/kratos/issues/1643))
- [ ] Regenerate lookup secrets as part of login when last secret was used
      ([kratos#1649](https://github.com/ory/kratos/issues/1649))
- [ ] Expand password validation error messages to facilitate i18n
      ([kratos#1071](https://github.com/ory/kratos/issues/1071))
- [ ] User deactivation does not deactivate user sessions
      ([kratos#1811](https://github.com/ory/kratos/issues/1811))
- [ ] Implement full-spec with e2e tests for react native app
      ([kratos#1820](https://github.com/ory/kratos/issues/1820))
- [x] Validate identity schema on load
      ([kratos#701](https://github.com/ory/kratos/issues/701)) -
      [@Alano Terblanche](https://github.com/Benehiko)
- [x] Add i18n support to mail templates
      ([kratos#834](https://github.com/ory/kratos/issues/834))
- [x] Allow account recovery for identities without email address
      ([kratos#1419](https://github.com/ory/kratos/issues/1419))
- [x] Field validation answer status code 422 instead of 400
      ([kratos#1559](https://github.com/ory/kratos/issues/1559))
- [x] Potentially encrypt settings InternalContext
      ([kratos#1647](https://github.com/ory/kratos/issues/1647))
- [x] Endpoint for fetching all schema IDs or all schemas
      ([kratos#1699](https://github.com/ory/kratos/issues/1699))
- [x] Email Courier SMTP Header Configuration
      ([kratos#1725](https://github.com/ory/kratos/issues/1725))

### [Docs](https://github.com/ory/kratos/labels/docs)

Affects documentation.

#### Issues

- [ ] Config JSON Schema needs example values
      ([kratos#179](https://github.com/ory/kratos/issues/179)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Tag version in docker-compose before commit&tag push
      ([kratos#1738](https://github.com/ory/kratos/issues/1738))

### [Rfc](https://github.com/ory/kratos/labels/rfc)

A request for comments to discuss and share ideas.

#### Issues

- [ ] Introduce prevent extension in Identity JSON schema
      ([kratos#47](https://github.com/ory/kratos/issues/47)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] improve multi schema handling in different auth flows
      ([kratos#765](https://github.com/ory/kratos/issues/765))
- [ ] Strategies to prevent mass email sending for several flows
      ([kratos#1835](https://github.com/ory/kratos/issues/1835))
- [x] Field validation answer status code 422 instead of 400
      ([kratos#1559](https://github.com/ory/kratos/issues/1559))
- [x] Potentially encrypt settings InternalContext
      ([kratos#1647](https://github.com/ory/kratos/issues/1647))

### [Blocking](https://github.com/ory/kratos/labels/blocking)

Blocks milestones or other issues or pulls.

#### Issues

- [ ] Implement Hydra integration
      ([kratos#273](https://github.com/ory/kratos/issues/273))

### [Ci](https://github.com/ory/kratos/labels/ci)

Affects Continuous Integration (CI).

#### Issues

- [x] Tag version in docker-compose before commit&tag push
      ([kratos#1738](https://github.com/ory/kratos/issues/1738))

## [v0.10.0-alpha.1](https://github.com/ory/kratos/milestone/13)

_This milestone does not have a description._

### [Feat](https://github.com/ory/kratos/labels/feat)

New feature or request.

#### Issues

- [ ] Selfservice account deletion
      ([kratos#596](https://github.com/ory/kratos/issues/596))
- [ ] Add ability to import user credentials
      ([kratos#605](https://github.com/ory/kratos/issues/605)) -
      [@hackerman](https://github.com/aeneasr)

## [v0.7.x](https://github.com/ory/kratos/milestone/12)

_This milestone does not have a description._

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [x] Submitting successful link verification again ends up with JSON response
      ([kratos#1546](https://github.com/ory/kratos/issues/1546)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Opening verification link in another browser causes GET request to fail
      due to CSRF issue
      ([kratos#1547](https://github.com/ory/kratos/issues/1547)) -
      [@hackerman](https://github.com/aeneasr)
- [x] 0.7.3.alpha.1, docker, panic if SESSION_COOKIE_NAME is set in
      docker-compose.yml
      ([kratos#1695](https://github.com/ory/kratos/issues/1695))
- [x] kratos identities import - include "state" property of identity
      ([kratos#1767](https://github.com/ory/kratos/issues/1767))
- [x] Panic on recovery for deactivated user
      ([kratos#1794](https://github.com/ory/kratos/issues/1794))

#### Pull Requests

- [x] fix(sdk): use proper annotation for genericError
      ([kratos#1611](https://github.com/ory/kratos/pull/1611)) -
      [@hackerman](https://github.com/aeneasr)

### [Docs](https://github.com/ory/kratos/labels/docs)

Affects documentation.

#### Issues

- [x] Update CSRF pitfall section for admin endpoints
      ([kratos#1557](https://github.com/ory/kratos/issues/1557)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Different payload for stub:500
      ([kratos#1568](https://github.com/ory/kratos/issues/1568))

## [v0.9.0-alpha.1](https://github.com/ory/kratos/milestone/11)

This milestone focuses on passwordless authentication and improved recovery and
verification flows.

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [ ] Recovery link doesn't log out existing users
      ([kratos#1077](https://github.com/ory/kratos/issues/1077))
- [ ] Email verification link is automatically opened by mail provider
      ([kratos#1202](https://github.com/ory/kratos/issues/1202))
- [ ] Delete self-service flows after completion
      ([kratos#952](https://github.com/ory/kratos/issues/952)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Dotenv files are respected and incorrectly override environment variables
      ([kratos#1856](https://github.com/ory/kratos/issues/1856))
- [ ] Ask to re-authenticate despite only updating non-privileged fields.
      ([kratos#1889](https://github.com/ory/kratos/issues/1889))
- [x] recovery link is open by email app
      ([kratos#1142](https://github.com/ory/kratos/issues/1142))

### [Feat](https://github.com/ory/kratos/labels/feat)

New feature or request.

#### Issues

- [ ] Support email verification paswordless login
      ([kratos#286](https://github.com/ory/kratos/issues/286))
- [ ] Prevent account enumeration for profile updates
      ([kratos#292](https://github.com/ory/kratos/issues/292)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Admin/Selfservice session management
      ([kratos#655](https://github.com/ory/kratos/issues/655)) -
      [@Alano Terblanche](https://github.com/Benehiko)
- [ ] Write new recovery and verification strategy that does not rely on links
      ([kratos#1451](https://github.com/ory/kratos/issues/1451))
- [ ] Delete self-service flows after completion
      ([kratos#952](https://github.com/ory/kratos/issues/952)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Documentation refactoring
      ([kratos#1904](https://github.com/ory/kratos/issues/1904)) -
      [@hackerman](https://github.com/aeneasr)

### [Docs](https://github.com/ory/kratos/labels/docs)

Affects documentation.

#### Issues

- [ ] Documentation refactoring
      ([kratos#1904](https://github.com/ory/kratos/issues/1904)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Document that identity information (traits, etc) are available to token
      holders and backend systems
      ([kratos#43](https://github.com/ory/kratos/issues/43)) -
      [@hackerman](https://github.com/aeneasr)
- [x] doc page for ui error messages
      ([kratos#1784](https://github.com/ory/kratos/issues/1784)) -
      [@hackerman](https://github.com/aeneasr)

## [v0.8.0-alpha.1](https://github.com/ory/kratos/milestone/10)

This milestone focuses on MFA with TOTP, WebAuthn, and Recovery Codes.

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [x] Unmable to use Auth0 as a generic OIDC provider
      ([kratos#609](https://github.com/ory/kratos/issues/609))
- [x] Typescript ErrorContainer type is incorrect
      ([kratos#782](https://github.com/ory/kratos/issues/782))
- [x] Add option to remove recovery codes
      ([kratos#1751](https://github.com/ory/kratos/issues/1751)) -
      [@hackerman](https://github.com/aeneasr)
- [x] return_to_query_param not honored on OIDC login
      ([kratos#1773](https://github.com/ory/kratos/issues/1773)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Hide SQLite migration message
      ([kratos#1791](https://github.com/ory/kratos/issues/1791)) -
      [@hackerman](https://github.com/aeneasr)

### [Feat](https://github.com/ory/kratos/labels/feat)

New feature or request.

#### Issues

- [x] Implement identity state and administrative deactivation, deletion of
      identities ([kratos#598](https://github.com/ory/kratos/issues/598)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Add TLS configuration
      ([kratos#791](https://github.com/ory/kratos/issues/791))
- [x] More meta information about the managed identity
      ([kratos#820](https://github.com/ory/kratos/issues/820))
- [x] Provide return_to value as part of flow entity
      ([kratos#1121](https://github.com/ory/kratos/issues/1121))
- [x] Add `return_to` to self-service SDK methods including logout
      ([kratos#1605](https://github.com/ory/kratos/issues/1605)) -
      [@hackerman](https://github.com/aeneasr)

#### Pull Requests

- [x] feat: ReactJS, better SPA APIs, 2FA with WebAuthn & TOTP & Lookup Secrets,
      better errors, typescript e2e tests
      ([kratos#1624](https://github.com/ory/kratos/pull/1624)) -
      [@hackerman](https://github.com/aeneasr)
- [x] fix: slow CLI start up time
      ([kratos#1878](https://github.com/ory/kratos/pull/1878))

### [Docs](https://github.com/ory/kratos/labels/docs)

Affects documentation.

#### Issues

- [x] Include release notes in CHANGELOG.md
      ([kratos#1442](https://github.com/ory/kratos/issues/1442)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Include changelog in docs navigation
      ([kratos#1443](https://github.com/ory/kratos/issues/1443)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Config reference has not been updated since 0.5.0
      ([kratos#1597](https://github.com/ory/kratos/issues/1597)) -
      [@hackerman](https://github.com/aeneasr)

### [Blocking](https://github.com/ory/kratos/labels/blocking)

Blocks milestones or other issues or pulls.

#### Issues

- [ ] Ory Kratos 0.8 Release Prep
      ([kratos#1663](https://github.com/ory/kratos/issues/1663)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Add option to remove recovery codes
      ([kratos#1751](https://github.com/ory/kratos/issues/1751)) -
      [@hackerman](https://github.com/aeneasr)

#### Pull Requests

- [x] feat: ReactJS, better SPA APIs, 2FA with WebAuthn & TOTP & Lookup Secrets,
      better errors, typescript e2e tests
      ([kratos#1624](https://github.com/ory/kratos/pull/1624)) -
      [@hackerman](https://github.com/aeneasr)
