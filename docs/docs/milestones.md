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
- [ ] Password policies: make api.pwnedpasswords.com configurable
      ([kratos#969](https://github.com/ory/kratos/issues/969))
- [ ] Access identity schema information (e.g. usernames) within the jsonnet
      mapper ([kratos#987](https://github.com/ory/kratos/issues/987))
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
- [x] Add option for disabling registration
      ([kratos#882](https://github.com/ory/kratos/issues/882)) -
      [@hackerman](https://github.com/aeneasr)
- [x] login ui returned without flowid
      ([kratos#1055](https://github.com/ory/kratos/issues/1055))
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

## [v0.10.0-alpha.1](https://github.com/ory/kratos/milestone/13)

_This milestone does not have a description._

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [ ] WebAuth docs are missing
      ([kratos#1951](https://github.com/ory/kratos/issues/1951)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Add missing cloud docs
      ([kratos#1952](https://github.com/ory/kratos/issues/1952)) -
      [@hackerman](https://github.com/aeneasr)

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
- [ ] Replace magic links with TOTP OOB codes
      ([kratos#1451](https://github.com/ory/kratos/issues/1451))
- [ ] Delete self-service flows after completion
      ([kratos#952](https://github.com/ory/kratos/issues/952)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Documentation refactoring
      ([kratos#1904](https://github.com/ory/kratos/issues/1904)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Update config key neames
      ([kratos#1979](https://github.com/ory/kratos/issues/1979)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Admin/Selfservice session management
      ([kratos#655](https://github.com/ory/kratos/issues/655)) -
      [@Patrik](https://github.com/zepatrik)

## [v0.8.0-alpha.1](https://github.com/ory/kratos/milestone/10)

This milestone focuses on MFA with TOTP, WebAuthn, and Recovery Codes.

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [ ] Recovery link answers with JSON payload for API flows
      ([kratos#2049](https://github.com/ory/kratos/issues/2049))
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
- [x] Resumable session is not cleared up on error
      ([kratos#2016](https://github.com/ory/kratos/issues/2016)) -
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

## [v0.7.0-alpha.1](https://github.com/ory/kratos/milestone/9)

_This milestone does not have a description._

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [x] Do not create system errors on duplicate credentials when linking oidc
      providers ([kratos#694](https://github.com/ory/kratos/issues/694))
- [x] Investigate why smtps fails but smtp does not
      ([kratos#781](https://github.com/ory/kratos/issues/781)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Reloading config values does not work
      ([kratos#804](https://github.com/ory/kratos/issues/804)) -
      [@hackerman](https://github.com/aeneasr)
- [x] handle 409 error in settings flow
      ([kratos#1107](https://github.com/ory/kratos/issues/1107))
- [x] Account Recovery API flow requires CSRF cookie
      ([kratos#1141](https://github.com/ory/kratos/issues/1141)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Receive the verification email every time when I update my password
      ([kratos#1221](https://github.com/ory/kratos/issues/1221))
- [x] No email validation for /verify and /recovery page. email queue is
      backlogged with invalid emails.
      ([kratos#1285](https://github.com/ory/kratos/issues/1285))
- [x] Verification submission endpoint (submitSelfServiceVerificationFlow) still
      requires CSRF cookie for API flows
      ([kratos#1368](https://github.com/ory/kratos/issues/1368)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Social sign uop should reduce confusion on sign up button
      ([kratos#1422](https://github.com/ory/kratos/issues/1422)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Update profile with field error returning 502
      ([kratos#1425](https://github.com/ory/kratos/issues/1425)) -
      [@hackerman](https://github.com/aeneasr)
- [x] fix: settings flow error handle should access schemas using configured
      "local" URLs instead of public URLs
      ([kratos#1448](https://github.com/ory/kratos/issues/1448))
- [x] Verification does not include success message
      ([kratos#1450](https://github.com/ory/kratos/issues/1450)) -
      [@hackerman](https://github.com/aeneasr)

#### Pull Requests

- [x] fix: continue button for oidc registration step
      ([kratos#1427](https://github.com/ory/kratos/pull/1427)) -
      [@hackerman](https://github.com/aeneasr)
- [x] fix: deprecate sessionCookie
      ([kratos#1428](https://github.com/ory/kratos/pull/1428)) -
      [@hackerman](https://github.com/aeneasr)
- [x] fix: use STARTTLS for smtps connections
      ([kratos#1430](https://github.com/ory/kratos/pull/1430)) -
      [@hackerman](https://github.com/aeneasr)
- [x] fix: incorrect openapi specification for verification submission
      ([kratos#1431](https://github.com/ory/kratos/pull/1431)) -
      [@hackerman](https://github.com/aeneasr)

### [Feat](https://github.com/ory/kratos/labels/feat)

New feature or request.

#### Issues

- [x] Gracefully handle CSRF errors
      ([kratos#91](https://github.com/ory/kratos/issues/91)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Feature Request: Webhooks
      ([kratos#271](https://github.com/ory/kratos/issues/271))
- [x] Implement Security Questions MFA
      ([kratos#469](https://github.com/ory/kratos/issues/469))
- [x] Implement React SPA sample app
      ([kratos#668](https://github.com/ory/kratos/issues/668)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Double slash in URLs causes CSRF issues
      ([kratos#779](https://github.com/ory/kratos/issues/779))
- [x] CSRF failure should start a new login/registration flow
      ([kratos#821](https://github.com/ory/kratos/issues/821)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Prevent accidental leak of PII when Copy & Pasting of Flow URLs which
      include Flow IDs
      ([kratos#1282](https://github.com/ory/kratos/issues/1282)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Write tests for domain aliasing in cookie handler
      ([kratos#1292](https://github.com/ory/kratos/issues/1292)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Document new CSRF Cookie requirement
      ([kratos#1472](https://github.com/ory/kratos/issues/1472)) -
      [@hackerman](https://github.com/aeneasr)

#### Pull Requests

- [x] feat: APIs for native integration with AJAX / SPAs / ReactJS / NextJS /
      ... ([kratos#1367](https://github.com/ory/kratos/pull/1367)) -
      [@hackerman](https://github.com/aeneasr)
- [x] feat: anti-CSRF measures when fetching flows
      ([kratos#1458](https://github.com/ory/kratos/pull/1458)) -
      [@hackerman](https://github.com/aeneasr)

### [Rfc](https://github.com/ory/kratos/labels/rfc)

A request for comments to discuss and share ideas.

#### Issues

- [x] Prevent accidental leak of PII when Copy & Pasting of Flow URLs which
      include Flow IDs
      ([kratos#1282](https://github.com/ory/kratos/issues/1282)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Separate OpenAPI tags into stable and experimental and rework admin
      strategy ([kratos#1424](https://github.com/ory/kratos/issues/1424)) -
      [@hackerman](https://github.com/aeneasr)

### [Blocking](https://github.com/ory/kratos/labels/blocking)

Blocks milestones or other issues or pulls.

#### Issues

- [x] Document new CSRF Cookie requirement
      ([kratos#1472](https://github.com/ory/kratos/issues/1472)) -
      [@hackerman](https://github.com/aeneasr)

#### Pull Requests

- [x] feat: APIs for native integration with AJAX / SPAs / ReactJS / NextJS /
      ... ([kratos#1367](https://github.com/ory/kratos/pull/1367)) -
      [@hackerman](https://github.com/aeneasr)

## [v0.6.0-alpha.1](https://github.com/ory/kratos/milestone/8)

_This milestone does not have a description._

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [x] Sending JSON to complete oidc/password strategy flows causes CSRF issues
      ([kratos#378](https://github.com/ory/kratos/issues/378))
- [x] Password reset emails sent twice by each of the two kratos pods in my
      cluster ([kratos#652](https://github.com/ory/kratos/issues/652))
- [x] Building From Source fails
      ([kratos#711](https://github.com/ory/kratos/issues/711))
- [x] Quickstart is failing to mount volume kratos.yml when SELinux is enabled
      using Podman ([kratos#831](https://github.com/ory/kratos/issues/831)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Add randomized constant time to every login request
      ([kratos#832](https://github.com/ory/kratos/issues/832))
- [x] Kratos Admin API return 409 when createIdentity is called simultaneously
      ([kratos#861](https://github.com/ory/kratos/issues/861)) -
      [@Patrik](https://github.com/zepatrik)
- [x] `make sdk` is broken
      ([kratos#950](https://github.com/ory/kratos/issues/950)) -
      [@hackerman](https://github.com/aeneasr)
- [x] CLI navigation reference is broken
      ([kratos#951](https://github.com/ory/kratos/issues/951))

#### Pull Requests

- [x] Implement FIDO2/MFA and refactor flow payloads and identity credentials
      and authenticators
      ([kratos#921](https://github.com/ory/kratos/pull/921)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Umbrella PR for Ory Kratos v0.6 with MFA and improved flows (#961)
      ([kratos#1012](https://github.com/ory/kratos/pull/1012)) -
      [@hackerman](https://github.com/aeneasr)

### [Feat](https://github.com/ory/kratos/labels/feat)

New feature or request.

#### Issues

- [x] Support remote argon2 execution
      ([kratos#357](https://github.com/ory/kratos/issues/357)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Feature request: adjustable thresholds on how many times a password has
      been in a breach according to haveibeenpwned
      ([kratos#450](https://github.com/ory/kratos/issues/450))
- [x] Add return_to after logout
      ([kratos#702](https://github.com/ory/kratos/issues/702)) -
      [@Patrik](https://github.com/zepatrik)
- [x] Write CLI helper for recommending Argon2 parameters
      ([kratos#723](https://github.com/ory/kratos/issues/723)) -
      [@Patrik](https://github.com/zepatrik)
- [x] Add possibility to configure the "claims" query parameter in the auth_url
      of OIDC providers to request individial id_token claims
      ([kratos#735](https://github.com/ory/kratos/issues/735))
- [x] Replace viper with Koanf
      ([kratos#894](https://github.com/ory/kratos/issues/894)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Support dynamic return_to address on verification flow
      ([kratos#1123](https://github.com/ory/kratos/issues/1123))

#### Pull Requests

- [x] docs: Initial set of documentation tests
      ([kratos#567](https://github.com/ory/kratos/pull/567)) -
      [@hackerman](https://github.com/aeneasr)
- [x] feat: add selinux compatible quickstart config
      ([kratos#889](https://github.com/ory/kratos/pull/889)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Umbrella PR for Ory Kratos v0.6 with MFA and improved flows (#961)
      ([kratos#1012](https://github.com/ory/kratos/pull/1012)) -
      [@hackerman](https://github.com/aeneasr)

### [Rfc](https://github.com/ory/kratos/labels/rfc)

A request for comments to discuss and share ideas.

#### Issues

- [x] Refactor form builder
      ([kratos#929](https://github.com/ory/kratos/issues/929)) -
      [@hackerman](https://github.com/aeneasr)

### [Blocking](https://github.com/ory/kratos/labels/blocking)

Blocks milestones or other issues or pulls.

#### Issues

- [x] Ory Kratos v0.6 pre-release list
      ([kratos#1222](https://github.com/ory/kratos/issues/1222)) -
      [@hackerman](https://github.com/aeneasr)

#### Pull Requests

- [x] Umbrella PR for Ory Kratos v0.6 with MFA and improved flows (#961)
      ([kratos#1012](https://github.com/ory/kratos/pull/1012)) -
      [@hackerman](https://github.com/aeneasr)
