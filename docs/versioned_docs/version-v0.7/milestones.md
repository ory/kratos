---
id: milestones
title: Milestones and Roadmap
---

## [v0.7.1-alpha.1](https://github.com/ory/kratos/milestone/12)

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

### [Docs](https://github.com/ory/kratos/labels/docs)

Affects documentation.

#### Issues

- [ ] Different payload for stub:500
      ([kratos#1568](https://github.com/ory/kratos/issues/1568))
- [x] Update CSRF pitfall section for admin endpoints
      ([kratos#1557](https://github.com/ory/kratos/issues/1557)) -
      [@hackerman](https://github.com/aeneasr)

## [v0.9.0-alpha.1](https://github.com/ory/kratos/milestone/11)

_This milestone does not have a description._

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [ ] Refresh Sessions Without Having to Log In Again
      ([kratos#615](https://github.com/ory/kratos/issues/615)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Fetching a settings request after error is missing identity data
      ([kratos#689](https://github.com/ory/kratos/issues/689)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Feature Request: Have access to username in email templates
      ([kratos#925](https://github.com/ory/kratos/issues/925))
- [ ] Implement email TTL for non-working/non-existant emails
      ([kratos#944](https://github.com/ory/kratos/issues/944)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] panic: a handle is already registered for path
      '/self-service/recovery/methods/link'
      ([kratos#1068](https://github.com/ory/kratos/issues/1068))
- [ ] Recovery link doesn't log out existing users
      ([kratos#1077](https://github.com/ory/kratos/issues/1077))
- [ ] recovery link is open by email app
      ([kratos#1142](https://github.com/ory/kratos/issues/1142))
- [ ] Email verification link is automatically opened by mail provider
      ([kratos#1202](https://github.com/ory/kratos/issues/1202))
- [ ] Relative path in ui_url or default_browser_return_url cause runtime crash
      ([kratos#1446](https://github.com/ory/kratos/issues/1446))
- [ ] /schemas/default does not work with base64 identity schemas
      ([kratos#1529](https://github.com/ory/kratos/issues/1529))

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
- [ ] Support email verification paswordless login
      ([kratos#286](https://github.com/ory/kratos/issues/286))
- [ ] Prevent account enumeration for profile updates
      ([kratos#292](https://github.com/ory/kratos/issues/292)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] SMTP Error spams the server logs
      ([kratos#402](https://github.com/ory/kratos/issues/402))
- [ ] How to sign in with Twitter
      ([kratos#517](https://github.com/ory/kratos/issues/517))
- [ ] Selfservice account deletion
      ([kratos#596](https://github.com/ory/kratos/issues/596))
- [ ] Add ability to import user credentials
      ([kratos#605](https://github.com/ory/kratos/issues/605)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Throttling repeated login requests
      ([kratos#654](https://github.com/ory/kratos/issues/654))
- [ ] Admin/Selfservice session management
      ([kratos#655](https://github.com/ory/kratos/issues/655)) -
      [@Alano Terblanche](https://github.com/Benehiko)
- [ ] Require identity deactivation before administrative deletion
      ([kratos#657](https://github.com/ory/kratos/issues/657))
- [ ] Self-service GDPR identity export
      ([kratos#658](https://github.com/ory/kratos/issues/658))
- [ ] Rename strategy to method in internal APIs and Documentation
      ([kratos#683](https://github.com/ory/kratos/issues/683)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Validate identity schema on load
      ([kratos#701](https://github.com/ory/kratos/issues/701)) -
      [@Alano Terblanche](https://github.com/Benehiko)
- [ ] Support OAuth2 and OpenID Connect in API-based flows
      ([kratos#707](https://github.com/ory/kratos/issues/707))
- [ ] improve multi schema handling in different auth flows
      ([kratos#765](https://github.com/ory/kratos/issues/765))
- [ ] Add i18n support to mail templates
      ([kratos#834](https://github.com/ory/kratos/issues/834))
- [ ] Add option for disabling registration
      ([kratos#882](https://github.com/ory/kratos/issues/882)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Password policies: make api.pwnedpasswords.com configurable
      ([kratos#969](https://github.com/ory/kratos/issues/969))
- [ ] Access identity schema information (e.g. usernames) within the jsonnet
      mapper ([kratos#987](https://github.com/ory/kratos/issues/987))
- [ ] login ui returned without flowid
      ([kratos#1055](https://github.com/ory/kratos/issues/1055))
- [ ] Allow account recovery for identities without email address
      ([kratos#1419](https://github.com/ory/kratos/issues/1419))
- [ ] Write new recovery and verification strategy that does not rely on links
      ([kratos#1451](https://github.com/ory/kratos/issues/1451))
- [ ] Field validation answer status code 422 instead of 400
      ([kratos#1559](https://github.com/ory/kratos/issues/1559))

### [Docs](https://github.com/ory/kratos/labels/docs)

Affects documentation.

#### Issues

- [ ] Document that identity information (traits, etc) are available to token
      holders and backend systems
      ([kratos#43](https://github.com/ory/kratos/issues/43)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] Config JSON Schema needs example values
      ([kratos#179](https://github.com/ory/kratos/issues/179)) -
      [@hackerman](https://github.com/aeneasr)

### [Rfc](https://github.com/ory/kratos/labels/rfc)

A request for comments to discuss and share ideas.

#### Issues

- [ ] Introduce prevent extension in Identity JSON schema
      ([kratos#47](https://github.com/ory/kratos/issues/47)) -
      [@hackerman](https://github.com/aeneasr)
- [ ] improve multi schema handling in different auth flows
      ([kratos#765](https://github.com/ory/kratos/issues/765))
- [ ] Field validation answer status code 422 instead of 400
      ([kratos#1559](https://github.com/ory/kratos/issues/1559))

### [Blocking](https://github.com/ory/kratos/labels/blocking)

Blocks milestones or other issues or pulls.

#### Issues

- [ ] Implement Hydra integration
      ([kratos#273](https://github.com/ory/kratos/issues/273))

## [v0.8.0-alpha.1](https://github.com/ory/kratos/milestone/10)

This milestone focuses on MFA with TOTP.

### [Bug](https://github.com/ory/kratos/labels/bug)

Something is not working.

#### Issues

- [x] Unmable to use Auth0 as a generic OIDC provider
      ([kratos#609](https://github.com/ory/kratos/issues/609))
- [x] Typescript ErrorContainer type is incorrect
      ([kratos#782](https://github.com/ory/kratos/issues/782))

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

### [Docs](https://github.com/ory/kratos/labels/docs)

Affects documentation.

#### Issues

- [x] Include release notes in CHANGELOG.md
      ([kratos#1442](https://github.com/ory/kratos/issues/1442)) -
      [@hackerman](https://github.com/aeneasr)
- [x] Include changelog in docs navigation
      ([kratos#1443](https://github.com/ory/kratos/issues/1443)) -
      [@hackerman](https://github.com/aeneasr)
