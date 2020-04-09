<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Changelog](#changelog)
  - [Unreleased](#unreleased)
  - [v0.1.1-alpha.1 (2020-02-18)](#v011-alpha1-2020-02-18)
  - [v0.1.0-alpha.6 (2020-02-16)](#v010-alpha6-2020-02-16)
  - [v0.1.0-alpha.5 (2020-02-06)](#v010-alpha5-2020-02-06)
  - [v0.1.0-alpha.4 (2020-02-06)](#v010-alpha4-2020-02-06)
  - [v0.1.0-alpha.3 (2020-02-06)](#v010-alpha3-2020-02-06)
  - [v0.1.0-alpha.2 (2020-02-03)](#v010-alpha2-2020-02-03)
  - [v0.1.0-alpha.1 (2020-01-31)](#v010-alpha1-2020-01-31)
  - [v0.0.3-alpha.15 (2020-01-31)](#v003-alpha15-2020-01-31)
  - [v0.0.3-alpha.14 (2020-01-31)](#v003-alpha14-2020-01-31)
  - [v0.0.3-alpha.12 (2020-01-31)](#v003-alpha12-2020-01-31)
  - [v0.0.3-alpha.13 (2020-01-31)](#v003-alpha13-2020-01-31)
  - [v0.0.3-alpha.11 (2020-01-31)](#v003-alpha11-2020-01-31)
  - [v0.0.3-alpha.10 (2020-01-31)](#v003-alpha10-2020-01-31)
  - [v0.0.3-alpha.8+oryOS.15 (2020-01-30)](#v003-alpha8oryos15-2020-01-30)
  - [v0.0.3-alpha.9 (2020-01-30)](#v003-alpha9-2020-01-30)
  - [v0.0.3-alpha.7 (2020-01-30)](#v003-alpha7-2020-01-30)
  - [v0.0.3-alpha.5 (2020-01-30)](#v003-alpha5-2020-01-30)
  - [v0.0.3-alpha.4 (2020-01-30)](#v003-alpha4-2020-01-30)
  - [v0.0.3-alpha.3 (2020-01-30)](#v003-alpha3-2020-01-30)
  - [v0.0.3-alpha.2 (2020-01-30)](#v003-alpha2-2020-01-30)
  - [v0.0.3-alpha.1 (2020-01-30)](#v003-alpha1-2020-01-30)
  - [v0.0.1-alpha.9 (2020-01-29)](#v001-alpha9-2020-01-29)
  - [v0.0.1-alpha.10+oryOS.15 (2020-01-29)](#v001-alpha10oryos15-2020-01-29)
  - [v0.0.1-alpha.11 (2020-01-29)](#v001-alpha11-2020-01-29)
  - [v0.0.1-alpha.7 (2020-01-29)](#v001-alpha7-2020-01-29)
  - [v0.0.1-alpha.8 (2020-01-29)](#v001-alpha8-2020-01-29)
  - [v0.0.2-alpha.1 (2020-01-29)](#v002-alpha1-2020-01-29)
  - [v0.0.1-alpha.6 (2020-01-29)](#v001-alpha6-2020-01-29)
  - [v0.0.1-alpha.5 (2020-01-29)](#v001-alpha5-2020-01-29)
  - [v0.0.1-alpha.3 (2020-01-28)](#v001-alpha3-2020-01-28)
  - [v0.0.1-alpha.2 (2020-01-28)](#v001-alpha2-2020-01-28)
  - [v0.0.1-alpha.1 (2020-01-28)](#v001-alpha1-2020-01-28)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Changelog

## [Unreleased](https://github.com/ory/kratos/tree/HEAD)

[Full Changelog](https://github.com/ory/kratos/compare/v0.1.1-alpha.1...HEAD)

**Implemented enhancements:**

- Have `dsn: memory` as an alias for SQLite in memory DSN [\#228](https://github.com/ory/kratos/issues/228)

**Fixed bugs:**

- shutdown doesn't complete [\#295](https://github.com/ory/kratos/issues/295)
- Same version of migration version 20191100000010 caused test failure [\#279](https://github.com/ory/kratos/issues/279)
- Email Verification Error when using PostgreSQL [\#269](https://github.com/ory/kratos/issues/269)
- HBIP check hangs when connection is slow or ends with a network error [\#261](https://github.com/ory/kratos/issues/261)
- Investigate MySQL empty timestamp issue on session [\#244](https://github.com/ory/kratos/issues/244)
- Return REST error when fetching expired login/registration/profile request [\#235](https://github.com/ory/kratos/issues/235)
- fix\(swagger\): Move nolint,deadcode instructions to own file [\#293](https://github.com/ory/kratos/pull/293) ([aeneasr](https://github.com/aeneasr))
- feat: Enable CockroachDB integration [\#260](https://github.com/ory/kratos/pull/260) ([aeneasr](https://github.com/aeneasr))
- fix: Resolve NULL value for seen\_at [\#259](https://github.com/ory/kratos/pull/259) ([aeneasr](https://github.com/aeneasr))

**Security fixes:**

- Regenerate CSRF Tokens on principal change [\#217](https://github.com/ory/kratos/issues/217)
- Implement Password Strength Meter API [\#136](https://github.com/ory/kratos/issues/136)

**Closed issues:**

- Quickstart app doubt for authentication only\(not authorization\) with mysql database  [\#297](https://github.com/ory/kratos/issues/297)
- Serve the schemas in the common API and have it documented [\#287](https://github.com/ory/kratos/issues/287)
- Quickstart broken, db.sqlite not writabel [\#281](https://github.com/ory/kratos/issues/281)
- Viper key for SMTP from address appears to be incorrect [\#277](https://github.com/ory/kratos/issues/277)
- MailSlurper is not sending the verification email [\#264](https://github.com/ory/kratos/issues/264)
- SQLite database errors in quickstart [\#263](https://github.com/ory/kratos/issues/263)
- Allow configuration of same-site cookie [\#257](https://github.com/ory/kratos/issues/257)
- CSRF token is missing or invalid [\#250](https://github.com/ory/kratos/issues/250)
- Enable CockroachDB test suite and integration [\#132](https://github.com/ory/kratos/issues/132)

**Merged pull requests:**

- chore: update docusaurus template [\#324](https://github.com/ory/kratos/pull/324) ([aeneasr](https://github.com/aeneasr))
- fix writing [\#322](https://github.com/ory/kratos/pull/322) ([gwind](https://github.com/gwind))
- chore: update docusaurus template [\#321](https://github.com/ory/kratos/pull/321) ([aeneasr](https://github.com/aeneasr))
- chore: update docusaurus template [\#320](https://github.com/ory/kratos/pull/320) ([aeneasr](https://github.com/aeneasr))
- chore: update docusaurus template [\#319](https://github.com/ory/kratos/pull/319) ([aeneasr](https://github.com/aeneasr))
- chore: update docusaurus template [\#318](https://github.com/ory/kratos/pull/318) ([aeneasr](https://github.com/aeneasr))
- refactor: move docs to this repository [\#317](https://github.com/ory/kratos/pull/317) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#315](https://github.com/ory/kratos/pull/315) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#314](https://github.com/ory/kratos/pull/314) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#313](https://github.com/ory/kratos/pull/313) ([aeneasr](https://github.com/aeneasr))
- chore: bump ory/x to have csv parsing from env vars [\#312](https://github.com/ory/kratos/pull/312) ([zepatrik](https://github.com/zepatrik))
- fix: move to ory sqa service [\#309](https://github.com/ory/kratos/pull/309) ([aeneasr](https://github.com/aeneasr))
- Fix the query parameter name in the get self-service error endpoint in API docs [\#308](https://github.com/ory/kratos/pull/308) ([sandhose](https://github.com/sandhose))
- chore: moved watchAndValidateViper to viperx [\#307](https://github.com/ory/kratos/pull/307) ([zepatrik](https://github.com/zepatrik))
- chore: update ory/x dependency and add test case [\#305](https://github.com/ory/kratos/pull/305) ([zepatrik](https://github.com/zepatrik))
- feat: allow configuring same-site for session cookies [\#303](https://github.com/ory/kratos/pull/303) ([zepatrik](https://github.com/zepatrik))
- fix: Linux install script [\#302](https://github.com/ory/kratos/pull/302) ([guillett](https://github.com/guillett))
- Document the schema API and serve it in the admin API [\#299](https://github.com/ory/kratos/pull/299) ([sandhose](https://github.com/sandhose))
- docs: Updates issue and pull request templates [\#298](https://github.com/ory/kratos/pull/298) ([aeneasr](https://github.com/aeneasr))
- fix:add graceful shutdown to courier handler [\#296](https://github.com/ory/kratos/pull/296) ([Gibheer](https://github.com/Gibheer))
- fix\(session\): Regenerate CSRF Token on principal change [\#290](https://github.com/ory/kratos/pull/290) ([aeneasr](https://github.com/aeneasr))
- feat: Return 410 when selfservice requests expire [\#289](https://github.com/ory/kratos/pull/289) ([aeneasr](https://github.com/aeneasr))
- fix: Use resilient client for HIBP lookup [\#288](https://github.com/ory/kratos/pull/288) ([aeneasr](https://github.com/aeneasr))
- Revert "fix: Use host volume mount for sqlite" [\#285](https://github.com/ory/kratos/pull/285) ([aeneasr](https://github.com/aeneasr))
- feat: add `dsn: memory` shorthand [\#284](https://github.com/ory/kratos/pull/284) ([zepatrik](https://github.com/zepatrik))
- fix\(session\): whoami endpoint now supports all HTTP methods [\#283](https://github.com/ory/kratos/pull/283) ([aeneasr](https://github.com/aeneasr))
- fix: rename migrations with same version [\#280](https://github.com/ory/kratos/pull/280) ([zepatrik](https://github.com/zepatrik))
- Fix smtp/stmp typo [\#278](https://github.com/ory/kratos/pull/278) ([jdnurmi](https://github.com/jdnurmi))
- fix\(sql/migrations\): change type of courier\_message.body to "text" [\#276](https://github.com/ory/kratos/pull/276) ([zepatrik](https://github.com/zepatrik))
- fix: Use host volume mount for sqlite [\#272](https://github.com/ory/kratos/pull/272) ([aeneasr](https://github.com/aeneasr))
- feat\(selfService/profile\): enable updating auth related traits [\#266](https://github.com/ory/kratos/pull/266) ([zepatrik](https://github.com/zepatrik))
- docs: Typo in README.md [\#265](https://github.com/ory/kratos/pull/265) ([cuttlefish](https://github.com/cuttlefish))
- feat\(selfservice/login\): enable reauthentication functionality [\#248](https://github.com/ory/kratos/pull/248) ([zepatrik](https://github.com/zepatrik))

## [v0.1.1-alpha.1](https://github.com/ory/kratos/tree/v0.1.1-alpha.1) (2020-02-18)

[Full Changelog](https://github.com/ory/kratos/compare/v0.1.0-alpha.6...v0.1.1-alpha.1)

**Fixed bugs:**

- fix: Resolve several verification problems [\#253](https://github.com/ory/kratos/pull/253) ([aeneasr](https://github.com/aeneasr))

**Merged pull requests:**

- fix: Update verify URLs [\#258](https://github.com/ory/kratos/pull/258) ([aeneasr](https://github.com/aeneasr))
- fix: Clean up docker quickstart [\#255](https://github.com/ory/kratos/pull/255) ([aeneasr](https://github.com/aeneasr))
- refactor\(persistence/sql\): move connection to context to enable transactions [\#254](https://github.com/ory/kratos/pull/254) ([zepatrik](https://github.com/zepatrik))
- fix: Add verify return to address [\#252](https://github.com/ory/kratos/pull/252) ([aeneasr](https://github.com/aeneasr))

## [v0.1.0-alpha.6](https://github.com/ory/kratos/tree/v0.1.0-alpha.6) (2020-02-16)

[Full Changelog](https://github.com/ory/kratos/compare/v0.1.0-alpha.5...v0.1.0-alpha.6)

**Implemented enhancements:**

- Make OIDC strategy capable of dealing with expiry errors [\#233](https://github.com/ory/kratos/issues/233)
- selfservice/updateProfileFlow: disable form fields that the user is not allowed to update [\#227](https://github.com/ory/kratos/issues/227)
- Use jsonschema everywhere [\#225](https://github.com/ory/kratos/issues/225)
- Implement Verification [\#27](https://github.com/ory/kratos/issues/27)
- feat: Implement email verification [\#245](https://github.com/ory/kratos/pull/245) ([aeneasr](https://github.com/aeneasr))

**Fixed bugs:**

- Mark fields required in login / registration methods [\#234](https://github.com/ory/kratos/issues/234)
- fix: Set AuthenticatedAt in session issuer hook [\#246](https://github.com/ory/kratos/pull/246) ([aeneasr](https://github.com/aeneasr))
- Resolve flaky SDK generation issues caused by UUID [\#240](https://github.com/ory/kratos/pull/240) ([aeneasr](https://github.com/aeneasr))

**Closed issues:**

- Require Levenshtein distance between identifiers and password [\#184](https://github.com/ory/kratos/issues/184)

**Merged pull requests:**

- feat: Add verification to quickstart [\#251](https://github.com/ory/kratos/pull/251) ([aeneasr](https://github.com/aeneasr))
- fix: Adapt quickstart to verify changes [\#247](https://github.com/ory/kratos/pull/247) ([aeneasr](https://github.com/aeneasr))
- fix\(SelfService/Strategy/oidc\): rework auth session expiry  [\#242](https://github.com/ory/kratos/pull/242) ([zepatrik](https://github.com/zepatrik))
- feat\(selfservice/profile\): Add disabled flag to identifier form fields [\#238](https://github.com/ory/kratos/pull/238) ([zepatrik](https://github.com/zepatrik))
- fix\(swagger\): Use correct annotations for request methods [\#237](https://github.com/ory/kratos/pull/237) ([aeneasr](https://github.com/aeneasr))
- feat: add levenshtein distance check for password validation [\#231](https://github.com/ory/kratos/pull/231) ([zepatrik](https://github.com/zepatrik))
- Use ory/jsonschema/v3 everywhere [\#229](https://github.com/ory/kratos/pull/229) ([aeneasr](https://github.com/aeneasr))

## [v0.1.0-alpha.5](https://github.com/ory/kratos/tree/v0.1.0-alpha.5) (2020-02-06)

[Full Changelog](https://github.com/ory/kratos/compare/v0.1.0-alpha.4...v0.1.0-alpha.5)

**Fixed bugs:**

- Mitigate expired login and registration requests [\#96](https://github.com/ory/kratos/issues/96)

**Merged pull requests:**

- feat: redirect to new auth session on expired auth sessions [\#230](https://github.com/ory/kratos/pull/230) ([zepatrik](https://github.com/zepatrik))

## [v0.1.0-alpha.4](https://github.com/ory/kratos/tree/v0.1.0-alpha.4) (2020-02-06)

[Full Changelog](https://github.com/ory/kratos/compare/v0.1.0-alpha.3...v0.1.0-alpha.4)

## [v0.1.0-alpha.3](https://github.com/ory/kratos/tree/v0.1.0-alpha.3) (2020-02-06)

[Full Changelog](https://github.com/ory/kratos/compare/v0.1.0-alpha.2...v0.1.0-alpha.3)

## [v0.1.0-alpha.2](https://github.com/ory/kratos/tree/v0.1.0-alpha.2) (2020-02-03)

[Full Changelog](https://github.com/ory/kratos/compare/v0.1.0-alpha.1...v0.1.0-alpha.2)

**Implemented enhancements:**

- Rework errors API [\#204](https://github.com/ory/kratos/issues/204)

**Fixed bugs:**

- refactor!: Improve user-facing error APIs [\#219](https://github.com/ory/kratos/pull/219) ([aeneasr](https://github.com/aeneasr))

**Closed issues:**

- Discrepancy in documentation [\#218](https://github.com/ory/kratos/issues/218)
- Support identity impersonation [\#201](https://github.com/ory/kratos/issues/201)
- Implement `--dev` flag [\#36](https://github.com/ory/kratos/issues/36)

**Merged pull requests:**

- Serve: add admin /self-service/errors route [\#226](https://github.com/ory/kratos/pull/226) ([zepatrik](https://github.com/zepatrik))
- fix: Set csrf token on public endpoints [\#224](https://github.com/ory/kratos/pull/224) ([aeneasr](https://github.com/aeneasr))
- ci: Switch to golangci orb [\#223](https://github.com/ory/kratos/pull/223) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#222](https://github.com/ory/kratos/pull/222) ([aeneasr](https://github.com/aeneasr))
- ci: Bump sdk and changelog versions [\#221](https://github.com/ory/kratos/pull/221) ([aeneasr](https://github.com/aeneasr))
- feat: Override semantic config [\#220](https://github.com/ory/kratos/pull/220) ([aeneasr](https://github.com/aeneasr))
- Add paths to sqa middleware [\#216](https://github.com/ory/kratos/pull/216) ([aeneasr](https://github.com/aeneasr))

## [v0.1.0-alpha.1](https://github.com/ory/kratos/tree/v0.1.0-alpha.1) (2020-01-31)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.15...v0.1.0-alpha.1)

**Implemented enhancements:**

- FormFields for Login, Registration, Profile requests should be array and not maps [\#186](https://github.com/ory/kratos/issues/186)
- Adopt CircleCI orbs for SDK, goreleaser, changelog [\#166](https://github.com/ory/kratos/issues/166)
- Improve `--dev` flag [\#162](https://github.com/ory/kratos/issues/162)
- Reintroduce SQL Migration plans [\#131](https://github.com/ory/kratos/issues/131)
- Refactor DBAL layer [\#128](https://github.com/ory/kratos/issues/128)
- Populate registration form with data from JSON Schema [\#120](https://github.com/ory/kratos/issues/120)
- Implement selfservice profile management [\#112](https://github.com/ory/kratos/issues/112)
- Implement SQL backend for errorx package [\#92](https://github.com/ory/kratos/issues/92)
- Schemas should be mirrored by hive at some well-known url [\#86](https://github.com/ory/kratos/issues/86)
- Add health endpoints [\#82](https://github.com/ory/kratos/issues/82)
- Check all credentials for uniqueness to support uniqueness in passwordless flows [\#78](https://github.com/ory/kratos/issues/78)
- Implement persistent DBAL using Postgres [\#66](https://github.com/ory/kratos/issues/66)
- Disable login and registration when session exists [\#63](https://github.com/ory/kratos/issues/63)
- Implement Admin CRUD for Identities [\#58](https://github.com/ory/kratos/issues/58)
- Write test for missing data during sign up with oidc [\#55](https://github.com/ory/kratos/issues/55)
- Write tests for selfservice.ErrorHandler [\#54](https://github.com/ory/kratos/issues/54)
- Add continuous integration [\#53](https://github.com/ory/kratos/issues/53)
- Support object stubs in form payloads [\#45](https://github.com/ory/kratos/issues/45)
- Implement form-based, self-service login and registration [\#29](https://github.com/ory/kratos/issues/29)
- Rework public and admin fetch strategy [\#203](https://github.com/ory/kratos/pull/203) ([aeneasr](https://github.com/aeneasr))
- Update HTTP routes for a consistent API naming [\#199](https://github.com/ory/kratos/pull/199) ([aeneasr](https://github.com/aeneasr))
- ss: Use JSON Schema to type assert form body [\#116](https://github.com/ory/kratos/pull/116) ([aeneasr](https://github.com/aeneasr))

**Fixed bugs:**

- Improve `/profile` and `/session` URLs [\#195](https://github.com/ory/kratos/issues/195)
- Profile Management requests sends Request ID in POST Body instead of Query Parameter [\#190](https://github.com/ory/kratos/issues/190)
- Key `traits\_schema\_url` not populated in profile management request [\#189](https://github.com/ory/kratos/issues/189)
- Update Quickstart Access Rules to include new CSS files for sample app [\#188](https://github.com/ory/kratos/issues/188)
- Send right field type in registration/login request information [\#175](https://github.com/ory/kratos/issues/175)
- Ensure that decoderx works with checkboxes [\#125](https://github.com/ory/kratos/issues/125)
- OIDC Credentials do not allow multiple connections [\#114](https://github.com/ory/kratos/issues/114)
- selfservice: Form BodyParser should assert types using JSON Schema [\#109](https://github.com/ory/kratos/issues/109)
- Using only numbers as password during sign up leads to error [\#98](https://github.com/ory/kratos/issues/98)
- Irrecoverable state when "securecookie" fails. [\#97](https://github.com/ory/kratos/issues/97)
- Do not echo headers in login/register request response [\#95](https://github.com/ory/kratos/issues/95)
- Registration values are not properly propagated [\#71](https://github.com/ory/kratos/issues/71)
- CSRF is broken for social sign in [\#68](https://github.com/ory/kratos/issues/68)
- Reset CSRF Token on Principal Change \(Sign Out\) [\#38](https://github.com/ory/kratos/issues/38)
- selfservice: Omit request header from login/registration request [\#106](https://github.com/ory/kratos/pull/106) ([aeneasr](https://github.com/aeneasr))
- selfservice: Explicitly whitelist form parser keys [\#105](https://github.com/ory/kratos/pull/105) ([aeneasr](https://github.com/aeneasr))

**Security fixes:**

- Rethink public fetch request protection [\#122](https://github.com/ory/kratos/issues/122)
- Prevent credentials from being filled in without validation [\#46](https://github.com/ory/kratos/issues/46)

**Closed issues:**

- OIDC method has "request" field in the form [\#180](https://github.com/ory/kratos/issues/180)
- Schemas should be tested [\#165](https://github.com/ory/kratos/issues/165)
- JSON Schema `ory.sh/kratos` keyword extension design document [\#118](https://github.com/ory/kratos/issues/118)
- Decide JSON Schema custom keyword prefix for custom logic [\#115](https://github.com/ory/kratos/issues/115)
- Implement profile and credentials management [\#108](https://github.com/ory/kratos/issues/108)
- hermes: Notification architecture [\#99](https://github.com/ory/kratos/issues/99)
- Omit oidc client secret, cookie secret, and dsn from k8s configmap [\#88](https://github.com/ory/kratos/issues/88)
- Document Self-Service state machine [\#52](https://github.com/ory/kratos/issues/52)
- Document how the form parser works [\#41](https://github.com/ory/kratos/issues/41)
- docs: Document that the password strategy lowercases identifiers [\#25](https://github.com/ory/kratos/issues/25)
- Dealing with missing data when using 3rd-party registration [\#23](https://github.com/ory/kratos/issues/23)

**Merged pull requests:**

- docs: Updates issue and pull request templates [\#215](https://github.com/ory/kratos/pull/215) ([aeneasr](https://github.com/aeneasr))
- Clean up and resolve packr2 issues [\#211](https://github.com/ory/kratos/pull/211) ([aeneasr](https://github.com/aeneasr))
- Resolve goreleaser build issues [\#208](https://github.com/ory/kratos/pull/208) ([aeneasr](https://github.com/aeneasr))
- ci: Bump sdk orb [\#206](https://github.com/ory/kratos/pull/206) ([aeneasr](https://github.com/aeneasr))
- ss/oidc: Remove obsolete request field from form [\#193](https://github.com/ory/kratos/pull/193) ([aeneasr](https://github.com/aeneasr))
- sql: Re-introduce migration plans to CLI command [\#192](https://github.com/ory/kratos/pull/192) ([aeneasr](https://github.com/aeneasr))
- courier: Implement message templates and SMTP delivery [\#146](https://github.com/ory/kratos/pull/146) ([aeneasr](https://github.com/aeneasr))
- Implement base features for v0.0.1 release [\#102](https://github.com/ory/kratos/pull/102) ([aeneasr](https://github.com/aeneasr))

## [v0.0.3-alpha.15](https://github.com/ory/kratos/tree/v0.0.3-alpha.15) (2020-01-31)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.14...v0.0.3-alpha.15)

## [v0.0.3-alpha.14](https://github.com/ory/kratos/tree/v0.0.3-alpha.14) (2020-01-31)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.12...v0.0.3-alpha.14)

## [v0.0.3-alpha.12](https://github.com/ory/kratos/tree/v0.0.3-alpha.12) (2020-01-31)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.13...v0.0.3-alpha.12)

## [v0.0.3-alpha.13](https://github.com/ory/kratos/tree/v0.0.3-alpha.13) (2020-01-31)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.11...v0.0.3-alpha.13)

**Merged pull requests:**

- Allow mounting SQLite in /home/ory/sqlite [\#212](https://github.com/ory/kratos/pull/212) ([aeneasr](https://github.com/aeneasr))

## [v0.0.3-alpha.11](https://github.com/ory/kratos/tree/v0.0.3-alpha.11) (2020-01-31)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.10...v0.0.3-alpha.11)

**Merged pull requests:**

- Improve field types [\#209](https://github.com/ory/kratos/pull/209) ([zepatrik](https://github.com/zepatrik))

## [v0.0.3-alpha.10](https://github.com/ory/kratos/tree/v0.0.3-alpha.10) (2020-01-31)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.8+oryOS.15...v0.0.3-alpha.10)

## [v0.0.3-alpha.8+oryOS.15](https://github.com/ory/kratos/tree/v0.0.3-alpha.8+oryOS.15) (2020-01-30)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.9...v0.0.3-alpha.8+oryOS.15)

## [v0.0.3-alpha.9](https://github.com/ory/kratos/tree/v0.0.3-alpha.9) (2020-01-30)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.7...v0.0.3-alpha.9)

## [v0.0.3-alpha.7](https://github.com/ory/kratos/tree/v0.0.3-alpha.7) (2020-01-30)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.5...v0.0.3-alpha.7)

## [v0.0.3-alpha.5](https://github.com/ory/kratos/tree/v0.0.3-alpha.5) (2020-01-30)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.4...v0.0.3-alpha.5)

**Merged pull requests:**

- ci: Resolve final docker build issues [\#210](https://github.com/ory/kratos/pull/210) ([aeneasr](https://github.com/aeneasr))

## [v0.0.3-alpha.4](https://github.com/ory/kratos/tree/v0.0.3-alpha.4) (2020-01-30)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.3...v0.0.3-alpha.4)

## [v0.0.3-alpha.3](https://github.com/ory/kratos/tree/v0.0.3-alpha.3) (2020-01-30)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.2...v0.0.3-alpha.3)

## [v0.0.3-alpha.2](https://github.com/ory/kratos/tree/v0.0.3-alpha.2) (2020-01-30)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.3-alpha.1...v0.0.3-alpha.2)

## [v0.0.3-alpha.1](https://github.com/ory/kratos/tree/v0.0.3-alpha.1) (2020-01-30)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.9...v0.0.3-alpha.1)

**Fixed bugs:**

- Order registration/login form fields according to schema [\#176](https://github.com/ory/kratos/issues/176)

**Merged pull requests:**

- Update quickstart [\#207](https://github.com/ory/kratos/pull/207) ([aeneasr](https://github.com/aeneasr))
- Take field order from schema [\#205](https://github.com/ory/kratos/pull/205) ([zepatrik](https://github.com/zepatrik))
- ss/profile: Use request ID as query param everywhere [\#202](https://github.com/ory/kratos/pull/202) ([aeneasr](https://github.com/aeneasr))

## [v0.0.1-alpha.9](https://github.com/ory/kratos/tree/v0.0.1-alpha.9) (2020-01-29)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.10+oryOS.15...v0.0.1-alpha.9)

## [v0.0.1-alpha.10+oryOS.15](https://github.com/ory/kratos/tree/v0.0.1-alpha.10+oryOS.15) (2020-01-29)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.11...v0.0.1-alpha.10+oryOS.15)

## [v0.0.1-alpha.11](https://github.com/ory/kratos/tree/v0.0.1-alpha.11) (2020-01-29)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.7...v0.0.1-alpha.11)

## [v0.0.1-alpha.7](https://github.com/ory/kratos/tree/v0.0.1-alpha.7) (2020-01-29)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.8...v0.0.1-alpha.7)

## [v0.0.1-alpha.8](https://github.com/ory/kratos/tree/v0.0.1-alpha.8) (2020-01-29)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.2-alpha.1...v0.0.1-alpha.8)

## [v0.0.2-alpha.1](https://github.com/ory/kratos/tree/v0.0.2-alpha.1) (2020-01-29)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.6...v0.0.2-alpha.1)

## [v0.0.1-alpha.6](https://github.com/ory/kratos/tree/v0.0.1-alpha.6) (2020-01-29)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.5...v0.0.1-alpha.6)

## [v0.0.1-alpha.5](https://github.com/ory/kratos/tree/v0.0.1-alpha.5) (2020-01-29)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.3...v0.0.1-alpha.5)

**Closed issues:**

- Issue with the quickstart build failure [\#198](https://github.com/ory/kratos/issues/198)

**Merged pull requests:**

- Make form fields an array [\#197](https://github.com/ory/kratos/pull/197) ([zepatrik](https://github.com/zepatrik))
- Resolve build issues with CGO [\#196](https://github.com/ory/kratos/pull/196) ([aeneasr](https://github.com/aeneasr))

## [v0.0.1-alpha.3](https://github.com/ory/kratos/tree/v0.0.1-alpha.3) (2020-01-28)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.2...v0.0.1-alpha.3)

## [v0.0.1-alpha.2](https://github.com/ory/kratos/tree/v0.0.1-alpha.2) (2020-01-28)

[Full Changelog](https://github.com/ory/kratos/compare/v0.0.1-alpha.1...v0.0.1-alpha.2)

## [v0.0.1-alpha.1](https://github.com/ory/kratos/tree/v0.0.1-alpha.1) (2020-01-28)

[Full Changelog](https://github.com/ory/kratos/compare/ab6f24a85276bdd8687f2fc06390c1279892b005...v0.0.1-alpha.1)

**Fixed bugs:**

- Contain security context for reading schemas from disk [\#163](https://github.com/ory/kratos/issues/163)
- strategy/oidc: Allow multiple OIDC Connections [\#191](https://github.com/ory/kratos/pull/191) ([aeneasr](https://github.com/aeneasr))

**Closed issues:**

- Registration/Login form fields should not include "request" [\#178](https://github.com/ory/kratos/issues/178)
- Fix broken CI test pipeline [\#151](https://github.com/ory/kratos/issues/151)
- Seprate out login & registeration POST hooks [\#149](https://github.com/ory/kratos/issues/149)
- Optionally allow only one active session per identity [\#139](https://github.com/ory/kratos/issues/139)
- Deleting user does not delete sessions [\#69](https://github.com/ory/kratos/issues/69)
- Support RISC [\#10](https://github.com/ory/kratos/issues/10)
- pool: Comparing email addresses properly [\#3](https://github.com/ory/kratos/issues/3)
- Sign here to help! [\#2](https://github.com/ory/kratos/issues/2)
- Rough feature-ideas \(wip\) [\#1](https://github.com/ory/kratos/issues/1)

**Merged pull requests:**

- Remove redundant return statement [\#194](https://github.com/ory/kratos/pull/194) ([aeneasr](https://github.com/aeneasr))
- Improve Docker Compose Quickstart [\#187](https://github.com/ory/kratos/pull/187) ([aeneasr](https://github.com/aeneasr))
- Registration/Login HTML form: remove request field and ensure method is set [\#183](https://github.com/ory/kratos/pull/183) ([zepatrik](https://github.com/zepatrik))
- Replace number with integer in config JSON Schema [\#177](https://github.com/ory/kratos/pull/177) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#174](https://github.com/ory/kratos/pull/174) ([aeneasr](https://github.com/aeneasr))
- Schema testing [\#171](https://github.com/ory/kratos/pull/171) ([zepatrik](https://github.com/zepatrik))
- Add goreleaser orb task [\#170](https://github.com/ory/kratos/pull/170) ([aeneasr](https://github.com/aeneasr))
- Add changelog generation task [\#169](https://github.com/ory/kratos/pull/169) ([aeneasr](https://github.com/aeneasr))
- Adopt new SDK pipeline [\#168](https://github.com/ory/kratos/pull/168) ([aeneasr](https://github.com/aeneasr))
- Improve dev flag [\#167](https://github.com/ory/kratos/pull/167) ([zepatrik](https://github.com/zepatrik))
- Serve json schemas [\#164](https://github.com/ory/kratos/pull/164) ([zepatrik](https://github.com/zepatrik))
- update to readme.md [\#160](https://github.com/ory/kratos/pull/160) ([tacurran](https://github.com/tacurran))
- Bump go-acc and resolve test issues [\#154](https://github.com/ory/kratos/pull/154) ([aeneasr](https://github.com/aeneasr))
- Docker compose [\#153](https://github.com/ory/kratos/pull/153) ([aeneasr](https://github.com/aeneasr))
- Separate post register/login hooks [\#150](https://github.com/ory/kratos/pull/150) ([nmlc](https://github.com/nmlc))
- Optionally allow only one active session per identity  [\#148](https://github.com/ory/kratos/pull/148) ([evalsocket](https://github.com/evalsocket))
- Update documentation images [\#145](https://github.com/ory/kratos/pull/145) ([jfcurran](https://github.com/jfcurran))
- Refactor selfservice modules and add profile management [\#126](https://github.com/ory/kratos/pull/126) ([aeneasr](https://github.com/aeneasr))
- Rebrand ORY Hive to ORY Kratos [\#111](https://github.com/ory/kratos/pull/111) ([aeneasr](https://github.com/aeneasr))
- vendor: Update to ory/x 0.0.80 [\#110](https://github.com/ory/kratos/pull/110) ([aeneasr](https://github.com/aeneasr))
- Update README.md [\#107](https://github.com/ory/kratos/pull/107) ([aeneasr](https://github.com/aeneasr))
- Fix broken tests and linter issues  [\#104](https://github.com/ory/kratos/pull/104) ([aeneasr](https://github.com/aeneasr))
- errorx: Move package to selfservice [\#103](https://github.com/ory/kratos/pull/103) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#59](https://github.com/ory/kratos/pull/59) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#40](https://github.com/ory/kratos/pull/40) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#39](https://github.com/ory/kratos/pull/39) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#8](https://github.com/ory/kratos/pull/8) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#7](https://github.com/ory/kratos/pull/7) ([aeneasr](https://github.com/aeneasr))



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
