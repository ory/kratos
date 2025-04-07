# Changelog

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

**Table of Contents**

- [ (2025-04-07)](#2025-04-07)
  - [Breaking Changes](#breaking-changes)
  - [Related issue(s)](#related-issues)
  - [Related issue(s)](#related-issues-1)
  - [Related issue(s)](#related-issues-2)
  - [Related issue(s)](#related-issues-3)
  - [Related issue(s)](#related-issues-4)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# [](https://github.com/ory/kratos/compare/v1.3.0...v) (2025-04-07)

## Breaking Changes

This patch changes the behavior of configuration item `foo` to do bar. To keep
the existing behavior please do baz.

```
-->

## Related issue(s)

<!--
If this pull request

1. is a fix for a known bug, link the issue where the bug was reported

This patch changes the behavior of configuration item `foo` to do bar. To keep the existing
behavior please do baz.
```

-->

## Related issue(s)

<!--
If this pull request

1. is a fix for a known bug, link the issue where the bug was reported

This patch changes the behavior of configuration item `foo` to do bar. To keep the existing
behavior please do baz.
```
-->

## Related issue(s)

<!--
If this pull request

1. is a fix for a known bug, link the issue where the bug was reported

This patch changes the behavior of configuration item `foo` to do bar. To keep the existing
behavior please do baz.
```
-->

## Related issue(s)

<!--
If this pull request

1. is a fix for a known bug, link the issue where the bug was reported

This patch changes the behavior of configuration item `foo` to do bar. To keep the existing
behavior please do baz.
```
-->

## Related issue(s)

<!--
If this pull request

1. is a fix for a known bug, link the issue where the bug was reported

This patch changes the behavior of configuration item `foo` to do bar. To keep the existing
behavior please do baz.
```
-->

## Related issue(s)

<!--
If this pull request

1. is a fix for a known bug, link the issue where the bug was reported

The total count header `x-total-count` will no longer be sent in response to `GET /admin/sessions` requests.

Closes https://github.com/ory-corp/cloud/issues/7177
Closes https://github.com/ory-corp/cloud/issues/7175
Closes https://github.com/ory-corp/cloud/issues/7176



### Bug Fixes

* Accept login challenge in session_issuer on SPA flows ([#4288](https://github.com/ory/kratos/issues/4288)) ([e13687a](https://github.com/ory/kratos/commit/e13687ad51cdb889f0e680a005145a0134086fc7))
* Accept login_challenge in SPA verification flows ([#4284](https://github.com/ory/kratos/issues/4284)) ([7ca3b6b](https://github.com/ory/kratos/commit/7ca3b6be14c53e16c3a8f4e7eb83efe0b0e7c88e))
* Account linking should only happen after 2fa when required ([#4174](https://github.com/ory/kratos/issues/4174)) ([8e29b68](https://github.com/ory/kratos/commit/8e29b68a595d2ef18e48c2a01072335cefa36d86))
* Account linking with 2FA ([#4188](https://github.com/ory/kratos/issues/4188)) ([4a870a6](https://github.com/ory/kratos/commit/4a870a678dd3676abda7afc9803399dec4411b05)):

    This fixes some edge cases with OIDC account linking for accounts with 2FA enabled.

* Add exists clause ([#4191](https://github.com/ory/kratos/issues/4191)) ([a313dd6](https://github.com/ory/kratos/commit/a313dd6ba6d823deb40f14c738e3b609dbaad56c))
* Add missing autocomplete attributes to identifier_first strategy ([#4215](https://github.com/ory/kratos/issues/4215)) ([e1f29c2](https://github.com/ory/kratos/commit/e1f29c2d3524f9444ec067c52d2c9f1d44fa6539))
* Add missing csrf_token ([#4363](https://github.com/ory/kratos/issues/4363)) ([f441f41](https://github.com/ory/kratos/commit/f441f41312b81a570e99348f69b88008f4516660))
* Add missing discriminator ([#4365](https://github.com/ory/kratos/issues/4365)) ([c10bb06](https://github.com/ory/kratos/commit/c10bb06bb9125fbc71863c5aa82194da2f2e2888))
* Add missing saml group ([#4268](https://github.com/ory/kratos/issues/4268)) ([44eb305](https://github.com/ory/kratos/commit/44eb305cf91672798f7d57550a026c6b970f7566))
* Add missing submit group ([#4354](https://github.com/ory/kratos/issues/4354)) ([106163d](https://github.com/ory/kratos/commit/106163d15e2eb84c3403d0ce8f829a9d9b3ce94f))
* Add resend node to after registration verification flow ([#4260](https://github.com/ory/kratos/issues/4260)) ([9bc83a4](https://github.com/ory/kratos/commit/9bc83a410b8de9d649b6393f136889dd14098b0d))
* Add transient payload to fedcm ([#4369](https://github.com/ory/kratos/issues/4369)) ([245f5dc](https://github.com/ory/kratos/commit/245f5dc1c83d35d7e228a45ee267d6bcdb705e98)), closes [#1234](https://github.com/ory/kratos/issues/1234) [#1234](https://github.com/ory/kratos/issues/1234):

    <!--
    Describe the big picture of your changes here to communicate to the
    maintainers why we should accept this pull request.
    
    This text will be included in the changelog. If applicable, include
    links to documentation or pieces of code.
    If your change includes breaking changes please add a code block
    documenting the breaking change:
    
    ```

* Allow patching some /credentials sub-paths ([#4277](https://github.com/ory/kratos/issues/4277)) ([aefa806](https://github.com/ory/kratos/commit/aefa80623ee942254653b03ef4b273ae2779af0e)), closes [#1234](https://github.com/ory/kratos/issues/1234) [#1234](https://github.com/ory/kratos/issues/1234):

    <!--
    Describe the big picture of your changes here to communicate to the
    maintainers why we should accept this pull request.
    
    This text will be included in the changelog. If applicable, include
    links to documentation or pieces of code.
    If your change includes breaking changes please add a code block
    documenting the breaking change:
    
    ```

* Also update identifiers ([#4321](https://github.com/ory/kratos/issues/4321)) ([7c63727](https://github.com/ory/kratos/commit/7c6372794a94868555f647f6160be8205072c506)):

    This fixes a bug where when an identity is merged into another, the
    identifier of the original identity was not updated.

* Apply strategy filters in identifier first as well ([#4352](https://github.com/ory/kratos/issues/4352)) ([ec3ecc5](https://github.com/ory/kratos/commit/ec3ecc562a4d6ab511e53210d14c143903176b8c))
* Cancel conditional passkey before trying again ([#4247](https://github.com/ory/kratos/issues/4247)) ([d9f6f75](https://github.com/ory/kratos/commit/d9f6f75b6a43aad996f6390f73616a2cf596c6e4))
* Check aal on sessions list endpoint ([#4305](https://github.com/ory/kratos/issues/4305)) ([44f97b8](https://github.com/ory/kratos/commit/44f97b85e36160b8cce272fd61fbe3ac7d810fbf)), closes [#3671](https://github.com/ory/kratos/issues/3671):

    The session check to list a user's own sessions now requires the same AAL level as the whoami check.

* Count MFA addresses in CountActiveMultiFactorCredentials for code method ([9860c9a](https://github.com/ory/kratos/commit/9860c9a4faa5bd5d725c742c4d4ce9473baa0963)), closes [ory/network#409](https://github.com/ory/network/issues/409)
* Div decoding ([#4362](https://github.com/ory/kratos/issues/4362)) ([ef9ee23](https://github.com/ory/kratos/commit/ef9ee235866c6f3958574d2e98f09de3ca1e83af))
* Do not roll back transaction on partial identity insert error ([#4211](https://github.com/ory/kratos/issues/4211)) ([82660f0](https://github.com/ory/kratos/commit/82660f04e2f33d0aa86fccee42c90773a901d400))
* Don't show oidc subject in login hints ([#4264](https://github.com/ory/kratos/issues/4264)) ([b95fd3f](https://github.com/ory/kratos/commit/b95fd3fa723521807824cad84e4a9ce812172311))
* Duplicate autocomplete trigger ([6bbf915](https://github.com/ory/kratos/commit/6bbf91593a37e4973a86f610290ebab44df8dc81))
* Enable b2b_sso hook in more places ([#4168](https://github.com/ory/kratos/issues/4168)) ([0c48ad1](https://github.com/ory/kratos/commit/0c48ad12b978bf58b6bc68b0684a7879f93ebf06)):

    fix: allow b2b_sso hook in more places

* Ensure context is not canceled during password hashing ([#4364](https://github.com/ory/kratos/issues/4364)) ([e9c6a18](https://github.com/ory/kratos/commit/e9c6a1803daa622e559d0b8904cde4dc8834f1e2)):

    Especially during large imports of plaintext passwords there can be a
    lot of useless hashing, even after the request timed out or got
    canceled.

* Ensure that auto_link_credentials markers are being properly overwritten ([#4320](https://github.com/ory/kratos/issues/4320)) ([a4fd8ac](https://github.com/ory/kratos/commit/a4fd8acbbbd0cd0ff054e0f8737b076745aa71c8)), closes [#1234](https://github.com/ory/kratos/issues/1234) [#1234](https://github.com/ory/kratos/issues/1234):

    <!--
    Describe the big picture of your changes here to communicate to the
    maintainers why we should accept this pull request.
    
    This text will be included in the changelog. If applicable, include
    links to documentation or pieces of code.
    If your change includes breaking changes please add a code block
    documenting the breaking change:
    
    ```

* Exclude orgs ([#4351](https://github.com/ory/kratos/issues/4351)) ([68500d1](https://github.com/ory/kratos/commit/68500d14509a2697d2832eafafa5608fd8cfbf47))
* Explicity set updated_at field when updating identity ([#4131](https://github.com/ory/kratos/issues/4131)) ([66afac1](https://github.com/ory/kratos/commit/66afac173dc08b1d6666b107cf7050a2b0b27774))
* Gracefully handle unused index ([#4196](https://github.com/ory/kratos/issues/4196)) ([3dbeb64](https://github.com/ory/kratos/commit/3dbeb64b3f99a3aeba5f7126c301b72fda4c3e3c))
* IdentityCreated is over-reporting on error inserts ([#4323](https://github.com/ory/kratos/issues/4323)) ([c3f4ecf](https://github.com/ory/kratos/commit/c3f4ecf2562ffe400e500da97a93327b6115ddb6)):

    `defer` was the incorrect code path here, as we should only record
    identity created if the transaction did not error (aka was rolled back).

* Ignore CSRF on all apple provider callback URLs ([#4291](https://github.com/ory/kratos/issues/4291)) ([b60edba](https://github.com/ory/kratos/commit/b60edba1f4642f07b411271b6c7a442665dc2a74))
* Improve linking on OIDC signup ([#4314](https://github.com/ory/kratos/issues/4314)) ([687d578](https://github.com/ory/kratos/commit/687d5787b12450895ba613ceee47da408917a0a7)), closes [#1234](https://github.com/ory/kratos/issues/1234) [#1234](https://github.com/ory/kratos/issues/1234):

    <!--
    Describe the big picture of your changes here to communicate to the
    maintainers why we should accept this pull request.
    
    This text will be included in the changelog. If applicable, include
    links to documentation or pieces of code.
    If your change includes breaking changes please add a code block
    documenting the breaking change:
    
    ```

* Incorrect if switch in previous sceen case in two step registration ([f8ee403](https://github.com/ory/kratos/commit/f8ee40396a36a2e7a348c9cf983dec7db13814c5)), closes [#374](https://github.com/ory/kratos/issues/374)
* Incorrect query plan ([#4218](https://github.com/ory/kratos/issues/4218)) ([7d0e78a](https://github.com/ory/kratos/commit/7d0e78a4f6631b0662beee3b8e9dd0d774b875ea))
* Order-by clause and span names ([#4200](https://github.com/ory/kratos/issues/4200)) ([b6278af](https://github.com/ory/kratos/commit/b6278af5c7ed7fb845a71ad0e64f8b87402a8f4b))
* Pass on correct context during verification ([#4151](https://github.com/ory/kratos/issues/4151)) ([7e0b500](https://github.com/ory/kratos/commit/7e0b500aada9c1931c759a43db7360e85afb57e3))
* Preview_credentials_identifier_similar ([#4246](https://github.com/ory/kratos/issues/4246)) ([5ee54ed](https://github.com/ory/kratos/commit/5ee54eda909638fa10c543f156042a217b34cba6))
* Registration post persist hooks should not be cancelable ([#4148](https://github.com/ory/kratos/issues/4148)) ([18056a0](https://github.com/ory/kratos/commit/18056a0f1cfdf42769e5a974b2526ccf5c608cc2))
* Rename b2b_sso hook ([#4349](https://github.com/ory/kratos/issues/4349)) ([d9e3295](https://github.com/ory/kratos/commit/d9e3295d98b0446a90a960d0f0e957e7a6513dfc))
* Return `return_to` code if already authenticated ([#4286](https://github.com/ory/kratos/issues/4286)) ([119841a](https://github.com/ory/kratos/commit/119841a304917e222d8c0fd4606419a520f481c1)):

    This fixes a bug in native OIDC login and registration flows, where the
    user already has a session in the browser the flow is continued with
    (usually a web view, but depending on the platform it already has a
    session cookie set). In the callback, we now correctly handle the case
    in `alreadyAuthenticated` to return the session token exchange code.

* Schema key ([#4332](https://github.com/ory/kratos/issues/4332)) ([306316f](https://github.com/ory/kratos/commit/306316fedf20467059776c003f8285880d272c95))
* **sdk:** Add missing captcha group ([#4254](https://github.com/ory/kratos/issues/4254)) ([241111b](https://github.com/ory/kratos/commit/241111b21f5d96b26ff8bc8106dc8a527c68063b))
* **sdk:** Remove incorrect attributes ([#4163](https://github.com/ory/kratos/issues/4163)) ([88c68aa](https://github.com/ory/kratos/commit/88c68aa07281a638c9897e76d300d1095b17601d))
* Send correct verification status in post-recovery hook ([#4224](https://github.com/ory/kratos/issues/4224)) ([7f50400](https://github.com/ory/kratos/commit/7f5040080578e194dde3605dbb1a344fe9ff27ae)):

    The verification status is now correctly being transported when executing a recovery hook.

* Set correct request url in acc linking and oidc flows ([#4282](https://github.com/ory/kratos/issues/4282)) ([07cb83c](https://github.com/ory/kratos/commit/07cb83c672326848162998a9cfbc8ca34af42bf0))
* Settings linking error override ([#4368](https://github.com/ory/kratos/issues/4368)) ([6e30865](https://github.com/ory/kratos/commit/6e30865e1314bc4c4fdc3b472b34c92019eadfa4))
* Show code email in most error states ([#4338](https://github.com/ory/kratos/issues/4338)) ([905d1e5](https://github.com/ory/kratos/commit/905d1e5dc8fcdc7f96afa14a5ee036060ea43056))
* Span names ([#4232](https://github.com/ory/kratos/issues/4232)) ([dbae98a](https://github.com/ory/kratos/commit/dbae98a26b8e2a3328d8510745ddb58c18b7ad3d))
* Stricter JSON patch checking for PATCH identities ([#4263](https://github.com/ory/kratos/issues/4263)) ([906f6c8](https://github.com/ory/kratos/commit/906f6c8fdf9ec0834993a44f8a19697b38dd63d2))
* Truncate updated at ([#4149](https://github.com/ory/kratos/issues/4149)) ([2f8aaee](https://github.com/ory/kratos/commit/2f8aaee0716835caaba0dff9b6cc457c2cdff5d4))
* Use context for readiness probes ([#4219](https://github.com/ory/kratos/issues/4219)) ([e6d2d4d](https://github.com/ory/kratos/commit/e6d2d4d0c04e60ab5b0658b9e5c4c52104446368))

### Chores

* Document test migration ([#4265](https://github.com/ory/kratos/issues/4265)) ([9959545](https://github.com/ory/kratos/commit/9959545cb9d90364e1928fcc4f01b3171e052360)), closes [#1234](https://github.com/ory/kratos/issues/1234) [#1234](https://github.com/ory/kratos/issues/1234):

    <!--
    Describe the big picture of your changes here to communicate to the
    maintainers why we should accept this pull request.
    
    This text will be included in the changelog. If applicable, include
    links to documentation or pieces of code.
    If your change includes breaking changes please add a code block
    documenting the breaking change:
    
    ```

* Upgrade to go 1.24 ([#4313](https://github.com/ory/kratos/issues/4313)) ([bdb046d](https://github.com/ory/kratos/commit/bdb046da36c290f775ad4cabdbe6191295252cc7)), closes [#1234](https://github.com/ory/kratos/issues/1234) [#1234](https://github.com/ory/kratos/issues/1234):

    <!--
    Describe the big picture of your changes here to communicate to the
    maintainers why we should accept this pull request.
    
    This text will be included in the changelog. If applicable, include
    links to documentation or pieces of code.
    If your change includes breaking changes please add a code block
    documenting the breaking change:
    
    ```


### Code Refactoring

* Hash comparator instantiation ([#4195](https://github.com/ory/kratos/issues/4195)) ([53a5a8b](https://github.com/ory/kratos/commit/53a5a8b93cec274456df3d988eb3bd12bc11fa87))
* Remove total count from listSessions and improve secondary indices ([#4173](https://github.com/ory/kratos/issues/4173)) ([e24f993](https://github.com/ory/kratos/commit/e24f993ea4236bac4e23bd4250c11b5932040fd9)):

    This patch changes sorting to improve performance on list session endpoints. It also removes the `x-total-count` header from list responses.

* Two-step registration ([#4348](https://github.com/ory/kratos/issues/4348)) ([f46aed1](https://github.com/ory/kratos/commit/f46aed12a244094e9e3e4014792543d6fb1a2a4b)):

    Refactors internals of the two-step registration to better fit into the architecture.


### Documentation

* Add return_to query parameter to OAS Verification Flow for Native Apps ([#4086](https://github.com/ory/kratos/issues/4086)) ([b22135f](https://github.com/ory/kratos/commit/b22135fa05d7fb47dfeaccd7cdc183d16921a7ac))
* Clarify facebook graph API versioning ([#4208](https://github.com/ory/kratos/issues/4208)) ([a90df58](https://github.com/ory/kratos/commit/a90df5852ba96704863cc576edcb8286eaa9b3f9))
* Defining oid as oidc subject_source ([#4270](https://github.com/ory/kratos/issues/4270)) ([b388a4a](https://github.com/ory/kratos/commit/b388a4ab9fe101def70ca604fbfbef2bcccd01a9))
* Improve SecurityError error message for ory elements local ([#4205](https://github.com/ory/kratos/issues/4205)) ([0062d45](https://github.com/ory/kratos/commit/0062d45b6c9a6323f9dccb10f63dce752836c29e))
* Remove unused SMS config from schema ([#4212](https://github.com/ory/kratos/issues/4212)) ([f076fe4](https://github.com/ory/kratos/commit/f076fe4e1487f67f355eaa7f238090abf3796578))
* Usage of `organization` parameter in native self-service flows ([#4176](https://github.com/ory/kratos/issues/4176)) ([cb71e38](https://github.com/ory/kratos/commit/cb71e38147d21f73e9bd1e081dc3443abb63353e))

### Features

* Add a policy callback to customize OIDC credential linking ([#4302](https://github.com/ory/kratos/issues/4302)) ([e2f878a](https://github.com/ory/kratos/commit/e2f878a3ed25b4617bf6bd43b6e147d87d3b8ca2))
* Add attributes to webhook events for better debugging ([#4206](https://github.com/ory/kratos/issues/4206)) ([00da05d](https://github.com/ory/kratos/commit/00da05da9f77bbfb68b364b3ba2a5d0a2d9e4f15))
* Add captcha group to first-step registration ([eca4ae9](https://github.com/ory/kratos/commit/eca4ae9dcce37d03bbd1bf5f0cd492466c02acde))
* Add context param to policy ([#4315](https://github.com/ory/kratos/issues/4315)) ([261596b](https://github.com/ory/kratos/commit/261596b7261c315b7d8291e886023c34fc9135c5))
* Add explicit config flag for secure cookies ([#4180](https://github.com/ory/kratos/issues/4180)) ([2aabe12](https://github.com/ory/kratos/commit/2aabe12e5329acc807c495445999e5591bdf982b)):

    Adds a new config flag  for session and all other cookies. Falls back to the previous behavior of using the dev mode to decide if the cookie should be secure or not.

* Add failure reason to events ([#4203](https://github.com/ory/kratos/issues/4203)) ([afa7618](https://github.com/ory/kratos/commit/afa76180e77df0ee0f96eef3b3f2b2d3fe08a33d))
* Add migrate sql up|down|status ([#4228](https://github.com/ory/kratos/issues/4228)) ([e6fa520](https://github.com/ory/kratos/commit/e6fa520058ca778e01d4e93a8ab4b31a74dd2e11)):

    This patch adds the ability to execute down migrations using:
    
    ```
    kratos migrate sql down -e --steps {num_of_steps}
    ```
    
    Please read `kratos migrate sql down --help` carefully.
    
    Going forward, please use the following commands
    
    ```
    kratos migrate sql up ...
    kratos migrate sql status ...
    ```
    
    instead of the previous, now deprecated
    
    ```
    kratos migrate sql ...
    kratos migrate status ...
    ```
    
    commands.
    
    See https://github.com/ory-corp/cloud/issues/7350

* Add new Division ui node attributes ([235af52](https://github.com/ory/kratos/commit/235af527dea47b87ad0f18ff04f9b807e4639ae3)):

    Division nodes may be used to hook dynamic scripts and are not actively used in the Ory Kratos open source.

* Add oid as subject source for microsoft ([#4171](https://github.com/ory/kratos/issues/4171)) ([77beb4d](https://github.com/ory/kratos/commit/77beb4de5209cee0bea4b63dfec21d656cf64473)), closes [#4170](https://github.com/ory/kratos/issues/4170):

    In the case of Microsoft, using `sub` as an identifier can lead to problems. Because the use of OIDC at Microsoft is based on an app registration, the content of `sub` changes with every new app registration. `Sub` is therefore not uniquely related to the user. It is therefore not possible to transfer users from one app registration to another without further problems.
    https://learn.microsoft.com/en-us/entra/identity-platform/id-token-claims-reference#payload-claims
    
    With the use of `oid` it is possible to identify a user by a unique id.

* Allow deleting password credentials ([#4304](https://github.com/ory/kratos/issues/4304)) ([f2212d4](https://github.com/ory/kratos/commit/f2212d48af47f24ca6e504ca98bc31afe6774241)):

    The admin API did not allow to delete passwords at all. The restriction is now lifted to only block deletion of the first-factor credential if it is the last one.

* Allow extra go migrations in persister ([#4183](https://github.com/ory/kratos/issues/4183)) ([7bec935](https://github.com/ory/kratos/commit/7bec935c33b9adb6033aaecfa9a6dbe6c9c3daa1))
* Allow listing identities by organization ID ([#4115](https://github.com/ory/kratos/issues/4115)) ([b4c453b](https://github.com/ory/kratos/commit/b4c453b0472f67d0a52b345691f66aa48777a897))
* Allow setting the org ID on creation ([#4306](https://github.com/ory/kratos/issues/4306)) ([bccd2fb](https://github.com/ory/kratos/commit/bccd2fb8c8efac96938e564f1f34cd711b41d0a1))
* Cache OIDC providers ([#4222](https://github.com/ory/kratos/issues/4222)) ([30485c4](https://github.com/ory/kratos/commit/30485c44e61c17231e0c46b321be842b19ea5a5f)):

    This change significantly reduces the number of requests to `/.well-known/openid-configuration` endpoints.

* Drop unused indices post index migration ([#4201](https://github.com/ory/kratos/issues/4201)) ([1008639](https://github.com/ory/kratos/commit/1008639428a6b72e0aa47bd13fe9c1d120aafb6e))
* Emit admin recovery code event ([#4230](https://github.com/ory/kratos/issues/4230)) ([a7cdc3a](https://github.com/ory/kratos/commit/a7cdc3a6911e265f4e78c780d8e4b8922066875c))
* Fast add credential type lookups ([#4177](https://github.com/ory/kratos/issues/4177)) ([eeb1355](https://github.com/ory/kratos/commit/eeb13552118504f17b48f2c7e002e777f5ee73f4))
* Fewer DB loads when linking credentials, add tracing ([2c5bb21](https://github.com/ory/kratos/commit/2c5bb21224e28d5218354349f77514f4fbe71762))
* Gracefully handle failing password rehashing during login ([#4235](https://github.com/ory/kratos/issues/4235)) ([3905787](https://github.com/ory/kratos/commit/39057879821b387b49f5d4f7cb19b9e02ec924a7)):

    This fixes an issue where we would successfully import long passwords (>72 chars), but fail when the user attempts to login with the correct password because we can't rehash it. In this case, we simply issue a warning to the logs, keep the old hash intact, and continue logging in the user.

* Improve QueryForCredentials ([#4181](https://github.com/ory/kratos/issues/4181)) ([ca0d6a7](https://github.com/ory/kratos/commit/ca0d6a7ea717495429b8bac7fd843ac69c1ebf16))
* Improve secondary indices for self service tables ([#4179](https://github.com/ory/kratos/issues/4179)) ([825aec2](https://github.com/ory/kratos/commit/825aec208d966b54df9eeac6643e6d8129cf2253))
* Improved tracing for courier ([85a7071](https://github.com/ory/kratos/commit/85a7071d20d0f072316c74bee82c76ee690276f8))
* Index hint for CRDB when deleting identity credentials ([#4276](https://github.com/ory/kratos/issues/4276)) ([c703a33](https://github.com/ory/kratos/commit/c703a338894f865c7dc1dcebc6e6980ad98eaa1d)):

    Ref https://support.cockroachlabs.com/hc/en-us/requests/25430

* Jackson provider ([#4242](https://github.com/ory/kratos/issues/4242)) ([f18d1b2](https://github.com/ory/kratos/commit/f18d1b24539f7d8dcf9c27986af861d0f8cb9683)):

    This adds a jackson provider to Kratos.

* Load session only once when middleware is used ([#4187](https://github.com/ory/kratos/issues/4187)) ([234b6f2](https://github.com/ory/kratos/commit/234b6f2f6435c62b7e161c032b888c4e2b3328d4))
* More extension points ([#4272](https://github.com/ory/kratos/issues/4272)) ([373a2e6](https://github.com/ory/kratos/commit/373a2e6552f0da0488638306a58d8bd63a6ca10a)):

    This adds more extension points to the Kratos registry.

* Optimize identity-related secondary indices ([#4182](https://github.com/ory/kratos/issues/4182)) ([53874c1](https://github.com/ory/kratos/commit/53874c1753940e08e0bf50753a1d3126add77af1))
* Passwordless SMS and expiry notice in code / link templates ([#4104](https://github.com/ory/kratos/issues/4104)) ([462cea9](https://github.com/ory/kratos/commit/462cea91448a00a0db21e20c2c347bf74957dc8f)):

    This feature allows Ory Kratos to use the SMS gateway for login and registration with code via SMS.
    
    Additionally, the default email and sms templates have been updated. We now also expose `ExpiresInMinutes` / `expires_in_minutes` in the templates, making it easier to remind the user how long the code or link is valid for.
    
    Closes https://github.com/ory/kratos/issues/1570
    Closes https://github.com/ory/kratos/issues/3779

* Refactor cmd/daemon ([#4371](https://github.com/ory/kratos/issues/4371)) ([7fe55d9](https://github.com/ory/kratos/commit/7fe55d9fec5e5f4048b211eaa56ac61e29635157))
* Remove duplicate queries during settings flow and use better index hint for credentials lookup ([#4193](https://github.com/ory/kratos/issues/4193)) ([c33965e](https://github.com/ory/kratos/commit/c33965e5735ead3acddac87ef84c3a730874f9ab)):

    This patch reduces duplicate GetIdentity queries as part of submitting the settings flow, and improves an index to significantly reduce credential lookup.
    
    For better debugging, more tracing ha been added to the settings module.

* Remove more unused indices ([#4186](https://github.com/ory/kratos/issues/4186)) ([b294804](https://github.com/ory/kratos/commit/b2948044de4eee1841110162fe874055182bd2d2))
* Rework the OTP code submit count mechanism ([#4251](https://github.com/ory/kratos/issues/4251)) ([4ca4d79](https://github.com/ory/kratos/commit/4ca4d79cff5185caad27eddee7e6f8d0e58463ba)):

    * feat: rework the OTP code submit count mechanism
    
    Unlike what the previous comment suggested, incrementing and checking the submit count inside the
    database transaction is not actually optimal peformance- or security-wise.
    
    We now check atomically increment and check the submit count as the first part of the operation,
    and abort as early as possible if we detect brute-forcing. This prevents a situation where the
    check works only on certain transaction isolation levels.
    
    * chore: bump dependencies

* Support android webauthn origins ([#4155](https://github.com/ory/kratos/issues/4155)) ([a82d288](https://github.com/ory/kratos/commit/a82d288014411ae4eb82c718bfe825ca55b4fab0)):

    This patch adds the ability to verify Android APK origins used during WebAuthn/Passkey exchange.
    
    Upgrades go-webauthn and includes fixes for Go 1.23 and workarounds for Swagger.

* Support importing more credentials ([#4361](https://github.com/ory/kratos/issues/4361)) ([9a6dadf](https://github.com/ory/kratos/commit/9a6dadfefaf0d54c227cdbab5a2cbe7da14faa96)):

    Adds support to import SAML credentials. SAML connections are only
    available in Ory Enterprise License / Ory Network.

* Update only necessary database columns in UpdateVerifiableAddress ([#4292](https://github.com/ory/kratos/issues/4292)) ([168a3f6](https://github.com/ory/kratos/commit/168a3f6c68b1fbc0ddcd455f8762f6de19879442)):

    This is an optimization to reduce database load.
    
    When we specify exactly which columns changed, we should be able to
    elide updates to the `identity_verifiable_addresses_status_via_uq_idx
    (nid,via,value)` index. Updating that index requires contacting remote
    regions.
    
    Also fixed a bug where we did not set the `verified_at` timestamp
    correctly sometimes.

* Use one transaction for `/admin/recovery/code` ([#4225](https://github.com/ory/kratos/issues/4225)) ([3e87e0c](https://github.com/ory/kratos/commit/3e87e0c4559736f9476eba943bac8d67cde91aad))
* Webhook header allowlist configuration option ([#4309](https://github.com/ory/kratos/issues/4309)) ([871f5aa](https://github.com/ory/kratos/commit/871f5aab6d7b2a655ebcd6f0f90e79635ffc85f6)), closes [#4290](https://github.com/ory/kratos/issues/4290):

    Adds a `clients.web_hook.header_allowlist` configuration option for
    configuring the webhook header allowlist.


### Tests

* Update snapshots ([#4167](https://github.com/ory/kratos/issues/4167)) ([b51f780](https://github.com/ory/kratos/commit/b51f780b7e4abc79a757ac1efe1cb65b3d35c8a4))


# [1.3.0](https://github.com/ory/kratos/compare/v1.2.0...v1.3.0) (2024-09-26)

We are thrilled to announce the release of [Ory Kratos v1.3.0](https://www.ory.sh/kratos)! This release includes significant updates, enhancements, and fixes to improve your experience with Ory Kratos.

![Ory Kratos 1.3.0 Release](https://www.ory.sh/images/newsletter/kratos-1.3.0/kratos-1.3-release.png)

Enhance your sign-in experience with Identifier First Authentication. This feature allows users to first identify themselves (e.g., by providing their email or username) and then proceed with the chosen authentication method, whether it be OTP code, passkeys, passwords, or social login. By streamlining the sign-in process, users can select the authentication method that best suits their needs, reducing friction and enhancing security. Identifier First Authentication improves user flow and reduces the likelihood of errors, resulting in a more user-friendly and efficient login experience.

![Identifier First Authentication](https://www.ory.sh/images/newsletter/kratos-1.3.0/identifier-first-demo.png)

The UI for OpenID Connect (OIDC) account linking has been improved to provide better user guidance and error messages during the linking process. As a result, account linking error rates have dropped significantly, making it easier for users to link multiple identities (e.g., social login and email-based accounts) to the same profile. This improvement enhances user convenience, reduces support inquiries, and offers a seamless multi-account experience.

You can now use Salesforce as an identity provider, expanding the range of supported identity providers. This integration allows organizations already using Salesforce for identity management to leverage their existing infrastructure, simplifying user management and enhancing the authentication experience.

Social sign-in has been enhanced with better detection and handling of double-submit issues, especially for platforms like Facebook and Apple mobile login. These changes make the social login process more reliable, reducing errors and improving the user experience. Additionally, Ory Kratos now supports social providers in credential discovery, offering more flexibility during sign-up and sign-in flows.

One-Time Password (OTP) MFA has been improved with more robust handling of code-based authentication. The enhancements ensure a smoother flow when using OTP for multi-factor authentication (MFA), providing clearer guidance to users and improving fallback mechanisms. These updates help to prevent users from being locked out due to misconfigurations or errors during the MFA process, increasing security without compromising user convenience.

- **Deprecated `via` Parameter for SMS 2FA**: The `via` parameter is now deprecated when performing SMS 2FA. If not included, users will see all their phone/email addresses to perform the flow. This parameter will be removed in a future version. Ensure your identity schema has the appropriate code configuration for passwordless or 2FA login.
- **Endpoint Change**: The `/admin/session/.../extend` endpoint will now return 204 No Content for new Ory Network projects. Returning 200 with the session body will be deprecated in future versions.

- **SDK Enhancements**: Added new methods and support for additional actions in the SDK, improving integration capabilities.
- **Password Migration Hook**: Added a password migration hook to facilitate migrating passwords where the hash is unavailable, easing the transition to Ory Kratos.
- **Partially Failing Batch Inserts:** When batch-inserting multiple identities, conflicts or validation errors of a subset of identities in the batch still allow the rest of the identities to be inserted. The returned JSON contains the error details that led to the failure.

- **Security Fixes**: Fixed a security vulnerability where the `code` method did not respect the `highest_available`setting. Refer to the [security advisory](https://github.com/ory/kratos/security/advisories/GHSA-wc43-73w7-x2f5) for more details.
- **Session Extension Issues**: Fixed issues related to session extension to prevent long response times on `/session/whoami` when extending sessions simultaneously.
- **OIDC and Social Sign-In**: Fixed UI and error handling for OpenID Connect and social sign-in flows, improving the overall experience.
- **Credential Identifier Handling**: Corrected handling of code credential identifiers, ensuring proper detection of phone numbers and correct functioning of SMS/email MFA.
- **Concurrent Updates for Webhooks**: Fixed concurrent map update issues for webhook headers, improving webhook reliability.

- **Passwordless & 2FA Login**: Before upgrading, ensure your identity schema has the appropriate code configuration when using the code method for passwordless or 2FA login.
- **Code Method for 2FA**: If you use the code method for 2FA or 1FA login but haven't configured the code identifier, set `selfservice.methods.code.config.missing_credential_fallback_enabled` to `true` to avoid user lockouts.

We hope you enjoy the new features and improvements in Ory Kratos v1.3.0. Please remember to leave a [GitHub star](https://github.com/ory/kratos) and check out our other [open-source projects](https://github.com/ory). Your feedback is valuable to us, so join the [Ory community](https://slack.ory.sh/) and help us shape the future of identity management.



## Breaking Changes

When using two-step registration, it was previously possible to send `method=profile:back` to get to the previous screen. This feature was not documented in the SDK API yet. Going forward, please instead use `screen=previous`.

Please note that the `via` parameter is deprecated when performing SMS 2FA. It will be removed in a future version. If the parameter is not included in the request, the user will see all their phone/email addresses from which to perform the flow.

Before upgrading, ensure that your identity schema has the appropriate code configuration when using the code method for passwordless or 2fa login.

If you are using the code method for 2FA login already, or you are using it for 1FA login but have not yet configured the code identifier, set `selfservice.methods.code.config.missing_credential_fallback_enabled` to `true` to prevent users from being locked out.

Please note that the `via` parameter is deprecated when performing SMS 2FA. It will be removed in a future version. If the parameter is not included in the request, the user will see all their phone/email addresses from which to perform the flow.

Before upgrading, ensure that your identity schema has the appropriate code configuration when using the code method for passwordless or 2fa login.

If you are using the code method for 2FA login already, or you are using it for 1FA login but have not yet configured the code identifier, set `selfservice.methods.code.config.missing_credential_fallback_enabled` to `true` to prevent users from being locked out.

Going forward, the `/admin/session/.../extend` endpoint will return 204 no content for new Ory Network projects. We will deprecate returning 200 + session body in the future.



### Bug Fixes

* Add continue with only for json browser requests ([#4002](https://github.com/ory/kratos/issues/4002)) ([e0a4010](https://github.com/ory/kratos/commit/e0a4010b84b43f364be14414a380c872b166274d))
* Add fallback to providerLabel ([#3999](https://github.com/ory/kratos/issues/3999)) ([d26f204](https://github.com/ory/kratos/commit/d26f2042eb5325a8d639c08d95a005724e61cb8e)):

    This adds a fallback to the provider label when trying to register a duplicate identifier with an oidc.
    
    Current error message:
    
    `Signing in will link your account to "test@test.com" at provider "". If you do not wish to link that account, please start a new login flow.`
    
    The label represents an optional label for the UI, but in my case it's always empty. I suggest we fallback to the provider when the label is not present. In case the label is present, the behaviour won't change.
    
    Fallback to provider:
    
    `Signing in will link your account to "test@test.com" at provider "google". If you do not wish to link that account, please start a new login flow.`

* Add missing JS triggers ([7597bc6](https://github.com/ory/kratos/commit/7597bc6345848b66161d5a9b7a42307bbc85c978))
* Add PKCE config key to config schema ([#4098](https://github.com/ory/kratos/issues/4098)) ([2c7ff3c](https://github.com/ory/kratos/commit/2c7ff3c8baab6aaa105e2d733a483fc07537470f))
* Batch identity created event ([#4111](https://github.com/ory/kratos/issues/4111)) ([340f698](https://github.com/ory/kratos/commit/340f698243bd908e217394710b475a7f686a8cf9))
* Concurrent map update for webhook header ([#4055](https://github.com/ory/kratos/issues/4055)) ([6ceb2f1](https://github.com/ory/kratos/commit/6ceb2f1213e1b28d3aa72380661e4aa985bfa437))
* Do not populate `id_first` first step for account linking flows ([#4074](https://github.com/ory/kratos/issues/4074)) ([6ab2637](https://github.com/ory/kratos/commit/6ab2637652013e0ff377f52355e2025d68c7b3d3))
* Downgrade go-webauthn ([#4035](https://github.com/ory/kratos/issues/4035)) ([4d1954a](https://github.com/ory/kratos/commit/4d1954ac74dee358f9a08e619848dfe94e4934ce))
* Emit SelfServiceMethodUsed in SettingsSucceeded event ([#4056](https://github.com/ory/kratos/issues/4056)) ([76af303](https://github.com/ory/kratos/commit/76af303b20ae5dffb932169a73667a55be3f3f80))
* Filter web hook headers ([#4048](https://github.com/ory/kratos/issues/4048)) ([ddb838e](https://github.com/ory/kratos/commit/ddb838e0e8f7d752cd1708c505e80b6c0ccc0b8a))
* Improve OIDC account linking UI ([#4036](https://github.com/ory/kratos/issues/4036)) ([2b4a618](https://github.com/ory/kratos/commit/2b4a618485c9d79762243f59b35f142083f5492c))
* Include duplicate credentials in account linking message ([#4079](https://github.com/ory/kratos/issues/4079)) ([122b63d](https://github.com/ory/kratos/commit/122b63d68a3ff2ad78107300869c5a6d2aa43354))
* Incorrect append of code credential identifier ([#4102](https://github.com/ory/kratos/issues/4102)) ([3215792](https://github.com/ory/kratos/commit/3215792df4cab494c05ef09e969b2fa0ed95a98b)), closes [#4076](https://github.com/ory/kratos/issues/4076)
* Jsonnet timeouts ([#3979](https://github.com/ory/kratos/issues/3979)) ([7c5299f](https://github.com/ory/kratos/commit/7c5299f1f832ebbe0622d0920b7a91253d26b06c))
* Move password migration hook config ([#3986](https://github.com/ory/kratos/issues/3986)) ([b5a66e0](https://github.com/ory/kratos/commit/b5a66e0dde3a8fa6fdeb727482481b6302589631)):

    This moves the password migration hook to
    
    ```yaml
    selfservice:
      methods:
        password:
          config:
            migrate_hook:
              ...
    ```

* Normalize code credentials and deprecate via parameter ([c417b4a](https://github.com/ory/kratos/commit/c417b4aa76a76d3aebb4474999d7bb072615bd9f)):

    Before this, code credentials for passwordless and mfa login were incorrectly stored and normalized. This could cause issues where the system would not detect the user's phone number, and where SMS/email MFA would not properly work with the `highest_available` setting.

* Passthrough correct organization ID to CompletedLoginForWithProvider ([#4124](https://github.com/ory/kratos/issues/4124)) ([ad1acd5](https://github.com/ory/kratos/commit/ad1acd51d8dd7582b05a3078b92f73970e1e2715))
* Password migration hook config ([#4001](https://github.com/ory/kratos/issues/4001)) ([50deedf](https://github.com/ory/kratos/commit/50deedfeecf7adbc948521371b181306a0c26cf1)):

    This fixes the config loading for the password migration hook.

* Pw migration param ([#3998](https://github.com/ory/kratos/issues/3998)) ([6016cc8](https://github.com/ory/kratos/commit/6016cc88a076eeea71a85d75cfb5191808b69844))
* Refactor internal API to prevent panics ([#4028](https://github.com/ory/kratos/issues/4028)) ([81bc152](https://github.com/ory/kratos/commit/81bc1525f09504729c666192d458cf2eaafab99f))
* Remove flows from log messages ([#3913](https://github.com/ory/kratos/issues/3913)) ([310a405](https://github.com/ory/kratos/commit/310a405202c6b44633b15ad30e1fdb8ebd153e4b))
* Replace submit with continue button for recovery and verification and add maxlength ([04850f4](https://github.com/ory/kratos/commit/04850f45cfbdc89223366ffa3b540d579a3b44be))
* Return credentials in FindByCredentialsIdentifier ([#4068](https://github.com/ory/kratos/issues/4068)) ([f949173](https://github.com/ory/kratos/commit/f949173b3ed3d45167bb4af8b95440d5e4a39636)):

    Instead of re-fetching the credentials later (expensive), we load them only once.

* Return error if invalid UUID is supplied to ids filter ([#4116](https://github.com/ory/kratos/issues/4116)) ([98140f2](https://github.com/ory/kratos/commit/98140f2fd43ccd889e2635e4f3e7582b92fe96ab))
* **security:** Code credential does not respect `highest_available` setting ([b0111d4](https://github.com/ory/kratos/commit/b0111d4bd561d0f0e2f5883f30fac36fcf7135d5)):

    This patch fixes a security vulnerability which prevents the `code` method to properly report it's credentials count to the `highest_available` mechanism.
    
    For more details on this issue please refer to the [security advisory](https://github.com/ory/kratos/security/advisories/GHSA-wc43-73w7-x2f5).

* Timestamp precision on mysql ([9a1f171](https://github.com/ory/kratos/commit/9a1f171c1a4a8d20dc2103073bdc11ee3fdc70af))
* Transient_payload is lost when verification flow started as part of registration ([#3983](https://github.com/ory/kratos/issues/3983)) ([192f10f](https://github.com/ory/kratos/commit/192f10f4ad9eb44a612baaccfc71765d52c7e1ed))
* Trigger oidc web hook on sign in after registration ([#4027](https://github.com/ory/kratos/issues/4027)) ([ad5fb09](https://github.com/ory/kratos/commit/ad5fb09687f863e7c5d45868d0b8f5ec2d965372))
* Typo in login link CLI error messages ([#3995](https://github.com/ory/kratos/issues/3995)) ([8350625](https://github.com/ory/kratos/commit/835062542077b9dd8d6a30836d0455adb015265d))
* Validate page tokens for better error codes ([#4021](https://github.com/ory/kratos/issues/4021)) ([32737dc](https://github.com/ory/kratos/commit/32737dc708c1ecf0ec0ceaa4bbc0ac09286186fd))
* Whoami latency ([#4070](https://github.com/ory/kratos/issues/4070)) ([ff6ed5b](https://github.com/ory/kratos/commit/ff6ed5b70b7f715fc38a41cedd17b5323aebd79e))

### Code Generation

* Pin v1.3.0 release commit ([0a49fd0](https://github.com/ory/kratos/commit/0a49fd05245f179501b117163cd574786f287fe8))

### Documentation

* Add google to supported providers in ID Token doc strings ([#4026](https://github.com/ory/kratos/issues/4026)) ([955bd8f](https://github.com/ory/kratos/commit/955bd8fbc1353d7a9f84d8f591c3af31781cf7b7))
* Typo in changelog ([c508980](https://github.com/ory/kratos/commit/c5089801af2a656e9c1fc371a11aeb23918ba359))

### Features

* Add additional messages ([735fc5b](https://github.com/ory/kratos/commit/735fc5b2c5a99746d3012cc38ee2e1b7cc3a67f2))
* Add browser return_to continue_with action ([7b636d8](https://github.com/ory/kratos/commit/7b636d860c6917cb1133d6d1d7401808adb890c7))
* Add if method to sdk ([612e3bf](https://github.com/ory/kratos/commit/612e3bf09dbffd3feba08d5100bffbc39cbd240a))
* Add redirect to continue_with for SPA flows ([99c945c](https://github.com/ory/kratos/commit/99c945c92d0c2745dc8df4402d755afd53e1b9aa)):

    This patch adds the new `continue_with` action `redirect_browser_to`, which contains the redirect URL the app should redirect to. It is only supported for SPA (not server-side browser apps, not native apps) flows at this point in time.

* Add social providers to credential discovery as well ([5f4a2bf](https://github.com/ory/kratos/commit/5f4a2bf619d540d45e96586129c8ee1e7850e745))
* Add support for Salesforce as identity provider ([#4003](https://github.com/ory/kratos/issues/4003)) ([3bf1ca9](https://github.com/ory/kratos/commit/3bf1ca9030555df90ef9903c34313ae4bd1fecae))
* Add tests for two step login ([#3959](https://github.com/ory/kratos/issues/3959)) ([8225e40](https://github.com/ory/kratos/commit/8225e40e3d767e945006b33eebdfc47fd242ff06))
* Allow deletion of an individual OIDC credential ([#3968](https://github.com/ory/kratos/issues/3968)) ([a43cef2](https://github.com/ory/kratos/commit/a43cef23c177acddbf8b03afef087feeaca51981)):

    This extends the existing `DELETE /admin/identities/{id}/credentials/{type}` API to accept an `?identifier=foobar` query parameter for `{type}==oidc` like such:
    
    `DELETE /admin/identities/{id}/credentials/oidc?identifier=github%3A012345`
    
    This will delete the GitHub OIDC credential with the identifier `github:012345` (`012345` is the subject as returned by GitHub).
    
    To find out which OIDC credentials exist, call `GET /admin/identities/{id}?include_credential=oidc` beforehand.
    
    This will allow you to delete individual OIDC credentials for users even if they have several set up.

* Allow partially failing batch inserts ([#4083](https://github.com/ory/kratos/issues/4083)) ([4ba7033](https://github.com/ory/kratos/commit/4ba70330cf9e0eda9044b0a5a504c34493ae17ed)):

    When batch-inserting multiple identities, conflicts or validation errors of a subset of identities in the batch still allow the rest of the identities to be inserted. The returned JSON contains the error details that lead to the failure.

* Better detection if credentials exist on identifier first login ([#3963](https://github.com/ory/kratos/issues/3963)) ([42ade94](https://github.com/ory/kratos/commit/42ade94e32a9a7ad6c0bda785e86d7209c46d8bb))
* Change `method=profile:back` to `screen=previous` ([#4119](https://github.com/ory/kratos/issues/4119)) ([2cd8483](https://github.com/ory/kratos/commit/2cd8483e809170d0524fe6a5d13837108d29fa54))
* Clarify session extend behavior ([#3962](https://github.com/ory/kratos/issues/3962)) ([af5ea35](https://github.com/ory/kratos/commit/af5ea35759e74d7a1637823abcc21dc8e3e39a9d))
* Client-side PKCE take 3 ([#4078](https://github.com/ory/kratos/issues/4078)) ([f7c1024](https://github.com/ory/kratos/commit/f7c102456a71b226d8353b9d59cc03fb2ba0af40)):

    * feat: client-side PKCE
    
    This change introduces a new configuration for OIDC providers: pkce with values auto (default), never, force.
    
    When auto is specified or the field is omitted, Kratos will perform autodiscovery and perform PKCE when the server advertises support for it. This requires the issuer_url to be set for the provider.
    
    never completely disables PKCE support. This is only theoretically useful: when a provider advertises PKCE support but doesn't actually implement it.
    
    force always sends a PKCE challenge in the initial redirect URL, regardless of what the provider advertises. This setting is useful when the provider offers PKCE but doesn't advertise it in his ./well-known/openid-configuration.
    
    Important: When setting pkce: force, you must whitelist a different return URL for your OAuth2 client in the provider's configuration. Instead of <base-url>/self-service/methods/oidc/callback/<provider>, you must use <base-url>/self-service/methods/oidc/callback (note missing last path segment). This is to enable the use of the same OAuth client ID+secret when configuring several Kratos OIDC providers, without having to whitelist individual redirect_uris for each Kratos provider config.
    
    * chore: regenerate SDK, bump DB versions, cleanup tool install
    
    * chore: get final organization ID from provider config during registration and login
    
    * chore: fixup OIDC function signatures and improve tests

* Emit events in identity persister ([#4107](https://github.com/ory/kratos/issues/4107)) ([20156f6](https://github.com/ory/kratos/commit/20156f651f2faa0a79842de8d2fb4a09ee7094c1))
* Enable new-style OIDC state generation ([#4121](https://github.com/ory/kratos/issues/4121)) ([eb97243](https://github.com/ory/kratos/commit/eb97243d6499e2d9f2338a2ce3f5e39579d19086))
* Identifier first auth ([1bdc19a](https://github.com/ory/kratos/commit/1bdc19ae3e1a3df38234cb892f65de4a2c95f041))
* Identifier first login for all first factor login methods ([638b274](https://github.com/ory/kratos/commit/638b27431312bcd91844ac4a00733a840976aa4f))
* Improve session extend performance ([#3948](https://github.com/ory/kratos/issues/3948)) ([4e3fad4](https://github.com/ory/kratos/commit/4e3fad4b4739b5cf00d658155350cb599f2cd06a)):

    This patch improves the performance for extending session lifespans. Lifespan extension is tricky as it is often part of the middleware of Ory Kratos consumers. As such, it is prone to transaction contention when we read and write to the same session row at the same time (and potentially multiple times).
    
    To address this, we:
    
    1. Introduce a locking mechanism on the row to reduce transaction contention;
    2. Add a new feature flag that toggles returning 204 no content instead of 200 + session.
    
    Be aware that all reads on the session table will have to wait for the transaction to commit before they return a value. This may cause long(er) response times on `/session/whoami` for sessions that are being extended at the same time.

* Password migration hook ([#3978](https://github.com/ory/kratos/issues/3978)) ([c9d5573](https://github.com/ory/kratos/commit/c9d55730a10b71ac61bb5097f5f9c33f144f2a95)):

    This adds a password migration hook to easily migrate passwords for which we do not have the hash.
    
    For each user that needs to be migrated to Ory Network, a new identity is created with a credential of type password with a config of {"use_password_migration_hook": true} .
    When a user logs in, the credential identifier and password will be sent to the password_migration web hook if all of these are true:
    The user’s identity’s password credential is {"use_password_migration_hook": true}
    The password_migration hook is configured
    After calling the password_migration hook, the HTTP status code will be inspected:
    On 200, we parse the response as JSON and look for {"status": "password_match"}. The password credential config will be replaced with the hash of the actual password.
    On any other status code, we assume that the password is not valid.

* **sdk:** Add missing profile discriminator to update registration ([0150795](https://github.com/ory/kratos/commit/0150795d902dcc7cfb2298c3b5a98da1c2541e46))
* **sdk:** Avoid eval with javascript triggers ([dd6e53d](https://github.com/ory/kratos/commit/dd6e53d62f343a317edf403218b20599539218c6)):

    Using `OnLoadTrigger` and `OnClickTrigger` one can now map the trigger to the corresponding JavaScript function.
    
    For example, trigger `{"on_click_trigger":"oryWebAuthnRegistration"}` should be translated to `window.oryWebAuthnRegistration()`:
    
    ```
    if (attrs.onClickTrigger) {
      window[attrs.onClickTrigger]()
    }
    ```

* Separate 2fa refresh from 1st factor refresh ([#3961](https://github.com/ory/kratos/issues/3961)) ([89355d8](https://github.com/ory/kratos/commit/89355d86258ace19c03fcb38dd3861f88e28af59))
* Set maxlength for totp input ([51042d9](https://github.com/ory/kratos/commit/51042d99fab301f0bb44665e56c5a2364e7d8866))

### Tests

* Add form hydration tests for code login ([37781a9](https://github.com/ory/kratos/commit/37781a93dda9b8f0127217a6b0ac2434dda1cc58))
* Add form hydration tests for idfirst login ([633b0ba](https://github.com/ory/kratos/commit/633b0ba7f724374f4c02128a5b0f748bd2e9413e))
* Add form hydration tests for oidc login ([df0cdcb](https://github.com/ory/kratos/commit/df0cdcb424cae6c49143ef2ef2d0b2c95f14fffb))
* Add form hydration tests for passkey login ([a777854](https://github.com/ory/kratos/commit/a777854e8d99336ab8f5755fdbc9d257e5edd1c0))
* Add form hydration tests for password login ([7186e7e](https://github.com/ory/kratos/commit/7186e7e060e04a4918e22e0b03fefbf4eb9f4a4b))
* Add form hydration tests for webauthn login ([8b68163](https://github.com/ory/kratos/commit/8b68163a3f293f7dceb58397f0ef555f1d8fd7c3))
* Add tests for idfirst ([5f76c15](https://github.com/ory/kratos/commit/5f76c1565e89bfb99f23c3f0f3a9beadbdfa270c))
* Additional code credential test case ([#4122](https://github.com/ory/kratos/issues/4122)) ([4f2c854](https://github.com/ory/kratos/commit/4f2c8542ab04b88c7112d7b564d91bcfd8f5791a))
* Deflake and parallelize persister tests ([#3953](https://github.com/ory/kratos/issues/3953)) ([61f87d9](https://github.com/ory/kratos/commit/61f87d90bd67e5bb1f00ee110d986e4f72fc4c91))
* Deflake session extend config side-effect ([#3950](https://github.com/ory/kratos/issues/3950)) ([b192c92](https://github.com/ory/kratos/commit/b192c92d6c969d470d6479bc33dbc351d327c1f9))
* Enable server-side config from context ([#3954](https://github.com/ory/kratos/issues/3954)) ([e0001b0](https://github.com/ory/kratos/commit/e0001b0db784457652581366bd7ead7cdf6b3898))
* Improve stability of refresh test ([#4037](https://github.com/ory/kratos/issues/4037)) ([68693a4](https://github.com/ory/kratos/commit/68693a43e4e1e3028f17789e72d0b79f6298d139))
* Resolve CI failures ([#4067](https://github.com/ory/kratos/issues/4067)) ([dbf7274](https://github.com/ory/kratos/commit/dbf7274f7a4be56c33b06559875c42725bf4a351))
* Resolve issues and update snapshots for all selfservice strategies ([e2e81ac](https://github.com/ory/kratos/commit/e2e81ac16726b180d33c57913e3cac099daf946b))
* Update incorrect usage of Auth0 in Salesforce tests ([#4007](https://github.com/ory/kratos/issues/4007)) ([6ce3068](https://github.com/ory/kratos/commit/6ce306824cec81890c50dcf23c2b8a5825f20a10))
* Verify redirect continue_with in hook executor for browser clients ([7b0b94d](https://github.com/ory/kratos/commit/7b0b94d30ec9069de6978427814d55a30e62adb8))

### Unclassified

* Merge commit from fork ([123e807](https://github.com/ory/kratos/commit/123e80782b392095631ee2e0d1bd6ec337c1fb79)):

    * fix(security): code credential does not respect `highest_available` setting
    
    This patch fixes a security vulnerability which prevents the `code` method to properly report it's credentials count to the `highest_available` mechanism.
    
    For more details on this issue please refer to the [security advisory](https://github.com/ory/kratos/security/advisories/GHSA-wc43-73w7-x2f5).
    
    * fix: normalize code credentials and deprecate via parameter
    
    Before this, code credentials for passwordless and mfa login were incorrectly stored and normalized. This could cause issues where the system would not detect the user's phone number, and where SMS/email MFA would not properly work with the `highest_available` setting.

* Update .github/workflows/ci.yaml ([2d60772](https://github.com/ory/kratos/commit/2d60772062a684c3a27f28b8836c3548f5b8cea9))
* Update Code QL action to v2 ([#4008](https://github.com/ory/kratos/issues/4008)) ([e3f1da0](https://github.com/ory/kratos/commit/e3f1da0f4bf41a8a8733758fcd9edb9910c55cfa))


# [1.2.0](https://github.com/ory/kratos/compare/v1.1.0...v1.2.0) (2024-06-05)

Ory Kratos v1.2 is the most complete, scalable, and secure open-source identity server available. We are thrilled to announce its release!

![Ory Kratos 1.2 released](https://www.ory.sh/images/newsletter/kratos-1.2.0/banner.png)

This release introduces two major features: two-step registration and full PassKey with resident key support.

Passkeys provide a secure and convenient authentication method, eliminating the need for passwords while ensuring strong security. With this release, we have added support for resident keys, enabling offline authentication. Credential discovery allows users to link existing passkeys to their Ory account seamlessly.

[Watch the PassKey demo video](https://github.com/aeneasr/web-next-deprecated/assets/3372410/e676c518-c82a-42a6-821e-28aecadb270c)

Two-step registration improves the user experience by dividing the registration process into two steps. Users first enter their identity traits, and then choose a credential method for authentication, resulting in a streamlined process. This feature is especially useful when enabling multiple authentication strategies, as it eliminates the need to repeat identity traits for each strategy.

![Two-Step Registration](https://ik.imagekit.io/launchnotes/production/tr:w-1640,c-at_max,f-auto/ngul9dzfjdt3pe8benegjjeeagi1)

The 107 commits since v1.1 include several improvements:
- **Webhooks** now carry session information if available.
- **Transient Payloads** are now available across all self-service flows.
- **Sign in with Twitter** is now available.
- **Sign in with LinkedIn** now includes an additional v2 provider compatible with LinkedIn's new SSO API.
- **Two-Step Registration**: An improved registration experience that separates entering profile information from choosing authentication methods.
- **User Credentials Meta-Information** can now be included on the list endpoint.
- **Social Sign-In** is now resilient to double-submit issues common with Facebook and Apple mobile login.

**Two-Step Registration Enabled by Default**: This is now the default setting. To disable, set `selfservice.flows.registration.enable_legacy_flow` to `true`.

- Improved account linking and credential discovery during sign-up.
- The `return_to` parameter is now respected in OIDC API flows.
- Adjustments to database indices.
- Enhanced error messages for security violations.
- Improved SDK types.
- The `verification` and `verification_ui` hooks are now available in the login flow.
- Webhooks now contain the correct identity state in the after-verification hook chain.

We are doing this survey to find out how we can support self-hosted Ory users better. We strive to provide you with the best product and service possible and your feedback will help us understand what we're doing well and where we can improve to better meet your needs. We truly value your opinion and thank you in advance for taking the time to share your thoughts with us!

Fill out the [survey now](https://share-eu1.hsforms.com/15DiCnJpcRuijnpAdnDhxxwextgn)!



## Breaking Changes

This feature enables two-step registration per default. Two-step registration is a significantly improved sign up flow and recommended when using more than one sign up methods. To disable two-step registration, set `selfservice.flows.registration.enable_legacy_flow` to `true`. This value defaults to `false`.



### Bug Fixes

* Add login succeeded event to post registration hook ([#3739](https://github.com/ory/kratos/issues/3739)) ([b685fa5](https://github.com/ory/kratos/commit/b685fa5477be2ba099fd2420b27b2411fafc7e51))
* Add missing env vars to set up guide ([#3855](https://github.com/ory/kratos/issues/3855)) ([da90502](https://github.com/ory/kratos/commit/da90502dc3bf8e3d34fb4ecc531834b1919989ad)):

    Closes https://github.com/ory/kratos/issues/3828

* Add missing indexes and remove unused index ([6d7372e](https://github.com/ory/kratos/commit/6d7372ee3d88ee4fc552b969dd0ff338dcc0544c))
* Add missing indexes and remove unused index ([#3756](https://github.com/ory/kratos/issues/3756)) ([c905f02](https://github.com/ory/kratos/commit/c905f02473c5d77ab309a45f10251b1ba7e88584))
* Add sms mfa via parameter to spec ([#3766](https://github.com/ory/kratos/issues/3766)) ([b291c95](https://github.com/ory/kratos/commit/b291c959c18c72f5edc55607ab23b4592faf8d53))
* Allow updating just the verified_at timestamp of addresses ([#3880](https://github.com/ory/kratos/issues/3880)) ([696cc1b](https://github.com/ory/kratos/commit/696cc1b59b18627fec63915070f4d8c5b3e3250d))
* Always issue session last ([#3876](https://github.com/ory/kratos/issues/3876)) ([e942507](https://github.com/ory/kratos/commit/e94250705e999567e2ed58cebdb3f6a9d589e3ef)):

    In post persist hooks, the session issuance hook always needs
    to come last. This fixes the getHooks function to ensure this.

* Audit issues ([#3797](https://github.com/ory/kratos/issues/3797)) ([7017490](https://github.com/ory/kratos/commit/7017490caa9c70e22d5c626773c0266521813ff5))
* Change return urls in quickstarts ([#3928](https://github.com/ory/kratos/issues/3928)) ([9730e09](https://github.com/ory/kratos/commit/9730e099a656d211389d8e993c64d8082784c929))
* Close res body ([#3870](https://github.com/ory/kratos/issues/3870)) ([cc39f8d](https://github.com/ory/kratos/commit/cc39f8df7c235af0df616432bc4f88681896ad85))
* CVEs in dependencies ([#3902](https://github.com/ory/kratos/issues/3902)) ([e5d3b0a](https://github.com/ory/kratos/commit/e5d3b0afde3c80c6c9cf8815c56d82e291ede663))
* Db index and duplicate credentials error ([#3896](https://github.com/ory/kratos/issues/3896)) ([9f34a21](https://github.com/ory/kratos/commit/9f34a21ea2035a5d33edd96753023a3c8c6c054c)):

    * fix: don't return password cred type if empty
    * fix: better index for config.user_handle on identity_credentials

* Do not require method to be  passkey in settings schema ([#3862](https://github.com/ory/kratos/issues/3862)) ([660f330](https://github.com/ory/kratos/commit/660f330ab69ef0e6fd21501fbc9dfed693d4a715))
* Don't require connection_uri in SMTP ([#3861](https://github.com/ory/kratos/issues/3861)) ([800f8f1](https://github.com/ory/kratos/commit/800f8f1036ef46a561d24dcdec45dd48803978d7))
* Don't treat passkeys as AAL2 ([#3853](https://github.com/ory/kratos/issues/3853)) ([8eee972](https://github.com/ory/kratos/commit/8eee972d89accb02b3caa053fca2f16ed2c876f1))
* Drop index if exists ([#3846](https://github.com/ory/kratos/issues/3846)) ([ad0619d](https://github.com/ory/kratos/commit/ad0619d803cd2842a67c56a545ec5ab252501b0f))
* Drop trigram index on identifiers ([#3827](https://github.com/ory/kratos/issues/3827)) ([8f8fd90](https://github.com/ory/kratos/commit/8f8fd90304886ecd689a85fc60c4712e47526cdd))
* Enum type of session expandables ([#3891](https://github.com/ory/kratos/issues/3891)) ([63d785e](https://github.com/ory/kratos/commit/63d785e5e73ff067ec804ecc2107fac1525d3688))
* Enum type of session expandables ([#3895](https://github.com/ory/kratos/issues/3895)) ([c435727](https://github.com/ory/kratos/commit/c435727c1e3c70c040b7fc7648ce621b136e5fc2))
* Execute verification & verification_ui properly in login flows ([#3847](https://github.com/ory/kratos/issues/3847)) ([5aad1c1](https://github.com/ory/kratos/commit/5aad1c1e6cc92f72af56511dacb9812edb600813))
* Ignore decrypt errors in WithDeclassifiedCredentials ([#3731](https://github.com/ory/kratos/issues/3731)) ([8f5192f](https://github.com/ory/kratos/commit/8f5192fbb74c4b952029a6856284de8d59027770))
* Improve SDK discriminators ([#3844](https://github.com/ory/kratos/issues/3844)) ([c08b3ad](https://github.com/ory/kratos/commit/c08b3ad76c5adb712c945cdbd92a9a51832e94b9))
* Include all creds in duplicate credential err ([#3881](https://github.com/ory/kratos/issues/3881)) ([e06c241](https://github.com/ory/kratos/commit/e06c241ffe3f0e696bb1cbc1d1080f9d4e09fbd2))
* Linkedin issuer override ([#3875](https://github.com/ory/kratos/issues/3875)) ([11d221a](https://github.com/ory/kratos/commit/11d221a4d33878930ca7025ae1b5c18b25dd1add))
* Make sure emails can still be sent with SMS enabled ([#3795](https://github.com/ory/kratos/issues/3795)) ([7c68c5a](https://github.com/ory/kratos/commit/7c68c5aa69ed76a84a37a37a3555277ddc772cf8))
* Missing indices and foreign keys ([#3800](https://github.com/ory/kratos/issues/3800)) ([0b32ce1](https://github.com/ory/kratos/commit/0b32ce113be47aa724d3468062ced09f8f60c52a))
* **oidc:** Grace period for continuity container on oidc callbacks ([#3915](https://github.com/ory/kratos/issues/3915)) ([1a9a096](https://github.com/ory/kratos/commit/1a9a096d619925dd3718ad9dd9daf77387572ece))
* Passing transient payloads ([#3838](https://github.com/ory/kratos/issues/3838)) ([d01b670](https://github.com/ory/kratos/commit/d01b6705bf36efb6e0f3d71ed22d0574ab8a98a4))
* Prevent SMTP URL leak on unparsable URL ([#3770](https://github.com/ory/kratos/issues/3770)) ([c5f39f4](https://github.com/ory/kratos/commit/c5f39f4bc481e400f736ede7f8f0be546a55eebf))
* Respect return_to in OIDC API flow error case ([#3893](https://github.com/ory/kratos/issues/3893)) ([e8f1bcb](https://github.com/ory/kratos/commit/e8f1bcb1342af994b8e08282aa4066ee00ffe7d4)):

    * fix: respect return_to in OIDC API flow error case
    
    This fix ensures that we redirect the user to the return_to URL
    when an error occurs during the OIDC login for native flows.
    
    Native flows are initialized through the API, and the browser
    URL is retrieved from a 422 response after a POST to submit the
    login flow. Successful OIDC flows already returned the `code` to
    the `return_to` URL. Now, unsuccessful flows return the `flow` with
    the current flow ID (which might have changed), so that the caller
    can retrieve the full flow and act accordingly.
    
    * fix: ignore trivvy CVE report
    
    Bump in distroless is still open

* **sdk:** Expand identity in session extension ([#3843](https://github.com/ory/kratos/issues/3843)) ([04f0231](https://github.com/ory/kratos/commit/04f02318d4de5290cbf100e9b301284d5ee40fe7)), closes [#3842](https://github.com/ory/kratos/issues/3842)
* **sdk:** Improve discriminators for node and Go ([#3821](https://github.com/ory/kratos/issues/3821)) ([9ddf7cc](https://github.com/ory/kratos/commit/9ddf7cc7c52313c4ee13ccdc2886ad94b5d1317f))
* Show error page on identity mismatch ([#3790](https://github.com/ory/kratos/issues/3790)) ([e6db689](https://github.com/ory/kratos/commit/e6db689e0de41067e6e78889c3dab9637a96236e))
* Test assertions on declassifying OIDC tokens ([#3773](https://github.com/ory/kratos/issues/3773)) ([7f8a7f1](https://github.com/ory/kratos/commit/7f8a7f142a91c8c74f32eadb41224fc4f69c2109))
* Tolerate more "truthy" values when creating new flows ([#3841](https://github.com/ory/kratos/issues/3841)) ([49d93c0](https://github.com/ory/kratos/commit/49d93c0e3383f602fe6be3c7bf749b54f344aa72)), closes [#3839](https://github.com/ory/kratos/issues/3839):

    Use strconv.ParseBool to accept multiple "truthy" values for the
    `refresh` and `return_session_token_exchange_code` query parameters when
    creating a new login flow.
    
    For some SDKs (e.g.: Python), these stringification of booleans is not
    user-controlled and these endpoints could not be used fully due to the
    backend ignoring any value other than `true` (all lowercase).

* Tweaks to UpsertSessions ([#3878](https://github.com/ory/kratos/issues/3878)) ([da51dcd](https://github.com/ory/kratos/commit/da51dcdb8c82a5dbd290ab2f48ad74a1c6dd18f0))
* Use correct post-verification identity state in post-hooks ([#3863](https://github.com/ory/kratos/issues/3863)) ([6e63d06](https://github.com/ory/kratos/commit/6e63d06db1cd1ab62f8a2d0b202ec74572420204))
* Webhook transient payload in OIDC login flows ([#3857](https://github.com/ory/kratos/issues/3857)) ([2cdfc70](https://github.com/ory/kratos/commit/2cdfc70c726a166790b98d419895f0396d13176f)):

    * fix: transient payload with OIDC login


### Code Generation

* Pin v1.2.0 release commit ([1a70648](https://github.com/ory/kratos/commit/1a70648c4d5b9b8d135dd7bea3842057e67b574e))

### Documentation

* Remove delete reference from batch patch identity ([#3906](https://github.com/ory/kratos/issues/3906)) ([cd01cb9](https://github.com/ory/kratos/commit/cd01cb9fb23a24e52d46538a9ea63c2144c3b145))

### Features

* Add `include_credential` query param to `/admin/identities` list call ([#3343](https://github.com/ory/kratos/issues/3343)) ([d94530a](https://github.com/ory/kratos/commit/d94530a716358895b01b65babd77226fab69f494))
* Add headers to web hooks ([#3849](https://github.com/ory/kratos/issues/3849)) ([4642de0](https://github.com/ory/kratos/commit/4642de0cfd1fb15bc48c7093be9449abd488755c))
* Add session to post login webhook ([#3877](https://github.com/ory/kratos/issues/3877)) ([386078e](https://github.com/ory/kratos/commit/386078e0b5c74c54ce2c7dc6fd12fd865817b87a))
* Add transient payloads to all flows ([#3738](https://github.com/ory/kratos/issues/3738)) ([b8b747b](https://github.com/ory/kratos/commit/b8b747b2adc59c8cf938a0ee30accdb4135634b8))
* Add twitter SSO ([#3778](https://github.com/ory/kratos/issues/3778)) ([930fb19](https://github.com/ory/kratos/commit/930fb19842e527e5e9c415efa983b36e02829516))
* Add verification hook to login flow ([#3829](https://github.com/ory/kratos/issues/3829)) ([43e4ead](https://github.com/ory/kratos/commit/43e4eadce7fa6e66bf1f9c03136d141bffd3094f))
* Allow admin to create API code recovery flows ([#3939](https://github.com/ory/kratos/issues/3939)) ([25d1ecd](https://github.com/ory/kratos/commit/25d1ecd90317193095e01b97ff21d92920035b02))
* Control edge cache ttl ([#3808](https://github.com/ory/kratos/issues/3808)) ([c9dcce5](https://github.com/ory/kratos/commit/c9dcce5a41137937df1aad7ac81170b443740f88))
* Linkedin v2 provider ([#3804](https://github.com/ory/kratos/issues/3804)) ([a6ad983](https://github.com/ory/kratos/commit/a6ad983ac83aa3ea65c4dc0c46b582096574c25a)):

    * feat: add linkedin-v2 provider
    
    * docs: document linkedin special-case

* PassKeys with Resident Keys and two-step registration ([#3748](https://github.com/ory/kratos/issues/3748)) ([3621411](https://github.com/ory/kratos/commit/3621411dc4386d841bc6766a5ab8d03e65812073))
* Send OIDC claim keys to tracing ([#3798](https://github.com/ory/kratos/issues/3798)) ([04390be](https://github.com/ory/kratos/commit/04390bee426befe51af2ee8177afabaa9ce4fa80))
* Use authenticate endpoint for x ([#3833](https://github.com/ory/kratos/issues/3833)) ([3d9ba5d](https://github.com/ory/kratos/commit/3d9ba5df85e0d0c4d8002365987e536b37678104)):

    Improves the "Log in with X" experience by not asking the user to re-authenticate every time.


### Tests

* Deflake session test ([#3864](https://github.com/ory/kratos/issues/3864)) ([6b275f3](https://github.com/ory/kratos/commit/6b275f35a0732ffb723d47df5b6afbdc06eaf71f))
* Resolve failing test for empty tokens ([#3775](https://github.com/ory/kratos/issues/3775)) ([7277368](https://github.com/ory/kratos/commit/7277368bc28df8f0badffc7e739cef20f05e9a02))
* Resolve flaky e2e tests ([#3935](https://github.com/ory/kratos/issues/3935)) ([a14927d](https://github.com/ory/kratos/commit/a14927dfa5f8d0fbda7e5a831f0a09a42369e06c)):

    * test: resolve flaky code registration tests
    
    * chore: don't fail logout if cookie is not found
    
    * chore: remove .only
    
    * chore: reduce wait
    
    * chore: u
    
    * chore: u
    
    * chore: u


### Unclassified

* Remove unnecessary COPY command from Dockerfile (#3771) ([087748c](https://github.com/ory/kratos/commit/087748c0651ff0fc93259f7ab6b10668c09f5eba)), closes [#3771](https://github.com/ory/kratos/issues/3771)


# [1.1.0](https://github.com/ory/kratos/compare/v1.0.0...v1.1.0) (2024-02-20)

![Ory Kratos v1.1.0](https://www.ory.sh/images/newsletter/kratos-1.1.0/banner.png)

Ory Kratos v1.1 is the most complete, most scalable, and most secure open-source identity server on the planet, and we are thrilled to announce its release! This release comes with over 270 commits and an incredible amount of new features and capabilities!

- **Phone Verification & 2FA with SMS**: Enhance convenient security with phone verification and two-factor authentication (2FA) via SMS, integrating easily with SMS gateways like Twilio. This feature not only adds a convenient layer of security but also offers a straightforward method for user verification, increasing your trust in user accounts.
- **Translations & Internationalization**: Ory Kratos now supports multiple languages, making it accessible to a global audience. This improvement enhances the user experience by providing a localized interface, ensuring users interact with the system in their preferred language.
- **Native Support for Sign in with Google and Apple on Android/iOS**: Get more sign-ups with native support for "Sign in with Google" and "Sign in with Apple" on mobile platforms. Great user experience matters!
- **Account Linking**: Simplify user management with new features that facilitate account linking. If a user registers with a password and later signs in with a social account sharing the same email, new screens make account linking straightforward, enhancing user convenience and reducing support inquiries.
- **Passwordless "Magic Code"**: Introduce a passwordless login method with "Magic Code," which sends a one-time code to the user's email for sign-up and login. This method can also serve as a fallback when users forget their password or their social login is unavailable, streamlining the login process and improving user accessibility.
- **Session to JWT Conversion**: Convert an Ory Session Cookie or Ory Session Token into a JSON Web Token (JWT), providing more flexibility in handling sessions and integrating with other systems. This feature allows for seamless authentication and authorization processes across different platforms and services.

**Note:** To ensure a seamless upgrade experience with minimal impact, some of these features are gated behind the `feature_flags` config parameter, allowing controlled deployment and testing.

The following features have been shipped exclusively to Ory Network for this version:

- **[B2B SSO](https://www.ory.sh/docs/kratos/organizations)** allows your customers to connect their LDAP / Okta / AD / … to your login. Ory selects the correct login provider based on the user’s email domain.
- [**Significantly better API performance](https://www.ory.sh/docs/api/eventual-consistency)** for expensive API operations by specifying the desired consistency (`strong`, `eventual`).
- **Finding users effortlessly** with our new fuzzy search for credential identifiers available for the [Identity List API](https://www.ory.sh/docs/kratos/reference/api#tag/identity/operation/listIdentities).

- Better reliability when sending out emails across different providers.
- Streamlining the HTTP API and improving related SDK methods.
- Better performance when calling the whoami API endpoint, updating identities, and listing identities.
- The performance of listing identities has significantly improved with the introduction of keyset pagination. Page pagination is still available but will be fully deprecated soon.
- Ability to list multiple identities in a batch call.
- Passkeys and WebAuthn now support multiple origins, useful when working with subdomains.
- The logout flow now redirects the user back to the `return_to` parameter set in the API call.
- When updating their settings, the user was sometimes incorrectly asked to confirm the changes by providing their password. This issue has now been fixed.
- When signing up with an account that already exists, the user will be shown a hint helping them sign in to their existing account.
- CORS configuration can now be hot-reloaded.
- The integration with Ory OAuth2 / Ory Hydra has improved for logout, login session management, verification, and recovery flows.
- A new passwordless method has been added: "Magic code". It sends a one-time code to the user's email during sign-up and log-in. This method can additionally be used as a fallback login method when the user forgets their password.
- Integration with social sign-in has improved, and it is now possible to use the email verified status from the social sign-in provider.
- Ory Elements and the default Ory Account Experience are now internationalized with translations.
- It is now possible to convert an Ory Session Cookie or Ory Session Token into a JSON Web Token.
- Recovery on native apps has improved significantly and no longer requires the user to switch to a browser for the recovery step.
- Administrators can now find users by their identifiers with fuzzy search - this feature is still in preview.
- Importing HMAC-hashed passwords is now possible.
- Webhooks can now update identity admin metadata.
- New screens have been added to make account linking possible when a user has registered with a password and later tries signing in with a social account sharing the same email.
- Ability to revoke all sessions of a user when they change their password.
- Webhooks are now available for all login, registration, and login methods, including Passkeys, TOTP, and others.
- The login screen now longer shows “ID” for the primary identifier, but instead extracts the correct label - for example, “Email” or “Username” from the Identity Schema.
- Login hints help users with guidance when they are unable to sign in (wrong social sign-in provider) but have an active account.
- Phone numbers can now be verified via an SMS gateway like Twilio.
- SMS OTP is now a two-factor option.
Ory Kratos 1.1 is a major release that marks a significant milestone in our journey.

We sincerely hope that you find these new features and improvements in Ory Kratos 1.1 valuable for your projects. To experience the power of the latest release, we encourage you to get the latest version of Ory Kratos [here](https://github.com/ory/kratos) or leverage Ory Kratos in [Ory Network](https://www.ory.sh/network/) — the easiest, simplest, and most cost-effective way to run Ory.

For organizations seeking to upgrade their self-hosted solution, **Ory offers enterprise support services to ensure a smooth transition**. Our team is ready to assist you throughout the migration process, ensuring uninterrupted access to the latest features and improvements. Additionally, we provide various [support plans](https://www.ory.sh/support/) specifically tailored for self-hosting organizations. These plans offer comprehensive assistance and guidance to optimize your Ory deployments and meet your unique requirements.
We extend our heartfelt gratitude to the vibrant and supportive Ory Community. Without your constant support, feedback, and contributions, reaching this significant milestone would not have been possible. As we continue on this journey, your feedback and suggestions are invaluable to us. Together, we are shaping the future of identity management and authentication in the digital landscape.

Contributors to this release in no particular order: [moose115](https://github.com/ory/kratos/commits?author=moose115), [K3das](https://github.com/ory/kratos/commits?author=K3das), [sidartha](https://github.com/ory/kratos/commits?author=sidartha), [efesler](https://github.com/ory/kratos/commits?author=efesler), [BrandonNoad](https://github.com/ory/kratos/commits?author=BrandonNoad) ,[Saancreed](https://github.com/ory/kratos/commits?author=Saancreed), [jpogorzelski](https://github.com/ory/kratos/commits?author=jpogorzelski), [dreksx](https://github.com/ory/kratos/commits?author=dreksx), [martinloesethjensen](https://github.com/ory/kratos/commits?author=martinloesethjensen), [cpoyatos1](https://github.com/ory/kratos/commits?author=cpoyatos1), [misamu](https://github.com/ory/kratos/commits?author=misamu), [tristankenney](https://github.com/ory/kratos/commits?author=tristankenney), [nxy7](https://github.com/ory/kratos/commits?author=nxy7), [anhnmt](https://github.com/ory/kratos/commits?author=anhnmt)

Are you passionate about security and want to make a meaningful impact in one of the biggest open-source communities? Join the [Ory community](https://slack.ory.sh/) and become a part of the new ID stack. Together, we are building the next generation of IAM solutions that empower organizations and individuals to secure their identities effectively.
Want to check out Ory Kratos yourself? Use these commands to get your Ory Kratos project running on the Ory Network:
```
brew install ory/tap/cli

scoop bucket add ory <https://github.com/ory/scoop.git>
scoop install ory

bash <(curl <https://raw.githubusercontent.com/ory/meta/master/install.sh>) -b . ory
sudo mv ./ory /usr/local/bin/

ory auth login

ory create project --name "My first Kratos project"

ory open account-experience registration

ory patch identity-config \
  --replace '/identity/default_schema_id="preset://username"' \
  --replace '/identity/schemas=[{"id":"preset://username","url":"preset://username"}]' \
  --format yaml

ory open account-experience registration
```



## Breaking Changes

Pagination parameters for the `list identities` CLI command have changed from arguments to flags `--page-token` and `page-size`:

```
- kratos list identities 1 100
+ kratos list identities --page-size 100 --page-token ...
```

Furthermore, the JSON / JSON pretty output of `list identities` has changed:

```patch
-[
-  { "id": "..." },
-  { /* ... */ },
-  // ...
-]
+{
+  "identities": [
+    {"id": "..."},
+    { /* ... */ },
+    // ...
+  ],
+  "next_page_token": "..."
+}
```

Closes https://github.com/ory/sdk/issues/284
Closes https://github.com/ory/kratos/pull/3480



### Bug Fixes

* `oidc` does not require a method in the payload ([#3564](https://github.com/ory/kratos/issues/3564)) ([b299abc](https://github.com/ory/kratos/commit/b299abcfa1ebdb8bbb6bb9339f61873d5c77c44f)):

    * fix: `oidc` does not require a method in the payload
    
    * refactor: only update strategies order in test
    
    * chore: update audit messages and comments

* Accept all 200 responses as OK in courier ([#3401](https://github.com/ory/kratos/issues/3401)) ([88237e2](https://github.com/ory/kratos/commit/88237e25b080a9643f6cbf7eedbf23988ba9ba7c)), closes [#3399](https://github.com/ory/kratos/issues/3399):

    * fix: accept all 200 responses as OK in courier

* Accept login_challenge after verification ([#3427](https://github.com/ory/kratos/issues/3427)) ([6b02350](https://github.com/ory/kratos/commit/6b02350c21aa65decd1bb16e559e1cc7dae42d55)):

    Part of https://github.com/ory/network/issues/320

* Add caching to Jsonnet snippet during session JWT tokenization ([#3699](https://github.com/ory/kratos/issues/3699)) ([1da8180](https://github.com/ory/kratos/commit/1da818072154baa5c0921134919afde595031e94))
* Add consistency flag ([#3733](https://github.com/ory/kratos/issues/3733)) ([fd79950](https://github.com/ory/kratos/commit/fd7995077307cc101550eda5d7724ea1f68fa98a))
* Add max-age to default cors headers ([#3584](https://github.com/ory/kratos/issues/3584)) ([c5b4aaa](https://github.com/ory/kratos/commit/c5b4aaa2df5d010b62a99ccf45850583daad3a66))
* Add missing tracing & attributes in oidc strategy ([#3429](https://github.com/ory/kratos/issues/3429)) ([09bcb71](https://github.com/ory/kratos/commit/09bcb71f1f0b3238e2d0f4376a1a2290d062c6c1))
* Add return_to parameter to API spec of createRecoveryLinkForIdentity ([#3711](https://github.com/ory/kratos/issues/3711)) ([757a5e4](https://github.com/ory/kratos/commit/757a5e43257e9ff28a16bfe76f8e737b656d3696))
* Add value code to authentication method enum ([#3546](https://github.com/ory/kratos/issues/3546)) ([95dc7a2](https://github.com/ory/kratos/commit/95dc7a20f49aa682f324b70e507ec56c20159ebb)):

    * fix: add value code to authentication method enum
    
    * chore: generate sdk

* Additional_id_token_audiences key in config schema ([#3622](https://github.com/ory/kratos/issues/3622)) ([9396bb0](https://github.com/ory/kratos/commit/9396bb0b586d1d1e74a85c0ae3bcf9de81214f1b))
* Adjust tracing verbosity ([976cd0d](https://github.com/ory/kratos/commit/976cd0dc3dd95c2c1992bfa82394e9fad39f34f2))
* Allow post recovery hooks to interrupt the flow ([#3393](https://github.com/ory/kratos/issues/3393)) ([6c1d2f1](https://github.com/ory/kratos/commit/6c1d2f1e4173cfb9a7abe2bfe4f20e47b7568d3b))
* Allow updating admin metadata from webhook responses ([#3569](https://github.com/ory/kratos/issues/3569)) ([22f61f0](https://github.com/ory/kratos/commit/22f61f015495c55e58db4f31ee6882444b9a3caf))
* Always return relative URLs in the Link header for pagination ([fb229c9](https://github.com/ory/kratos/commit/fb229c982c6f7d7a4f5f0f84ffc971a576906160))
* Auto migrate old accounts to use code credential ([#3581](https://github.com/ory/kratos/issues/3581)) ([569b14a](https://github.com/ory/kratos/commit/569b14aba864761236bd3d5a48e4e69f10ea6c86))
* Carry `oauth2_login_challenge` over to registration flow ([#3419](https://github.com/ory/kratos/issues/3419)) ([76241be](https://github.com/ory/kratos/commit/76241bee3dc7fec4690346ee85bc4b9f897fdd34)):

    Fixes https://github.com/ory/kratos/issues/3321

* Change ListIdentities to keyset pagination ([e16fed1](https://github.com/ory/kratos/commit/e16fed1f8563509aac30886386668bb85e6dc797))
* Change shebangs and makefile from /bin/bash to /usr/bin/env bash ([#3597](https://github.com/ory/kratos/issues/3597)) ([1343bbb](https://github.com/ory/kratos/commit/1343bbbfa11ff3e7fcbc0f233b858d13fd40c66d)):

    * makefile fix
    
    
    
    * shebangs changed to /usr/bin/env bash
    
    Signed-off-by: nxy7 <lolnoxy@gmail.com>

* Check whoami aal before accepting hydra login request ([#3669](https://github.com/ory/kratos/issues/3669)) ([a2f79c3](https://github.com/ory/kratos/commit/a2f79c31f3208b88024897fc8bf1307ccac6f895))
* Code method on registration and 2fa ([#3481](https://github.com/ory/kratos/issues/3481)) ([7aa2e29](https://github.com/ory/kratos/commit/7aa2e293175d0f4b6c13552cc3781f54f8caf3a0))
* Consider OIDC registration flows errored with duplicate credential to be completed by strategy ([#3525](https://github.com/ory/kratos/issues/3525)) ([3e3c789](https://github.com/ory/kratos/commit/3e3c78967523676cbce9a227d574c2f7f4ea314d)):

    Returning anything else here may cause Kratos to respond with two concatenated JSON objects: new login flow with actual error message as the first one and a very confusing '500, aborted registration hook execution' as the second one.

* Csrf token regenerate on browser flows ([#3706](https://github.com/ory/kratos/issues/3706)) ([e4908db](https://github.com/ory/kratos/commit/e4908dbe4a42fad5a80c4d46004e1e3710cabeb7)), closes [#3705](https://github.com/ory/kratos/issues/3705)
* Data race in test ([ab6dc31](https://github.com/ory/kratos/commit/ab6dc3121535d27668fed58804a218b17b17ae43))
* Do not encode full config in multiple places ([#3500](https://github.com/ory/kratos/issues/3500)) ([57a3273](https://github.com/ory/kratos/commit/57a3273055c6e8627dd0b736e881dba3fb0fe75d))
* Do not generate CSRF token for api flows ([#3704](https://github.com/ory/kratos/issues/3704)) ([d93570d](https://github.com/ory/kratos/commit/d93570d330155c27a9315d1f530a0002a459910a))
* Do not initialize parts of the registry in parallel ([#3534](https://github.com/ory/kratos/issues/3534)) ([ff177db](https://github.com/ory/kratos/commit/ff177db8a97f27abc3e883e79832685348602334))
* Don't list org SSOs in settings ([#3637](https://github.com/ory/kratos/issues/3637)) ([6c7068c](https://github.com/ory/kratos/commit/6c7068cf41df51cde5fe9fc79cca84ec6124d38a))
* Don't require code credential for MFA flows ([#3753](https://github.com/ory/kratos/issues/3753)) ([40ed809](https://github.com/ory/kratos/commit/40ed809db631149874864f216a106c43ea5df670))
* Don't require session for OIDC verification ([#3443](https://github.com/ory/kratos/issues/3443)) ([e08f831](https://github.com/ory/kratos/commit/e08f831c2715e515bf58dc2dbb47fc3576421a5c))
* Don't return 500 on conflict for POST /admin/identities ([#3437](https://github.com/ory/kratos/issues/3437)) ([1429949](https://github.com/ory/kratos/commit/142994932e449d9948148804502c98ef73daafff))
* Don't return nil if code is invalid ([#3662](https://github.com/ory/kratos/issues/3662)) ([df8ec2b](https://github.com/ory/kratos/commit/df8ec2b9b77a53beb32e3f94a8fccb711896d8e7)):

    * fix: don't return nil if code is invalid
    
    * chore: add test

* Error handling on identity import ([#3520](https://github.com/ory/kratos/issues/3520)) ([83bfb2d](https://github.com/ory/kratos/commit/83bfb2d2a9c69bf3a3442500b9484c1a69f8c794)):

    When importing identities without any traits, or with malformed traits, 500s are returned. This improves the error handling and messaging.

* False-positives for requiring re-authentication on update ([#3421](https://github.com/ory/kratos/issues/3421)) ([ce8139f](https://github.com/ory/kratos/commit/ce8139f2325a8317388cbcaaa98f3f83d626657b))
* Http courier using should use lower case json ([#3740](https://github.com/ory/kratos/issues/3740)) ([84149c4](https://github.com/ory/kratos/commit/84149c4b420ea89f0a16a579c017a8e7e1670204))
* Identity list pagination in CLI command and SDK ([#3482](https://github.com/ory/kratos/issues/3482)) ([1e8b1ae](https://github.com/ory/kratos/commit/1e8b1aeb4bf866892788986f62a31255372de999)):

    Adds correct pagination parameters to the SDK methods for listing identities and sessions.

* Ignore CSRF middleware on Apple OIDC callback ([309c506](https://github.com/ory/kratos/commit/309c50694c11162cad070337f9b1d4e0fcdf444b))
* Ignore more cloudflare cookies ([#3499](https://github.com/ory/kratos/issues/3499)) ([f124ab5](https://github.com/ory/kratos/commit/f124ab5586781cdbfc0a0cfd11b4355bfc8a115c))
* Improved SSRF protection ([#3629](https://github.com/ory/kratos/issues/3629)) ([6d08576](https://github.com/ory/kratos/commit/6d08576bbc2c06014192f05e0129b95eb6c9fd80)):

    This also improves tracing in the OIDC strategy.

* Incorrect login accept challenge ([#3658](https://github.com/ory/kratos/issues/3658)) ([b5dede3](https://github.com/ory/kratos/commit/b5dede329247d0962688b15872a6caf027cf910f))
* Incorrect sdk generator path ([#3488](https://github.com/ory/kratos/issues/3488)) ([ed996c0](https://github.com/ory/kratos/commit/ed996c0d25e68e8a2c7de861c546f0b0e42e9e6e))
* Incorrect SMTP error handling ([#3636](https://github.com/ory/kratos/issues/3636)) ([ee138ec](https://github.com/ory/kratos/commit/ee138ec4e1ba55ef077858653220db9e6b0c7254))
* Incorrect swagger spec for filter parameter ([#3684](https://github.com/ory/kratos/issues/3684)) ([2c1470a](https://github.com/ory/kratos/commit/2c1470ab3556e639f06a01ac1646a6b90c7ecac7)), closes [#3676](https://github.com/ory/kratos/issues/3676) [#3675](https://github.com/ory/kratos/issues/3675)
* Increase connection-level timeouts and shutdown timeouts ([#3570](https://github.com/ory/kratos/issues/3570)) ([200b413](https://github.com/ory/kratos/commit/200b4138a429d113ee045d16031bb0a6312c1c01)):

    The admin API is generally expected to require longer timeouts, for example during bulk identity import.

* Issue session after verification after registration with OIDC SSO ([#3467](https://github.com/ory/kratos/issues/3467)) ([a28b523](https://github.com/ory/kratos/commit/a28b523238743f3873b51479eea3b86d684092f9))
* Lint ([e8740c3](https://github.com/ory/kratos/commit/e8740c3498446dcaeab2990604a317e61dc170df))
* Lower-case recovery & verification emails on import ([#3571](https://github.com/ory/kratos/issues/3571)) ([e2ac9ff](https://github.com/ory/kratos/commit/e2ac9ff4e2101788f1fca1b8c83f8791cce446e2)):

    Emails that contained upper-case characters would be overwritten by the identity schema extension runner, because there all emails are lower-cased.

* Mark identity as optional in session struct ([#3463](https://github.com/ory/kratos/issues/3463)) ([7ae02ba](https://github.com/ory/kratos/commit/7ae02ba697f68c9cfae5fe8f696b2c55a3ba9ddc)), closes [#3461](https://github.com/ory/kratos/issues/3461):

    The identity is not always available in the session struct, for example when AAL2 is required.

* Omit irrelevant OIDC providers in forced refresh login flows ([#3608](https://github.com/ory/kratos/issues/3608)) ([912dccd](https://github.com/ory/kratos/commit/912dccdf04a550604c5bfeb53ccf79c5f1133ef2)):

    Whenever an user is asked to reauthenticate (e.g. because they wish to execute settings flow touching their credentials and their session is no longer privileged) they are asked to provide their credentials again. The forced-refresh login flow generated for such cases already excludes some strategies that are enabled in Kratos but cannot be used to authenticate as current identity, and for example the form presented to the user will not have a password field if the identity does not have a password credential.
    
    This, however, does not currently apply to OIDC providers; the user will always see the full set even if some of them can't be used to sign in as current identity. This change causes forced refresh login flows to also omit irrelevant OIDC providers in generated form in order to avoid confunding the user about which strategies/providers are valid and can actually be used to reauthenticate.

* On verification required after registration, preserve return_to ([#3589](https://github.com/ory/kratos/issues/3589)) ([6a0a914](https://github.com/ory/kratos/commit/6a0a9149b9828ba994bec9b48a43f9d70245f43f)):

    * fix: on verification required after registration, preserve return_to
    
    * test: return_to on verification flow
    
    * chore: refactor
    
    

* Panic in recovery ([#3639](https://github.com/ory/kratos/issues/3639)) ([c25ddff](https://github.com/ory/kratos/commit/c25ddffd2270a8d0861e2fc78cd0ba26e63af4eb))
* Pass context ([#3452](https://github.com/ory/kratos/issues/3452)) ([c492bdc](https://github.com/ory/kratos/commit/c492bdcd0c5dbdf527ae523d879a6c1eeb9c4cdf))
* Properly normalize OIDC verified emails ([#3450](https://github.com/ory/kratos/issues/3450)) ([703b910](https://github.com/ory/kratos/commit/703b910927d879558bfeb0fd2c3339b1d301fac8))
* Redirect to verification URL even if login_challenge is set ([#3412](https://github.com/ory/kratos/issues/3412)) ([cd9e6a0](https://github.com/ory/kratos/commit/cd9e6a0e1e4cb4957d2a50ae3d288ebb0591e42d)):

    Fixes https://github.com/ory/network/issues/320

* Reduce db lookups in whoami for aal check ([#3372](https://github.com/ory/kratos/issues/3372)) ([d814a48](https://github.com/ory/kratos/commit/d814a4864d5c25c4f320daca733873577d517331)):

    Significantly improves performance by reducing the amount of queries we need to do when checking for the different AAL levels.

* Registration code ui nodes group ([#3505](https://github.com/ory/kratos/issues/3505)) ([6220184](https://github.com/ory/kratos/commit/622018459ddb16c182da49dfd91fd1c6ef8c6b73)):

    * fix: registration code ui nodes group
    
    * style: format

* Registration should accept hydra login ([#3592](https://github.com/ory/kratos/issues/3592)) ([7a47827](https://github.com/ory/kratos/commit/7a47827cfd58ef68ebfbbeaf5ed86c394ba2bd5e)):

    * fix: registration should accept hydra login
    
    * fix: oauth2 registration flow with session
    
    * wip: registration oauth flow tests
    
    * wip: refactor oauth flows test
    
    * wip: refactor op_registration_test
    
    * wip: oauth provider registration test
    
    * wip: refactor oauth flows test
    
    * fix(test): oauth provider login
    
    * style: format

* Registration with verification ([#3451](https://github.com/ory/kratos/issues/3451)) ([77c3196](https://github.com/ory/kratos/commit/77c3196fd60c5927b84e9a7f6546f80ac2d78ee5))
* Reject obviously invalid email addresses from courier ([8cb9e4c](https://github.com/ory/kratos/commit/8cb9e4cae9dffd4c25d52920186f9c5fbe2bd0fe))
* Remove `earliest_possible_extend` default in schema ([#3464](https://github.com/ory/kratos/issues/3464)) ([7e05b7d](https://github.com/ory/kratos/commit/7e05b7db3c01efc96185ac18042e971e33da37c8))
* Remove duplicate message ID usage ([#3468](https://github.com/ory/kratos/issues/3468)) ([dfcbe22](https://github.com/ory/kratos/commit/dfcbe226bc53b91f3a6c9837496a159b85c2e68a))
* Remove requirement for smtp section ([#3405](https://github.com/ory/kratos/issues/3405)) ([59a3f14](https://github.com/ory/kratos/commit/59a3f1469b8412e49846a500493cb02fc6eb34b1))
* Remove slow queries from update identities ([#3553](https://github.com/ory/kratos/issues/3553)) ([d138abb](https://github.com/ory/kratos/commit/d138abb6278ebb232e120bee0fb956a0f2816b8d))
* Rename "phone" courier channel to "sms" ([#3680](https://github.com/ory/kratos/issues/3680)) ([eb8d1b9](https://github.com/ory/kratos/commit/eb8d1b9abd6d2b3eb86ab11d48d9ebd059586b67))
* Respect gomail.SendError in mail queue ([#3600](https://github.com/ory/kratos/issues/3600)) ([9c608b9](https://github.com/ory/kratos/commit/9c608b991874d839782d9219f2fc27d0d4a398af))
* Respond with 422 when SPA identity requires AAL2 ([#3572](https://github.com/ory/kratos/issues/3572)) ([df18c09](https://github.com/ory/kratos/commit/df18c09e0089743e8aee17540d277b9572252e06)):

    If you submit a browser login flow with an `Accept` header of `application/json`, but the login flow requires AAL2, then there is no way for the code to know it needs to redirect the user to the 2FA page. Instead of responding with the `Session` in this scenario, this PR changes the behaviour to respond with a `browser_location_change_required` error (status `422`) to indicate that the browser needs to open a specific URL, /self-service/login/browser?aal=aal2. 
    
    

* Return 400 bad request for invalid login challenge ([#3404](https://github.com/ory/kratos/issues/3404)) ([ca34e9b](https://github.com/ory/kratos/commit/ca34e9b744482b41d65082f3bed52e9c4ebd7ba4))
* Return HTTP 400 if key unmarshal fails ([#3594](https://github.com/ory/kratos/issues/3594)) ([fdf4956](https://github.com/ory/kratos/commit/fdf4956d9218cfa1d2227c4880e48f9bbdaeb95d)):

    * fix: return HTTP 400 if key unmarshal fails
    
    * fix: apply reviewer's suggestion, prepare for  bump
    
    * fix: follow up reviewer suggestion from ory/x
    
    * chore: bump ory/x

* Schema test errors ([#3528](https://github.com/ory/kratos/issues/3528)) ([bee0341](https://github.com/ory/kratos/commit/bee0341c5bf5708a2210146fc59f050a1b9df663))
* Set iss from userinfo claims if missing ([#3744](https://github.com/ory/kratos/issues/3744)) ([241a911](https://github.com/ory/kratos/commit/241a911af74e8ad7353d6e3cab86db20758b86fc))
* Specify correct minimum versions in migratest ([18b89ea](https://github.com/ory/kratos/commit/18b89ea588d129fa88379f7b0d7f4fd00ec6023d))
* Tracing context passing in /sessions/whoami ([1254bf5](https://github.com/ory/kratos/commit/1254bf5a38dbe2c0e2798e07dd0ee5e4b2f63d6e))
* Tracing improvements ([c804cb2](https://github.com/ory/kratos/commit/c804cb2bebbefc97073cf3b8fa250c3eefc58894))
* Type-assert all interfaces that WebHook implements ([ffda1a0](https://github.com/ory/kratos/commit/ffda1a0dab661c5f11ad849b9287094313561b79))
* Ui node input attributes key added ([#3561](https://github.com/ory/kratos/issues/3561)) ([9eff0f3](https://github.com/ory/kratos/commit/9eff0f3a611f32af7aa7f27587b3d3f4448ce915)):

    * fix: ui node InputAttributes.Key added
    
    * fix: selfservice recovery flow add React unique key and numeric pattern
    
    * fix: remove React related key addition
    
    * test: update snapshot

* Use ID label on login with multiple identifiers ([#3657](https://github.com/ory/kratos/issues/3657)) ([be907db](https://github.com/ory/kratos/commit/be907dbbd841025fd854344b77d3368b2ff8089f))
* Use org ID from session if available in login flow ([#3545](https://github.com/ory/kratos/issues/3545)) ([1b3647c](https://github.com/ory/kratos/commit/1b3647c2acdad966f920c2b9e6e657c52aa50c6e))
* Use provider label in link message ([#3661](https://github.com/ory/kratos/issues/3661)) ([fa5ec93](https://github.com/ory/kratos/commit/fa5ec93e8ae7d971d07f0e9b3acaa0840b9ac7de))
* Use registry client for schema loading ([#3471](https://github.com/ory/kratos/issues/3471)) ([3a57726](https://github.com/ory/kratos/commit/3a577269980213e4415fd5fa713882990e2e7640))
* Using first name as last name ([#3556](https://github.com/ory/kratos/issues/3556)) ([df80377](https://github.com/ory/kratos/commit/df80377f5fe6180fba5904baa5be1ba1d68eb2aa))
* Wrong continue_with enum declaration ([#3522](https://github.com/ory/kratos/issues/3522)) ([4c34c24](https://github.com/ory/kratos/commit/4c34c2417db0cb1f79b42db5f33544c90b38ad87))

### Code Generation

* Pin v1.1.0 release commit ([f47675b](https://github.com/ory/kratos/commit/f47675b82012e0ff74b05b9b7e713b3aa2fdda54))

### Documentation

* Add example for `allowed_return_urls` to include wildcard url ([#3533](https://github.com/ory/kratos/issues/3533)) ([39b0c3c](https://github.com/ory/kratos/commit/39b0c3c03df0aec254b32c840730452d4856872b)), closes [#1528](https://github.com/ory/kratos/issues/1528)
* Improve enum handling and completeness ([#3714](https://github.com/ory/kratos/issues/3714)) ([4b881ca](https://github.com/ory/kratos/commit/4b881cae4359bfa068261d2d0765ce3daadcbcf2))
* Remove experimental warnings ([#3406](https://github.com/ory/kratos/issues/3406)) ([d4d26e6](https://github.com/ory/kratos/commit/d4d26e6e1510c8e09346e95251f420f95ec54998)):

    See https://github.com/ory/kratos/discussions/3388

* Update link to hashed password formats ([#3484](https://github.com/ory/kratos/issues/3484)) ([8ca3adc](https://github.com/ory/kratos/commit/8ca3adcb8a5db2906fbeb92f4b74aa4242fabdef))

### Features

* Add ability to convert session to JWT when calling whoami ([#3472](https://github.com/ory/kratos/issues/3472)) ([57b7bb8](https://github.com/ory/kratos/commit/57b7bb846c8072f786ea6b80cd688fdee75805da)), closes [#2487](https://github.com/ory/kratos/issues/2487):

    This patch adds a query parameter `tokenize_as` to `/session/whoami` which encodes the session to a JWT. It is possible to customize the JWT claims by using a JsonNet template, and furthermore change the expiry of the token.
    
    The tokenize feature supports multiple templates, which makes it easy to use the resulting JWT in a variety of use cases.

* Add event ([#3524](https://github.com/ory/kratos/issues/3524)) ([75031e6](https://github.com/ory/kratos/commit/75031e67bc82a820a6aba134115e8d5f93303638))
* Add GetID member functions to RecoveryAddress and Credentials ([#3474](https://github.com/ory/kratos/issues/3474)) ([085d500](https://github.com/ory/kratos/commit/085d5002df27d455057d33bd2d93dfbca0de4872))
* Add ID Token sign in with Google Android/iOS SDK ([#3515](https://github.com/ory/kratos/issues/3515)) ([055ed92](https://github.com/ory/kratos/commit/055ed9226d9d12f5142542be2e18438ff708c2e2))
* Add OpenTelemetry span for password hash comparison ([#3383](https://github.com/ory/kratos/issues/3383)) ([e3fcf0c](https://github.com/ory/kratos/commit/e3fcf0c31db9742ed61bcf783e37ee119ed19d42))
* Add request URL to email and SMS templates ([bf5f8c3](https://github.com/ory/kratos/commit/bf5f8c3cfb2eb523a77239addb8249adf9f8b31d))
* Add sms verification for phone numbers ([#3649](https://github.com/ory/kratos/issues/3649)) ([e3a3c4f](https://github.com/ory/kratos/commit/e3a3c4fe0d6697f6864283daf4be8a8f8971c7b4))
* Add support for recovery on native flows ([#3273](https://github.com/ory/kratos/issues/3273)) ([e363889](https://github.com/ory/kratos/commit/e363889732c0a1cb801fd12b2e0e8546006e9714))
* Add WebhookSucceeded event ([aa8c936](https://github.com/ory/kratos/commit/aa8c93677a8f682f7693afe69f1baf1887355e0a))
* Added various new text messages ([ea91483](https://github.com/ory/kratos/commit/ea914834e6bb626de2977e228af2b40935ccc980)):

    To improve i18n and message customization, we added a bunch of new messages. Integrations that do message customization should probably handle those new message codes:
    
    - 1010014
    - 1010015
    - 1040005
    - 1040006
    - 1070012
    - 1070013
    - 4000028
    - 4000029
    - 4000030
    - 4000031
    - 4000032
    - 4000033
    - 4000034
    - 4000035
    - 4000036
    - 4010007
    - 4010008
    - 4040002
    - 4040003
    
    Additionally, these messages got more context:
    
    - 1050014
    - 1050018
    - 1070002
    - 4000001
    - 4000003
    - 4000004
    - 4000017
    - 4000018
    - 4000019
    - 4000020
    - 4000021
    - 4000022
    - 4000023
    - 4000024
    - 4000025
    - 4000026
    - 4010001
    - 4040001
    - 4050001
    - 4060005
    - 4070005
    - 5000001

* Allow additional id token audiences ([#3616](https://github.com/ory/kratos/issues/3616)) ([0fa648d](https://github.com/ory/kratos/commit/0fa648d9f7b837a35de9b230a05b5951e95d5874))
* Allow extra migrations in NewPersister ([96c1ff7](https://github.com/ory/kratos/commit/96c1ff7747ea38e23a3892f74b75ee555ed49c88))
* Allow fuzzy-search on credential identifiers ([#3526](https://github.com/ory/kratos/issues/3526)) ([2cb3ea2](https://github.com/ory/kratos/commit/2cb3ea2eaff909ac936611d5653f69e713f41b64)):

    This PR adds the ability to search for sub-strings and similar strings in credential identifiers.
    
    Note that the **postgres** and **CRDB** migrations create special indexes useful for this feature. To use [online schema changes](https://www.cockroachlabs.com/docs/v23.1/online-schema-changes) with cockroach, we recommend to manually copy the index definition and run it before applying migrations. The migration will then be a no-op.
    
    If you run on **mysql** (or **sqlite**), no special index is created. If desired, you can create such an index manually, and it would be highly appreciated if you could contribute its definition.
    
    This feature is a preview and will change in behavior! Similarity search is not expected to return deterministic results but are useful for humans.

* Allow importing hmac hashed passwords ([#3544](https://github.com/ory/kratos/issues/3544)) ([0a0e1f7](https://github.com/ory/kratos/commit/0a0e1f7200e226ef24de062811a05bcdd02b6acd)), closes [#2422](https://github.com/ory/kratos/issues/2422):

    The basic format is `$hmac-<hashfunction>$<base64 encoded hash>$<base64 encoded key>`:
    
    ```
    # password = test; key=key; hash function=sha
    $hmac-sha1$NjcxZjU0Y2UwYzU0MGY3OGZmZTFlMjZkY2Y5YzJhMDQ3YWVhNGZkYQ==$a2V5
    ```

* Allow marking OIDC provider-verified addresses as verified during registration ([#3448](https://github.com/ory/kratos/issues/3448)) ([e7b33a1](https://github.com/ory/kratos/commit/e7b33a168bf0c0fe0492901abd3df8b6d6a08a68)), closes [#3445](https://github.com/ory/kratos/issues/3445) [#3424](https://github.com/ory/kratos/issues/3424) [#1057](https://github.com/ory/kratos/issues/1057):

    This feature allows marking emails provided by social sign in providers as verified.

* Batch list identities ([#3598](https://github.com/ory/kratos/issues/3598)) ([8ad54f1](https://github.com/ory/kratos/commit/8ad54f1be53b30fdb24b616be0c52fd66829f201)), closes [#2448](https://github.com/ory/kratos/issues/2448):

    This change allows to filter `GET /admin/identities` by ID with the following syntax:
    
    ```
    /admin/identities?ids=id1&ids=id2&ids=id3
    ```

* **changelog:** Add support for native recovery ([#3624](https://github.com/ory/kratos/issues/3624)) ([492808c](https://github.com/ory/kratos/commit/492808cae0e804793aef9a02a902fce988f9fc6d)):

    Adds the ability to complete the recovery flow properly on API flows. This PR also streamlines the behavior for SPA flows to not return 422 errors anymore. To enable this new behavior, set the features.use_continue_with_transitions flag in the config to `true`.
    
    See also https://github.com/ory/kratos/pull/3273

* Claims from userinfo endpoint ([#3718](https://github.com/ory/kratos/issues/3718)) ([90bdc61](https://github.com/ory/kratos/commit/90bdc61d28466f10e4e609df014b220afbee0478)):

    * feat: claims from userinfo endpoint
    
    * chore: update libraries
    
    * test: improve coverage

* Emit error details when we find stray cookies in an API flow ([#3496](https://github.com/ory/kratos/issues/3496)) ([df74339](https://github.com/ory/kratos/commit/df74339802d98a292abb32806eca35fb2554960b))
* Eventually consistency API controls ([#3558](https://github.com/ory/kratos/issues/3558)) ([00cf11c](https://github.com/ory/kratos/commit/00cf11c071344103c603c078f07196401d091780)):

    Adds a feature used in Ory Network which enables trading faster reads for slightly stale data.
    
    This feature depends on Cockroach functionality and configuration, and is not possible for MySQL or PostgreSQL.

* Extend Microsoft Graph API capabilities ([#3609](https://github.com/ory/kratos/issues/3609)) ([4a7bcc9](https://github.com/ory/kratos/commit/4a7bcc9322be37e6fd141e411bd65e3977eeb692)):

    This change queries for all user information available with the `User.Read` scope
    during OIDC, and populates the `RawClaims` field.

* Extract identifier label for login from default identity schema ([#3645](https://github.com/ory/kratos/issues/3645)) ([180828e](https://github.com/ory/kratos/commit/180828eb507ab239a9c6589f747a6816b6e50074))
* Fine-grained hooks for all available flow methods ([#3519](https://github.com/ory/kratos/issues/3519)) ([a37f6bd](https://github.com/ory/kratos/commit/a37f6bddc48443b2fc464699fa5c2922f64d81f6)):

    Adds fine-grained hook configurations to the post-settings flow for methods totp, webauthn, lookup_secret and the post-login flow for totp, lookup_secret, and code.

* Hook to revoke sessions after password changed ([#3514](https://github.com/ory/kratos/issues/3514)) ([e6af6db](https://github.com/ory/kratos/commit/e6af6db37ff5de33a656ce7804c813451395459d)), closes [#3513](https://github.com/ory/kratos/issues/3513):

    Currently, the Kratos system does not automatically log out or invalidate other active sessions when a user changes their password. This poses a significant security risk as it allows potentially unauthorized individuals to maintain access to the account even after the password has been updated.
    
    This PR provides the option to add the `revoke_active_sessions` hook to the actions sections of the selfservice settings.

* Hot-reload CORS origins ([#3423](https://github.com/ory/kratos/issues/3423)) ([157d934](https://github.com/ory/kratos/commit/157d9345aeb04f371f9d85b70c89e8646e781333))
* Improve messages for easier i18n ([#3457](https://github.com/ory/kratos/issues/3457)) ([37f1657](https://github.com/ory/kratos/commit/37f16577d92ba88869bf15fb1ea54e819b062724))
* Improve performance by computing password hashes while validating ([#3508](https://github.com/ory/kratos/issues/3508)) ([a9786c5](https://github.com/ory/kratos/commit/a9786c599d09f61e2e07df5066ce94feb2d99bac))
* Improved webhook tracing ([#3746](https://github.com/ory/kratos/issues/3746)) ([9d7021d](https://github.com/ory/kratos/commit/9d7021d87f47690c2c1f8000e87b425e49bc9496))
* Jsonnet caching for OIDC claims mapper, webhooks, JWT session tokenizer ([#3701](https://github.com/ory/kratos/issues/3701)) ([1d26e09](https://github.com/ory/kratos/commit/1d26e097b273aeda36f73637765da5bdb2aa4a66))
* Link oidc credentials when login ([#3563](https://github.com/ory/kratos/issues/3563)) ([b784949](https://github.com/ory/kratos/commit/b784949d03b849d9d1d594977f75f5843b7b5da8)), closes [#2727](https://github.com/ory/kratos/issues/2727) [#3222](https://github.com/ory/kratos/issues/3222):

    When user tries to login with OIDC for the first time but has already registered before with email/password a credentials identifier conflict may be detected by Kratos. In this case user needs to login with email/password first and then link OIDC credentials on a settings screen.
    This PR simplifies UX and allows user to link OIDC credentials to existing account right in the login flow, without
    switching to settings flow.

* List by OIDC cred ([#3721](https://github.com/ory/kratos/issues/3721)) ([bff9c61](https://github.com/ory/kratos/commit/bff9c61b147648ab139e7e86cda4336b5d1cfd39))
* Login with code on any credential type ([#3549](https://github.com/ory/kratos/issues/3549)) ([ceed7d5](https://github.com/ory/kratos/commit/ceed7d5478c5cca894587698c57f676dda100b27)):

    Should be able to login with the `code` credential even if the user did not register on the `code` credential. 
    Only `identifier` matching is done and validation based on the identity schema.

* One-time code native flows ([#3516](https://github.com/ory/kratos/issues/3516)) ([9b0fee3](https://github.com/ory/kratos/commit/9b0fee30f980d860fd548e7589fa6a06e593537a))
* Order sessions by created_at ([#3696](https://github.com/ory/kratos/issues/3696)) ([688111c](https://github.com/ory/kratos/commit/688111c9a6bf9872657cf6aada77f55fa2520e00))
* Parametrize courier worker ([#3601](https://github.com/ory/kratos/issues/3601)) ([0e4be57](https://github.com/ory/kratos/commit/0e4be57e41e1152f4be22f490541c2c099cfe3fe)):

    Allows one to parametrize how many messages the courier will fetch and how often it will fetch messages.

* Passwordless browser login and registration via code to email ([#3378](https://github.com/ory/kratos/issues/3378)) ([eaaf375](https://github.com/ory/kratos/commit/eaaf37519917612671238412a633847386d7c613)), closes [#2029](https://github.com/ory/kratos/issues/2029) [ory-corp/cloud#3573](https://github.com/ory-corp/cloud/issues/3573):

    This feature adds passwordless email code login. When a user signs up, or signs in, a code is sent to their email address which they can use to complete the authentication process.
    
    This feature is currently only working for browser facing APIs.

* Pooled process-isolated Jsonnet VM ([9a52ddf](https://github.com/ory/kratos/commit/9a52ddfbe7c24c41b6aa3ddc3c79c6fcbfb8db02))
* Provide login hints when registration fails due to duplicate credentials/addresses ([#3430](https://github.com/ory/kratos/issues/3430)) ([8b28469](https://github.com/ory/kratos/commit/8b284697e4a26fb01ad57d2e9ebd8f714be49f33)):

    * feat: provide login hints when registration fails due to duplicate credentials or identifiers
    
    * feat: identify edge cases and write tests
    
    * chore: synchronize workspaces
    
    * feat: make login hints configurable
    
    * chore: synchronize workspaces
    
    * chore: synchronize workspaces
    
    * chore: synchronize workspaces
    
    * chore: synchronize workspaces

* Support auth_type parameter ([#3487](https://github.com/ory/kratos/issues/3487)) ([fc30304](https://github.com/ory/kratos/commit/fc303040b71139f512fd1491ce30f80837b940b9)):

    The Facebook OIDC provider supports an auth_type parameter that
    when set to "reauthenticate" will force the user to
    reauthenticate (similar to `prompt=login` for other Providers).

* Support for B2B SSO ([#3489](https://github.com/ory/kratos/issues/3489)) ([0ec037a](https://github.com/ory/kratos/commit/0ec037ab298ed28fb0ac84db6a4d2b14b81e57df))
* Support MFA via SMS ([#3682](https://github.com/ory/kratos/issues/3682)) ([1516cf6](https://github.com/ory/kratos/commit/1516cf64e346819dccace1cc25aaccac38b9e47c))
* Support multiple origins for WebAuthN ([#3380](https://github.com/ory/kratos/issues/3380)) ([013f335](https://github.com/ory/kratos/commit/013f335881831bbf90ac31b219b57118fc089fe6)):

    Users can now supply a list of origins for webauthn in the configuration.

* Support native social sign using apple sdk ([#3476](https://github.com/ory/kratos/issues/3476)) ([f561013](https://github.com/ory/kratos/commit/f561013dd737dadcc82c4ec049fde12861e91e43))
* Transmit current session ID to Hydra when accepting the login ([#3426](https://github.com/ory/kratos/issues/3426)) ([610c76d](https://github.com/ory/kratos/commit/610c76d9140f2f43217ac55094051a994ea83ecc)):

    * chore: change react-native port to 19006
    
    * feat: transmit current session ID when accepting login
    
    * fix: upgrade hydra in tests

* Webhook analytic events ([9c8a25e](https://github.com/ory/kratos/commit/9c8a25eb0d3e06df182565d3d959d57e5dccfed8))

### Reverts

* Revert "chore: simplify courier code (#3603)" ([7c54c9f](https://github.com/ory/kratos/commit/7c54c9f36c86142c8e071a5359c71cf6213a1a69)), closes [#3603](https://github.com/ory/kratos/issues/3603):

    This reverts commit 316cd4aacfe31efafa7d737a7c476e2c794e9c9b.


### Tests

* Add test for link + oidc challenge ([#3720](https://github.com/ory/kratos/issues/3720)) ([67360cf](https://github.com/ory/kratos/commit/67360cf39482b935604f088a4b7a83cc4deab375))
* **e2e:** Logout return_to ([#3418](https://github.com/ory/kratos/issues/3418)) ([c348c12](https://github.com/ory/kratos/commit/c348c12ab3c9cdb4ce8159fe774ed179ff6a4d8a))
* Fix cypress setup ([#3527](https://github.com/ory/kratos/issues/3527)) ([70c8ddd](https://github.com/ory/kratos/commit/70c8ddd49c8abb9c10f2ca349e01061b791c5e7b))
* Fix e2e failures and speed up e2e tests ([#3483](https://github.com/ory/kratos/issues/3483)) ([70a6171](https://github.com/ory/kratos/commit/70a617194d61763f4b75691b22cfa76ba71ab019))
* Fix hydra tests on master ([#3737](https://github.com/ory/kratos/issues/3737)) ([12166b4](https://github.com/ory/kratos/commit/12166b4370d607a069f268227752bb7b18a50b57))
* Reduce logging in go tests ([#3562](https://github.com/ory/kratos/issues/3562)) ([05de3a2](https://github.com/ory/kratos/commit/05de3a29fed020593c44ea7a7b29e45197fef4f7))
* Resolve cypress issues ([#3531](https://github.com/ory/kratos/issues/3531)) ([4206d26](https://github.com/ory/kratos/commit/4206d2605dfa30b19e132be31b85b1a35f8dca78))

### Unclassified

* Revert "feat: extend Microsoft Graph API capabilities (#3609)" (#3717) ([549308d](https://github.com/ory/kratos/commit/549308db1f7dca42004631ed6156cae5f827b8fe)), closes [#3609](https://github.com/ory/kratos/issues/3609) [#3717](https://github.com/ory/kratos/issues/3717):

    This reverts commit 4a7bcc9322be37e6fd141e411bd65e3977eeb692.
    
    



# [1.0.0](https://github.com/ory/kratos/compare/v0.13.0...v1.0.0) (2023-07-12)

We are thrilled to announce Ory Kratos v1.0, the powerful Identity, User Management, and Authentication system! With this major update, Ory Kratos brings a host of enhancements and fixes that greatly improve the user experience and overall performance.

Several compelling reasons led to label Ory Kratos as a major release, like successfully processing over 100 million API requests daily and having about 100 million Docker Pulls. We have maintained stability within the Ory Kratos APIs for nearly two years, demonstrating their robustness and reliability. No breaking changes mean that developers can trust the stability of Ory Kratos in production.

Ory Kratos 1.0 introduces a variety of new features while focusing on stability, robustness, and improved performance. Major enhancements include support for social login and single-sign-on via OpenID connect in native apps, emails sent through HTTP rather than SMTP, and full compatibility with Ory Hydra v2.2.0. Users will also find multi-region support in the Ory Network for broader geographic reach, improved export functionality for all credential types, and enhanced session management with the introduction of the "provider ID" parameter. Other additions comprise distroless images for leaner resource utilization and faster deployment and support for the Lark OIDC provider.

Significant improvements and fixes accompany these new features. Enhanced OIDC flows now include the ability to forward prompt upstream parameters, offering developers increased flexibility and customization options. The logout flow also supports the `return_to` parameter, facilitating more flexible redirection post-user logout. Performance has been a key focus, with Ory Kratos 1.0 now capable of handling hundreds of millions of active users monthly. Critical bug fixes have been applied to prevent users from being redirected to incorrect destinations, ensuring smoother authentication and authorization. Additionally, there's more support for legacy systems via implemented crypt(3) hashers and a fix for metadata patching has been deployed to ensure consistent user metadata management. For a detailed view of all changes, refer to the [changelog on GitHub]( https://github.com/ory/kratos/blob/master/CHANGELOG.md). Feedback and support are, as always, greatly appreciated.

Ory Kratos 1.0 is a major release that marks a significant milestone in our journey.

We sincerely hope that you find these new features and improvements in Ory Kratos 1.0 valuable for your projects. To experience the power of the latest release, we encourage you to get the latest version of Ory Kratos [here](https://github.com/ory/kratos) or leverage Kratos in [Ory Network](https://www.ory.sh/network/) — the easiest, simplest, and most cost-effective way to run Ory.

For organizations seeking to upgrade their self-hosted solution, **Ory offers dedicated support services to ensure a smooth transition**. Our team is ready to assist you throughout the migration process, ensuring uninterrupted access to the latest features and improvements. Additionally, we provide various [support plans](https://www.ory.sh/support/) specifically tailored for self-hosting organizations. These plans offer comprehensive assistance and guidance to optimize your Ory deployments and meet your unique requirements.

We extend our heartfelt gratitude to the vibrant and supportive Ory Community. Without your constant support, feedback, and contributions, reaching this significant milestone would not have been possible. As we continue on this journey, your feedback and suggestions are invaluable to us. Together, we are shaping the future of identity management and authentication in the digital landscape.

Contributors to this release in alphabetical order: [borisroman](https://github.com/ory/kratos/commits?author=borisroman), [ci42](https://github.com/ory/kratos/commits?author=ci42), [CNLHC](https://github.com/ory/kratos/commits?author=CNLHC), [David-Wobrock](https://github.com/ory/kratos/commits?author=David-Wobrock), [giautm](https://github.com/ory/kratos/commits?author=giautm), [IchordeDionysos](https://github.com/ory/kratos/commits?author=IchordeDionysos), [indietyp](https://github.com/ory/kratos/commits?author=indietyp), [jossbnd](https://github.com/ory/kratos/commits?author=jossbnd), [kralicky](https://github.com/ory/kratos/commits?author=kralicky), [PhakornKiong](https://github.com/ory/kratos/commits?author=PhakornKiong), [sunakan](https://github.com/ory/kratos/commits?author=sunakan), [steverusso](https://github.com/ory/kratos/commits?author=steverusso)

Are you passionate about security and want to make a meaningful impact in one of the biggest open-source communities? Join the [Ory community](https://slack.ory.sh) and become a part of the new ID stack. Together, we are building the next generation of IAM solutions that empower organizations and individuals to secure their identities effectively.

Want to check out Ory Kratos yourself? Use these commands to get your Ory Kratos project running on the Ory Network:

```shell
brew install ory/tap/cli

scoop bucket add ory https://github.com/ory/scoop.git
scoop install ory

bash <(curl <https://raw.githubusercontent.com/ory/meta/master/install.sh>) -b . ory
sudo mv ./ory /usr/local/bin/

ory auth

ory create project --name "My first Kratos project"

ory open account-experience registration

ory patch identity-config \\
  --replace '/identity/default_schema_id="preset://username"' \\
  --replace '/identity/schemas=[{"id":"preset://username","url":"preset://username"}]' \\
  --format yaml

ory open account-experience registration
```





### Bug Fixes

* Ability to patch metadata even if it is `null` ([#3304](https://github.com/ory/kratos/issues/3304)) ([3c04d8f](https://github.com/ory/kratos/commit/3c04d8fb63cacf91774864450b02d6d1eb90d856))
* Accept OIDC login request in browser+JSON login flow ([#3271](https://github.com/ory/kratos/issues/3271)) ([ad54093](https://github.com/ory/kratos/commit/ad540930df96e84fb65a36616d5081ec0bb46df5)):

    * fix: OIDC login in browser JSON flow
    
    * test: add test for OIDC+JSON continuity cookie

* Add error checking when creating verification code ([#3328](https://github.com/ory/kratos/issues/3328)) ([7182eca](https://github.com/ory/kratos/commit/7182eca074c8e84be325d62c75b62d22698878be))
* Add missing SessionIssued event for api flows ([#3348](https://github.com/ory/kratos/issues/3348)) ([adf78e0](https://github.com/ory/kratos/commit/adf78e09f336b2ac83f8ff1ba5ca382c7cfbec23)):

    * fix: missing SessionIssued event for api flows
    * chore: add SessionIssued event to post registration hook
    * chore: format
    * chore: move sessionissued event to persister

* Bump quickstart version ([#3257](https://github.com/ory/kratos/issues/3257)) ([6db70a8](https://github.com/ory/kratos/commit/6db70a81afac5860a86c31881a6fc988096ff0e4))
* Cypress TOTP test ([eac908c](https://github.com/ory/kratos/commit/eac908c4fc14831288e6fd5b3c65ac197d2f58e1))
* Do not require items to be unique ([#3349](https://github.com/ory/kratos/issues/3349)) ([17be30d](https://github.com/ory/kratos/commit/17be30dd84c667e5d1ae13bd79827b7ca9cdd2de))
* Don't assume the login challenge to be a UUID ([#3317](https://github.com/ory/kratos/issues/3317)) ([3172862](https://github.com/ory/kratos/commit/3172862929ad68011fc940a6e0876fa07187a275)):

    For compatibility with https://github.com/ory/hydra/pull/3515, which
    now encodes the whole flow in the login challenge, we cannot further
    assume that the challenge is a UUID.

* **e2e:** Install kratos-selfservice-ui-node peer deps ([#3354](https://github.com/ory/kratos/issues/3354)) ([ce20063](https://github.com/ory/kratos/commit/ce20063a858acecb5d9124792fe6d3899bf95c1c))
* Identity list pagination ([#3325](https://github.com/ory/kratos/issues/3325)) ([9d3ef0d](https://github.com/ory/kratos/commit/9d3ef0df9333aff2c587005df0cdd263028029f3)):

    Resolves a pesky issue that would skip the last page.

* IdentityCreated event ([#3314](https://github.com/ory/kratos/issues/3314)) ([78e31cb](https://github.com/ory/kratos/commit/78e31cb82a28e240a6176c8d3d9ef3bc64559e75))
* Incorrect override in identity hydrate ([#3368](https://github.com/ory/kratos/issues/3368)) ([eaa3f3c](https://github.com/ory/kratos/commit/eaa3f3c19feaf9048e800cc5a5f1e28d3708c624))
* Increase size for request url ([#3366](https://github.com/ory/kratos/issues/3366)) ([10713cc](https://github.com/ory/kratos/commit/10713cc703457cb6f4a1b38482c836e54a0cb224))
* Minor refactorings in package hash ([#3186](https://github.com/ory/kratos/issues/3186)) ([831fb19](https://github.com/ory/kratos/commit/831fb19e1c98b9fade3ff61d26ad249c548292d6))
* Missing id for login event ([#3315](https://github.com/ory/kratos/issues/3315)) ([b6b80a3](https://github.com/ory/kratos/commit/b6b80a3af1162e4009fa8c7c5e9ae7225e941849))
* Properly normalize uppercase mail addresses ([4984e0f](https://github.com/ory/kratos/commit/4984e0fb329291484a54344255f797008142b7cc)):

    Fixes https://github.com/ory/kratos/issues/3187
    Fixes https://github.com/ory/kratos/issues/3289

* Provide index hint in QueryForCredentials ([#3329](https://github.com/ory/kratos/issues/3329)) ([4ba530e](https://github.com/ory/kratos/commit/4ba530ef593272d3cc0a9e1d354e81db495e8686)):

    * fix: provide index hint in QueryForCredentials
    
    * feat: remove customizable join predicate in QueryForCredentials
    
    * chore: remove obsolete config tracer

* Reduce lookups in whoami call ([#3364](https://github.com/ory/kratos/issues/3364)) ([5bb7b0c](https://github.com/ory/kratos/commit/5bb7b0c83b330ee893bdeb4e636655179bd29e39))
* Reintroduce ExpandAll ([#3369](https://github.com/ory/kratos/issues/3369)) ([8f9bff5](https://github.com/ory/kratos/commit/8f9bff527528780b623bf8e4801f7f3c37a5a6f3))
* Remove codeball ([aa29606](https://github.com/ory/kratos/commit/aa296067e2736cad329814f7acffd816ce0d74a3))
* Remove duplicate SessionIssued event ([#3351](https://github.com/ory/kratos/issues/3351)) ([b1e78ad](https://github.com/ory/kratos/commit/b1e78ad3e39418695639e521ddceb64589455d87))
* Return HTTP 400 instead of 500 for bad query parameters ([58258eb](https://github.com/ory/kratos/commit/58258eba99aa15f2ac852123c0200f56518ecb2a))
* **sdk:** Add cookie for updateLogoutFlow ([#3284](https://github.com/ory/kratos/issues/3284)) ([95ed2b9](https://github.com/ory/kratos/commit/95ed2b94cc99d40af6bbe57e5356ec0f28cb9b78)):

    Closes https://github.com/ory/sdk/issues/255

* **sdk:** Update the API spec to reflect the 204 NoContent in DeleteIdentityCredentials ([#3347](https://github.com/ory/kratos/issues/3347)) ([f3dee86](https://github.com/ory/kratos/commit/f3dee869bef0e0dd2d36541823ae57d54ba5788e))
* Settings should persist `return_to` after required mfa login flow ([#3263](https://github.com/ory/kratos/issues/3263)) ([0ed1abd](https://github.com/ory/kratos/commit/0ed1abd391b6b5369862ee5db8faa4f4aaf68b09)):

    * fix: get settings should persist `return_to` when redirecting to aal2
    
    * feat(e2e): verify `return_to` persists in recovery flows
    
    * test: recovery strategy with mfa account
    
    * test: code recovery return to persists to settings with aal2
    
    * u
    
    * fix: return to settings flow after mfa login
    
    * fix(test): login handler
    
    * fix: flow between settings and mfa
    
    * fix: get settings endpoint should redirect to settings ui instead of to itself
    
    * feat(test): preserve URL from various settings flows through login mfa flow
    
    * chore: cleanup
    
    * fix(e2e): recovery return to spa tests
    
    * fix: e2e proxy
    
    * fix: do not always redirect back to settings on mfa
    
    * fix: new settings flow with required mfa shouldn't be added to login flow return_to unless it contains a return_to parameter
    
    * fix(e2e): let test dynamically handle required_aal
    
    * chore: cleanup unused code
    
    * test: `DoesSessionSatisfy` with method options
    
    * test: recovery strategy with aal2

* String to enum for updateVerificationFlowWithLinkMethod Method ([#3279](https://github.com/ory/kratos/issues/3279)) ([34ff1d2](https://github.com/ory/kratos/commit/34ff1d2912e7f7aefb35dae759dce2eb37ecb790)), closes [#2943](https://github.com/ory/kratos/issues/2943)
* Update correct typo ([#3281](https://github.com/ory/kratos/issues/3281)) ([0fea75c](https://github.com/ory/kratos/commit/0fea75c4093d2c7edc84c14f0ab5bebf33a58970)):

    The text for verification code input should be `Verification code` not `Verify code`.

* Update README ([#3363](https://github.com/ory/kratos/issues/3363)) ([c426014](https://github.com/ory/kratos/commit/c4260140966489a05169a0197e209ff98181bc2e))
* Use RETURNING clause for batch create ([#3293](https://github.com/ory/kratos/issues/3293)) ([8ae8783](https://github.com/ory/kratos/commit/8ae8783935292fb011b1018ac7417ed77eb6abb7))
* Use the correct redirect_uri for linkedin social login ([#3269](https://github.com/ory/kratos/issues/3269)) ([27ccecc](https://github.com/ory/kratos/commit/27ccecc1cd490eaa71da7f8235b4b0057b8f14fe))
* Webhook config parse for settings flow ([#3305](https://github.com/ory/kratos/issues/3305)) ([95ad94d](https://github.com/ory/kratos/commit/95ad94d08efdbb369caecaa64cd0a30058c34ed3))

### Code Generation

* Pin v1.0.0 release commit ([41b7c51](https://github.com/ory/kratos/commit/41b7c51c1c6b3bdff9e9ea8bb5e455e3c15c5256))

### Documentation

* Fix typo in readme ([#3299](https://github.com/ory/kratos/issues/3299)) ([b40544e](https://github.com/ory/kratos/commit/b40544e427891f20cea6838e79f4dee5b52ea5d1))

### Features

* Add “provider id” parameter to kratos session ([#3292](https://github.com/ory/kratos/issues/3292)) ([387f5a2](https://github.com/ory/kratos/commit/387f5a2711ca8eee97ad0f6bb2575ec9ba4797d9)), closes [#3283](https://github.com/ory/kratos/issues/3283)
* Add distroless and static images ([#3350](https://github.com/ory/kratos/issues/3350)) ([1e65662](https://github.com/ory/kratos/commit/1e65662c92b107290466c20de38bbdc0571b596a))
* Add return_to parameters to the `createLogout` handler ([#3336](https://github.com/ory/kratos/issues/3336)) ([08fed36](https://github.com/ory/kratos/commit/08fed36973274ef294491d00811bc867f1537d62)):

    * feat: add return_to parameters to the `createLogout` handler
    
    * test: logout take over return_to from create to update
    
    * test(e2e): logout return to
    
    * test(e2e): logout return to
    
    * test: logout return_to isnt applicable to react

* Allow customization of JOIN predicate in QueryForCredentials ([#3253](https://github.com/ory/kratos/issues/3253)) ([8785166](https://github.com/ory/kratos/commit/87851668e776404aabbfbc67af73a43ea3ee28fc))
* Emit events for login/logout and registration ([#3235](https://github.com/ory/kratos/issues/3235)) ([c784b7e](https://github.com/ory/kratos/commit/c784b7e7ed2834ca83c6db2326b735e78e5a75f2))
* Forward `prompt` upstream parameter during OIDC flow ([#3276](https://github.com/ory/kratos/issues/3276)) ([d290cb0](https://github.com/ory/kratos/commit/d290cb05bb4f63d04ec3763db127060e13c350dc)), closes [#2709](https://github.com/ory/kratos/issues/2709)
* Implement `crypt(3)` hashers ([#3303](https://github.com/ory/kratos/issues/3303)) ([afe06db](https://github.com/ory/kratos/commit/afe06db95663cc0cb9704ba4f7014ed9bfb4de09)), closes [#3291](https://github.com/ory/kratos/issues/3291):

    This PR implements md5crypt, sha256crypt, sha512crypt, which are considered legacy (like md5), but are used in legacy systems looking to convert to ory. They use the existing format of crypt(5) (which is compliant to PHC).

* Improve event types and capture more events ([#3297](https://github.com/ory/kratos/issues/3297)) ([835fe13](https://github.com/ory/kratos/commit/835fe13d9ce81f7c0ed91dd2863a740fbb0c6209))
* Lark OIDC provider ([#2925](https://github.com/ory/kratos/issues/2925)) ([f884dfb](https://github.com/ory/kratos/commit/f884dfbaa8aeba58b3b1595bd45e41f9b3e5a0e0))
* Return to oauth flow after switching from login to other flows ([#3212](https://github.com/ory/kratos/issues/3212)) ([a1fea6c](https://github.com/ory/kratos/commit/a1fea6c353768bbf154900766fbbe51f2a148554)):

    * feat: return to oauth flow after switching from login to other flows
    
    * feat(e2e): flows should have return_to set to hydra request_url
    
    * u
    
    * fix: override return_to URL on OAuth flows
    
    * style: format
    
    * fix: TestOAuth2Provider
    
    * feat: config to opt into using OAuth request url as return_to
    
    * chore: cleanup
    
    * fix(e2e): oauth2 login flow switching to recovery
    
    * feat(test): oauth2 login flow to recovery through oidc provider
    
    * fix(e2e): oidc-provider registration
    
    * chore: rename `oauth2_provider.return_to_enabled` to `oauth2_provider.override_return_to`
    
    * style: format
    
    * chore: nit config description
    
    

* Sort sessions by authenticated_at ([#3324](https://github.com/ory/kratos/issues/3324)) ([46f92ff](https://github.com/ory/kratos/commit/46f92ffebf14d1cf4133ca37a2151e8c3aef9d2d)):

    Closes https://github.com/ory/network/issues/295

* Sqa metrics v2 ([#3300](https://github.com/ory/kratos/issues/3300)) ([98fe73f](https://github.com/ory/kratos/commit/98fe73faa75c56be47c19c61a780578ef24e7267))
* Support exporting of all credential types ([#3290](https://github.com/ory/kratos/issues/3290)) ([de6c857](https://github.com/ory/kratos/commit/de6c8574c9c6070458303f9b5caf7e8533f06b69)):

    It's now possible to export all credential types (including passwords) when calling the `getIdentity` SDK method.

* Support OIDC flows for native apps ([#3216](https://github.com/ory/kratos/issues/3216)) ([cb10609](https://github.com/ory/kratos/commit/cb106097210ac9a146738d06c20a4306c2345923)), closes [#707](https://github.com/ory/kratos/issues/707):

    Implements Social Sign In and OpenID Connect for native apps.


### Tests

* Run Playwright in CI ([#3259](https://github.com/ory/kratos/issues/3259)) ([342edec](https://github.com/ory/kratos/commit/342edeced4080a1b914000dfb8427196abebc596)):

    * run Playwright in CI
    
    * add cleanup for session token exchangers
    
    * fixup: ci
    
    * fix: compatibility between OIDC+code and other flows
    
    This improves the compatibility between OIDC+code and other
    flows such as TOTP, settings, password auth.
    
    * Update persistence/sql/persister_cleanup_test.go
    
    
    
    * fix: error handling with OIDC+Code
    
    * fix: increase playwright timeout


### Unclassified

* @barnarddt @hperl feat: send emails via http api endpoint instead of smtp (#1030) (#3341) ([28b7b04](https://github.com/ory/kratos/commit/28b7b04a34eeba2d84de5c543f5ba8b41b38a129)), closes [#1030](https://github.com/ory/kratos/issues/1030) [#3341](https://github.com/ory/kratos/issues/3341) [#1030](https://github.com/ory/kratos/issues/1030) [#3008](https://github.com/ory/kratos/issues/3008):

    This change adds a new delivery method to the courier called `mailer`. Similar to SMS functionality it posts a templated Data model to a API endpoint.  This API can then send emails via a CRM or any other mechanism that it wants.
    
    `Mailer` still uses the existing email data models so any new email added will automatically be sent to the API/CRM as well.
    
    ## Related issue(s)
    Resolves https://github.com/ory/kratos/issues/2825



# [0.13.0](https://github.com/ory/kratos/compare/v0.11.1...v0.13.0) (2023-04-18)

We’re excited to announce the release of Ory Kratos v0.13.0! This update brings many enhancements and fixes, improving the user experience and overall performance. Here are the highlights:

- We’ve added new social sign-in options with Patreon OIDC and LinkedIn providers, making it even easier for your users to register and log in. Furthermore, we’ve introduced a new admin API that allows you to remove specific 2nd factor credentials, giving you more control over your user accounts.
- Performance has been a key focus in this release. We’ve optimized the whoami calls, parallelized the getIdentity and getSession calls, and made asynchronous webhooks fully async. These improvements will result in faster response times and a smoother experience for your users. Additionally, we’ve implemented better tracing to help you diagnose and resolve issues more effectively.
- We’ve also made several updates to the webhook system. A new response.parse configuration has been introduced, allowing you to update identity data during registration. This includes admin/public metadata, identity traits, enabling/disabling identity, and modifying verified/recovery addresses. Please note that can_interrupt is now deprecated in favor of response.parse.
- Lastly, we’ve made several important fixes, such as resolving the wrong message ID on resend code buttons, implementing the offline scope as Google expects, and improving the OIDC flow on duplicate account registration. We’ve also added the ability to configure whether the system should notify unknown recipients when attempting to recover an account or verify an address, enhancing security with “anti-account-enumeration measures.”

We hope you enjoy these new features and improvements in Ory Kratos v0.13.0! All features are already live on the Ory Network - the simplest, fastest and most scalable way to run Ory.

Please note that the v0.12.0 release was skipped due to CI issues.

Head over to the changelog at [https://github.com/ory/kratos/blob/master/CHANGELOG.md](https://github.com/ory/kratos/blob/master/CHANGELOG.md) to read all the details. As always, we appreciate your feedback and support!



## Breaking Changes

By default, Kratos no longer sends out these Emails. If you want to keep notifying unknown addresses (keep the current behavior), set `selfservice.flows.recovery.notify_unknown_recipients` to `true` for recovery, or `selfservice.flows.verification.notify_unknown_recipients` for verification flows.



### Bug Fixes

* Access rules example ([#3178](https://github.com/ory/kratos/issues/3178)) ([a206772](https://github.com/ory/kratos/commit/a206772d78efed6febe783ee88dae92de80063d0))
* Account experience redirects to verification page ([#3195](https://github.com/ory/kratos/issues/3195)) ([2e96d75](https://github.com/ory/kratos/commit/2e96d75c2e0a1c9a884e2d3342725fb1983b495d))
* Account settings broken on OIDC removal ([#3185](https://github.com/ory/kratos/issues/3185)) ([61ae531](https://github.com/ory/kratos/commit/61ae531ba86636e1ad4d63e37df47ef76dfa5f29)), closes [ory-corp/cloud#3514](https://github.com/ory-corp/cloud/issues/3514)
* Add `after_verification_return_to` to sdk and api docs ([#3097](https://github.com/ory/kratos/issues/3097)) ([c70704c](https://github.com/ory/kratos/commit/c70704cebafff7a92f32928273e4570abb3b1c3d)), closes [#3096](https://github.com/ory/kratos/issues/3096)
* Add `HydraLoginRequest` on flow creation ([#3152](https://github.com/ory/kratos/issues/3152)) ([09312dd](https://github.com/ory/kratos/commit/09312dd2d7f89eadbae603e4c8891f39630a2570)), closes [#3108](https://github.com/ory/kratos/issues/3108):

    The oauth2_login_request field was missing when initially creating the login flow.

* Add missing `code` discriminator in updateVerificationFlow ([#3213](https://github.com/ory/kratos/issues/3213)) ([21576be](https://github.com/ory/kratos/commit/21576bebc0d8c3796a4a16b1972ff42889814d61))
* Add missing index ([#3181](https://github.com/ory/kratos/issues/3181)) ([756bed4](https://github.com/ory/kratos/commit/756bed4db3789428117ec105ac0713a52d610938))
* Add mutex to test SMTP server setup/teardown ([20c2359](https://github.com/ory/kratos/commit/20c2359407044c81850759e27b03c371cb0e4886))
* Avoid unchecked casts from IdentityPool to PrivilegedIdentityPool ([71d35dd](https://github.com/ory/kratos/commit/71d35ddd582b3c7081f66e0cdc0c43457816ab25))
* Correctly apply patches to identity metadata ([#3103](https://github.com/ory/kratos/issues/3103)) ([1193a56](https://github.com/ory/kratos/commit/1193a5681fbc25d03c1e26a4296fa0b9abd2452b)), closes [#2950](https://github.com/ory/kratos/issues/2950)
* Do not omit last page on identity list ([#3169](https://github.com/ory/kratos/issues/3169)) ([f95f48a](https://github.com/ory/kratos/commit/f95f48a79395b7b99c7482c0974bc5188e007cc0))
* Don't return 500 if active strategy is disabled ([#3197](https://github.com/ory/kratos/issues/3197)) ([3a734c2](https://github.com/ory/kratos/commit/3a734c2dc2bd848033dbdc7d6116b8b6db6fa760))
* Don't reuse ports in courier/SMTP tests ([#3156](https://github.com/ory/kratos/issues/3156)) ([e260fcf](https://github.com/ory/kratos/commit/e260fcf06181ce9339edc729ab74826aa4be78cf))
* Don't treat missing session as error in tracing ([290d28a](https://github.com/ory/kratos/commit/290d28ada1a55b599af7e41e638de699a474f1d8))
* Error messages in OpenAPI/Swagger / improve error messages from failed webhooks and client timeouts ([#3218](https://github.com/ory/kratos/issues/3218)) ([b1bdcd3](https://github.com/ory/kratos/commit/b1bdcd32828fcdbf65bc43b85b64df210ba4c646))
* Handle upstream errors in patreon provider ([#3032](https://github.com/ory/kratos/issues/3032)) ([39fa31f](https://github.com/ory/kratos/commit/39fa31f85deb3f015aa0f1b30b4a17e4b51d461b))
* Identity.CopyWithoutCredentials ([989c99d](https://github.com/ory/kratos/commit/989c99d6a32e02759a8a7a07606a90832afec460))
* Implement offline scope in the way google expects ([#3088](https://github.com/ory/kratos/issues/3088)) ([39043d4](https://github.com/ory/kratos/commit/39043d451e154af44123ba031381f0e3c10fbb00))
* Improve webhook resilience ([#3200](https://github.com/ory/kratos/issues/3200)) ([0a05d99](https://github.com/ory/kratos/commit/0a05d9941c6be549acfe65a78f4a8b21d6efbcdc)):

    * fix: improve webhook logging
    * chore: bump x
    * feat: decouple context in PostRegistrationPostPersist hook

* Invalid SQL syntax in ListIdentities ([#3202](https://github.com/ory/kratos/issues/3202)) ([162ab9b](https://github.com/ory/kratos/commit/162ab9b5634329135b1b729ad401701019aca222)):

    PostgresQL does not support `... WHERE x IN ( )` with an empty argument list.

* Issuer missing from netid claims ([#3080](https://github.com/ory/kratos/issues/3080)) ([dec7cbc](https://github.com/ory/kratos/commit/dec7cbc4286cbbe2d787b1f8998ee57054d7c95b)):

    The NetID provider omits the issuer claim in the userinfo response. To resolve this issue, the ID token returned by NetID is now validated and its `sub` and `iss` values are used.

* Lint errors and unused code ([ae49ef0](https://github.com/ory/kratos/commit/ae49ef04ed24c23406a5639d34c2e81ab0130c75))
* Make async webhooks fully async ([#3111](https://github.com/ory/kratos/issues/3111)) ([342bfb0](https://github.com/ory/kratos/commit/342bfb0332d235a2d535493d586192815b7d4974))
* Make session AAL satisfaction check resilient against a nil identity in the session ([5ab1a56](https://github.com/ory/kratos/commit/5ab1a56cfd41e95fbb30b8f93426a27e510c62c7)):

    Also fix tracing.

* Missing issuer regression in OIDC ([#3220](https://github.com/ory/kratos/issues/3220)) ([52f0740](https://github.com/ory/kratos/commit/52f07402edac2624cb37c72c768737a785658d29)):

    Closes https://github.com/ory/kratos/issues/3182
    Closes https://github.com/ory/kratos/issues/3040

* Nolint comment ([93e6501](https://github.com/ory/kratos/commit/93e6501c63a253336c081f156ada58458b83ef92))
* Only return one result set for credentials_identifier ([#3107](https://github.com/ory/kratos/issues/3107)) ([59f35d1](https://github.com/ory/kratos/commit/59f35d11e61a246d1079ac02cb8958ba81b37f75)), closes [#3105](https://github.com/ory/kratos/issues/3105)
* Orphaned webhook spans ([a7f9414](https://github.com/ory/kratos/commit/a7f9414460eb214a8f2b2ff96a2b6b303721f806))
* Re-use existing CSRF token in verification flows ([#3188](https://github.com/ory/kratos/issues/3188)) ([08a3447](https://github.com/ory/kratos/commit/08a344761e049c64cffafca2f94c942468201d24)):

    * fix: re-use existing CSRF token in verification flows
    
    * chore: fix if/else

* Reduce SQL tracing noise ([1650426](https://github.com/ory/kratos/commit/1650426a2b59cd46035e5556ff8f69994602e88e))
* Remove `http.Redirect` from `show_verification_ui` hook ([#3238](https://github.com/ory/kratos/issues/3238)) ([054705b](https://github.com/ory/kratos/commit/054705b8c6c933d20b8fb45fcb2593a451cee685))
* Remove network omit flag ([#3066](https://github.com/ory/kratos/issues/3066)) ([c629b72](https://github.com/ory/kratos/commit/c629b72be42001e3e1671d61cc8348373b686844))
* Report correct errors for json schema validation ([#3085](https://github.com/ory/kratos/issues/3085)) ([9477ea4](https://github.com/ory/kratos/commit/9477ea4a7bde6efa73ed94f61c2d4ed66fd43a08)):

    - Implemented the translation of `jsonschema.ValidationError` to errors codes documented [here](https://www.ory.sh/docs/kratos/concepts/ui-user-interface#machine-readable-format)
    - Added missing error codes for relevant schema errors
      | Validation         | Name                            | ID      |
      | ------------------ | ------------------------------- | ------- |
      | `maxLength`        | ErrorValidationMaxLength        | 4000017 |
      | `minimum`          | ErrorValidationMinimum.         | 4000018 |
      | `exclusiveMinimum` | ErrorValidationExclusiveMinimum | 4000019 |
      | `maximum`          | ErrorValidationMaximum          | 4000020 |
      | `exclusiveMaximum` | ErrorValidationExclusiveMaximum | 4000021 |
      | `multipleOf`       | ErrorValidationMultipleOf       | 4000022 |
      | `maxItems`         | ErrorValidationMaxItems         | 4000023 |
      | `minItems`         | ErrorValidationMinItems         | 4000024 |
      | `uniqueItems`      | ErrorValidationUniqueItems      | 4000025 |
      | `type`             | ErrorValidationWrongType        | 4000026 |
    - Updated e2e tests to check these IDs explicitly

* Respect the after recovery return to URL from config ([#3141](https://github.com/ory/kratos/issues/3141)) ([3467fd3](https://github.com/ory/kratos/commit/3467fd3b860dd2ad915449e3fff7e4da2d2c61ca)):

    Fixes https://github.com/ory-corp/cloud/issues/1405

* Set DB connection max idle time ([8d4762c](https://github.com/ory/kratos/commit/8d4762c1bffad14c94ac69575e488fc67d3f5dde))
* Set proper maxAge for session cookies ([#3209](https://github.com/ory/kratos/issues/3209)) ([1180c05](https://github.com/ory/kratos/commit/1180c051b34eb5de786d6b4e4bd94e863f60d06a)), closes [#3208](https://github.com/ory/kratos/issues/3208)
* Sqa config values unified across projects ([#3237](https://github.com/ory/kratos/issues/3237)) ([523b93f](https://github.com/ory/kratos/commit/523b93fd1fe8715d06aeedc2db0ac072dfcafb71))
* Test contract names ([e9ac00b](https://github.com/ory/kratos/commit/e9ac00b3941641a955f5d8f32f25a4031c87a726))
* Use correct names in WebAuthN dialogs ([#3215](https://github.com/ory/kratos/issues/3215)) ([3bc1ff0](https://github.com/ory/kratos/commit/3bc1ff0e63c885c1db08e3d1332d959799edb0a8))
* Use type alias instead of type definition ([#3148](https://github.com/ory/kratos/issues/3148)) ([dba3803](https://github.com/ory/kratos/commit/dba38032d5939ff7286560ec19d83a89fe0410ce))
* Webhook tracing and missing defers ([#3145](https://github.com/ory/kratos/issues/3145)) ([46eb063](https://github.com/ory/kratos/commit/46eb063f414a0ad9b901407cf781002ccb97ad93))
* Wrong context in logout trace span ([#3168](https://github.com/ory/kratos/issues/3168)) ([b9ccccf](https://github.com/ory/kratos/commit/b9ccccf0f1b6a5ba903293133b2be15b528c8308))

### Code Generation

* Pin v0.13.0 release commit ([349d0ee](https://github.com/ory/kratos/commit/349d0ee1899e2ff0f81587b528c04fa0287e5546))

### Code Refactoring

* Identity persistence ([#3101](https://github.com/ory/kratos/issues/3101)) ([ceb5cc2](https://github.com/ory/kratos/commit/ceb5cc2b8a78be2f5b65d9a026c01ff0afe106af))

### Documentation

* Fix broken docs links and code example to get verification flow ([#3170](https://github.com/ory/kratos/issues/3170)) ([bdbddcc](https://github.com/ory/kratos/commit/bdbddcce2909b290e2e04dee493519b842715ab4))
* Update security email ([#3164](https://github.com/ory/kratos/issues/3164)) ([9252f5a](https://github.com/ory/kratos/commit/9252f5a3c746927a2f537efc39cb1eb0aba167a5))

### Features

* Add a new admin API to remove a specific 2nd factor credential ([#2962](https://github.com/ory/kratos/issues/2962)) ([44556a4](https://github.com/ory/kratos/commit/44556a468ef233b18fd0f16a83a4e1b2e5f05dcf)), closes [#2505](https://github.com/ory/kratos/issues/2505)
* Add API to batch insert identities ([#3157](https://github.com/ory/kratos/issues/3157)) ([829bda7](https://github.com/ory/kratos/commit/829bda701acfd6706ffd72845414d177895ff8fe)), closes [ory/network#266](https://github.com/ory/network/issues/266)
* Add Inspect option to driver ([8aa75e9](https://github.com/ory/kratos/commit/8aa75e97e4bfee37e7cf551173b516c6244786ff))
* Add patreon oidc provider ([#3021](https://github.com/ory/kratos/issues/3021)) ([20ea29e](https://github.com/ory/kratos/commit/20ea29e018b33231cf6b2743de74d2233f756c2a))
* Add test to verify GetIdentityConfidential expands everything ([#3217](https://github.com/ory/kratos/issues/3217)) ([f088ccd](https://github.com/ory/kratos/commit/f088ccdf462f5e6373aceb142caa181d98975a09))
* Add token prefixes to session and logout tokens ([#3132](https://github.com/ory/kratos/issues/3132)) ([8210cd0](https://github.com/ory/kratos/commit/8210cd09200d370b101072649fddd1ad9a7f32a9)):

    This feature adds token prefixes to Ory session and logout tokens:
    
    * `ory_st_`: Ory session token prefix
    * `ory_lt_`: Logout token prefix

* Add upstream parameters to oidc provider ([#3138](https://github.com/ory/kratos/issues/3138)) ([b6b1679](https://github.com/ory/kratos/commit/b6b1679c3bd053cd08ff8f26c762735e380fed67)), closes [#3127](https://github.com/ory/kratos/issues/3127) [#2069](https://github.com/ory/kratos/issues/2069):

    This PR introduces the upstream OIDC query parameters `login_hint` and `hd`.
    
    To send additional upstream parameters the form can post this on a login, registration or settings link submit.
    For example the form below does an OIDC flow to Google. We can now add additional parameters such as `login_hint` and `hd` to the upstream request to Google login with a pre-filled email `email@example.com`:
    
    ```html
    <form action="https://kratos/self-service/login?flow=">
      <input type="submit" name="provider" value="google" />
      <input type="hidden" name="upstream_parameters.login_hint" value="email@example.com" />
      <input type="hidden" name="upstream_parameters.hd" value="example.com" />
    </form>
    ```

* Allow importing (salted) SHA hashing algorithms ([#2741](https://github.com/ory/kratos/issues/2741)) ([132255e](https://github.com/ory/kratos/commit/132255eff24a3f5a7fc2249a0ecf9b8716a8f1e7)), closes [#2422](https://github.com/ory/kratos/issues/2422)
* Allow passing transient data from registration to webhook ([#3104](https://github.com/ory/kratos/issues/3104)) ([4a3a076](https://github.com/ory/kratos/commit/4a3a07657d2eb2a39d777565b58882cb48e928fa))
* Don't pre-generate UUIDs for transient objects ([e17f307](https://github.com/ory/kratos/commit/e17f307732f8ced34727d5f3a70929866a0595e0))
* Drop unused index ([#3165](https://github.com/ory/kratos/issues/3165)) ([852dea9](https://github.com/ory/kratos/commit/852dea90881a7c9abdbfc127a2e8d1cc0aacb166))
* Even more tracing of hidden HTTP requests ([9d8b1e2](https://github.com/ory/kratos/commit/9d8b1e223072e66d284c9e7890060678b77c1d4f))
* Identity by identifier ([#3077](https://github.com/ory/kratos/issues/3077)) ([c288d4d](https://github.com/ory/kratos/commit/c288d4d136bca1a9ed3931b4827967eb44e80ede))
* Improve tracing span naming in hooks ([bf828d3](https://github.com/ory/kratos/commit/bf828d3f5d56a963529e98958f4039f0dc569979))
* Improve webhook diagnostics ([d4eb2f6](https://github.com/ory/kratos/commit/d4eb2f6b728a211f1e1454559c2eff73f2f77936))
* Improved oidc flow on duplicate account registration ([#3151](https://github.com/ory/kratos/issues/3151)) ([4d2fda4](https://github.com/ory/kratos/commit/4d2fda453b16349589e941af06fcce312c2e5c37)):

    This PR improves the OIDC registration flow when a duplicate account error happens. 
    
    Currently the flow looks as follows:
    
    1. User registers with password (or other credentials)
    2. User forgot they registered with password and tries to login through an OIDC provider (e.g. Google)
    3. Kratos attempts a registration since the OIDC credentials do not exist
    4. (optional) User needs to add missing traits (e.g. full name) which could not be retrieved from the OIDC provider
    5. User gets a duplicate account error with a "Continue" button.
    6. After submitting the "Continue" button the flow continues again to the OIDC provider, back to Kratos and redirects to UI with duplicate error (Steps 3 to 5)
    
    Instead of causing a confusing redirect loop we should show the user the error with a fresh login flow (since the account exists). This also gives the user the option to do a recovery flow.
    
    1. User registers with password (or other credentials)
    2. User forgot they registered with password and tries to login through an OIDC provider (e.g. Google)
    3. Kratos attempts a registration since the OIDC credentials do not exist
    4. (optional) User needs to add missing traits
    5. User is returned to a Login flow with the duplication error

* Let DB generate ID for session devices ([62402c7](https://github.com/ory/kratos/commit/62402c7bed3c57ef5b957572e4b84f56d9c530ae))
* Make notification to unknown recipients configurable ([#3075](https://github.com/ory/kratos/issues/3075)) ([1a5ead4](https://github.com/ory/kratos/commit/1a5ead43a60e7a0388617877a9f16d1dec61459b)), closes [#2345](https://github.com/ory/kratos/issues/2345) [#2585](https://github.com/ory/kratos/issues/2585):

    Added the ability to configure whether the system should notify unknown recipients, if some tries to recover their account or verify their address ("anti-account-enumeration measures").

* Make password validator (HIBP check) cancelable and add tracing ([28f8914](https://github.com/ory/kratos/commit/28f8914bfb8276d38e08b9be9a3ad1c59d1410bb))
* Parallelize get identity and session calls ([#3023](https://github.com/ory/kratos/issues/3023)) ([6393519](https://github.com/ory/kratos/commit/6393519977bc3d804673b5669166e07c561f1c79))
* Refactor credentials fetching ([#3183](https://github.com/ory/kratos/issues/3183)) ([590269f](https://github.com/ory/kratos/commit/590269f91e24203f987124cfbf11d31c04c1d35c)):

    This change revamps the way we fetch identity credentials. We no longer need most of the helper fields for gobuffalo/pop inside the `Identity` and `Credentials` structures, and we collect all the credentials in one joined query rather than using pop's `EagerPreload` functionality.

* Return hydra error messages ([b3d037b](https://github.com/ory/kratos/commit/b3d037b33b248f1873f09d641e5d61376bcfde80))
* Return verification flow ID after registration flow ([#3144](https://github.com/ory/kratos/issues/3144)) ([eb854be](https://github.com/ory/kratos/commit/eb854becd9fe75213fba6ebe4283cc4ed2c9d128)), closes [#2975](https://github.com/ory/kratos/issues/2975)
* Show "continue" screen after successful verification ([#3090](https://github.com/ory/kratos/issues/3090)) ([fb6b160](https://github.com/ory/kratos/commit/fb6b1600d3d75e5d11fb98445c499a6218e6b869)):

    The `link` strategy for verification now shows a confirmation screen with a "continue" link after successful verification, aligning its behavior to the `code` strategy.
    
    Also fixes a bug, where the `default_browser_return_url` of the verification flow was not respected when using the code strategy.
    
    Closes https://github.com/ory-corp/cloud#3925
    Fixes https://github.com/ory/network#228
    Fixes https://github.com/ory/network/issues/224

* Social sign in via linkedin ([#3079](https://github.com/ory/kratos/issues/3079)) ([5de6bf4](https://github.com/ory/kratos/commit/5de6bf46aba6c13f927ef1c4c425322a34063ca9)), closes [#2856](https://github.com/ory/kratos/issues/2856):

    Adds LinkedIn as a social sign in provider.

* Webhooks that update identities ([2cbee3e](https://github.com/ory/kratos/commit/2cbee3e8eea6bac376faf9382bf5b15acb732f03)), closes [#2161](https://github.com/ory/kratos/issues/2161):

    Introduces a new configuration `response.parse` in webhooks. This enables updating of identity data during registration, including admin/public metadata, identity traits, enabling/disabling identity, and modifying verified/recovery addresses.
    
    Please note that `can_interrupt` is being deprecated in favor of `response.parse`.


### Tests

* **e2e:** Fix compile errors in commands ([#3179](https://github.com/ory/kratos/issues/3179)) ([0002668](https://github.com/ory/kratos/commit/00026682b548b1f33e255a8ee865d90ea127a254))
* Parallelize several unit tests ([#3081](https://github.com/ory/kratos/issues/3081)) ([5403f86](https://github.com/ory/kratos/commit/5403f863d21a6fb5ba4b8572fb054d52e5a8205d))

### Unclassified

* Revert "fix: do not omit last page on identity list (#3169)" (#3184) ([73b5f13](https://github.com/ory/kratos/commit/73b5f13935ef051aae5538cf3d189bb430ea49ae)), closes [#3169](https://github.com/ory/kratos/issues/3169) [#3184](https://github.com/ory/kratos/issues/3184):

    This reverts commit f95f48a79395b7b99c7482c0974bc5188e007cc0.



# [0.11.1](https://github.com/ory/kratos/compare/v0.11.0...v0.11.1) (2023-01-14)

* Fixed several bugs to improve overall stability.
* Optimized performance for faster load times and smoother operation.
* Improved tracing capabilities for better debugging and issue resolution.

We are constantly working to improve Ory Kratos and this release is no exception. Thank you for using Ory and please let us know if you have any feedback or encounter any issues.



## Breaking Changes

The `/admin/courier/messages` endpoint now uses `keysetpagination` instead.



### Bug Fixes

* Add missing indexes ([#2973](https://github.com/ory/kratos/issues/2973)) ([bbb3995](https://github.com/ory/kratos/commit/bbb399572926bd433928b22764f7b3558bb0c21d))
* Add missing indexes for identity delete ([#2952](https://github.com/ory/kratos/issues/2952)) ([dc311f9](https://github.com/ory/kratos/commit/dc311f9a9dc0dbb26e2375b3cd4232a4e8cccb61)):

    This significantly improves the performance of identity deletes.
    
    

* Cors headers not added to the response [#2922](https://github.com/ory/kratos/issues/2922) ([#2934](https://github.com/ory/kratos/issues/2934)) ([1ed6839](https://github.com/ory/kratos/commit/1ed6839369baeecc99610d9f04d78dfee53ad72a))
* Dont reset to false ([#2965](https://github.com/ory/kratos/issues/2965)) ([ae8ad7b](https://github.com/ory/kratos/commit/ae8ad7be5b6f3dbb9142bee55448a71c7df44e52))
* Flaky test now stable ([4e5dcd0](https://github.com/ory/kratos/commit/4e5dcd0df6baffda8b15eda37fd7a247793f3297))
* Listing sessions query ([#2958](https://github.com/ory/kratos/issues/2958)) ([3e06c99](https://github.com/ory/kratos/commit/3e06c991ad557f4629ef7412c256ede2386a7bed)), closes [#2930](https://github.com/ory/kratos/issues/2930)
* Missing index on courier list count ([#3002](https://github.com/ory/kratos/issues/3002)) ([3b50711](https://github.com/ory/kratos/commit/3b507110d6e0296e90d3c495515bf2a066b7c09b))
* Pin geckodriver version to bypass GitHub API quota ([#2972](https://github.com/ory/kratos/issues/2972)) ([585cb9e](https://github.com/ory/kratos/commit/585cb9e79be5de8b3d684313edb72bb703ffaa78))
* Quickstart demos ([#2940](https://github.com/ory/kratos/issues/2940)) ([a7720b2](https://github.com/ory/kratos/commit/a7720b2ba389c08c83c4f3118b83e1fc044773cc))
* Remove duplicate query in GetIdentity ([#2987](https://github.com/ory/kratos/issues/2987)) ([33b01bb](https://github.com/ory/kratos/commit/33b01bbb0e53fc8ac0127531de72ee1b680be656))
* Remove unused x-session-cookie parameter ([#2983](https://github.com/ory/kratos/issues/2983)) ([56b5c26](https://github.com/ory/kratos/commit/56b5c26e666af2442b3e99449b62b2f76a3a4677)):

    This patch removes the undocumented and experimental `X-Session-Cookie` header from the `/sessions/whoami` endpoint.

* Resilient social sign in ([#3011](https://github.com/ory/kratos/issues/3011)) ([ca35b45](https://github.com/ory/kratos/commit/ca35b45a26c6781be81086a7677344fc165dac9f))
* Respect `return_to` URL parameter in registration flow when the user is already registered ([#2957](https://github.com/ory/kratos/issues/2957)) ([3462ce1](https://github.com/ory/kratos/commit/3462ce1512d03529b613421a69bcf4c1d5e98e08))
* Set accept header for GitLab ([#2998](https://github.com/ory/kratos/issues/2998)) ([e892113](https://github.com/ory/kratos/commit/e892113cc00a010490492def7f128bfb5c15b8de))
* Set config at the start ([e58bc6e](https://github.com/ory/kratos/commit/e58bc6e9bacd5c9c6ee9369beb843a4c54059ae2))
* Spurious cancelation of async webhooks, better tracing ([#2969](https://github.com/ory/kratos/issues/2969)) ([72de640](https://github.com/ory/kratos/commit/72de640bad75da29424222bd613a21d10e1811ec)):

    Previously, async webhooks (response.ignore=true) would be canceled
    early once the incoming Kratos request was served and it's associated
    context released. We now dissociate the cancellation of async hooks
    from the normal request processing flow.

* TOTP internal context after saving settings ([#2960](https://github.com/ory/kratos/issues/2960)) ([8b647b1](https://github.com/ory/kratos/commit/8b647b1f54bb674982b982ce483fbd877e42c43a)), closes [#2680](https://github.com/ory/kratos/issues/2680)
* Update pquerna/otp to fix TOTP URL encoding ([#2951](https://github.com/ory/kratos/issues/2951)) ([7248636](https://github.com/ory/kratos/commit/72486368f5403c02772e4a99ed9edc34e84c217c)):

    v1.4.0 fixes generating TOTP URLs. Query params now use %20 instead of +
    to encode spaces. + was not correctly interpreted by some Android
    authenticator apps, and would show up in the issuer name, e.g. "My+Issuer"
    instead of "My Issuer".
    
    

* Update year ([d77e2cf](https://github.com/ory/kratos/commit/d77e2cf56ceab4c73e1c2fd579d43ae25a19d345))
* Webhook tracing instrumentation+memory leak ([f0044a3](https://github.com/ory/kratos/commit/f0044a365b39a5f940d6d268977744f8fcb2e49b))

### Code Generation

* Pin v0.11.1 release commit ([41595c5](https://github.com/ory/kratos/commit/41595c52cf48e2bae81b1a901577062cc6e3dc06))

### Documentation

* Improve api headline ([#2989](https://github.com/ory/kratos/issues/2989)) ([fc2787b](https://github.com/ory/kratos/commit/fc2787ba9a5cb9088a76b7ec25752d75ef399281))

### Features

* Add client IP to span events ([7ce3a74](https://github.com/ory/kratos/commit/7ce3a7471243898e111ca3e2b5d1346131c55dae))
* Add NID to logs in courier ([#2956](https://github.com/ory/kratos/issues/2956)) ([b407aa9](https://github.com/ory/kratos/commit/b407aa9427382f38dd8a992a6998202a7b6ba83a))
* Improve error message when no session is found ([#2988](https://github.com/ory/kratos/issues/2988)) ([7ad2b97](https://github.com/ory/kratos/commit/7ad2b970089cee2209b3afeaaffd7e04f803918d))
* Improve tracing ([#2992](https://github.com/ory/kratos/issues/2992)) ([04d0280](https://github.com/ory/kratos/commit/04d0280ca1338b93ac6e3026a8a2d852fbb46ef2))
* Remove duplicate queries from whoami calls ([#2995](https://github.com/ory/kratos/issues/2995)) ([b50a222](https://github.com/ory/kratos/commit/b50a22298eedef30a45979866163921604bc698a)), closes [#2402](https://github.com/ory/kratos/issues/2402):

    Introduces an expand API to the identity persister which greatly improves whoami performance.

* Require verification on login ([#2927](https://github.com/ory/kratos/issues/2927)) ([efb8ae8](https://github.com/ory/kratos/commit/efb8ae89cbc31477c2696a0df4c89d6dbf856d27))
* Store errors of courier message ([#2914](https://github.com/ory/kratos/issues/2914)) ([fc7aa86](https://github.com/ory/kratos/commit/fc7aa86545f9e74c22738891af92abafe0030d7f))

### Tests

* Improve parallelization ([e8e8ce5](https://github.com/ory/kratos/commit/e8e8ce5eb3713f28ce1c9a05564ec7f74b48ab4d))
* Regenerate csrf if verification flow expired ([#2455](https://github.com/ory/kratos/issues/2455)) ([7025081](https://github.com/ory/kratos/commit/7025081b76171ce0a8f312a7b671aead1bb21215))
* Update integrity snapshots ([#3000](https://github.com/ory/kratos/issues/3000)) ([6d26e5c](https://github.com/ory/kratos/commit/6d26e5c735a28ecb8b2d8cd142751ef679e19e86))


# [0.11.0](https://github.com/ory/kratos/compare/v0.11.0-alpha.0.pre.2...v0.11.0) (2022-12-02)

The 2022 winter release of Ory Kratos is here, and we are extremely excited to share with you some of the highlights included:

* Ory Kratos now supports verification and recovery codes, which replace are now the default strategy and should be used instead of magic links.
* Import of MD5-hashed passwords is now supported.
* Ory Kratos can now act as the login app for the Ory Hydra Consent & Login Flow using the `oauth2_provider.url` configuration value.
* Ory Kratos' SDK is now released as version 1. Learn more in the [upgrade guide](https://www.ory.sh/docs/guides/upgrade/sdk-v1).
* New APIs are available to manage Ory Sessions.
* Ory Sessions now contain device information.
* Added all claims to the Social Sign-In data mapper as well as the option to customize admin and public metadata.
* Add webhooks that can block the request, useful to do some additional validation.
* Add asynchronous webhooks which do not block the request.
* A CLI helper to clean up stale data.

Please read the changelog carefully to identify changes which might affect you. Always test upgrading with a copy of your production system before applying the upgrade in production.





### Code Generation

* Pin v0.11.0 release commit ([59c30b6](https://github.com/ory/kratos/commit/59c30b6860b56990e132416366e0ae6abe7a275f))

### Features

* Forward parsed request cookies to webhook Jsonnet snippet ([#2917](https://github.com/ory/kratos/issues/2917)) ([70ed068](https://github.com/ory/kratos/commit/70ed068debe7a711ba36e2eb4fcf60be8cae4681)):

    Request cookies were already available in raw form in
    the ctx.request_headers top-level argument to the Jsonnet snippet.
    Parsing cookies in Jsonnet is tedious and error-prone, though, so
    we parse them internally for convenience.



# [0.11.0-alpha.0.pre.2](https://github.com/ory/kratos/compare/v0.10.1...v0.11.0-alpha.0.pre.2) (2022-11-28)

autogen: pin v0.11.0-alpha.0.pre.2 release commit



## Breaking Changes

This patch changes the behavior of the recovery flow. It introduces a new strategy for account recovery that sends out short "one-time passwords" (`code`) that a user can use to prove ownership of their account and recovery access to it. This PR also updates the default recovery strategy to `code`.

This patch invalidates recovery flows initiated using the Admin API. Please re-generate any admin-generated recovery flows and tokens. 

This is a breaking change, as it removes the `courier.message_ttl` config key and replaces it with a counter `courier.message_retries`.

Closes https://github.com/ory/kratos/issues/402
Closes https://github.com/ory/kratos/issues/1598

SDK Method `getJsonSchema` was renamed to `getIdentitySchema`.



### Bug Fixes

* Active attribute based off IsActive checks ([#2901](https://github.com/ory/kratos/issues/2901)) ([bcbf68e](https://github.com/ory/kratos/commit/bcbf68e716aa62f684acbe91e8c35f6c006a4706))
* Add issuerURL for apple id ([#2565](https://github.com/ory/kratos/issues/2565)) ([2aeb0a2](https://github.com/ory/kratos/commit/2aeb0a210e6e6433f1a9d9e6a75b21b8e3083239)):

    No issuer url was specified when using the Apple ID provider,
    this forced usersers to manually enter it in the provider config.
    
    This PR adds the Apple ID issuer url to the provider simplifying the setup.

* Add missing go.mod to docker build ([7c4964e](https://github.com/ory/kratos/commit/7c4964ef65769b40f1ec572a87c2c4106a800bf9))
* Add support for verified Graph API calls for facebook oidc provider ([#2547](https://github.com/ory/kratos/issues/2547)) ([1ba7c66](https://github.com/ory/kratos/commit/1ba7c66fc4897b676690f0ac701a0b68aee4f151))
* Admin recovery CSRF & duplicate form elements ([#2846](https://github.com/ory/kratos/issues/2846)) ([de80b7f](https://github.com/ory/kratos/commit/de80b7f508afdd56f5d8396f03919bd9a98e49d3))
* Bump docker image ([#2594](https://github.com/ory/kratos/issues/2594)) ([071c885](https://github.com/ory/kratos/commit/071c885d8231a1a66051002ecfcff5c8e5237085))
* Bump graceful to deal with http header timeouts ([9ce2d26](https://github.com/ory/kratos/commit/9ce2d260338f020e2da077e81464e520883f582b))
* Cache migration status ([#2631](https://github.com/ory/kratos/issues/2631)) ([9020738](https://github.com/ory/kratos/commit/902073836e4dcf6dc87776921e7988d795943718)):

    See https://github.com/ory-corp/cloud/issues/2691

* Check return code of ms graphapi /me request. ([#2647](https://github.com/ory/kratos/issues/2647)) ([3f490a3](https://github.com/ory/kratos/commit/3f490a31cddc53ce5d9958454f41c352580904c9))
* **cli:** Dry up code ([#2572](https://github.com/ory/kratos/issues/2572)) ([d1b6b40](https://github.com/ory/kratos/commit/d1b6b40aa9dcc7a3ec9237eec28c4fa55f0b8627))
* Codecov ([#2879](https://github.com/ory/kratos/issues/2879)) ([e446c5a](https://github.com/ory/kratos/commit/e446c5a53dbe9963e8047a3e9ca443fa6a7e64eb))
* Correct name of span on recovery code deletion ([#2823](https://github.com/ory/kratos/issues/2823)) ([44f775f](https://github.com/ory/kratos/commit/44f775f45d47eff63379d77a2339b824a6ede235))
* Correctly calculate `expired_at` timestamp for FlowExpired errors ([#2836](https://github.com/ory/kratos/issues/2836)) ([ddde43e](https://github.com/ory/kratos/commit/ddde43ec0d77a1214cd03e1f3e48ab4c34193779))
* Debugging Docker setup ([#2616](https://github.com/ory/kratos/issues/2616)) ([aaabe75](https://github.com/ory/kratos/commit/aaabe754659b96d2a5b727c4cada3ec300624434))
* Disappearing title label on verification and recovery flow ([#2613](https://github.com/ory/kratos/issues/2613)) ([29aa3b6](https://github.com/ory/kratos/commit/29aa3b6c37b3a173dcfeb02fdad4abc83774bc0b)), closes [#2591](https://github.com/ory/kratos/issues/2591)
* Distinguish credential types properly when collecting identifiers ([#2873](https://github.com/ory/kratos/issues/2873)) ([705f7b1](https://github.com/ory/kratos/commit/705f7b105c98b1d68b3e35d6e6893e9cfb661548))
* Do not crash process on invalid smtp url ([#2890](https://github.com/ory/kratos/issues/2890)) ([c5d3ebc](https://github.com/ory/kratos/commit/c5d3ebc6927f7293ee05b65aee745a19ec96ce77)):

    Closes https://github.com/ory-corp/cloud/issues/3321

* Do not double-commit webhooks on registration ([#2888](https://github.com/ory/kratos/issues/2888)) ([88e75d9](https://github.com/ory/kratos/commit/88e75d997348450b1a2a3e4619bcbd614a5582e8))
* Do not invalidate recovery addr on update ([#2699](https://github.com/ory/kratos/issues/2699)) ([1689bb9](https://github.com/ory/kratos/commit/1689bb9f0a52387f699568da6bc773929b1201ae))
* **docker:** Add missing dependencies ([#2643](https://github.com/ory/kratos/issues/2643)) ([c589520](https://github.com/ory/kratos/commit/c589520ff865cefdb287e597b9e858851a778755))
* **docker:** Update images ([b5f80c1](https://github.com/ory/kratos/commit/b5f80c1198e4bb9ed392521daca934548eb21ee6))
* Duplicate messages in recovery flow ([#2592](https://github.com/ory/kratos/issues/2592)) ([43fcc51](https://github.com/ory/kratos/commit/43fcc51b9bf6996fc4f7b0ef797189eb8f3978dc))
* Express e2e tests for new account experience ([#2708](https://github.com/ory/kratos/issues/2708)) ([84ea0cf](https://github.com/ory/kratos/commit/84ea0cf4c72b14f246835d435d22a31f96d9e644))
* Format ([0934def](https://github.com/ory/kratos/commit/0934defff7a0d56e712af98c1cec87c60b3c934b))
* Format check stage in the CI ([#2737](https://github.com/ory/kratos/issues/2737)) ([bbe4463](https://github.com/ory/kratos/commit/bbe44632de77cfb3d4983b68647107d914cd4c46))
* Gosec false positives ([e3e7ed0](https://github.com/ory/kratos/commit/e3e7ed08f5ce47fc794bd5c093018cee51baf689))
* Identity sessions list response includes pagination headers ([#2763](https://github.com/ory/kratos/issues/2763)) ([0c2efa2](https://github.com/ory/kratos/commit/0c2efa2d4345c035649208a71332a64c225313c3)), closes [#2762](https://github.com/ory/kratos/issues/2762)
* **identity:** Migrate identity_addresses to lower case ([#2517](https://github.com/ory/kratos/issues/2517)) ([c058e23](https://github.com/ory/kratos/commit/c058e23599d994e12b676e87f7282c1f2b2e089c)), closes [#2426](https://github.com/ory/kratos/issues/2426)
* Ignore commata in HIBP response ([0856bd7](https://github.com/ory/kratos/commit/0856bd719b7e06a6d2163bf428ff6513d86376db))
* Ignore CSRF for session extension on public route ([866b472](https://github.com/ory/kratos/commit/866b472750fba7bf498d359796f24867af7270ad))
* Ignore error explicitly ([772d596](https://github.com/ory/kratos/commit/772d5968d5a0cb7ac9415cfb2b1e9e86ae3a3131))
* Improve migration status speed ([#2637](https://github.com/ory/kratos/issues/2637)) ([a2e3c41](https://github.com/ory/kratos/commit/a2e3c41f9e513e1de47f6320f6a10acd1fed5eea))
* Include flow id in use recovery token query ([#2679](https://github.com/ory/kratos/issues/2679)) ([d56586b](https://github.com/ory/kratos/commit/d56586b028d79387886f880c1455edb5e4df2209)):

    This PR adds the `selfservice_recovery_flow_id` to the query used when "using" a token in the recovery flow.
    
    This PR also adds a new enum field for `identity_recovery_tokens` to distinguish the two flows: admin versus self-service recovery.

* Include metadata_admin in admin identity list response ([#2791](https://github.com/ory/kratos/issues/2791)) ([aa698e0](https://github.com/ory/kratos/commit/aa698e03a3a96abf1563aea24273735bd9cc412d)), closes [#2711](https://github.com/ory/kratos/issues/2711)
* Incorrect swagger annotation for `getSession` ([#2891](https://github.com/ory/kratos/issues/2891)) ([797ea68](https://github.com/ory/kratos/commit/797ea6857e29e5477e0769af5dd51dd7e43080b2))
* **lint:** Fixed lint error causing ci failures ([4aab5e0](https://github.com/ory/kratos/commit/4aab5e0114dd02b8b0ce45376a0fe4bf11e38221))
* Make `courier.TemplateType` an enum ([#2875](https://github.com/ory/kratos/issues/2875)) ([65aeb0a](https://github.com/ory/kratos/commit/65aeb0a7fd90bfbc81f68b77141f8271aef011fe))
* Make hydra consistently localhost ([70211a1](https://github.com/ory/kratos/commit/70211a17a452d5ced8317822afda3f8e6185cc71))
* Make ID field in VerifiableAddress struct optional ([#2507](https://github.com/ory/kratos/issues/2507)) ([0844b47](https://github.com/ory/kratos/commit/0844b47c30851c548d46273927afee103cdc0e97)), closes [#2506](https://github.com/ory/kratos/issues/2506)
* Make servicelocator explicit ([4f841da](https://github.com/ory/kratos/commit/4f841dae5423acf3514d50add9e99d28bc339fbb))
* Make swagger/openapi go 1.19 compatible ([fec6772](https://github.com/ory/kratos/commit/fec6772739129e0d5bb4103c717b1ac60df45aa8))
* Mark gosec false positives ([13eaddb](https://github.com/ory/kratos/commit/13eaddb7babe630750361c6d8f3ffc736898ddec))
* Metadata should not be required ([05afd68](https://github.com/ory/kratos/commit/05afd68381abe58c5e7cdd51cbf0ae409f5f0eb0))
* Migration error detection ([a115486](https://github.com/ory/kratos/commit/a11548603a4c9b46ba238d2a7ee58fffb7f6d857))
* Missing usage to recovery_code_invalid template ([#2798](https://github.com/ory/kratos/issues/2798)) ([5ac7553](https://github.com/ory/kratos/commit/5ac7553d191885957215b5a63f3bbdc2d020f3fe))
* Not cleared field validation message ([#2800](https://github.com/ory/kratos/issues/2800)) ([cdaf68d](https://github.com/ory/kratos/commit/cdaf68db8e6dd7bacfdb5fc6ff28e5d960f75c2c))
* Panic ([1182278](https://github.com/ory/kratos/commit/11822789c1561b27c2d769c9ea53a81835702f4a))
* Patch invalidates credentials ([#2721](https://github.com/ory/kratos/issues/2721)) ([c4d95af](https://github.com/ory/kratos/commit/c4d95afac590136acd14efa093f48c301fd07164)), closes [ory/cloud#148](https://github.com/ory/cloud/issues/148)
* Potentially resolve tx issue in crdb ([#2595](https://github.com/ory/kratos/issues/2595)) ([9d22035](https://github.com/ory/kratos/commit/9d22035695b6a793ac4bc5e2bd0a68b3aeea039c))
* Preserve return_to param between flows ([#2644](https://github.com/ory/kratos/issues/2644)) ([f002649](https://github.com/ory/kratos/commit/f002649d45658a1486fac551d8ca6b37b3d03026))
* Proper annotation for patch ([#2784](https://github.com/ory/kratos/issues/2784)) ([0cbfe41](https://github.com/ory/kratos/commit/0cbfe410c50cfe551693683881b4145d115c1aa3))
* Re-add service to quickstart ([8c52c33](https://github.com/ory/kratos/commit/8c52c33cf277eda82c9b00b77cd9e03f1e5b4602))
* Re-issue outdated cookie in /whoami ([#2598](https://github.com/ory/kratos/issues/2598)) ([bf6f27e](https://github.com/ory/kratos/commit/bf6f27e37b8aa342ae002e0a9f227a31e0f7c279)), closes [#2562](https://github.com/ory/kratos/issues/2562)
* Remove jackc rewrites ([#2634](https://github.com/ory/kratos/issues/2634)) ([fe00c5b](https://github.com/ory/kratos/commit/fe00c5be72b0cdcc8d462a97aa04c413f758e8e3))
* Remove jsonnet import support ([d708c81](https://github.com/ory/kratos/commit/d708c81abbec424e4376a68140e5008bdba4eaaf))
* Remove newline sign from email subject ([#2576](https://github.com/ory/kratos/issues/2576)) ([ca3d9c2](https://github.com/ory/kratos/commit/ca3d9c24e25ce501e9eae23547f87e1c35b2ea97))
* Remove rust workaround ([355ec43](https://github.com/ory/kratos/commit/355ec431a304eef236a088571e2414f96c49d862))
* Replace io/util usage by io and os package ([e2d805b](https://github.com/ory/kratos/commit/e2d805b7e336d202f7cf3c2e0ce586d78ac03cc0))
* Resolve bug where 500s in web hooks are not properly retried ([e572e81](https://github.com/ory/kratos/commit/e572e8185e17839addabf2a72f4e9921bda8b47a))
* Respect more http sources for computing request URL ([66a9448](https://github.com/ory/kratos/commit/66a94488eb2fc778a00a5c69916e7958b3535440))
* Return browser to 'return_to' when logging in without registered account using oidc.  ([#2496](https://github.com/ory/kratos/issues/2496)) ([a4194f5](https://github.com/ory/kratos/commit/a4194f58dd4ccecca6698d5b43284d857a70a221)), closes [#2444](https://github.com/ory/kratos/issues/2444)
* Return empty array not null when there are no sessions ([#2548](https://github.com/ory/kratos/issues/2548)) ([fffba47](https://github.com/ory/kratos/commit/fffba473440fec3118a3951b697d5a0d2d4e30d6))
* Revert Go 1.19 formatting changes ([7fb085b](https://github.com/ory/kratos/commit/7fb085b6ca4fbfe2978998bea868959966ae193d))
* Revert removal of required field in uiNodeInputAttributes ([#2623](https://github.com/ory/kratos/issues/2623)) ([fee154b](https://github.com/ory/kratos/commit/fee154b28dfb3007f8d20a807cfd6d362c3bd9e7))
* **sdk:** Identity metadata is nullable ([#2841](https://github.com/ory/kratos/issues/2841)) ([4c70578](https://github.com/ory/kratos/commit/4c7057823b5292cb38f43bd5a96041aed178ad0a)):

    Closes https://github.com/ory/sdk/issues/218

* **sdk:** Make InputAttributes.Type an enum ([ff6190f](https://github.com/ory/kratos/commit/ff6190f31f538cf8ed735dfd1bb3b7afcd944c36))
* **sdk:** Rust compile issue with required enum ([#2619](https://github.com/ory/kratos/issues/2619)) ([8800085](https://github.com/ory/kratos/commit/8800085d5bde32367217170d00f7141b7ea46733))
* Send out correct verification invalid email in code strategy ([#2908](https://github.com/ory/kratos/issues/2908)) ([d2bb67a](https://github.com/ory/kratos/commit/d2bb67af64d031613f2516b4848208d4f709e7b4))
* Set cache default to false ([#2906](https://github.com/ory/kratos/issues/2906)) ([e407f92](https://github.com/ory/kratos/commit/e407f92572b7823f70df17d463400807f14c8ae8))
* Take over return_to param from unauthorized settings to login flow ([#2787](https://github.com/ory/kratos/issues/2787)) ([504fb36](https://github.com/ory/kratos/commit/504fb36b6e72900808666dde778906a069f3c48b))
* Unable to find JSON Schema ID: default ([#2393](https://github.com/ory/kratos/issues/2393)) ([f43396b](https://github.com/ory/kratos/commit/f43396bdc03f89812f026c2a94b0b50100134c23))
* Use correct download location for golangci-lint ([c36ca53](https://github.com/ory/kratos/commit/c36ca53d4552596e62ec323795c3bf21438d4f26))
* Use errors instead of fatal for serve cmd ([02f7e9c](https://github.com/ory/kratos/commit/02f7e9cfd17ab60c3f38aab3ae977c427b26990d))
* Use full URL for webhook payload ([72595ad](https://github.com/ory/kratos/commit/72595adcb68a1a2d350c4687328653e28d888847))
* Use process-isolated Jsonnet VM ([#2869](https://github.com/ory/kratos/issues/2869)) ([9eeedc0](https://github.com/ory/kratos/commit/9eeedc06408c447077b630fff65e9ca4ed1ec59a))
* Verification redirect & continue label ([#2905](https://github.com/ory/kratos/issues/2905)) ([e1119e8](https://github.com/ory/kratos/commit/e1119e8f2e0372152d7d8367e7843fd5a49bf728)):

    This PR resolves an issue with the redirect after a successful verification, if not specified.

* Wrap migration error in WithStack ([#2636](https://github.com/ory/kratos/issues/2636)) ([4ce9f1e](https://github.com/ory/kratos/commit/4ce9f1ebb39cccfd36c4f0fb4a2ae2a17fbc18cc))
* Wrong config key in admin recovery documentation ([#2815](https://github.com/ory/kratos/issues/2815)) ([154b61b](https://github.com/ory/kratos/commit/154b61b9ff50306c540eb0904ae012195e735da4))
* X-forwarded-for header parsing ([#2807](https://github.com/ory/kratos/issues/2807)) ([4682afa](https://github.com/ory/kratos/commit/4682afaca3655dc809582b775a5a1c56205a4b4a))

### Code Generation

* Pin v0.11.0-alpha.0.pre.2 release commit ([624e1f0](https://github.com/ory/kratos/commit/624e1f0d23b1c58bc28b2eaf845d4ef63e64bdba))

### Code Refactoring

* Hot reloading ([b0d8f38](https://github.com/ory/kratos/commit/b0d8f3853886228a64e82437643a82b3970d6ff7))
* Make embedding easier with internal sdk ([e9aa21f](https://github.com/ory/kratos/commit/e9aa21f02b4bb7b09e268197334beb9c5772d13d))
* SDK v1 naming ([11f9d30](https://github.com/ory/kratos/commit/11f9d30a5d245b4dfc922a766853eaac2a20a8f5)):

    Find the full [upgrade guide in our documentation](https://www.ory.sh/docs/guides/upgrade/sdk).

* **sdk:** Rename `getJsonSchema` to `getIdentitySchema` ([#2606](https://github.com/ory/kratos/issues/2606)) ([8dc2ecf](https://github.com/ory/kratos/commit/8dc2ecf4919c9a14ef0bd089677de66ab3cfed92))
* Use gotemplates for command usage ([baa84c6](https://github.com/ory/kratos/commit/baa84c681b0c7fa29d653bd7226e792a5f44cb4c))
* Use gotemplates for command usage ([#2770](https://github.com/ory/kratos/issues/2770)) ([1d22b23](https://github.com/ory/kratos/commit/1d22b235291ce7102dd186a53a431b55780973d3))

### Documentation

* Cleanup v0alpha2 endpoint summaries ([db9a95b](https://github.com/ory/kratos/commit/db9a95b6d28f7db3416c9d1530be4fd63a17ac6b))
* Cypress on arm based mac ([#2795](https://github.com/ory/kratos/issues/2795)) ([d8514b5](https://github.com/ory/kratos/commit/d8514b50b5df9c098c77c5cb817602657b2a02ea))
* Enable 2FA methods in docker-compose quickstart setup ([#2828](https://github.com/ory/kratos/issues/2828)) ([8f52e8b](https://github.com/ory/kratos/commit/8f52e8b728bf8e2a99807f4d4899c2eaaca9e7e5))
* Fix badge ([dbb7506](https://github.com/ory/kratos/commit/dbb7506ec1a5a2b5bef21cb7838b6c86e755f0f9))
* Importing credentials supported ([4e8b5cf](https://github.com/ory/kratos/commit/4e8b5cf775c1bfe4c2eb5588bfebe900d1c390eb))
* **sdk:** Identifier is actually required ([#2593](https://github.com/ory/kratos/issues/2593)) ([f89d279](https://github.com/ory/kratos/commit/f89d2794d8a2122e3f86eeb8aa5d554da32e753e))
* **sdk:** Incorrect URL ([#2521](https://github.com/ory/kratos/issues/2521)) ([ac6c4cc](https://github.com/ory/kratos/commit/ac6c4ccfc1901d38855ecd9991ef8de80e9d7c40))
* Update README ([5da4c6b](https://github.com/ory/kratos/commit/5da4c6b934b1b820d4a6ca67621855e87ecef773))
* Update readme badges ([7136e94](https://github.com/ory/kratos/commit/7136e94028dc64877e887776a1ccafb8826ce23c))
* Write messages as single json document ([#2519](https://github.com/ory/kratos/issues/2519)) ([3d8cf38](https://github.com/ory/kratos/commit/3d8cf38ef05c6ca5edf1161846c63bd3a23d9adc)), closes [#2498](https://github.com/ory/kratos/issues/2498)

### Features

* Add "success" UITextType ([#2900](https://github.com/ory/kratos/issues/2900)) ([2ff34b6](https://github.com/ory/kratos/commit/2ff34b604757c46aae5cf3cbb23f39f982341486))
* Add admin get api for session ([#2855](https://github.com/ory/kratos/issues/2855)) ([1aa1321](https://github.com/ory/kratos/commit/1aa13211d1459e7453c2ba8fec69fee1c79aecbc))
* Add api endpoint to fetch messages ([#2651](https://github.com/ory/kratos/issues/2651)) ([5fddcbf](https://github.com/ory/kratos/commit/5fddcbf6554264766301e63ed3889ba746f0cd1a)):

    Closes https://github.com/ory/kratos/issues/2639
    
    

* Add autocomplete attributes ([#2523](https://github.com/ory/kratos/issues/2523)) ([6284a9a](https://github.com/ory/kratos/commit/6284a9a5152924018d85f306e5758e9d8d759283)), closes [#2396](https://github.com/ory/kratos/issues/2396)
* Add cache headers ([#2817](https://github.com/ory/kratos/issues/2817)) ([71e2449](https://github.com/ory/kratos/commit/71e2449d7038594e107f39934e4716f845be7bb7))
* Add codecov yaml ([90da0bb](https://github.com/ory/kratos/commit/90da0bb4aeb50ed697c998342300cc56de5d5e1c))
* Add DingTalk social login ([#2494](https://github.com/ory/kratos/issues/2494)) ([7b966bd](https://github.com/ory/kratos/commit/7b966bd16333f419b2a57f2a0b8684d6d86b34e6))
* Add flow id check to use verification token ([#2695](https://github.com/ory/kratos/issues/2695)) ([54c64fc](https://github.com/ory/kratos/commit/54c64fcea40ede17a87253042259fd97eeb780fe))
* Add handler with openapi def for admin revoke session ([#2867](https://github.com/ory/kratos/issues/2867)) ([2438ca0](https://github.com/ory/kratos/commit/2438ca0c9aed997870dcf60d41dad783838dd840))
* Add identity id to "account disabled" error ([#2557](https://github.com/ory/kratos/issues/2557)) ([f09b1b3](https://github.com/ory/kratos/commit/f09b1b3701c6deda4d25cebb7ccf2e97089be32a))
* Add missing config entry ([8fe9de6](https://github.com/ory/kratos/commit/8fe9de6d60a381611e07226614241a83b0010126))
* Add missing cookie headers to SDK methods ([#2720](https://github.com/ory/kratos/issues/2720)) ([32e32d1](https://github.com/ory/kratos/commit/32e32d1b98404ac14a44b2f0ccefa8c02d38c5f7)):

    See https://github.com/ory/kratos/discussions/2583

* Add OpenTelemetry span events ([#2858](https://github.com/ory/kratos/issues/2858)) ([37b1a3b](https://github.com/ory/kratos/commit/37b1a3bb0cf2ea859d672674ca0e95893e63301b))
* Add PATCH to adminUpdateIdentity ([#2380](https://github.com/ory/kratos/issues/2380)) ([#2471](https://github.com/ory/kratos/issues/2471)) ([94a3741](https://github.com/ory/kratos/commit/94a37416011086582e309f62dc2c45ca84083a33))
* Add pre-hooks to settings, verification, recovery ([c0ceaf3](https://github.com/ory/kratos/commit/c0ceaf31f9327cca903c19b77597cae4587737e6))
* Add session cache header feature flag ([#2899](https://github.com/ory/kratos/issues/2899)) ([02a92b4](https://github.com/ory/kratos/commit/02a92b4d8ab5ced5d0d9387b38491990fa7cb724)), closes [ory-corp/cloud#3283](https://github.com/ory-corp/cloud/issues/3283)
* Add support for firebase scrypt hashes on identity import and login hash upgrade ([#2734](https://github.com/ory/kratos/issues/2734)) ([3852eb4](https://github.com/ory/kratos/commit/3852eb460251a079bad68d08bee2aef23516d168)), closes [#2422](https://github.com/ory/kratos/issues/2422)
* Add verification via `code` ([#2838](https://github.com/ory/kratos/issues/2838)) ([a82ee92](https://github.com/ory/kratos/commit/a82ee9295681b8dde96c3c6fb156e791df68613c)), closes [#2824](https://github.com/ory/kratos/issues/2824):

    The new `code` strategy is now supported as a verification strategy. If enabled, the strategy sends a code, instead of a magic link to the user's address, which they can use to verify their address.

* Adding admin session listing api ([#2818](https://github.com/ory/kratos/issues/2818)) ([59588d2](https://github.com/ory/kratos/commit/59588d2e290a8b72125021fa899661622e4cd946))
* Adding device information to the session ([#2715](https://github.com/ory/kratos/issues/2715)) ([82bc9ce](https://github.com/ory/kratos/commit/82bc9ce00d44085287e6d8d9e3fb67e107be2503)):

    Closes https://github.com/ory/kratos/issues/2091
    See https://github.com/ory-corp/cloud/issues/3011
    
    
    Co-authored-by: Patrik <zepatrik@users.noreply.github.com>

* Allow importing scrypt hashing algorithm ([#2689](https://github.com/ory/kratos/issues/2689)) ([3e3b59e](https://github.com/ory/kratos/commit/3e3b59e53de8cb89e9fd01cfec75a0f8a601035b)), closes [#2422](https://github.com/ory/kratos/issues/2422):

    It is now possible to import scrypt-hashed passwords.

* Allow setting public and admin metadata with the jsonnet data mapper ([#2569](https://github.com/ory/kratos/issues/2569)) ([aa6eb13](https://github.com/ory/kratos/commit/aa6eb13c1c42c11354074553fac9c90ee0a8999e)), closes [#2552](https://github.com/ory/kratos/issues/2552)
* Automatic TLS certificate reloading ([#2744](https://github.com/ory/kratos/issues/2744)) ([09751e6](https://github.com/ory/kratos/commit/09751e6a03783701af60ce606633694ef67deacc))
* Change code length to 6 numbers ([#2894](https://github.com/ory/kratos/issues/2894)) ([56feb07](https://github.com/ory/kratos/commit/56feb079c3b99856c03cd8beb950673c10310520))
* **cli:** Helper for cleaning up stale records ([#2406](https://github.com/ory/kratos/issues/2406)) ([29d6376](https://github.com/ory/kratos/commit/29d6376e22e4de617ec63ca0a5dcb4dbf34c7c37)), closes [#952](https://github.com/ory/kratos/issues/952)
* Handler for update API with credentials ([#2423](https://github.com/ory/kratos/issues/2423)) ([561187d](https://github.com/ory/kratos/commit/561187dafe2fea324d55c4efe3ffa6b65f9bed72)), closes [#2334](https://github.com/ory/kratos/issues/2334)
* Immutable cookie session values ([#2761](https://github.com/ory/kratos/issues/2761)) ([a6f2793](https://github.com/ory/kratos/commit/a6f27935ce17a7ff5b3deaa4973d72a7d83454fb)), closes [#2701](https://github.com/ory/kratos/issues/2701)
* Implement blocking webhooks ([#1585](https://github.com/ory/kratos/issues/1585)) ([e48e9fa](https://github.com/ory/kratos/commit/e48e9fac7ab6a982e0e941bfea1d15569eb53582)), closes [#1724](https://github.com/ory/kratos/issues/1724) [#1483](https://github.com/ory/kratos/issues/1483)
* Improve cache handling ([6e8579b](https://github.com/ory/kratos/commit/6e8579b835d54d5ebb5371297ea60f24e915882d))
* Improve state generation logic ([546ee3d](https://github.com/ory/kratos/commit/546ee3dc900874bc0614923b10697388c4e7676b))
* Ingest hydra bugfix ([3c11216](https://github.com/ory/kratos/commit/3c112165e553161696cf746befb9e03c2e6e07fb))
* OAuth2 integration ([#2804](https://github.com/ory/kratos/issues/2804)) ([7c6eb2a](https://github.com/ory/kratos/commit/7c6eb2a5128c6bc76ac7306edafaa54c4893ea82)):

    This feature allows Ory Kratos to act as a login provider for Ory Hydra using the `oauth2_provider.url` configuration value.
    
    Closes https://github.com/ory/kratos/issues/273
    Closes https://github.com/ory/kratos/discussions/2293
    See https://github.com/ory/kratos-selfservice-ui-node/pull/50
    See https://github.com/ory/kratos-selfservice-ui-node/pull/68
    See https://github.com/ory/kratos-selfservice-ui-node/pull/108
    See https://github.com/ory/kratos-selfservice-ui-node/pull/111
    See https://github.com/ory/kratos-selfservice-ui-node/pull/149
    See https://github.com/ory/kratos-selfservice-ui-node/pull/170
    See https://github.com/ory/kratos-selfservice-ui-node/pull/198
    See https://github.com/ory/kratos-selfservice-ui-node/pull/207

* Parse all id token claims into raw_claims ([#2765](https://github.com/ory/kratos/issues/2765)) ([1da0cf6](https://github.com/ory/kratos/commit/1da0cf62b3f0ed8a81bca22123474baa7cf6de65)), closes [#2528](https://github.com/ory/kratos/issues/2528):

    All ID Token claims resulting from the Social Sign In flow are now available in `raw_claims` and can be used in the Social Sign In JsonNet Mapper.

* Replace magic links with one time codes in recovery flow ([#2645](https://github.com/ory/kratos/issues/2645)) ([a1532ba](https://github.com/ory/kratos/commit/a1532ba79722ccfc9c8608ef6f51a6d9ecb24a8e)), closes [#1451](https://github.com/ory/kratos/issues/1451):

    This feature introduces a new `code` strategy to recover an account. 
    
    Currently, if a user needs to initiate a recovery flow to recover a lost password/MFA/etc., they’ll receive an email containing a “magic link”. This link contains a flow_id and a recovery_token. This is problematic because some antivirus software opens links in emails to check for malicious content, etc.
    
    Instead of the magic link, we send an 8-digit code that is clearly displayed in the email or SMS. A user can now copy/paste or type it manually into the text-field that is shown after the user clicks “submit” on the initiate flow page.

* Replace message_ttl with static max retry count ([#2638](https://github.com/ory/kratos/issues/2638)) ([b341756](https://github.com/ory/kratos/commit/b341756130ee808ddcc003163884f09e3f006d0a)):

    This PR replaces the `courier.message_ttl` configuration option with a `courier.message_retries` option to limit how often the sending of a message is retried before it is marked as `abandoned`. 

* Standardize license headers ([#2790](https://github.com/ory/kratos/issues/2790)) ([8406eaf](https://github.com/ory/kratos/commit/8406eaf92006d9812108bd3ae57245f01e627bfc))
* Support ip exceptions ([de46c08](https://github.com/ory/kratos/commit/de46c08534dfae6165f6a570cc59829f367c0b57))
* Support md5 hash import ([#2725](https://github.com/ory/kratos/issues/2725)) ([d1b4e17](https://github.com/ory/kratos/commit/d1b4e1748f66c0dc8033235f1a9c155aac0d5caa))
* Trace WebHooks ([#2911](https://github.com/ory/kratos/issues/2911)) ([665605b](https://github.com/ory/kratos/commit/665605bbc4f6ca838f0180680cdd68905f07d482)):

    Previously the context was not propagated to the http client. As a result the (instrumented) client did not find the existing span and the sapns for outgoing http request have been orphains.
    
    With this simple Fix they are now children of the corresponding webhook spans.

* Update for the Ory Network ([#2814](https://github.com/ory/kratos/issues/2814)) ([3e09e58](https://github.com/ory/kratos/commit/3e09e58a695cf5d9d57b9f773e0f50b1fd794915))
* Upgrade hydra to v2 ([fdb108f](https://github.com/ory/kratos/commit/fdb108fe2542569202bfb39ef55e1a7e8c5b5ebf))

### Reverts

* Revert "autogen(openapi): regenerate swagger spec and internal client" ([24eddfb](https://github.com/ory/kratos/commit/24eddfb2adc67e22d34efdc6b6a6723c7be64237)):

    This reverts commit 4159b93ae3f8175cf7ccf77d34e4a7a2d0181d4f.


### Tests

* **e2e:** Add typescript ([37018c0](https://github.com/ory/kratos/commit/37018c0161d0affe88c9f2574d043f337579e4a9))
* **e2e:** Fix flaky assertions ([21a8487](https://github.com/ory/kratos/commit/21a8487f984168abbc7279c590c66822414c718e))
* **e2e:** Fix issuer config ([32454d2](https://github.com/ory/kratos/commit/32454d2fbd169a7839fc3d02786376ef4c7c986d))
* **e2e:** Fix webauthn regression ([26001e7](https://github.com/ory/kratos/commit/26001e7544b60ad0004153773a21c1d04abf9987))
* **e2e:** Improve webauthn test reliability ([4d323d0](https://github.com/ory/kratos/commit/4d323d01b53b9f7b0dc346211ac4fda0626d357a))
* **e2e:** Migrate to cypress 10.x ([317fab0](https://github.com/ory/kratos/commit/317fab0fe76a2762a77b3d2f8a75735598cb1c0e))
* **e2e:** Resolve flaky hydra configuration ([d8c82da](https://github.com/ory/kratos/commit/d8c82dabad4f04874647c48ecbf0eda91c7c90fa))
* **e2e:** Resolve max-age and issuer regression ([0ee4cf0](https://github.com/ory/kratos/commit/0ee4cf058cbda2bef52b3fa830f3db411f442197))
* **e2e:** Resolve max-age regression ([904f75d](https://github.com/ory/kratos/commit/904f75d254e9513aa3edad4fa3f9ead4d80e46df))
* **e2e:** Use correct dir ([907dbe3](https://github.com/ory/kratos/commit/907dbe3f605d5be5038ddc06029082b2df0914e2))
* Fix broken assertions ([e5f1311](https://github.com/ory/kratos/commit/e5f131138243ad5806c7927dd5a642d029cfad6c))
* Fix oidc test regression ([6c14b68](https://github.com/ory/kratos/commit/6c14b682d0984175495051308985281d72c0988e))
* Improve e2e tooling ([390ccaa](https://github.com/ory/kratos/commit/390ccaac18023979ff36bc7ee2df6c0d4a90d8c8))
* Parallelize and speed up config tests ([#2611](https://github.com/ory/kratos/issues/2611)) ([d8dea01](https://github.com/ory/kratos/commit/d8dea0138b09d4dff3c30aa14e0e99e423b355fe))
* Resolve builder regression ([934c30d](https://github.com/ory/kratos/commit/934c30d6064d1e7dfc59f4eef43d096e977c113e))
* Try and recover from allocated port error ([3b5ac5f](https://github.com/ory/kratos/commit/3b5ac5ff03b653191c1979fe1e4e9a4ea3ed7d36))
* Update snapshots ([#2877](https://github.com/ory/kratos/issues/2877)) ([cbaaceb](https://github.com/ory/kratos/commit/cbaaceb9ef73a91e1b4ce5e4f7b9d7bac04d4c03))

### Unclassified

* Revert "refactor: use gotemplates for command usage (#2770)" (#2778) ([d612612](https://github.com/ory/kratos/commit/d612612313dc26f1ddaaa84dbca65139b967d52c)), closes [#2770](https://github.com/ory/kratos/issues/2770) [#2778](https://github.com/ory/kratos/issues/2778):

    This reverts commit 1d22b235291ce7102dd186a53a431b55780973d3.

* Remove empty script (#2739) ([1515b83](https://github.com/ory/kratos/commit/1515b839f52044d6c9674d4a2df43dfeda3bb15b)), closes [#2739](https://github.com/ory/kratos/issues/2739)


# [0.10.1](https://github.com/ory/kratos/compare/v0.10.0...v0.10.1) (2022-06-01)

Re-release the SDK.





### Bug Fixes

* Bump ory cli ([12ceae0](https://github.com/ory/kratos/commit/12ceae005749c5dd01959720925418d643f13070))

### Code Generation

* Pin v0.10.1 release commit ([ab16580](https://github.com/ory/kratos/commit/ab16580b4326250885b920198b280456eb873a6b))


# [0.10.0](https://github.com/ory/kratos/compare/v0.9.0-alpha.3...v0.10.0) (2022-05-30)

We achieved a major milestone - Ory Kratos is out of alpha! Ory Kratos had no major changes in the APIs for the last months and feel confident that no large breaking changes will need to be introduced in the near future.

This release focuses on quality-of-live improvements, resolves several bugs, irons out developer experience issues, and introduces session renew capabilities!



## Breaking Changes

Please be aware that the SDK method signatures for `submitSelfServiceRecoveryFlow`, `submitSelfServiceRegistrationFlow`, `submitSelfServiceLoginFlow`, `submitSelfServiceSettingsFlow`, `submitSelfServiceVerificationFlow` might have changed in your SDK.

This patch moves several CLI command to comply with the Ory CLI command structure:

```patch
- ory identities get ...
+ ory get identity ...

- ory identities delete ...
+ ory delete identity ...

- ory identities import ...
+ ory import identity ...

- ory identities list ...
+ ory list identities ...

- ory identities validate ...
+ ory validate identity ...

- ory jsonnet format ...
+ ory format jsonnet ...

- ory jsonnet lint ...
+ ory lint jsonnet ...
```

This patch moves several CLI command to comply with the Ory CLI command structure:

```patch
- ory identities get ...
+ ory get identity ...

- ory identities delete ...
+ ory delete identity ...

- ory identities import ...
+ ory import identity ...

- ory identities list ...
+ ory list identities ...

- ory identities validate ...
+ ory validate identity ...

- ory jsonnet format ...
+ ory format jsonnet ...

- ory jsonnet lint ...
+ ory lint jsonnet ...
```



### Bug Fixes

* Add flow id when return_to is passed to the verification ([#2482](https://github.com/ory/kratos/issues/2482)) ([c2b1c23](https://github.com/ory/kratos/commit/c2b1c2303cd0587b9419d500f2e3d5f9c9c80ad4))
* Add indices for slow queries ([e0cdbc9](https://github.com/ory/kratos/commit/e0cdbc9ab3389de0f65b37758d86bea56d294d64))
* Add legacy session value ([ecfd052](https://github.com/ory/kratos/commit/ecfd05216f5ebb70f1617595d2d398cf1fa3c660)), closes [#2398](https://github.com/ory/kratos/issues/2398)
* **auth0:** Created_at workaround ([#2492](https://github.com/ory/kratos/issues/2492)) ([52a965d](https://github.com/ory/kratos/commit/52a965dc7e4ac868d21261cb44576846426bffa5)), closes [#2485](https://github.com/ory/kratos/issues/2485)
* Avoid excessive memory allocations in HIBP cache ([#2389](https://github.com/ory/kratos/issues/2389)) ([ee2d410](https://github.com/ory/kratos/commit/ee2d41057a7e6cb2c57c6304c2e7bbf5ad7c56da)), closes [#2354](https://github.com/ory/kratos/issues/2354)
* Change SQLite database mode to 0600 ([#2344](https://github.com/ory/kratos/issues/2344)) ([0e5d3b7](https://github.com/ory/kratos/commit/0e5d3b7726a8923fbc2a4c10ec18f0ba97ffbcff)):

    The default mode is 0644, which is allows broader access than necessary.

* Compile issues from merge conflict ([#2419](https://github.com/ory/kratos/issues/2419)) ([85a90c8](https://github.com/ory/kratos/commit/85a90c892d785b834cbdf8d029315550210444e2))
* Correct location ([b249aaa](https://github.com/ory/kratos/commit/b249aaad97eabc88c269265359a33cea920ef7f2))
* **courier:** Add ability to specify backoff ([#2349](https://github.com/ory/kratos/issues/2349)) ([bf970f3](https://github.com/ory/kratos/commit/bf970f32f571164b8081f09f602a3473e079194e))
* Do not expose debug in a response when a schema is not found ([#2348](https://github.com/ory/kratos/issues/2348)) ([aee2b1e](https://github.com/ory/kratos/commit/aee2b1ed1189b57fcbb1aaa456444d5121be94b1))
* Do not fail release if no changes needed ([114c93e](https://github.com/ory/kratos/commit/114c93eb48c242702b72d7785da70bd31d858214))
* **Dockerfile:** Use existing builder base image ([#2390](https://github.com/ory/kratos/issues/2390)) ([37de25a](https://github.com/ory/kratos/commit/37de25a541a24e03407ecf344fb750775e48c782))
* Embed schema ([b797bba](https://github.com/ory/kratos/commit/b797bba5910dfd925a11fb86e2dbd14b5dd839d9))
* Get user first name and last name from Apple ([#2331](https://github.com/ory/kratos/issues/2331)) ([4779909](https://github.com/ory/kratos/commit/47799098b35ea1cf5a1163f57d872a5bb2242d97))
* Improve error reporting from OpenAPI ([8a1009b](https://github.com/ory/kratos/commit/8a1009b16653df13485bab8e33926967c449bf4e))
* Improve performance of identity schema call ([af28de2](https://github.com/ory/kratos/commit/af28de267f21cd72953f3f353d8fd587937b2249))
* Internal Server Error on Empty PUT /identities/id body ([#2417](https://github.com/ory/kratos/issues/2417)) ([5a50231](https://github.com/ory/kratos/commit/5a50231b553aaa64bd90a3d2cd1be9d2e3aba9ac))
* Load return_to and append to errors ([#2333](https://github.com/ory/kratos/issues/2333)) ([5efe4a3](https://github.com/ory/kratos/commit/5efe4a33e35e74d248d4eec43dc901b7b6334037)), closes [#2275](https://github.com/ory/kratos/issues/2275) [#2279](https://github.com/ory/kratos/issues/2279) [#2285](https://github.com/ory/kratos/issues/2285)
* Make delete formattable ([0005f35](https://github.com/ory/kratos/commit/0005f357a049ecbf94d76a1e73434837753a04ea))
* Mark body as required ([#2479](https://github.com/ory/kratos/issues/2479)) ([c9ae117](https://github.com/ory/kratos/commit/c9ae1175340993cfc93db436c06462c80935ea2a))
* New issue templates ([b9ad684](https://github.com/ory/kratos/commit/b9ad684311ee8c654b2fa382010315e892581f5c))
* Openapi regression ([#2465](https://github.com/ory/kratos/issues/2465)) ([37a3369](https://github.com/ory/kratos/commit/37a3369cea8ed5af34e8324a291a7d7dba0eb43a))
* Quickstart docker-compose ([#2490](https://github.com/ory/kratos/issues/2490)) ([9717762](https://github.com/ory/kratos/commit/97177629c715028affbc294bdd432fd6c954d5ad)), closes [#2488](https://github.com/ory/kratos/issues/2488)
* Refresh is always false when session exists ([d3436d7](https://github.com/ory/kratos/commit/d3436d7fa17589d91e25c9f0bd66bc3bb5b150fa)), closes [#2341](https://github.com/ory/kratos/issues/2341)
* Remove required legacy field ([#2410](https://github.com/ory/kratos/issues/2410)) ([638d45c](https://github.com/ory/kratos/commit/638d45caf480b7287c9762cbf3c593217f40e3e8))
* Remove wrong templates ([4fe2d25](https://github.com/ory/kratos/commit/4fe2d25dd68033a8d7b3dd5f62d87b23a7ba361d))
* Reorder transactions ([78ca4c6](https://github.com/ory/kratos/commit/78ca4c6ca5a49b0800d9c34954638a926d80078b))
* Resolve index naming issues ([d5550b5](https://github.com/ory/kratos/commit/d5550b5ddc4e1677e4c4f808578f573760c6581e))
* Resolve MySQL index issues ([50bdba9](https://github.com/ory/kratos/commit/50bdba9f1117c60e80e153416bc997187b4a60b7))
* Resolve otelx panics ([6613a02](https://github.com/ory/kratos/commit/6613a02b8fd5f6f06e9b6301bdc39037771b3d9b))
* **sdk:** Improved OpenAPI specifications for UI nodes ([#2375](https://github.com/ory/kratos/issues/2375)) ([a42a0f7](https://github.com/ory/kratos/commit/a42a0f772af3625c457032d6dcc34289a62acc61)), closes [#2357](https://github.com/ory/kratos/issues/2357)
* Serve.admin.request_log.disable_for_health behaviour ([#2399](https://github.com/ory/kratos/issues/2399)) ([0a381fa](https://github.com/ory/kratos/commit/0a381fa3d702f77e614d0492dafa3ac2cd102c7e))
* **sql:** Add additional join argument to resolve MySQL query issue ([854e5cb](https://github.com/ory/kratos/commit/854e5cba80cad52b58571587980c00c038ff6596)), closes [#2262](https://github.com/ory/kratos/issues/2262)
* Unreliable HIBP caching strategy ([#2468](https://github.com/ory/kratos/issues/2468)) ([93bf1e2](https://github.com/ory/kratos/commit/93bf1e2cd53f3a4de3ff414017c17813d36b56da))
* Use `path` instead of `filepath` to join http route paths ([16b1244](https://github.com/ory/kratos/commit/16b12449c841bf7a237fe436b884b4b5012cd022)), closes [#2292](https://github.com/ory/kratos/issues/2292)
* Use JOIN instead of iterative queries ([0998cfb](https://github.com/ory/kratos/commit/0998cfb2fdda27ba8baeebcc603aae5fbe5c901f)), closes [#2402](https://github.com/ory/kratos/issues/2402)
* Use pointer of string for PasswordIdentifier in example code ([#2421](https://github.com/ory/kratos/issues/2421)) ([61f12e7](https://github.com/ory/kratos/commit/61f12e7579c7c337d0f415ac2b4029790c659c3d))
* Use predictable SQLite in memory DSNs ([#2415](https://github.com/ory/kratos/issues/2415)) ([51a13f7](https://github.com/ory/kratos/commit/51a13f712d38a942772b3f4c014971ecb4658d7a)), closes [#2059](https://github.com/ory/kratos/issues/2059)

### Code Generation

* Pin v0.10.0 release commit ([87e0de7](https://github.com/ory/kratos/commit/87e0de7a10b2a7478d8113ca028bfdb6525bc8e5))

### Code Refactoring

* Deprecate fizz renderer ([5277668](https://github.com/ory/kratos/commit/5277668b1324173df95db5e9e4b96ed841ff088b))
* Move CLI commands to match Ory CLI structure ([d11a9a9](https://github.com/ory/kratos/commit/d11a9a9dafdebb53ed9a8359496eb70b8adb99dd))
* Move CLI commands to match Ory CLI structure ([73910a3](https://github.com/ory/kratos/commit/73910a329b1ee46de2607c7ab1958ef2fb6de5f4))

### Documentation

* Add docs about change in default schema ([#2447](https://github.com/ory/kratos/issues/2447)) ([5093cd4](https://github.com/ory/kratos/commit/5093cd47f22311c2e1fdbffd82f0494806076f08))
* Remove notice importing credentials not possible ([#2418](https://github.com/ory/kratos/issues/2418)) ([b80ed69](https://github.com/ory/kratos/commit/b80ed6955518003ae6b7f647dffd2d49cc999fbc))

### Features

* Add certificate based authentication for smtp client ([#2351](https://github.com/ory/kratos/issues/2351)) ([7200037](https://github.com/ory/kratos/commit/72000375c028f5f7f9cb0d0b1b02f8aa09503e4f))
* Add ID to the recovery error when already logged in ([#2483](https://github.com/ory/kratos/issues/2483)) ([29e4a51](https://github.com/ory/kratos/commit/29e4a51cc5344dcb44839f8aa57197c41aeeb78d))
* Add localName to smtp config ([#2445](https://github.com/ory/kratos/issues/2445)) ([27336b6](https://github.com/ory/kratos/commit/27336b63b0c11c1667d5a07230bed82283475aa4)), closes [#2425](https://github.com/ory/kratos/issues/2425)
* Add render-schema script ([a0c006e](https://github.com/ory/kratos/commit/a0c006e40fb00608d682b74f44725883b9c7bf4f))
* Add session renew capabilities ([#2146](https://github.com/ory/kratos/issues/2146)) ([4348b86](https://github.com/ory/kratos/commit/4348b8640a282cd61fe30961faba5753e2af8bb0)), closes [#615](https://github.com/ory/kratos/issues/615)
* Add support for netID provider  ([#2394](https://github.com/ory/kratos/issues/2394)) ([ee7fc79](https://github.com/ory/kratos/commit/ee7fc79d49cd6d8f2985809585d1675c8e2ed376))
* Add tracing to persister ([391c54e](https://github.com/ory/kratos/commit/391c54eb3ba721e4912a7a4676acc2f630be2a72))
* **identity:** Add admin and public metadata fields ([562e340](https://github.com/ory/kratos/commit/562e340fe980e7c65ab3fc41f82a2a8899a33bfa)), closes [#2388](https://github.com/ory/kratos/issues/2388) [#47](https://github.com/ory/kratos/issues/47):

    This patch adds two new keys to identities, `metadata_public` and `metadata_admin` that can be used to store additional metadata about identities in Ory.

* Read subject id from https://graph.microsoft.com/v1.0/me for microsoft ([#2347](https://github.com/ory/kratos/issues/2347)) ([852f24f](https://github.com/ory/kratos/commit/852f24fb5cd8576f3f6d35017ce85e4fa1c51c95)):

    Adds the ability to read the OIDC subject ID from the `https://graph.microsoft.com/v1.0/me` endpoint. This introduces a new field `subject_source` to the OIDC configuration.
    
    Closes https://github.com/ory/kratos/pull/2153
    
    

* **sdk:** Add cookie headers to all form submissions ([#2467](https://github.com/ory/kratos/issues/2467)) ([9a969fd](https://github.com/ory/kratos/commit/9a969fd927ae8436a863e91ecb6574cb3bb1c3a6)), closes [#2003](https://github.com/ory/kratos/issues/2003) [#2454](https://github.com/ory/kratos/issues/2454)
* **sdk:** Add csrf cookie for login flow submission ([#2454](https://github.com/ory/kratos/issues/2454)) ([2bffee8](https://github.com/ory/kratos/commit/2bffee81f0e8a98851a3e11b4fc4969d95e9b445))
* Support argon2i password ([#2395](https://github.com/ory/kratos/issues/2395)) ([8fdadf9](https://github.com/ory/kratos/commit/8fdadf9d1724d28ae11996304703e06671549660))
* Switch to opentelemetry tracing ([#2318](https://github.com/ory/kratos/issues/2318)) ([121a4d3](https://github.com/ory/kratos/commit/121a4d3fc0f396e8da50ad1985cacf68a5c85a12))
* **tracing:** Improved tracing for requests ([#2475](https://github.com/ory/kratos/issues/2475)) ([b90a558](https://github.com/ory/kratos/commit/b90a5582284f1ceb0e97575e3b3562603b65ec5f))
* Upgrade to Go 1.18 ([725d202](https://github.com/ory/kratos/commit/725d202e6ae15b3b5c3282e03c03a40480a2e310))

### Tests

* Fix incorrect assertion ([b5b1361](https://github.com/ory/kratos/commit/b5b1361defa8faa6ea36d50a8d940c76f70c4ddd))
* Resolve regressions ([dd44593](https://github.com/ory/kratos/commit/dd44593a51a9277c717170360f9794837e4f910c))

### Unclassified

* BREAKING CHANGES: This patch group updates the tracing provider from OpenTracing to OpenTelemetry. Due to these changes, tracing providers Zipkin, DataDog, Elastic APM have been deactivated temporarily. The best way to re-add support for them is to make a pull request at https://github.com/ory/x/tree/master/otelx and check the status of https://github.com/ory/x/issues/499 ([7165fa0](https://github.com/ory/kratos/commit/7165fa04fa1c9442cad8da5c5814453e1ca0ba7b)):

    The configuration has not changed, and thus no changes to your system are required if you use Jaeger.



# [0.9.0-alpha.3](https://github.com/ory/kratos/compare/v0.9.0-alpha.2...v0.9.0-alpha.3) (2022-03-25)

Resolves an issue in the quickstart.



## Breaking Changes

Calling /self-service/recovery without flow ID or with an invalid flow ID while authenticated will now respond with an error instead of redirecting to the default page.

Closes https://github.com/ory-corp/cloud/issues/2173

Co-authored-by: aeneasr <3372410+aeneasr@users.noreply.github.com>



### Bug Fixes

* Accept recovery link from authenticated users ([#2195](https://github.com/ory/kratos/issues/2195)) ([0fa64dd](https://github.com/ory/kratos/commit/0fa64dd7fdaaadf92bddb600bbf201fb6e9d1fed)):

    When a recovery link is opened while the user already has a session cookie (possibly for another account), the endpoint will now correctly complete the recovery process and issue new cookies.

* Quickstart ([73b461c](https://github.com/ory/kratos/commit/73b461c6ea45e0feaab734d0eb0ce380993e95d4)):

    Closes https://github.com/ory/kratos/issues/2339

* Resolve issue where CF cookies would mingle with CSRF detection in API flows ([011219a](https://github.com/ory/kratos/commit/011219a40027d2c1b06c2797951a55e2f07c0845))
* Typo in error message ([#2332](https://github.com/ory/kratos/issues/2332)) ([b075a5b](https://github.com/ory/kratos/commit/b075a5b30b47e79af1330238a3b5ea97a3c2ac4b))
* Update v0.9.0-alpha.2 config schema path ([#2328](https://github.com/ory/kratos/issues/2328)) ([55705c7](https://github.com/ory/kratos/commit/55705c7ce0ff76dc7ddda24524db919dcb51225a))
* **version schema:** Require version or fall back to latest ([52c9824](https://github.com/ory/kratos/commit/52c98247d4c170f79fa25a019d7f4a73b3e5fdc4))

### Code Generation

* Pin v0.9.0-alpha.3 release commit ([32e36d4](https://github.com/ory/kratos/commit/32e36d4e75f888e69653625a52171200b4968a6c))

### Documentation

* Add missing error codes ([b854bb8](https://github.com/ory/kratos/commit/b854bb8a33794bba684abbfe5abc6b8da1c54f44))
* Clarify 410 error for api payloads ([2c7ac3b](https://github.com/ory/kratos/commit/2c7ac3b15a65e629ba25c0170fce68aa9eb3a80a))


# [0.9.0-alpha.2](https://github.com/ory/kratos/compare/v0.9.0-alpha.1...v0.9.0-alpha.2) (2022-03-22)

Resolves an issue in the SDK release pipeline.





### Bug Fixes

* Swag location ([5b51bfb](https://github.com/ory/kratos/commit/5b51bfbb10592c9e7dce14689f48530427c34edc))

### Code Generation

* Pin v0.9.0-alpha.2 release commit ([f5501cf](https://github.com/ory/kratos/commit/f5501cf575a74884555e0e1e4cba39c552f4868f))


# [0.9.0-alpha.1](https://github.com/ory/kratos/compare/v0.8.3-alpha.1.pre.0...v0.9.0-alpha.1) (2022-03-21)

Ory Kratos v0.9 is here! We're extremely happy to announce that the new release is out and once again it's been made even better thanks to the incredible contributions from our awesome community. <3

Enjoy!

Here's an overview of things you can expect from the v0.9 release:

1. We introduced 1:1 compatibility between self-hosting Ory Kratos and using Ory Cloud. The configuration works the same across all modes of operation and deployment!
2. Passwordless login with WebAuthn is now available! Authentication with YubiKeys, TouchID, FaceID, Microsoft Hello, and other WebAuthn-supported methods is now available. The refactored infrastructure lays a foundation for more passwordless flows to come.
3. All the docs are now available in a single repo. Go to the [ory/docs](https://github.com/ory/docs) repository to find docs for all Ory projects.
4. You can now load custom email templates that'll make your essential messaging like project invitations or password recovery emails look slick.
5. We've laid the foundation for adding SMS-dependant flows.
6. Security is always a top priority. We've made changes and updates such as CSP nonces, SSRF defenses, session invalidation hooks, and more.
7. Kratos now gracefully handles cookie errors.
8. Password policies are now configurable.
9. Added configuration to control the flow of webhooks. Now you can cancel flows & run them in the background.
10. You can import identities along with their credentials (password, social sign-in connections, WebAuthn, ...).
11. Infra: we migrated all of our CIs from CircleCI to GitHub Actions.
12. We moved the admin API from `/` to `admin`. **This is a breaking change**. Please read the explanation and proceed with caution!
13. Bugfix: fixed a bug in the handling of secrets. **This is a breaking change**. Please read the explanation and proceed with caution!
14. Bugfix: several bugs in different self-service flows are no more.

As you can see, this release introduces breaking changes. We tried to keep the HTTP API as backward-compatible as possible by introducing HTTP redirects and other measures, but this update requires you to take extra care. Make sure you've read the release notes and understand the risk before updating.

You must apply SQL migrations for this release. **Make sure to create backup before you start!**



## Breaking Changes

Configuration key `selfservice.whitelisted_return_urls` has been renamed to `allowed_return_urls`.

All endpoints at the Admin API are now exposed at `/admin/`. For example, endpoint `https://kratos:4434/identities` is now exposed at `https://kratos:4434/admin/identities`. This change makes it easier to configure reverse proxies and API Gateways. Additionally, it introduces 1:1 compatibility between Ory Cloud's APIs and self-hosted Ory Kratos. Please note that nothing has changed in terms of the port. To make the migration less painful, we have set up redirects from the old endpoints to the new `/admin` endpoints, so your APIs, SDKs, and clients should continue working as they were working before. This change is marked as a breaking change as it touches many endpoints and might be confusing when encountering the redirect for the first time.

If you are using two or more secrets for the `secrets.session`, this patch might break existing Ory Session Cookies. This has the effect that users will need to re-authenticate when visiting your app.

The `password_identifier` form field of the password login strategy has been renamed to `identifier` to make compatibility with passwordless flows possible. Field name `password_identifier` will still be accepted. Please note that the UI node for displaying the "username" / "email" field has this `name="identifier"` going forward. Additionally, the `traits` of the password strategy are no longer within group `password` but instead in group `profile` going forward!

The following OpenID Connect configuration keys have been renamed to better explain their purpose:

```patch
- private_key_id
+ apple_private_key_id

- private_key
+ apple_private_key

- team_id
+ apple_team_id

- tenant
+ microsoft_tenant
```

A major issue has been lingering in the configuration for a while. What happens to your identities when you update a schema? The answer was, it depends on the change. If the change is incompatible, some things might break!

To resolve this problem we changed the way you define schemas. Instead of having a global `default_schema_url` which developers used to update their schema, you now need to define the `default_schema_id` which must reference schema ID in your config. To update your existing configuration, check out the patch example below:

```patch
identity:
-  default_schema_url: file://stub/identity.schema.json
+  default_schema_id: default
+  schemas:
+  - id: default
+    url: file://stub/identity.schema.json
```

Ideally, you would version your schema and update the `default_schema_id` with every change to the new version:

```yaml
identity:
  default_schema_id: user_v1
  schemas:
    - id: user_v0
      url: file://path/to/user_v0.json
    - id: user_v1
      url: file://path/to/user_v1.json
```



### Bug Fixes

* Add CourierConfig to default registry ([#2243](https://github.com/ory/kratos/issues/2243)) ([2e1fba3](https://github.com/ory/kratos/commit/2e1fba3ca88e273362978fe29197fe44a879813e))
* Add DispatchMessage to interface ([df2ca7a](https://github.com/ory/kratos/commit/df2ca7a7c97a28d40c6a8af082f99ff7706ee9db))
* Add missing enum ([#2223](https://github.com/ory/kratos/issues/2223)) ([4b7d7d0](https://github.com/ory/kratos/commit/4b7d7d0011207614ab12f52bb3a911b62581ebe9)):

    Closes https://github.com/ory/sdk/issues/147

* Add output-dir input to cli-next ([#2230](https://github.com/ory/kratos/issues/2230)) ([1eb3f18](https://github.com/ory/kratos/commit/1eb3f189f29cc032c44cbd9803acbf99362e5a62))
* Added malformed config test ([5a3c9c1](https://github.com/ory/kratos/commit/5a3c9c162bd1da5c7bb938192a5e82789bac52cc))
* Appropriately pass context around ([#2241](https://github.com/ory/kratos/issues/2241)) ([668f6b2](https://github.com/ory/kratos/commit/668f6b246db1f61b9800f7581bedba4fa25318c4)):

    Closes https://github.com/ory/cloud/issues/56

* Base redirect URL decoding ([acdefa7](https://github.com/ory/kratos/commit/acdefa7464825e5307132eab5cd2752e1841c3de))
* Base64 encode identity schema URLs ([ad44e4d](https://github.com/ory/kratos/commit/ad44e4d5f2cea86a95cc376c94fb5f5ac5bc1b82)):

    Previously, identity schema IDs with special characters could lead to broken URLs. This patch introduces a change where identity schema IDs are base64 encoded to address this issue. Schema IDs that are not base64 encoded will continue working.

* Broken links API spec ([e1e7516](https://github.com/ory/kratos/commit/e1e75165785f48f5a154c899e1c4168bcbb7d8c3))
* Cloud config issue ([135b29c](https://github.com/ory/kratos/commit/135b29c647c87569cc85e8a72babb8d6777ebd24))
* Correct recovery hook ([c7682a8](https://github.com/ory/kratos/commit/c7682a8fd97fdac87d59d3e7fb798384b018c40f))
* **courier:** Improve composability ([d47150e](https://github.com/ory/kratos/commit/d47150e8440a03ce34d6085fb693bddf2c02620b))
* Do not error when HIBP behaves unexpectedly ([#2251](https://github.com/ory/kratos/issues/2251)) ([a431c1e](https://github.com/ory/kratos/commit/a431c1e1976f740bedb2fec4ce88b7d1b832e42c)), closes [#2145](https://github.com/ory/kratos/issues/2145)
* Do not remove all credentials when remove all security keys ([#2233](https://github.com/ory/kratos/issues/2233)) ([ecd715a](https://github.com/ory/kratos/commit/ecd715a0437c0b068aa0c6a17cd2ba53fe034354))
* Don't inherit flow type in recovery and verification flows ([#2250](https://github.com/ory/kratos/issues/2250)) ([c5b444a](https://github.com/ory/kratos/commit/c5b444aa2bf46b3a86d08f693ab200a30bd4a609)), closes [#2049](https://github.com/ory/kratos/issues/2049)
* **embed:** Disallow additional props ([b2018ce](https://github.com/ory/kratos/commit/b2018ce3b1667fffc9d0a2c4c82cfafed7f3cac5))
* **embed:** Do not require plaintext/html in email config ([dfe4140](https://github.com/ory/kratos/commit/dfe4140dda44d4b64988b94272b4776e362abde5))
* Ensure no internal networks can be called in SMS sender ([65e42e5](https://github.com/ory/kratos/commit/65e42e5cb3a9a3a81e3c623fa066a7651dfb0699))
* **identity:** Slow query performance on MySQL ([731b3c7](https://github.com/ory/kratos/commit/731b3c7ba48271e2fb6bbd53b0281d5269012332)), closes [#2278](https://github.com/ory/kratos/issues/2278)
* Improve password error resilience on settings flow ([e614f6e](https://github.com/ory/kratos/commit/e614f6e94e1d0f66f48bd058b015ab467d6b1b07))
* Improve soundness of credential identifier normalization ([e475163](https://github.com/ory/kratos/commit/e475163330d06ca02cd0419e4b7216f03218e8c5))
* Incorrect makefile rule ([#2222](https://github.com/ory/kratos/issues/2222)) ([83a0ce7](https://github.com/ory/kratos/commit/83a0ce7d20e59c2fb1a35fa071a3d11a9280bcad))
* **login:** Put passwordless login before password ([df9245f](https://github.com/ory/kratos/commit/df9245fbc403e1b8f2dd1378678963cc0d71ef1a))
* **lookup:** Resolve credentials counting regression ([50782c6](https://github.com/ory/kratos/commit/50782c68c77ce1c0d8c092678a6710e0be6fa18d))
* Lower-case jsonnet context for sms ([8c58e94](https://github.com/ory/kratos/commit/8c58e94707122a9b50873ca1acaa32659b5b8416))
* Mark struct as used ([33f3dfe](https://github.com/ory/kratos/commit/33f3dfeba5af3808f34b16241d74993ceed788be))
* Mark width and height as required ([#2322](https://github.com/ory/kratos/issues/2322)) ([37f2f22](https://github.com/ory/kratos/commit/37f2f220ce699e031018777c9976cafa22faa984)):

    Closes https://github.com/ory/sdk/issues/157

* Move to new post-release steps ([#2206](https://github.com/ory/kratos/issues/2206)) ([10778fd](https://github.com/ory/kratos/commit/10778fdd16a116b5dc8f4c2bdc96a895728d9aec))
* Mr comment fix ([96c917e](https://github.com/ory/kratos/commit/96c917e3c1b02b13be55056bfd94b517007fc206))
* **oidc:** Improve empty credential handling ([124d4ce](https://github.com/ory/kratos/commit/124d4ce9fe949dcea4fd5ff8e45530835d38cb3c))
* **oidc:** Incorrect error handling ([c8d789c](https://github.com/ory/kratos/commit/c8d789c10e2be11dfc8c3eea01a339637f89ea63))
* Order regression ([2cb5d2b](https://github.com/ory/kratos/commit/2cb5d2bf2d645a0e63cf289c966ee8557edbf333))
* Pass context to registration flow ([c8d55b3](https://github.com/ory/kratos/commit/c8d55b339647cdca3c9beace760dc3a9beac31c1))
* Pass docs output dir as a separate argument ([78c69a2](https://github.com/ory/kratos/commit/78c69a2790c957bf8102260150d69b1844899ed9))
* Pass token to render-version-schema ([#2246](https://github.com/ory/kratos/issues/2246)) ([4d117e5](https://github.com/ory/kratos/commit/4d117e51abef739d686e48dede63a030a753be41))
* **password:** Schema regressions ([271d5fa](https://github.com/ory/kratos/commit/271d5fa93f96721d7bf8aa841c700dfec1de4104))
* Properly check for not found ([77ac199](https://github.com/ory/kratos/commit/77ac199f00f04eb7fd40db6fb546921271026e20))
* Properly pass context ([#2300](https://github.com/ory/kratos/issues/2300)) ([fab8a93](https://github.com/ory/kratos/commit/fab8a939c97e61c028143e37e2a78d3edd569da0))
* Provide access to root path and error page ([#2317](https://github.com/ory/kratos/issues/2317)) ([f360ee8](https://github.com/ory/kratos/commit/f360ee8e65dc64983181746d1059eac53588e029))
* Rebase regressions ([d1c5085](https://github.com/ory/kratos/commit/d1c508570032c620a654b896111215a76a811517))
* **registration:** Order for passwordless webauthn ([8427322](https://github.com/ory/kratos/commit/8427322b31fb5206a55e9f62823745fcc6983a22))
* Remove non-hermetic sprig functions ([#2201](https://github.com/ory/kratos/issues/2201)) ([17e0acc](https://github.com/ory/kratos/commit/17e0acc527cfbb703d9d44b776138da23b217ca4)):

    Closes https://github.com/ory/kratos/issues/2087

* Resolve issues with the CI pipeline ([d15bd90](https://github.com/ory/kratos/commit/d15bd90433ed191c2eb41f119ed288906827334e))
* Resolve merge regression ([d8ca4f3](https://github.com/ory/kratos/commit/d8ca4f327499f94c811c55237f210288fb6a9dd5))
* Resolve prettier issues ([32bf052](https://github.com/ory/kratos/commit/32bf052f0084860623ea815ed913e94261c89070))
* Resolve remaining passwordless regressions ([151c8cf](https://github.com/ory/kratos/commit/151c8cfb53402aaf2518a471579c25c3785b13d2))
* Resovle lint errors ([afb7aaf](https://github.com/ory/kratos/commit/afb7aaf7b019756a624e7f1b2e35fd575882570a))
* Return 400 instead of 404 on admin recovery ([ae2509c](https://github.com/ory/kratos/commit/ae2509cf7a95f940d33945271ac1fe8fc255506b)), closes [#1664](https://github.com/ory/kratos/issues/1664)
* **sdk:** Add all available discriminators ([5d70f9c](https://github.com/ory/kratos/commit/5d70f9c70a39067c2d6c0b1f127ff28ca39e77a9)), closes [#2287](https://github.com/ory/kratos/issues/2287) [#2288](https://github.com/ory/kratos/issues/2288)
* **sdk:** Add webauth and lookup_secret to identityCredentialsType ([#2276](https://github.com/ory/kratos/issues/2276)) ([61ce3c0](https://github.com/ory/kratos/commit/61ce3c0c35366f587bfee5c89496fa15432bb241))
* **sdk:** Correct minimum page to 1 ([a28362e](https://github.com/ory/kratos/commit/a28362e054cf12441ed25d8927cd63e3264bfed6)), closes [#2286](https://github.com/ory/kratos/issues/2286)
* **selfservice:** Cannot login after remove security keys and all other 2FA settings ([#2181](https://github.com/ory/kratos/issues/2181)) ([5ff6773](https://github.com/ory/kratos/commit/5ff6773ab8512bdfb8d2c7b650970711cbb012ba)), closes [#2180](https://github.com/ory/kratos/issues/2180)
* **selfservice:** Login self service flow with TOTP does not pass on return_to URL ([#2175](https://github.com/ory/kratos/issues/2175)) ([3eaa88e](https://github.com/ory/kratos/commit/3eaa88e74e1540b14b6e41df2881346c60b92046)), closes [#2172](https://github.com/ory/kratos/issues/2172)
* **session:** Correctly calculate aal for passwordless webauthn ([c7eb970](https://github.com/ory/kratos/commit/c7eb970ed252577e06d3d769d2545d5e8e98175a))
* **session:** Properly declare session secrets ([6312afd](https://github.com/ory/kratos/commit/6312afd2eb0d1dc808d600a902eb1e16b07fd9cb)), closes [#2272](https://github.com/ory/kratos/issues/2272):

    Previously, a misconfiguration of Gorilla's session store caused incorrect handling of the configured secrets. From now on, cookies will also be properly encrypted at all times.

* Snapshot regression ([6481441](https://github.com/ory/kratos/commit/6481441fe7df1a2fc43ff153697e9bd2160c49b3))
* Static analysis ([a1d3254](https://github.com/ory/kratos/commit/a1d3254346ec0bcc0a8c42bf66a8171e027f0d97))
* **test:** Parallelization issues ([dbcf3fb](https://github.com/ory/kratos/commit/dbcf3fb616db64e1b1f4cb5066113f703ca0b2ee))
* **text:** Incorrect IDs for different messages ([0833321](https://github.com/ory/kratos/commit/0833321e04e9865046294b051376bed415a41441)), closes [#2277](https://github.com/ory/kratos/issues/2277)
* **totp:** Resolve credentials counting regression ([737bb3f](https://github.com/ory/kratos/commit/737bb3f71e91f7c735231d0131072aca4f5622ea))
* Typo ([fbc8b4f](https://github.com/ory/kratos/commit/fbc8b4f9901e7761bef9a7f74a483cb077007cf8))
* Typo ([3bb0d41](https://github.com/ory/kratos/commit/3bb0d41e3696be90cfc12f1bf00a546536e283b6))
* Unstable ordering ([bee26c6](https://github.com/ory/kratos/commit/bee26c65c9511af82b9ed2051ab4f45b9570602d))
* Unstable webauthn order ([6262160](https://github.com/ory/kratos/commit/626216098fcd9411c1b4b7cb3b42784146b29924))
* Updated oathkeeper+kratos example ([#2273](https://github.com/ory/kratos/issues/2273)) ([567a3d7](https://github.com/ory/kratos/commit/567a3d765aa2115951f6af5b4ed4d2c791231de0))
* URL with hash sign in after_verification_return_to stays encoded ([#2173](https://github.com/ory/kratos/issues/2173)) ([fb1cb8a](https://github.com/ory/kratos/commit/fb1cb8a993cbf6cb050d7dce91672b05efd53224)), closes [#2068](https://github.com/ory/kratos/issues/2068)
* Use actions/checkout for ui repos ([f0136ca](https://github.com/ory/kratos/commit/f0136cac639862bf50933063b7dc38973739139b))
* Use correct dir for clidoc ([8c8a1ab](https://github.com/ory/kratos/commit/8c8a1ab7b41fa026189cec8d1f77e2e89c696d11))
* Use HTTP 303 instead of 302 for selfservice redirects ([#2215](https://github.com/ory/kratos/issues/2215)) ([50b6bd8](https://github.com/ory/kratos/commit/50b6bd892ae6efba34773811ef488f15fc95154f)), closes [#1969](https://github.com/ory/kratos/issues/1969)
* Use latest hydra version ([ffb3f20](https://github.com/ory/kratos/commit/ffb3f20e67d357160c024f5e58ebf63a9aec41ff))
* **webauthn:** Resolve missing identifier bug ([93a1ae4](https://github.com/ory/kratos/commit/93a1ae4fe98487a0bca00d2afdc5e7b07c0e1c46))
* **webauthn:** Schema regressions ([970e861](https://github.com/ory/kratos/commit/970e861714ec01c5cfe19545871798d9ad0ae70c))
* **webauth:** SPA regressions for login ([be378ff](https://github.com/ory/kratos/commit/be378ffa5ddbd56a00b471dce861ec074eed5192))
* Yq version ([41b6f18](https://github.com/ory/kratos/commit/41b6f1879f23866c070100dd1767f841bff3a815))

### Code Generation

* Pin v0.9.0-alpha.1 release commit ([72bd2ed](https://github.com/ory/kratos/commit/72bd2ed67559a64415b2686e8f67c42df888e49e))

### Code Refactoring

* All admin endpoints are now exposed under `/admin/` on the admin port ([8acb4cf](https://github.com/ory/kratos/commit/8acb4cfaa61ef52619e889b8c862191c6b92e5eb))
* Distinguish between first and multi factor credentials ([8de9d01](https://github.com/ory/kratos/commit/8de9d01d9edae485f5a6ea7c68584ba4019a24d6))
* Identity.default_schema_url is now `identity.default_schema_id` ([#1964](https://github.com/ory/kratos/issues/1964)) ([e4f205d](https://github.com/ory/kratos/commit/e4f205d69bec07a71bf1d34d97ab3a6b99a4cc46))
* **identity:** Move credentials counter ([c9875a7](https://github.com/ory/kratos/commit/c9875a7582accc740061e6a19d7b4b0998899f3f))
* Mimic credentials config on import ([c3eb7ce](https://github.com/ory/kratos/commit/c3eb7ce60597954a60b8903ac011a643d0facf12))
* Move credential configs for oidc and password ([50ac851](https://github.com/ory/kratos/commit/50ac851cc4534aa474a76c208f15483548ec8631))
* Move docs to ory/docs ([57151da](https://github.com/ory/kratos/commit/57151da6adc85753d54c108637298642ccbc8347))
* **oidc:** Credentials counting ([b75a639](https://github.com/ory/kratos/commit/b75a6390de85e10db8e9e17a74e95dd6dd716442))
* **password:** DRY up registration helpers ([8a51839](https://github.com/ory/kratos/commit/8a51839ba85ddb5a345fef65f30b4325103ce38a))
* **password:** Internals and deprecated fields ([a7784bd](https://github.com/ory/kratos/commit/a7784bdb52aff0ac171e59b2301755b65c842813))
* Rename `password_identifier` field to `identifier` ([4dbe0ea](https://github.com/ory/kratos/commit/4dbe0ea41f49e198840292fc101258a4bdca826e))
* Rename `whitelisted_return_urls` to `allowed_return_urls` ([#2299](https://github.com/ory/kratos/issues/2299)) ([686c9ba](https://github.com/ory/kratos/commit/686c9ba08ff1db8a310eaed5c4b3aec69e0f84da))
* **session:** Aal computation ([a136de9](https://github.com/ory/kratos/commit/a136de99a0f8fe78ee344f2243359c781b166378))
* Update apple and microsoft config key names ([#2261](https://github.com/ory/kratos/issues/2261)) ([6da2370](https://github.com/ory/kratos/commit/6da2370b4e6833ef61ca03214261e45c4786cb44)), closes [#1979](https://github.com/ory/kratos/issues/1979)

### Documentation

* Add debug tip ([#2186](https://github.com/ory/kratos/issues/2186)) ([a1ada22](https://github.com/ory/kratos/commit/a1ada2255d132b1f3ea8cb494620b9c17b42f161))
* Add react example code ([#2185](https://github.com/ory/kratos/issues/2185)) ([0689cc7](https://github.com/ory/kratos/commit/0689cc73ccc9a472c5610f1e011c6ccbc5e0c20d))
* Cloud ([8d1d65d](https://github.com/ory/kratos/commit/8d1d65d9d12a894bd25c82394e0392e228fe383d))
* Fix broken links ([d88c56f](https://github.com/ory/kratos/commit/d88c56fc0ebf042d1270d04a2382784e5200654d))
* Fix broken links API doc ([#2296](https://github.com/ory/kratos/issues/2296)) ([47eaae5](https://github.com/ory/kratos/commit/47eaae575023469834c0c3a4aac64dc6d880e164))
* Fix versions ([7186ff3](https://github.com/ory/kratos/commit/7186ff354b9c3d0fbd3fb809546075fcfcd0c57f))
* Replace all mentions of Ory Kratos SDK with Ory SDK ([#2187](https://github.com/ory/kratos/issues/2187)) ([4e6897f](https://github.com/ory/kratos/commit/4e6897ff2220b5668d784a16dd1f48db30f271f0))
* Update readme ([e7d9da1](https://github.com/ory/kratos/commit/e7d9da199825fb15ae720c0496a257590b353a26))

### Features

* Abandon courier messages after configurable timeout ([#2257](https://github.com/ory/kratos/issues/2257)) ([bff92f7](https://github.com/ory/kratos/commit/bff92f73b3f12d2dffa2061eb0e51e746eba2185))
* Add `webauthn` to list of identifiers ([1a8b256](https://github.com/ory/kratos/commit/1a8b256cca33aa9cbb143e7e8fc1efc8217e9b8a)):

    This patch adds the key `webauthn` to the list of possible identifiers in the Identity JSON Schema. Use this key to specify what field is used to find the WebAuthn credentials on passwordless login flows.

* Add credential migrator pattern ([77afc6f](https://github.com/ory/kratos/commit/77afc6f8ea868eaba7853adfcb9ed159b44ecbc8))
* Add message for missing webauthn credentials ([303dc6b](https://github.com/ory/kratos/commit/303dc6bc33c20cd619d2542180247bd7b7f02092))
* Add new messages ([09e6fd1](https://github.com/ory/kratos/commit/09e6fd16bb6be0ff3ee209bbfe69e967546f70da))
* Add npm install step ([3d253e5](https://github.com/ory/kratos/commit/3d253e58ec7d4464d9749efe6ecc4a5c1d9be789))
* Add versioning and improve compatibility for credential migrations ([78ce668](https://github.com/ory/kratos/commit/78ce668a38c914939028be42cd30eefa566ed09a))
* Added sms sending support to courier ([687eca2](https://github.com/ory/kratos/commit/687eca24aac7a7b89cc949693271343573107898))
* Allow empty version string ([419f94b](https://github.com/ory/kratos/commit/419f94bc1065771e49982faf56f8ef90a30bc306))
* Cancelable web hooks ([44a5323](https://github.com/ory/kratos/commit/44a5323f835860dccd11460d666f620026e8b58d)):

    Introduces the ability to cancel web hooks by calling `error "cancel"` in JsonNet.

* **config:** Add option to mark webauthn as passwordless-able ([0455e3f](https://github.com/ory/kratos/commit/0455e3fe901cff6ff314fd59a35864886672327c)):

    Adds option `passwordless` to `selfservice.methods.webauthn.config`, making it possible to use WebAuthn for first-factor authentication, or so-called "passwordless" authentication.

* Courier template configs ([#2156](https://github.com/ory/kratos/issues/2156)) ([799b6a8](https://github.com/ory/kratos/commit/799b6a81add747d3001a1758e08ee7b4c6463d64)), closes [#2054](https://github.com/ory/kratos/issues/2054):

    It is now possible to override individual courier email templates using the configuration system!

* **courier:** Expose setters again ([598dc3a](https://github.com/ory/kratos/commit/598dc3a4d7c27838e9058382378972a1c0330bde))
* **e2e:** Add passwordless flows and fix bugs ([ef3871b](https://github.com/ory/kratos/commit/ef3871bd9b3e7e5f4360da8d1b7749cc005b4e19))
* **identity:** Add identity credentials helpers ([b7be327](https://github.com/ory/kratos/commit/b7be327a370368932ff390968acffaa1ce6d55a0))
* **identity:** Add versioning to credentials ([aaf779a](https://github.com/ory/kratos/commit/aaf779ac1c29b24ece6d5f3d7892a3bf08277653))
* Ignore web hook response ([ae87914](https://github.com/ory/kratos/commit/ae87914512025c05d814a1200eda66d8f931ce44)):

    Introduces the ability to ignore responses from web hooks in favor of faster and non-blocking execution.

* Make sensitive log value redaction text configurable ([#2321](https://github.com/ory/kratos/issues/2321)) ([9b66e43](https://github.com/ory/kratos/commit/9b66e437d0aeed61643b76aea7d49cad001dc8cf))
* **oidc:** Customizable base redirect uri ([fa1f234](https://github.com/ory/kratos/commit/fa1f23469f2fecfa82fa38147f601d969bd9aaa4)):

    Closes https://github.com/ory-corp/cloud/issues/2003

* Password, social sign, verified email in import ([41a27b1](https://github.com/ory/kratos/commit/41a27b1e15e090d3e99cdcfc3c1ba8eac76097a4)), closes [#605](https://github.com/ory/kratos/issues/605):

    This patch introduces the ability to import passwords (cleartext, PKBDF2, Argon2, BCrypt) and Social Sign In connections when creating identities!

* **recovery:** Allow invalidation of existing sessions ([5029884](https://github.com/ory/kratos/commit/502988474e2bce46752f7fc7885bc1b91423bbdd)), closes [#1077](https://github.com/ory/kratos/issues/1077):

    You can now use the `revoke_active_sessions` hook in the recovery flow. It invalidates all of an identity's sessions on successful account recovery.

* **schema:** Add functionality to disallow internal HTTP requests ([6e08416](https://github.com/ory/kratos/commit/6e08416235bd821493df4d9cda2e8bd76d507871)):

    See https://github.com/ory-corp/cloud/issues/1261

* **security:** Add e2e tests for various private network SSRF defenses ([b049bc3](https://github.com/ory/kratos/commit/b049bc304cd79568ee82f1423e583949f63d3377))
* **security:** Add SSRF defenses in OIDC ([d37dc5d](https://github.com/ory/kratos/commit/d37dc5d7946252783463bc9e99f7f792e2735614))
* **session:** Add webauthn to extension validation ([049fd8e](https://github.com/ory/kratos/commit/049fd8edc382f344018398027a4e0b3915116ff2))
* **session:** Webauthn can now be a first factor as well ([861bee0](https://github.com/ory/kratos/commit/861bee0f029e3bb3f6b7218be19eaf6c26562b76))
* Trace web hook calls ([#2154](https://github.com/ory/kratos/issues/2154)) ([98ee300](https://github.com/ory/kratos/commit/98ee300e065c6e81e6128a509af3f48612cda88a))
* **webauthn:** Add error preventing deleting last webauthn credential ([1209eda](https://github.com/ory/kratos/commit/1209edacaf1b7dea32bd1bd124c86910bc2553c6))
* **webauthn:** Add new decoder schemas ([c3e1501](https://github.com/ory/kratos/commit/c3e1501bf5170416a034130eb68d1db456a47239))
* **webauthn:** Add passwordless credentials indicator ([6e3057a](https://github.com/ory/kratos/commit/6e3057a96a34d22cac193e5c17b4a3c01d2ca045))
* **webauthn:** Add swagger type ([14c2b74](https://github.com/ory/kratos/commit/14c2b745e951a185dee600f6f2e8f93788c67285))
* **webauthn:** Count passwordless credentials ([145af23](https://github.com/ory/kratos/commit/145af23aef8f5c9ffdcec47bac5758da709d4646))
* **webauthn:** Implement refresh using webauth ([bf10868](https://github.com/ory/kratos/commit/bf108688ed146211da3cc2ec4bf0df015e535220)), closes [#2284](https://github.com/ory/kratos/issues/2284):

    This change introduces the ability to refresh a session (for example when entering "sudo" mode") using WebAuthn credentials. In this case, it does not matter whether the WebAuthN credentials are for MFA or passwordless flows.

* **webauthn:** Improve schema ([790dcf3](https://github.com/ory/kratos/commit/790dcf3a7079d57a088d399c03d040af1019a3aa))
* **webauthn:** Manage webauthn passwordless keys ([5a62ced](https://github.com/ory/kratos/commit/5a62ced175248a85b1e843b4017757aa86d62d23))
* **webauthn:** Passwordless login ([b4c4fd2](https://github.com/ory/kratos/commit/b4c4fd2c25ae5d55350ce573df8295fe6d8c42a1))
* **webauthn:** Update messages and nodes ([22534d8](https://github.com/ory/kratos/commit/22534d8253384f2002033a5b2bbdcf573779a49c))
* **webauthn:** Use plain bytes for wrapped user ([97c8c9e](https://github.com/ory/kratos/commit/97c8c9e25234847622f1ab508cd5d50758d323c0))

### Tests

* Add data for new migration ([b0488ef](https://github.com/ory/kratos/commit/b0488efa600024f40b2c019fa0f492dd39c8bfa9))
* Add tests for new sms options ([799fa10](https://github.com/ory/kratos/commit/799fa106cd0fed33afbe76903911df9292d49bf6))
* **cmd:** Fix regressions ([4b92be9](https://github.com/ory/kratos/commit/4b92be9325d02e605e12d96c7990774234ed1d1d))
* **driver:** Fix regressions ([c6f5137](https://github.com/ory/kratos/commit/c6f51377f253275bf7321c67a5e949699ac12adb))
* **e2e:** Add import tests ([ed90f39](https://github.com/ory/kratos/commit/ed90f394d32ee0a3e42c3a9c1c066f94a05d02c1))
* **e2e:** Reenable hydra ([055a491](https://github.com/ory/kratos/commit/055a4912d3e7712d4bc3a3f5cf9c68d1834998dc))
* **e2e:** Resolve privileged regression ([f7dd5ab](https://github.com/ory/kratos/commit/f7dd5aba26b43aa9f60d8429a7d256f48f228578))
* **e2e:** Resolve regression ([b5053c9](https://github.com/ory/kratos/commit/b5053c902331ae166824eb92b89295e693bf0dc7))
* **e2e:** Resolve regressions ([da154c5](https://github.com/ory/kratos/commit/da154c5e549f79ca5703209852981ded07281f43))
* **e2e:** Resolve regressions ([d46d435](https://github.com/ory/kratos/commit/d46d435c40c383bbd844af8fead283ee46a137fb))
* **e2e:** Resolve regressions and flakes ([a607385](https://github.com/ory/kratos/commit/a60738510875f770f9dbb0b3449dbcf2d473ada3))
* **e2e:** Wait for initial network requests ([#2242](https://github.com/ory/kratos/issues/2242)) ([c5a04b5](https://github.com/ory/kratos/commit/c5a04b5f174e06faca99ebc7461c8ebe8e1f694d))
* Extract common registration helpers to library ([5c1f11b](https://github.com/ory/kratos/commit/5c1f11b2ae65dd73d572e456b522a7d83ac1f473))
* Fix concurrent database access ([46f6fb7](https://github.com/ory/kratos/commit/46f6fb7d246b384e561bdf8952185855f25cce56))
* Fix regression ([f96e48f](https://github.com/ory/kratos/commit/f96e48fa6d4d8b341bcd3f52228b7abff8b934fb))
* **identity:** Ensure migrations run when fetching identities ([322d467](https://github.com/ory/kratos/commit/322d467ac11dcdf4e3210f947b80029c77662065))
* **identity:** Fix regressions ([f492f0e](https://github.com/ory/kratos/commit/f492f0e1d112813d926eac48b5ad5d2e1857a382))
* Re-enable MySQL ([cbe8f6e](https://github.com/ory/kratos/commit/cbe8f6ea4fe48fe84a5cbc8915754f83e7eff428))
* Remove obsolete test ([cd644ae](https://github.com/ory/kratos/commit/cd644aef9175fe21024c37a381722503fcd88555))
* Remove obsolete test failure ([f8fd480](https://github.com/ory/kratos/commit/f8fd48041404344636c51b63d55a668209bed0e0))
* Remove only ([87b3bce](https://github.com/ory/kratos/commit/87b3bce3433601dd918f76c0bc2d25ea4af6e482))
* Remove unnecessary test ([2fa33e4](https://github.com/ory/kratos/commit/2fa33e4f28759b5dc5de78e00e42ed8cc4ccce89))
* Resolve potential panic ([d44af28](https://github.com/ory/kratos/commit/d44af289e9c09a981e80b6f69d22a5cce6b1dbfa))
* **schema:** Resolve regressions ([c6d0810](https://github.com/ory/kratos/commit/c6d08105a270fafd21a14a19e412d7081dedc754))
* Significantly reduce persister run time ([647d6ef](https://github.com/ory/kratos/commit/647d6ef73797462020c2f59ece15e645561182b0))
* Update fixtures ([21462b7](https://github.com/ory/kratos/commit/21462b7eb8cbac719d8ae531969b0fd9d42b5e0c))
* Update fixtures ([299c6e3](https://github.com/ory/kratos/commit/299c6e3be7c120bb769a4b2572ebe42c5ab3ddb1))
* **webauthn:** Add passwordless profile ([88199ea](https://github.com/ory/kratos/commit/88199ea28e8b3460ccc585e5fd1713d398cae15c))
* **webauthn:** Passwordless registration ([c9b6280](https://github.com/ory/kratos/commit/c9b6280720c2fd08191994c86e85ceb1f52a27d2))

### Unclassified

* Move login hinting to own package ([1eb2604](https://github.com/ory/kratos/commit/1eb260423491af917edb1256d260ca3d3fb198dc))


# [0.8.3-alpha.1.pre.0](https://github.com/ory/kratos/compare/v0.8.2-alpha.1...v0.8.3-alpha.1.pre.0) (2022-01-21)

autogen: pin v0.8.3-alpha.1.pre.0 release commit



## Breaking Changes

This patch removes the ability to use domain aliases, an obscure feature rarely used that had several issues and inconsistencies.



### Bug Fixes

* Add `identity_id` index to `identity_verifiable_addresses` table ([#2147](https://github.com/ory/kratos/issues/2147)) ([86fd942](https://github.com/ory/kratos/commit/86fd942e9a80e36dd65ef4ac57c5a5546f94995a)):

    The verifiable addresses are loaded eagerly into the identity. When that happens, the `identity_verifiable_addresses` table is queried by `nid` and `identity_id`. This index should greatly improve performance, especially of the `/sessions/whoami` endpoint.

* Add ability to resume continuity sessions from several cookies ([#2131](https://github.com/ory/kratos/issues/2131)) ([8b87bdb](https://github.com/ory/kratos/commit/8b87bdb1967654b5fbfbf9799948485b2a9a6af0)), closes [#2016](https://github.com/ory/kratos/issues/2016) [#1786](https://github.com/ory/kratos/issues/1786)
* Add hiring notice to README ([#2074](https://github.com/ory/kratos/issues/2074)) ([0c1e816](https://github.com/ory/kratos/commit/0c1e816693ad4a6c3fdb7206bbc95c81cdfdf3c0))
* Add missing version tag in quickstart.yml ([#2110](https://github.com/ory/kratos/issues/2110)) ([1d281ea](https://github.com/ory/kratos/commit/1d281ea69e551cc3d40415f5405690f445891bb6))
* Adjust scan configuration ([#2140](https://github.com/ory/kratos/issues/2140)) ([8506fcf](https://github.com/ory/kratos/commit/8506fcf59d572851b24041b48af6a04b31520a32)), closes [#2083](https://github.com/ory/kratos/issues/2083)
* Admin endpoint `/schemas` not redirecting to public endpoint ([#2133](https://github.com/ory/kratos/issues/2133)) ([413833f](https://github.com/ory/kratos/commit/413833f128c0674f4e8dbb9e73698a9df04cfc1a)), closes [#2084](https://github.com/ory/kratos/issues/2084)
* Choose correct CSRF cookie when multiple are set ([633076b](https://github.com/ory/kratos/commit/633076be008104afd50186ebe60722ef21999d5d)), closes [ory/kratos#2121](https://github.com/ory/kratos/issues/2121) [ory-corp/cloud#1786](https://github.com/ory-corp/cloud/issues/1786):

    Resolves an issue where, when multiple CSRF cookies are set, a random one would be used to verify the CSRF token. Now, regardless of how many conflicting CSRF cookies exist, if one of them is valid, the request will pass and clean up the cookie store.

* **continuity:** Properly reset cookies that became invalid ([8e4b4fb](https://github.com/ory/kratos/commit/8e4b4fb3d6dbe668cf0166f4cff49eae753d481c)), closes [#2121](https://github.com/ory/kratos/issues/2121) [ory-corp/cloud#1786](https://github.com/ory-corp/cloud/issues/1786):

    Resolves several reports related to incorrect handling of invalid continuity issues.

* **continuity:** Remove cookie on any error ([428ac03](https://github.com/ory/kratos/commit/428ac03b582184dbbbc0c9c3ffd399273fd8e1a5))
* Do not send session after registration without hook ([#2094](https://github.com/ory/kratos/issues/2094)) ([3044229](https://github.com/ory/kratos/commit/3044229227229e81a4ba770eec241a748dd0945c)), closes [#2093](https://github.com/ory/kratos/issues/2093)
* Docker-compose standalone definition ([3c7065a](https://github.com/ory/kratos/commit/3c7065ad32ff314c8cbdad8ed89fd9a9f5928f72))
* Explain mitigations in cookie error messages ([ef4b01a](https://github.com/ory/kratos/commit/ef4b01a80ea91114b182ff26759d98cd5ba2cd02))
* Expose network wrapper ([a570607](https://github.com/ory/kratos/commit/a570607d460e7c5f9d49ce38ba7a4e06ae172359))
* Faq ([#2101](https://github.com/ory/kratos/issues/2101)) ([311f906](https://github.com/ory/kratos/commit/311f9066a524308b970afc81d98d1a14b78bf63d)):

    This patch 
    - moves the FAQ to the Debug & Help section
    - renames it to Tips & Troubleshooting
    - moves many of the questions to documents where they fit better, reformatted and with added information where needed.
    - also some other spelling/format fixes
    
    See also https://github.com/ory/docusaurus-template/pull/87

* Ignore whitespace around identifier with password strategy ([#2160](https://github.com/ory/kratos/issues/2160)) ([45335c5](https://github.com/ory/kratos/commit/45335c50f719af504974fe54e504d7653db03c78)), closes [#2158](https://github.com/ory/kratos/issues/2158)
* Improve courier test signature ([b8888e3](https://github.com/ory/kratos/commit/b8888e3c93a602635b396503b7301396ce740ff8))
* Include missing type string in config schema ([#2142](https://github.com/ory/kratos/issues/2142)) ([ec2c88a](https://github.com/ory/kratos/commit/ec2c88ac2d65ea1db1146101519cdbb709ebdbbb)):

    Inside the config.schema.json under the CORS setting, add the missing type (string) for the items of the allowed_origins array

* **login:** Error handling when failed to prepare for an expired flow ([#2120](https://github.com/ory/kratos/issues/2120)) ([fdad834](https://github.com/ory/kratos/commit/fdad834e7577e298887b83b693ddf20632cd7c43))
* Minor fixes in FAQ update ([#2130](https://github.com/ory/kratos/issues/2130)) ([b53eec7](https://github.com/ory/kratos/commit/b53eec721489514a80719b73bc5c758dc2adedfd))
* Quickstart standalone service definition ([#2149](https://github.com/ory/kratos/issues/2149)) ([872b06e](https://github.com/ory/kratos/commit/872b06e1f798deacfef101edc3ab33fd75af9b29))
* Resolve configx regression ([672c0ff](https://github.com/ory/kratos/commit/672c0ffc7f5edd1fd238dcdd0c5d0430b30966c6))
* **selfservice:** Recovery self service flow passes on return_to URL ([#1920](https://github.com/ory/kratos/issues/1920)) ([b925d35](https://github.com/ory/kratos/commit/b925d351dd0ce48cb6aed046dcf2698796453751)), closes [#914](https://github.com/ory/kratos/issues/914)
* Send 404 instead of null response for unknown verification flows ([#2102](https://github.com/ory/kratos/issues/2102)) ([c9490c8](https://github.com/ory/kratos/commit/c9490c8927209b686aafe54b8a16207a8ef47ebe)), closes [#2099](https://github.com/ory/kratos/issues/2099):

    Fixes the verification handler to write the error, instead of nil object, when the flow does not exist. Adds tests for every handler to check proper behavior in that regard.

* Support setting complex configs from the environment ([c45bf83](https://github.com/ory/kratos/commit/c45bf83a9e6744a0b3f2f24e3b07a6f0131d9a40)):

    Closes https://github.com/ory/kratos/issues/1535
    Closes https://github.com/ory/kratos/issues/1792
    Closes https://github.com/ory/kratos/issues/1801

* Update download urls according to the new names ([#2078](https://github.com/ory/kratos/issues/2078)) ([86ae016](https://github.com/ory/kratos/commit/86ae0166c8893b809929c7c45a2ba84416ddf228))

### Code Generation

* Pin v0.8.3-alpha.1.pre.0 release commit ([b1f1da2](https://github.com/ory/kratos/commit/b1f1da2c0b4fbf6e6b4259c58b39a3e88e990142))

### Code Refactoring

* Deprecate domain aliases ([894a2cc](https://github.com/ory/kratos/commit/894a2cc39671fbc9d2c13b1fc1b45b217da5145d))

### Documentation

* Fix incorrect port ([c9a3587](https://github.com/ory/kratos/commit/c9a358717a99af436c6802f45c9c1f6edc77585f)), closes [#2095](https://github.com/ory/kratos/issues/2095)
* Fix link ([c245ed4](https://github.com/ory/kratos/commit/c245ed40d443e3068bc5eee902e6b14f6ae777c6)):

    Closes https://github.com/ory/kratos-selfservice-ui-node/issues/164

* Ory cloud mentions + spelling ([#2100](https://github.com/ory/kratos/issues/2100)) ([0c2fa5b](https://github.com/ory/kratos/commit/0c2fa5bdb98b95877ef740297b6d96a931a3430f))
* Pagination ([#2143](https://github.com/ory/kratos/issues/2143)) ([0807a03](https://github.com/ory/kratos/commit/0807a03fba8ff9a3123cd038a472e90895502e82)), closes [#2039](https://github.com/ory/kratos/issues/2039)
* Typo ([#2073](https://github.com/ory/kratos/issues/2073)) ([e1a54f9](https://github.com/ory/kratos/commit/e1a54f9129d41b34cc8864c8ac38d1448e1f9372))
* Typo ([#2114](https://github.com/ory/kratos/issues/2114)) ([a7a16d7](https://github.com/ory/kratos/commit/a7a16d7c91d89e274ea5fd79787cd4671d825532))
* Update docker guide ([072ca4d](https://github.com/ory/kratos/commit/072ca4d990cf4060555c8b2626f39ff18172d064)), closes [#2086](https://github.com/ory/kratos/issues/2086)
* Upgrade guide ([#2132](https://github.com/ory/kratos/issues/2132)) ([4a4ab05](https://github.com/ory/kratos/commit/4a4ab05573ebb20f82f62bfd38767de68d7708e9)):

    Closes https://github.com/ory/kratos/discussions/2104


### Features

* Add preset CSP nonce ([#2096](https://github.com/ory/kratos/issues/2096)) ([8913292](https://github.com/ory/kratos/commit/8913292c1193c416e5a54997e3635bef87affc01)):

    Closes https://github.com/ory/kratos-selfservice-ui-node/issues/162

* Added phone number identifier ([#1938](https://github.com/ory/kratos/issues/1938)) ([294dfa8](https://github.com/ory/kratos/commit/294dfa85b4552b9266c44bb3376b8610c1ff5521)), closes [#137](https://github.com/ory/kratos/issues/137)
* Allow registration to be disabled ([#2081](https://github.com/ory/kratos/issues/2081)) ([864b00d](https://github.com/ory/kratos/commit/864b00d6ecddefdb06ac22fda04670bfa43f2fd5)), closes [#882](https://github.com/ory/kratos/issues/882)
* Courier templates fs support ([#2164](https://github.com/ory/kratos/issues/2164)) ([13689a7](https://github.com/ory/kratos/commit/13689a7135311a05b17383486f5fdab2e7a412d0))
* **courier:** Override default link base URL ([cc99096](https://github.com/ory/kratos/commit/cc99096d07408c8b713ef9a7b17b8345597a9129)):

    Added a new configuration value `selfservice.methods.link.config.base_url` which allows to change the default base URL of recovery and verification links. This is useful when the email should send a link which does not match the globally configured base URL.
    
    See https://github.com/ory-corp/cloud/issues/1766

* **docker:** Add jaeger ([27ec2b7](https://github.com/ory/kratos/commit/27ec2b74ee42697102c6a9a79bc5ca3c09756d94))
* Enable Buildkit ([#2079](https://github.com/ory/kratos/issues/2079)) ([f40df5c](https://github.com/ory/kratos/commit/f40df5cd932aa3185b2155368db51a49b7f05991)):

    Looks like this was attempted before but the magic comment was not on the first line.

* Expose courier template load ([#2082](https://github.com/ory/kratos/issues/2082)) ([790716e](https://github.com/ory/kratos/commit/790716e58a4be06f04f3cbc5b974f16d873ae0d8))
* Generalise courier tests ([#2125](https://github.com/ory/kratos/issues/2125)) ([75c6053](https://github.com/ory/kratos/commit/75c60537e366760fe87b7b8978e9854873b7f702))
* Make the password policy more configurable ([#2118](https://github.com/ory/kratos/issues/2118)) ([70c627b](https://github.com/ory/kratos/commit/70c627b9feb3ec55765070b7c6c3fd64f2640e59)), closes [#970](https://github.com/ory/kratos/issues/970)
* **security:** Add option to disallow private IP ranges in webhooks ([05f1e5a](https://github.com/ory/kratos/commit/05f1e5a99426ed54cb70514554e64d851f0ba8d6)), closes [#2152](https://github.com/ory/kratos/issues/2152)
* Selfservice and administrative session management ([#2011](https://github.com/ory/kratos/issues/2011)) ([0fe4155](https://github.com/ory/kratos/commit/0fe4155b878102b77f7f13de5f0754ff75961498)), closes [#655](https://github.com/ory/kratos/issues/655) [#2007](https://github.com/ory/kratos/issues/2007)

### Tests

* Update cypress ([#2090](https://github.com/ory/kratos/issues/2090)) ([883a1b1](https://github.com/ory/kratos/commit/883a1b1ea33a1d3ef8b33342328382b59e4f18c3))


# [0.8.2-alpha.1](https://github.com/ory/kratos/compare/v0.8.1-alpha.1...v0.8.2-alpha.1) (2021-12-17)

This release addresses further important security updates in the base Docker Images. We also resolved all issues related to ARM support on both Linux and macOS and fixed a bug that prevent the binary from compiling on FreeBSD.

This release also makes use of our new build architecture which means that the Docker Images names have changed. We removed the "scratch" images as we received frequent complaints about them. Additionally,
all Docker Images have now, per default, SQLite support built-in. If you are relying on the SQLite images, update your Docker Pull commands as follows:

```patch
- docker pull oryd/kratos:{version}-sqlite
+ docker pull oryd/kratos:{version}
```

Additionally, all passwords now have to be at least 8 characters long, following recommendations from Microsoft and others.

In v0.8.1-alpha.1 we failed to include all the exciting things that landed, so we'll cover them now!

1. Advanced E-Mail templating support with sprig - makes it possible to translate emails as well!
2. Support wildcards for allowing redirection targets.
3. Account Recovery initiated by the Admin API now works even if identities have no email address.

Enjoy this release!





### Bug Fixes

* Add missing sample app paths to oathkeeper config ([#2058](https://github.com/ory/kratos/issues/2058)) ([a527db4](https://github.com/ory/kratos/commit/a527db4487c4efd2e96f8bf84d48a3cca30a14a1)):

    Add "welcome,registration,login,verification" and "**.png" to the paths oathkeeper forwards to self service ui.

* Add section on webauthn constraints ([#2072](https://github.com/ory/kratos/issues/2072)) ([23663b5](https://github.com/ory/kratos/commit/23663b50afce59cec2cfcaa4d3f50ae0abcf6310))
* After release hooks ([56c2e61](https://github.com/ory/kratos/commit/56c2e61195b6e6808ed76b9fd5dee0da1f489ce9))
* Dockerfile clean up ([52420cc](https://github.com/ory/kratos/commit/52420ccc17a8d395f0b13c0ad03ac334434c4b0e)), closes [#2070](https://github.com/ory/kratos/issues/2070)
* Goreleaser after hook ([c763f2b](https://github.com/ory/kratos/commit/c763f2b394543a142f35b022d9c9d154c8e8489c))
* Goreleaser config ([7099af2](https://github.com/ory/kratos/commit/7099af20929ad003968e7fc9e47a4fe745984fbb)):

    See https://github.com/goreleaser/goreleaser/issues/2762

* Release hook ([90bd769](https://github.com/ory/kratos/commit/90bd7698380168b88ee301d9f343054052b208fd))

### Code Generation

* Pin v0.8.2-alpha.1 release commit ([627f4a1](https://github.com/ory/kratos/commit/627f4a1ddb378db84510a85013c4580a9d8024ad))

### Documentation

* Fix bodged release ([032b23a](https://github.com/ory/kratos/commit/032b23aba3fa04e5e2a638b78b806ca49a6a8e1c))
* Quickstart update ([#2060](https://github.com/ory/kratos/issues/2060)) ([3387cf6](https://github.com/ory/kratos/commit/3387cf6f111db5944fbff536fd0a9a67bc388f9a)), closes [#2032](https://github.com/ory/kratos/issues/2032) [#1916](https://github.com/ory/kratos/issues/1916)


# [0.8.1-alpha.1](https://github.com/ory/kratos/compare/v0.8.0-alpha.4.pre.0...v0.8.1-alpha.1) (2021-12-13)

This maintenance release important security updates for the base Docker Images (e.g. Alpine). Additionally, several hiccups with the new ARM support have been resolved and the binaries are now downloadable for all major platforms. Please note that passwords now have to be at least 8 characters long, following recommendations from Microsoft and others.

Enjoy this release!





### Bug Fixes

* Bodget docs commit ([f9d2f82](https://github.com/ory/kratos/commit/f9d2f8245bc94aaf21ddc9e5516b64e7887dae4b))
* Build docs on release ([2cf137a](https://github.com/ory/kratos/commit/2cf137a0540b81f4e405920cafd251db71d2f9fa))
* De-duplicate message IDs ([#1973](https://github.com/ory/kratos/issues/1973)) ([9d8e197](https://github.com/ory/kratos/commit/9d8e19720fcc2e5b5371c2ddea4e2501304a93fd))
* Docs links ([#2008](https://github.com/ory/kratos/issues/2008)) ([8515e17](https://github.com/ory/kratos/commit/8515e17938570770ca4cbf93028782925e28f431))
* Require minimum length of 8 characters password ([#2009](https://github.com/ory/kratos/issues/2009)) ([bb5846e](https://github.com/ory/kratos/commit/bb5846ecb446b9e58b2a4949c678fddac4bbac4f)):

    Kratos follows [NIST Digital Identity Guidelines - 5.1.1.2 Memorized Secret Verifiers](https://pages.nist.gov/800-63-3/sp800-63b.html) and [password policy](https://www.ory.sh/kratos/docs/concepts/security#password-policy) says
    
    > Passwords must have a minimum length of 8 characters and all characters (unicode, ASCII) must be allowed.
    
    
    

* Resolve freebsd build issue ([#2004](https://github.com/ory/kratos/issues/2004)) ([9c75fe9](https://github.com/ory/kratos/commit/9c75fe9e7ab4ff27f8d1f2399a58baaadefaaa0d)), closes [#1645](https://github.com/ory/kratos/issues/1645)
* Revert tag ([f1d7b9e](https://github.com/ory/kratos/commit/f1d7b9e2db2cab4acdcaacbae06a85c42417b334)), closes [#1945](https://github.com/ory/kratos/issues/1945)
* Set dockerfile ([c860b99](https://github.com/ory/kratos/commit/c860b992aee6a63d9696377ed9047e8cdeef0098))
* Skip docs publishing for pre releases ([eb6d8cd](https://github.com/ory/kratos/commit/eb6d8cdb2d3d400eb3b9398a15825ecdb10d3cf8))
* Support complex lifespans ([#2050](https://github.com/ory/kratos/issues/2050)) ([0edbebe](https://github.com/ory/kratos/commit/0edbebed896e79fd2979a54756932ea27c2ddb99))
* Update docs after release ([850be90](https://github.com/ory/kratos/commit/850be9065b64bcf268b42e4018f60b25a7a73da5))
* Verification error code ([#1967](https://github.com/ory/kratos/issues/1967)) ([44411ab](https://github.com/ory/kratos/commit/44411ab4ac5f184c7f42e6ece0ccb2ae7cbdc42c)), closes [#1956](https://github.com/ory/kratos/issues/1956)

### Code Generation

* Pin v0.8.1-alpha.1 release commit ([8247416](https://github.com/ory/kratos/commit/82474161f61a3a22afad478838ffe8fe837d41ac))

### Documentation

* Add `Content-Type` to recommended CORS allowed headers ([#2015](https://github.com/ory/kratos/issues/2015)) ([dd890ab](https://github.com/ory/kratos/commit/dd890ab96727d7a2c8c2f52279dc3516096213f0))
* **debug:** Fix typo ([#1976](https://github.com/ory/kratos/issues/1976)) ([0647554](https://github.com/ory/kratos/commit/0647554179d7b0119ed01d353cd0ea9eb8317752))
* Fix incorrect tag ([bbd2355](https://github.com/ory/kratos/commit/bbd2355bbb220389021b596eec339a25652d932a)), closes [#2032](https://github.com/ory/kratos/issues/2032) [#2028](https://github.com/ory/kratos/issues/2028)
* Fixed date format example ([#2038](https://github.com/ory/kratos/issues/2038)) ([fc4703a](https://github.com/ory/kratos/commit/fc4703aa34066a56fa3cf3b664a0d032157e477a))
* Improve text around bcrypt ([#2037](https://github.com/ory/kratos/issues/2037)) ([ba6981e](https://github.com/ory/kratos/commit/ba6981e344e880936b5e995c433dae85659ba780))
* Levenshtein-Distance has been released ([#2040](https://github.com/ory/kratos/issues/2040)) ([393b6b3](https://github.com/ory/kratos/commit/393b6b38cdc4758e838eec20e81d486662f7b4a7))
* Minor fixes ([#2010](https://github.com/ory/kratos/issues/2010)) ([12918db](https://github.com/ory/kratos/commit/12918dbf4b0edb2857e06736aee9cccf1a5f76ff))
* Password-strength meter has been dropped ([#2041](https://github.com/ory/kratos/issues/2041)) ([9848fb3](https://github.com/ory/kratos/commit/9848fb3b40c12799eafc73d2ec0f410bf5b22aa8))
* This has been done ([#2045](https://github.com/ory/kratos/issues/2045)) ([7e8c91a](https://github.com/ory/kratos/commit/7e8c91ace5229fdc394461b3453acb3f01da0a6c))
* Totp unlink image in 2fa docs ([#1957](https://github.com/ory/kratos/issues/1957)) ([7afb731](https://github.com/ory/kratos/commit/7afb731c15ebbd6bab54a133f2e80e938dd937d4))
* Update email template docs ([#1960](https://github.com/ory/kratos/issues/1960)) ([#1968](https://github.com/ory/kratos/issues/1968)) ([b0f25a9](https://github.com/ory/kratos/commit/b0f25a9a6013f1e450163f5c08b221d328c210be))
* Webhooks have landed ([#2035](https://github.com/ory/kratos/issues/2035)) ([80e53eb](https://github.com/ory/kratos/commit/80e53eb83d0dc84d2082ee343bfcecd2bfd99e13))

### Features

* Add alpine dockerfile ([587eaee](https://github.com/ory/kratos/commit/587eaeee60cab2f539af8f309800f5a6e9cdfe6f))
* Add x-total-count to paginated pages ([b633ec3](https://github.com/ory/kratos/commit/b633ec3da6ccca196cd9d78c3c43d9797bd8d982))
* Buildkit with multi stage build ([#2025](https://github.com/ory/kratos/issues/2025)) ([57ab7f7](https://github.com/ory/kratos/commit/57ab7f784674c2cef2b1cef4b6922e9834213e3d))
* **cmd:** Add OIDC credential include ([#2017](https://github.com/ory/kratos/issues/2017)) ([1482844](https://github.com/ory/kratos/commit/148284485db8a86aa10c5aefb34373f9a8c7d95a)):

    With this change, the `kratos identities get` CLI can additionally fetch OIDC credentials.
    
    

* Generalise courier ([#2019](https://github.com/ory/kratos/issues/2019)) ([1762a73](https://github.com/ory/kratos/commit/1762a730886707be3549bc6789f65c66d755e1d0))
* **oidc:** Add spotify provider ([#2024](https://github.com/ory/kratos/issues/2024)) ([0064e35](https://github.com/ory/kratos/commit/0064e350ccb417fefee6f48ca5895f3d75247bb3))

### Tests

* Add web hook test cases ([#2051](https://github.com/ory/kratos/issues/2051)) ([316e940](https://github.com/ory/kratos/commit/316e940a70684084c857e80a2ffaf334a64aee94))
* **e2e:** Split e2e script into setup and test phase ([#2027](https://github.com/ory/kratos/issues/2027)) ([1761418](https://github.com/ory/kratos/commit/176141860f3aa946519073d0e35bf3acacd6c685))
* Fix changed message ID ([#2013](https://github.com/ory/kratos/issues/2013)) ([0bb66de](https://github.com/ory/kratos/commit/0bb66de582ebcb501c161655ae00e276a1d7d5d2))


# [0.8.0-alpha.4.pre.0](https://github.com/ory/kratos/compare/v0.8.0-alpha.3...v0.8.0-alpha.4.pre.0) (2021-11-09)

autogen: pin v0.8.0-alpha.4.pre.0 release commit



## Breaking Changes

To celebrate this change, we cleaned up the ways you install Ory software, and will roll this out to all other projects soon:

There is now one central brew / bash curl repository:

```patch
-brew install ory/kratos/kratos
+brew install ory/tap/kratos

-bash <(curl https://raw.githubusercontent.com/ory/kratos/master/install.sh)
+bash <(curl https://raw.githubusercontent.com/ory/meta/master/install.sh) kratos
```



### Bug Fixes

* Add base64 to ReadSchema ([#1918](https://github.com/ory/kratos/issues/1918)) ([8c8815b](https://github.com/ory/kratos/commit/8c8815b7ced0051eb0120198ae75b8fcf0fce2ba)), closes [#1529](https://github.com/ory/kratos/issues/1529)
* Add error.id to invalid cookie/token settings flow ([#1919](https://github.com/ory/kratos/issues/1919)) ([73610d4](https://github.com/ory/kratos/commit/73610d4cfb16789385d2660e278419664b1ea3f3)), closes [#1888](https://github.com/ory/kratos/issues/1888)
* Adds missing webauthn authentication method ([#1914](https://github.com/ory/kratos/issues/1914)) ([44892f3](https://github.com/ory/kratos/commit/44892f379c1aa9ffd7f5c92c9c1b32cc34a0dada))
* Allow use of relative URLs in config ([#1754](https://github.com/ory/kratos/issues/1754)) ([5f73bb0](https://github.com/ory/kratos/commit/5f73bb0784aeb7c4f3b1ed949926f9d9aed968d1)), closes [#1446](https://github.com/ory/kratos/issues/1446)
* Do not use csrf for meta endpoints ([#1927](https://github.com/ory/kratos/issues/1927)) ([fd14798](https://github.com/ory/kratos/commit/fd147989a55357248a37a30548c5d4c104bcf0f7))
* E2e test regression ([#1937](https://github.com/ory/kratos/issues/1937)) ([c9be009](https://github.com/ory/kratos/commit/c9be009112b03291ea76dd4de0911f495cf1e1ac))
* Include text label for link email field ([07a1dbb](https://github.com/ory/kratos/commit/07a1dbb95156ca50116219dc837ca61e3d597df1)), closes [#1909](https://github.com/ory/kratos/issues/1909)
* Panic on webhook with nil body ([#1890](https://github.com/ory/kratos/issues/1890)) ([4bf1825](https://github.com/ory/kratos/commit/4bf18250373b7255e26e95d51a257e5280ad3148)), closes [#1885](https://github.com/ory/kratos/issues/1885)
* Paths ([8c852c7](https://github.com/ory/kratos/commit/8c852c73136e130d163e2c9c5e0ca8a3449f4e26))
* Speed up git clone ([d3e4bde](https://github.com/ory/kratos/commit/d3e4bdefd252131b6a1b84917962ff07284e3f9f))
* Update sdk orb ([94e12e6](https://github.com/ory/kratos/commit/94e12e6d767ffa46d9060fdfb463adb83806990b))
* Use bcrypt for password hashing in example ([a9196f2](https://github.com/ory/kratos/commit/a9196f27791c30d32743e6b69a86595d76362f29))
* Use new ory installation method ([09cfc7e](https://github.com/ory/kratos/commit/09cfc7e2c23885270ef02193b4fdddc5550f3c23))

### Code Generation

* Pin v0.8.0-alpha.4.pre.0 release commit ([3e443b7](https://github.com/ory/kratos/commit/3e443b77ef63d72e5bf0b806790c86841a140afc))

### Documentation

* Add subdomain configuration in csrf page ([#1896](https://github.com/ory/kratos/issues/1896)) ([681750f](https://github.com/ory/kratos/commit/681750f92d7fe517e7cc184cb4b65e6a21903ee9)):

    Add some instructions as to how kratos can be configured to work across subdomains.

* Remove unintended characters in subdomain section in csrf page ([#1897](https://github.com/ory/kratos/issues/1897)) ([dfb9007](https://github.com/ory/kratos/commit/dfb900797fc98ca7900631ccf8018858c4e43e85))

### Features

* Add new goreleaser build chain ([#1932](https://github.com/ory/kratos/issues/1932)) ([cf1714d](https://github.com/ory/kratos/commit/cf1714dafaa0cda98640c772106620586dae7763)):

    This patch adds full compatibility with ARM architectures, including Apple Silicon (M1). We additionally added cryptographically signed signatures verifiable using [cosign](https://github.com/sigstore/cosign) for both binaries as well as docker images.

* Add quickstart mimicking hosted ui ([813fb4c](https://github.com/ory/kratos/commit/813fb4cf48df1154ea334cca751cb55f7b3c77eb))
* Advanced e-mail templating support ([#1859](https://github.com/ory/kratos/issues/1859)) ([54b97b4](https://github.com/ory/kratos/commit/54b97b45506eff9cfafe338842ddf818b0c81f62)), closes [#834](https://github.com/ory/kratos/issues/834) [#925](https://github.com/ory/kratos/issues/925)
* Allow wildcard domains for redirect_to checks ([#1528](https://github.com/ory/kratos/issues/1528)) ([349cdcf](https://github.com/ory/kratos/commit/349cdcf4b1298d9e544344705ecd8e7b5eada48c)), closes [#943](https://github.com/ory/kratos/issues/943):

    Support wildcard domains in redirect_to checks.

* Configurable health endpoints access logging ([#1934](https://github.com/ory/kratos/issues/1934)) ([1301f68](https://github.com/ory/kratos/commit/1301f689bb0f1f44b66a057c8915f77ac71f30cc)):

    This PR introduces a new boolean configuration parameter that allows turning off logging of health endpoints requests in the access log. The implementation is basically a rip-off from Ory Hydra and the configuration parameter is the same:
    
    ```
    serve.public.request_log.disable_for_health
    serve.admin.request_log.disable_for_health
    ```
    
    The default value is _false_.
    
    

* Integrate sbom generation to goreleaser ([#1850](https://github.com/ory/kratos/issues/1850)) ([305bb28](https://github.com/ory/kratos/commit/305bb28d689dabc4d211baac5e6babd34862af5f))
* Make admin recovery to work without emails [#1419](https://github.com/ory/kratos/issues/1419) ([#1750](https://github.com/ory/kratos/issues/1750)) ([db00e85](https://github.com/ory/kratos/commit/db00e85e65c31b2bc497f0f4b4a28684b9f8bb9a))

### Tests

* **e2e:** Improved SDK set up and arm fix ([#1933](https://github.com/ory/kratos/issues/1933)) ([c914ba1](https://github.com/ory/kratos/commit/c914ba10a85e89c031e7acfb73bf22c53201e287))
* Update snapshots ([a820653](https://github.com/ory/kratos/commit/a820653718475656b7ae44a1bc7235a8fb97b8b5))


# [0.8.0-alpha.3](https://github.com/ory/kratos/compare/v0.8.0-alpha.2...v0.8.0-alpha.3) (2021-10-28)

Resolves issues in the quickstart.





### Bug Fixes

* Resolve quickstart issues ([#1900](https://github.com/ory/kratos/issues/1900)) ([d047009](https://github.com/ory/kratos/commit/d0470095f3263e287f76e8be0abb8df332492dd9)):

    Closes https://github.com/ory/kratos/discussions/1899


### Code Generation

* Pin v0.8.0-alpha.3 release commit ([a307deb](https://github.com/ory/kratos/commit/a307deb6779dacd2ce54e161a00d347600d2c583))


# [0.8.0-alpha.2](https://github.com/ory/kratos/compare/v0.8.0-alpha.1...v0.8.0-alpha.2) (2021-10-28)

Resolves an issue in the SDK release pipeline.





### Code Generation

* Pin v0.8.0-alpha.2 release commit ([2178929](https://github.com/ory/kratos/commit/217892978c4fa9897a88b140276c2d27622c5de4))


# [0.8.0-alpha.1](https://github.com/ory/kratos/compare/v0.7.6-alpha.1...v0.8.0-alpha.1) (2021-10-27)

We are extremely excited to share this next generation of Ory Kratos! The project is truly maturing and the community is getting larger by the hour.

On this special occasion, we would like to bring to your attention that the [**Ory Summit is happening tomorrow and on Friday!**](https://events.hubilo.com/ory-summit/register?mtm_campaign=ory-summit-2021&mtm_kwd=banner-landingpage) You will hear gripping talks from the Ory Community and Ory maintainers! And the best part, tickets are free and we are covering multiple time zones!

This release is truly the best version of Ory Kratos to date and we want to give you a tl;dr of the 345 commits and 1152 files changed, and what you can expect from this release:

- Full multi-factor authentication with different enforcement policies (soft/hard MFA).
- Support for WebAuthn (FIDO2 / U2F) two-factor authentication - from fingerprints to hardware tokens every FIDO2 device is supported!
- Ability to fetch the initial OAuth2 Access and Refresh and OpenID Connect ID Tokens an identity receives when performing social sign up. Optionally, these tokens are stored encrypted in the database (XChaCha20Poly1305 or AES-GCM)!
- Support for TOTP (Google Authenticator) two-factor verification/authentication.
- Advanced two-factor recovery with lookup secrets.
- [A complete reference implementation of the Ory Kratos end-user (self-service) facing UI in ReactJS & VercelJS](https://github.com/ory/kratos-react-nextjs-ui).
- "Native" support for Single-Page App Single Sign-On.
- Much improved single-page app and native app APIs for all self-service flows.
- Support for PKBDF2 password hashing, which will help import user passwords from other systems in the future.
- Bugfixes and improvements to the OpenAPI spec and auto-generated SDKs.
- ARM Docker Images.
- Greatly improved internal e2e test pipeline using Cypress 8.x.
- Improved functional tests with cupaloy snapshot testing.
- Documentation on different error codes and message identifiers to easier translate messages in your own UI.
- Better form decoding and ability to mark required JSON Schema fields as required in the UI.
- Bug fixes that could result in users ending up in irrecoverable UI states.
- Better support for `return_to` across flows (e.g. OIDC) and in custom UIs.
- SBOM Software Supply Chain scanning & reporting.
- Docker Image vulnerability checking as part of the release pipeline.
- Support sending emails via AWS SES SMTP.
- A REST endpoint to invalidate all an identity's sessions.

As you can see, much has happened and we are grateful for all the great interactions we have with you, every day!

Let's take a look at some of the breaking changes. Even though much was added, little has changed in breaking ways! This is a testament that Ory Kratos' internals and APIs are becoming more stable!

This release requires you to run SQL migrations. Please, as always, create a backup of your database first!

The SDKs are now generated with tag v0alpha2 to reflect that some signatures have changed in a breaking fashion. Please update your imports from `v0alpha1` to `v0alpha2`.

The SMTPS scheme used in courier config URL with cleartext/StartTLS/TLS SMTP connection types is now only supporting implicit TLS. For StartTLS and cleartext SMTP, please use the SMTP scheme instead.

Example:
- SMTP Cleartext: `smtp://foo:bar@my-mailserver:1234/?disable_starttls=true`
- SMTP with StartTLS: `smtps://foo:bar@my-mailserver:1234/` -> `smtp://foo:bar@my-mailserver:1234/`
- SMTP with implicit TLS: `smtps://foo:bar@my-mailserver:1234/?legacy_ssl=true` -> `smtps://foo:bar@my-mailserver:1234/We are extremely excited to share this next generation of Ory Kratos! The project is truly maturing and the community is getting larger by the hour.

On this special occasion, we would like to bring to your attention that the [**Ory Summit is happening tomorrow and on Friday!**](https://events.hubilo.com/ory-summit/register?mtm_campaign=ory-summit-2021&mtm_kwd=banner-landingpage) You will hear gripping talks from the Ory Community and Ory maintainers! And the best part, tickets are free and we are covering multiple time zones!

This release is truly the best version of Ory Kratos to date and we want to give you a tl;dr of the 345 commits and 1152 files changed, and what you can expect from this release:

- Full multi-factor authentication with different enforcement policies (soft/hard MFA).
- Support for WebAuthn (FIDO2 / U2F) two-factor authentication - from fingerprints to hardware tokens every FIDO2 device is supported!
- Ability to fetch the initial OAuth2 Access and Refresh and OpenID Connect ID Tokens an identity receives when performing social sign up. Optionally, these tokens are stored encrypted in the database (XChaCha20Poly1305 or AES-GCM)!
- Support for TOTP (Google Authenticator) two-factor verification/authentication.
- Advanced two-factor recovery with lookup secrets.
- [A complete reference implementation of the Ory Kratos end-user (self-service) facing UI in ReactJS & VercelJS](https://github.com/ory/kratos-react-nextjs-ui).
- "Native" support for Single-Page App Single Sign-On.
- Much improved single-page app and native app APIs for all self-service flows.
- Support for PKBDF2 password hashing, which will help import user passwords from other systems in the future.
- Bugfixes and improvements to the OpenAPI spec and auto-generated SDKs.
- ARM Docker Images.
- Greatly improved internal e2e test pipeline using Cypress 8.x.
- Improved functional tests with cupaloy snapshot testing.
- Documentation on different error codes and message identifiers to easier translate messages in your own UI.
- Better form decoding and ability to mark required JSON Schema fields as required in the UI.
- Bug fixes that could result in users ending up in irrecoverable UI states.
- Better support for `return_to` across flows (e.g. OIDC) and in custom UIs.
- SBOM Software Supply Chain scanning & reporting.
- Docker Image vulnerability checking as part of the release pipeline.
- Support sending emails via AWS SES SMTP.
- A REST endpoint to invalidate all an identity's sessions.

As you can see, much has happened and we are grateful for all the great interactions we have with you, every day!

Let's take a look at some of the breaking changes. Even though much was added, little has changed in breaking ways! This is a testament that Ory Kratos' internals and APIs are becoming more stable!

This release requires you to run SQL migrations. Please, as always, create a backup of your database first!

The SDKs are now generated with tag v0alpha2 to reflect that some signatures have changed in a breaking fashion. Please update your imports from `v0alpha1` to `v0alpha2`.

The SMTPS scheme used in courier config URL with cleartext/StartTLS/TLS SMTP connection types is now only supporting implicit TLS. For StartTLS and cleartext SMTP, please use the SMTP scheme instead.

Example:
- SMTP Cleartext: `smtp://foo:bar@my-mailserver:1234/?disable_starttls=true`
- SMTP with StartTLS: `smtps://foo:bar@my-mailserver:1234/` -> `smtp://foo:bar@my-mailserver:1234/`
- SMTP with implicit TLS: `smtps://foo:bar@my-mailserver:1234/?legacy_ssl=true` -> `smtps://foo:bar@my-mailserver:1234/We are extremely excited to share this next generation of Ory Kratos! The project is truly maturing and the community is getting larger by the hour.

On this special occasion, we would like to bring to your attention that the [**Ory Summit is happening tomorrow and on Friday!**](https://events.hubilo.com/ory-summit/register?mtm_campaign=ory-summit-2021&mtm_kwd=banner-landingpage) You will hear gripping talks from the Ory Community and Ory maintainers! And the best part, tickets are free and we are covering multiple time zones!

This release is truly the best version of Ory Kratos to date and we want to give you a tl;dr of the 345 commits and 1152 files changed, and what you can expect from this release:

- Full multi-factor authentication with different enforcement policies (soft/hard MFA).
- Support for WebAuthn (FIDO2 / U2F) two-factor authentication - from fingerprints to hardware tokens every FIDO2 device is supported!
- Ability to fetch the initial OAuth2 Access and Refresh and OpenID Connect ID Tokens an identity receives when performing social sign up. Optionally, these tokens are stored encrypted in the database (XChaCha20Poly1305 or AES-GCM)!
- Support for TOTP (Google Authenticator) two-factor verification/authentication.
- Advanced two-factor recovery with lookup secrets.
- [A complete reference implementation of the Ory Kratos end-user (self-service) facing UI in ReactJS & VercelJS](https://github.com/ory/kratos-react-nextjs-ui).
- "Native" support for Single-Page App Single Sign-On.
- Much improved single-page app and native app APIs for all self-service flows.
- Support for PKBDF2 password hashing, which will help import user passwords from other systems in the future.
- Bugfixes and improvements to the OpenAPI spec and auto-generated SDKs.
- ARM Docker Images.
- Greatly improved internal e2e test pipeline using Cypress 8.x.
- Improved functional tests with cupaloy snapshot testing.
- Documentation on different error codes and message identifiers to easier translate messages in your own UI.
- Better form decoding and ability to mark required JSON Schema fields as required in the UI.
- Bug fixes that could result in users ending up in irrecoverable UI states.
- Better support for `return_to` across flows (e.g. OIDC) and in custom UIs.
- SBOM Software Supply Chain scanning & reporting.
- Docker Image vulnerability checking as part of the release pipeline.
- Support sending emails via AWS SES SMTP.
- A REST endpoint to invalidate all an identity's sessions.

As you can see, much has happened and we are grateful for all the great interactions we have with you, every day!

Let's take a look at some of the breaking changes. Even though much was added, little has changed in breaking ways! This is a testament that Ory Kratos' internals and APIs are becoming more stable!

This release requires you to run SQL migrations. Please, as always, create a backup of your database first!

The SDKs are now generated with tag v0alpha2 to reflect that some signatures have changed in a breaking fashion. Please update your imports from `v0alpha1` to `v0alpha2`.

The SMTPS scheme used in courier config URL with cleartext/StartTLS/TLS SMTP connection types is now only supporting implicit TLS. For StartTLS and cleartext SMTP, please use the SMTP scheme instead.

Example:
- SMTP Cleartext: `smtp://foo:bar@my-mailserver:1234/?disable_starttls=true`
- SMTP with StartTLS: `smtps://foo:bar@my-mailserver:1234/` -> `smtp://foo:bar@my-mailserver:1234/`
- SMTP with implicit TLS: `smtps://foo:bar@my-mailserver:1234/?legacy_ssl=true` -> `smtps://foo:bar@my-mailserver:1234/`



## Breaking Changes

The location of the homebrew tap has changed from `ory/ory/kratos` to `ory/tap/kratos`.

To stay consistent with other query parameter's, the self-service login flow's `forced` key has been renamed to `refresh`.

The SDKs are now generated with tag v0alpha2 to reflect that some signatures have changed in a breaking fashion. Please update your imports from `v0alpha1` to `v0alpha2`.

To support 2FA on non-browser (e.g. native mobile) apps we have added the Ory Session Token as a possible parameter to both `initializeSelfServiceLoginFlowWithoutBrowser` and `submitSelfServiceLoginFlow`. Depending on the SDK generator, the order of the arguments may have changed. In JavaScript:

```patch
- .submitSelfServiceLoginFlow(flow.id, payload)
+ .submitSelfServiceLoginFlow(flow.id, sessionToken, payload)
+ // or if the user has no session yet:
+ .submitSelfServiceLoginFlow(flow.id, undefined, payload)
```

To improve the overall API design we have changed the result of `POST /self-service/settings`. Instead of having flow be a key, the flow is now the response. The updated identity payload stays the same!

```patch
 {
-  "flow": {
-    "id": "flow-id-..."
-    ...
-  },
+  "id": "flow-id-..."
+  ...
   "identity": {
     "id": "identity-id-..."
   }
 }
```

The SMTPS scheme used in courier config url with cleartext/StartTLS/TLS SMTP connection types is now only supporting implicit TLS. For StartTLS and cleartext SMTP, please use the smtp scheme instead.

Example:
- SMTP Cleartext: `smtp://foo:bar@my-mailserver:1234/?disable_starttls=true`
- SMTP with StartTLS: `smtps://foo:bar@my-mailserver:1234/` -> `smtp://foo:bar@my-mailserver:1234/`
- SMTP with implicit TLS: `smtps://foo:bar@my-mailserver:1234/?legacy_ssl=true` -> `smtps://foo:bar@my-mailserver:1234/`

This patch changes the naming and number of prometheus metrics (see: https://github.com/ory/x/pull/379). In short: all metrics will have now `http_` prefix to conform to Prometheus best practices.



### Bug Fixes

* Add error id ([1442784](https://github.com/ory/kratos/commit/1442784264d1f5032830a0646b853b925bb19c62))
* Add mfa e2e test scenarios and resolve found issues ([436992d](https://github.com/ory/kratos/commit/436992ddf2ace68b247c708fc955fccb95cf6fd2))
* Add middleware earlier [#1775](https://github.com/ory/kratos/issues/1775) ([#1776](https://github.com/ory/kratos/issues/1776)) ([b9d253e](https://github.com/ory/kratos/commit/b9d253ef05ff7cd616111a817d03a17e39f8f4a8))
* Allow refresh and aal upgrade at the same time ([2ec801f](https://github.com/ory/kratos/commit/2ec801f262cd8f6dcdf8121a20897257e3b74ad3))
* API client leaks stack trace with an error ([#1772](https://github.com/ory/kratos/issues/1772)) ([d3aff6d](https://github.com/ory/kratos/commit/d3aff6d3eb11942fbfd6f2de71f4399053075b62)), closes [#1771](https://github.com/ory/kratos/issues/1771)
* Better const handling for internal context ([1e457e3](https://github.com/ory/kratos/commit/1e457e3b3dea9ea9a05c12740578af2d45902aba))
* Correct swagger path for /identities/:id/session endpoint ([#1756](https://github.com/ory/kratos/issues/1756)) ([d614f2a](https://github.com/ory/kratos/commit/d614f2a737eef90ad60a4bdedae248b74131ff35))
* Decoder regression in registration ([febf75a](https://github.com/ory/kratos/commit/febf75ae959a2b67c19fcd1705b591f22ff5314b))
* Deterministic clidoc dates ([e48d90a](https://github.com/ory/kratos/commit/e48d90ad5a178ab3317d89800526c516aad6e274))
* Disable totp per default ([7278589](https://github.com/ory/kratos/commit/7278589ff2460a13302650b5e3fae01d774f9684))
* Docs autogen should not use `time.Now` ([a830f5b](https://github.com/ory/kratos/commit/a830f5b3b535bc375e879c797626b6084b76776e))
* Ensure correct error propagation ([77ce709](https://github.com/ory/kratos/commit/77ce709d53d88f70c892ab0892c13e16f5b761a5))
* Ensure refresh issues a new session when the identity changes ([a10b385](https://github.com/ory/kratos/commit/a10b385510a0102ede5850f9be30b7deba810acf))
* Ensure return_to works for OIDC flows ([d615734](https://github.com/ory/kratos/commit/d615734c312db6f7fa48fb8c7b4090a80c9e5ce7)), closes [#1773](https://github.com/ory/kratos/issues/1773)
* Explicit validation for return to in new flows ([284cf29](https://github.com/ory/kratos/commit/284cf29a6be82530b55c24a15c465ec9f1b6a210))
* Follow chrome webauthn best practice recommendation ([0a7c812](https://github.com/ory/kratos/commit/0a7c8128bb0b78f8dc236af06ca9be038b201829))
* Githup-app name in config ([#1822](https://github.com/ory/kratos/issues/1822)) ([1b50963](https://github.com/ory/kratos/commit/1b50963525ceaceea9afb8d1236d728de3107a8e))
* Handle return errors on the frontend and break early ([0e8d481](https://github.com/ory/kratos/commit/0e8d481cc220777aa56faf2e716da15537fa27fc)):

    Closes https://github.com/ory-corp/cloud/issues/1426

* Identity credential identifiers are now unique per method ([57fd99a](https://github.com/ory/kratos/commit/57fd99ac05d29fc0362f14e5910641944232d61e))
* Improve schema validation error tracing ([f793fe5](https://github.com/ory/kratos/commit/f793fe56182f3f195a57fe5f4b54f7fcf8402c81))
* Incorrect JSON response for browser flows ([1501f56](https://github.com/ory/kratos/commit/1501f5627ed12d2d149f1fcf49fcf326120e6b0b))
* Kill modd as well ([e5a98e5](https://github.com/ory/kratos/commit/e5a98e54ec68f122615dd902df9ebac788fdb579))
* **link:** Resolve incorrect response types when opening API recovery link in browser ([35ea8db](https://github.com/ory/kratos/commit/35ea8db300c2d3eeaf7d8f0e29c604ecc455cd2b))
* **login:** Properly handle refresh ([8dc7059](https://github.com/ory/kratos/commit/8dc7059222fa12dd0bca0183f42306b5169addb6))
* **lookup:** Ensure correct fields are set ([5ed4c55](https://github.com/ory/kratos/commit/5ed4c5572f9cbb35461e45dfc6b7c5eb4bce7434))
* **lookup:** Resolve reuse scenarios ([dbfe475](https://github.com/ory/kratos/commit/dbfe475ba5f0d2b9d4b0b67d0d8e7cb99e89ad5d))
* **lookup:** Set up codes correctly ([2f373f3](https://github.com/ory/kratos/commit/2f373f344326fbd5dbebf6233dbf5b56252b7e95))
* OIDC provider field in spec ([#1809](https://github.com/ory/kratos/issues/1809)) ([11b25de](https://github.com/ory/kratos/commit/11b25deb46b73c7d0ab95a77ff2ab60c032c1942))
* **oidc:** Ensure nested keys work on login ([71583c5](https://github.com/ory/kratos/commit/71583c57f1334bee1e5c9be1fae6a1b241ea3d6d))
* Omitempty for VerifiedAt and StateChangedAt ([#1736](https://github.com/ory/kratos/issues/1736)) ([bf2ec6e](https://github.com/ory/kratos/commit/bf2ec6e6ae8d656ea6dcac037dedd3603ad12915)):

    Closes https://github.com/ory/sdk/issues/95
    
    

* Only respect required modules for SDK ([4c5677f](https://github.com/ory/kratos/commit/4c5677f3ea48bd87e5d7a1f95e3807b7884a0b64))
* Panic when recovering deactivated user ([0a49f27](https://github.com/ory/kratos/commit/0a49f2714991a3f397dc5c721fe22d11846d3db5)), closes [#1794](https://github.com/ory/kratos/issues/1794) [#1826](https://github.com/ory/kratos/issues/1826)
* Potentially resolve hanging postgres connection closing ([693a928](https://github.com/ory/kratos/commit/693a9286b02c2329dcfd358a038857901193b459))
* Properly encode aal error ([49b6288](https://github.com/ory/kratos/commit/49b6288c2345840a7517272e9616c2c20a254edb))
* Properly open recovery endpoints in browser if flow was initiated via API ([23c12e5](https://github.com/ory/kratos/commit/23c12e55d24591ca69c9178017355a9262fa35eb))
* Remove duplicate schema error ([4e69123](https://github.com/ory/kratos/commit/4e691238da3bf3ee8d9a92d4d9507b27fce20199))
* Remove initial_value again as it was not useful outside of booleans ([0cc984b](https://github.com/ory/kratos/commit/0cc984b85baff3db500fb656bd541cfa0396df98))
* Remove obsolete openapi patch ([11618ec](https://github.com/ory/kratos/commit/11618ecc6681a9108ee70a3e0d1ab3d21e33f9db))
* Remove unnecessary cmd reference ([351760e](https://github.com/ory/kratos/commit/351760ece01d421687179b8e3f6f48a720247a1d))
* Replace 302 with 303 ([2e2b0f8](https://github.com/ory/kratos/commit/2e2b0f840450c6d23f3e51e5885d0908685ef3f6))
* Resolve clidoc generation issue ([1aaaa03](https://github.com/ory/kratos/commit/1aaaa035f863852799575e1f65e9d9ed276a3160))
* Resolve merge issues ([1dc7497](https://github.com/ory/kratos/commit/1dc74976c785afca8079379cd5060116b5f3d831))
* Resolve openapi issues and regenerate clients ([f7d60c0](https://github.com/ory/kratos/commit/f7d60c02392d2ad664c73ee4ff6bb108a4cb04e2))
* Resolve swagger regression ([02b9d47](https://github.com/ory/kratos/commit/02b9d470df012ae9818a8516a5549aee83c0963d))
* Run format on ts files ([f55f6f6](https://github.com/ory/kratos/commit/f55f6f69bf0df88d001fda791b330bdcbf5d92b2))
* Slow CLI start-up time ([ae20c17](https://github.com/ory/kratos/commit/ae20c17777eb57363f811b57d782db88b2de91ae)):

    Found a deeply nested dependency which was importing `https://github.com/markbates/pkger`, causing unreasonable CPU consumption and significant delay at start up time. With this patch, start up time was reduced from almost 3s to ~0.01s.
    
    ```
    $ time kratos
    kratos  2.55s user 2.46s system 508% cpu 0.986 total
    
    $ time ./kratos-patch
    ./kratos-patch  0.00s user 0.00s system 64% cpu 0.001 total
    ```

* **test:** OIDC storategy test ([#1836](https://github.com/ory/kratos/issues/1836)) ([b877dbe](https://github.com/ory/kratos/commit/b877dbecaf84e2d102bcceff4ad85c5b4efe18c5))
* **totp:** Reorder QR ([d096df7](https://github.com/ory/kratos/commit/d096df734ba8cf7dcfb872af03a19550d320c8b7))
* Try and reduce cookie flakyness ([e7ae8d6](https://github.com/ory/kratos/commit/e7ae8d63a16df69fd43afdf41691b9c1d3efe439))
* Typo ([8c4d8a2](https://github.com/ory/kratos/commit/8c4d8a2284f7a52a2dca7e7fd5e686756d410647))
* **ui:** Use correct type for anchor ([a6595e4](https://github.com/ory/kratos/commit/a6595e49c38a302f4a603dd46f5a0764680a24b1))
* Update schema config location ([539ae73](https://github.com/ory/kratos/commit/539ae7303158f14ca42165c12f9d3e8ef9dcdbdf))
* Use parallelism of 1 in go test ([8736334](https://github.com/ory/kratos/commit/8736334bf11fc9a742e2972aa97ee56c407c7c0c))
* **webauthn:** Support react-based webauth ([b6123b4](https://github.com/ory/kratos/commit/b6123b4840547b295be44272e76454462a0f60c4))
* X-session-token must not be mandatory ([05d73be](https://github.com/ory/kratos/commit/05d73beed26f1be31c6f2a62499c7c71d7d54bec))

### Code Generation

* Pin v0.8.0-alpha.1 release commit ([c2c902c](https://github.com/ory/kratos/commit/c2c902c1bd8d910843d747c25b99ee1bcc6f962d))

### Code Refactoring

* **courier:** Support SMTP schemes for implicit TLS, explicit StartTLS, and cleartext SMTP ([#1831](https://github.com/ory/kratos/issues/1831)) ([4cb082c](https://github.com/ory/kratos/commit/4cb082ce1e15ddd1d992a2def9e7d6410142cc02)), closes [#1770](https://github.com/ory/kratos/issues/1770) [#1769](https://github.com/ory/kratos/issues/1769)
* Homogenize error messages ([421a319](https://github.com/ory/kratos/commit/421a3190d1d4f6f5d96ef8ad87c3a2a667b57a28))
* Improved prometheus metrics ([#1830](https://github.com/ory/kratos/issues/1830)) ([0be993b](https://github.com/ory/kratos/commit/0be993bebeb9e50d90806ad13f60bb8d72c3b2d3)), closes [#1735](https://github.com/ory/kratos/issues/1735):

    This will add new prometheus metrics for Kratos that are more useful for alerting and increase overall observability.

* Login flow `forced` renamed to `refresh` ([92087e5](https://github.com/ory/kratos/commit/92087e5f00b4fcce1706442c9edf1b466f9a23c9))
* **login:** Rename forced -> refresh ([8d1e54b](https://github.com/ory/kratos/commit/8d1e54bd79cf617985602997f1121e168f58c389))
* **login:** Support 2FA for non-browser SDKs ([df4846d](https://github.com/ory/kratos/commit/df4846d3867599f49e58b6b4d59b338916f37cbf))
* Move expired error into top-level flow module ([01a2602](https://github.com/ory/kratos/commit/01a26025375f1d958a7e345c61fb6ba5e3403efe))
* Move homebrew tap to ory/tap ([0ee67c3](https://github.com/ory/kratos/commit/0ee67c388a1fea8aa9633cbf684e1f62e16d61cc))
* Move node identifiers to node package ([b0a86dc](https://github.com/ory/kratos/commit/b0a86dc6e5005017a9a0fa2120560f668ab2432f))
* Revert decision to return 422 errors and streamline 401/403 ([8aa5318](https://github.com/ory/kratos/commit/8aa53187f1e78d693463a47fcd9aedab30d1b55f))
* Sdk API is no v0alpha2 ([3f06738](https://github.com/ory/kratos/commit/3f067386e32ad3baeec48fd21dd51659a5725970))
* **session:** CreateAndIssueCookie is now UpsertAndIssueCookie ([a6d134d](https://github.com/ory/kratos/commit/a6d134de7710c7e92e51f735f13b7757eb7011e5))
* **session:** CreateSession is now UpsertSession ([3ec81a2](https://github.com/ory/kratos/commit/3ec81a2cc401ff18052abd2a9ba060e665f0baa2))
* **settings:** Change settings success response ([12f98f2](https://github.com/ory/kratos/commit/12f98f2884294669bbb7eab7e8ed73a5372386f6))

### Documentation

* Add 2fa credentials ([f7899a7](https://github.com/ory/kratos/commit/f7899a761aaf59d2cfddc2c330a805456cfca947))
* Add 2fa guide ([b4eed76](https://github.com/ory/kratos/commit/b4eed76305ecf1de3461525fd2ea748ec94da53c))
* Add a commandline example for the logout ([#1753](https://github.com/ory/kratos/issues/1753)) ([81ba264](https://github.com/ory/kratos/commit/81ba2647a66fca99b7ed2e56a67deec75ac06b89))
* Add admin ui guide ([ac88060](https://github.com/ory/kratos/commit/ac88060ed7390f0a34db637880c1660b8c45b352))
* Add advanced custom UI documentation ([5e3a2cd](https://github.com/ory/kratos/commit/5e3a2cdbedf0005c89db717d1136c56ab3304ede))
* Add image assets ([6bc93ca](https://github.com/ory/kratos/commit/6bc93ca79283bd993b0176dda11ad9d5860a5e4f))
* Add missing angle bracket ([#1799](https://github.com/ory/kratos/issues/1799)) ([4270140](https://github.com/ory/kratos/commit/427014052ef905c2003e2cd0133d57bf83819776))
* Add ory sessions as a concept ([626c0c9](https://github.com/ory/kratos/commit/626c0c90bd2d683048618452ba421e40be92f587))
* Add powershell to deps ([#1853](https://github.com/ory/kratos/issues/1853)) ([e945336](https://github.com/ory/kratos/commit/e94533690b658c4afba81e694a052579f0ffff42)), closes [#1848](https://github.com/ory/kratos/issues/1848)
* **credentials:** Add AAL explanation ([c1f501e](https://github.com/ory/kratos/commit/c1f501e9ec3ba203fb252fbd56cb87843d667b17))
* Enhance error return values ([3799c24](https://github.com/ory/kratos/commit/3799c24fbc0397876df4f1c530e325bd1212d750))
* Fix invalid syntax ([#1819](https://github.com/ory/kratos/issues/1819)) ([8cd6428](https://github.com/ory/kratos/commit/8cd6428e40610fa40b9c59414beb3d5c614dddaa))
* Fix the flow links used for rendering ([#1752](https://github.com/ory/kratos/issues/1752)) ([131d2c2](https://github.com/ory/kratos/commit/131d2c284d4191ee979077937ea3b48fce772f3c))
* Fix the invalid links ([#1868](https://github.com/ory/kratos/issues/1868)) ([6d621ec](https://github.com/ory/kratos/commit/6d621ec89d1a7c37daf4622b06a0ad94f2d77b31))
* Remove obsolete file ([b7f9052](https://github.com/ory/kratos/commit/b7f905278edf4aed1e2984aa3d2d94a41368d6d8))
* Update generated docs ([72afb81](https://github.com/ory/kratos/commit/72afb81be8bfaa36236087ec7715bca1804aa62c))
* Update quickstart curl examples ([#1778](https://github.com/ory/kratos/issues/1778)) ([6c677c4](https://github.com/ory/kratos/commit/6c677c49df8fa8d48e7c0bbf91bbd18874f4c514))
* Use correct link ([f007919](https://github.com/ory/kratos/commit/f007919b7bd86c1d1b20b3625709e01b5f123302)), closes [#1793](https://github.com/ory/kratos/issues/1793)

### Features

* Add `intended_for_someone_else` error code ([572a131](https://github.com/ory/kratos/commit/572a1315aec7d1103c8d4fb9c128644ea2af6d3b))
* Add aal fallback for existing sessions ([a5c7b11](https://github.com/ory/kratos/commit/a5c7b1143bca7029bf94fc42fe638534961a06bc))
* Add authenticators after set up ([035c276](https://github.com/ory/kratos/commit/035c276152a22a2c9c7159b1cf89dbe7724728dd))
* Add DeleteCredentialsType to identity struct including tests ([b12bf52](https://github.com/ory/kratos/commit/b12bf523e4213e49f545206c457a1739f493d385))
* Add e2e tests for react native 2fa ([a3ac253](https://github.com/ory/kratos/commit/a3ac253bdb9c42df6dce9288a2e7c2dada24d255))
* Add error ids for csrf-related errors ([dc2adbf](https://github.com/ory/kratos/commit/dc2adbf52f7ee845778ee5c3c943b9e10c41e181))
* Add error ids for redirect-related errors ([246a045](https://github.com/ory/kratos/commit/246a0453e65e70635331c95ff02ec9133ae81e46))
* Add error ids for session-related errors ([087d907](https://github.com/ory/kratos/commit/087d90731185b71cd88cc6451ce360d6c1dada34))
* Add explicit return_to to flow objects and API parameters ([50d04ea](https://github.com/ory/kratos/commit/50d04eaa455932a9a5cc31f812f66518e1d4ad3b)), closes [#1605](https://github.com/ory/kratos/issues/1605) [#1121](https://github.com/ory/kratos/issues/1121):

    This patch adds a `return_to` field to the flow objects which contains the original `?return_to=...` value. It uses the Flow's `request_url` for that purpose.

* Add ids for user-facing errors for login, registration, settings ([787558b](https://github.com/ory/kratos/commit/787558b48fd7405ac61a48d3c18c7252ac1aaf19)):

    This patch adds a new field `id` to JSON error payloads. This helps tremendously in implementing better client-side (native / SPA) apps as the API now returns error IDs like `no_active_session`, `orbidden_return_to`, `no_verified_address` and more. UIs can use these IDs to decide what to do next in the application - for example redirecting to a particular endpoint or showing an error message.

* Add initial value to bool checkboxes ([63dba73](https://github.com/ory/kratos/commit/63dba737376dbe2f15c5afb5df22c593328c6483))
* Add internal context to login and registration ([723e6ee](https://github.com/ory/kratos/commit/723e6eee731d34f85bc4346a1040f2f121662ae9))
* Add internal context to settings flow ([afb6895](https://github.com/ory/kratos/commit/afb6895daa8743edbf4fca957b2a156e676ef63a))
* Add lookup node to disable lookup ([d0836be](https://github.com/ory/kratos/commit/d0836beb53709c88eb9ed78df39e95a3204c7cec)):

    See https://github.com/ory/cloud/issues/12

* Add lookup to config ([14119b6](https://github.com/ory/kratos/commit/14119b623941b6f5e795ef0d369ee9e3adb73207))
* Add lookup to identity ([ead3833](https://github.com/ory/kratos/commit/ead3833e254b4939f2b86b34e954580960cc7ea1))
* Add lookup to migrations ([dac4f75](https://github.com/ory/kratos/commit/dac4f759a0b92c1eebca177c3c931fb3146e7dee))
* Add MFA enforcment option to whoami and settings ([554d725](https://github.com/ory/kratos/commit/554d72552702818c8f1fc45fd1daf9d93c0d2cad))
* Add mfa for non-browser ([4096fd3](https://github.com/ory/kratos/commit/4096fd3fbdb430fd325b38fe9102defa31dd1b6d))
* Add missing migrations ([ccc64d8](https://github.com/ory/kratos/commit/ccc64d87935c6b5ad506dce6a5f903d56541f864))
* Add option to disable recovery codes ([9d3daa6](https://github.com/ory/kratos/commit/9d3daa656a5361ef8a90fe3511f9c1a6e9015969)):

    Closes https://github.com/ory/cloud/issues/12

* Add ory cli config ([5b959be](https://github.com/ory/kratos/commit/5b959beaba4d03e143f7701c30bc30e25f2c51cc))
* Add schema patch for new initial_value field ([131e380](https://github.com/ory/kratos/commit/131e3803ff6d04af9ec668286c8e6fcf88467214)):

    The field sets a node input's initial value. This is primarily used for fields which are e.g. checkboxes or buttons (active/inactive). If this field is set on a button, it implies that clicking the button should trigger the "value" to be set.

* Add script type and discriminator for attributes ([de0af95](https://github.com/ory/kratos/commit/de0af955904894d97997cf598686b6d33cd88bd4)):

    See https://github.com/ory/sdk/issues/72

* Add smtp headers config option ([#1747](https://github.com/ory/kratos/issues/1747)) ([7ffe0e9](https://github.com/ory/kratos/commit/7ffe0e9766e930615dbb6833e650b73a8975a544)), closes [#1725](https://github.com/ory/kratos/issues/1725)
* Add support for onclick javascript in ui nodes ([7cc7efa](https://github.com/ory/kratos/commit/7cc7efa00ff0e8107f1369573bfdf766fcfc0e93))
* Add totp strategy for settings flow ([d1d6617](https://github.com/ory/kratos/commit/d1d6617013fbcc37eaf48cf19061b86955fc5d5e)):

    This patch allows adding a TOTP device in the settings, and also removing it when no longer needed.

* Add webauthn identity credential ([f8b9582](https://github.com/ory/kratos/commit/f8b95828ea41d29c7f3577cc7772168135bc5514))
* Adding Dockle Container Linter ([#1852](https://github.com/ory/kratos/issues/1852)) ([3c0d519](https://github.com/ory/kratos/commit/3c0d519dd47657c6adca3d64bca8b3ed02cb7a8f))
* Adjust to new aal error handling ([b8956bc](https://github.com/ory/kratos/commit/b8956bc0fc8a45e88dd51f79608f9d6c34e2b6f3))
* API to return access, refresh, id tokens from social sign in ([#1818](https://github.com/ory/kratos/issues/1818)) ([198991a](https://github.com/ory/kratos/commit/198991a9ce25fbaccc927be3bd3f6b1593771bec)), closes [#1518](https://github.com/ory/kratos/issues/1518) [#397](https://github.com/ory/kratos/issues/397):

    This patch introduces the new `include_credential` query parameter to the `GET /identities` endpoint which allows administrators to receive the initial access, refresh, and ID tokens from Social Sign In (OpenID Connect / OAuth 2.0) flows.
    
    These tokens can be stored in an encrypted format (XChaCha20Poly1305 or AES-GCM) in the database if an appropriate encryption secret is set. To get started easily these values are not encrypted per default.
    
    For more information head [over to the docs](https://kratos/docs/guides/retrieve-social-sign-in-access-refresh-id-token).

* Auto-generate list of messages ([cf46339](https://github.com/ory/kratos/commit/cf46339b9a07cd72b4d01e40c2df72e6c8104e9b)), closes [#1784](https://github.com/ory/kratos/issues/1784)
* Endpoint to list all identity schemas ([#1703](https://github.com/ory/kratos/issues/1703)) ([aa23d5d](https://github.com/ory/kratos/commit/aa23d5d5af28d8a7789b4a0c7e97197c7758ad98)), closes [#1699](https://github.com/ory/kratos/issues/1699)
* Generate sdks and update versions ([c9d22d9](https://github.com/ory/kratos/commit/c9d22d91f5fe49b5f2818160ade58bfd265f03e5))
* **hash:** PBKDF2 password hash verification ([#1774](https://github.com/ory/kratos/issues/1774)) ([33cc7e0](https://github.com/ory/kratos/commit/33cc7e02d9bcc24ae1de438102660cc89fd008d6)), closes [#1659](https://github.com/ory/kratos/issues/1659)
* Identity schema validation on startup ([#1779](https://github.com/ory/kratos/issues/1779)) ([99db3f0](https://github.com/ory/kratos/commit/99db3f03afd4b2525cbce54133a1abd1d49d2886)), closes [#701](https://github.com/ory/kratos/issues/701)
* **identity:** Add AAL constants ([882573d](https://github.com/ory/kratos/commit/882573df5621446e799b17ca0ab09d3934e44437))
* Implement AAL for login and sessions ([45467e0](https://github.com/ory/kratos/commit/45467e0caba7ed31e2ebde71a8b32ecd5f8db7c2))
* Implement endpoint for invalidating all sessions for a given identity ([#1740](https://github.com/ory/kratos/issues/1740)) ([dbd1689](https://github.com/ory/kratos/commit/dbd1689c11fd0a3d999ea09b553dd4a14a7a6972)), closes [#655](https://github.com/ory/kratos/issues/655):

    This PR introduces endpoint to destroy all sessions for a given identity which effectively logouts user from all devices/sessions. This is useful when for some security concern we want to make sure there are no "old" sessions active or other "staff" related actions (such as force logout after password change etc.).

* Implement lookup code settings and login ([8f3ce7b](https://github.com/ory/kratos/commit/8f3ce7b33390fcae85e605193806364ca9d099c9))
* Improve detection of AAL errors and return 422 instead of 403 ([e2bfbea](https://github.com/ory/kratos/commit/e2bfbea1541aca983eb835d3da2b5fe70ac4b7a5))
* Improve labels for totp and lookup ([b92e00e](https://github.com/ory/kratos/commit/b92e00e345da1f8ab76750e3f0ae1301977bbae0))
* Improve session device annotations ([87907b8](https://github.com/ory/kratos/commit/87907b8d29dc9cd7140535e81ea62c2d7f8e41c3))
* In docker debug support with delve ([#1789](https://github.com/ory/kratos/issues/1789)) ([37325a1](https://github.com/ory/kratos/commit/37325a18d9430130d0062674433fa0d3f9a59eb3))
* Introduce cve scanning ([#1798](https://github.com/ory/kratos/issues/1798)) ([ade13ea](https://github.com/ory/kratos/commit/ade13ea082ee11e9c1005de3ccb3ae6b5f02bb49))
* **logout:** Add logout token to browser response ([#1758](https://github.com/ory/kratos/issues/1758)) ([d3f1177](https://github.com/ory/kratos/commit/d3f1177a9a82dc2c4f930f15c6ec87c3ec5a1d53))
* Mark recovery email address verified ([#1665](https://github.com/ory/kratos/issues/1665)) ([e3efc5d](https://github.com/ory/kratos/commit/e3efc5d0673106115a236e38b5d76d6672d64d20)), closes [#1662](https://github.com/ory/kratos/issues/1662)
* Mark required fiels as required ([34cd5e8](https://github.com/ory/kratos/commit/34cd5e8e638be3d48ed8174112417bc36400e8cb)):

    Closes https://github.com/ory-corp/cloud/issues/1328
    Closes https://github.com/ory/kratos/issues/400
    Closes https://github.com/ory/kratos/issues/1058
    See https://ory-community.slack.com/archives/C012RJ2MQ1H/p1631825476159000

* Natively support social sign in for single-page apps ([1a1a350](https://github.com/ory/kratos/commit/1a1a350a9f0df85195505690fc52086eddf78371))
* **persistence:** Add new columns for mfa ([6184fe3](https://github.com/ory/kratos/commit/6184fe385cf87b260117290089b06445e5b6b205))
* Potentially add arm64 docker support ([68112de](https://github.com/ory/kratos/commit/68112defb97db1c6f4b8bf65e2e522b22e27d280))
* Proper enum and type assertions for openapi ([c4d8516](https://github.com/ory/kratos/commit/c4d8516fb93c2127c6d0c28a914ed7b8f8646832))
* Publish webauthn as loadable script instead of eval ([2717c59](https://github.com/ory/kratos/commit/2717c5958ab3f088821fdf96fdf6d44d48fea310))
* Redirect on login if session aal is not matched ([8feff8d](https://github.com/ory/kratos/commit/8feff8daaf4ac744fab22627d9bdab45740570d5))
* Respect webauthn in session aal ([869b4a5](https://github.com/ory/kratos/commit/869b4a5a812b840196eaf1e591aeb685d7f0e904))
* **session:** Respect 2fa enforcement in whoami ([3a82c88](https://github.com/ory/kratos/commit/3a82c8806931a2b4cd05142a6dae8040a76658bc))
* Sign in with apple ([#1833](https://github.com/ory/kratos/issues/1833)) ([16ed123](https://github.com/ory/kratos/commit/16ed123adba06167f70eb952ae3877d4476f8c71)), closes [#1782](https://github.com/ory/kratos/issues/1782):

    Adds an adapter and configuration options for enabling Social Sign In with Apple.

* Sort totp nodes ([5c9a494](https://github.com/ory/kratos/commit/5c9a49487f45af5b7edf069edf9c3d37ef293cd5))
* Stubable time in text package ([22e4ed1](https://github.com/ory/kratos/commit/22e4ed15e2eecb51b393762077872b19f6f2acd2))
* Support apple m1 ([54b4fb6](https://github.com/ory/kratos/commit/54b4fb698c6a087afef8821fa8300798e484ae18))
* Support setting the identity state via the admin API ([#1805](https://github.com/ory/kratos/issues/1805)) ([29c060b](https://github.com/ory/kratos/commit/29c060bd348733eeafee98d5f255c737a8cbcad0)), closes [#1767](https://github.com/ory/kratos/issues/1767)
* Support strategy return to ui for settings ([74670bb](https://github.com/ory/kratos/commit/74670bb4b0cc45626537e5ac63283fd14f05dee1))
* Support webauthn for mfa ([e8f4d3c](https://github.com/ory/kratos/commit/e8f4d3cb899d44c777b094f2ae4d84ff68532bf9))
* **totp:** Add width and height to QR code ([a648ba3](https://github.com/ory/kratos/commit/a648ba3de9a0ba707ce39c37fa5d5e38c4da74d3))
* **totp:** Support account name setting from schema ([19a6bcc](https://github.com/ory/kratos/commit/19a6bcc9d8940acb2a5f0eb4a6cc7f28801a2f92))
* Treat lookup as aal2 in session ([3269028](https://github.com/ory/kratos/commit/3269028d46d0ef23de3f905c325d514f24db43b8))
* Use discriminators for ui node types in spec ([59e808e](https://github.com/ory/kratos/commit/59e808e8dc6339da59bbe08ebbcf7b840e3fdd50))
* Use initial_value in lookup strategy ([efe272f](https://github.com/ory/kratos/commit/efe272f06966edc4858602d94740b6ed36c12e57))

### Reverts

* 3745014 ([d493d10](https://github.com/ory/kratos/commit/d493d1049f90ca6ee7b85931e3652aa9fdeb0254))

### Tests

* Aal in login.NewFlow ([5986e38](https://github.com/ory/kratos/commit/5986e38e6ab9eec1761e4c723c807dc0ef2a3dfa))
* AcceptToRedirectOrJSON ([2ca153f](https://github.com/ory/kratos/commit/2ca153f027599c18583ce0ebacb5ed577b56ddf3))
* Add credentials test ([58b388c](https://github.com/ory/kratos/commit/58b388c70d5ff32822e8ac5f3a394e683273ac6a))
* Add expired test to login handler ([3bdb8ab](https://github.com/ory/kratos/commit/3bdb8abb558c0f8c4b33f712678f5da02d0ef4ee))
* Add identity change test to settings submit ([5eb090b](https://github.com/ory/kratos/commit/5eb090b2564192deb77e64dd74a07b96c381391d))
* Add initial spa e2e test ([20617f6](https://github.com/ory/kratos/commit/20617f628ac84981c3b47ce9e9ab193b8ff426d0))
* Add initial totp integration tests ([c9d456b](https://github.com/ory/kratos/commit/c9d456bf03cb33baf0745fe9a511f84b4c9427e3))
* Add login tests ([a71cadd](https://github.com/ory/kratos/commit/a71cadde91bdaf960caf30dcfa957a2646da86a2))
* Add migrations tests for new tables ([3c96ab0](https://github.com/ory/kratos/commit/3c96ab059af9bf6002b341c5db51d1b3ca5da655))
* Add react app to e2e tests ([1214eee](https://github.com/ory/kratos/commit/1214eeee24b06e6e72c55cfed2176860ecbf3c13))
* Add schema test for totp config ([c4f05ba](https://github.com/ory/kratos/commit/c4f05ba60af1d7ca31b4cf54097cbefa88085704))
* Add session amr test ([eedb60b](https://github.com/ory/kratos/commit/eedb60bec9bebfb0a4ffb67dd484d2e6b466e776))
* Add settings tests ([6959565](https://github.com/ory/kratos/commit/6959565212dc5e7296aad7f1365a944379dd5d6d))
* Add test for TOTPIssuer ([14731c4](https://github.com/ory/kratos/commit/14731c4e7809c2202c9298422c005358b7b26fc3))
* Add test for ui error page ([3977a9c](https://github.com/ory/kratos/commit/3977a9c4d6f98ef6d8f7f4c88d55b46579401ba8))
* Add TestEnsureInternalContext ([152bfc7](https://github.com/ory/kratos/commit/152bfc7294078081ca9f8fc6dd194db6d2e699ad))
* Add totp registry tests ([817e3ec](https://github.com/ory/kratos/commit/817e3ecb213454e4ce3f987ce8a8714301ee8165))
* Add totp settings tests ([c5a0d0f](https://github.com/ory/kratos/commit/c5a0d0f8435690786eaf719bb1376f7da15a6203))
* Add TOTP to profile ([7431e9f](https://github.com/ory/kratos/commit/7431e9fcf4e9c9853ec4d378221c7a3744b3b239))
* Add update session test ([47bd057](https://github.com/ory/kratos/commit/47bd057da0fbf849d643c27c6eb75ef09c5075fb))
* Additional checks for flow hydration ([a40d7fe](https://github.com/ory/kratos/commit/a40d7fe4340ff61c3fa9ac0a70dc5f7e4641a15e))
* Amr persistence ([b0b2d81](https://github.com/ory/kratos/commit/b0b2d8174ca46e066e8eb912a24d9e6efeea0ce8))
* Check if internal context is validated in store ([a23d851](https://github.com/ory/kratos/commit/a23d8518fc65f645cae9c196ff70df4efca67266))
* CheckAAL ([03b37e7](https://github.com/ory/kratos/commit/03b37e7675e369817d2bb226047ec9f26b18a456))
* Complete TOTP login integration tests ([6e503cf](https://github.com/ory/kratos/commit/6e503cff28428e707b3812cd2bf8e44ccc487b89))
* **e2e:** Add baseurl ([159b25f](https://github.com/ory/kratos/commit/159b25f7ab0ac659033d861868f472183b852167))
* **e2e:** Add checkboxes to schemas ([0c91f0c](https://github.com/ory/kratos/commit/0c91f0c89081726e7451d5411a6adeb631ae2edb))
* **e2e:** Add config for proxy to simplify cy.visit logic ([7d87985](https://github.com/ory/kratos/commit/7d8798560947227d64a35d2dd69623bc1a1ddc8f))
* **e2e:** Add mfa profile ([a60d157](https://github.com/ory/kratos/commit/a60d157bfeb79cb527bf73b3fc38e1ba5388cbed))
* **e2e:** Add modd to build ([48cd8ae](https://github.com/ory/kratos/commit/48cd8aeb851d02e2fd31e73e044befb45242e953))
* **e2e:** Add more helpers and ts defs ([21b35b0](https://github.com/ory/kratos/commit/21b35b025a21b1f6ab3ac8be79339f1734b3033a))
* **e2e:** Add more helpers for various flows and proxy settings ([755ac60](https://github.com/ory/kratos/commit/755ac60cb1a54cd188ab07d9448598d738c5e866))
* **e2e:** Add more routes to registry ([30423c9](https://github.com/ory/kratos/commit/30423c92ba27709e003e88e58072b78ef3e2aa04))
* **e2e:** Add more typings for cypress helpers ([60bd63f](https://github.com/ory/kratos/commit/60bd63f31d6b639af19048cc3d1e392b885213e0))
* **e2e:** Add plugin for using got ([8fafc40](https://github.com/ory/kratos/commit/8fafc40dff8a0d9d5d678b59ecf4c13755906a4f))
* **e2e:** Add proxy capabilities for react native app ([b5668df](https://github.com/ory/kratos/commit/b5668df755e186f12c0e543715bc2e16011583a6))
* **e2e:** Add recovery tests for SPA ([b6014ee](https://github.com/ory/kratos/commit/b6014eee8b507abf6e3b4324097b3015f722cbe3))
* **e2e:** Add spa as allowed redirect url ([2625d16](https://github.com/ory/kratos/commit/2625d1689d47fb1cdbe34708be27f2317cdc7bea))
* **e2e:** Add SPA tests for login and refactor tests to typescript ([d9a25df](https://github.com/ory/kratos/commit/d9a25df1ba34cbefd416dccfdb2f5fc93e0290b9))
* **e2e:** Add SPA tests for logout and refactor tests to typescript ([b0c6776](https://github.com/ory/kratos/commit/b0c67769e4afcdbc05d2c1966e38faa18404a5db))
* **e2e:** Add SPA tests for registration and refactor tests to typescript ([a61ed1e](https://github.com/ory/kratos/commit/a61ed1edb41df64f58e23f8c88894fb742fd275d))
* **e2e:** Add support functions and type definitions ([c82d68d](https://github.com/ory/kratos/commit/c82d68db36563b16623a63be9efaf6b25322f855))
* **e2e:** Clean up helper ([4806add](https://github.com/ory/kratos/commit/4806add17a5dd0ea8c8fded644a6c240b17861b3))
* **e2e:** Complete SPA tests for all mfa flows ([2196129](https://github.com/ory/kratos/commit/219612903bd4dce208e2074e4595980c1cb60711))
* **e2e:** Default and empty values and required fields ([72f2c5f](https://github.com/ory/kratos/commit/72f2c5fbd8227e19d62f26aeddfb1bd14d7c768b))
* **e2e:** Ensure advanced types work in forms also ([287269c](https://github.com/ory/kratos/commit/287269c9992390b52ff380b31eda3bb7ad205f09))
* **e2e:** Ensure correct app ([a9ff545](https://github.com/ory/kratos/commit/a9ff5457cb48a90668b62e54d0b08cb1e9108994))
* **e2e:** Finalize mobile tests ([acf5c3d](https://github.com/ory/kratos/commit/acf5c3d649e51edfd9e1e3755222d9c7161a92e7))
* **e2e:** Force port ([a49eda8](https://github.com/ory/kratos/commit/a49eda8e0405954d62058d8c1410a62f72bfb7ae))
* **e2e:** Homogenize profiles ([7798e19](https://github.com/ory/kratos/commit/7798e193aa3cce0347e5ca018e09685b6fda0ba2))
* **e2e:** Hot reload ory kratos on changes ([841da09](https://github.com/ory/kratos/commit/841da091689f9a3fceb5509490d7a2f4828b926f))
* **e2e:** Implement recovery tests for SPA ([3dea57f](https://github.com/ory/kratos/commit/3dea57ff986702b9a31621198794e1cc94e4881e))
* **e2e:** Implement required verification tests for SPA ([fb55f34](https://github.com/ory/kratos/commit/fb55f3475f25ab3aa6f7b1765ec5b9f13ef72b15))
* **e2e:** Improve stability for login tests ([43df22b](https://github.com/ory/kratos/commit/43df22bdd52305b2b5d98a0db1c09751bd3ebb4f))
* **e2e:** Improve stability for registration tests ([a1c59a3](https://github.com/ory/kratos/commit/a1c59a349cab3819e5f869dc89eba3c05100f1b8))
* **e2e:** Improve test reliability ([061a7e3](https://github.com/ory/kratos/commit/061a7e340c86b580abde02de3cb521dda7c23efb))
* **e2e:** Migrate email tests to new proxy set up ([54d8cd6](https://github.com/ory/kratos/commit/54d8cd65b8b19f7a643bf9d4060906b818fc91d6))
* **e2e:** Migrate settings tests to typescript and add SPA tests ([566336d](https://github.com/ory/kratos/commit/566336d910f0b3deb4675e1413bfd0182bde6a79))
* **e2e:** Move config to lower level and publish as package ([c21fa26](https://github.com/ory/kratos/commit/c21fa2688e560bb9c714d2078dbc9a72a1da125f))
* **e2e:** Move registration tests to new proxy set up ([eddeb85](https://github.com/ory/kratos/commit/eddeb8510ca4cb13d0644d7083d436778828d0bd))
* **e2e:** Port mobile test to typescript ([db42346](https://github.com/ory/kratos/commit/db4234694723b7dc965c9e2cf4ba792bad0374e9))
* **e2e:** Port remaining e2e tests to typescript ([5853d1a](https://github.com/ory/kratos/commit/5853d1a64b3f7b20af79cc6ebbc381de0d213139))
* **e2e:** Potentially resolve flaky login test ([e237d66](https://github.com/ory/kratos/commit/e237d66adbc3cce972d8e4689a88d02b9a925354))
* **e2e:** Potentially resolve webauthn startup issues ([eae6f5d](https://github.com/ory/kratos/commit/eae6f5d1e9dc08dc8f7152a9c441e029dd4351f3))
* **e2e:** Prototype typescript implementation ([2e869cf](https://github.com/ory/kratos/commit/2e869cff7b1cb87e15013a86b54fda16a01e0267))
* **e2e:** Recreate identities per flow ([1a560a3](https://github.com/ory/kratos/commit/1a560a37c13240d9ae16d34188a6221f589ebbbc))
* **e2e:** Reduce flaky tests ([cae86e7](https://github.com/ory/kratos/commit/cae86e7f6a4fcc9e1433b9c063efe3745273f2dc))
* **e2e:** Reduce test flakes in lookup codes ([bfea354](https://github.com/ory/kratos/commit/bfea354f45858e5be0a588840f6e8125819a244c))
* **e2e:** Refactor and add support for SPA app ([7609219](https://github.com/ory/kratos/commit/7609219448effde35844675533e71583babe1d14))
* **e2e:** Remove wait condition ([af10b03](https://github.com/ory/kratos/commit/af10b03ebca03cdb5654c116efbd3c23b47c7594))
* **e2e:** Resolve broken test ([c7cf134](https://github.com/ory/kratos/commit/c7cf134fbfbbb59b276aa00d02bbad3886f78dee))
* **e2e:** Resolve flaky test ([de7cc59](https://github.com/ory/kratos/commit/de7cc59f07a6b77e3bbf3d98a7b2104b60ce708c))
* **e2e:** Resolve flaky test issues ([1627745](https://github.com/ory/kratos/commit/162774567d44336c8999ee0c1362adb191855d0c))
* **e2e:** Resolve next not starting ([2a2a3cb](https://github.com/ory/kratos/commit/2a2a3cb016e820f651f3cf6cd33123672e5977cb))
* **e2e:** Resolve regression ([d62f0c0](https://github.com/ory/kratos/commit/d62f0c02315702f55b998d4c48d4ca8c6a41827f))
* **e2e:** Resolve regressions ([aaff34e](https://github.com/ory/kratos/commit/aaff34ed66165f787103292ac0a034a0cdaf1308))
* **e2e:** Resolve regressions ([af9aedc](https://github.com/ory/kratos/commit/af9aedc8d29678f480b1b6bad128aefbacd6a373))
* **e2e:** Revert proxy changes ([293d920](https://github.com/ory/kratos/commit/293d92084a7614ae0cd7d5326dc82a209a0841be))
* **e2e:** Stabilize e2e tests ([a5dca28](https://github.com/ory/kratos/commit/a5dca2839ef66217b0046262a7e1fc886276509f))
* **e2e:** Temporarily add totp to default profile ([8ffac9d](https://github.com/ory/kratos/commit/8ffac9d138656eb2322913992b350cea31ed7e87))
* **e2e:** Update e2e profiles to new proxy set up ([a3204cf](https://github.com/ory/kratos/commit/a3204cf9b85e274441c02592288a4f322481e894))
* **e2e:** Use 127.0.0.1 to prevent ipv6 issues ([6f4b534](https://github.com/ory/kratos/commit/6f4b5340d33b31a5e4582858b544beb9c82181c7))
* **e2e:** Wait for oidc to trigger ([9c67c49](https://github.com/ory/kratos/commit/9c67c49235a562430da7ae60426d60cfd6120fca))
* Enable cookie debug ([81c3064](https://github.com/ory/kratos/commit/81c3064d69f8a233b8e0b78e103f2a23ae63cb63))
* Ensure aal and amr is set on recovery ([5cbab54](https://github.com/ory/kratos/commit/5cbab54fe5780689f0b64700567ac4632eb04c0b)), closes [#1322](https://github.com/ory/kratos/issues/1322)
* Ensure aal2 can not be used for oidc ([cbbcdd2](https://github.com/ory/kratos/commit/cbbcdd2e86c2d4da14c478637105eb8a36ae06c0))
* Ensure aal2 can not be used for password ([d9d39f0](https://github.com/ory/kratos/commit/d9d39f0bdda0725989a0a8261a449cf1a71afb6b))
* Ensure authenticated_at after all upgrade ([80408b4](https://github.com/ory/kratos/commit/80408b4c90229c61138411be8534fc577b8f0f33))
* Ensure redirect_url in password strategy ([9eafc10](https://github.com/ory/kratos/commit/9eafc10189ca88724fa6d75748299c2dd2c470b1))
* ErrStrategyAsksToReturnToUI behavior ([f739018](https://github.com/ory/kratos/commit/f7390184b02d526bb6e3ff496abc4522afc39d5a))
* Finalize webauthn tests ([97e59e6](https://github.com/ory/kratos/commit/97e59e61ee8be263199c3749e27dd81344777166))
* Fix regressions in the tests ([246c580](https://github.com/ory/kratos/commit/246c580222acd193eea784a6cbfd1e75181a484f))
* Fix tests in cmd/serve ([#1755](https://github.com/ory/kratos/issues/1755)) ([b704d08](https://github.com/ory/kratos/commit/b704d08382a9059157c2a649872e88943d66a99f))
* ID methods of node attributes ([ff9ff04](https://github.com/ory/kratos/commit/ff9ff048ddfa13ae73571064a36b33a867727392))
* Login form submission with AAL ([4d54fbb](https://github.com/ory/kratos/commit/4d54fbb37349126418274de8e21473c2ff81f785))
* **lookup:** Add secret_disable to snapshots ([68d6a87](https://github.com/ory/kratos/commit/68d6a876a4f1a0fd74789798397bd325a68d71d6))
* **lookup:** Ensure context is cleaned up after use ([8a210c4](https://github.com/ory/kratos/commit/8a210c41696d1865cce4c589a7cb3e52283fe24d))
* **lookup:** Refresh and reuse scenarios ([89736ed](https://github.com/ory/kratos/commit/89736ed9ba8667314313ca549a6377faddcc3d80))
* **migration:** Resolve mysql migration issue with empty array ([71a5649](https://github.com/ory/kratos/commit/71a5649a52036e29b351b6b4ee220ec7ce3aed05))
* Move to cupaloy for snapshots ([0cce70f](https://github.com/ory/kratos/commit/0cce70f47712da44d891c6d2890e818da6d9971b))
* Properly refresh mobile session ([c31915d](https://github.com/ory/kratos/commit/c31915de32e4b3db4af8ca8f3b5ecb0adf01a510))
* Registry regression ([25c88b5](https://github.com/ory/kratos/commit/25c88b55577b016aa77d2df3c595410633d0eefe))
* Remove todo items ([f60050e](https://github.com/ory/kratos/commit/f60050e0e30b1bf5441c95ada5777743719d65f1))
* Resolve flaky config test ([147c670](https://github.com/ory/kratos/commit/147c6704a9d38b5687eb8aba5661f24f99e577e3))
* Resolve flaky config test ([#1832](https://github.com/ory/kratos/issues/1832)) ([db98d01](https://github.com/ory/kratos/commit/db98d010639bfc387ef927c4f80ff6cd0ebc9588))
* Resolve flaky example tests ([#1817](https://github.com/ory/kratos/issues/1817)) ([0e700d8](https://github.com/ory/kratos/commit/0e700d89c0aaa99b9eec7ce070b7974373377f03))
* Resolve flaky tests ([2bd9100](https://github.com/ory/kratos/commit/2bd910037efd20ab1829784ee087c533e5e8b177))
* Resolve migratest regressions ([e9a1ed1](https://github.com/ory/kratos/commit/e9a1ed188a8f2556e1f60d1c171506dc0dd931d4))
* Resolve regressions ([1502ca1](https://github.com/ory/kratos/commit/1502ca1eb6c2e7ab698dc94675a50db63c326a41))
* Resolve regressions ([1a93b2f](https://github.com/ory/kratos/commit/1a93b2fba1fc41a6ba314253387af9770fd36f5a))
* Resolve regressions ([64850ed](https://github.com/ory/kratos/commit/64850ed3277185ebf68b50449721c903c01eab89))
* Resolve remaining regressions ([f02804c](https://github.com/ory/kratos/commit/f02804c567a532a30eaa228b0ba784b7f7fb0d9a))
* Resolve remaining regressions ([0224c22](https://github.com/ory/kratos/commit/0224c22ebda566c69363ae09dea9d42368c86f48))
* Resolve remaining regressions ([1fa2aa5](https://github.com/ory/kratos/commit/1fa2aa5b60d0b81e2035ae18c60d199b060a4c1f))
* Resolve time locality issues ([53b8b2a](https://github.com/ory/kratos/commit/53b8b2a22e5bad12dabf90c7bcbaf05b13a73a55))
* Restructure session struct tests ([50d3f66](https://github.com/ory/kratos/commit/50d3f66f82cb4e85a213fd86dc20bfadafefae23))
* Session AAL handling ([6fea3e5](https://github.com/ory/kratos/commit/6fea3e5aec6556697092c9a9d12295ed7e4d408b))
* Session activate ([c86fa03](https://github.com/ory/kratos/commit/c86fa03d3b2390403dcb14ef93307adc61ac7c79))
* **sql:** Fix incorrect UUID ([ea2894e](https://github.com/ory/kratos/commit/ea2894ed0f12de011fd5ce304dd614579ea5e96c))
* Temporarily enable lookup globally ([458f559](https://github.com/ory/kratos/commit/458f559ec816e64c6c9f53ecacdb4ae30fc9f8f7))
* **totp:** Ensure context is cleaned up after use ([1905883](https://github.com/ory/kratos/commit/19058830c0541f717360d3f599760b2a5cf47c4e))
* Upgrade cypress to 8.x ([c8a1dfc](https://github.com/ory/kratos/commit/c8a1dfcae3d42555b1215ad7eaa03a521bdcb1da))
* Use different return handler ([e489a43](https://github.com/ory/kratos/commit/e489a439e56dcd4218cf81284beaca0ef2ecd35e))
* Various aal combinations for newflow ([b095b99](https://github.com/ory/kratos/commit/b095b990224cbbd5ffa272b8f443b3345634d353))
* Webauth settings flow ([4c82772](https://github.com/ory/kratos/commit/4c82772ae28643ce69a5778c37f3c67644ef6f4c))
* Webauthn aal2 login ([60ace8b](https://github.com/ory/kratos/commit/60ace8b36c033ac4f9cd7e8cd929921e2e882946))
* Webauthn credentials ([c3e1184](https://github.com/ory/kratos/commit/c3e1184e719cd2041df8894edd4bd921bf2c3b00))
* Webauthn credentials counter ([f7701f6](https://github.com/ory/kratos/commit/f7701f629d5553e229546b00d3c345a8d74dd627))
* **webauthn:** Ensure context is cleaned up after use ([7a8055b](https://github.com/ory/kratos/commit/7a8055be357a64a1f4074fe28b249fbaf05cf519))

### Unclassified

* test(e2e) improve reliability ([763dd00](https://github.com/ory/kratos/commit/763dd0063f3166fad323b25a1b0e7bdf9850e519))
* Correct session godoc ([7108e65](https://github.com/ory/kratos/commit/7108e65447c37cc6f2937083a2a61442e0a43cb8))


# [0.7.6-alpha.1](https://github.com/ory/kratos/compare/v0.7.5-alpha.1...v0.7.6-alpha.1) (2021-09-12)

Resolves further issues in the SDK and release pipeline.





### Code Generation

* Pin v0.7.6-alpha.1 release commit ([8b0d1ee](https://github.com/ory/kratos/commit/8b0d1ee66f1ee2b9f37cd178ac2bcbd8980d6f1d))


# [0.7.5-alpha.1](https://github.com/ory/kratos/compare/v0.7.4-alpha.1...v0.7.5-alpha.1) (2021-09-11)

Primarily resolves issues in the SDK pipeline.





### Code Generation

* Pin v0.7.5-alpha.1 release commit ([3a741a5](https://github.com/ory/kratos/commit/3a741a5ed5cff78e0e060bc98f8526537e8719d7))


# [0.7.4-alpha.1](https://github.com/ory/kratos/compare/v0.7.3-alpha.1...v0.7.4-alpha.1) (2021-09-09)

This release adds the GitHub-app provider, improves SQL instrumentation, resolves an expired flow bug, and resolves documentation issues.





### Bug Fixes

* Corret sdk annotations for enums ([6152363](https://github.com/ory/kratos/commit/6152363cda20992a9b894e618c3a438f30808a97))
* Do not panic if cookiemanager returns a nil cookie ([6ea5678](https://github.com/ory/kratos/commit/6ea56785fa0354d8d9479a699304a4b933d6c294)), closes [#1695](https://github.com/ory/kratos/issues/1695)
* Respect return_to in expired flows ([#1697](https://github.com/ory/kratos/issues/1697)) ([394a8de](https://github.com/ory/kratos/commit/394a8de9c0cdd33df91d56008eac12510ff14e07)), closes [#1251](https://github.com/ory/kratos/issues/1251)

### Code Generation

* Pin v0.7.4-alpha.1 release commit ([67ff8a9](https://github.com/ory/kratos/commit/67ff8a947b5b339648aeb4c22aba89205c61382b))

### Documentation

* Add e2e quickstart ([2b749d3](https://github.com/ory/kratos/commit/2b749d39fcb0d320d193290966a558ee2c5734d1))
* Browser redirects ([#1700](https://github.com/ory/kratos/issues/1700)) ([a44089a](https://github.com/ory/kratos/commit/a44089a506f5ea9daa406fcb862ad707f569c2bb))
* Mark logout_url always available ([9021805](https://github.com/ory/kratos/commit/9021805c4399beb73f234726f8f5f3bfd312482c))
* Minor improvements ([#1707](https://github.com/ory/kratos/issues/1707)) ([79c132c](https://github.com/ory/kratos/commit/79c132c5a0737ea1632655d8aea0af63c4200d37))

### Features

* Making use of the updated instrumentedsql version ([#1723](https://github.com/ory/kratos/issues/1723)) ([9e6fbdd](https://github.com/ory/kratos/commit/9e6fbdd06a75d7207b4801d1148267b3a1a0a0c7))
* **oidc:** Github-app provider ([#1711](https://github.com/ory/kratos/issues/1711)) ([fb1fe8c](https://github.com/ory/kratos/commit/fb1fe8c468bb6f8275618b84c5fa157a314c345f))

### Tests

* **session:** Resolve incorrect assertion ([0531220](https://github.com/ory/kratos/commit/05312203ab12eec44e59dcd9210160f2781a69b4))


# [0.7.3-alpha.1](https://github.com/ory/kratos/compare/v0.7.1-alpha.1...v0.7.3-alpha.1) (2021-08-28)

This patch resolves a regression issue with Facebook login, a memory leak issue introduced by an external dependency, adds a "requires verification" login hook, and improves performance for some endpoints.

Also, Ory Kratos SDKs are now published in individual [GitHub repositories for every language](https://github.com/ory?q=kratos-client).





### Bug Fixes

* Add new message when refresh parameter is true ([#1560](https://github.com/ory/kratos/issues/1560)) ([0525623](https://github.com/ory/kratos/commit/05256232bf85d68e068eece6c883f46a447ba5bd)), closes [#1117](https://github.com/ory/kratos/issues/1117)
* Add session in spa registration if session cook is configured ([#1657](https://github.com/ory/kratos/issues/1657)) ([639a7dd](https://github.com/ory/kratos/commit/639a7dd52d43c57e9708ed3e7360c17d6efde6a5)), closes [#1604](https://github.com/ory/kratos/issues/1604)
* **docs:** Ensure config reference is updated ([f6b3aa4](https://github.com/ory/kratos/commit/f6b3aa45b1f39ca5e9ee7ef4cd96de1970b2ed71)), closes [#1597](https://github.com/ory/kratos/issues/1597)
* Facebook sign in regression ([#1689](https://github.com/ory/kratos/issues/1689)) ([85337bf](https://github.com/ory/kratos/commit/85337bf65af767d7296b14e8fd21bab5c64d23e2)), closes [#1687](https://github.com/ory/kratos/issues/1687) [#1686](https://github.com/ory/kratos/issues/1686)
* Http context memory leak ([b21bd22](https://github.com/ory/kratos/commit/b21bd224059e8a42da9814237572a118297c5210)):

    Ory Kratos was using `gorilla/sessions` prior to version v1.2 which had a dependency on `gorilla/context`, a deprecated library with known memory management issues. Even though we used `gorilla/context`'s clean up middleware, it appears that `r.Context()` was not properly cleaned up, causing memory leaks.
    
    On average, the memory leak is pretty small, but depending on what gets added to `r.Context()` it could significantly increase the memory leak.
    
    By replacing `gorilla/sessions` with v1.2.1 we:
    
    1. Increased the HTTP API throughput by an estimate of 4 times;
    2. Brought average memory use back down to about 12MB;
    
    Closes https://github.com/ory-corp/cloud/issues/1292

* Outdated label ([#1681](https://github.com/ory/kratos/issues/1681)) ([149101e](https://github.com/ory/kratos/commit/149101ed145dae2b75e5150013efc478f5fd0cc3))
* Register argon2 CLI commands properly ([#1592](https://github.com/ory/kratos/issues/1592)) ([45c28d9](https://github.com/ory/kratos/commit/45c28d99064baf8051521a1078ac2b59bb3206ec))
* Remove session cookie on logout ([#1587](https://github.com/ory/kratos/issues/1587)) ([cdb30bb](https://github.com/ory/kratos/commit/cdb30bb65ac932a17e4924b4efc8952113452513)), closes [#1584](https://github.com/ory/kratos/issues/1584):

    Before, the logout endpoint would invalidate the session cookie, but not remove it. This was a regression introduced in 0.7.0. This patch resolves that issue.

* **sdk:** Use proper annotation for genericError ([#1611](https://github.com/ory/kratos/issues/1611)) ([da214b2](https://github.com/ory/kratos/commit/da214b2933ae2a91d8c5bf6aa8eea613a2078b9d)), closes [#1609](https://github.com/ory/kratos/issues/1609)
* Skip prompt on discord authorization by default ([#1594](https://github.com/ory/kratos/issues/1594)) ([a667255](https://github.com/ory/kratos/commit/a6672554b02378eb2dac7b1af99ea2915395867b)):

    When a value for prompt is not provided, Discord defaults to `prompt="consent"`. This change makes it so that if the request is not forced, prompt is explicitly set to "none".

* Static parameter for warning message in config.baseURL(...) ([#1673](https://github.com/ory/kratos/issues/1673)) ([db54a1b](https://github.com/ory/kratos/commit/db54a1bd0c93d7a5845ee09d0a16cbc3b8f26a4a)), closes [#1672](https://github.com/ory/kratos/issues/1672)
* Update csrf token cookie name ([#1601](https://github.com/ory/kratos/issues/1601)) ([64c90bf](https://github.com/ory/kratos/commit/64c90bf5e5cec6545a81f88ad5fabb29e9e80850)):

    See https://github.com/ory-corp/cloud/issues/1252

* Use eager preloading for list identites endpoint ([#1588](https://github.com/ory/kratos/issues/1588)) ([de5fb3e](https://github.com/ory/kratos/commit/de5fb3e52af9f2d0f1209eed217403a5d7d1ae2d))

### Code Generation

* Pin v0.7.3-alpha.1 release commit ([b5ad53e](https://github.com/ory/kratos/commit/b5ad53eca933438126eda3c6c647d99e05e37695))

### Documentation

* Change model to schema ([#1639](https://github.com/ory/kratos/issues/1639)) ([09c403e](https://github.com/ory/kratos/commit/09c403e55482e91a5bfe9a253e514b7a90826709))
* Fix func naming for Logout flow ([#1676](https://github.com/ory/kratos/issues/1676)) ([bbeb613](https://github.com/ory/kratos/commit/bbeb6132ba82e28057bc14bf35ea99b70f0c4118)):

    rename createSelfServiceLogoutUrlForBrowsers  to createSelfServiceLogoutFlowUrlForBrowsers

* Fix stub error example ([#1642](https://github.com/ory/kratos/issues/1642)) ([9bc2fd0](https://github.com/ory/kratos/commit/9bc2fd088ed9b3e7334713e63bae3c7bbcb922db)), closes [#1568](https://github.com/ory/kratos/issues/1568)
* Fixes incorrect yaml identation ([#1641](https://github.com/ory/kratos/issues/1641)) ([6b58278](https://github.com/ory/kratos/commit/6b582784b49c1d103bbf7a6843cdf197fbd93931))
* Identity traits are visible to user ([#1621](https://github.com/ory/kratos/issues/1621)) ([641eba6](https://github.com/ory/kratos/commit/641eba675bdc583661565a6378776bfad26067c6))
* Make qickstart URLs consistent (playground vs. localhost) ([#1626](https://github.com/ory/kratos/issues/1626)) ([bae1847](https://github.com/ory/kratos/commit/bae1847eba0d925f28a010876e35e3c2093bc8c6)):

    Since the quick-start describes how to run Kratos locally the actual location of the redirect is `http://127.0.0.1:4433/self-service/login/browser`.

* Update docker.md - Outdated information ([#1627](https://github.com/ory/kratos/issues/1627)) ([dc32720](https://github.com/ory/kratos/commit/dc32720de25f52b7deb3e32f7530c7827a6ce5df)), closes [#1619](https://github.com/ory/kratos/issues/1619):

    Kratos does not automatically use a config file that exists at `$HOME/.kratos.yaml`, or any other similar pattern. The documentation in the Docker Images section of the guides could lead developers to believe that the --config flag is unnecessary if they are binding the directory the configuration file is in to $HOME or using a custom docker image to provide the file.


### Features

* Allow multiple webhook body sources ([#1606](https://github.com/ory/kratos/issues/1606)) ([51b1311](https://github.com/ory/kratos/commit/51b131177c9e0db018eced939fef43742c9e86cf)):

    This patch adds support for loading webhooks from the local filesystem, base64 encoded inline string, and remote (http/https) sources. Please note that support for relative/absolute paths without an URI scheme are deprecated and will eventually be removed.

* Require verified address ([#1355](https://github.com/ory/kratos/issues/1355)) ([1cf61cd](https://github.com/ory/kratos/commit/1cf61cdeedbd8bf5b66310793249681ff976baab)), closes [#1328](https://github.com/ory/kratos/issues/1328)


# [0.7.1-alpha.1](https://github.com/ory/kratos/compare/v0.7.0-alpha.1...v0.7.1-alpha.1) (2021-07-22)

This release addresses regressions introduced in Ory Kratos v0.7.0 and resolves some bugs and documentation inconsistencies.





### Bug Fixes

* Automatic tagging for node ui ([fe5056e](https://github.com/ory/kratos/commit/fe5056e11d1f8e4355cafa72ed1ff953077181cc)), closes [#1537](https://github.com/ory/kratos/issues/1537)
* Bump kratos ui image for quickstart ([aedbb5a](https://github.com/ory/kratos/commit/aedbb5a259ea8ee63fb06c36fb1c7af78bb63ffc)), closes [#1537](https://github.com/ory/kratos/issues/1537)
* Cleanup lint errors and add doc to x ([#1545](https://github.com/ory/kratos/issues/1545)) ([3cfd784](https://github.com/ory/kratos/commit/3cfd7845730685a4493c2b5d1974b79d873eea86))
* Correct meta schema ([8d4f3ff](https://github.com/ory/kratos/commit/8d4f3ff22d4ade6ae3f923c33303002e5f534cff))
* Do not reset link method ([#1573](https://github.com/ory/kratos/issues/1573)) ([835fb31](https://github.com/ory/kratos/commit/835fb3127bc10b1642b4a7573722e5dce63fedc7))
* Do not set csrf cookies on /sessions/whoami ([#1580](https://github.com/ory/kratos/issues/1580)) ([36bbd43](https://github.com/ory/kratos/commit/36bbd434114d120006d49785787a3c94c7f103f9))
* Export extensionschemas ([#1553](https://github.com/ory/kratos/issues/1553)) ([6af7638](https://github.com/ory/kratos/commit/6af76387caf37160ded75d83dc09ba0bc177a895))
* Generate CSRF token on validation creation ([#1549](https://github.com/ory/kratos/issues/1549)) ([6612c5f](https://github.com/ory/kratos/commit/6612c5f62e5cc242a808032def5714715ce49d11)), closes [#1547](https://github.com/ory/kratos/issues/1547)
* Identity extension meta schema ([#1554](https://github.com/ory/kratos/issues/1554)) ([ba5ca64](https://github.com/ory/kratos/commit/ba5ca642d01917b43d49e009bf140ae13b4f1313)):

    Up until now the extension meta schema was only applied to top level keys. This fix now recursively checks the extension schema on any depth.

* Remove domain alias config constraint ([#1542](https://github.com/ory/kratos/issues/1542)) ([c6145db](https://github.com/ory/kratos/commit/c6145dbfb278369c8e3ad6eae7e8574ed49ba193))
* Resolve wrong openapi types ([b07927c](https://github.com/ory/kratos/commit/b07927cd23cbfce23f3b0676303a2d0ca564143b))
* Update identity state openapi spec ([0217737](https://github.com/ory/kratos/commit/0217737f5a2860e299ccec4387a2cc83aaac1557))
* Use legacy ssl in quickstart config ([6c13c2b](https://github.com/ory/kratos/commit/6c13c2bedd45c10713907e24976658d4a4b88de6)), closes [#1569](https://github.com/ory/kratos/issues/1569)

### Code Generation

* Pin v0.7.1-alpha.1 release commit ([4fe76af](https://github.com/ory/kratos/commit/4fe76af1302d45ddf4cf3c2c5949311c9cf1f8b8))

### Documentation

* Add instruction for creating user  ([#1541](https://github.com/ory/kratos/issues/1541)) ([c2a1b6d](https://github.com/ory/kratos/commit/c2a1b6df95bcb5dfe2b238be5903f483b9e701b5)), closes [#1530](https://github.com/ory/kratos/issues/1530)
* Clarify flags in schema which are not available in config file ([e5ea5fe](https://github.com/ory/kratos/commit/e5ea5fee31eb2f70dc7c33565f791da9e2e87cc2)), closes [#1514](https://github.com/ory/kratos/issues/1514)
* Fix formatting of Email and Phone Verification Flow tab content ([#1536](https://github.com/ory/kratos/issues/1536)) ([0bfac67](https://github.com/ory/kratos/commit/0bfac67a06ef0d96ffd6a487c90edb44d3a40710))
* Fix typo ([#1543](https://github.com/ory/kratos/issues/1543)) ([b25bae7](https://github.com/ory/kratos/commit/b25bae7f2cdcbb60384808041744edd718a2a814))
* Fix typo ([#1544](https://github.com/ory/kratos/issues/1544)) ([547788d](https://github.com/ory/kratos/commit/547788de74794a1dcf43e5190cdfc9d2e1a2dc92))
* Update csrf pitfall flow section ([#1558](https://github.com/ory/kratos/issues/1558)) ([cc7ed4b](https://github.com/ory/kratos/commit/cc7ed4b5f65d2971a45d5d0ec6188908d070d915)), closes [#1557](https://github.com/ory/kratos/issues/1557)

### Tests

* Longer wait time for e2e boot ([3a85a33](https://github.com/ory/kratos/commit/3a85a33ad8a8eec2ebf57d5a47937499141b6bc0))


# [0.7.0-alpha.1](https://github.com/ory/kratos/compare/v0.6.3-alpha.1...v0.7.0-alpha.1) (2021-07-13)

About two months ago we released Ory Kratos v0.6. Today, we are excited to announce the next iteration of Ory Kratos v0.7! This release includes 215 commits from 24 contributors with over 770 files and more than 100.000 lines of code changed!

Ory Kratos v0.7 brings massive developer experience improvements:

- A reworked, tested, and standardized SDK based on OpenAPI 3.0.3 ([#1477](https://github.com/ory/kratos/pull/1477), [#1424](https://github.com/ory/kratos/issues/1424));
- Native support of Single-Page-Apps (ReactJS, AngularJS, ...) for all self-service flows ([#1367](https://github.com/ory/kratos/pull/1367));
- Sign in with Yandex, VK, Auth0, Slack;
- An all-new, secure logout flow ([#1433](https://github.com/ory/kratos/pull/1433));
- Important security updates to the self-service GET APIs ([#1458](https://github.com/ory/kratos/pull/1458), [#1282](https://github.com/ory/kratos/issues/1282));
- Built-in support for TLS ([#1466](https://github.com/ory/kratos/pull/1466));
- Improved documentation and Go Module structure;
- Resolving a case-sensitivity bug in self-service recovery and verification flows;
- Improved performance for listing identities;
- Support for Instant tracing ([#1429](https://github.com/ory/kratos/pull/1429));
- Improved control for SMTPS, supporting SSL and STARTTLS ([#1430](https://github.com/ory/kratos/pull/1430));
- Ability to run Ory Kratos in networks without outbound requests ([#1445](https://github.com/ory/kratos/pull/1445));
- Improved control over HTTP Cookie behavior ([#1531](https://github.com/ory/kratos/pull/1531));
- Several smaller user experience improvements and bug fixes;
- Improved e2e test pipeline.

In the next iteration of Ory Kratos, we will focus on providing a NextJS example application for the SPA integration as well as the long-awaited MFA flows!

Please be aware that upgrading to Ory Kratos 0.7 requires you to apply SQL migrations. Make sure to back up your database before migration!

For more details on breaking changes and patch notes, see below.



## Breaking Changes

Prior to this change it was not possible to specify the verification/recovery link lifetime. Instead, it was bound to the flow expiry. This patch changes that and adds the ability to configure the lifespan of the link individually:

```patch
 selfservice:
   methods:
     link:
       enabled: true
       config:
+        # Defines how long a recovery link is valid for (default 1h)
+        lifespan: 15m
```

This is a breaking change because the link strategy no longer respects the recovery / verification flow expiry time and, unless set, will default to one hour.

This change introduces a better SDK. As part of this change, several breaking changes with regards to the SDK have been introduced. We recommend reading this section carefully to understand the changes and how they might affect you.

Before, the SDK was structured into tags `public` and `admin`. This stems from the fact that we have two ports in Ory Kratos - one administrative and one public port.

While serves as a good overview when working with Ory Kratos, it does not express:

- What module the API belongs to (e.g. self-service, identity, ...)
- What maturity the API has (e.g. experimental, alpha, beta, ...)
- What version the API has (e.g. v0alpha0, v1beta0, ...)

This patch replaces the current `admin` and `public` tags with a versioned approach indicating the maturity of the API used. For example, `initializeSelfServiceSettingsForBrowsers` would no longer be under the `public` tag but instead under the `v0alpha1` tag:

```patch
import {
  Configuration,
- PublicApi
+ V0Alpha1
} from '@ory/kratos-client';

- const kratos = new PublicApi(new Configuration({ basePath: config.kratos.public }));
+ const kratos = new V0Alpha1(new Configuration({ basePath: config.kratos.public }));
```

To avoid confusion when setting up the SDK, and potentially using the wrong endpoints in your codebase and ending up with strange 404 errors, Ory Kratos now redirects you to the correct port, given that `serve.(public|admin).base_url` are configured correctly. This is a significant improvement towards a more robust API experience!

Further, all administrative functions require, in the Ory SaaS, authorization using e.g. an Ory Personal Access Token. In the open source, we do not know what developers use to protect their APIs. As such, we believe that it is ok to have admin and public functions under one common API and differentiate with an `admin` prefix. Therefore, the following patches should be made in your codebase:

```patch
import {
- AdminApi,
+ V0Alpha1,
  Configuration
} from '@ory/kratos-client';

-const kratos = new AdminApi(new Configuration({ basePath: config.kratos.admin }));
+const kratos = new V0Alpha1(new Configuration({ basePath: config.kratos.admin }));

-kratos.createIdentity({
+kratos.adminCreateIdentity({
  schema_id: 'default',
  traits: { /* ... */ }
})
```

Further, we have introduced a [style guide for writing SDKs annotations](https://www.ory.sh/docs/ecosystem/contributing#openapi-spec-and-go-swagger) governing how naming conventions should be chosen.

We also streamlined how credentials are used. We now differentiate between:

- Per-request credentials such as the Ory Session Token / Cookie
    ```
    - public getSelfServiceRegistrationFlow(id: string, cookie?: string, options?: any) {}
    + public getSelfServiceSettingsFlow(id: string, xSessionToken?: string, cookie?: string, options?: any) {}
    ```
- Global credentials such as the Ory (SaaS) Personal Access Token.
    ```typescript
    const kratos = new V0Alpha0(new Configuration({ basePath: config.kratos.admin, accessToken: 'some-token' }));

    kratosAdmin.adminCreateIdentity({
      schema_id: 'default',
      traits: { /* ... */ },
    });
    ```

We hope you enjoy the vastly improved experience! There are still many things that we want to iterate on. For full context, we recommend reading the proposal and discussion around these changes at [kratos#1424](https://github.com/ory/kratos/issues/1424).

Additionally, the Self-Service Error endpoint was updated. First, the endpoint `/self-service/errors` is now located at the public port only with the admin port redirecting to it. Second, the parameter `?error` was renamed to `?id` for better SDK compatibility. Parameter `?error` is still working but will be deprecated at some point. Third, the response no longer contains an error array in `errors` but instead just a single error under `error`:

```patch
{
  "id": "60208346-3a61-4880-96ae-0419cde8fca8",
- "errors": [{
+ "error": {
    "code": 404,
    "status": "Not Found",
    "reason": "foobar",
    "message": "The requested resource could not be found"
- }],
+ },
  "created_at": "2021-07-07T11:20:15.310506+02:00",
  "updated_at": "2021-07-07T11:20:15.310506+02:00"
}
```

This patch introduces CSRF countermeasures for fetching all self-service flows. This ensures that users can not accidentally leak sensitive information when copy/pasting e.g. login URLs (see #1282). If a self-service flow for browsers is requested, the CSRF cookie must be included in the call, regardless if it is a client-side browser app or a server-side browser app calling. This **does not apply** for API-based flows.

As part of this change, the following endpoints have been removed:

- `GET <ory-kratos-admin>/self-service/login/flows`;
- `GET <ory-kratos-admin>/self-service/registration/flows`;
- `GET <ory-kratos-admin>/self-service/verification/flows`;
- `GET <ory-kratos-admin>/self-service/recovery/flows`;
- `GET <ory-kratos-admin>/self-service/settings/flows`.

Please ensure that your server-side applications use the public port (e.g. `GET <ory-kratos-public>/self-service/login/flows`) for fetching self-service flows going forward.

If you use the SDKs, upgrading is easy by adding the `cookie` header when fetching the flows. This is only required when **using browser flows on the server side**.

The following example illustrates a ExpressJS (NodeJS) server-side application fetching the self-service flows.

```patch
app.get('some-route', (req: Request, res: Response) => {
-   kratos.getSelfServiceLoginFlow(flow).then((flow) => /* ... */ )
+   kratos.getSelfServiceLoginFlow(flow, req.header('cookie')).then((flow) => /* ... */ )

-   kratos.getSelfServiceRecoveryFlow(flow).then((flow) => /* ... */ )
+   kratos.getSelfServiceRecoveryFlow(flow, req.header('cookie')).then((flow) => /* ... */ )

-   kratos.getSelfServiceRegistrationFlow(flow).then((flow) => /* ... */ )
+   kratos.getSelfServiceRegistrationFlow(flow, req.header('cookie')).then((flow) => /* ... */ )

-   kratos.getSelfServiceVerificationFlow(flow).then((flow) => /* ... */ )
+   kratos.getSelfServiceVerificationFlow(flow, req.header('cookie')).then((flow) => /* ... */ )

-   kratos.getSelfServiceSettingsFlow(flow).then((flow) => /* ... */ )
+   kratos.getSelfServiceSettingsFlow(flow, undefined, req.header('cookie')).then((flow) => /* ... */ )
})
```

For concrete details, check out [the changes in the NodeJS app](https://github.com/ory/kratos-selfservice-ui-node/commit/e7fa292968111e06401fcfc9b1dd0e8e285a4d87).

This patch refactors the logout functionality for browsers and APIs. It adds increased security and DoS-defenses to the logout flow.

Previously, calling `GET /self-service/browser/flows/logout` would remove the session cookie and redirect the user to the logout endpoint. Now you have to make a call to `GET /self-service/logout/browser` which returns a JSON response including a `logout_url` URL to be used for logout. The call to `/self-service/logout/browser` must be made using AJAX with cookies enabled or by including the Ory Session Cookie in the `X-Session-Cookie` HTTP Header. You may also use the SDK method `createSelfServiceLogoutUrlForBrowsers` to do that.

Additionally, the endpoint `DELETE /sessions` has been moved to `DELETE /self-service/logout/api`. Payloads and responses stay equal. The SDK method `revokeSession` has been renamed to `submitSelfServiceLogoutFlowWithoutBrowser`.

We listened to your feedback and have improved the naming of the SDK method `initializeSelfServiceRecoveryForNativeApps` to better match what it does: `initializeSelfServiceRecoveryWithoutBrowser`. As in the previous release you may still use the old SDK if you do not want to deal with the SDK breaking changes for now.

We listened to your feedback and have improved the naming of the SDK method `initializeSelfServiceVerificationForNativeApps` to better match what it does: `initializeSelfServiceVerificationWithoutBrowser`. As in the previous release you may still use the old SDK if you do not want to deal with the SDK breaking changes for now.

We listened to your feedback and have improved the naming of the SDK method `initializeSelfServiceSettingsForNativeApps` to better match what it does: `initializeSelfServiceSettingsWithoutBrowser`. As in the previous release you may still use the old SDK if you do not want to deal with the SDK breaking changes for now.

We listened to your feedback and have improved the naming of the SDK method `initializeSelfServiceregistrationForNativeApps` to better match what it does: `initializeSelfServiceregistrationWithoutBrowser`. As in the previous release you may still use the old SDK if you do not want to deal with the SDK breaking changes for now.

We listened to your feedback and have improved the naming of the SDK method `initializeSelfServiceLoginForNativeApps` to better match what it does: `initializeSelfServiceLoginWithoutBrowser`. As in the previous release you may still use the old SDK if you do not want to deal with the SDK breaking changes for now.



### Bug Fixes

* Add json detection to setting error subbranches ([fb83dcb](https://github.com/ory/kratos/commit/fb83dcb8ae7463079ddb33c04673cf4556f6058c))
* Add verification success message ([#1526](https://github.com/ory/kratos/issues/1526)) ([126698c](https://github.com/ory/kratos/commit/126698c0b531ca304bb323c825cbeb86b5814f31)), closes [#1450](https://github.com/ory/kratos/issues/1450)
* Cache migration status ([5be2f14](https://github.com/ory/kratos/commit/5be2f149cd79ddfbe8496eccf5d5aacb6a9a0b8e)), closes [#1337](https://github.com/ory/kratos/issues/1337)
* Change SMTP config validation from URI to a Regex pattern ([#1436](https://github.com/ory/kratos/issues/1436)) ([5ab1e8f](https://github.com/ory/kratos/commit/5ab1e8f17bcbc229fada2c584b2c1f576b819761)), closes [#1435](https://github.com/ory/kratos/issues/1435)
* Check filesystem before fallback to bundled templates ([#1401](https://github.com/ory/kratos/issues/1401)) ([22d999e](https://github.com/ory/kratos/commit/22d999e78eb4f67d2f3ba07e62fd28ffb3331d6d))
* Continue button for oidc registration step ([2aad5ac](https://github.com/ory/kratos/commit/2aad5ac8f7055f39f4f434d26fbca74cdbe75337)), closes [#1422](https://github.com/ory/kratos/issues/1422) [#1320](https://github.com/ory/kratos/issues/1320):

    When signing up with an OIDC provider and the traits model is missing some fields, the submit button shows all OIDC options. Instead, it should show just one option called "Continue".

* Deprecate sessionCookie ([#1428](https://github.com/ory/kratos/issues/1428)) ([eccad74](https://github.com/ory/kratos/commit/eccad741a1702181d4b207aad954a950906a808b)), closes [#1426](https://github.com/ory/kratos/issues/1426)
* Do not cache incomplete migrations ([#1434](https://github.com/ory/kratos/issues/1434)) ([154c26f](https://github.com/ory/kratos/commit/154c26f6da4bb7040deabdc352c90cdae42c69fe))
* Do not run network migrations when booting ([12bbab9](https://github.com/ory/kratos/commit/12bbab9d3cf788998cd4a9be50ac8c7a9d2232bd)), closes [#1399](https://github.com/ory/kratos/issues/1399)
* Format test files ([0468aa1](https://github.com/ory/kratos/commit/0468aa19ebfb0f68de5d9d1e59180d953f197cc0))
* Improve identity list performance ([f76886f](https://github.com/ory/kratos/commit/f76886fe7436f71fbef00081888a2f8d0106ba98)), closes [#1412](https://github.com/ory/kratos/issues/1412)
* Incorrect openapi specification for verification submission  ([#1431](https://github.com/ory/kratos/issues/1431)) ([ecb0a01](https://github.com/ory/kratos/commit/ecb0a01f61441aa97751943b5e9ddcc28f783d91)), closes [#1368](https://github.com/ory/kratos/issues/1368)
* Link t docker guide ([953c6d6](https://github.com/ory/kratos/commit/953c6d60f6b6d82ac1406e84c2d87119e63dac48))
* Mark ui node message as optional ([#1365](https://github.com/ory/kratos/issues/1365)) ([7b8d59f](https://github.com/ory/kratos/commit/7b8d59f48ed14a6d0672238645d8675d4bf7fd77)), closes [#1361](https://github.com/ory/kratos/issues/1361) [#1362](https://github.com/ory/kratos/issues/1362)
* Mark verified_at as omitempty ([77b258e](https://github.com/ory/kratos/commit/77b258e57a3d53fe437838a5e9c57805e9c970aa)):

    Closes https://github.com/ory/sdk/issues/46

* Panic if contextualizer is not set ([760035a](https://github.com/ory/kratos/commit/760035a6c5efa08561b93daff57ebb4655032b2a))
* Panic on error in issue session ([5fbd855](https://github.com/ory/kratos/commit/5fbd8557e1f907dd400bfcd26c187db16dc344ba)), closes [#1384](https://github.com/ory/kratos/issues/1384)
* Prometheus metrics fix ([#1299](https://github.com/ory/kratos/issues/1299)) ([ac5d00d](https://github.com/ory/kratos/commit/ac5d00d472a87ab51e7c6834e2cb59f107fc3b3b))
* Recovery email case sensitive ([#1357](https://github.com/ory/kratos/issues/1357)) ([bce14c4](https://github.com/ory/kratos/commit/bce14c487450bd668859f362b98704644fa4c72a)), closes [#1329](https://github.com/ory/kratos/issues/1329)
* Remove changelog ([7affb7a](https://github.com/ory/kratos/commit/7affb7a25bc84082e0ad8096e6c0e4b3933ac5f6))
* Remove obsolete ADD for corp module ([#1455](https://github.com/ory/kratos/issues/1455)) ([0fa3a53](https://github.com/ory/kratos/commit/0fa3a539fbe1ae498434b200c3b636de10d73a7c))
* Remove typing from node.attribute.value ([63a5e08](https://github.com/ory/kratos/commit/63a5e08afab76dafbfe13e6126e165af28492aad)):

    Closes https://github.com/ory/sdk/issues/75
    Closes https://github.com/ory/sdk/issues/74
    Closes https://github.com/ory/sdk/issues/72

* Rename client package for external consumption ([cba8b00](https://github.com/ory/kratos/commit/cba8b00c8b755cc0bdc7818bc9d7390ff3532ce1))
* Resolve build issues on release ([7c265a8](https://github.com/ory/kratos/commit/7c265a8b909dcc07ceeeda546a748ad28ab0c746))
* Resolve driver issues ([47b1c8d](https://github.com/ory/kratos/commit/47b1c8dce57a023e89a2b178bc8a033496ef4ff2))
* Resolve network regression ([8f96b1f](https://github.com/ory/kratos/commit/8f96b1fe4d0846a3ad97a45bc972ece04109289d))
* Resolve network regressions ([8fc52c0](https://github.com/ory/kratos/commit/8fc52c034ed9978c2a04cc66bccc9b795c9bbefa))
* Testhelper regressions ([bf3b04f](https://github.com/ory/kratos/commit/bf3b04fd2c7f9162073cb584d6fb0d59e868ecbf))
* Use correct url in submitSelfServiceVerificationFlow ([ab8a600](https://github.com/ory/kratos/commit/ab8a600080ac0d6a6235806b74c5b9e3dc1c2d60))
* Use local schema URL for sorting UI nodes ([#1449](https://github.com/ory/kratos/issues/1449)) ([a003885](https://github.com/ory/kratos/commit/a0038853f30cd7d139d42d1d4601c8cf49d03934))
* Use session cookie path settings for csrf cookie ([#1493](https://github.com/ory/kratos/issues/1493)) ([c6d08ed](https://github.com/ory/kratos/commit/c6d08edae32fd94877fb58355d3c711460c7d1a2)), closes [#1292](https://github.com/ory/kratos/issues/1292):

    This PR adds configuration option for CSRF cookies and improves the domain alias logic as well as adding tests for it.

* Use STARTTLS for smtps connections ([#1430](https://github.com/ory/kratos/issues/1430)) ([c21bb80](https://github.com/ory/kratos/commit/c21bb80a749df7b224a8ac3f15fa62523a78d805)), closes [#781](https://github.com/ory/kratos/issues/781)
* Version schema ([#1359](https://github.com/ory/kratos/issues/1359)) ([8c4bac7](https://github.com/ory/kratos/commit/8c4bac71674e45e440d916c6c947ed018a8ea29a)), closes [#1331](https://github.com/ory/kratos/issues/1331) [#1101](https://github.com/ory/kratos/issues/1101) [ory/hydra#2427](https://github.com/ory/hydra/issues/2427)

### Code Generation

* Pin v0.7.0-alpha.1 release commit ([53a0e38](https://github.com/ory/kratos/commit/53a0e38c2b5d7003786a8386a9c4cf129acc06aa))

### Code Refactoring

* Corp package ([#1402](https://github.com/ory/kratos/issues/1402)) ([0202dc5](https://github.com/ory/kratos/commit/0202dc57aacc0d48e4c1ee4e68c91654451f63fa))
* Finalize SDK refactoring ([e772641](https://github.com/ory/kratos/commit/e772641f9bcfa462aa5111cf1329a479e3cdff99)), closes [#1424](https://github.com/ory/kratos/issues/1424)
* Identity SDKs ([d8658dc](https://github.com/ory/kratos/commit/d8658dc887a76d82e3cf23386c03b5ebf7053189)), closes [#1477](https://github.com/ory/kratos/issues/1477)
* Improve session sdk ([7207af4](https://github.com/ory/kratos/commit/7207af4cdf6c78dd3f0fd42b6727d7e320d252e6))
* Introduce DefaultContextualizer in corp package ([#1390](https://github.com/ory/kratos/issues/1390)) ([944d045](https://github.com/ory/kratos/commit/944d045aa7fc59eadfdd18951f0d4937b1ea79df)), closes [#1363](https://github.com/ory/kratos/issues/1363)
* Move cleansql to separate package ([7c203dc](https://github.com/ory/kratos/commit/7c203dc8219afe07f180143f832158615b51f60a))
* Openapi.json -> api.json ([6df0de5](https://github.com/ory/kratos/commit/6df0de5d0b4c952576bf9e14c18d521934edd9bb))
* Self-service error APIs ([65c482f](https://github.com/ory/kratos/commit/65c482fba62c2782b03a3b840124eac062499266))

### Documentation

* Add docs for registration SPA flow ([84458f1](https://github.com/ory/kratos/commit/84458f1a9dfe8be6a97bddd832fcc508b60b8498))
* Add go sdk examples ([e948fad](https://github.com/ory/kratos/commit/e948faddce3a1f52df964c701f6ba2a28f5dfe03))
* Add kratos quickstart config notes ([#1490](https://github.com/ory/kratos/issues/1490)) ([2f8094c](https://github.com/ory/kratos/commit/2f8094c50eaf7e1cd964067172adcad407713764))
* Add replit instructions ([8ab8607](https://github.com/ory/kratos/commit/8ab8607dee433f6e708ade296a6c26d0a87d0aae))
* Add tested and running go sdk examples ([3b56bb5](https://github.com/ory/kratos/commit/3b56bb5fd37d0e7d4479967aa0b5721a68a267f2))
* Correct CII badge ([#1447](https://github.com/ory/kratos/issues/1447)) ([048aec3](https://github.com/ory/kratos/commit/048aec39295f0a3534df5e43e3cd7684d4fbd758))
* Fix broken link ([9eaf764](https://github.com/ory/kratos/commit/9eaf764b28f3ca1dae2816d4c0a985c4866c409b))
* Fix building from source ([#1473](https://github.com/ory/kratos/issues/1473)) ([af54d5b](https://github.com/ory/kratos/commit/af54d5bb9e36f90d272d293817f0d6d7eb2e79a8))
* Fix typo in "Sign in/up with ID & assword" ([#1383](https://github.com/ory/kratos/issues/1383)) ([f39739d](https://github.com/ory/kratos/commit/f39739d94e97f20b94630b957371d11294dc8300))
* Mark login endpoints as experimental ([6faf0f6](https://github.com/ory/kratos/commit/6faf0f65bb05bbafdee6b1274a719695fd5b4173))
* Refactor documentation and adopt changes for [#1477](https://github.com/ory/kratos/issues/1477) ([f5e96cd](https://github.com/ory/kratos/commit/f5e96cd5054e734c319ed32992357fcd73ac44a1)), closes [#1472](https://github.com/ory/kratos/issues/1472)
* Remove changelog from docs folder ([5a7e3d8](https://github.com/ory/kratos/commit/5a7e3d83a5fb7f3e6945f37d42abca14d2982e72))
* Resolve build issues ([b51bb55](https://github.com/ory/kratos/commit/b51bb555d829ab020e593a764cbce4c5ba4885a2))
* Resolve typos and docs react issues ([2d640e4](https://github.com/ory/kratos/commit/2d640e4b9b556fd866c29c83564cb1c7702ab9ff))
* Update docs for all flows ([d29ea69](https://github.com/ory/kratos/commit/d29ea69f6bb908b529502030942b1ced52227372))
* Update documentation for plaintext templates ([#1369](https://github.com/ory/kratos/issues/1369)) ([419784d](https://github.com/ory/kratos/commit/419784dd0d4ddc338830ed0d77a7d99f8f440777)), closes [#1351](https://github.com/ory/kratos/issues/1351)
* Update error documentation ([7d83609](https://github.com/ory/kratos/commit/7d8360973a3359bec321a60f4f3a4202ac7d2430))
* Update login flow documentation ([a27de91](https://github.com/ory/kratos/commit/a27de91e9e06f8501ae9cb70446ed0aae5a39f71))
* Update path ([f0384d9](https://github.com/ory/kratos/commit/f0384d9c11085230fd16290c524d22fac6002870))
* Update README.md Go instructions ([#1464](https://github.com/ory/kratos/issues/1464)) ([8db4b4a](https://github.com/ory/kratos/commit/8db4b4a966c5c418cf9d9169b66d7dacff256113))
* Update remaining self service documentation ([bcc6284](https://github.com/ory/kratos/commit/bcc62846297a67216e01e8c31d375d376c1b7cef))
* Update sdk use ([bcb8c06](https://github.com/ory/kratos/commit/bcb8c06ee324c639e548fc06315d9e952f470582))
* Update settings documentation ([258ceaf](https://github.com/ory/kratos/commit/258ceaf84e6ee15b8eee2f203f456f73e7d406d5))
* Use correct path ([#1333](https://github.com/ory/kratos/issues/1333)) ([e401135](https://github.com/ory/kratos/commit/e401135cf415d7e3e6a8ca463dd47e46fe399b33))

### Features

* Add examples for usage of go sdk ([870c2bd](https://github.com/ory/kratos/commit/870c2bd316a3e5b7ce9d526ebf369e41dbea2630))
* Add GetContextualizer ([ac32717](https://github.com/ory/kratos/commit/ac3271742c9c2b968b08dd2b35a5d120c5befcd9))
* Add helper for starting kratos e2e ([#1469](https://github.com/ory/kratos/issues/1469)) ([b9c7674](https://github.com/ory/kratos/commit/b9c7674c30df8200bcd7223c2fa6b058e833bb8a))
* Add instana as possible tracing provider ([#1429](https://github.com/ory/kratos/issues/1429)) ([abe48a9](https://github.com/ory/kratos/commit/abe48a97ee75567979a70f00dd73ff698efcc75d)), closes [#1385](https://github.com/ory/kratos/issues/1385)
* Add redoc ([#1502](https://github.com/ory/kratos/issues/1502)) ([492266d](https://github.com/ory/kratos/commit/492266de9c9b7b775a7b21b5890361380d911da4))
* Add vk and yandex providers to oidc providers and documentation ([#1339](https://github.com/ory/kratos/issues/1339)) ([22a3ef9](https://github.com/ory/kratos/commit/22a3ef98181eb5922cc0f1c016d42ce46732d0a2)), closes [#1234](https://github.com/ory/kratos/issues/1234)
* Anti-CSRF measures when fetching flows ([#1458](https://github.com/ory/kratos/issues/1458)) ([5171557](https://github.com/ory/kratos/commit/51715572ea08f654d1e97d760b9c3d3a9113aa3d)), closes [#1282](https://github.com/ory/kratos/issues/1282)
* Configurable recovery/verification link lifetime ([f80d4e3](https://github.com/ory/kratos/commit/f80d4e3bf7df603b73589dbc6805c69d049921e0))
* Disable HaveIBeenPwned validation when HaveIBeenPwnedEnabled is set to false ([#1445](https://github.com/ory/kratos/issues/1445)) ([44002f4](https://github.com/ory/kratos/commit/44002f4fa93b40a6bb18f1e759bb416d082cec08)), closes [#316](https://github.com/ory/kratos/issues/316):

    This patch introduces an option to disable HaveIBeenPwned checks in environments where outbound network calls are disabled.

* **identities:** Add a state to identities ([#1312](https://github.com/ory/kratos/issues/1312)) ([d22954e](https://github.com/ory/kratos/commit/d22954e2fdb7b2dd5206651b6dd5cf96185a33ba)), closes [#598](https://github.com/ory/kratos/issues/598)
* Improve contextualization in serve/daemon ([f83cd35](https://github.com/ory/kratos/commit/f83cd355422fb4b422f703406473bda914d8419c))
* Include Credentials Metadata in admin api ([#1274](https://github.com/ory/kratos/issues/1274)) ([c8b6219](https://github.com/ory/kratos/commit/c8b62190fca53db4e1b3a4ddb5253fbd2fd46002)), closes [#820](https://github.com/ory/kratos/issues/820)
* Include Credentials Metadata in admin api Missing changes in handler ([#1366](https://github.com/ory/kratos/issues/1366)) ([a71c220](https://github.com/ory/kratos/commit/a71c2208dedac45d32dab578e62a5e3105c8dee0))
* Natively support SPA for login flows ([6ff67af](https://github.com/ory/kratos/commit/6ff67afa8b0fc0a95cec44d3dda2cbc1987b51dd)), closes [#1138](https://github.com/ory/kratos/issues/1138) [#668](https://github.com/ory/kratos/issues/668):

    This patch adds the long-awaited capabilities for natively working with SPAs and AJAX requests. Previously, requests to the `/self-service/login/browser` endpoint would always end up in a redirect. Now, if the `Accept` header is set to `application/json`, the login flow will be returned as JSON instead. Accordingly, changes to the error and submission flow have been made to support `application/json` content types and SPA / AJAX requests.

* Natively support SPA for recovery flows ([5461244](https://github.com/ory/kratos/commit/5461244943286081e13c304a3b38413b8ee6fdf2)):

    This patch adds the long-awaited capabilities for natively working with SPAs and AJAX requests. Previously, requests to the `/self-service/recovery/browser` endpoint would always end up in a redirect. Now, if the `Accept` header is set to `application/json`, the registration flow will be returned as JSON instead. Accordingly, changes to the error and submission flow have been made to support `application/json` content types and SPA / AJAX requests.

* Natively support SPA for registration flows ([57d3c57](https://github.com/ory/kratos/commit/57d3c5786a88f0648e7fa57f181f060a057ec19f)), closes [#1138](https://github.com/ory/kratos/issues/1138) [#668](https://github.com/ory/kratos/issues/668):

    This patch adds the long-awaited capabilities for natively working with SPAs and AJAX requests. Previously, requests to the `/self-service/registration/browser` endpoint would always end up in a redirect. Now, if the `Accept` header is set to `application/json`, the registration flow will be returned as JSON instead. Accordingly, changes to the error and submission flow have been made to support `application/json` content types and SPA / AJAX requests.

* Natively support SPA for settings flows ([ea4395e](https://github.com/ory/kratos/commit/ea4395ed25d5668e4ce365336cd7a5e13e0ba1cc)):

    This patch adds the long-awaited capabilities for natively working with SPAs and AJAX requests. Previously, requests to the `/self-service/settings/browser` endpoint would always end up in a redirect. Now, if the `Accept` header is set to `application/json`, the registration flow will be returned as JSON instead. Accordingly, changes to the error and submission flow have been made to support `application/json` content types and SPA / AJAX requests.

* Natively support SPA for verification flows ([c151500](https://github.com/ory/kratos/commit/c1515009dcd1b5946a93733feedb01753de91c3d)):

    This patch adds the long-awaited capabilities for natively working with SPAs and AJAX requests. Previously, requests to the `/self-service/verification/browser` endpoint would always end up in a redirect. Now, if the `Accept` header is set to `application/json`, the registration flow will be returned as JSON instead. Accordingly, changes to the error and submission flow have been made to support `application/json` content types and SPA / AJAX requests.

* Protect logout against CSRF ([#1433](https://github.com/ory/kratos/issues/1433)) ([1a7a74c](https://github.com/ory/kratos/commit/1a7a74c3fe425f139a87bb68fbc07f8862c00e58)), closes [#142](https://github.com/ory/kratos/issues/142)
* Sign in with Auth0 ([#1352](https://github.com/ory/kratos/issues/1352)) ([f618a53](https://github.com/ory/kratos/commit/f618a53fb971ad16121aa8728cfec54253bb3f44)), closes [#609](https://github.com/ory/kratos/issues/609)
* Support api in settings error ([23105db](https://github.com/ory/kratos/commit/23105dbb836d920b8766536b65de58932f53d6f6))
* Support reading session token from X-Session-Token HTTP header ([dcaefd9](https://github.com/ory/kratos/commit/dcaefd94a0b2cf819424f2e10b3bdae63b256726))
* Team id in slack oidc ([#1409](https://github.com/ory/kratos/issues/1409)) ([e4d021a](https://github.com/ory/kratos/commit/e4d021a037a6b44f8bd66372e9c260c640e87b9d)), closes [#1408](https://github.com/ory/kratos/issues/1408)
* TLS support for public and admin endpoints ([#1466](https://github.com/ory/kratos/issues/1466)) ([7f44f81](https://github.com/ory/kratos/commit/7f44f819a5989a699e403e02c69541369573078f)), closes [#791](https://github.com/ory/kratos/issues/791)
* Update openapi specs and regenerate ([cac507e](https://github.com/ory/kratos/commit/cac507eb5b1f39d003d72e57912dbbfe6f92deb1))

### Tests

* Add tests for cookie behavior of API and browser endpoints ([d1b1521](https://github.com/ory/kratos/commit/d1b15217867cfb92a615c793b26fad288f5e5742))
* **e2e:** Greatly improve test performance ([#1421](https://github.com/ory/kratos/issues/1421)) ([2ffad9e](https://github.com/ory/kratos/commit/2ffad9ee751471451e2151719a2e70d5f89437b0)):

    Instead of running the individual profiles as separate Cypress instances, we now use one singular instance which updates the Ory Kratos configuration depending on the test context. This ensures that hot-reloading is properly working while also signficantly reducing the amount of time spent on booting up the service dependencies.

* **e2e:** Resolve flaky test issues related to timeouts and speed ([b083791](https://github.com/ory/kratos/commit/b083791858bc26a02250d7f5a4e8883cd7392a58))
* **e2e:** Resolve recovery regression ([72c47d6](https://github.com/ory/kratos/commit/72c47d65415efbb53d5d680bd9d78156d577b67f))
* **e2e:** Resolve test config regressions ([eb9c4f9](https://github.com/ory/kratos/commit/eb9c4f98f2e30ac420ed1e3f18a3f0d9ff23846e))
* Remove obsolete console.log ([3ecc869](https://github.com/ory/kratos/commit/3ecc869ebfef5c97334ae4334fb4af98ca9baf97))
* Resolve e2e regressions ([b0d3b82](https://github.com/ory/kratos/commit/b0d3b82f301942bebe3c0027c8b3160749f907af))
* Resolve migratest panic ([89d05ae](https://github.com/ory/kratos/commit/89d05ae0c376c4ea1f23708cccf95c9754a29c94))
* Resolve mobile regressions ([868e82e](https://github.com/ory/kratos/commit/868e82e3d7aec4cde80d7c1d0ce4601e40695f27))
* Resolve oidc regressions ([2403082](https://github.com/ory/kratos/commit/2403082701ac5d667706afd893a6d406496f67fa))

### Unclassified

* add CoC shield (#1439) ([826ed1a](https://github.com/ory/kratos/commit/826ed1a6deafdc2631a5c72f0bfacc91b06a3435)), closes [#1439](https://github.com/ory/kratos/issues/1439)
* u ([b03549b](https://github.com/ory/kratos/commit/b03549b6340ec0bf4f9d741ce145ca90bbc09968))
* u ([318a31d](https://github.com/ory/kratos/commit/318a31d400b97653b4f377c67df4ae0afea189d9))
* Format ([eca7aff](https://github.com/ory/kratos/commit/eca7aff2be96c673dd6be5dc36ab1f4850cc44f0))
* Format ([5cc9fc3](https://github.com/ory/kratos/commit/5cc9fc3a6e91a96225d016d60c8da5cef647ac18))
* Format ([e525805](https://github.com/ory/kratos/commit/e525805246431075d26c3f47596ae93f6580d8ee))
* Format ([4a692ac](https://github.com/ory/kratos/commit/4a692acc7db160068ed7d81461b173bc957e4736))
* Format ([169c0cd](https://github.com/ory/kratos/commit/169c0cd8d424babef69a52ddf65e2b75ded09a46))


# [0.6.3-alpha.1](https://github.com/ory/kratos/compare/v0.6.2-alpha.1...v0.6.3-alpha.1) (2021-05-17)

This release addresses some minor bugs and improves the SDK experience. Please be aware that the Ory Kratos SDK v0.6.3+ have breaking changes compared to Ory Kratos SDK v0.6.2. If you do not wish to update your code, you can keep using the Ory Kratos v0.6.2 SDK and upgrade to v0.6.3+ SDKs at a later stage, as only naming conventions have changed!



## Breaking Changes

Unfortunately, some method signatures have changed in the SDKs. Below is a list of changed entries:

- Error `genericError` was renamed to `jsonError` and now includes more information and better typing for errors;
- The following functions have been renamed:
   - `initializeSelfServiceLoginViaAPIFlow` -> `initializeSelfServiceLoginForNativeApps`
   - `initializeSelfServiceLoginViaBrowserFlow` -> `initializeSelfServiceLoginForBrowsers`
   - `initializeSelfServiceRegistrationViaAPIFlow` -> `initializeSelfServiceRegistrationForNativeApps`
   - `initializeSelfServiceRegistrationViaBrowserFlow` -> `initializeSelfServiceRegistrationForBrowsers`
   - `initializeSelfServiceSettingsViaAPIFlow` -> `initializeSelfServiceSettingsForNativeApps`
   - `initializeSelfServiceSettingsViaBrowserFlow` -> `initializeSelfServiceSettingsForBrowsers`
   - `initializeSelfServiceRecoveryViaAPIFlow` -> `initializeSelfServiceRecoveryForNativeApps`
   - `initializeSelfServiceRecoveryViaBrowserFlow` -> `initializeSelfServiceRecoveryForBrowsers`
   - `initializeSelfServiceVerificationViaAPIFlow` -> `initializeSelfServiceVerificationForNativeApps`
   - `initializeSelfServiceVerificationViaBrowserFlow` -> `initializeSelfServiceVerificationForBrowsers`
- Some type names have changed, for example `traits` -> `identityTraits`.



### Bug Fixes

* Improve settings oas definition ([867abfc](https://github.com/ory/kratos/commit/867abfc813b08142786f71bfe28e373d4754c959))
* Properly handle CSRF for API flows in recovery and verification strategies ([461c829](https://github.com/ory/kratos/commit/461c829dc4d7f7b70620abee2263efba78ce463a)), closes [#1141](https://github.com/ory/kratos/issues/1141)
* **session:** Use specific headers before bearer use ([82c0b54](https://github.com/ory/kratos/commit/82c0b545b29b30fcf3521d9621ec5c5f1a23dc96))
* Use correct api spec path ([5f41f87](https://github.com/ory/kratos/commit/5f41f87bea2919cdf4e9f55c6ad938c5bc08b619))
* Use correct openapi path for validation ([#1340](https://github.com/ory/kratos/issues/1340)) ([a0f5673](https://github.com/ory/kratos/commit/a0f5673d6aa4e60bab06ef699dce231f0bf4aeff))

### Code Generation

* Pin v0.6.3-alpha.1 release commit ([5edf952](https://github.com/ory/kratos/commit/5edf9524d812795ac5712e4a9541b34359234724))

### Code Refactoring

* Improve SDK experience ([71b8511](https://github.com/ory/kratos/commit/71b8511ae1f6f77b2996a01a55accc99d171cfaf)):

    This patch resolves UX issues in the auto-generated SDKs by using consistent naming and introducing a test suite for the Ory SaaS.



# [0.6.2-alpha.1](https://github.com/ory/kratos/compare/v0.6.1-alpha.1...v0.6.2-alpha.1) (2021-05-14)

Resolves an issue in the Go SDK.





### Code Generation

* Pin v0.6.2-alpha.1 release commit ([99c1b1d](https://github.com/ory/kratos/commit/99c1b1d674df3bd8263f7cbf1ed2bdfae6281f69))

### Documentation

* Update link to example email template. ([#1326](https://github.com/ory/kratos/issues/1326)) ([28a1723](https://github.com/ory/kratos/commit/28a17234b557cabf17b592ee68041aec695f6d20))


# [0.6.1-alpha.1](https://github.com/ory/kratos/compare/v0.6.0-alpha.2...v0.6.1-alpha.1) (2021-05-11)

This release primarily addresses issues in the SDK CI pipeline.





### Code Generation

* Pin v0.6.1-alpha.1 release commit ([1df82da](https://github.com/ory/kratos/commit/1df82daaf3f9cfd3a470d7c9bf8d96abbd52b872))

### Features

* Allow changing password validation API DNS name ([#1009](https://github.com/ory/kratos/issues/1009)) ([ced85e8](https://github.com/ory/kratos/commit/ced85e8091b06d864cc55c9975f8b006f6be1ce4))


# [0.6.0-alpha.2](https://github.com/ory/kratos/compare/v0.6.0-alpha.1...v0.6.0-alpha.2) (2021-05-07)

This release addresses issues with the SDK pipeline and also closes a bug related to email sending.





### Bug Fixes

* Update node image ([eef307e](https://github.com/ory/kratos/commit/eef307e6bc33c9ec36ed9138f99c19f72c7be575))

### Code Generation

* Pin v0.6.0-alpha.2 release commit ([a3658ba](https://github.com/ory/kratos/commit/a3658badb848656b61d54b3ee35114972afc1f35))

### Features

* Fix unexpected emails when update profile ([#1300](https://github.com/ory/kratos/issues/1300)) ([7b24485](https://github.com/ory/kratos/commit/7b2448566f82e69d555997654ee410f9b4ff3939)), closes [#1221](https://github.com/ory/kratos/issues/1221)


# [0.6.0-alpha.1](https://github.com/ory/kratos/compare/v0.5.5-alpha.1...v0.6.0-alpha.1) (2021-05-05)

Today Ory Kratos v0.6 has been released! We are extremely happy with this release where we made many changes that pave the path for exciting future additions such as integrating 2FA more easily! We would like to thank the awesome community for the many contributions.

Kratos v0.6 includes an insane amount of work spread over the last five months - 480 commits and over 4200 files changed. The team at Ory would like to thank all the amazing contributors that made this release possible!

Here is a summary of the most important changes:

- Ory Kratos now support highly customizable web hooks - contributed by [@dadrus](https://github.com/dadrus) and [@martinei](https://github.com/martinei);
- Ory Kratos Courier can now be run as a standalone task using `kratos courier watch -c your/config.yaml`. To use the mail courier as a background task of the server run `kratos serve --watch-courier` - contributed by [@mattbonnell](https://github.com/mattbonnell);
- Reworked migrations to ensure stable migrations in production systems - backward compatibility is ensured and tested;
- Upgraded to Go 1.16 and removed all static file packers, greatly improving build time;
- Refactored our SDK pipeline from Swagger 2.0 to OpenAPI Spec 3.0. Ory's SDKs are now properly typed and bugs can easily be addressed using a patch process. Due to this, we had to move away from go-swagger client generation for the Go SDK and replace it with openapi-generator. This, unfortunately, introduced breaking changes in the Go SDK APIs. If you have problems migrating, or have a tutorial on how to migrate, please share it with the community on GitHub!
- Created reliable health and status checks by ensuring that e.g. migrations have completed;
- Made resilient CLI client commands e.g. kratos identities list;
- Better support for cookies in multi-domain setups called [domain aliasing](https://www.ory.sh/kratos/docs/guides/configuring-cookies);
- A new, [dynamically generated FAQ](https://www.ory.sh/kratos/docs/next/faq);
- Enhanced GitHub and Google claims parsing;
- Faster and more resilient CI/CD pipeline;
- Improvements for running Ory Kratos in secure Kubernetes environments;
- Better Helm Charts for Ory Kratos;
- Support for BCrypt hashing, which is now the default hashing implementation. Existing Argon2id hashes will be automatically translated to BCrypt hashes when the user signs in the next time. We recommend using Argon2id in use cases where password hashing is required to take at least 2 seconds. For regular web workloads (200ms) BCrypt is recommended - contributed by [@seremenko-wish](https://github.com/seremenko-wish);
- The Argon2 memory configuration is now human readable: `hashers.argon2.memory: 131072` ->  `hashers.argon2.memory: 131072B` (supports kb, mb, kib, mib, ...).
- Add possibility to keep track of the return_to URLs for verification_flows after sign up using the new `after_verification_return_to` query parameter (e.g. `http://foo.com/registration?after_verification_return_to=verification_callback`) - contributed by [@mattbonnell](https://github.com/mattbonnell);
- Emails are now populated at delivery time, offering more flexibility in terms of templating;
- Emails contain a plaintext variant for email clients that do not display HTML emails - contributed by [@mattbonnell](https://github.com/mattbonnell);
- Mitigation for password hash timing attacks by adding a random delay to login attempts where the user does not exist;
- Resolving SDKs issues for whoami requests;
- Simplified database schema for faster processing, significantly reducing the amount of data stored and latency as several JOINS have been removed;
- Support for binding the HTTP server on UNIX sockets - contributed by [@sloonz](https://github.com/sloonz);

There are even more contributions by [@NickUfer](https://github.com/NickUfer) and [harnash](https://github.com/harnash). In total, [33 people contributed to this release](https://github.com/ory/kratos/graphs/contributors?from=2020-12-09&to=2021-05-04&type=c)! Thank you all!

*IMPORTANT:* Please be aware that the database schema has changed significantly. Applying migrations might, depending on the size of your tables, take a long time. If your database does not support online schema migrations, you will experience downtimes. Please test the migration process before applying it to production!

The probably biggest and most significant change is the refactoring of how self-service flows work and what their payloads look like. This took the most amount of time and introduces the biggest breaking changes in our APIs. We did this refactoring to support several flows planned for Ory Kratos 0.7:

1. Displaying QR codes (images) in login, registration, settings flows - necessary for TOTP 2FA;
2. Asking the login/registration/... UI to render JavaScript - necessary for CAPTCHA, WebAuthN, and more;
3. Refactoring the form submission API to use one endpoint per flow instead of one endpoint per flow per method. This allows us to process several registration/settings/login/... methods such as password + 2FA in one Go.

[Check out how we migrated the NodeJS app](https://github.com/ory/kratos-selfservice-ui-node/commit/53ad90b6c82cde48994feebcc75d754ba74929ec) from the Ory Kratos 0.5 to Ory Kratos 0.6 SDK.

Let's take a look into how these payloads have changed (the flows have identical configuration):

**Ory Kratos v0.5**

*Login*

```json
{
  "id": "ee6e1565-d3c3-4f3a-a6ff-0ba6b3a6481b",
  "type": "browser",
  "expires_at": "2020-09-13T10:49:54.8295242Z",
  "issued_at": "2020-09-13T10:39:54.8295242Z",
  "request_url": "http://127.0.0.1:4433/self-service/login/browser",
  "methods": {
    "password": {
      "method": "password",
      "config": {
        "action": "http://127.0.0.1:4433/self-service/login/methods/password?flow=ee6e1565-d3c3-4f3a-a6ff-0ba6b3a6481b",
        "method": "POST",
        "fields": [
          {
            "name": "identifier",
            "type": "text",
            "required": true,
            "value": ""
          },
          {
            "name": "password",
            "type": "password",
            "required": true
          },
          {
            "name": "csrf_token",
            "type": "hidden",
            "required": true,
            "value": "lNrB8sW2fZY6xnnA91V7ISYrUVcJbmRCOoGHjsnsfI7MsIL5RTbuWFm5TRv1azQW+7IRCfnt2Ch6pC42/45sJQ=="
          }
        ]
      }
    }
  },
  "forced": false
}
```

*Registration*

```json
{
  "id": "2b1f8c5d-e830-4068-97b8-35f776df9217",
  "type": "browser",
  "expires_at": "2020-09-13T10:53:15.1774019Z",
  "issued_at": "2020-09-13T10:43:15.1774019Z",
  "request_url": "http://127.0.0.1:4433/self-service/registration/browser",
  "active": "password",
  "messages": null,
  "methods": {
    "password": {
      "method": "password",
      "config": {
        "action": "http://127.0.0.1:4433/self-service/registration/methods/password?flow=2b1f8c5d-e830-4068-97b8-35f776df9217",
        "method": "POST",
        "fields": [
          {
            "name": "csrf_token",
            "type": "hidden",
            "required": true,
            "value": "1IlHWNjkAZxuYhO82WPgNTgujKsUSaW87j6og/20i2uM4wRTWGSSUg0dJ2fbXa8C5bfM9eTKGdauGwE7y9abwA=="
          },
          {
            "name": "password",
            "type": "password",
            "required": true,
            "messages": [
              {
                "id": 4000005,
                "text": "The password can not be used because the password has been found in at least 23597311 data breaches and must no longer be used..",
                "type": "error",
                "context": {
                  "reason": "the password has been found in at least 23597311 data breaches and must no longer be used."
                }
              }
            ]
          },
          {
            "name": "traits.email",
            "type": "text",
            "value": "foo@ory.sh"
          },
          {
            "name": "traits.name.first",
            "type": "text",
            "value": "Ory"
          },
          {
            "name": "traits.name.last",
            "type": "text",
            "value": "Corp"
          }
        ]
      }
    }
  }
}
```

**Ory Kratos v0.6**

*Login*

As you can see below, the input name `identifier` has changed to `password_identifier`.

```json
{
  "id": "07016811-917d-4788-bb9c-fc297897af6c",
  "type": "browser",
  "expires_at": "2021-04-28T08:37:53.924337873Z",
  "issued_at": "2021-04-28T08:27:53.924337873Z",
  "request_url": "http://127.0.0.1:4433/self-service/login/browser",
  "ui": {
    "action": "http://127.0.0.1:4433/self-service/login?flow=07016811-917d-4788-bb9c-fc297897af6c",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "csrf_token",
          "type": "hidden",
          "value": "IuiHo8fajl6Nwi2CfR33bmC7ZI+geYY44oinK/npkS9gaeV6DlkzS0voYZuyGawsCruvlawFl/pY6/Ph6d9JVg==",
          "required": true,
          "disabled": false
        },
        "messages": null,
        "meta": {}
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "password_identifier",
          "type": "text",
          "value": "",
          "required": true,
          "disabled": false
        },
        "messages": null,
        "meta": {
          "label": {
            "id": 1070004,
            "text": "ID",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "password",
          "type": "password",
          "required": true,
          "disabled": false
        },
        "messages": null,
        "meta": {
          "label": {
            "id": 1070001,
            "text": "Password",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "method",
          "type": "submit",
          "value": "password",
          "disabled": false
        },
        "messages": null,
        "meta": {
          "label": {
            "id": 1010001,
            "text": "Sign in",
            "type": "info",
            "context": {}
          }
        }
      }
    ]
  },
  "forced": false
}
```

*Registration*

```json
{
  "id": "f0c0830a-f5b2-4c2d-a37f-2e70152a4f7c",
  "type": "browser",
  "expires_at": "2021-04-28T08:54:12.951178972Z",
  "issued_at": "2021-04-28T08:44:12.951178972Z",
  "request_url": "http://127.0.0.1:4433/self-service/registration/browser",
  "ui": {
    "action": "http://127.0.0.1:4433/self-service/registration?flow=f0c0830a-f5b2-4c2d-a37f-2e70152a4f7c",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "csrf_token",
          "type": "hidden",
          "value": "408SIAOvpKxW/WbcYfKue26MlLTMbON7T7JT1yhiSemhznD5yiwZuZDXKsWu9vU5BIxfrsAQ8rn10QcdOFSRkA==",
          "required": true,
          "disabled": false
        },
        "messages": null,
        "meta": {}
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.email",
          "type": "email",
          "disabled": false
        },
        "messages": null,
        "meta": {
          "label": {
            "id": 1070002,
            "text": "E-Mail",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "password",
          "type": "password",
          "required": true,
          "disabled": false
        },
        "messages": null,
        "meta": {
          "label": {
            "id": 1070001,
            "text": "Password",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.name.first",
          "type": "text",
          "disabled": false
        },
        "messages": null,
        "meta": {
          "label": {
            "id": 1070002,
            "text": "First Name",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.name.last",
          "type": "text",
          "disabled": false
        },
        "messages": null,
        "meta": {
          "label": {
            "id": 1070002,
            "text": "Last Name",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "method",
          "type": "submit",
          "value": "password",
          "disabled": false
        },
        "messages": null,
        "meta": {
          "label": {
            "id": 1040001,
            "text": "Sign up",
            "type": "info",
            "context": {}
          }
        }
      }
    ]
  }
}
```

These changes are analogous to settings, recovery, verification as well!

We hope you enjoy these new features as much as we do, even if we were not able to deliver 2FA in time for 0.6!

On the last note, Ory Platform, a SaaS is launching in May as early access. It includes Ory Kratos as a managed service and we plan on adding all the other Ory open source technology soon. In our view, Ory is a 10x improvement to the existing "IAM" ecosystem:

1. The major components of Ory Platform are and will remain Apache 2.0 licensed open source. We are *not changing our approach or commitment to open source*. The SaaS model allows us to keep commercialization and open source in harmony;
2. Affordable pricing - Ory does not charge on a per identity basis;
3. Supporting migrations from the Ory Platform (SaaS) to the open-source and vice versa;
4. Offering a planet-scale service with ultra-low latencies no matter where your users are;
5. The largest set of features and APIs of any Identity Product, including Identity and Credentials Management (Ory Kratos), Permissions and Access Control (Ory Keto), Zero-Trust Networking (Ory Oathkeeper), OAuth2, and OpenID Connect (Ory Hydra) plus integrations with Stripe, Mailchimp, Salesforce, and much more.
6. Data aggregation for threat mitigation, auditing, and other use cases (e.g. integration with Snowflake, AWS RedShift, GCP BigQuery, ...)
7. All the advantages of the open source projects - headless, fully customizable, strong security, built with a community;
If you wish to become a part of the preview, please write a short email to [sales@ory.sh](mailto:sales@ory.sh). Early access adopters are also eligible for Ory Hypercare - helping you integrate with Ory fast and designing your security architecture following industry best practices.

Thank you for being a part of our community!



## Breaking Changes

BCrypt is now the default hashing alogrithm. If you wish to continue using Argon2id please set `hashers.algorithm` to `argon2`.

This implies a significant breaking change in the verification flow payload. Please consult the new ui documentation. In essence, the login flow's `methods` key was replaced with a generic `ui` key which provides information for the UI that needs to be rendered.

To apply this patch you must apply SQL migrations. These migrations will drop the flow method table implying that all verification flows that are ongoing will become invalid. We recommend purging the flow table manually as well after this migration has been applied, if you have users doing at least one self-service flow per minute.

This implies a significant breaking change in the recovery flow payload. Please consult the new ui documentation. In essence, the login flow's `methods` key was replaced with a generic `ui` key which provides information for the UI that needs to be rendered.

To apply this patch you must apply SQL migrations. These migrations will drop the flow method table implying that all recovery flows that are ongoing will become invalid. We recommend purging the flow table manually as well after this migration has been applied, if you have users doing at least one self-service flow per minute.

This implies a significant breaking change in the settings flow payload. Please consult the new ui documentation. In essence, the login flow's `methods` key was replaced with a generic `ui` key which provides information for the UI that needs to be rendered.

To apply this patch you must apply SQL migrations. These migrations will drop the flow method table implying that all settings flows that are ongoing will become invalid. We recommend purging the flow table manually as well after this migration has been applied, if you have users doing at least one self-service flow per minute.

This implies a significant breaking change in the registration flow payload. Please consult the new ui documentation. In essence, the login flow's `methods` key was replaced with a generic `ui` key which provides information for the UI that needs to be rendered.

To apply this patch you must apply SQL migrations. These migrations will drop the flow method table implying that all registration flows that are ongoing will become invalid. We recommend purging the flow table manually as well after this migration has been applied, if you have users doing at least one self-service flow per minute.

This implies a significant breaking change in the login flow payload. Please consult the new ui documentation. In essence, the login flow's `methods` key was replaced with a generic `ui` key which provides information for the UI that needs to be rendered.

To apply this patch you must apply SQL migrations. These migrations will drop the flow method table implying that all login flows that are ongoing will become invalid. We recommend purging the flow table manually as well after this migration has been applied, if you have users doing at least one self-service flow per minute.

This change introduces a new feature: UI Nodes. Previously, all self-service flows (login, registration, ...) included form fields (e.g. `methods.password.config.fields`). However, these form fields lacked support for other types of UI elements such as links (for e.g. "Sign in with Google"), images (e.g. QR codes), javascript (e.g. WebAuthn), or text (e.g. recovery codes). With this patch, these new features have been introduced. Please be aware that this introduces significant breaking changes which you will need to adopt to in your UI. Please refer to the most recent documentation to see what has changed. Conceptionally, most things stayed the same - you do however need to update how you access and render the form fields.

Please be also aware that this patch includes SQL migrations which **purge existing self-service forms** from the database. This means that users will need to re-start the login/registration/... flow after the SQL migrations have been applied! If you wish to keep these records, make a back up of your database prior!

This change introduces a new feature: UI Nodes. Previously, all self-service flows (login, registration, ...) included form fields (e.g. `methods.password.config.fields`). However, these form fields lacked support for other types of UI elements such as links (for e.g. "Sign in with Google"), images (e.g. QR codes), javascript (e.g. WebAuthn), or text (e.g. recovery codes). With this patch, these new features have been introduced. Please be aware that this introduces significant breaking changes which you will need to adopt to in your UI. Please refer to the most recent documentation to see what has changed. Conceptionally, most things stayed the same - you do however need to update how you access and render the form fields.

Please be also aware that this patch includes SQL migrations which **purge existing self-service forms** from the database. This means that users will need to re-start the login/registration/... flow after the SQL migrations have been applied! If you wish to keep these records, make a back up of your database prior!

The configuration value for `hashers.argon2.memory` is now a string representation of the memory amount including the unit of measurement. To convert the value divide your current setting (KB) by 1024 to get a result in MB or 1048576 to get a result in GB. Example: `131072` would now become `128MB`.

Co-authored-by: aeneasr <3372410+aeneasr@users.noreply.github.com>
Co-authored-by: aeneasr <aeneas@ory.sh>

Please run SQL migrations when applying this patch.

The following configuration keys were updated:

```patch
selfservice.methods.password.config.max_breaches
```
- `password.max_breaches` -> `selfservice.methods.password.config.max_breaches`
- `password.ignore_network_errors` -> `selfservice.methods.password.config.ignore_network_errors`

After battling with [spf13/viper](https://github.com/spf13/viper) for several years we finally found a viable alternative with [knadh/koanf](https://github.com/knadh/koanf). The complete internal configuration infrastructure has changed, with several highlights:

1. Configuration sourcing works from all sources (file, env, cli flags) with validation against the configuration schema, greatly improving developer experience when changing or updating configuration.
2. Configuration reloading has improved significantly and works flawlessly on Kubernetes.
3. Performance increased dramatically, completely removing the need for a cache layer between the configuration system and ORY Hydra.
4. It is now possible to load several config files using the `--config` flag.
5. Configuration values are now sent to the tracer (e.g. Jaeger) if tracing is enabled.

Please be aware that ORY Kratos might complain about an invalid configuration, because the validation process has improved significantly.



### Bug Fixes

* Add include stub go files ([6d725b1](https://github.com/ory/kratos/commit/6d725b1461a26d99c8b179be8ca219ba83ba0f17))
* Add index to migration status ([8c6ec27](https://github.com/ory/kratos/commit/8c6ec2741535c090aae16f02a744f56c15923e2b))
* Add node_modules to format tasks ([e5f6b36](https://github.com/ory/kratos/commit/e5f6b36caeff080905d15566cf55f8fe4905dbc0))
* Add titles to identity schema ([73c15d2](https://github.com/ory/kratos/commit/73c15d23840aa83d2c99c013cad52ad7df285f18))
* Adopt to new go-swagger changes ([5c45bd9](https://github.com/ory/kratos/commit/5c45bd9f354bfe19b8cbcd7eb4eaebf22c441f42))
* Allow absolute file URLs as config values ([#1069](https://github.com/ory/kratos/issues/1069)) ([4bb4f67](https://github.com/ory/kratos/commit/4bb4f679d1fe0a49edb0c0189bb7a2188d4f850d))
* Allow hashtag in ui urls ([#1040](https://github.com/ory/kratos/issues/1040)) ([7591f07](https://github.com/ory/kratos/commit/7591f07f7d48376a03e9eacfdb6f4a93fd26c0d5))
* Avoid unicode-escaping ampersand in recovery URL query string ([#1212](https://github.com/ory/kratos/issues/1212)) ([d172368](https://github.com/ory/kratos/commit/d17236870af490f043d87e220179b35c9eb2dd4e))
* Bcrypt regression in credentials counting ([23fc13b](https://github.com/ory/kratos/commit/23fc13ba778e0045ca30c00d673ebd6c2f2b7fb7))
* Broken make quickstart-dev task ([#980](https://github.com/ory/kratos/issues/980)) ([999828a](https://github.com/ory/kratos/commit/999828ae036f20bde6d12fe89851e1fde9bdaca6)), closes [#965](https://github.com/ory/kratos/issues/965)
* Broken make sdk task ([#977](https://github.com/ory/kratos/issues/977)) ([5b01c7a](https://github.com/ory/kratos/commit/5b01c7a368c5bcfaa3af218d42f15288f51ab3e4)), closes [#950](https://github.com/ory/kratos/issues/950)
* Call contextualized test helpers ([e1f3f78](https://github.com/ory/kratos/commit/e1f3f7835696b039409c9d05f63665aba7a179ae))
* **cmd:** Make HTTP calls resilient ([e8ed61f](https://github.com/ory/kratos/commit/e8ed61fc3e806453f78b8fa629e96ff7b320bf95))
* Code integer parsing bit size ([#1178](https://github.com/ory/kratos/issues/1178)) ([31e9632](https://github.com/ory/kratos/commit/31e9632bcd6ec3bdeabe862a4cce89021c6dd361)):

    In some cases we had a wrong bitsize of `64`, while the var was later cast to `int`. Replaced with a bitsize of `0`, which is the value to cast to `int`.

* Contextualize identity persister ([f8640c0](https://github.com/ory/kratos/commit/f8640c04f0c5873c39c8af4652d16bfbd347b79e))
* Convert all identifiers to lower case on login ([#815](https://github.com/ory/kratos/issues/815)) ([d64b575](https://github.com/ory/kratos/commit/d64b5757c710c436d6789dbdb33ed04dc11cbdf9)), closes [#814](https://github.com/ory/kratos/issues/814)
* Courier adress ([#1198](https://github.com/ory/kratos/issues/1198)) ([ebe4e64](https://github.com/ory/kratos/commit/ebe4e643150f7603a1e3a3cf6f909135097b3f49)), closes [#1194](https://github.com/ory/kratos/issues/1194)
* Courier message dequeue race condition ([#1024](https://github.com/ory/kratos/issues/1024)) ([5396a82](https://github.com/ory/kratos/commit/5396a82c34eef5d42444b5c4371bd4f820fe3eb0)), closes [#652](https://github.com/ory/kratos/issues/652) [#732](https://github.com/ory/kratos/issues/732):

    Fixes the courier message dequeuing race condition by modifying `*sql.Persister.NextMessages(ctx context.Context, limit uint8)` to retrieve only messages with status `MessageStatusQueued` and update the status of the retrieved messages to `MessageStatusProcessing` within a transaction. On message send failure, the message's status is reset to `MessageStatusQueued`, so that the message can be dequeued in a subsequent `NextMessages` call. On message send success, the status is updated to `MessageStatusSent` (no change there).

* Define credentials types as sql template and resolve crdb issue ([a2d6eeb](https://github.com/ory/kratos/commit/a2d6eeb2928c9750741237f559197fd80494310d))
* Dereference pointer types from new flow structures ([#1019](https://github.com/ory/kratos/issues/1019)) ([efedc92](https://github.com/ory/kratos/commit/efedc920e592bd6e963726e6b123ddc40df93a59))
* Do not include smtp in tracing ([#1268](https://github.com/ory/kratos/issues/1268)) ([bbfcbf9](https://github.com/ory/kratos/commit/bbfcbf9ce595d842a53a3ea21c286d5899eeb28f))
* Do not publish version at public endpoint ([3726ed4](https://github.com/ory/kratos/commit/3726ed4d145a949b25f5b5da5f58d4f448a2a90f))
* Do not reset registration method ([554bb0b](https://github.com/ory/kratos/commit/554bb0b4e62e4ac2a321fa4dbf89ffdf37b188df))
* Do not return system errors for missing identifiers ([1fcc855](https://github.com/ory/kratos/commit/1fcc8557bfee0f7ba562a635670b61dc9acb3530)), closes [#1286](https://github.com/ory/kratos/issues/1286)
* Export mailhog dockertest runner ([1384148](https://github.com/ory/kratos/commit/138414873ad319c6c32c6cc64a73547540dffc74))
* Fix random delay norm distribution math ([#1131](https://github.com/ory/kratos/issues/1131)) ([bd9d28f](https://github.com/ory/kratos/commit/bd9d28fe354710957f4ebaf71d1fffeae3968364))
* Fork audit logger from root logger ([68a09e7](https://github.com/ory/kratos/commit/68a09e7f3dc3ded9a477bb309c68ac8c4e2c2836))
* Gitlab oidc flow ([#1159](https://github.com/ory/kratos/issues/1159)) ([0bb3eb6](https://github.com/ory/kratos/commit/0bb3eb6db1144a09f4ac356cc45e1644d862bb70)), closes [#1157](https://github.com/ory/kratos/issues/1157)
* Give specific message instead of only 404 when method is disabled ([#1025](https://github.com/ory/kratos/issues/1025)) ([2f62041](https://github.com/ory/kratos/commit/2f62041a62588f5b3b062092c57053facb858e62)):

    Enabled strategies are not only used for handlers but also in other areas
    (e.g. populating the flow methods). So we should keep the logic to get
    enabled strategies and add new functions for getting all strategies.

* **hashing:** Make bcrypt default hashing algorithm ([04abe77](https://github.com/ory/kratos/commit/04abe774ada1ef4bf318658fcf84c1d39a2a922d))
* Ignore unset domain aliases ([ada6997](https://github.com/ory/kratos/commit/ada6997ff3dc7e48fd098e40267db5f231a5201f))
* Improve cli error output ([43e9678](https://github.com/ory/kratos/commit/43e967887280b57639565dabd92a07f02fbddeb5))
* Improve error stack trace ([4351773](https://github.com/ory/kratos/commit/43517737109088eda3b1d7f5b42f78bd5eb701d2))
* Improve error tracing ([#1005](https://github.com/ory/kratos/issues/1005)) ([456fd25](https://github.com/ory/kratos/commit/456fd254485fc80b9ae02dfca672a9fea8ae0134))
* Improve test contextualization ([2f92a70](https://github.com/ory/kratos/commit/2f92a7066d72535d32146a98207996fda45e0b96))
* Initialize randomdelay with seeded source ([9896289](https://github.com/ory/kratos/commit/9896289216f10b808a8c78b86d9c27b8d74379de))
* Insert credentials type constants as part of migrations ([#865](https://github.com/ory/kratos/issues/865)) ([92b79b8](https://github.com/ory/kratos/commit/92b79b86762edddf2ad6529b98b3383b641148d5)), closes [#861](https://github.com/ory/kratos/issues/861)
* Linking a connection may result in system error ([#990](https://github.com/ory/kratos/issues/990)) ([be02a70](https://github.com/ory/kratos/commit/be02a70c3cd60adbcc13559e1cb5dc01a8572da4)), closes [#694](https://github.com/ory/kratos/issues/694)
* Marking whoami auhorization parameter as 'in header' ([#1244](https://github.com/ory/kratos/issues/1244)) ([62d8b85](https://github.com/ory/kratos/commit/62d8b85223a0535b07620b08d35c6c3f6b127642)), closes [#1215](https://github.com/ory/kratos/issues/1215)
* Move schema loaders to correct file ([029781f](https://github.com/ory/kratos/commit/029781f69448e8abc85607a03b4bd2055158cf2c))
* Move to new transaction-safe migrations ([#1063](https://github.com/ory/kratos/issues/1063)) ([2588fb4](https://github.com/ory/kratos/commit/2588fb489d76939aeec2986d30fde9075b373831)):

    This patch introduces a new SQL transaction model for running SQL migrations. This fix is particularly targeted at CockroachDB which has limited support for mixing DDL and DML statements. 
    
    Previously it could happen that migrations failure needed manual intervention. This has now been resolved. The new migration model is compatible with the old one and should work without a problem.

* Pass down context to registry ([0879446](https://github.com/ory/kratos/commit/08794461ed95965a9e5460ded2b4c04ab0f5e2e8))
* Re-enable SDK generation ([1d5854d](https://github.com/ory/kratos/commit/1d5854d6298e3d21f85a8fa01d3004166c4b3f50))
* Record cypress runs ([db35d8f](https://github.com/ory/kratos/commit/db35d8ff6bb44dc9e9acf131cb0a14a7f4a7d160))
* Rehydrate settings form on successful submission ([3457e1a](https://github.com/ory/kratos/commit/3457e1a46f48ed79eabff76f8af08b82f12ecc89)), closes [#1305](https://github.com/ory/kratos/issues/1305)
* Remove absolete 'make pack' from Dockerfile ([#1172](https://github.com/ory/kratos/issues/1172)) ([b8eb908](https://github.com/ory/kratos/commit/b8eb908529cc72a3147ad28e4eeee71850a8e431))
* Remove continuity cookies on errors ([85eea67](https://github.com/ory/kratos/commit/85eea6748be6ae8cdfc10cabaa6b677e4efd63eb))
* Remove include stubs ([1764e3a](https://github.com/ory/kratos/commit/1764e3a08a24db82dc391a77fdea09a91faffb5f))
* Remove obsolete clihelpers ([230fd13](https://github.com/ory/kratos/commit/230fd138d1bc7ec57647ea8eeca8e17baaacce0a))
* Remove record from bash script ([84a9315](https://github.com/ory/kratos/commit/84a9315a824cacd29d30b98b65725343af22732d))
* Remove stray non-ctx configs ([#1053](https://github.com/ory/kratos/issues/1053)) ([1fe137e](https://github.com/ory/kratos/commit/1fe137e0d6314bd0af47a29c00e2f72564e71cef))
* Remove trailing double-dot from error ([59581e3](https://github.com/ory/kratos/commit/59581e3fede0fd43028a5f064c350c3cc833b5b0))
* Remove unused sql migration ([1445d1d](https://github.com/ory/kratos/commit/1445d1d1b4b0b5e8ef3426a98ced9573063d8646))
* Remove unused var ([30a8cee](https://github.com/ory/kratos/commit/30a8cee22238d9f400e6d315a9bc99f710945f81))
* Remove verify hook ([98cfec6](https://github.com/ory/kratos/commit/98cfec6d72c2e7bf2db2e8dd6f8875e885923ba8)), closes [#1302](https://github.com/ory/kratos/issues/1302):

    The verify hook is automatically used when verification is enabled and has been removed as a configuration option.

* Replace jwt module ([#1254](https://github.com/ory/kratos/issues/1254)) ([3803c8c](https://github.com/ory/kratos/commit/3803c8ce43e35c51a9c1d7ab55bc662c398cf0d8)), closes [#1250](https://github.com/ory/kratos/issues/1250)
* Resolve build and release issues ([fb582aa](https://github.com/ory/kratos/commit/fb582aa06ad55ca3fd4e2b083e1e9bbb4ba7c715))
* Resolve clidoc issues ([599e9f7](https://github.com/ory/kratos/commit/599e9f773a743f811329cc57cea2748831105e58))
* Resolve compile issues ([63063c1](https://github.com/ory/kratos/commit/63063c15c17f4d3aca96b106275a3478a8ed717e))
* Resolve contextualized table issues ([5a4f0d9](https://github.com/ory/kratos/commit/5a4f0d92800df7fb5ca0df18203a6d73416814e1))
* Resolve crdb migration issue ([9f6edfd](https://github.com/ory/kratos/commit/9f6edfd1f544d5f85e5f5558a08672f40e928136))
* Resolve double hook invokation for registration ([032322c](https://github.com/ory/kratos/commit/032322c66fb6925d8f1473746cb4bfd800d60590))
* Resolve incorrect field types on oidc sign up completion ([f88b6ab](https://github.com/ory/kratos/commit/f88b6abe202605739092a8230fbdebaebcd4407a))
* Resolve lint issues ([0348825](https://github.com/ory/kratos/commit/03488250bcdbfda6ef6a536b4de6117fa8924dc8))
* Resolve lint issues ([75a995b](https://github.com/ory/kratos/commit/75a995b3f69778655611929b65ae22bd77c5370b))
* Resolve linting issues and disable nancy ([c8396f6](https://github.com/ory/kratos/commit/c8396f6007831240d83f77433876c5971a2191ef))
* Resolve mail queue issues ([b968bc4](https://github.com/ory/kratos/commit/b968bc4ed8962d421175adbcaa2dba6eaeea2245))
* Resolve merge regressions ([9862ac7](https://github.com/ory/kratos/commit/9862ac72e0877df4cf17c93e140c354e1ddbd0e7))
* Resolve oidc e2e regressions ([f28087a](https://github.com/ory/kratos/commit/f28087aaf133c116a81213f787dc6f2e982564c0))
* Resolve oidc regressions and e2e tests ([f5091fa](https://github.com/ory/kratos/commit/f5091fac161db0b1401b340a002278bc26891251))
* Resolve potential fsnotify leaks ([3159c0a](https://github.com/ory/kratos/commit/3159c0abe109ea4e3832770278c4e9bc4ca3b3e1))
* Resolve regressions and test failures ([8bae356](https://github.com/ory/kratos/commit/8bae3565ea5410b60c3e638a49f5454fac8e63d3))
* Resolve regressions in cookies and payloads ([9e34bf2](https://github.com/ory/kratos/commit/9e34bf2f6a2f3b007069a5415643c448798207a6))
* Resolve settings sudo regressions ([4b611f3](https://github.com/ory/kratos/commit/4b611f34755369eafcbafa2fc16da13ea3b82370))
* Resolve test regressions ([e3fb028](https://github.com/ory/kratos/commit/e3fb0281dd9be123271d11f2934cfb08fdc470b7))
* Resolve ui issues with nested form objects ([8e744b9](https://github.com/ory/kratos/commit/8e744b931954283cf5f5cbf3ebaca3fa94e035ed))
* Resolve update regression ([d0d661a](https://github.com/ory/kratos/commit/d0d661aaffcba8b039738b773c891ee6e8f6449e))
* Return delay instead of sleeping to improve tests ([27b977e](https://github.com/ory/kratos/commit/27b977ebbaa25b95caa7e3e4536a09ea0bfa61c3))
* Revert generator changes ([c18b97f](https://github.com/ory/kratos/commit/c18b97f333a638d4b4495678013c55faca4b04d0))
* Run correct error handler for registration hooks ([0d80447](https://github.com/ory/kratos/commit/0d80447102d5092e310ca728012f083147c0c5c9))
* Simplify data breaches password error reason ([#1136](https://github.com/ory/kratos/issues/1136)) ([33d29bf](https://github.com/ory/kratos/commit/33d29bf72af03aea77f1d318c19f5087a506719f)):

    This PR simplifies the error reason given when a password has appeared in data breaches to not include the actual number and rather just show "this password has appeared in data breaches and must not be used".

* Support form and json formats in decoder ([d420fe6](https://github.com/ory/kratos/commit/d420fe6e8a491b20063d4bfeaa0a841058087d32))
* Update openapi definitions for signup ([eb0b69d](https://github.com/ory/kratos/commit/eb0b69d50ce834b170186a39bbc9cda4d3366c36))
* Update quickstart node image ([c19b2f4](https://github.com/ory/kratos/commit/c19b2f4c57307e27ce289d44eff34f5aec1341da)):

    See https://github.com/ory/kratos/discussions/1301

* Update to new goreleaser config ([4c2a1b7](https://github.com/ory/kratos/commit/4c2a1b7f5a0059a6e0c28779808ffb27e8910553))
* Update to new healthx ([6ec987a](https://github.com/ory/kratos/commit/6ec987ae81ef0c05f2c4d1eb836c40f9d15950b2))
* Use equalfold ([1c0e52e](https://github.com/ory/kratos/commit/1c0e52ec36ff95b53e3537c5ef457f1c818d7f6b))
* Use new TB interface ([d75a378](https://github.com/ory/kratos/commit/d75a378e700a206753f2cb17032315f2981960e7))
* Use numerical User ID instead of name to avoid k8s security warnings ([#1151](https://github.com/ory/kratos/issues/1151)) ([468a12e](https://github.com/ory/kratos/commit/468a12e56f22cfdf7bd05d68159cc735e75211b2)):

    Our docker image scanner does not allow running processes inside
    container using non-numeric User spec (to determine if we are trying
    to run docker image as root).

* Use remote dependencies ([1e56457](https://github.com/ory/kratos/commit/1e56457d49e1cde69baa41e3111ca113aa49ee3c))

### Code Generation

* Pin v0.6.0-alpha.1 release commit ([507d13a](https://github.com/ory/kratos/commit/507d13a8ec9cd89c9933fc8814a8a99921da69fb))

### Code Refactoring

* Adapt new sdk in testhelpers ([6e15f6f](https://github.com/ory/kratos/commit/6e15f6f86c0f146e846a384ffd6eac78406178bc))
* Add nid everywhere ([407fd95](https://github.com/ory/kratos/commit/407fd95889f416f0d76d6f3f43644a6fafa13b44))
* Contextualize everything ([7ebc3a9](https://github.com/ory/kratos/commit/7ebc3a9a1a2cd85d28c5a9adf2c0c8c10cbd072e)):

    This patch contextualizes all configuration and DBAL models.

* Do not use prefixed node names ([fc42ece](https://github.com/ory/kratos/commit/fc42ece24107dcb6e6a416cc54a2fb5de524fd94))
* Improve Argon2 tooling ([#961](https://github.com/ory/kratos/issues/961)) ([3151187](https://github.com/ory/kratos/commit/315118720419194be8baf5e5e64d7bf190179568)), closes [#955](https://github.com/ory/kratos/issues/955):

    This adds a load testing CLI that allows to adjust the hasher parameters under simulated load.

* Move faker to exportable module ([09f8ae5](https://github.com/ory/kratos/commit/09f8ae5755c9978574e91676bf5df6a23a2feb78))
* Move migratest helpers to ory/x ([7eca67e](https://github.com/ory/kratos/commit/7eca67eb9ec3e4ab065af7221911a74ed16c7c48))
* Move password config to selfservice ([cd0e0eb](https://github.com/ory/kratos/commit/cd0e0ebb0de372ff31c982ef023fe1979addb05a))
* Move to go 1.16 embed ([43c4a13](https://github.com/ory/kratos/commit/43c4a13c25be4a3a23a1ffdbecfaa0f9eda1a11d)):

    This patch replaces packr and pkged with the Go 1.16 embed feature.

* Remove password node attribute prefix ([e27fae4](https://github.com/ory/kratos/commit/e27fae4b0d7a91ff3964804963d4885178b80803))
* Remove profile node attribute prefix ([a3ff6f7](https://github.com/ory/kratos/commit/a3ff6f7eec45b1a9a1e7eb8569793fbc6a047d4f))
* Rename config structs and interfaces ([4a2f419](https://github.com/ory/kratos/commit/4a2f41977439354415118df3e37dd0cde8dac1aa))
* Rename form to container ([5da155a](https://github.com/ory/kratos/commit/5da155a07d3737cefabaf98c4ff650115f662480))
* Replace flow's forms with new ui node module ([647eb1e](https://github.com/ory/kratos/commit/647eb1e66850c67e539d0338cca6cb8ae476ee55))
* Replace flow's forms with new ui node module ([f74a5c2](https://github.com/ory/kratos/commit/f74a5c25af60936b59caee0866a21637a5c0ae6f))
* Replace login flow methods with ui container ([d4ca364](https://github.com/ory/kratos/commit/d4ca364fd8905cfb205ee047a9cb831064a6b9d0))
* Replace recovery flow methods with ui container ([cac0456](https://github.com/ory/kratos/commit/cac04562f2e4e77875275fcfd82c039d787607fb))
* Replace registration flow methods with ui container ([3f6388d](https://github.com/ory/kratos/commit/3f6388d03f91cfad17bd74ebca4d924b4b546668))
* Replace settings flow methods with ui container ([0efd17e](https://github.com/ory/kratos/commit/0efd17e76ba0a0cbd46916a7644b7bdf19bd4ab4))
* Replace verification flow methods with ui container ([dbf2668](https://github.com/ory/kratos/commit/dbf2668747922c93dd967961cd843354afbecfde))
* Replace viper with koanf config management ([5eb1bc0](https://github.com/ory/kratos/commit/5eb1bc0bff7c5d0f83c604484b8e845701112cad))
* Update RegisterFakes calls ([6268310](https://github.com/ory/kratos/commit/626831069ab4f971094ba0bc0b43ac9ff618d91d))
* Use underscore in webhook auth types ([26829d2](https://github.com/ory/kratos/commit/26829d21911cccd4a87c8693b6089af661c1bfe3))

### Documentation

* Add docker to docs main ([8ce8b78](https://github.com/ory/kratos/commit/8ce8b785e2246557253420ea97cf6b7d5ee75d58))
* Add docker to sidebar ([ed38c88](https://github.com/ory/kratos/commit/ed38c88bdbadcdcd2527a2b5270390251742bbe4))
* Add dotnet sdk ([#1183](https://github.com/ory/kratos/issues/1183)) ([32d874a](https://github.com/ory/kratos/commit/32d874a04bb384259aeb544a3fcd6b3a8b23acdd))
* Add faq sidebar ([#1105](https://github.com/ory/kratos/issues/1105)) ([10697aa](https://github.com/ory/kratos/commit/10697aa4ab5dc3e2ab90d1c037dfbe3492bf2bdf))
* Add log docs to schema config ([4967f11](https://github.com/ory/kratos/commit/4967f11d8df177ebdae855eb745e90d21ce38e9f))
* Add more HA docs ([cbb2e27](https://github.com/ory/kratos/commit/cbb2e27f8919a8991c4797a3f1c192ec364f0dd3))
* Add Rust and Dart SDKs ([6d96952](https://github.com/ory/kratos/commit/6d969528e13350ef099669510d3d37df1c007c82)):

    We now support for Rust and Dart SDKs!

* Add SameSite help ([2df6729](https://github.com/ory/kratos/commit/2df6729b4acc70532024658e8874682de64b06b3))
* Add shell-session language ([d16db87](https://github.com/ory/kratos/commit/d16db87802ae2f230a02e4deed189f473588552c))
* Add ui node docs ([e48a07d](https://github.com/ory/kratos/commit/e48a07d03c19a0677d3a56f9e57294b358f24501))
* Adding double colons ([#1187](https://github.com/ory/kratos/issues/1187)) ([fc712f4](https://github.com/ory/kratos/commit/fc712f4530066c429242491c19d1534ffb267b0c))
* Bcrypt is default and add 72 char warning ([29ae53a](https://github.com/ory/kratos/commit/29ae53a96b4472ff549b34241894d72d439c8ea1))
* Better import identities examples ([#997](https://github.com/ory/kratos/issues/997)) ([2e2880a](https://github.com/ory/kratos/commit/2e2880ac057b5c98cd69481c4f6f36b564b5871d))
* Change forum to discussions readme ([#1220](https://github.com/ory/kratos/issues/1220)) ([ae39956](https://github.com/ory/kratos/commit/ae399561ea6ed89aaadd4128bc564254984520e8))
* Describe more about Kratos login/browser flow on quickstart doc ([#1047](https://github.com/ory/kratos/issues/1047)) ([fe725ad](https://github.com/ory/kratos/commit/fe725ad12b5aed5faa8f95bec24ed3aa82512de8))
* Docker file links ([#1182](https://github.com/ory/kratos/issues/1182)) ([4d9b6a3](https://github.com/ory/kratos/commit/4d9b6a3fd5de81310016a811126e40a263ecd27c))
* Document hash timing attack mitigation ([ec86993](https://github.com/ory/kratos/commit/ec869930a9c0e6f6f56c2614835894e0a6a3eaab))
* Explain how to use `after_verification_return_to` ([7e1546b](https://github.com/ory/kratos/commit/7e1546be1fd20baca10507d642d4f209eb88dcbc))
* FAQ improvements ([#1135](https://github.com/ory/kratos/issues/1135)) ([44d0bc9](https://github.com/ory/kratos/commit/44d0bc968a7c0ba5c0793b2349820fa8133bada3))
* FAQ item & minor changes ([#1174](https://github.com/ory/kratos/issues/1174)) ([11cf630](https://github.com/ory/kratos/commit/11cf630082b56c80d12f5915f8e34aa03a7e8c54))
* Fix broken link ([#1037](https://github.com/ory/kratos/issues/1037)) ([6b9aae8](https://github.com/ory/kratos/commit/6b9aae8af5aa3bd614c99b32e341fbd533caf116))
* Fix failing build ([0de328f](https://github.com/ory/kratos/commit/0de328ff0053605e6bded589a79d3ab938d55b31))
* Fix formatting ([#966](https://github.com/ory/kratos/issues/966)) ([687251a](https://github.com/ory/kratos/commit/687251a24e796322b43f8aed6b1fb3d7900e3271))
* Fix identity state bullets ([#1095](https://github.com/ory/kratos/issues/1095)) ([f476334](https://github.com/ory/kratos/commit/f476334c4693277656ad88e768f66b59cbcba126))
* Fix known/unknown email account recovery ([#1211](https://github.com/ory/kratos/issues/1211)) ([e208ca5](https://github.com/ory/kratos/commit/e208ca50ba4f03d5410c9644aaa3b04bdf1b8dbd))
* Fix link ([7f6d7f5](https://github.com/ory/kratos/commit/7f6d7f501d7118dfe6868c9d923fb5ecc5eded48))
* Fix link ([#1128](https://github.com/ory/kratos/issues/1128)) ([e7043e9](https://github.com/ory/kratos/commit/e7043e9b99260eaff2b48ca6f457af46a1521654))
* Fix link to blogpost ([#949](https://github.com/ory/kratos/issues/949)) ([4622e32](https://github.com/ory/kratos/commit/4622e3228fb12231222c7e6b602458111f35f727)), closes [#945](https://github.com/ory/kratos/issues/945)
* Fix link to self-service flows overview ([#995](https://github.com/ory/kratos/issues/995)) ([2be8778](https://github.com/ory/kratos/commit/2be877847644a3df2645ac3be4bbd7704db30b17))
* Fix note block in third party login guide ([#920](https://github.com/ory/kratos/issues/920)) ([745cea0](https://github.com/ory/kratos/commit/745cea02d0e9940f689e668bbd814b29fd53bf37)):

    Allows the document to render properly

* Fix npm links ([#991](https://github.com/ory/kratos/issues/991)) ([4ce4468](https://github.com/ory/kratos/commit/4ce4468132dde21c1692e3a834ad7780bee12b90))
* Fix self-service code flows labels ([#1253](https://github.com/ory/kratos/issues/1253)) ([f2ed424](https://github.com/ory/kratos/commit/f2ed424289cdd2a0edc1736888dd15be6df65f11))
* Fix typo in README ([#1122](https://github.com/ory/kratos/issues/1122)) ([e500707](https://github.com/ory/kratos/commit/e5007078c3cd597cea669827b96c7e6f205f2f32))
* Link to argon2 blogpost and add cross-references ([#1038](https://github.com/ory/kratos/issues/1038)) ([9ab7c3d](https://github.com/ory/kratos/commit/9ab7c3df59ecd94a74a7bf18af9c0ded5305e042))
* Make explicit the ID of the default schema ([#1173](https://github.com/ory/kratos/issues/1173)) ([cc6e9ff](https://github.com/ory/kratos/commit/cc6e9ffbac7118436d85078720cde2de98a68044))
* Minor cosmetics ([#1050](https://github.com/ory/kratos/issues/1050)) ([34db06f](https://github.com/ory/kratos/commit/34db06fd4f83d415c09109b06dfd3b82ce03705e))
* Minor improvements ([#1052](https://github.com/ory/kratos/issues/1052)) ([f0672b5](https://github.com/ory/kratos/commit/f0672b5cb8cca41fa914db21798d20f00a5699f9))
* ORY -> Ory ([ea30979](https://github.com/ory/kratos/commit/ea309797bf59f3da5c5cd184e45f2e585144be56))
* **prometheus:** Update codedoc ([47146ea](https://github.com/ory/kratos/commit/47146ea8ce169ee908aa4d33b59a01e9df4bae10))
* Reformat settings code samples ([cdbbf4d](https://github.com/ory/kratos/commit/cdbbf4df5fa3fa667a78d5cf682bc7fa36693e9d))
* Remove unnecessary and wrong docker pull commands ([#1203](https://github.com/ory/kratos/issues/1203)) ([2b0342a](https://github.com/ory/kratos/commit/2b0342ad7607d705bcebfafd5a78e4e09e57a940))
* Resolve duplication error ([a3d8284](https://github.com/ory/kratos/commit/a3d8284ab20ae76bccba361601b7290af20bdde6))
* Update build from source ([9b5754f](https://github.com/ory/kratos/commit/9b5754f36661f6de9c95f30c06f28164fe5be48b)), closes [#979](https://github.com/ory/kratos/issues/979)
* Update email template docs ([1778cb9](https://github.com/ory/kratos/commit/1778cb9a293feb2c91c0b1921ab78a0395cdca98)), closes [#897](https://github.com/ory/kratos/issues/897)
* Update identity-data-model links ([b5fd9a3](https://github.com/ory/kratos/commit/b5fd9a3a0821215f94da168c9c6f87dceba8c8f4))
* Update identity.ID field documentation ([4624f03](https://github.com/ory/kratos/commit/4624f03a5e9249a5449992a1f0b7ec80dc3499fd)):

    See https://github.com/ory/kratos/discussions/956

* Update kratos video link ([#1073](https://github.com/ory/kratos/issues/1073)) ([e86178f](https://github.com/ory/kratos/commit/e86178f4ee66e5053e0da2fab2c21ecb2e730ada))
* Update login code samples ([695a30f](https://github.com/ory/kratos/commit/695a30f6c80f277676bf04b4665efeb7ea4db618))
* Update login code samples ([ce6c755](https://github.com/ory/kratos/commit/ce6c75587bea80ef83855d764fed79a9d6c948d3))
* Update quickstart samples ([c3fcaba](https://github.com/ory/kratos/commit/c3fcaba65899d9d46a08ca8b60ec0c010f70b16c))
* Update recovery code samples ([d9fbb62](https://github.com/ory/kratos/commit/d9fbb62faff5144f587136935f15d24b6399f29c))
* Update registration code samples ([317810f](https://github.com/ory/kratos/commit/317810ffd8ba6faf87f2248263b6c82cf4e9ffd8))
* Update self-service code samples ([6415011](https://github.com/ory/kratos/commit/6415011ab83a19972c6f52467055fbdcef23a0cc))
* Update settings code samples ([bbd6266](https://github.com/ory/kratos/commit/bbd6266c22097fae195654957cbab589d04892c7))
* Update verification code samples ([4285dec](https://github.com/ory/kratos/commit/4285dec59a8fc31fa3416b594c765f5da9a9de1c))
* Use correct extension for identity-data-model ([acab3e8](https://github.com/ory/kratos/commit/acab3e8b489d9865e4bf0805895f0b7ae9e6f1b8)):

    See https://github.com/ory/kratos/pull/1197#issuecomment-819455322


### Features

* Add email template specification in doc ([#898](https://github.com/ory/kratos/issues/898)) ([4230d9e](https://github.com/ory/kratos/commit/4230d9e0fc35c651b0d2cbdbbf9e1f1c514743f8))
* Add error for when no login strategy was found ([6bae66c](https://github.com/ory/kratos/commit/6bae66cde362c4e2995c9d06a0d3ffee403feb74))
* Add facebook provider to oidc providers and documentation ([#1035](https://github.com/ory/kratos/issues/1035)) ([905bb03](https://github.com/ory/kratos/commit/905bb032520189212bd88f29641903945ae03608)), closes [#1034](https://github.com/ory/kratos/issues/1034)
* Add FAQ to docs ([#1096](https://github.com/ory/kratos/issues/1096)) ([9c6b68c](https://github.com/ory/kratos/commit/9c6b68c454f472b26c34e1975b6a67b24b218f47))
* Add gh login to claims ([49deb2e](https://github.com/ory/kratos/commit/49deb2e166362a5d051bc08523ef44425f144bdd))
* Add login strategy text message ([7468c83](https://github.com/ory/kratos/commit/7468c835d4800c207035897fc9962860d8ab7803))
* Add more tests for multi domain args ([e99803b](https://github.com/ory/kratos/commit/e99803b62a847bcee52bcd87fa8088124b4deae2))
* Add Prometheus monitoring to Public APIs ([#1022](https://github.com/ory/kratos/issues/1022)) ([75a4f1a](https://github.com/ory/kratos/commit/75a4f1a5472ffd780fed43a7395a191ed495c6e9))
* Add random delay to login flow ([#1088](https://github.com/ory/kratos/issues/1088)) ([cb9894f](https://github.com/ory/kratos/commit/cb9894fefc694a4092215d3981e80f287021542f)), closes [#832](https://github.com/ory/kratos/issues/832)
* Add return_url to verification flow ([#1149](https://github.com/ory/kratos/issues/1149)) ([bb99912](https://github.com/ory/kratos/commit/bb99912d823e9bcffa41edf50a01dcae40117fe6)), closes [#1123](https://github.com/ory/kratos/issues/1123) [#1133](https://github.com/ory/kratos/issues/1133)
* Add sql migrations for new login flow ([e947edf](https://github.com/ory/kratos/commit/e947edf497b36bc576061c9ae38049e84ee48575))
* Add sql tracing ([3c4cc1c](https://github.com/ory/kratos/commit/3c4cc1cec170df14331288170a94ada770d3289f))
* Add tracing to config schema ([007dde4](https://github.com/ory/kratos/commit/007dde4482d11f22b8527c94b002da675152a872))
* Add transporter with host modification ([2c41b81](https://github.com/ory/kratos/commit/2c41b81be947f9972638d082105f0f5c83078b91))
* Add workaround template for go openapi ([5d72d10](https://github.com/ory/kratos/commit/5d72d10f6c6948c48c5701fe348084a668c8311a))
* Adds slack sogial login ([#974](https://github.com/ory/kratos/issues/974)) ([7c66053](https://github.com/ory/kratos/commit/7c66053390b3086fe7233625038a78431a61e507)), closes [#953](https://github.com/ory/kratos/issues/953)
* Allow session cookie name configuration ([77ce316](https://github.com/ory/kratos/commit/77ce3162ba97cf5c516c26ef499d9fa892162f0a)), closes [#268](https://github.com/ory/kratos/issues/268)
* Allow specifying sender name in smtp.from_address ([#1100](https://github.com/ory/kratos/issues/1100)) ([5904fe3](https://github.com/ory/kratos/commit/5904fe319f75f8138783434d568db6fc7c55b301))
* Bcrypt algorithm support ([#1169](https://github.com/ory/kratos/issues/1169)) ([b2612ee](https://github.com/ory/kratos/commit/b2612eefbad98d29482d364f670549f470d0a6f5)):

    This patch adds the ability to use BCrypt instead of Argon2id for password hashing. We recommend using BCrypt for web workloads where password hashing should take around 200ms. For workloads where login takes >= 2 seconds, we recommend to continue using Argon2id.
    
    To use bcrypt for password hashing, set your config as follows:
    
     ```
    hashers:
     bcrypt:
        cost: 12
      algorithm: bcrypt
     ```
    
    Switching the hashing algorithm will not break existing passwords!
    
    
    Co-authored-by: Patrik <zepatrik@users.noreply.github.com>

* Check migrations in health check ([c6ef7ad](https://github.com/ory/kratos/commit/c6ef7ad16b70310c645550f7e41b3c8aff847de3))
* Configure domain alias as query param ([9d8563e](https://github.com/ory/kratos/commit/9d8563eeb3293c42cce440ad74f025b304cccbbe))
* Contextualize configuration ([d3d5327](https://github.com/ory/kratos/commit/d3d5327a3622318265a063be4782caa25e645a05))
* Contextualize health checks ([8145a1c](https://github.com/ory/kratos/commit/8145a1c9acaeab441e787118d40ccd448ea82fe4))
* Contextualize http client in cli calls ([3b3ef8f](https://github.com/ory/kratos/commit/3b3ef8f025d75b244d9285036e66f79af7d5ee35))
* Contextualize persitence testers ([6440373](https://github.com/ory/kratos/commit/64403736ad9f8b264567e1f8eed1af710cab6046))
* Courier foreground worker with "kratos courier watch" ([#1062](https://github.com/ory/kratos/issues/1062)) ([500b8ba](https://github.com/ory/kratos/commit/500b8bacd9fd541afd053f42fec66443cfebabda)), closes [#1033](https://github.com/ory/kratos/issues/1033) [#1024](https://github.com/ory/kratos/issues/1024):

    BREACKING CHANGES: This patch moves the courier watcher (responsible for sending mail) to its own foreground worker, which can be executed as a, for example, Kubernetes job.
    
    It is still possible to have the previous behaviour which would run the worker as a background task when running `kratos serve` by using the `--watch-courier` flag.
    
    To run the foreground worker, use `kratos courier watch -c your/config.yaml`.

* **courier:** Allow sending individual messages ([cbb2c0b](https://github.com/ory/kratos/commit/cbb2c0bef63323a177589e9d2a809c84b4f1acdd))
* Do not enforce bcrypt 12 for dev envs ([bbf44d8](https://github.com/ory/kratos/commit/bbf44d887ae5cdb5975516149c74b3ba10896209))
* Email input validation ([#1287](https://github.com/ory/kratos/issues/1287)) ([cd56b73](https://github.com/ory/kratos/commit/cd56b73df363dd37485f07d31fef11fd4d9f40a6)), closes [#1285](https://github.com/ory/kratos/issues/1285)
* Export and add config options ([4391fe5](https://github.com/ory/kratos/commit/4391fe572eb6a766afe9808396847ca5fdca07f5))
* Expose courier worker ([f50969e](https://github.com/ory/kratos/commit/f50969ecba757dea558e9e8b9dd142f5f564d53a))
* Expose crdb ui ([504d518](https://github.com/ory/kratos/commit/504d5181f5e391bb8d67768b314a0348ed252c8b))
* Global docs sidebar ([#1258](https://github.com/ory/kratos/issues/1258)) ([7108262](https://github.com/ory/kratos/commit/71082624e093b8c100e71ae59050f89b35ac20a2))
* Implement and test domain aliasing ([1516a54](https://github.com/ory/kratos/commit/1516a54657df485627251de4e7019bc16353c956)):

    This patch adds a feature called domain aliasing. For more information, head over to http://ory.sh/docs/kratos/next/guides/multi-domain-cookies

* Improve oas spec and fix mobile tests ([4ead2c8](https://github.com/ory/kratos/commit/4ead2c826a2f1a307e327b9736dd8ac99ef52743))
* Improve sorting of ui fields ([797b49d](https://github.com/ory/kratos/commit/797b49d0175280f85f568014cf3083e9bc42d354)):

    See https://github.com/ory/kratos/discussions/1196

* Include schema ([348a493](https://github.com/ory/kratos/commit/348a493c9e5381830b76e57cad803a308e6ce53a))
* Make cli commands consumable in Ory Cloud ([#926](https://github.com/ory/kratos/issues/926)) ([fed790b](https://github.com/ory/kratos/commit/fed790b0f71f028f6d92e8ebceee188dbdb20770))
* Migrate to openapi v3 ([595224b](https://github.com/ory/kratos/commit/595224b1efd5a225702ef236a87f08180a7118b8))
* **oidc:** Support google hd claim ([#1097](https://github.com/ory/kratos/issues/1097)) ([1f20a5c](https://github.com/ory/kratos/commit/1f20a5ceba7682719112d24a3b18bf046fb2ac22))
* Populate email templates at delivery time, add plaintext defaults ([#1155](https://github.com/ory/kratos/issues/1155)) ([7749c7a](https://github.com/ory/kratos/commit/7749c7a75a4386c1fd53db57626355467b698c2f)), closes [#1065](https://github.com/ory/kratos/issues/1065)
* **schema:** Add totp errors ([a61f881](https://github.com/ory/kratos/commit/a61f8814101401dbb422967e37b6c6c1ae85d113))
* Sort and label nodes with easy to use defaults ([cbec27c](https://github.com/ory/kratos/commit/cbec27c957a733411e4c1d511ed5854855b7236e)):

    Ory Kratos takes a guess based on best practices for
    
    - ordering UI nodes (e.g. email, password, submit button)
    - grouping UI nodes (e.g. keep password and oidc nodes together)
    - labeling UI nodes (e.g. "Sign in with GitHub")
    - using the "title" attribute from the identity schema to label trait fields
    
    This greatly simplifies front-end code on your end and makes it even easier to integrate with Ory Kratos! If you want a custom experience with e.g. translations or other things you can always adjust this in your UI integration!

* Support base64 inline schemas ([815a248](https://github.com/ory/kratos/commit/815a24890a118f4128ac083241a93d8df27042f7))
* Support contextual csrf cookies ([957ef38](https://github.com/ory/kratos/commit/957ef38b69fc6ab071b91262736e6c191be3a4b8))
* Support domain aliasing in session cookie ([0681c12](https://github.com/ory/kratos/commit/0681c123f2d856ca27caee645dadc9e6e3731d2c))
* Support label in oidc config ([a99cdcd](https://github.com/ory/kratos/commit/a99cdcddaa0c4bd7b679884b232c2ef8f2dcd978))
* Support retryable CRDB transactions ([f0c21d7](https://github.com/ory/kratos/commit/f0c21d7e0a6ed85818d0e9025a451cb8cbdee086))
* Unix sockets support ([#1255](https://github.com/ory/kratos/issues/1255)) ([ad010de](https://github.com/ory/kratos/commit/ad010de240ddd9219f0cfb2ca3fbb180d2d3a697))
* Web hooks support (recovery) ([#1289](https://github.com/ory/kratos/issues/1289)) ([3e181fe](https://github.com/ory/kratos/commit/3e181fe3d7750a715ab31eb8347fbb4bdb89d6e6)), closes [#271](https://github.com/ory/kratos/issues/271):

    feat: web hooks for self-service flows
    
    This feature adds the ability to define web-hooks using a mixture of configuration and JsonNet. This allows integration with services like Mailchimp, Stripe, CRMs, and all other APIs that support REST requests. Additional to these new changes it is now possible to define hooks for verification and recovery as well!
    
    For more information, head over to the [hooks documentation](https://www.ory.sh/kratos/docs/self-service/hooks).


### Tests

* Add case to ensure correct behavior when verifying a different email address ([#999](https://github.com/ory/kratos/issues/999)) ([f95a117](https://github.com/ory/kratos/commit/f95a117677c9c59436ad10aa8951fe875c39a64f)), closes [#998](https://github.com/ory/kratos/issues/998)
* Add oasis test case ([f80691b](https://github.com/ory/kratos/commit/f80691b9dd77566857c4284e2639cc94d5b8c333))
* Bump poll interval ([b3dc925](https://github.com/ory/kratos/commit/b3dc925a5d43557293745ee81c0ffb3db37b6342))
* Bump video quality ([b7f8d04](https://github.com/ory/kratos/commit/b7f8d042646037e1589ae2d03602bd63a5cec2fe))
* Bump wait times ([b2e43f8](https://github.com/ory/kratos/commit/b2e43f8b0b64784f60e5f57d9a0f5d2928c2b891))
* Clean up hydra env before restart ([cf49414](https://github.com/ory/kratos/commit/cf494149e6a46b15e3b174185e1e87cfcd6f9f7a))
* **e2e:** Significantly reduce wait and idle times ([f525fc5](https://github.com/ory/kratos/commit/f525fc53afec6f5232ce507fe25ddec1b9069196))
* Longer wait times ([4bec9ef](https://github.com/ory/kratos/commit/4bec9ef50f14f22342a311f09ba1b59cde47befc))
* Reliable migration tests on crdb ([2e3764b](https://github.com/ory/kratos/commit/2e3764ba66c156d810de66fba2b0e142dced6f4d))
* Remove old noop test ([16dca3f](https://github.com/ory/kratos/commit/16dca3f78b2021c09ec83e81ab6d2e68c42ca081))
* Resolve compile issues ([c1b5ba4](https://github.com/ory/kratos/commit/c1b5ba42171ec522579df9dfaff27b5b74a1566a))
* Resolve flaky tests ([cb670a8](https://github.com/ory/kratos/commit/cb670a854cbb09b8437bfed7e4a6908ff6dcfd27))
* Resolve json parser test regression ([a1b9b9a](https://github.com/ory/kratos/commit/a1b9b9a95d58583dc7ecf6d2a501da52f84dd6bb))
* Resolve login integration regressions ([388b5b2](https://github.com/ory/kratos/commit/388b5b27d6dee7770e5f37d6d83c532044a4e984))
* Resolve migration regression ([2051a71](https://github.com/ory/kratos/commit/2051a716cb4b8cf334dd65f2ccddb31e5fbed545))
* Resolve more json parser test regressions ([ff791c4](https://github.com/ory/kratos/commit/ff791c41a1d9ce25af4e883469d3f8c0ef9eb302))
* Resolve more regressions ([c5a23af](https://github.com/ory/kratos/commit/c5a23af81427480088651833d904e3403a969fab))
* Resolve order regression ([40a849c](https://github.com/ory/kratos/commit/40a849ca35f4700185322e9ac4f6a4b70132851c))
* Resolve regression ([e2b0ad3](https://github.com/ory/kratos/commit/e2b0ad3c1845da80f078b11b327b9a0376cbb7c5))
* Resolve regression ([f0c9e5f](https://github.com/ory/kratos/commit/f0c9e5ff105d76d6bc9478c98522b2440c7181df))
* Resolve regressions ([4b9da3c](https://github.com/ory/kratos/commit/4b9da3c9d98d40f7b71a56c51543fc115974630d))
* Resolve stub regressions ([82650cf](https://github.com/ory/kratos/commit/82650cf1843f6bfde015f556f4452a7b6fd52b11))
* Resolve test migrations ([de0b65d](https://github.com/ory/kratos/commit/de0b65d96daef0e31c12b3b6915f283a8e71244b))
* Resolve test regression issues ([ccf9fed](https://github.com/ory/kratos/commit/ccf9feddade11f9fcaaf1c37dd3efeb2c4df6649))
* Speed up tests ([a16737c](https://github.com/ory/kratos/commit/a16737cccc36a14444711660f1737913ffd7ba01))
* Update schema tests for webhooks ([d1ddfa8](https://github.com/ory/kratos/commit/d1ddfa80742728b28dc5710ca5b6e7282a2dec55))
* Update test description ([55fb37f](https://github.com/ory/kratos/commit/55fb37f62fc3ab7c0d5324ed31ef3e7f66a73aa2))
* Use bcrypt cost 4 to reduce CI times ([cabe97d](https://github.com/ory/kratos/commit/cabe97d0656858fd1ee0442b40881417e91294f3))
* Use fast bcrypt for e2e ([d90cf13](https://github.com/ory/kratos/commit/d90cf13230632e76eb74965c0945573b4f2e98ff))

### Unclassified

*  fix: resolve clidoc issues (#976) ([346bc73](https://github.com/ory/kratos/commit/346bc73921655d52861b8803eb3351c4205657ee)), closes [#976](https://github.com/ory/kratos/issues/976) [#951](https://github.com/ory/kratos/issues/951)
* :bug: fix ory home directory path (#897) ([2fca2be](https://github.com/ory/kratos/commit/2fca2bedaa907691bef324c11545e007b51d4881)), closes [#897](https://github.com/ory/kratos/issues/897)
* Fix typo in config schema ([16337f1](https://github.com/ory/kratos/commit/16337f13e4388a715c8109c29cf198c82a848a16))
* Format ([e4b7e79](https://github.com/ory/kratos/commit/e4b7e79f4ee91dadfcd008a5b3e318b6bfedad10))
* Format ([193d266](https://github.com/ory/kratos/commit/193d2668ae0955a1346390057539a8b796d17afd))
* Format ([1ebfbde](https://github.com/ory/kratos/commit/1ebfbdea75f27c8eeafa7d3aff45de133ea340bb))
* Format ([ba1eeef](https://github.com/ory/kratos/commit/ba1eeef4f232c4ab59343a2ca3c7cf0eb6dfd110))
* Format ([ada5dbb](https://github.com/ory/kratos/commit/ada5dbb58c45502b8275850a3bc0876debc66888))
* Format ([17a0bf5](https://github.com/ory/kratos/commit/17a0bf5872b33eac615afc675c7d92d7c7441b2e))
* Initial documentation tests via Text-Runner ([#567](https://github.com/ory/kratos/issues/567)) ([c30eb26](https://github.com/ory/kratos/commit/c30eb26f76ab70a6098c0b40c9a04726d36d72f2))


# [0.5.5-alpha.1](https://github.com/ory/kratos/compare/v0.5.4-alpha.1...v0.5.5-alpha.1) (2020-12-09)

The ORY Community is proud to present you the next iteration of ORY Kratos. In this release, we focused on improving production stability!





### Bug Fixes

* CSRF token is required when using the Revoke Session API endpoint ([#839](https://github.com/ory/kratos/issues/839)) ([d3218a0](https://github.com/ory/kratos/commit/d3218a0f23de7293b0a4a966ad21369a92b68b1a)), closes [#838](https://github.com/ory/kratos/issues/838)
* Incorrect home path ([#848](https://github.com/ory/kratos/issues/848)) ([5265af0](https://github.com/ory/kratos/commit/5265af00c92fe505819300caddfcc64004d45c65))
* Make password policy configurable ([#888](https://github.com/ory/kratos/issues/888)) ([7a00483](https://github.com/ory/kratos/commit/7a00483908bb623efdf281e76005c4485ea6b1ab)), closes [#450](https://github.com/ory/kratos/issues/450) [#316](https://github.com/ory/kratos/issues/316):

    Allows configuring password breach thresholds and optionally enforces checks against the HIBP API.

* Remove obsolete types ([#887](https://github.com/ory/kratos/issues/887)) ([b8bac7a](https://github.com/ory/kratos/commit/b8bac7aa56c16cd98f76a95a5e0d01fb1bbde6b7)), closes [#716](https://github.com/ory/kratos/issues/716)
* Set samesite attribute to lax if in dev mode ([#824](https://github.com/ory/kratos/issues/824)) ([91d6698](https://github.com/ory/kratos/commit/91d6698e4ce05ee59bb72fc84b54af9d1d204b41)), closes [#821](https://github.com/ory/kratos/issues/821)
* Use working cache-control header for cdn/proxies/cache ([#869](https://github.com/ory/kratos/issues/869)) ([d8e3d40](https://github.com/ory/kratos/commit/d8e3d40001ffdc64da2288f3cffd53cf3bfdf781)), closes [#601](https://github.com/ory/kratos/issues/601)

### Code Generation

* Pin v0.5.5-alpha.1 release commit ([83aedcb](https://github.com/ory/kratos/commit/83aedcb885acb96c5deb39fff675d5f0528af32d))

### Documentation

* Add contributing to sidebar ([#866](https://github.com/ory/kratos/issues/866)) ([44f33f9](https://github.com/ory/kratos/commit/44f33f97d43f2a3c553a65ebb2986e0731c0e5f2)):

    The same change as in https://github.com/ory/hydra/pull/2209

* Add newsletter to config ([1735ca2](https://github.com/ory/kratos/commit/1735ca2ced104971de4e97524d0a23d57ba045f2))
* Add recovery flow  ([#868](https://github.com/ory/kratos/issues/868)) ([d95cfe9](https://github.com/ory/kratos/commit/d95cfe9759d3ffc08c24048a064c0c800abdf4b4)), closes [#864](https://github.com/ory/kratos/issues/864):

    Added a short section for the recovery flow on managing-user-identities.

* Fix account recovery click instruction ([#870](https://github.com/ory/kratos/issues/870)) ([383de9e](https://github.com/ory/kratos/commit/383de9ecf6f6504dbb9c20fb4cb984e934f0751e))
* Fix broken link ([#893](https://github.com/ory/kratos/issues/893)) ([dec38a2](https://github.com/ory/kratos/commit/dec38a28964aaa13827d356e5bfa12c2a6d1400e)), closes [#835](https://github.com/ory/kratos/issues/835)
* Fix oidc config example structure ([#845](https://github.com/ory/kratos/issues/845)) ([c102a68](https://github.com/ory/kratos/commit/c102a6844db29f994b67d23bb04e64ee71376264))
* Fix redirect ([#802](https://github.com/ory/kratos/issues/802)) ([b868782](https://github.com/ory/kratos/commit/b86878229f343e6b11521596b04040f892d1e2c3))
* Fix typo ([#847](https://github.com/ory/kratos/issues/847)) ([9b3da9f](https://github.com/ory/kratos/commit/9b3da9f0fe2ce71743115844d8c91a1dc9c4cbae))
* Fix typo ([#881](https://github.com/ory/kratos/issues/881)) ([3078293](https://github.com/ory/kratos/commit/3078293717a2ce21c4b939de4c2c4886c75303b5))
* Fix typo MKFA to MFA ([#826](https://github.com/ory/kratos/issues/826)) ([a5613d0](https://github.com/ory/kratos/commit/a5613d08aa21f90f4d192e5663ba4977b3de16c3))
* Remove workaround note ([#886](https://github.com/ory/kratos/issues/886)) ([05409bc](https://github.com/ory/kratos/commit/05409bc13f527398e3de01f29437e5d4353ef8d4)), closes [#718](https://github.com/ory/kratos/issues/718)
* Swagger specs for selfservice settings browser flow ([#825](https://github.com/ory/kratos/issues/825)) ([28d50f4](https://github.com/ory/kratos/commit/28d50f45ab14d561609be7047cac13902394b547))
* Update oidc provider with json conf support ([#833](https://github.com/ory/kratos/issues/833)) ([670eb37](https://github.com/ory/kratos/commit/670eb37d19674f33a36402cd9a88d61ca7327751))

### Features

* Add return_to parameter to logout flow ([#823](https://github.com/ory/kratos/issues/823)) ([1c146dd](https://github.com/ory/kratos/commit/1c146dd21d616a56f510019abadd37402782bb39)), closes [#702](https://github.com/ory/kratos/issues/702)
* Add selinux compatible quickstart config ([#889](https://github.com/ory/kratos/issues/889)) ([0f87948](https://github.com/ory/kratos/commit/0f879481df209ed96b778799adcc2a9424449b37)), closes [#831](https://github.com/ory/kratos/issues/831)

### Tests

* Ensure registration runs only once ([#872](https://github.com/ory/kratos/issues/872)) ([5ffc036](https://github.com/ory/kratos/commit/5ffc036ac82f36ad6ef499e217971275a35fc23a))

### Unclassified

* docs: fix link and typo in Configuring Cookies (#883) ([c51ed6b](https://github.com/ory/kratos/commit/c51ed6b789d2e3a8fe4e93565c3bded37d298f98)), closes [#883](https://github.com/ory/kratos/issues/883)


# [0.5.4-alpha.1](https://github.com/ory/kratos/compare/v0.5.3-alpha.1...v0.5.4-alpha.1) (2020-11-11)

This release introduces the new CLI command `kratos hashers argon2 calibrate 500ms`. This command will choose the best parameterization for Argon2. Check out the [Choose Argon2 Parameters for Secure Password Hashing and Login](https://www.ory.sh/choose-recommended-argon2-parameters-password-hashing/) blog article for more insights!





### Bug Fixes

* Case in settings handler method ([#798](https://github.com/ory/kratos/issues/798)) ([83eb4e0](https://github.com/ory/kratos/commit/83eb4e0021621014d2b543e57a01401381f07fe4))
* Force brew install statement ([#796](https://github.com/ory/kratos/issues/796)) ([ad542ad](https://github.com/ory/kratos/commit/ad542ad5919205ac26a757145474e5a46f3937ec)):

    Closes https://github.com/ory/homebrew-kratos/issues/1


### Code Generation

* Pin v0.5.4-alpha.1 release commit ([b02926c](https://github.com/ory/kratos/commit/b02926c42aee2748bc37ce2600596bd0c2537a0d))

### Code Refactoring

* Move pkger and ioutil helpers to ory/x ([60a0fc4](https://github.com/ory/kratos/commit/60a0fc449d90ead6065ca00926536a989d8b2a2b))

### Documentation

* Fix another broken link ([15bae9f](https://github.com/ory/kratos/commit/15bae9f893c2e2910167326d987455246c110001))
* Fix broken links ([#795](https://github.com/ory/kratos/issues/795)) ([0ab0e7e](https://github.com/ory/kratos/commit/0ab0e7eca8e95d6c26d028c177cbbd1f06b68871)), closes [#793](https://github.com/ory/kratos/issues/793)
* Fix broken relative link ([#812](https://github.com/ory/kratos/issues/812)) ([b32b173](https://github.com/ory/kratos/commit/b32b173fe30b7c5c43700abfa4ddb3409a33556b))
* Fix links ([#800](https://github.com/ory/kratos/issues/800)) ([5fcc272](https://github.com/ory/kratos/commit/5fcc272e625de9e583b2ec24d5679895a6d24c1b))
* Fix oidc config examples ([#799](https://github.com/ory/kratos/issues/799)) ([8a4f480](https://github.com/ory/kratos/commit/8a4f480121995d9899668f037382086fcdd2da4c))
* Fix self-service recovery flow typo ([#807](https://github.com/ory/kratos/issues/807)) ([800110d](https://github.com/ory/kratos/commit/800110d87c9df70a5ec79b58d9fcb9ae39ff76b9))
* Remove duplicate words & fix spelling ([#810](https://github.com/ory/kratos/issues/810)) ([4e1b966](https://github.com/ory/kratos/commit/4e1b96667d9f08dbafeb2f5ce144ca43309de8e0))
* Remove leftover category from reference sidebar ([#813](https://github.com/ory/kratos/issues/813)) ([94fde51](https://github.com/ory/kratos/commit/94fde5101d00b9e1f7228e9d122ef0a8e4719355))
* Use correct links ([#797](https://github.com/ory/kratos/issues/797)) ([a4de293](https://github.com/ory/kratos/commit/a4de29399e4f1b5d0a33acc85478f2d38579a174))

### Features

* Add helper for choosing argon2 parameters ([#803](https://github.com/ory/kratos/issues/803)) ([ca5a69b](https://github.com/ory/kratos/commit/ca5a69b798635d0e5361fd5b0cc369b035dca738)), closes [#723](https://github.com/ory/kratos/issues/723) [#572](https://github.com/ory/kratos/issues/572) [#647](https://github.com/ory/kratos/issues/647):

    This patch adds the new command "hashers argon2 calibrate" which allows one to pick the desired hashing time for password hashing and then chooses the optimal parameters for the hardware the command is running on:
    
    ```
    $ kratos hashers argon2 calibrate 500ms
    Increasing memory to get over 500ms:
        took 2.846592732s in try 0
        took 6.006488824s in try 1
      took 4.42657975s with 4.00GB of memory
    [...]
    Decreasing iterations to get under 500ms:
        took 484.257775ms in try 0
        took 488.784192ms in try 1
      took 486.534204ms with 3 iterations
    Settled on 3 iterations.
    
    {
      "memory": 1048576,
      "iterations": 3,
      "parallelism": 32,
      "salt_length": 16,
      "key_length": 32
    }
    ```



# [0.5.3-alpha.1](https://github.com/ory/kratos/compare/v0.5.2-alpha.1...v0.5.3-alpha.1) (2020-10-27)

This release improves the developer and user experience around CSRF counter-measures. It should now be possible to use the self-service API flows without having to explicitly disable cookie features in your SDKs and integrations. Additionally, another issue in the CGO pipeline was resolved which finally allows running ORY Kratos without CGO if the target database is not SQLite.

Further improvements to default config values have been made and a full end-to-end test suite for the exemplary [kratos-selfservice-ui-react-native](kratos-selfservice-ui-react-native) app. The app is now available in the iTunes store as well - just search for "ORY Profile App"!





### Bug Fixes

* Add "x-session-token" to default allowed headers ([3c912e4](https://github.com/ory/kratos/commit/3c912e4c7d46fd45c00cabb68ed7770bd44f7d07))
* Do not set cookies on api endpoints ([2f67c28](https://github.com/ory/kratos/commit/2f67c28718856ea03ea2effa89b28a8c4b3b8ae0))
* Do not set csrf cookies on potential api endpoints ([4d97a95](https://github.com/ory/kratos/commit/4d97a95d084ea99f5aca158609e197acd256cdd7))
* Ignore unsupported migration dialects ([12bb8d1](https://github.com/ory/kratos/commit/12bb8d14ae1edef18591996411be67d5693e5101)), closes [#778](https://github.com/ory/kratos/issues/778):

    Skips sqlite3 migrations when support is lacking.

* Improve semver regex ([584c0b5](https://github.com/ory/kratos/commit/584c0b5043e85e88ac2648cf699d60fed3e775a9))
* Properly set nosurf context even when ignored ([0dcb774](https://github.com/ory/kratos/commit/0dcb774157bcbfd41a5d9df3914c31162226da75))
* Update cypress ([ba8b172](https://github.com/ory/kratos/commit/ba8b1729477233f79d099e5d7b397430ac1c6ace))
* Use correct regex for version replacement ([ce870ab](https://github.com/ory/kratos/commit/ce870ababdf089344a9428d3a405e18504a3c906)), closes [#787](https://github.com/ory/kratos/issues/787)

### Code Generation

* Pin v0.5.3-alpha.1 release commit ([64dc91a](https://github.com/ory/kratos/commit/64dc91af54cdf3eba158a50690240cdc8f7cb43b))

### Documentation

* Fix docosaurus admonitions ([#788](https://github.com/ory/kratos/issues/788)) ([281a7c9](https://github.com/ory/kratos/commit/281a7c9289570d4bee33447655281b610cbe7e52))
* Pin download script version ([e4137a6](https://github.com/ory/kratos/commit/e4137a6a41d68b1480af2075bda8c5f46c42cd22))
* Remove trailing garbage from quickstart ([#787](https://github.com/ory/kratos/issues/787)) ([7e70924](https://github.com/ory/kratos/commit/7e709242ada28b7781c6ace272f60f9d1b9d5b2f))

### Features

* Improve makefile install process and update deps ([d1eb37f](https://github.com/ory/kratos/commit/d1eb37f5d9d0f16e7864b5f8f08a44ba80853fa5))

### Tests

* Add e2e tests for mobile ([d481d51](https://github.com/ory/kratos/commit/d481d51f5f4de96cbbc7c347f5dbff381b44462d))
* Add option to disable csrf protection in apis ([a0077f1](https://github.com/ory/kratos/commit/a0077f12adf94ff428b502b69bbb0eaafd05be66))
* Bump wait time ([7a719e1](https://github.com/ory/kratos/commit/7a719e17c5641f4df47314f6f0ac2cf73dddc8bb))
* Install expo-cli globally ([db21cfa](https://github.com/ory/kratos/commit/db21cfa1c589a2dab829a4c8eaf1db15d14d965e))
* Install expo-cli in cci config with sudo ([d255f46](https://github.com/ory/kratos/commit/d255f462402f2d2c2278dcba1a139d0064343b22))
* Log wait-on output ([62b5ba9](https://github.com/ory/kratos/commit/62b5ba92d56e9f6b98adb8fb9c4daff03be08f2e))
* Output web server address ([cb41ca7](https://github.com/ory/kratos/commit/cb41ca78367b1943d230fa9ac116fcf3cf69b1c1))
* Resolve csrf test issues in settings ([ef8ba7d](https://github.com/ory/kratos/commit/ef8ba7dc93d6ba84f22b7aa65d00797e33b520a3))
* Resolve test panic ([6f6461f](https://github.com/ory/kratos/commit/6f6461fe3690576015ded9146c065a1e5d950be1))
* Revert delay increase and improve install scripts ([1eafcaa](https://github.com/ory/kratos/commit/1eafcaa86be194e412b0470a759bff6afc6c21af))


# [0.5.2-alpha.1](https://github.com/ory/kratos/compare/v0.5.1-alpha.1...v0.5.2-alpha.1) (2020-10-22)

This release addresses bugs and user experience issues.





### Bug Fixes

* Add debug quickstart yml ([#780](https://github.com/ory/kratos/issues/780)) ([16e6b4d](https://github.com/ory/kratos/commit/16e6b4d76d297182ea9a1f5dc6367570f02f7b42))
* Gracefully handle double slashes in URLs ([aeb9414](https://github.com/ory/kratos/commit/aeb941477910b5ab54429a6aab7a3e1e388c48c5)), closes [#779](https://github.com/ory/kratos/issues/779)
* Merge gobuffalo CGO fix ([fea2e77](https://github.com/ory/kratos/commit/fea2e77ca0f9b20185c7a7704854fdcf29b7ab33))
* Remove obsolete recovery_token and add link to schema ([acf6ac4](https://github.com/ory/kratos/commit/acf6ac4e11c755e56c7d40728088257de367f7ff))
* Return correct error in login csrf ([dd9cab0](https://github.com/ory/kratos/commit/dd9cab0e02400c88e89877f755f03c6179013123)), closes [#785](https://github.com/ory/kratos/issues/785)
* Use correct assert package ([76be5b0](https://github.com/ory/kratos/commit/76be5b0a5d94c251f5f07eee9f700ec11b341e2e))

### Code Generation

* Pin v0.5.2-alpha.1 release commit ([79fcd8a](https://github.com/ory/kratos/commit/79fcd8a6949886f847f7be0c9ba2aba7554ab204))

### Documentation

* Small improvements to discord oidc provider guide ([#783](https://github.com/ory/kratos/issues/783)) ([6a3c453](https://github.com/ory/kratos/commit/6a3c45330885eb95015fa7ee9b58a72c38132499))

### Tests

* Add tests for csrf behavior ([48993e2](https://github.com/ory/kratos/commit/48993e2c496fb8af7e7b9e2752ba7078a134a75a)), closes [#785](https://github.com/ory/kratos/issues/785)
* Mark link as enabled in e2e test ([c214b81](https://github.com/ory/kratos/commit/c214b81a7026b06aaca062b2aa77951d01b0e237))
* Resolve schema test regression ([bb7af1b](https://github.com/ory/kratos/commit/bb7af1b759d6c812755956ef872bcbd31b9c50be))


# [0.5.1-alpha.1](https://github.com/ory/kratos/compare/v0.5.0-alpha.1...v0.5.1-alpha.1) (2020-10-20)

This release resolves an issue where ORY Kratos Docker Images without CGO and SQLite support would fail to boot even when SQLite was not used as a data source.





### Bug Fixes

* Do not require sqlite without build tag ([2ee787b](https://github.com/ory/kratos/commit/2ee787bc1e97bdc11d0c92d55664d59e777f7ed1))
* Use extra dc config file for quickstart-dev ([72c03f9](https://github.com/ory/kratos/commit/72c03f9bcb91d30d5ff6b94030f2cbb6144fbf8d))

### Code Generation

* Pin v0.5.1-alpha.1 release commit ([b85b36b](https://github.com/ory/kratos/commit/b85b36b967d91c13b6d70ed668f17d3474eafae7))

### Documentation

* Fix spelling mistake ([14e7f65](https://github.com/ory/kratos/commit/14e7f6535e69f4bee2e3ca611a8d1a36bfd5f8f8))
* Fix spelling mistake ([#772](https://github.com/ory/kratos/issues/772)) ([bf401a2](https://github.com/ory/kratos/commit/bf401a26ee4422a8ea1b52f642885b0d8bac1272))
* Improve schemas ([#773](https://github.com/ory/kratos/issues/773)) ([e614859](https://github.com/ory/kratos/commit/e6148590577e1688d58534b8559d3bc602f9c2e7))

### Features

* Auto-update docker and git tags on release ([08084a9](https://github.com/ory/kratos/commit/08084a987501939544da1a1c7ee102819e2480ce))
* Use fixed versions for docker-compose ([e73c4ce](https://github.com/ory/kratos/commit/e73c4ce6f328376ad310b8f6d5c391ea06573003))

### Tests

* Increase waittime ([5e911d6](https://github.com/ory/kratos/commit/5e911d687247e4878bdcf82e5b008617f0bbdf4e))
* Reduce flakes by increasing wait time for expiry test ([cddf29e](https://github.com/ory/kratos/commit/cddf29e7dc5304c497d5ba7c1e6a2d63c9b6c137))

### Unclassified

* Format ([8be02c8](https://github.com/ory/kratos/commit/8be02c8938769dfcd7c9b7ed5e72e4ded3b1924b))


# [0.5.0-alpha.1](https://github.com/ory/kratos/compare/v0.4.6-alpha.1...v0.5.0-alpha.1) (2020-10-15)

The ORY team and community is very proud to present the next ORY Kratos iteration!

ORY Kratos is now capable of handling native (iOS, Android, Windows, macOS, ...) login, registration, settings, recovery, and verification flows. As a goodie on top, we released a reference React Native application which you can find on [GitHub](http://github.com/ory/kratos-selfservice-ui-react-native).

We co-released our reference React Native application which acts as a reference on implementing these flows:

![Registration](http://ory.sh/images/newsletter/kratos-0.5.0/registration-screen.png)

![Welcome](http://ory.sh/images/newsletter/kratos-0.5.0/welcome-screen.png)

![Settings](http://ory.sh/images/newsletter/kratos-0.5.0/settings-screen.png)

In total, almost 1200 files were changed in about 480 commits. While you can find a list of all changes in the changelist below, these are the changes we are most proud of:

- We renamed login, registration, ... requests to "flows" consistently across the code base, APIs, and data storage. We now:
  - Initiate a login, registration, ... flow;
  - Fetch a login, registration, ... flow; and
  - Complete a login, registration, ... flow using a login flow method such as "Log in with username and password".
- All self-service flows are now capable of handling API-based requests that do not originate from Browser such as Chrome. This is set groundwork for handling native flows (see above)!
- The self service documentation has been refactored and simplified. We added code samples, screenshots, payloads, and curl commands to make things easier and clearer to understand. Video guides have also been added to help you and the community get things done faster!
- Documentation for rotating important secrets such as the cookie and session secrets was added.
- The need for reverse proxies was removed by adding the ability to change the ORY Kratos Session Cookie domain and path! The [kratos-selfservice-ui-node](https://github.com/ory/kratos-selfservice-ui-node) reference implementation no longer requires HTTP Request piping which greatly simplifies the network layout and codebase!
- The ORY Kratos CLI is now capable of managing identities with an interface that works almost like the Docker CLI we all love!
- Admins are now able to initiate account recovery for identities.
- Email verification and account recovery were refactored. It is now possible to add additional strategies (e.g. recovery codes) in the future, greatly increasing the feature set and security capabilities of future ORY Kratos versions!
- Lookup to Have I Been Pwnd is no longer a hard requirement, allowing registration processes to complete when the service is unavailable or the network is slow.
- We contributed several issues and features in upstream projects such as justinas/nosurf, gobuffalo/pop, and many more!
- The build pipeline has been upgraded to support cross-compilation of CGO with Go 1.15+.
- Fetching flows no longer requires CSRF cookies to be set, improving developer experience while not compromising on security!
- ORY Kratos now has ORY Kratos Session Cookies (set in the HTTP Cookie header) and ORY Kratos Session Tokens (set as a HTTP Bearer Authorization token or the `X-Session-Token` HTTP Header).

Additionally tons of bugs were fixed, tests added, documentation improved, and much more. Please note that several things have changed in a breaking fashion. You can find details for the individual breaking changes in the changelog below.

We would like to thank all community members who contributed towards this release (in no particular order):

- https://github.com/kevgo
- https://github.com/NickUfer
- https://github.com/drwatsno
- https://github.com/alsuren
- https://github.com/wezzle
- https://github.com/sherbang
- https://github.com/perryao
- https://github.com/jikunchong
- https://github.com/err0r500
- https://github.com/debrutal
- https://github.com/c0depwn
- https://github.com/aschepis
- https://github.com/jakhog

Have fun exploring the new release, we hope you like it! If you haven't already, join the [ORY Community Slack](http://slack.ory.sh) where we hold weekly community hangouts via video chat and answer your questions, exchange ideas, and present new developments!



## Breaking Changes

The "common" keyword has been removed from the Swagger 2.0 spec which deprecates the `common` module / package / class (depending on the generated SDK). Please use `public` or `admin` instead!

Additionally, the SDK for TypeScript now uses the `fetch` API which allows the SDK to be used in both client-side as well as server-side contexts. Please note that several methods and parameters in the generated TypeScript SDK have changed. Please check the TypeScript results to see what needs to be changed!

This patch changes the OpenID Connect and OAuth2 ("Sign in with Google, Facebook, ...") Callback URL from `http(s)://<kratos-public>/self-service/browser/flows/strategies/oidc/<provider>` to `http(s)://<kratos-public>/self-service/methods/oidc/<provider>`. To apply this patch, you need to update these URLs at the OAuth2 Client configuration pages of the individual OpenID Conenct providers (e.g. GitHub, Google).

Configuration key `selfservice.strategies` was renamed to `selfservice.methods`.

This patch significantly changes how email verification works. The Verification Flow no longer uses its own system but now re-uses the API and Browser flows and flow methods established in other components such as login, recovery, registration.

Due to the many changes these patch notes does not cover how to upgrade this particular flow. We instead want to kindly ask you to check out the updated documentation for this flow at: https://www.ory.sh/kratos/docs/self-service/flows/verify-email-account-activation

This patch changes the SQL schema and thus requires running the SQL Migration command (e.g. `... migrate sql`).
Never apply SQL migrations without backing up your database prior.

Configuration items `selfservice.flows.<name>.request_lifespan` have been renamed to `selfservice.flows.<name>.lifespan` to match the new flow semantics.

Wording has changed from "Self-Service Recovery Request" to "Self-Service Recovery Flow" to follow community feedback and practice already applied in the documentation. Additionally, fetching a recovery flow over the public API no longer requires Anti-CSRF cookies to be sent.

This patch renames several important recovery flow endpoints:

- `/self-service/browser/flows/recovery` is now `/self-service/recovery/browser` without functional changes.
- `/self-service/browser/flows/requests/recovery?request=abcd` is now `/self-service/recovery/flows?id=abcd` and no longer needs anti-CSRF cookies to be available.

Additionally, the URL for completing the password and oidc recovery method has been moved. Given that this endpoint is typically not manually called, you can probably ignore this change:

- `/self-service/browser/flows/recovery/link?request=abcd` is now `/self-service/recovery/methods/link?flow=abcd` without functional changes.

The Recovery UI Endpoint no longer receives a `?request=abcde` query parameter but instead a `?flow=abcde` query parameter. Functionality did not change however.

As part of this change SDK methods have been renamed:

```
  const kratos = new CommonApi(config.kratos.public)
  // ...
- kratos.completeSelfServiceBrowserRecoveryLinkStrategyFlow(req.query.request)
+ kratos.completeSelfServiceRecoveryFlowWithLinkMethod(req.query.flow)
```

This patch requires you to run SQL migrations.

Wording has changed from "Self-Service Settings Request" to "Self-Service Settings Flow" to follow community feedback and practice already applied in the documentation.

This patch renames several important settings flow endpoints:

- `/self-service/browser/flows/settings` is now `/self-service/settings/browser` without functional changes.
- `/self-service/browser/flows/requests/settings?request=abcd` is now `/self-service/settings/flows?id=abcd` and no longer needs anti-CSRF cookies to be available.

Additionally, the URL for completing the password, profile, and oidc settings method has been moved. Given that this endpoint is typically not manually called, you can probably ignore this change:

- `/self-service/browser/flows/login/strategies/password?request=abcd` is now `/self-service/login/methods/password?flow=abcd` without functional changes.
- `/self-service/browser/flows/strategies/oidc?request=abcd` is now `/self-service/methods/oidc?flow=abcd` without functional changes.
- `/self-service/browser/flows/settings/strategies/profile?request=abcd` is now `/self-service/settings/methods/profile?flow=abcd` without functional changes.

The Settings UI Endpoint no longer receives a `?request=abcde` query parameter but instead a `?flow=abcde` query parameter. Functionality did not change however.

As part of this change SDK methods have been renamed:

```
  const kratos = new CommonApi(config.kratos.public)
  // ...
- kratos.getSelfServiceBrowserSettingsRequest(req.query.request)
+ kratos.getSelfServiceSettingsFlow(req.query.flow)

  // You will most likely not be using this:
  const kratos = new PublicApi(config.kratos.public)
- kratos.completeSelfServiceBrowserSettingsPasswordStrategyFlow //...
- kratos.completeSelfServiceSettingsFlowWithPasswordMethod //..
- kratos.completeSelfServiceBrowserSettingsProfileStrategyFlow //...
- kratos.completeSelfServiceSettingsFlowWithProfileMethod //..
```

This patch requires you to run SQL migrations.

This patch makes the reverse proxy functionality required in prior versions of the self-service UI example obsolete. All examples work now with a simple set up and documentation has been added to assist in subdomain scenarios.

The session field `sid` has been renamed to `id` to stay consistent with other APIs which also use `id` terminology to clarify identifiers. The payload of, for example, `/session/whoami` has changed as follows:

```patch
  {
-   "sid": "abcde",
+   "id": "abcde",
    "expires_at": "..."
    "identity": {
      // ..
    }
  }
```

Wording has changed from "Self-Service Registration Request" to "Self-Service Registration Flow" to follow community feedback and practice already applied in the documentation. Additionally, fetching a login flow over the public API no longer requires Anti-CSRF cookies to be sent.

This patch renames several important registration flow endpoints:

- `/self-service/browser/flows/registration` is now `/self-service/registration/browser` without behavioral change.
- `/self-service/browser/flows/requests/registration?request=abcd` is now `/self-service/registration/flows?id=abcd` and no longer needs anti-CSRF cookies to be available.

Additionally, the URL for completing the password registration method has been moved. Given that this endpoint is typically not manually called, you can probably ignore this change:

- `/self-service/browser/flows/registration/strategies/password?request=abcd` is now `/self-service/registration/methods/password?flow=abcd` without functional changes.
- `/self-service/browser/flows/strategies/oidc?request=abcd` is now `/self-service/methods/oidc?flow=abcd` without functional changes.

The Registration UI Endpoint no longer receives a `?request=abcde` query parameter but instead a `?flow=abcde` query parameter. Functionality did not change however.

As part of this change SDK methods have been renamed:

```
  const kratos = new CommonApi(config.kratos.public)
  // ...
- kratos.getSelfServiceBrowserRegistrationRequest(req.query.request)
+ kratos.getSelfServiceRegistrationFlow(req.query.flow)
```

This patch requires you to run SQL migrations.

Existing login sessions will no longer be valid because the session cookie data model changed. If you apply this patch, your users will need to sign in again.

Wording has changed from "Self-Service Login Request" to "Self-Service Login Flow" to follow community feedback and practice already applied in the documentation. Additionally, fetching a login flow over the public API no longer requires Anti-CSRF cookies to be sent.

This patch renames several important login flow endpoints:

- `/self-service/browser/flows/login` is now `/self-service/login/browser` without functional changes.
- `/self-service/browser/flows/requests/login?request=abcd` is now `/self-service/login/flows?id=abcd` and no longer needs anti-CSRF cookies to be available.

Additionally, the URL for completing the password and oidc login method has been moved. Given that this endpoint is typically not manually called, you can probably ignore this change:

- `/self-service/browser/flows/login/strategies/password?request=abcd` is now `/self-service/login/methods/password?flow=abcd` without functional changes.
- `/self-service/browser/flows/strategies/oidc?request=abcd` is now `/self-service/methods/oidc?flow=abcd` without functional changes.

The Login UI Endpoint no longer receives a `?request=abcde` query parameter but instead a `?flow=abcde` query parameter. Functionality did not change however.

As part of this change SDK methods have been renamed:

```
  const kratos = new CommonApi(config.kratos.public)
  // ...
- kratos.getSelfServiceBrowserLoginRequest(req.query.request)
+ kratos.getSelfServiceLoginFlow(req.query.flow)
```

This patch requires you to run SQL migrations.

Configuraiton value `session.cookie_same_site` has moved to `session.cookie.same_site`. There was no functional change.



### Bug Fixes

* Add missing 'recovery' path in oathkeeper access-rules.yml ([#763](https://github.com/ory/kratos/issues/763)) ([f180dba](https://github.com/ory/kratos/commit/f180dba2207638e83e4a23ebc213cddaecb5677f))
* Add missing error handling ([43c1446](https://github.com/ory/kratos/commit/43c14464efa7b736695e2144b031daf6fca87703))
* Add ory-prettier-styles to main repo ([#744](https://github.com/ory/kratos/issues/744)) ([aeaddbc](https://github.com/ory/kratos/commit/aeaddbcb27f89d61b076bdd9ad1739fb1da2ffd9))
* Add remote help description ([f66bbe1](https://github.com/ory/kratos/commit/f66bbe18cfad1e8725ecbcf6e2843b34c3d5119f))
* Add serve help description ([2eb072b](https://github.com/ory/kratos/commit/2eb072b71e5602895d4232e197bfd76180fcdcd7))
* Allow using json with form layout in password registration ([bd2225c](https://github.com/ory/kratos/commit/bd2225c0fff3e0363716d2096346d59046838bb7))
* Annotate whoami endpoint with cookie and token ([a8a781c](https://github.com/ory/kratos/commit/a8a781c00847c74c65558b55e882e12c1e69d8c8))
* Bump datadog version to fix build failure ([4dfd322](https://github.com/ory/kratos/commit/4dfd322290313ec8467ebe8b385b56004b2417bd))
* Change KRATOS_ADMIN_ENDPOINT to KRATOS_ADMIN_URL ([763fdc5](https://github.com/ory/kratos/commit/763fdc56d19d12fa2b83eed2757fbf178d9288b1))
* Clarify fetch use ([8eb2e6f](https://github.com/ory/kratos/commit/8eb2e6f222788a9a579774772696c77987f3cf97))
* Complete verification by redirecting to UI with success ([f0ecf51](https://github.com/ory/kratos/commit/f0ecf5144970f666643aa7c00a3f4ca73f4ab047))
* Correct cookie domain on logout ([#646](https://github.com/ory/kratos/issues/646)) ([6d77e04](https://github.com/ory/kratos/commit/6d77e043ce3bec0864b8abdee371a101f68e4335)), closes [#645](https://github.com/ory/kratos/issues/645)
* Correct help message for import ([a5f46d2](https://github.com/ory/kratos/commit/a5f46d260b43d15f8e77b04cb36c589e103468bf))
* Correct password and profile swagger annotations ([668c184](https://github.com/ory/kratos/commit/668c1847c4c4236ca28f9dcd5147b523a2f60832))
* Correct password registration method api spec ([08dd582](https://github.com/ory/kratos/commit/08dd582195cdb6a891d2428ba5d02cd956555e48))
* Correct PHONY spelling ([#739](https://github.com/ory/kratos/issues/739)) ([e3d3617](https://github.com/ory/kratos/commit/e3d3617b8d82812b0ad67cc1cb02ff86c2c0c66c))
* Cover more test cases for persister ([37d2e08](https://github.com/ory/kratos/commit/37d2e0839b88792733387f26abb98c51bd1e1395))
* Create decoder only once ([34dc43b](https://github.com/ory/kratos/commit/34dc43b0c75303f88d2c304225c027faf5366c1f))
* Deprecate packr2 dependency in makefile ([be9a84d](https://github.com/ory/kratos/commit/be9a84dcffbccd5f0e073a38264cf11a404d3b66)), closes [#711](https://github.com/ory/kratos/issues/711) [#750](https://github.com/ory/kratos/issues/750)
* Do not propagate parent validation error ([bf6093d](https://github.com/ory/kratos/commit/bf6093d442d9779b4df051031565d020ef628ded))
* Don't resend verification emails once verified ([#583](https://github.com/ory/kratos/issues/583)) ([a4d9969](https://github.com/ory/kratos/commit/a4d99694525e65b58d49197c96324b27fb8c31c2)), closes [#578](https://github.com/ory/kratos/issues/578)
* Enforce endpoint to be set ([171ac18](https://github.com/ory/kratos/commit/171ac18d73eaa0822b45f544a9034d6734400f31))
* Escape jsx characters in api documentation ([0946094](https://github.com/ory/kratos/commit/09460948a24918b2a84804cafa86cf88189af919))
* Exit with code 1 on unimplemented CLI commands ([66943d7](https://github.com/ory/kratos/commit/66943d7e5b47fc477a378d8a7cf2b2009ccfceb3))
* Explicitly ignore fprint return values ([f50e582](https://github.com/ory/kratos/commit/f50e5823f4ee047fdc3e276b80b4fb08c9128d99))
* Explicitly ignore fprintf results ([a83dc50](https://github.com/ory/kratos/commit/a83dc509970b3be46d832743481357f336fecc35))
* Fallback to default return url if logout after url is not defined ([#594](https://github.com/ory/kratos/issues/594)) ([7edd367](https://github.com/ory/kratos/commit/7edd367dc64a01dbe252ca0ab8cf4d3926a35014))
* Favor packr2 over pkger ([ac18a45](https://github.com/ory/kratos/commit/ac18a45ea55929c34ca20953e3baa197363483bc)):

    See https://github.com/markbates/pkger/issues/117

* Find and replace "request" references ([41fb673](https://github.com/ory/kratos/commit/41fb673e38779cb27d4400f70458617eb7e5b93c))
* Force exe buildmode for windows CGO ([e017bb5](https://github.com/ory/kratos/commit/e017bb579cd29ad1a634cd552e2601295ff9c104))
* Html form parse regression issue ([6b07cbb](https://github.com/ory/kratos/commit/6b07cbb657702d36423d1fa66fe8a149222c8772))
* Ignore x/net false positives ([7044b95](https://github.com/ory/kratos/commit/7044b95f6188c4ffbfff42c666dee6ebaba055c8))
* Improve debugging output for login hook and restructure files ([dabac40](https://github.com/ory/kratos/commit/dabac40f82407f72071780840f468d0b5b389777))
* Improve debugging output for registration hook and restructure files ([ec11775](https://github.com/ory/kratos/commit/ec117754f5dd41e5a3a43b3807c05796396ced55))
* Improve expired error responses ([124a92e](https://github.com/ory/kratos/commit/124a92ee98d62abeb695e1e271ee2536a69d6047))
* Improve hook tests ([55ba485](https://github.com/ory/kratos/commit/55ba48530a890fdd55ed7da380940f2791148f26))
* Improve makefile dependency building ([8e1d69a](https://github.com/ory/kratos/commit/8e1d69a024414196b39eb3d419f4850cd547e3b5))
* Improve pagination when listing identities ([c60bf44](https://github.com/ory/kratos/commit/c60bf440b9c85b4f2e871237e3d7725571151efe))
* Improve post login hook log and audit messages ([ddd5d5a](https://github.com/ory/kratos/commit/ddd5d5a253d01d2b7b74239a1c7c701759084140))
* Improve post registration hook log and audit messages ([2495629](https://github.com/ory/kratos/commit/24956296dd91cf6f5b110a17f65f9f60d8a7aa78))
* Improve registration hook tests ([8163152](https://github.com/ory/kratos/commit/8163152a4d9595b1ea73d2887205e7ba80b016f9))
* Improve session max-age behavior ([65189fe](https://github.com/ory/kratos/commit/65189fe4a2f84f832240cd67366400e44bb7f09a)), closes [#42](https://github.com/ory/kratos/issues/42)
* Keep HTML form type on registration error ([#698](https://github.com/ory/kratos/issues/698)) ([6c9e756](https://github.com/ory/kratos/commit/6c9e7564efffe1452004d4eda42e1b9ec9feac6b)), closes [#670](https://github.com/ory/kratos/issues/670)
* Lowercase emails on login ([244b4dd](https://github.com/ory/kratos/commit/244b4dd825b9a2448cc61465cef81bd9dcb051db))
* Mark flow methods' fields as required ([#708](https://github.com/ory/kratos/issues/708)) ([834c607](https://github.com/ory/kratos/commit/834c60738ca7bb26e982ff73134b7b0e85a72076))
* Merge public and admin login flow fetch handlers ([48c4906](https://github.com/ory/kratos/commit/48c4906a606396d889e057a03dc83b619220db54))
* Missing write in registration error handler ([3b2af53](https://github.com/ory/kratos/commit/3b2af5397048d63099eace092bf2e50e84a4c610))
* Properly annotate swagger password parameters ([2ef57c4](https://github.com/ory/kratos/commit/2ef57c4323eb2623f4115bee0e44ee27dd1648a9))
* Properly fetch identity for session ([7be4086](https://github.com/ory/kratos/commit/7be4086045fddfacc38813ca3dd7fbcc7039391f))
* Recursive loop on network errors in password validator ([#589](https://github.com/ory/kratos/issues/589)) ([b4d5a42](https://github.com/ory/kratos/commit/b4d5a42346510e40222b8eb59b455b585f0a05cf)), closes [#316](https://github.com/ory/kratos/issues/316):

    The old code no error when ignoreNetworkErrors was set to true, but did not set a hash result which caused an infinite loop.

* Remove incorrect security specs ([4c3d46d](https://github.com/ory/kratos/commit/4c3d46dac20363202f0ccd043e1c9d6bf97fb1f8))
* Remove obsolete tests ([f102f95](https://github.com/ory/kratos/commit/f102f95f420c8a03520602880d096616069c9233)):

    The test is no longer valid as CSRF checks now happen after checking for login sessions in settings flows.

* Remove redirector from code base ([6689ecf](https://github.com/ory/kratos/commit/6689ecf110b11ba15ec39af822906c2b4b17369e))
* Remove stray debug statements ([a8e1ec4](https://github.com/ory/kratos/commit/a8e1ec42cda6ebc664e9434bb5ba7e4dd7c21b4c))
* Rename import to put ([8003e0f](https://github.com/ory/kratos/commit/8003e0f42a5d1b77e326d1dba0a70fcd44c704c0))
* Rename quickstart config files and path ([#671](https://github.com/ory/kratos/issues/671)) ([be8b9e5](https://github.com/ory/kratos/commit/be8b9e5f1ca70b1aa06b77bb2ca35644d8cd3c00))
* Rename quickstart schema file name ([e943c90](https://github.com/ory/kratos/commit/e943c9018a495b39b72ae463fd4727b1798d5ba2))
* Rename recovery models and generate SDKs ([d764435](https://github.com/ory/kratos/commit/d7644359c39732e0b25f43e122d05c1566fb837b))
* Resolve and test for missing data when updating flows ([045ecab](https://github.com/ory/kratos/commit/045ecab11ec185ca688a10de75e506fe413afa26))
* Resolve broken csrf tests ([6befe2e](https://github.com/ory/kratos/commit/6befe2ec08c01c6c9fb397ba119ecebdcecf7db3))
* Resolve broken docs links ([56f4a39](https://github.com/ory/kratos/commit/56f4a397a715b6c0428ae63baa0d2e4bc936f737))
* Resolve broken migrations and bump fizz ([1ed9c70](https://github.com/ory/kratos/commit/1ed9c700b946a090bce9587a57eeb9ac64f04c59))
* Resolve broken OIDC tests and disallow API flows ([9986d8f](https://github.com/ory/kratos/commit/9986d8f818934bd5e073f59bf7a73c6b7a74b6e2))
* Resolve cookie issues ([6e2b6d2](https://github.com/ory/kratos/commit/6e2b6d2f0ce2fb6df7d3e26d6cc8e755e6593a81))
* Resolve e2e headless test failures ([82d506e](https://github.com/ory/kratos/commit/82d506e9d35bbbe4c1578f72e5bcf380ebc97142))
* Resolve e2e test failures ([2627db2](https://github.com/ory/kratos/commit/2627db26089e8f8e4c18782ff59b4cb2068b276f))
* Resolve failing test cases ([f8647b4](https://github.com/ory/kratos/commit/f8647b4c637b4aee29d68df2336fd216306ec78c))
* Resolve flaky passwort setting tests ([#582](https://github.com/ory/kratos/issues/582)) ([c42d936](https://github.com/ory/kratos/commit/c42d936ef51d2ffb48b491b99988d048442e3b8b)), closes [#581](https://github.com/ory/kratos/issues/581) [#577](https://github.com/ory/kratos/issues/577)
* Resolve handler testing issue ([4f6bafd](https://github.com/ory/kratos/commit/4f6bafdc84ba4d878c68700dc243cd3cfe8fe530))
* Resolve identity admin api issues ([#586](https://github.com/ory/kratos/issues/586)) ([feef8a7](https://github.com/ory/kratos/commit/feef8a7d4454c1b343c34a96fa4dadd56149b0cd)), closes [#435](https://github.com/ory/kratos/issues/435) [#500](https://github.com/ory/kratos/issues/500):

    This patch resolves several issues that occurred when creating or updating identities using the Admin API. Now, all hooks are running properly and updating privileged properties no longer causes errors.

* Resolve interface type issues ([064b305](https://github.com/ory/kratos/commit/064b305ab31dc003ccb5992eb1ed2804f85085b9))
* Resolve logout csrf issues ([#761](https://github.com/ory/kratos/issues/761)) ([74c0aac](https://github.com/ory/kratos/commit/74c0aac3b94446c3824ae52b04b6f69395938b81))
* Resolve migratest failures ([e2f34d3](https://github.com/ory/kratos/commit/e2f34d3f411bac042079d7f5425063ef117fae77))
* Resolve migratest ordering failing tests ([dffecc0](https://github.com/ory/kratos/commit/dffecc0e80810ffae57870fd313ee0103ad3f60c))
* Resolve migration issues ([b545e15](https://github.com/ory/kratos/commit/b545e15eeaa3e6e1f4a8fe0f8e1890012ac62c94))
* Resolve panic on `serve` ([ae34155](https://github.com/ory/kratos/commit/ae341555e7b2b622cf58d09d3eb6a78d833dfdcc))
* Resolve panic when DSN="memory" ([#574](https://github.com/ory/kratos/issues/574)) ([05e55f3](https://github.com/ory/kratos/commit/05e55f3584e20ae5d39cfda6e542d4da40d718e4)):

    Executing the migration logic in registry.go cause a panic as the registry is not initalized at that point. Therefore we decided to move the handling to driver_default.go, after the registry has been initialized.

* Resolve pkger issues ([294066c](https://github.com/ory/kratos/commit/294066c41be1d508681caa435afda4858a37b7f1))
* Resolve remaining testing issues ([af40d93](https://github.com/ory/kratos/commit/af40d933b2f663adb6a537b32546b43ba13ae237))
* Resolve SQL persistence tester issues ([4952df4](https://github.com/ory/kratos/commit/4952df43e0aba067c06cdedb1fc2c2d9a2a81a40))
* Resolve swagger issues and regenerate SDK ([be4c7e4](https://github.com/ory/kratos/commit/be4c7e4ea72d2ad7cec67b1d6709858d5a1b3d61))
* Resolve template loading issue ([145fb20](https://github.com/ory/kratos/commit/145fb204d9a8ca189480f9f2221527ccc62980a0))
* Resolve test issues introduced by new csrf protection ([625ef5e](https://github.com/ory/kratos/commit/625ef5e4781700449af0c4e4f1f6cb8aa1787764))
* Resolve verification sql errors ([784da53](https://github.com/ory/kratos/commit/784da53ddefe59aea90254be40ae63e919b4b419))
* Resolves a bug that prevents sessions from expiring ([#612](https://github.com/ory/kratos/issues/612)) ([86b281a](https://github.com/ory/kratos/commit/86b281a46b676d80c8f70bfc42c91d988997c21c)), closes [#611](https://github.com/ory/kratos/issues/611)
* Revert disabling `swagger flatten` during sdk generation ([98c7915](https://github.com/ory/kratos/commit/98c7915cc493ad99c959244eef68b70bc9baa971))
* Set correct path for kratos in oathkeeper set up ([414259f](https://github.com/ory/kratos/commit/414259f9383f30b762051c712763d484f5358075))
* Set quickstart logging to trace ([d3e9192](https://github.com/ory/kratos/commit/d3e919249ae59b449367511d3cc8adef839f31c9))
* Support browser flows only in redirector ([cab5280](https://github.com/ory/kratos/commit/cab5280859b0fc7fc7fec2b2ec9945f457910b20))
* Swagger models ([1b5f9ab](https://github.com/ory/kratos/commit/1b5f9abd5d82251ab93a05d4ff26b4c48c8151ca)):

    The `swagger:parameters <id>` definitions for `updateIdentity` and `createIdentity` where defined two times with the same ID. They had some old definition swagger used. The `internal/httpclient` should now work again as expected.

* Tell tls what the smtps server name is ([#634](https://github.com/ory/kratos/issues/634)) ([b724038](https://github.com/ory/kratos/commit/b724038a67e84ca71b146bf4b9b044be2dc8c0b4))
* Type ([e264c69](https://github.com/ory/kratos/commit/e264c69a07e569429b5e835b1e15c318eff23339))
* Update cli documentation examples ([216ea7f](https://github.com/ory/kratos/commit/216ea7f926798ff03d211447200919f9ef3c8b39))
* Update contrib samples ([79d24b4](https://github.com/ory/kratos/commit/79d24b4472017a75854cce4a45b4c762e5390a67))
* Update crdb quickstart version ([249a6ba](https://github.com/ory/kratos/commit/249a6bae32ccaa6cf002eaab921388e8cb10e58f))
* Update import description ([aef1e1a](https://github.com/ory/kratos/commit/aef1e1acf757637590fe19644952a44d1994ba18))
* Update quickstart kratos config ([e3246e5](https://github.com/ory/kratos/commit/e3246e5d56b95750529239663bab03168789cc09))
* Update recovery token field and column names ([42abfa1](https://github.com/ory/kratos/commit/42abfa1dea2a6291c5b723baf25f35a66f2af835))
* Update status help description ([b147831](https://github.com/ory/kratos/commit/b1478316d2f601843133fd33d75c3b047384f283))
* Update swagger names and fix broken tests ([85b7fb1](https://github.com/ory/kratos/commit/85b7fb1d466bc4dcee97ad75cc92b8bea8e44d9f))
* Update version help description ([8bf4a79](https://github.com/ory/kratos/commit/8bf4a79064a93cb53ef8aee3433b24602bc9f30a))
* Use and test for csrf tokens and prevent api misuse ([a4e3bc5](https://github.com/ory/kratos/commit/a4e3bc55e43ba42582a33551c1cc2e83ecd865fa))
* Use correct HTTP method for password login ([4f4fcee](https://github.com/ory/kratos/commit/4f4fcee8931ab4998e974106b8d88e0c61736e3f))
* Use correct log message ([53c384a](https://github.com/ory/kratos/commit/53c384a542a583259a75315b2602cf4fb41a0ef0))
* Use correct redirection for registration ([8d47113](https://github.com/ory/kratos/commit/8d47113a5f7c0c25dc5f92c683b560763cfd47c9))
* Use correct security annotation ([c9bebe0](https://github.com/ory/kratos/commit/c9bebe00452a73d1c831831e5a95cb4ed8de37b9))
* Use correct swagger tags and regenerate ([df99d8c](https://github.com/ory/kratos/commit/df99d8cbe6e0f2f6a5da872f66db557b2a5e9f70))
* Use helpers to create flow ([aba8610](https://github.com/ory/kratos/commit/aba861097d2c67ce9ebff85df59fce8018862516))
* Use nosurf fork to address VerifyToken bug ([cd84e51](https://github.com/ory/kratos/commit/cd84e51b7b1861ca9bd2312a4dfc5e84afd890cf))
* Use params per_page and page for pagination ([5dfb6e3](https://github.com/ory/kratos/commit/5dfb6e32c44420ed49d652733b9099a41c9347f2))
* Use proper pwd in makefile ([52e22c3](https://github.com/ory/kratos/commit/52e22c3b5c0130afd3e235aba9847389369f435e))
* Use public instead of common sdk ([dcb4a36](https://github.com/ory/kratos/commit/dcb4a36f9fb3c25ace9a252b7e05f7ab71d2e21f))
* Use relative threshold to judge longest common substring in password policy ([#585](https://github.com/ory/kratos/issues/585)) ([3e9f8cc](https://github.com/ory/kratos/commit/3e9f8cce4b058b05d69c73fff514f3b8e46c2be3)), closes [#581](https://github.com/ory/kratos/issues/581)
* Whoami returns 401 not 403 ([3b3b78c](https://github.com/ory/kratos/commit/3b3b78c04bbbbb7b7fb05635d96b4f7c7fa7776f)), closes [#729](https://github.com/ory/kratos/issues/729)

### Code Generation

* Pin v0.5.0-alpha.1 release commit ([557d37d](https://github.com/ory/kratos/commit/557d37d1139adb14a25abe40d0174d47d4e18fee))

### Code Refactoring

* Add flow methods to verification ([00ee828](https://github.com/ory/kratos/commit/00ee828842bd4bc6f917ba2446b1374d28b62000)):

    Completely refactors the verification flow to support other methods. The original email verification flow now moved to the "link" method also used for recovery.
    
    Additionally, several upstream bugs in gobuffalo/pop and gobuffalo/fizz have been addressed, patched, and merged which improves support for SQLite and CockroachDB migrations:
    
    - https://github.com/gobuffalo/fizz/pull/97
    - https://github.com/gobuffalo/fizz/pull/96

* Add method and rename request to flow ([006bf56](https://github.com/ory/kratos/commit/006bf56671d8162cdb5bcce630c027b67935263d))
* Change oidc callback URL ([36d9380](https://github.com/ory/kratos/commit/36d9380b2123d27219c908b51ad97574ee11bc57))
* Complete login flow refactoring ([ad2b3db](https://github.com/ory/kratos/commit/ad2b3db4493085b80889cbc0dce9562288ec6896))
* Dry up login.NewFlow ([f261c44](https://github.com/ory/kratos/commit/f261c442dbe74e3b9887193b74e36fe70306f9d8))
* Improve CSRF infrastructure ([7e367e7](https://github.com/ory/kratos/commit/7e367e7f45481147d5c231d0ea8cbb30b738226f))
* Improve login test reuse ([b4184e5](https://github.com/ory/kratos/commit/b4184e5f1525a9918bc795f2353b186141ce5399))
* Improve NewFlowExpiredError ([1caefac](https://github.com/ory/kratos/commit/1caefac6e0e82aa2b12458ef16d7f5af24014bf9))
* Improve registration tests with testhelpers ([9bf4530](https://github.com/ory/kratos/commit/9bf45303be908449b78c68c7382eab5cfc5c40fa))
* Improve selfservice method tests ([df4d06d](https://github.com/ory/kratos/commit/df4d06d553852cdb8b914810c19bdd0fcc845c9c))
* Improve settings helper functions ([fda17ca](https://github.com/ory/kratos/commit/fda17ca5ea7824c4bf5010218cace7d5fbc7ad5b))
* Move samesite config to cookie parent-key ([753eb86](https://github.com/ory/kratos/commit/753eb86c904c4af9e7d91e46ff4c836dcce35807))
* Moved clihelpers to ory/x ([#756](https://github.com/ory/kratos/issues/756)) ([6ccffa8](https://github.com/ory/kratos/commit/6ccffa8a1cc5b9fd33435187720257bb66323546)):

    Contributes to https://github.com/ory/hydra/issues/2124.
    
    

* Profile settings method is now API-able ([c5f361f](https://github.com/ory/kratos/commit/c5f361ff418336cfcaa452eded4bd61132808b16))
* Remove common keyword from API spec ([6619562](https://github.com/ory/kratos/commit/6619562667ef0e363d14c57cfbcd15c16f292853))
* Remove need for reverse proxy in selfservice-ui ([beb4c32](https://github.com/ory/kratos/commit/beb4c3284e552fe51c3a8cebb20a8c2bfc07cdf8)), closes [#661](https://github.com/ory/kratos/issues/661)
* Rename `session.sid` to `session.id` ([809fe73](https://github.com/ory/kratos/commit/809fe7334e4a308405c1f03ada1dbef6ed33c01a))
* Rename login request to login flow ([9369d1b](https://github.com/ory/kratos/commit/9369d1bb637fc80b5d5980140693d5bcac0c76bb)), closes [#635](https://github.com/ory/kratos/issues/635):

    As part of this change, fetching a login flow over the public API no longer requires Anti-CSRF cookies to be sent.

* Rename LoginRequestErrorHandler to LoginFlowErrorHandler ([66ae029](https://github.com/ory/kratos/commit/66ae029f49aecdfba5fa6905cfccfcdad992dd5a))
* Rename package recoverytoken to link ([f87fb54](https://github.com/ory/kratos/commit/f87fb549f6d8a10ba5adffddeb2fe12060d520ab))
* Rename recovery request to flow internally ([16c5618](https://github.com/ory/kratos/commit/16c5618644e78cf1081f966e01b570a36eea709b))
* Rename recovery request to recovery flow ([b0f433d](https://github.com/ory/kratos/commit/b0f433d4cb65d79acba789394d828663e873a833)), closes [#635](https://github.com/ory/kratos/issues/635):

    As part of this change, fetching a login flow over the public API no longer requires Anti-CSRF cookies to be sent.

* Rename registration request to flow ([8437ebc](https://github.com/ory/kratos/commit/8437ebcf4deb2844562ec701af3bbbb2a9b5dea4))
* Rename registration request to registration flow ([0470956](https://github.com/ory/kratos/commit/0470956128d03921d8554c43af2c5a0003abe82f)), closes [#635](https://github.com/ory/kratos/issues/635):

    As part of this change, fetching a registration flow over the public API no longer requires Anti-CSRF cookies to be sent.

* Rename request_lifespan to lifespan ([#677](https://github.com/ory/kratos/issues/677)) ([3c8d5e0](https://github.com/ory/kratos/commit/3c8d5e02b04686a1e0bfbd28caa0bc536e3414e4)), closes [#666](https://github.com/ory/kratos/issues/666)
* Rename strategies to methods ([8985189](https://github.com/ory/kratos/commit/89851896d563518909bc2b47a7ff91683eec4958)):

    This patch renames `strategies` such as "Username/Email & Password" to methods.

* Rename verify to verificaiton ([#597](https://github.com/ory/kratos/issues/597)) ([0ecd69a](https://github.com/ory/kratos/commit/0ecd69a60f741fc334c9b060b6aeaafc39e048b1))
* Replace all occurrences of login request to flow ([1b3c491](https://github.com/ory/kratos/commit/1b3c49174a7a2eff51dd531f3a49afc15c31c536))
* Replace all registration request occurrences with registration flow ([308ef47](https://github.com/ory/kratos/commit/308ef47846c9ab4f18a598ef6ef78514fad77c42))
* Replace packr2 with pkger fork ([4e2acae](https://github.com/ory/kratos/commit/4e2acae7c4fc17880cf88ef05cf7cca5f20f5be3))
* Restructure login package ([c99e2a2](https://github.com/ory/kratos/commit/c99e2a2f23c3c2aabaae55de67e40ab7fb2dd307))
* Use session token as cookie identifier ([60fd9c2](https://github.com/ory/kratos/commit/60fd9c2efa881fcdd769a8967abe73c05a198868))

### Documentation

* Add administrative user management guide ([b97e0c6](https://github.com/ory/kratos/commit/b97e0c69bb1115bdec88b218e8cdda34f137d798))
* Add code samples to session checking ([eba8eda](https://github.com/ory/kratos/commit/eba8eda70423aa802eace278889a5e8d2e0bc513))
* Add configuring introduction ([#630](https://github.com/ory/kratos/issues/630)) ([b8cfb35](https://github.com/ory/kratos/commit/b8cfb351c2dca783e355f39d25ce17b65fef7dd4))
* Add descriptions to cobra commands ([607b76d](https://github.com/ory/kratos/commit/607b76d109d1fa519235fe9d6af78c8315b9c4fc))
* Add documentation for configuring cookies ([e3dbc8a](https://github.com/ory/kratos/commit/e3dbc8acc055f6e2d78bc959be7356f9a66ac90f)), closes [#516](https://github.com/ory/kratos/issues/516)
* Add domain, subdomain, multi-domain cookie guides ([3eb1e59](https://github.com/ory/kratos/commit/3eb1e5987df56993c792684a6a2bc11f5eb570b8)), closes [#661](https://github.com/ory/kratos/issues/661)
* Add github video tutorial ([#622](https://github.com/ory/kratos/issues/622)) ([0c4222c](https://github.com/ory/kratos/commit/0c4222c0d12df4e971fd7e5099006484e0bcb317))
* Add guide for cors ([a8ae759](https://github.com/ory/kratos/commit/a8ae759565d94ebd9d0f758b7eb6efbddf486372))
* Add guide for cors ([91fd278](https://github.com/ory/kratos/commit/91fd278d1a6720576998b115dedb882b90915561))
* Add guide for dealing with login sessions ([4e2718c](https://github.com/ory/kratos/commit/4e2718c779031c0e3b877e9df1747ccb2371927b))
* Add identity state ([fb4aedb](https://github.com/ory/kratos/commit/fb4aedb9a95367e25080491b54aab11de491d819))
* Add login session to navbar ([b212d64](https://github.com/ory/kratos/commit/b212d6484e40c9f2cce10f2ba4aaf4e2a72f03a1))
* Add milestones to sidebar ([aae13ec](https://github.com/ory/kratos/commit/aae13ec141a2c315aff1a53aa005bb9465efcdc0))
* Add missing GitLab provider to the list of supported OIDC providers ([#766](https://github.com/ory/kratos/issues/766)) ([a43ed33](https://github.com/ory/kratos/commit/a43ed335262fd542f349224aef918af5263c384d))
* Add missing TOC entries ([#748](https://github.com/ory/kratos/issues/748)) ([bd7edfb](https://github.com/ory/kratos/commit/bd7edfbebd19f01af337c34293ebc2865f2b077d))
* Add pagination docs ([7fe0901](https://github.com/ory/kratos/commit/7fe0901ee5d0e829e110bd0c4fdecb24bfc27768))
* Add secret key rotation guide ([3d6e21a](https://github.com/ory/kratos/commit/3d6e21af2f726944468299c326600a8ab0e4e885))
* Add sequence diagrams for browser/api flows ([590d767](https://github.com/ory/kratos/commit/590d767352b9253b7550eaba56fea99400399cd7))
* Add session hook to ssi guide ([#623](https://github.com/ory/kratos/issues/623)) ([1bbed39](https://github.com/ory/kratos/commit/1bbed390ffedd811afdb5fcfe69047554419d8ce))
* Add terminology section ([29b81a7](https://github.com/ory/kratos/commit/29b81a78fcf880cd6d9d3b2cbb03f955b701ffbd))
* Add theme helpers and decouple mermaid ([7c3eb32](https://github.com/ory/kratos/commit/7c3eb32df5d9287845258bf25d6719733f6c4227))
* Add video to OIDC guide ([#619](https://github.com/ory/kratos/issues/619)) ([f286980](https://github.com/ory/kratos/commit/f286980c29ce8460ba550e5d74b8dee23602e920))
* Added sidebar cli label ([5d24a29](https://github.com/ory/kratos/commit/5d24a2998b412159295feca40421b8b11cf02274)):

    `clidoc.Generate` expects to find an entry under `sidebar.json/Reference` that contains the substring "CLI" in it's label. Because that was missing, a new entry was appended on every regeneration of the file.

* Added sidebar item ([#639](https://github.com/ory/kratos/issues/639)) ([8574761](https://github.com/ory/kratos/commit/857476112d12b8ab79ef49054452a950ff81bc23)):

    Added Kratos Video Tutorial Transcripts document to sidebar.

* Added transcript ([#627](https://github.com/ory/kratos/issues/627)) ([cec7f1f](https://github.com/ory/kratos/commit/cec7f1fc4955b02d21d772e748ec791f31bad24e)):

    Added Login with Github Transcript

* Adds twitch oidc provider guide ([#760](https://github.com/ory/kratos/issues/760)) ([339e622](https://github.com/ory/kratos/commit/339e62202170bf21d469d1a2bfe6b053a78c374d))
* Bring oidc docs up to date ([7d0e470](https://github.com/ory/kratos/commit/7d0e47058cd6dca1763f01e45ed46cee49321240))
* Changed transcript location ([#642](https://github.com/ory/kratos/issues/642)) ([c52764d](https://github.com/ory/kratos/commit/c52764d4394181b24dffbf8301418530ba5dbcc2)):

    Changed the location so it is in the right place.

* Clarify 302 redirect on expired login flows ([ca31b53](https://github.com/ory/kratos/commit/ca31b53837e8eb2b811bf384da3724fdf61b423b))
* Clarify api flow use ([a38b4a1](https://github.com/ory/kratos/commit/a38b4a1684cfbc385ca21005c91a47e57df5a35d))
* Clarify feature-set ([2266ae7](https://github.com/ory/kratos/commit/2266ae7ea92207cdc4fcb58ef1384e287a5b34dc))
* Clarify kratos config snippet ([e7732f3](https://github.com/ory/kratos/commit/e7732f3283d82a1678076cd2463ef5ff33dd30ea))
* Clean up docs and correct samples ([8627ec5](https://github.com/ory/kratos/commit/8627ec58edb15118e0c4ce2cfcef7a5573482c5a))
* Complete registration documentation ([b3af02b](https://github.com/ory/kratos/commit/b3af02b0ea4cbf16ea282b7ce5f5057d99044ac3))
* Consistent formatting of badges ([#745](https://github.com/ory/kratos/issues/745)) ([b391a03](https://github.com/ory/kratos/commit/b391a036f3b49cd6c1915444c9f26dead4855a7c))
* Correct settings and verification redir ([30e25e7](https://github.com/ory/kratos/commit/30e25e7287a2579da99a6a6dc2f890e7e06fcc81))
* Docker image documentation ([#573](https://github.com/ory/kratos/issues/573)) ([bfe032e](https://github.com/ory/kratos/commit/bfe032e2b6bfd8b9415d466011bdd7e36efa4146))
* Document APi flows in self-service overview ([71ed0bd](https://github.com/ory/kratos/commit/71ed0bd2027d61c2e5cebf6b031fe66469bdf97e))
* Document how to check for login sessions ([9ad73b8](https://github.com/ory/kratos/commit/9ad73b8dab06c6796933448cb93ae4e55d9f2c51))
* Explain high-level API and browser flows ([fe3ee0a](https://github.com/ory/kratos/commit/fe3ee0a0c8681a99dc6b61b90cff547c6a7fc6d2))
* Fix logout url ([#593](https://github.com/ory/kratos/issues/593)) ([f0971d4](https://github.com/ory/kratos/commit/f0971d44a911caed8a6071358fa6b7ebc0fcf145))
* Fix sidebar missing comment ([d90123a](https://github.com/ory/kratos/commit/d90123ae31edbae6a39a1f039cc9362f9acdfdcb))
* Fix typo ([c2f94da](https://github.com/ory/kratos/commit/c2f94daa4143a70c13426ccd5366ec891182e4d0))
* Fix typo on index page ([#656](https://github.com/ory/kratos/issues/656)) ([907add5](https://github.com/ory/kratos/commit/907add5edb526adb4de57d35da16929ac08041e1))
* Fix url of admin-api /recovery/link ([#650](https://github.com/ory/kratos/issues/650)) ([e68c7cb](https://github.com/ory/kratos/commit/e68c7cbdc2191565570d0ee6812318ac9ad3421d))
* Fixed link ([c2aebbd](https://github.com/ory/kratos/commit/c2aebbd898f38388d849954938d56212c88d280f))
* Fixed link ([#629](https://github.com/ory/kratos/issues/629)) ([ad1276f](https://github.com/ory/kratos/commit/ad1276f2b2cf3cbbecba4dee1d6d433999286946))
* Fixed typos/readability ([#620](https://github.com/ory/kratos/issues/620)) ([7fd3ce0](https://github.com/ory/kratos/commit/7fd3ce0d8c52346ba3504ce5777321937baf8d1e)):

    Fixed a few typos, and moved some sentences around to improve readability.

* Fixed typos/readability ([#621](https://github.com/ory/kratos/issues/621)) ([c4fc75f](https://github.com/ory/kratos/commit/c4fc75f7dca59fa8f31d068f57179f49bf798b6a))
* Import mermaid ([#696](https://github.com/ory/kratos/issues/696)) ([6f75004](https://github.com/ory/kratos/commit/6f750047d41add6bd2d30adb1c654181c9636d2d))
* Improve charts and examples in self-service overview ([312c91d](https://github.com/ory/kratos/commit/312c91de3ae3c086f836ec3928735d787ad40dde))
* Improve documentation and add tests ([3dde956](https://github.com/ory/kratos/commit/3dde956e09d1f3f6411046b12f8684d8760f9b91))
* Improve long messages and render cli documentation ([e5fc02f](https://github.com/ory/kratos/commit/e5fc02ff22836e074a1dfca043d4b4b8ad64c747))
* Make assumptions neutral in concepts overview ([e89d980](https://github.com/ory/kratos/commit/e89d98099bd3fc5c8361f9015e44668494211152))
* Move development section ([2e6f643](https://github.com/ory/kratos/commit/2e6f6430f88105efd5618482043809c6d643216b))
* Move hooks ([c02b588](https://github.com/ory/kratos/commit/c02b58867ee2c0a386b2b741375ec8cd76122461))
* Move to json sidebar ([504af3b](https://github.com/ory/kratos/commit/504af3b89d728eb11bf42f4a2037c78b3b7cb788))
* Password login and registration methods for API clients ([5a44356](https://github.com/ory/kratos/commit/5a4435643ae3463df85458f22f87730c11af10ab))
* Prettify all files ([#743](https://github.com/ory/kratos/issues/743)) ([d9d1bfd](https://github.com/ory/kratos/commit/d9d1bfdff70ad835629a2dba00579925fcb3094d))
* Quickstart next steps ([#676](https://github.com/ory/kratos/issues/676)) ([ee9dd0d](https://github.com/ory/kratos/commit/ee9dd0d58a4146a0e131f6a7b74943bb39d26c0b)):

    Added a section outlining some easy config changes, that users can apply to the quickstart to test out different scenarios and configurations.

* Refactor login and registration documentation ([c660a04](https://github.com/ory/kratos/commit/c660a04ed6a70aefca18896662331fcc5d1919cf))
* Refactor settings and recovery documentation ([11ca9f7](https://github.com/ory/kratos/commit/11ca9f7d1b858dcda3a96e1e1d2607ba64f7fbbe))
* Refactor verification docs ([70f2789](https://github.com/ory/kratos/commit/70f2789363773fccc4bd8691597ff588ac6892c6))
* Regenerate clidocs with up-to-date binary ([e53289c](https://github.com/ory/kratos/commit/e53289c8e9f34a02ec66ec7ee03e2269a4a13c42))
* Remove `make tools` task ([ec6e664](https://github.com/ory/kratos/commit/ec6e6641234191d4eb39e1ad17bc7fcc03c2a0b5)), closes [#711](https://github.com/ory/kratos/issues/711) [#750](https://github.com/ory/kratos/issues/750):

    This task does not exist any more and the dependency building is much smarter now.

* Remove contraction ([#747](https://github.com/ory/kratos/issues/747)) ([cd4f21d](https://github.com/ory/kratos/commit/cd4f21dbfa2b3824468146677f542fbab2417c42))
* Remove duplicate word ([b84e659](https://github.com/ory/kratos/commit/b84e659af29aa1b129f33ccf5ca9e0d54353c019))
* Remove duplicate word ([#700](https://github.com/ory/kratos/issues/700)) ([a12100e](https://github.com/ory/kratos/commit/a12100e7644b535c4bd3073e03c48229bb81e7b2))
* Remove react native guide for now ([daa5f2e](https://github.com/ory/kratos/commit/daa5f2e3de3fe8380a91f594e034afcadc6e6ba5))
* Rename self service and add admin section ([639c424](https://github.com/ory/kratos/commit/639c424d3bde0557f7edd7edc489a476f1aa60b3))
* Replace ampersand ([#749](https://github.com/ory/kratos/issues/749)) ([8337b80](https://github.com/ory/kratos/commit/8337b80a13e8cf0cb2848241c93bb151420ac6a4))
* Resolve regression issues ([0470fd7](https://github.com/ory/kratos/commit/0470fd734fb30170033e10758d99cf5711c80eb1))
* Resolve typo in message IDs ([562cfc4](https://github.com/ory/kratos/commit/562cfc4392ba1c9c1fb8854ea0ac85bd44d0fac9))
* Resolve typo in message IDs ([#607](https://github.com/ory/kratos/issues/607)) ([f7688f0](https://github.com/ory/kratos/commit/f7688f0ab07b579a375ce4cc25361b360e82dd88))
* Update cli docs ([085efca](https://github.com/ory/kratos/commit/085efcae895b3aa3c76c819dca0f080ea79d57cd))
* Update link to mfa issue ([d03a706](https://github.com/ory/kratos/commit/d03a706307be21b83d18601223fb0d1430459a29))
* Update links ([a06fd88](https://github.com/ory/kratos/commit/a06fd88b0dcb747808ffea450bf1ac74dd941769))
* Update MFA link to issue ([#690](https://github.com/ory/kratos/issues/690)) ([7a744ad](https://github.com/ory/kratos/commit/7a744ad7b62540dd5789aee8532c1f97ddcab32d)):

    MFA issue was pushed to a later milestone. Update the documentation to point to the issue instead of the milestone.

* Update repository templates ([f422485](https://github.com/ory/kratos/commit/f4224852ceeb054405251b21895efa493e1abc9c))
* Update repository templates ([#678](https://github.com/ory/kratos/issues/678)) ([bdb6875](https://github.com/ory/kratos/commit/bdb6875e55aed454cda061969e1dd4f712e09bb5))
* Update sidebar ([ea15c20](https://github.com/ory/kratos/commit/ea15c2093fc66e4cfc0a66aabf7dfad6965777dc))
* Update ts examples ([65cb46e](https://github.com/ory/kratos/commit/65cb46e57595b920bd6544f9a9a4f7b886462be0))
* Use correct id for multi-domain-cookies ([b49288a](https://github.com/ory/kratos/commit/b49288a351647c91a3c7d4a62537146d4a9f1bd0))
* Use correct path in 0.4 docs ([9fcaac4](https://github.com/ory/kratos/commit/9fcaac4048e05500d0456eb3cd9cd11cc123e370)), closes [#588](https://github.com/ory/kratos/issues/588)
* Use NYT Capitalization for all Swagger headlines ([#675](https://github.com/ory/kratos/issues/675)) ([6c96429](https://github.com/ory/kratos/commit/6c9642959dab8cf042ad227711609d5726328394)), closes [#664](https://github.com/ory/kratos/issues/664)

### Features

* Add ability to configure session cookie domain/path ([faeb332](https://github.com/ory/kratos/commit/faeb3328dab343c6ef3974065ba0c5c590a8817e)), closes [#516](https://github.com/ory/kratos/issues/516)
* Add and improve settings testhelpers ([10a43fc](https://github.com/ory/kratos/commit/10a43fc518bd5c764712b549e6d35bf7159d757a))
* Add bearer helper ([ec6ca20](https://github.com/ory/kratos/commit/ec6ca20279d839dc10e7e3bc80e0442a630e586b))
* Add config version schema ([#608](https://github.com/ory/kratos/issues/608)) ([d218662](https://github.com/ory/kratos/commit/d218662388ef4fb7ea3bfee7b29c5cc8d34f1c8c)), closes [#590](https://github.com/ory/kratos/issues/590)
* Add discord oidc provider ([#767](https://github.com/ory/kratos/issues/767)) ([487296d](https://github.com/ory/kratos/commit/487296dd39d2e59d61b63f00f3d61fea9b8aed8c))
* Add enum to form field type ([96028d8](https://github.com/ory/kratos/commit/96028d8c80414cdcea177150ba6e986d0ecb29c6))
* Add flow type to login ([ce9133b](https://github.com/ory/kratos/commit/ce9133b0ff6d03738a5d27cf9c6a213496d75772))
* Add HTTP request flow validator ([1a6e847](https://github.com/ory/kratos/commit/1a6e84774b65ee7be9294baaaff77192cec8f0f2))
* Add new prometheus metrics endpoint [#672](https://github.com/ory/kratos/issues/672) ([#673](https://github.com/ory/kratos/issues/673)) ([0f5c436](https://github.com/ory/kratos/commit/0f5c436ce6e4aa78ca52ae63e58812e6703a1ab7)):

    Adds endpoint `/metrics` for prometheus metrics collection to the Admin API Endpoint.

* Add nocache helpers ([54dcc4d](https://github.com/ory/kratos/commit/54dcc4da2ff22bdb17e53dd6eac1c0bd54a20390))
* Add pagination tests ([e3aa81b](https://github.com/ory/kratos/commit/e3aa81b7da55108f43ea6e16c817c97e2f8a1d50))
* Add session token security definition ([d36c26f](https://github.com/ory/kratos/commit/d36c26f2edd66ddbd8338de4901957a9b9b7342e)):

    Adds the new Session Token as a Swagger security definition to allow setting the session token as a Bearer token when calling `/sessions/whoami`.

* Add stub errors to errorx ([5d452bb](https://github.com/ory/kratos/commit/5d452bb582e6a9e3b893424ec135d0cbdf875659)), closes [#610](https://github.com/ory/kratos/issues/610)
* Add test helper for fetching settings requests ([3646383](https://github.com/ory/kratos/commit/36463838d81d8b108aa9ded8c1ec6bc8f48f2267))
* Add tests and helpers to test recovery/verifiable addresses ([#579](https://github.com/ory/kratos/issues/579)) ([29979e6](https://github.com/ory/kratos/commit/29979e6c4934b71c7fb158cfa5b85e97be3ea8fc)), closes [#576](https://github.com/ory/kratos/issues/576)
* Add tests to cover auth ([c9d3a15](https://github.com/ory/kratos/commit/c9d3a1525cc74976d16b483e0ab5c48909b84022))
* Add texts for settings ([795548c](https://github.com/ory/kratos/commit/795548c25507c34c7fc37ce1c1a8ecc076c34ef4))
* Add the already declared (and settable) tracer as a middleware ([#614](https://github.com/ory/kratos/issues/614)) ([e24fffe](https://github.com/ory/kratos/commit/e24fffe3f13c353e3c07214c1e056a849533a9f6))
* Add token to session ([08c8c78](https://github.com/ory/kratos/commit/08c8c7837dbf799e6ba01d1820812c9e792d7850))
* Add type to all flows in SQL ([5515776](https://github.com/ory/kratos/commit/551577659f6a416ff6ef032c35af224b517df413))
* Allow import/validation of arrays ([d11ac32](https://github.com/ory/kratos/commit/d11ac32db6ddc0dce73067ffe7d4d0a734a3f991))
* Bump cli and migration render tasks ([6dcb42a](https://github.com/ory/kratos/commit/6dcb42a487476371a545b72f7ee7e820b815bbee))
* Finalize tests for registration flow refactor ([8e52c3a](https://github.com/ory/kratos/commit/8e52c3a99bd39b3429ff476340b5df49e0a85707))
* Finish off client cli ([36d60c7](https://github.com/ory/kratos/commit/36d60c7e7bc38d83726b4b4a3061ba6353dd1978))
* Implement administrative account recovery ([f5f9c43](https://github.com/ory/kratos/commit/f5f9c43e10dd3a9547e87776164d2d4a171f35ce))
* Implement API flow for recovery link method ([d65bf66](https://github.com/ory/kratos/commit/d65bf66781bdd2fae73e75c0ba39287b1575c45a))
* Implement API-based tests for password method settings flows ([60664aa](https://github.com/ory/kratos/commit/60664aaf05dbd6b228f420688d0171e5789246be))
* Implement max-age for session cookie ([2e642ff](https://github.com/ory/kratos/commit/2e642ff13c59a7e23babe9209c1a114ef0163bad)), closes [#326](https://github.com/ory/kratos/issues/326)
* Implement tests and anti-csrf for API settings flows ([8b8b6e5](https://github.com/ory/kratos/commit/8b8b6e5367e05f49950b851ea6834a9f18e896e7))
* Implement tests for new migrations ([e08ece9](https://github.com/ory/kratos/commit/e08ece9bb1c8c52580c15cf9152b4203821a0a0e))
* Improve test readability for password method ([a896d9b](https://github.com/ory/kratos/commit/a896d9b55596d2925941a6b6a91b8a6e4ef2caa1))
* Log successful hook execution ([f6026cf](https://github.com/ory/kratos/commit/f6026cfb0418767d99d18cd50529c2b71b21d775))
* Log successful hook execution ([1e7d044](https://github.com/ory/kratos/commit/1e7d044603b204632d2ec73c2e54db896992300b))
* Make login error handle JSON aware ([88f581f](https://github.com/ory/kratos/commit/88f581ff40a183cb96b5fb6d1ba398c58a9792d1))
* Make password settings method API-able ([0cf6027](https://github.com/ory/kratos/commit/0cf60274f87f098d5eb57531f5071cd407b65f4d))
* Make public cors configurable ([863a0d4](https://github.com/ory/kratos/commit/863a0d4f4696b05209b16f2e0c3daa9e8f4c1945)), closes [#712](https://github.com/ory/kratos/issues/712)
* Oidc provider claims config option ([#753](https://github.com/ory/kratos/issues/753)) ([bf94a40](https://github.com/ory/kratos/commit/bf94a40acd52128303c0b878ddb92d56abc4ceaf)), closes [#735](https://github.com/ory/kratos/issues/735)
* Reply with cache-control: 0 for browser-facing APIs ([1a45b53](https://github.com/ory/kratos/commit/1a45b5341e0ab4580208bfb6a505859d1e5d2faf)), closes [#360](https://github.com/ory/kratos/issues/360)
* Schemas are now static assets ([1776d58](https://github.com/ory/kratos/commit/1776d58278c42094b2c703e269a5901a96617051))
* Support and document api flow in session issuer hook ([91f3cc7](https://github.com/ory/kratos/commit/91f3cc7a559b1ea1279216f8dc81abd8e6f73776))
* Support application/json in registration ([3476b97](https://github.com/ory/kratos/commit/3476b978fdaee90358cc5505e20a0526f812a460)), closes [#44](https://github.com/ory/kratos/issues/44)
* Support custom session token header ([56bec76](https://github.com/ory/kratos/commit/56bec760fd1b94428ba296395a11358664d9e830)):

    The `/sessions/whoami` endpoint now accepts the ORY Kratos Session Token in the `X-Session-Token` HTTP header.

* Support GitLab OIDC Provider  ([#519](https://github.com/ory/kratos/issues/519)) ([8580d96](https://github.com/ory/kratos/commit/8580d96b7e345cc85a646f2945c3931f831afebf)), closes [#518](https://github.com/ory/kratos/issues/518)
* Support json payloads for login and password ([354e8b2](https://github.com/ory/kratos/commit/354e8b2cd63ee8feb1fd8a4ed8b033490155d90c))
* Support JSON payloads in password login flow ([dd32c23](https://github.com/ory/kratos/commit/dd32c23121da42e7eb3294fc8cb940fb7982723b))
* Support session token bearer auth and lifecycle ([c12600a](https://github.com/ory/kratos/commit/c12600a7243b541a91631169ec09d618a45c72dc)):

    This patch adds support for issuing, validating, and revoking session tokens. Session tokens carry a reference to a session, and are equal to session cookies but can be used on environments which do not support cookies (e.g. React Native) by sending them in the Bearer Authorization.

* Update migration tests ([fb28173](https://github.com/ory/kratos/commit/fb28173afa46ee828a3090981f394043c075f1ec))
* Use uri-reference for ui_url etc. to allow relative urls ([#617](https://github.com/ory/kratos/issues/617)) ([2dba450](https://github.com/ory/kratos/commit/2dba4503266436a615f4c1c18e07aa36ec713498))
* Write request -> flow rename migrations ([d7189a9](https://github.com/ory/kratos/commit/d7189a99c9d3e0ce33b4cc9846e6b2530ddfe5ec))

### Tests

* Add handler update tests ([aea1fb8](https://github.com/ory/kratos/commit/aea1fb807a16acd8406b94a72c3b39be8c3e1280)), closes [#325](https://github.com/ory/kratos/issues/325)
* Add init browser flow tests ([f477ece](https://github.com/ory/kratos/commit/f477ecebc73741b638cd62ef8aa2adb8b7adb8f2))
* Add test for no-cache on public router ([b8aa63b](https://github.com/ory/kratos/commit/b8aa63b7ebd269a87578e8a5c6b2df27e18f9efa))
* Add test for registration request ([79ed63c](https://github.com/ory/kratos/commit/79ed63cb4536499712796dab52999bcb73fe8466))
* Add tests for registration flows ([4772f71](https://github.com/ory/kratos/commit/4772f710f66d1ee36b52eca120d617a354f72413))
* Complete test suite for API-based auth ([fb9d62f](https://github.com/ory/kratos/commit/fb9d62f658165aa80bd117e1f827bbcc7c635150))
* Implement API login password tests ([8bfd5f2](https://github.com/ory/kratos/commit/8bfd5f294ff03280bcf01c5066acefe767eabc73))
* Implement API registration password tests ([db178b7](https://github.com/ory/kratos/commit/db178b73b097820c8dcd8760eec041a6fd0740aa))
* Replace e2e-memory with unit test ([52bd839](https://github.com/ory/kratos/commit/52bd839ea9fe8de1aac4663b9dc0a88ae18a5765)), closes [#580](https://github.com/ory/kratos/issues/580)
* Resolve broken decoder tests ([07add1b](https://github.com/ory/kratos/commit/07add1b3e4f46e4aff52174ce43d6970f60cf3ee))
* Use correct hook in test ([421320c](https://github.com/ory/kratos/commit/421320ca4ad5b346c6dfb6ef0a9d14d7cf23fded))

### Unclassified

* u ([e207a6a](https://github.com/ory/kratos/commit/e207a6adb98f639413accce383633d7e74ca4db9))
* As part of this change, fetching a settings flow over the public API no longer requires Anti-CSRF cookies to be sent. ([31d560e](https://github.com/ory/kratos/commit/31d560e47d55b087519355081cbca20b2a49da4e)), closes [#635](https://github.com/ory/kratos/issues/635)
* Create labels.json ([68b1f6f](https://github.com/ory/kratos/commit/68b1f6f5a35c66cc71f74f1473796fa16a852366))
* Add codedoc to identifier hint block ([6fe840f](https://github.com/ory/kratos/commit/6fe840f9c7a27ed97593e01936913e2239fd9446))
* Format ([e61a51d](https://github.com/ory/kratos/commit/e61a51dd6e2d5e003165a0b7906a9c86ebbc87d9))
* Format ([1e5b738](https://github.com/ory/kratos/commit/1e5b738f0765ec110c3ee70d7fc90fad0d1c89ac))
* Format code ([c3b5ff5](https://github.com/ory/kratos/commit/c3b5ff5d3bc3a1e72f48498fbed60bae9f159617))


# [0.4.6-alpha.1](https://github.com/ory/kratos/compare/v0.4.5-alpha.1...v0.4.6-alpha.1) (2020-07-13)

Resolves build and install issues and includes a few bugfixes.





### Bug Fixes

* Use proper binary name in dockerfile ([d36bbb0](https://github.com/ory/kratos/commit/d36bbb0875177ccd68747f4a17e59c981a7a6464))

### Code Generation

* Pin v0.4.6-alpha.1 release commit ([ad90e77](https://github.com/ory/kratos/commit/ad90e772cf59a33b213bc0fb782959a1685d9741)):

    Bumps from v0.4.4-alpha.1



# [0.4.5-alpha.1](https://github.com/ory/kratos/compare/v0.4.4-alpha.1...v0.4.5-alpha.1) (2020-07-13)

Resolves build and install issues and includes a few bugfixes.





### Bug Fixes

* Ensure default_browser_return_url for flows is configured in after ([#570](https://github.com/ory/kratos/issues/570)) ([cf9753c](https://github.com/ory/kratos/commit/cf9753c690c67e6401be52d2c1ce69f168aae6e8)), closes [#569](https://github.com/ory/kratos/issues/569)
* Require selfservice.default_browser_return_url to be set in config ([#571](https://github.com/ory/kratos/issues/571)) ([af2af7d](https://github.com/ory/kratos/commit/af2af7d35ba8b10dcd6d7636b044b0f7761a719d))

### Code Generation

* Pin v0.4.5-alpha.1 release commit ([3ea7fd3](https://github.com/ory/kratos/commit/3ea7fd3e7fd2c0b4aef638aa30e2b5b05c1bad26)):

    Bumps from v0.4.4-alpha.1



# [0.4.4-alpha.1](https://github.com/ory/kratos/compare/v0.4.3-alpha.1...v0.4.4-alpha.1) (2020-07-10)

The purpose of this release is to resolve issues with install scripts, homebrew, and scoop.





### Bug Fixes

* Detection of SQLite memory mode ([#564](https://github.com/ory/kratos/issues/564)) ([605cd57](https://github.com/ory/kratos/commit/605cd579895f3b765d398074cfdb37fa3eae0c4e))
* Improve goreleaser config ([0f8a0d8](https://github.com/ory/kratos/commit/0f8a0d8afa6489383800d3eff1b7b1da01fbef08))

### Code Generation

* Pin v0.4.4-alpha.1 release commit ([154d543](https://github.com/ory/kratos/commit/154d543eef29ab67be8637a96d8d06620974094f))

### Documentation

* Add description for subkeys of serve ([#562](https://github.com/ory/kratos/issues/562)) ([deae005](https://github.com/ory/kratos/commit/deae005a259747872f678d355b49cca21904e565))
* Add section about password expiry ([19c2414](https://github.com/ory/kratos/commit/19c2414c3defe79fe6e80e50dd0e85026ecd60e6))
* Specify the use of secrets ([#565](https://github.com/ory/kratos/issues/565)) ([7680450](https://github.com/ory/kratos/commit/7680450cfa44049759b27ec09d5bebc236b19a29))
* Update upgrade guide ([a40b1ec](https://github.com/ory/kratos/commit/a40b1ec18e7801f2862aad4e37becb7ce8f99c37))


# [0.4.3-alpha.1](https://github.com/ory/kratos/compare/v0.4.2-alpha.1...v0.4.3-alpha.1) (2020-07-08)

We are very happy to announce the 0.4 release of ORY Kratos with 163 commits and 817 changed files with 52,681 additions and 9,876 deletions.

There have been many improvements and bugfixes merged. The biggest changes are:

1. Account recovery ("reset password") has been implemented.
2. Documentation has been improved with easier to understand examples - currently only for account recovery so let us know what you think!
3. The configuration has been simplified a lot. It is now much easier to enable account recovery and email verification. This is a breaking change - please read the breaking changes section with care!
4. The Identity Traits JSON Schema has been renamed to the Identity JSON Schema. This is a breaking change - please read the breaking changes section with care!
5. `prompt=login` has been renamed to `refresh=true`. This is a breaking change - please read the breaking changes section with care!
6. We have reworked how (error) messages are returned. They now include an ID and all the parameters required for translating and customizing UI messages. This is a breaking change - please read the breaking changes section with care!
7. Instead of keeping track of `update_successful` with booleans, flows (e.g. the settings flow) that have more than one state now include a state machine. This is a breaking change - please read the breaking changes section with care!
8. Tons of tests have been added.
9. We have reworked and fully tested the migration pipeline to prevent breaking schema changes in future versions.
10. ORY Kratos now supports login with Azure AD and the Microsoft Identity Platform.

Before upgrading, please make a backup of your database and read the section "Breaking Changes" with care!





### Bug Fixes

* Resolve goreleaser build issues ([223571b](https://github.com/ory/kratos/commit/223571bca15f507067d20bedb104923331f88e59))
* Update install.sh script ([883d99b](https://github.com/ory/kratos/commit/883d99ba42de084018a32eaa094b5ae1a8ad4fc2))

### Code Generation

* Pin v0.4.3-alpha.1 release commit ([a3a34b1](https://github.com/ory/kratos/commit/a3a34b1e43b2d010ed85e098cd7cea31127df311)):

    Bumps from v0.4.0-alpha.1



# [0.4.2-alpha.1](https://github.com/ory/kratos/compare/v0.4.0-alpha.1...v0.4.2-alpha.1) (2020-07-08)

We are very happy to announce the 0.4 release of ORY Kratos with 153 commits and 760 changed files with 36,223 additions and 9,754 deletions.

There have been many improvements and bugfixes merged. The biggest changes are:

1. Account recovery ("reset password") has been implemented.
2. Documentation has been improved with easier to understand examples - currently only for account recovery so let us know what you think!
3. The configuration has been simplified a lot. It is now much easier to enable account recovery and email verification. This is a breaking change - please read the breaking changes section with care!
4. The Identity Traits JSON Schema has been renamed to the Identity JSON Schema. This is a breaking change - please read the breaking changes section with care!
5. `prompt=login` has been renamed to `refresh=true`. This is a breaking change - please read the breaking changes section with care!
6. We have reworked how (error) messages are returned. They now include an ID and all the parameters required for translating and customizing UI messages. This is a breaking change - please read the breaking changes section with care!
7. Instead of keeping track of `update_successful` with booleans, flows (e.g. the settings flow) that have more than one state now include a state machine. This is a breaking change - please read the breaking changes section with care!
8. Tons of tests have been added.
9. We have reworked and fully tested the migration pipeline to prevent breaking schema changes in future versions.
10. ORY Kratos now supports login with Azure AD and the Microsoft Identity Platform.

Before upgrading, please make a backup of your database and read the section "Breaking Changes" with care!





### Bug Fixes

* Ignore pkged generated files ([1d385e4](https://github.com/ory/kratos/commit/1d385e4d1a004405099242c3003006d1713a24c6))

### Code Generation

* Pin v0.4.2-alpha.1 release commit ([20024cb](https://github.com/ory/kratos/commit/20024cbbb44b4f556004ef752a7f37e70a070e6a)):

    Bumps from v0.4.0-alpha.1



# [0.4.0-alpha.1](https://github.com/ory/kratos/compare/v0.3.0-alpha.1...v0.4.0-alpha.1) (2020-07-08)

We are very happy to announce the 0.4 release of ORY Kratos with 153 commits and 760 changed files with 36,223 additions and 9,754 deletions.

There have been many improvements and bugfixes merged. The biggest changes are:

1. Account recovery ("reset password") has been implemented.
2. Documentation has been improved with easier to understand examples - currently only for account recovery so let us know what you think!
3. The configuration has been simplified a lot. It is now much easier to enable account recovery and email verification. This is a breaking change - please read the breaking changes section with care!
4. The Identity Traits JSON Schema has been renamed to the Identity JSON Schema. This is a breaking change - please read the breaking changes section with care!
5. `prompt=login` has been renamed to `refresh=true`. This is a breaking change - please read the breaking changes section with care!
6. We have reworked how (error) messages are returned. They now include an ID and all the parameters required for translating and customizing UI messages. This is a breaking change - please read the breaking changes section with care!
7. Instead of keeping track of `update_successful` with booleans, flows (e.g. the settings flow) that have more than one state now include a state machine. This is a breaking change - please read the breaking changes section with care!
8. Tons of tests have been added.
9. We have reworked and fully tested the migration pipeline to prevent breaking schema changes in future versions.
10. ORY Kratos now supports login with Azure AD and the Microsoft Identity Platform.

Before upgrading, please make a backup of your database and read the section "Breaking Changes" with care! This release requires running SQL migrations when upgrading!



## Breaking Changes

This patch renames the Identity Traits JSON Schema to Identity JSON Schema.

The identity payload has changed from

```
 {
-  "traits_schema_url": "...",
-  "traits_schema_id": "...",
+  "schema_url": "...",
+  "schema_id": "...",
 }
```

Additionally, it is now expected that your Identity JSON Schema includes a "traits" key at the
root level.

**Before (example)**

```
{
  "$id": "https://schemas.ory.sh/presets/kratos/quickstart/email-password/identity.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "email": {
      "type": "string",
      "format": "email",
      "title": "E-Mail",
      "minLength": 3,
      "ory.sh/kratos": {
        "credentials": {
          "password": {
            "identifier": true
          }
        },
        "verification": {
          "via": "email"
        },
        "recovery": {
          "via": "email"
        }
      }
    }
  },
  "required": [
    "email"
  ],
  "additionalProperties": false
}
```

**After (example)**

```
{
  "$id": "https://schemas.ory.sh/presets/kratos/quickstart/email-password/identity.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "email": {
          "type": "string",
          "format": "email",
          "title": "E-Mail",
          "minLength": 3,
          "ory.sh/kratos": {
            "credentials": {
              "password": {
                "identifier": true
              }
            },
            "verification": {
              "via": "email"
            },
            "recovery": {
              "via": "email"
            }
          }
        }
      },
      "required": [
        "email"
      ],
      "additionalProperties": false
    }
  }
}
```

You also need to remove the `traits` key from your ORY Kratos config like this:

```
 identity:
-   traits:
-     default_schema_url: http://test.kratos.ory.sh/default-identity.schema.json
-     schemas:
-       - id: other
-         url: http://test.kratos.ory.sh/other-identity.schema.json
+   default_schema_url: http://test.kratos.ory.sh/default-identity.schema.json
+   schemas:
+     - id: other
+       url: http://test.kratos.ory.sh/other-identity.schema.json
```

Do not forget to also update environment variables for the Identity JSON Schema as well if set.

To address these refactorings, the configuration had to be changed and with breaking changes
as keys have moved or have been removed.

Hook configuration has also changed. It is no longer required to include hooks such as `verification` to get
verification working. Instead, verification is enabled globally (`selfservice.flows.verification.enabled`).
Also, the `redirect` hook has been removed as it lead to confusion because there are already default redirect
URLs configurable. You will find more information in the details below.

**Session Management**

```diff
-ttl:
-  session: 1h
-security:
-  session:
-    cookie:
-      same_site: Lax
+session:
+  lifespan: 1h
+  cookie_same_site: Lax
```

**Secrets**

```diff
-secrets:
-  session:
-    - secret-to-encrypt-session-cookies
-    - old-session-cookie-secret-that-has-been-rotated
+secrets:
+  default:
+    # This secret is used as default and will also be used for encrypting e.g. cookies when a dedicated cookie secret (as shown below) is not defined.
+    - default-secret-to-encrypt-stuff
+  cookie:
+    - secret-to-encrypt-session-cookies
+    - old-session-cookie-secret-that-has-been-rotated
```

**URLs**

The Base URL configuration has moved to `serve.public` and `serve.admin`. They are also no longer required and fall
back to defaults based on the machine's hostname, port configuration, and other settings:

```diff
-urls:
-  self:
-    public: https://kratos.my-website.com/
-    admin: https://admin.kratos.cluster.localnet/
+serve:
+  public:
+    base_url: https://kratos.my-website.com/
+  admin:
+    base_url: https://admin.kratos.cluster.localnet/
```

The UI URLs have moved from `urls` to their respective self-service flows:

```diff
-urls:
-  login_ui: http://127.0.0.1:4455/auth/login
-  registration_ui: http://127.0.0.1:4455/auth/registration
-  settings_ui: http://127.0.0.1:4455/settings
-  verify_ui: http://127.0.0.1:4455/verify
-  error_ui: http://127.0.0.1:4455/error
+selfservice:
+  flows:
+    login:
+      ui_url: http://127.0.0.1:4455/auth/login
+    registration:
+      ui_url: http://127.0.0.1:4455/auth/registration
+    settings:
+      ui_url: http://127.0.0.1:4455/settings
+    # please note that `verify` has changed to `verification`!
+    verification:
+      ui_url: http://127.0.0.1:4455/verify
+    error:
+      ui_url: http://127.0.0.1:4455/error
```

The default redirect URL as well as whitelisted redirect URLs have also changed their location:

```diff
-urls:
-  default_return_to: https://self-service/dashboard
-  whitelisted_return_to_urls:
-    - https://self-service/some-other-url
-    - https://example.org/another-url
+selfservice:
+  default_browser_return_url: https://self-service/dashboard
#  Please note that the `to` has been removed (`whitelisted_return_to_urls` -> `whitelisted_return_urls`)
+  whitelisted_return_urls:
+    - https://self-service/some-other-url
+    - https://example.org/another-url
```

**Self-Service Login**

`selfservice.login` has moved to `selfservice.flow.login`:

```diff
 selfservice:
-  login:
+  flows:
+    login:
```

On top of this change, a few keys under `login` have changed as well:

```diff
 selfservice
   flows:
     login:
+      ui_url: http://127.0.0.1:4455/auth/login
       request_lifespan: 99m
-      before:
-        hooks:
-          - hook: redirect
-            config:
-              default_redirect_url: http://test.kratos.ory.sh:4000/
-              allow_user_defined_redirect: false
+      # The before hooks have been removed because there were no good use cases for them. If
+      # this is a problem for you feel free to open an issue!

     after:
-      default_return_to: https://self-service/login/return_to
+      default_browser_return_url: https://self-service/login/return_to
       password:
-         default_return_to: https://self-service/login/password/return_to
+         default_browser_return_url: https://self-service/login/password/return_to
          hooks:
            - hook: revoke_active_sessions
       oidc:
-         default_return_to: https://self-service/login/podc/return_to
+         default_browser_return_url: https://self-service/login/podc/return_to
          hooks:
            - hook: revoke_active_sessions
```

**Self-Service Registration**

`selfservice.registration` has moved from to `selfservice.flow.registration`:

```diff
 selfservice:
-  registration:
+  flows:
+    registration:
```

On top of this change, a few keys under `registration` have changed as well:

```diff
 selfservice
   flows:
     registration:
+      ui_url: http://127.0.0.1:4455/auth/registration
       request_lifespan: 99m
-      before:
-        hooks:
-          - hook: redirect
-            config:
-              default_redirect_url: http://test.kratos.ory.sh:4000/
-              allow_user_defined_redirect: false
+      # The before hooks have been removed because there were no good use cases for them. If
+      # this is a problem for you feel free to open an issue!

     after:
-      default_return_to: https://self-service/registration/return_to
+      default_browser_return_url: https://self-service/registration/return_to
       password:
-        default_return_to: https://self-service/registration/password/return_to
+        default_browser_return_url: https://self-service/registration/password/return_to
         hooks:
           - hook: revoke_active_sessions
+          # The verify hook is now executed automatically when verification is turned on.
-          - hook: verify
+          # The redirect hook was confusing as it aborts the registration flow and does not solve redirection on
+          # success. It has thus been removed.
-          - hook: redirect
       oidc:
-        default_return_to: https://self-service/registration/podc/return_to
+        default_browser_return_url: https://self-service/registration/podc/return_to
         hooks:
           - hook: revoke_active_sessions
+          # The verify hook is now executed automatically when verification is turned on.
-          - hook: verify
+          # The redirect hook was confusing as it aborts the registration flow and does not solve redirection on
+          # success. It has thus been removed.
-          - hook: redirect
```

**Self-Service Settings**

`selfservice.settings` has moved from to `selfservice.flow.settings`:

```diff
 selfservice:
-  settings:
+  flows:
+    settings:
```

On top of this change, a few keys under `settings` have changed as well:

```diff
 selfservice
   flows:
     settings:
+      ui_url: http://127.0.0.1:4455/settings
       request_lifespan: 99m
       privileged_session_max_age: 99m
-      default_return_to: https://self-service/settings/return_to
       after:
+        default_browser_return_url: https://self-service/settings/return_to
+        # The profile/password after hooks have been removed as verification is now executed automatically
+        # when turned on.
-        password:
-          hooks:
-            - hook: verify
-        profile:
-          hooks:
-            - hook: verify
```

**Self-Service Verification**

`selfservice.verify` has moved from to `selfservice.flow.verification`:

```diff
 selfservice:
-  verify:
+  flows:
+    verification:
```

Instead of configuring verification with hooks and other components, it can now be enabled
in a central place. If enabled, a SMTP server must be configured in the `courier` section.
You are still required to mark a field as verifiable in your Identity JSON Schema.

```diff
 selfservice:
   flows:
     verification:
+      enabled: true # defaults to true
+      ui_url: http://127.0.0.1:4455/recovery
       request_lifespan: 1m
-      default_return_to: https://self-service/verification/return_to
       after:
+        default_browser_return_url: https://self-service/verification/return_to
```

Replaces the `update_successful` field of the settings request
with a field called `state` which can be either `show_form` or `success`.

Flows, request methods, form fields have had a key errors to show e.g. validation errors such as ("not an email address", "incorrect username/password", and so on. The `errors` key is now called `messages`. Each message now has a `type` which can be `error` or `info`, an `id` which can be used to translate messages, a `text` (which was previously errors[*].message). This affects all login, request, settings, and recovery flows and methods.

To refresh a login session it is now required to append `refresh=true` instead of `prompt=login` as the second has implications for revoking an existing issue and might be confusing when used in combination with OpenID Connect.

* Applying this patch requires running SQL Migrations.
* The field `identity.addresses` has moved to `identity.verifiable_addresses`.
* Configuration key `selfservice.verification.link_lifespan`
has been merged with  `selfservice.verification.request_lifespan`.



### Bug Fixes

* Account recovery can't use recovery token ([#526](https://github.com/ory/kratos/issues/526)) ([379f24e](https://github.com/ory/kratos/commit/379f24e96e50a3e5c71b53a11195bdd84a8dc957)), closes [#525](https://github.com/ory/kratos/issues/525)
* Add and document recovery to quickstart ([c229c54](https://github.com/ory/kratos/commit/c229c54603bdc3efb863fd76b64096ae599d1aac))
* Add pkger to docker builds ([d3ef5a0](https://github.com/ory/kratos/commit/d3ef5a0fe90f430999d0d94cb2f55acc8d628212))
* Allow linking oidc credentials without existing oidc connection ([#548](https://github.com/ory/kratos/issues/548)) ([39c1234](https://github.com/ory/kratos/commit/39c1234f8ff3f6c7b0923053c8a317677d6cb667)), closes [#532](https://github.com/ory/kratos/issues/532)
* Bump pop version ([#558](https://github.com/ory/kratos/issues/558)) ([9e46cea](https://github.com/ory/kratos/commit/9e46ceabec8d5c1995321b62cbba9ac3900de446)), closes [#556](https://github.com/ory/kratos/issues/556)
* Clear error messages after updating settings successfully ([#421](https://github.com/ory/kratos/issues/421)) ([7eec388](https://github.com/ory/kratos/commit/7eec38829449237cffe345d8bec67578764559be)), closes [#420](https://github.com/ory/kratos/issues/420)
* Do not send debug on session/whoami ([16d3670](https://github.com/ory/kratos/commit/16d3670070bf46170c4540203e8380ad81bfb4c3)), closes [#483](https://github.com/ory/kratos/issues/483)
* Document login refresh parameter in swagger ([#482](https://github.com/ory/kratos/issues/482)) ([6b94993](https://github.com/ory/kratos/commit/6b949936725a6100a31851a5d879c877c2c76cbf))
* Embedded video link properly ([#514](https://github.com/ory/kratos/issues/514)) ([962bbc6](https://github.com/ory/kratos/commit/962bbc6e4af0797c190418b812f6298372dabdde))
* Embedded video link properly ([#515](https://github.com/ory/kratos/issues/515)) ([821ca93](https://github.com/ory/kratos/commit/821ca93838a360551378e336e9ce10cfe13369ec))
* Enable recovery for quickstart ([0ccc651](https://github.com/ory/kratos/commit/0ccc651f809b1e39dd6c41b88f1a10c67451eae2))
* Improve grammar of similar password error ([#471](https://github.com/ory/kratos/issues/471)) ([39873bf](https://github.com/ory/kratos/commit/39873bfad89a654fe12e101b54e9b0c2f95714ec))
* Improvements to Dockerfiles ([#552](https://github.com/ory/kratos/issues/552)) ([6023877](https://github.com/ory/kratos/commit/6023877184efeadd6ec27a050a6969b6d0dd6caa)):

    - expose ory home as volume to simplify passing in own config file
    - declare Kratos default ports in Dockerfile

* Initialize verification request with correct state ([3264ecf](https://github.com/ory/kratos/commit/3264ecfbb8f7b34d9dbb22237df8d9f591ac09f3)), closes [#543](https://github.com/ory/kratos/issues/543)
* Re-add all databases to persister ([#527](https://github.com/ory/kratos/issues/527)) ([b04d178](https://github.com/ory/kratos/commit/b04d17815b5a28b5fe73a6a94ce1d907a63115e1))
* Re-add redirect targets for quickstart ([3c48ad2](https://github.com/ory/kratos/commit/3c48ad26961560d6e10a627a64052e316d9ffdc7))
* Reduce docker bloat by ignoring docs and others ([ecc555b](https://github.com/ory/kratos/commit/ecc555b5ad0fa888a8d5ba39cc09094fd251e655))
* Resolve broken redirect in verify flow ([a9ca8fd](https://github.com/ory/kratos/commit/a9ca8fd793347ed8e4404a4bd29e330a3f1ef684)), closes [#436](https://github.com/ory/kratos/issues/436)
* Respect multiple secrets and fix used flag ([#526](https://github.com/ory/kratos/issues/526)) ([b16c2b8](https://github.com/ory/kratos/commit/b16c2b80edfc78afca0c72fa8da7d73b51b3075a)), closes [#525](https://github.com/ory/kratos/issues/525)
* Respect self-service enabled flag ([#470](https://github.com/ory/kratos/issues/470)) ([b198faf](https://github.com/ory/kratos/commit/b198fafce9d96fbb644300243e6a757242fbbd06)), closes [#417](https://github.com/ory/kratos/issues/417):

    Respects the `enabled` flag for self-service strategies.
    
    Also a new testhelper function was needed, to defer route registration
    (because whether strategies are enabled or not is determined only once:
    at route registration)

* Typo accent -> account ([984d978](https://github.com/ory/kratos/commit/984d978cf44763d916a9329742d046e00f21577b))
* Use correct brew replacements ([fd269b1](https://github.com/ory/kratos/commit/fd269b1afa784becac7ee79cd7a6f9d2bbe39121)), closes [#423](https://github.com/ory/kratos/issues/423)
* Write migration tests ([#499](https://github.com/ory/kratos/issues/499)) ([d32413a](https://github.com/ory/kratos/commit/d32413a1fcd0ce1a82d2529f18b5d4334a490a2a)), closes [#481](https://github.com/ory/kratos/issues/481)

### Code Generation

* Pin v0.4.0-alpha.1 release commit ([e8690c4](https://github.com/ory/kratos/commit/e8690c4037ba5d80aa2459625be553c5bc2d2152))

### Code Refactoring

* Improve and simplify configuration ([#536](https://github.com/ory/kratos/issues/536)) ([8e7f9f5](https://github.com/ory/kratos/commit/8e7f9f5ec3ac6f5675584974e8d189247b539634)), closes [#432](https://github.com/ory/kratos/issues/432)
* Move schema packing to pkger ([173f9d2](https://github.com/ory/kratos/commit/173f9d2b09d597376490b5d4588f7c0a4f525857))
* Move verify fallback to verification ([1ce6469](https://github.com/ory/kratos/commit/1ce64695ec61c3a31e00875069d2847be502744b))
* Rename identity traits schema to identity schema  ([#557](https://github.com/ory/kratos/issues/557)) ([949e743](https://github.com/ory/kratos/commit/949e743ef9ddbc6e711f0174593f59f4fa3a1171)), closes [#531](https://github.com/ory/kratos/issues/531)
* Rename prompt=login to refresh=true ([#478](https://github.com/ory/kratos/issues/478)) ([c04346e](https://github.com/ory/kratos/commit/c04346e0f01aa7ce5627c0b7135032b225e7faf9)), closes [#477](https://github.com/ory/kratos/issues/477)
* Replace settings update_successful with state ([#488](https://github.com/ory/kratos/issues/488)) ([ca3b3f4](https://github.com/ory/kratos/commit/ca3b3f4dbdcd75ceb13c9a1b2c8dc991aba7c7e4)), closes [#449](https://github.com/ory/kratos/issues/449)
* Text errors to text messages ([#476](https://github.com/ory/kratos/issues/476)) ([8106951](https://github.com/ory/kratos/commit/81069514e5ef1d851f76d44bb45d6a896d4985a6)), closes [#428](https://github.com/ory/kratos/issues/428):

    This patch implements a better way to deal with text messages by giving them a unique ID, a context, and a default message.


### Documentation

* Add azure to next docs ([e1dd3fa](https://github.com/ory/kratos/commit/e1dd3fad30a07be6f105201a8478642e9792df46))
* Add fixme note for viper workaround ([7e3eef6](https://github.com/ory/kratos/commit/7e3eef6d36dcbb1a06ce0a20e2de0874a7dc5d38)):

    See https://github.com/ory/x/issues/169

* Add guide for setting up account recovery ([bbf3762](https://github.com/ory/kratos/commit/bbf37620d5b47fd18cb754c8ed43856652ee33c0))
* Add guide for setting up email verification ([1435cbc](https://github.com/ory/kratos/commit/1435cbcea5d45c9cde1a0eb7e5ebb66ce65c4b82))
* Add guide for SSO via Google ([#424](https://github.com/ory/kratos/issues/424)) ([5c45b16](https://github.com/ory/kratos/commit/5c45b1653791cc3ab5d4e4694da98da7543e816d))
* Add new guides to sidebar ([24c5cbc](https://github.com/ory/kratos/commit/24c5cbc129ad185ec02883c3451d7e573409b865))
* Added video tutorials to guides ([#513](https://github.com/ory/kratos/issues/513)) ([956731d](https://github.com/ory/kratos/commit/956731d562f33f2849197b2e692a4f20b18279f9))
* Added youtube manual ([#490](https://github.com/ory/kratos/issues/490)) ([ec232f7](https://github.com/ory/kratos/commit/ec232f72d7204b2cdf946874d51f7473a10a76a4))
* Connecting Kratos to AzureAD ([#433](https://github.com/ory/kratos/issues/433)) ([7660bcd](https://github.com/ory/kratos/commit/7660bcd2ba90d83c4ab0683a2f011e6841b2c810))
* Correct claims.email in github guide ([#422](https://github.com/ory/kratos/issues/422)) ([052a622](https://github.com/ory/kratos/commit/052a622de79d34e32ccab9c7da12a1275c7be51b)):

    There is no email_primary in claims, and the selfservice strategy is currently using claims.email.

* Correct claims.email in github guide ([#422](https://github.com/ory/kratos/issues/422)) ([58f7e15](https://github.com/ory/kratos/commit/58f7e15093d2461d4322fe68adb0723ae244bed9)):

    There is no email_primary in claims, and the selfservice strategy is currently using claims.email.

* Correct link in user-settings ([d13317d](https://github.com/ory/kratos/commit/d13317d9bf71db775067a7c17f4c98cdbf1cc7e5))
* Correct SDK use in quickstart ([#480](https://github.com/ory/kratos/issues/480)) ([dfdf975](https://github.com/ory/kratos/commit/dfdf9751d9333994a49537d82a15b780ebd8bc76)), closes [#430](https://github.com/ory/kratos/issues/430)
* Correct stray dot ([e820f41](https://github.com/ory/kratos/commit/e820f41e63aff1a85094a9e14dfd968353ae6b1b))
* Correct user settings render form ([197e246](https://github.com/ory/kratos/commit/197e24603fc67707131e54e52e1bfb52011ca839))
* Delete old redirect homepage ([b6d9244](https://github.com/ory/kratos/commit/b6d9244b5d683f5baf27e9af5970596261a4fd20))
* Document new account recovery feature ([2252a86](https://github.com/ory/kratos/commit/2252a8676e573b9ade85814acc40b212dcfd48c1)), closes [#436](https://github.com/ory/kratos/issues/436)
* Document refresh=true for login ([#479](https://github.com/ory/kratos/issues/479)) ([2ab5ead](https://github.com/ory/kratos/commit/2ab5ead77517ab5b750835195ab6673e219da71a)), closes [#464](https://github.com/ory/kratos/issues/464)
* Embedded quickstart video ([#491](https://github.com/ory/kratos/issues/491)) ([ee80346](https://github.com/ory/kratos/commit/ee80346a30ebc2c7b06292e58bd3578e002e242a))
* Fix broken link ([d20816e](https://github.com/ory/kratos/commit/d20816e5335abb8bcde5c6d68b17eaabae5d01b0))
* Fix broken link ([aa9d3e6](https://github.com/ory/kratos/commit/aa9d3e6347375170a84ba53b2a9050c9544e7e2a))
* Fix broken link ([#506](https://github.com/ory/kratos/issues/506)) ([dac8dfd](https://github.com/ory/kratos/commit/dac8dfd970255f8e79e7fc7811f563e6903f6fc9)):

    The rest api is no longer under sdk but under reference.

* Fix broken link ([#554](https://github.com/ory/kratos/issues/554)) ([e80d691](https://github.com/ory/kratos/commit/e80d691e256326aacfa89b391583e0494d8a6872))
* Fix code sample comment ([781a76b](https://github.com/ory/kratos/commit/781a76bb6de20767d6150b1fcb5236f4f376edd7))
* Fix copy paste errors in code docs ([e456a4e](https://github.com/ory/kratos/commit/e456a4e435265eade7026fd899c4bc7d2b28a5c9))
* Fix iframe syntax ([#520](https://github.com/ory/kratos/issues/520)) ([0cb36ca](https://github.com/ory/kratos/commit/0cb36ca9d8459dc8027358190e6e8aa8764bffe4))
* Fix typo ([#535](https://github.com/ory/kratos/issues/535)) ([c57d270](https://github.com/ory/kratos/commit/c57d270758a97315c874df3fae867b0031300501))
* Fix typo in base docs ([#503](https://github.com/ory/kratos/issues/503)) ([6668048](https://github.com/ory/kratos/commit/666804812d707b1d50ea160877bdb3878ddfe6b0))
* Fix typo in oauth sign in documentation ([#504](https://github.com/ory/kratos/issues/504)) ([886e24d](https://github.com/ory/kratos/commit/886e24d93a5eb233062b8c7d562c8208f7a4f48f))
* Fix typos ([81903a5](https://github.com/ory/kratos/commit/81903a5137d87588531391623b92afde70abc3ea))
* Fix typos ([#489](https://github.com/ory/kratos/issues/489)) ([57a7bc8](https://github.com/ory/kratos/commit/57a7bc89961612fea0255202d3dd6a535921ef3c))
* Fix ui url keys everywhere ([b75debb](https://github.com/ory/kratos/commit/b75debb0ee4f87dd9910b30bd76d8c6ad382fb38))
* Fix username example by renaming property and removing format ([#508](https://github.com/ory/kratos/issues/508)) ([4573426](https://github.com/ory/kratos/commit/45734260bcead3087aadcaaf3033cc1e89bc1844))
* Fix wording in settings flow graph ([e2a0084](https://github.com/ory/kratos/commit/e2a00842cb5bd3cfbddd0e5117c7f3f968e9f2df))
* Fixed broken link ([#452](https://github.com/ory/kratos/issues/452)) ([d1ddbd1](https://github.com/ory/kratos/commit/d1ddbd1ee465a7d3e29815fcfd9c75b5decbb5f9))
* Fixed broken link ([#455](https://github.com/ory/kratos/issues/455)) ([4f3d179](https://github.com/ory/kratos/commit/4f3d17906f3fa2aea3a0b0505047da6aa54938e4))
* Fixed broken link ([#456](https://github.com/ory/kratos/issues/456)) ([4b43e99](https://github.com/ory/kratos/commit/4b43e993df62d2bf54fa39624651f081eb75bbb0))
* Fixed broken link ([#460](https://github.com/ory/kratos/issues/460)) ([7da304c](https://github.com/ory/kratos/commit/7da304caf0de93442f047872cdd30d7fc316218e))
* Fixed broken link ([#461](https://github.com/ory/kratos/issues/461)) ([c248e4e](https://github.com/ory/kratos/commit/c248e4e2a48a409b53ed02644abfc27e3cebeb11))
* Fixed broken link ([#462](https://github.com/ory/kratos/issues/462)) ([ceacac3](https://github.com/ory/kratos/commit/ceacac30eda7d94cb24403c1fb988d4dd5fcd21f))
* Fixed broken links ([#451](https://github.com/ory/kratos/issues/451)) ([193a781](https://github.com/ory/kratos/commit/193a781576031818006d6e2b72418293cf94dda1)):

    Fixed a few broken links, .md in the url was the problem.

* Fixed broken links ([#453](https://github.com/ory/kratos/issues/453)) ([59d00eb](https://github.com/ory/kratos/commit/59d00ebb87564cc9ff9c5ae12bcd7d25fb0b26c9))
* Fixed broken links ([#457](https://github.com/ory/kratos/issues/457)) ([00ec00d](https://github.com/ory/kratos/commit/00ec00d09ca5318c75832caff5e7a97d640ac083))
* Fixed broken links ([#458](https://github.com/ory/kratos/issues/458)) ([f960887](https://github.com/ory/kratos/commit/f9608876e30dbdd7c67ee70dcf5d9a1985b80f0f))
* Fixed broken links ([#459](https://github.com/ory/kratos/issues/459)) ([2749596](https://github.com/ory/kratos/commit/27495964c7cd34e9bf914b19c83157e484c9cde4))
* Fixed broken markdown ([#474](https://github.com/ory/kratos/issues/474)) ([22d5be1](https://github.com/ory/kratos/commit/22d5be16f91ed9df206310c6f04d843cd79328ca))
* Format guides ([407c70f](https://github.com/ory/kratos/commit/407c70f23d815380d98ee9252f263e07c1f0f4a9))
* Improve grammar and wording ([#448](https://github.com/ory/kratos/issues/448)) ([a19adf3](https://github.com/ory/kratos/commit/a19adf30426ff8df03a3eb725ae0101ebb6c4ab1))
* Improve grammar, clarify sections, update images ([#419](https://github.com/ory/kratos/issues/419)) ([79019d1](https://github.com/ory/kratos/commit/79019d1246b1517b3297996a207a3d2f517fab01))
* Make whitelisted_return_to_urls examples an array ([#426](https://github.com/ory/kratos/issues/426)) ([7ed5605](https://github.com/ory/kratos/commit/7ed56057f533f23ca18cab5a2614429554e877e2)), closes [#425](https://github.com/ory/kratos/issues/425)
* Minor fixes ([#467](https://github.com/ory/kratos/issues/467)) ([8d15307](https://github.com/ory/kratos/commit/8d153079ee44f0765993640500bbe746dc0a34aa))
* Move security questions to own document ([2b77fba](https://github.com/ory/kratos/commit/2b77fba79b724dcd68ff0cd739cd65517aea4325))
* Properly annotate forms disabled field ([#486](https://github.com/ory/kratos/issues/486)) ([be1acb3](https://github.com/ory/kratos/commit/be1acb3d161412d18599c970364f0c91fa6ebffb)):

    See https://github.com/ory/kratos/pull/467#discussion_r434764266

* Remove rogue slash and fix closing tag ([#521](https://github.com/ory/kratos/issues/521)) ([3fd1076](https://github.com/ory/kratos/commit/3fd1076929eeecffb7e8aa8e906970774283daeb))
* Rename redirect page to browser-redirect-flow-completion ([ae77d48](https://github.com/ory/kratos/commit/ae77d48a3435069556382b9403cb1ad45a9d7c07))
* Replace mailhog references with mailslurper ([#509](https://github.com/ory/kratos/issues/509)) ([d0e5a0f](https://github.com/ory/kratos/commit/d0e5a0fa64e2d46437fb2abd17dc306bdec34a91))
* Run format ([2b3f299](https://github.com/ory/kratos/commit/2b3f29913be844498a02b9869789c2b2d4aaacf8))
* Typo correction in credentials.md ([#551](https://github.com/ory/kratos/issues/551)) ([3b7e104](https://github.com/ory/kratos/commit/3b7e104c2bcba52326f89761c9e3da14b4f06d08))
* Typos and stale links ([29fb466](https://github.com/ory/kratos/commit/29fb466d9881b6574ee697d7e25e45785f07114b))
* Typos and stale links ([#510](https://github.com/ory/kratos/issues/510)) ([7557ab8](https://github.com/ory/kratos/commit/7557ab85ddf8501935d70e2558682dff2024897b))
* Update repository templates ([4c89834](https://github.com/ory/kratos/commit/4c89834ce59195c5b59da5bc5b41db7ed03bf1c4))
* Use central banner repo for README ([d1e8a82](https://github.com/ory/kratos/commit/d1e8a8272cd536b6e12326778258bfbe0b7e8af7))
* Use shorthand closing tag for Mermaid ([f9f2dbc](https://github.com/ory/kratos/commit/f9f2dbc063f82a852b540013ddff81501f7c1222))

### Features

* Add support for Multitenant Azure AD as an OIDC provider ([#434](https://github.com/ory/kratos/issues/434)) ([a8f1179](https://github.com/ory/kratos/commit/a8f117985217c753cfca52905e43b640e89a6bd1))
* Add tests for defaults ([a16fc51](https://github.com/ory/kratos/commit/a16fc5121b36353cf2e684190eda976a1ea53a8f))
* Add User ID to a header when calling whoami ([#530](https://github.com/ory/kratos/issues/530)) ([183b4d0](https://github.com/ory/kratos/commit/183b4d075a9ff50c1f9f53d108a48789e49a5138))
* Implement account recovery ([#428](https://github.com/ory/kratos/issues/428)) ([e169a3e](https://github.com/ory/kratos/commit/e169a3e4079b1ef3a18564e0723baf81c44c38ec)), closes [#37](https://github.com/ory/kratos/issues/37):

    This patch implements the account recovery with endpoints such as "Init Account Recovery", a new config value `urls.recovery_ui` and so on. A new identity field has been added `identity.recovery_addresses` containing all recovery addresses.
    
    Additionally, some refactoring was made to DRY code and make naming consistent. As part of dependency upgrades, structured logging has also improved and an audit trail prototype has been added (currently streams to stderr only).


### Unclassified

* docs:fixed broken link (#454) ([22720c6](https://github.com/ory/kratos/commit/22720c6c5e3d31acc175980223183e2336b3751d)), closes [#454](https://github.com/ory/kratos/issues/454)
* Allow kratos to talk to databases in docker-compose quickstart ([#522](https://github.com/ory/kratos/issues/522)) ([8bf9a1a](https://github.com/ory/kratos/commit/8bf9a1ac4162c677a455c2f02de658bd5d146905)):

    All of the databases must exist on the same docker network to allow the
    main kratos applications to communicate with them.

* Fixed typo ([#472](https://github.com/ory/kratos/issues/472)) ([31263b6](https://github.com/ory/kratos/commit/31263b68ab8d81d264e0fa375a915f8f82d70bb3))


# [0.3.0-alpha.1](https://github.com/ory/kratos/compare/v0.2.1-alpha.1...v0.3.0-alpha.1) (2020-05-15)

This release finalizes the OpenID Connect and OAuth2 login, registration, and settings strategy with JsonNet data transformation! From now on, "Sign in with Google, Github, ..." is officially supported! It's also possible to link and unlink these connections using the Self-Service Settings Flow! The documentation has been updated to reflect those changes and includes guides to setting up "Sign in with GitHub" in under 5 Minutes! Please be aware that existing OpenID Connect connections will stop working. Check out the "Breaking Changes" section for more info! Want to learn more? Check [out the docs](https://www.ory.sh/kratos/docs/concepts/credentials/openid-connect-oidc-oauth2)!

We also changed the config validation output, making it easier than ever to find bugs in your config:

```
% kratos --config invalid-config.yml serve
INFO[0001] Config file loaded successfully.              path=invalid-config.yml
ERRO[0001] The provided configuration is invalid and could not be loaded. Check the output below to understand why.  config_file=invalid-config.yml

dsn: <nil>
     ^-- one or more required properties are missing

urls.whitelisted_return_to_urls: https://selfservice.office.example.com
                                 ^-- expected array, but got string

FATA[0001] The services failed to start because the configuration is invalid. Check the output above for more details.
```

This release concludes over 50 commits and 16.000 lines of code changed.



## Breaking Changes

If you upgrade and have existing Social Sign In connections, it will no longer be possible to use them to sign in. Because the oidc strategy was undocumented and not officially released we do not provide an upgrade guide. If you run into this issue on a production system you may need to use SQL to change the config of those identities. If this is a real issue for you that you're unable to solve, please create an issue on GitHub.

This is a breaking change as previous OIDC configurations will not work. Please consult the newly written documentation on OpenID Connect to learn how to use OIDC in your login and registration flows. Since the OIDC feature was not publicly broadcasted yet we have chosen not to provide an upgrade path. If you have issues, please reach out on the forums or slack.



### Bug Fixes

* Access rules of oathkeeper for quick start ([#390](https://github.com/ory/kratos/issues/390)) ([5ed6d05](https://github.com/ory/kratos/commit/5ed6d05b3e13027e4e7ffef1ff10ab2fb948093d)), closes [#389](https://github.com/ory/kratos/issues/389):

    To access `/` as dashboard

* Active field should not be required ([#401](https://github.com/ory/kratos/issues/401)) ([aed2a5c](https://github.com/ory/kratos/commit/aed2a5c3c8e39132df53ae8f0eecfb7924296796)), closes [ory/sdk#14](https://github.com/ory/sdk/issues/14)
* Adopt jsonnet in e2e oidc tests ([5e518fb](https://github.com/ory/kratos/commit/5e518fb2de678e27fcc0e4fff020a4d575f1c109))
* Detect postgres unique constraint ([3a777af](https://github.com/ory/kratos/commit/3a777af00244066a42751005d832e4058ddad8d2))
* Fix oidc strategy jsonnet test ([f6c48bf](https://github.com/ory/kratos/commit/f6c48bf2c64cea1f111e5777de22878e0be5f03c))
* Improve config validation error message ([#414](https://github.com/ory/kratos/issues/414)) ([d1e6896](https://github.com/ory/kratos/commit/d1e6896b3870cad49217ee78f6024a8a5c416f46)), closes [#413](https://github.com/ory/kratos/issues/413)
* Reset request id after parse ([9550205](https://github.com/ory/kratos/commit/9550205a35364473e0f620ef2b2a7eac223dbfff))
* Resolve flaky swagger generation ([#416](https://github.com/ory/kratos/issues/416)) ([ac4acfc](https://github.com/ory/kratos/commit/ac4acfcd7f4e686b5d5c01136158fdf1687329ac))
* Resolve regression issues and bugs ([e6d5369](https://github.com/ory/kratos/commit/e6d53693e146ec6e0d9de2ea366323721af3d8fb))
* Return correct error on id mismatch ([5915f28](https://github.com/ory/kratos/commit/5915f2882d2a481ea357d50b0058093ba3ddb51b))
* Test and implement mapper_url for jsonnet ([40ac3dc](https://github.com/ory/kratos/commit/40ac3dc7b5828ac775055fed3c0bd9ff393e5d86))
* Transaction usage in the identity persister ([#404](https://github.com/ory/kratos/issues/404)) ([7f5072d](https://github.com/ory/kratos/commit/7f5072dc2d4fbf1f48cdf4d199ce4e89683a87b1))

### Chores

* Pin v0.3.0-alpha.1 release commit ([43b693a](https://github.com/ory/kratos/commit/43b693a449bf7cd219eb6901acf36725ace1c41c))

### Code Refactoring

* Adopt new request parser ([ad16cc9](https://github.com/ory/kratos/commit/ad16cc917c8067eb1c4b89ef8192287be1c912c8))
* Dry config and oidc tests ([3e98756](https://github.com/ory/kratos/commit/3e9875612ea895f9b565d34f4d5b0f80d136868f))
* Improve oidc flows and payloads and add e2e tests ([#381](https://github.com/ory/kratos/issues/381)) ([f9a5079](https://github.com/ory/kratos/commit/f9a50790637a848897ba275373bc538728e09f3d)), closes [#387](https://github.com/ory/kratos/issues/387):

    This patch improves the OpenID Connect login and registration user experience by simplifying the network flows and introduces e2e tests using ORY Hydra.

* Move cypress files to test/e2e ([df8e627](https://github.com/ory/kratos/commit/df8e627d81d69682e01ec5670c7088ba564df578))
* Moved scanner json to ory/x ([#412](https://github.com/ory/kratos/issues/412)) ([8a0967d](https://github.com/ory/kratos/commit/8a0967daef4329981b01e6c2b8bb55a8105b4829))
* Partition files and change creds structure ([4f1eb94](https://github.com/ory/kratos/commit/4f1eb946fe1e74e537fc2166fc000180a11c2048)):

    This patch changes the data model of the OpenID Connect strategy. Instead of using an array of providers as the base config item (e.g. `{"type":"oidc","config":[{"provider":"google","subject":"..."}]}`) the credentials config is now an object with a `providers` key: `{"type":"oidc","config":{"providers":[{"provider":"google","subject":"..."}]}}`. This change allows introduction of future changes to the schema without breaking compatibility.

* Replace oidc jsonschema with jsonnet ([2b45e79](https://github.com/ory/kratos/commit/2b45e7953787ad46a6937fe44cb24b6c786eb223)), closes [#380](https://github.com/ory/kratos/issues/380):

    This patch replaces the previous methodology of merging OIDC data which used JSON Schema with Extensions and JSON Path in favor of a much easier to use approach with JSONNet.

* **settings:** Use common request parser ([ad6c402](https://github.com/ory/kratos/commit/ad6c4026e5fd15924dc906cdc9cb6c9de2fc4daa))

### Documentation

* Document account enumeration defenses for oidc ([266329c](https://github.com/ory/kratos/commit/266329cd2969627c823418c1267360193e6342df)), closes [#32](https://github.com/ory/kratos/issues/32)
* Document new oidc jsonnet mapper ([#392](https://github.com/ory/kratos/issues/392)) ([088b30f](https://github.com/ory/kratos/commit/088b30feb6845863e6651489e0c963cde7e10516))
* Document oidc strategy ([#415](https://github.com/ory/kratos/issues/415)) ([9f079f4](https://github.com/ory/kratos/commit/9f079f4f77e54f7be67ac59e13e8ec2696522637)), closes [#409](https://github.com/ory/kratos/issues/409) [#124](https://github.com/ory/kratos/issues/124) [#32](https://github.com/ory/kratos/issues/32)
* Explain that form data is merged with oidc data ([#394](https://github.com/ory/kratos/issues/394)) ([b0dbec4](https://github.com/ory/kratos/commit/b0dbec403c96af41346b6b14fc74b7010e7f8e8a)), closes [#127](https://github.com/ory/kratos/issues/127)
* Fix links in README ([efb6102](https://github.com/ory/kratos/commit/efb610239ac2ae828db26ee84c4c5a83c54c0a6a)), closes [#403](https://github.com/ory/kratos/issues/403)
* Improve social sign in guide ([#393](https://github.com/ory/kratos/issues/393)) ([647ced3](https://github.com/ory/kratos/commit/647ced3084d203e9954ca037afea34316f2080d8)), closes [#49](https://github.com/ory/kratos/issues/49):

    This patch changes the social sign in guide to represent more use cases such as Google and Facebook. Additionally, the example has been updated to work with Jsonnet.
    
    This patch also documents limitations around merging user data from GitHub.

* Improve the identity data model page ([#410](https://github.com/ory/kratos/issues/410)) ([2915b8f](https://github.com/ory/kratos/commit/2915b8faf3530fe7b9d252094c3aeb9fdbe9dd08))
* Include redirect doc in nav ([5aaebff](https://github.com/ory/kratos/commit/5aaebffd8c03e613ec60735536b6ef38d4da39e3)), closes [#406](https://github.com/ory/kratos/issues/406)
* Prepare v0.3.0-alpha.1 ([d6a6f43](https://github.com/ory/kratos/commit/d6a6f432f375018a2dc79d6b60de18455057c25a))
* Ui should show only active form sections ([#395](https://github.com/ory/kratos/issues/395)) ([4db674d](https://github.com/ory/kratos/commit/4db674de14bc50e782321c7bd88ac8077db2bf75))
* Update github templates ([#408](https://github.com/ory/kratos/issues/408)) ([6e646b0](https://github.com/ory/kratos/commit/6e646b033e0d43499bf37579a2f04b726af0e3f7))

### Features

* Add format and lint for JSONNet files ([0a1b244](https://github.com/ory/kratos/commit/0a1b244a6fd2f714a12d101071b3c0f82b4da584)):

    This patch adds two commands `kratos jsonnet format` and `kratos jsonnet lint` that help with formatting and linting JSONNet code.

* Implement oidc settings e2e tests ([919925c](https://github.com/ory/kratos/commit/919925c87be561064300c3981b5a230c6cada4f7))
* Introduce leaklog for debugging oidc map payloads ([238d7a4](https://github.com/ory/kratos/commit/238d7a493566bcc28f08b1b2bf6463f95b100254))
* Write tests and fix bugs for oidc settings ([575a61f](https://github.com/ory/kratos/commit/575a61f58a887fefa6b2917761c06304c94c9892))

### Unclassified

* Format code ([bc7557a](https://github.com/ory/kratos/commit/bc7557a4247ede1fdb4141f2670532aec7cbd456))


# [0.2.1-alpha.1](https://github.com/ory/kratos/compare/v0.2.0-alpha.2...v0.2.1-alpha.1) (2020-05-05)

Resolves a bug in the kratos-selfservice-ui-node application.





### Chores

* Pin v0.2.1-alpha.1 release commit ([16463ea](https://github.com/ory/kratos/commit/16463ead91a009f33373150d10095aa3857b38f4))

### Documentation

* Fix quickstart hero sections ([7c6c439](https://github.com/ory/kratos/commit/7c6c4397bccd2b505fc04cc8d3b0944ceca18982))
* Fix typo in upgrade guide ([a1b1d7c](https://github.com/ory/kratos/commit/a1b1d7c9cbe5fad3b1112a16eced4f3064cfdda0))


# [0.2.0-alpha.2](https://github.com/ory/kratos/compare/v0.1.1-alpha.1...v0.2.0-alpha.2) (2020-05-04)

This is a heavy release with over hundreds of commits and files changed! Let's
take a look at some of the highlights!

**ORY Oathkeeper now optional**

Using ORY Oathkeeper to protect your API is now optional. The basic quickstart
now uses a much simpler set up. Go
[check it out](https://www.ory.sh/kratos/docs/quickstart) now!

**PostgreSQL, MySQL, CockroachDB support now tested and official!**

All three databases now pass acceptance tests and are thus officially supported!

**Self-Service Profile Flow**

The self-service profile flow has been refactored into a more generic flow
allowing users to make modifications to their traits and credentials. Check out
the [docs to learn
more](https://www.ory.sh/kratos/docs/self-service/flows/user-settings-profile-management)
about the flow and it's features.

Please keep in mind that the flow's APIs have changed. We recommend re-reading
the docs!

**Managing Privileged Profile Fields**

Flows such as changing ones profile or primary email address should not be
possible unless the login session is fresh. This prevents your colleague or evil
friend to take over your account while you make yourself a coffee.

ORY Kratos now supports this by redirecting the user to the login screen if
changes to sensitive fields are made. The changes will only be applied after
successful reauthentication.

**Changes to Hooks**

This patch focuses on refactoring how self-service flows terminate and changes
how hooks behave and when they are executed.

Before this patch, it was not clear whether hooks run before or after an
identity is persisted. This caused problems with multiple writes on the HTTP
ResponseWriter and other bugs.

This patch removes certain hooks from after login, registration, and profile
flows. Per default, these flows now respond with an appropriate payload (
redirect for browsers, JSON for API clients) and deprecate the `redirect` hook.
This patch includes documentation which explains how these hooks work now.

Additionally, the documentation was updated. Especially the sections about hooks
have been refactored. The login and user registration docs have been updated to
reflect the latest changes as well.

BREAKING CHANGE: Please remove the `redirect` hook from both login,
registration, and settings after configuration. Please remove the `session` hook
from your login after configuration. Hooks have moved down a level and are now
configured at `selfservice.<login|registration|settings>.<after|before>.hooks`
instead of `selfservice.<login|registration|settings>.<after|before>.hooks`.
Hooks are now identified by `hook:` instead of `job:`. Please rename those
sections accordingly.

We recommend re-reading the
[Hooks Documentation](https://www.ory.sh/kratos/docs/self-service/hooks/index).

**Changing Passwords**

It's now possible to change your password using the Self-Service Settings Flow!
Lean more about this flow
[here](https://www.ory.sh/kratos/docs/self-service/flows/user-settings-profile-management)

**End-To-End Tests**

We added tons of end-to-end and integration tests to find and fix pesky bugs.



## Breaking Changes

Please remove the `redirect` hook from both login,
registration, and settings after configuration. Please remove
the `session` hook from your login after configuration. Hooks
have moved down a level and are now configured at
`selfservice.<login|registration|settings>.<after|before>.hooks`
instead of
`selfservice.<login|registration|settings>.<after|before>.hooks`.
Hooks are now identified by `hook:` instead of `job:`. Please
rename those sections accordingly.

Several profile-related URLs have and payloads been updated. Please consult the most recent documentation.

The payloads of the Profile Management Request API
that previously were set in `{ "methods": { "traits": { ... } }}` have now moved to
`{ "methods": { "profile": { ... } }}`.

This patch introduces a refactor that is needed
for the profile management API to be capable of handling (password,
oidc, ...) credential changes as well.

To implement this, the payloads of the Profile Management Request API
that previously were set in `{"form": {...} }` have now moved to
`{"methods": { "traits": { ... } }}`.

In the future, as more credential updates are handled, there will
be additional keys in the forms key
`{"methods": { "traits": { ... }, "password": { ... } }}`.



### Bug Fixes

* Allow setting new password in profile flow ([3b5fd5c](https://github.com/ory/kratos/commit/3b5fd5ca8c09b2344c0262547f2b387bda362362))
* Automatically append multiStatements parameter to mySQL URI ([#374](https://github.com/ory/kratos/issues/374)) ([39f77bb](https://github.com/ory/kratos/commit/39f77bb29637db048b15c097d869d8828b0d292b))
* **config:** Rename config key stmp to smtp ([#278](https://github.com/ory/kratos/issues/278)) ([ef95811](https://github.com/ory/kratos/commit/ef95811bb891afe3a0ef3b19514f13a56a32ea3b))
* Create pop connection without parsed connection options ([#366](https://github.com/ory/kratos/issues/366)) ([10b6481](https://github.com/ory/kratos/commit/10b6481774aaff42b70b9c6af3ed776ac8f7734c))
* Declare proper vars for setting version ([#383](https://github.com/ory/kratos/issues/383)) ([2fc7556](https://github.com/ory/kratos/commit/2fc7556b70b11e519162326ded0ba2638b6d32df))
* Decouple quickstart scenarios ([#336](https://github.com/ory/kratos/issues/336)) ([17363b3](https://github.com/ory/kratos/commit/17363b312deff8b92fc1b0d158dc70670d5938e5)), closes [#262](https://github.com/ory/kratos/issues/262):

    Creates several docker compose examples which include various
    scenarios of the quickstart.
    
    The regular quickstart guide now works without ORY Oathkeeper
    and uses the standalone mode of the example app instead.
    
    Additionally, the Makefile was improved and now automatically pulls
    required dependencies in the appropriate version.

* **docker:** Throw away build artifacts ([481ec1b](https://github.com/ory/kratos/commit/481ec1ba14480ced39516f6e0c47a40b6a44a631))
* Document Schema API and serve over admin endpoint ([#299](https://github.com/ory/kratos/issues/299)) ([4be417c](https://github.com/ory/kratos/commit/4be417c0ee18622247a15d2803f7f436cfe3c229)), closes [#287](https://github.com/ory/kratos/issues/287)
* Exempt whomai from csrf protection ([#329](https://github.com/ory/kratos/issues/329)) ([31d4065](https://github.com/ory/kratos/commit/31d4065c2b0cbd6c8d2b0031ce8f6f157ff967cf))
* Fix swagger annotation ([#331](https://github.com/ory/kratos/issues/331)) ([5c5c78f](https://github.com/ory/kratos/commit/5c5c78f404a11d5df25cb68584b826b685bf5385)):

    Closes https://github.com/ory/sdk/issues/10

* Move to ory sqa service ([#309](https://github.com/ory/kratos/issues/309)) ([7c244e0](https://github.com/ory/kratos/commit/7c244e0a28a010e56e07d061132dad7a0309ea75))
* Properly annotate error API ([a6f1300](https://github.com/ory/kratos/commit/a6f1300951010e7c862c410e93653f7c02c2e79f))
* Remove unused returnTo ([e64e5b0](https://github.com/ory/kratos/commit/e64e5b0cecceedda29a525f683cbf6070a9ef1eb))
* Resolve docker build permission issues ([f3612e8](https://github.com/ory/kratos/commit/f3612e8f82018bae17c9146d273fe7e82ceb033d))
* Resolve failing test issues ([2e968e5](https://github.com/ory/kratos/commit/2e968e52d3ae3396a3f2e212c0dab22677b4b5fd))
* Resolve linux install script archive naming ([#302](https://github.com/ory/kratos/issues/302)) ([c98b8aa](https://github.com/ory/kratos/commit/c98b8aa4cd3ab881b904e9dc4cdcb6383a8ad09b))
* Resolve NULL value for seen_at ([#259](https://github.com/ory/kratos/issues/259)) ([a7d1e86](https://github.com/ory/kratos/commit/a7d1e86844a9cdd0c58353e1f1e4340dac4260b3)), closes [#244](https://github.com/ory/kratos/issues/244):

    Previously, errorx tests were not executed which caused several bugs.

* Resolve password continuity issues ([56a44fa](https://github.com/ory/kratos/commit/56a44fa33d325eea9fddec4269e34e632310f77b))
* Revert use host volume mount for sqlite ([#272](https://github.com/ory/kratos/issues/272)) ([#285](https://github.com/ory/kratos/issues/285)) ([a7477ab](https://github.com/ory/kratos/commit/a7477ab1db0d986f96e754946607d05888de4c97)):

    This reverts commit 230ab2d83f4d187f410e267c6d68554e82514948.

* Self-service error query parameter name ([#308](https://github.com/ory/kratos/issues/308)) ([be257f5](https://github.com/ory/kratos/commit/be257f5448abaa48e25735a088757f3fd6dc6d22)):

    The query parameter for the self-service errors endpoint was named `id`
    in the API docs, whereas it is the `error` param that is used by the
    handler.

* **session:** Regenerate CSRF Token on principal change ([#290](https://github.com/ory/kratos/issues/290)) ([1527ef4](https://github.com/ory/kratos/commit/1527ef4209b937e2175b60d56efd019f17b33b04)), closes [#217](https://github.com/ory/kratos/issues/217)
* **session:** Whoami endpoint now supports all HTTP methods ([#283](https://github.com/ory/kratos/issues/283)) ([4bf645b](https://github.com/ory/kratos/commit/4bf645b66c7a128182ff55e52fdad7f53d752ce7)), closes [#270](https://github.com/ory/kratos/issues/270)
* Show log in ui only when unauthenticated or forced ([df77310](https://github.com/ory/kratos/commit/df77310ffbe7cfc90fa3bc5dad0450e79c34ebef)), closes [#323](https://github.com/ory/kratos/issues/323)
* **sql:** Rename migrations with same version ([#280](https://github.com/ory/kratos/issues/280)) ([07e46b9](https://github.com/ory/kratos/commit/07e46b9c9e57940bec904d744ffdd272d610a77b)), closes [#279](https://github.com/ory/kratos/issues/279)
* **swagger:** Move nolint,deadcode instructions to own file ([#293](https://github.com/ory/kratos/issues/293)) ([1935510](https://github.com/ory/kratos/commit/1935510ad9b0f387eb3b2e690e31c5313a06883e)):

    Closes https://github.com/ory/docs/pull/279

* Use host volume mount for sqlite ([#272](https://github.com/ory/kratos/issues/272)) ([230ab2d](https://github.com/ory/kratos/commit/230ab2d83f4d187f410e267c6d68554e82514948))
* Use resilient client for HIBP lookup ([#288](https://github.com/ory/kratos/issues/288)) ([735b435](https://github.com/ory/kratos/commit/735b43508392c6966a57907c20caa7cf9df4fc4d)), closes [#261](https://github.com/ory/kratos/issues/261)
* Use semver-regex replacer func ([d5c9a47](https://github.com/ory/kratos/commit/d5c9a47800fc2a55b96c7b9330f68b0a2db328cb))
* Use sqlite tag on make install ([2c82784](https://github.com/ory/kratos/commit/2c82784cd69e0468a72354f6898945032d826306))
* Verified_at field should not be required ([#353](https://github.com/ory/kratos/issues/353)) ([15d5e26](https://github.com/ory/kratos/commit/15d5e268d2ec397f0647d2407d86404c4ee8bfa3)):

    Closes https://github.com/ory/sdk/issues/11
    
    


### Chores

* Pin v0.2.0-alpha.2 release commit ([ab91689](https://github.com/ory/kratos/commit/ab916894b761b18c53e4ed1fd0e42d9f5aa0817c))

### Code Refactoring

* Move docs to this repository ([#317](https://github.com/ory/kratos/issues/317)) ([aa0d726](https://github.com/ory/kratos/commit/aa0d72639ecae3b0649761e6ee881a59b2f3e94e))
* Prepare profile management payloads for credentials ([44493f3](https://github.com/ory/kratos/commit/44493f3ddbb449981576ec317ac45530ca3be14d))
* Rename traits method to profile ([4f1e033](https://github.com/ory/kratos/commit/4f1e0339ecc1efbdfa3d3680ad64b7683e90e447))
* Rework hooks and self-service flow completion ([#349](https://github.com/ory/kratos/issues/349)) ([a7c7fef](https://github.com/ory/kratos/commit/a7c7fef758e843393b0dc1e60bee11b88b8c9b4a)), closes [#348](https://github.com/ory/kratos/issues/348) [#347](https://github.com/ory/kratos/issues/347) [#179](https://github.com/ory/kratos/issues/179) [#51](https://github.com/ory/kratos/issues/51) [#50](https://github.com/ory/kratos/issues/50) [#31](https://github.com/ory/kratos/issues/31):

    This patch focuses on refactoring how self-service flows terminate and
    changes how hooks behave and when they are executed.
    
    Before this patch, it was not clear whether hooks run before or
    after an identity is persisted. This caused problems with multiple
    writes on the HTTP ResponseWriter and other bugs.
    
    This patch removes certain hooks from after login, registration, and profile flows.
    Per default, these flows now respond with an appropriate payload (
    redirect for browsers, JSON for API clients) and deprecate
    the `redirect` hook. This patch includes documentation which explains
    how these hooks work now.
    
    Additionally, the documentation was updated. Especially the sections
    about hooks have been refactored. The login and user registration docs
    have been updated to reflect the latest changes as well.
    
    Also, some other minor, cosmetic, changes to the documentation have been made.


### Documentation

* Add banner kratos ([8a9dfbb](https://github.com/ory/kratos/commit/8a9dfbbd54bac14778cc84ec13326eb1ef80f5b3))
* Add csrf and cookie debug section ([#342](https://github.com/ory/kratos/issues/342)) ([cac2948](https://github.com/ory/kratos/commit/cac2948685ed2a3c3edbc8eb4696bbfb8523dfeb)), closes [#341](https://github.com/ory/kratos/issues/341)
* Add database connection documentation ([#332](https://github.com/ory/kratos/issues/332)) ([4f9e8b0](https://github.com/ory/kratos/commit/4f9e8b00bacda3612db3f48b81fabd562075470a))
* Add HA docs ([2e5c591](https://github.com/ory/kratos/commit/2e5c59158915d1ccbb90363e23f73a09c227b6f7))
* Add hook changes to upgrade guide ([55b5fe0](https://github.com/ory/kratos/commit/55b5fe00c0472f5f6f7408eee76bf9a39318db7e))
* Add info to oidc ([#382](https://github.com/ory/kratos/issues/382)) ([6eeeb5d](https://github.com/ory/kratos/commit/6eeeb5dbe98d2f31fd922d60a35d9d8f81d0b2a8))
* Add more examples to config schema ([#372](https://github.com/ory/kratos/issues/372)) ([ed2ccb9](https://github.com/ory/kratos/commit/ed2ccb935fdcfcb11999996cd582726bba096435)), closes [#345](https://github.com/ory/kratos/issues/345)
* Add quickstart notes for docker debugging ([74f082a](https://github.com/ory/kratos/commit/74f082a407ee73741453ff6a394f47790e79b667))
* Add settings docs and improve flows ([#375](https://github.com/ory/kratos/issues/375)) ([478cd9c](https://github.com/ory/kratos/commit/478cd9c5b5755030307d1f11e9bcbd4e171ee0d6)), closes [#345](https://github.com/ory/kratos/issues/345)
* **concepts:** Fix typo ([a49184c](https://github.com/ory/kratos/commit/a49184c30d9c2ccff5a2d41d3aff61b24e7d2ea9)):

    Closes https://github.com/ory/docs/pull/296

* **concepts:** Properly close code tag ([1c841c2](https://github.com/ory/kratos/commit/1c841c213bdbc79a6aa41e8450444d8d6c1f0284))
* Declare api frontmatter properly ([df7591f](https://github.com/ory/kratos/commit/df7591f7b70c94cfe62042a598eceb36b6a4f29a))
* Document 0.2.0 high-level changes ([9be1064](https://github.com/ory/kratos/commit/9be1064500dd86489b79e1abd9cbf1268b97853a))
* Document multi-tenant set up ([891594d](https://github.com/ory/kratos/commit/891594df488e42ce30a81465f10f2936d152cb55)), closes [#370](https://github.com/ory/kratos/issues/370)
* Fix broken images in quickstart ([52aa4cf](https://github.com/ory/kratos/commit/52aa4cf0b6967108fa58f58b6b151e6f6118bcc9))
* Fix broken link ([bf7843c](https://github.com/ory/kratos/commit/bf7843cd96795a894488a0910529c847cf7eee19)), closes [#327](https://github.com/ory/kratos/issues/327)
* Fix broken link ([c2adc73](https://github.com/ory/kratos/commit/c2adc734a73758d858d50d8738dc2a556110f26c)), closes [#327](https://github.com/ory/kratos/issues/327)
* Fix broken mermaid links ([f24fc1b](https://github.com/ory/kratos/commit/f24fc1bbba234d71098298bcddbba236ac4297f3))
* Fix spelling in quickstart ([#356](https://github.com/ory/kratos/issues/356)) ([3ce6b4a](https://github.com/ory/kratos/commit/3ce6b4a1b0722a96bcbae79b7261616f20741494))
* Improve changelog ([#384](https://github.com/ory/kratos/issues/384)) ([a973ca7](https://github.com/ory/kratos/commit/a973ca7719cd820bb196ec5732c85418528be1d0))
* Improve profile section and restructure nav ([#373](https://github.com/ory/kratos/issues/373)) ([3cc0979](https://github.com/ory/kratos/commit/3cc097934edc81d4c6d853594eed5e68e9e48445)), closes [#345](https://github.com/ory/kratos/issues/345)
* Regenerate and update changelog ([7d4ed98](https://github.com/ory/kratos/commit/7d4ed9873f25b14b59f727002fb08a8b8a4e91a6))
* Regenerate and update changelog ([175b626](https://github.com/ory/kratos/commit/175b626f74b4471e068bd79259c6d479fd6c1a7d))
* Regenerate and update changelog ([e60e2df](https://github.com/ory/kratos/commit/e60e2df5d5cc4c1ef8a6a7f13487d4ebbf54741e))
* Regenerate and update changelog ([41eeb75](https://github.com/ory/kratos/commit/41eeb7587fad864f64c4179ac20847f902c438b3))
* Regenerate and update changelog ([468105a](https://github.com/ory/kratos/commit/468105a6080b861f1e02db3a404f2bac7f2f5eb6))
* Regenerate and update changelog ([8414520](https://github.com/ory/kratos/commit/8414520c995cb2405ed051952357d37ca8111f25))
* Regenerate and update changelog ([85d5866](https://github.com/ory/kratos/commit/85d5866df403b3cfa5566cef5cb983714b395505))
* Regenerate and update changelog ([e8d2d10](https://github.com/ory/kratos/commit/e8d2d1019bbc05fbe4eeaaee7a8eb1e8f2d18cf9))
* Regenerate and update changelog ([4c58b6d](https://github.com/ory/kratos/commit/4c58b6de4a3a39b1e94516abd1ea8ed7b09c1fe4))
* Regenerate and update changelog ([a726eb2](https://github.com/ory/kratos/commit/a726eb202a070038148612f98f12e5d22170d1ec))
* Regenerate and update changelog ([87b47ba](https://github.com/ory/kratos/commit/87b47baa9cdc0175c58ccbb20e67b458ce6a445f))
* Regenerate and update changelog ([537d496](https://github.com/ory/kratos/commit/537d496d2043a17c68f31a8744c39bc76f76314c))
* Regenerate and update changelog ([00e6af9](https://github.com/ory/kratos/commit/00e6af96060ec38059c449ac5e8b3c1df5bb8c95))
* Regenerate and update changelog ([48a2eca](https://github.com/ory/kratos/commit/48a2eca2dcd274ca73d55132efca4a6dae63efdf))
* Regenerate and update changelog ([8a71948](https://github.com/ory/kratos/commit/8a719481b54957681aa21eff5415229f3e5d4bff))
* Regenerate and update changelog ([ad3d510](https://github.com/ory/kratos/commit/ad3d5101dad3c8a2725083c63f155638905b6e8c))
* Regenerate and update changelog ([48bcc70](https://github.com/ory/kratos/commit/48bcc704ed22d8c78620aa3a5f8ecb5b41937759))
* Regenerate and update changelog ([816a55c](https://github.com/ory/kratos/commit/816a55c81a27b53d5bd823392751853b68d3f607))
* Regenerate and update changelog ([4ed74d2](https://github.com/ory/kratos/commit/4ed74d25c45f6e439377329d42cd7ae0acf9d0f1))
* Regenerate and update changelog ([367927e](https://github.com/ory/kratos/commit/367927e716e7c1c6898151a5f14876fb30070dd3))
* Regenerate and update changelog ([38f4019](https://github.com/ory/kratos/commit/38f40190f54264808c7a2716555876d05cdf560f))
* Typo in README.md ([#265](https://github.com/ory/kratos/issues/265)) ([9f865a2](https://github.com/ory/kratos/commit/9f865a2ebace801414b2de17fe2f627d91f23474))
* Update banner url ([292c986](https://github.com/ory/kratos/commit/292c986729d83187f7e77365e11ef74a6f3cadf6))
* Update forum and chat links ([3039191](https://github.com/ory/kratos/commit/30391919d7ea58609dd3cd37db2709495e7abc76))
* Update github templates ([#338](https://github.com/ory/kratos/issues/338)) ([57dbc77](https://github.com/ory/kratos/commit/57dbc77b548383522ca428e899dfde461334216c))
* Update github templates ([#343](https://github.com/ory/kratos/issues/343)) ([eb13dc1](https://github.com/ory/kratos/commit/eb13dc1285cb16515d1c63b99cc389147508a31e))
* Update github templates ([#350](https://github.com/ory/kratos/issues/350)) ([faf2f30](https://github.com/ory/kratos/commit/faf2f305aea1826e3d5f0b2614313920ac2b585b))
* Update github templates ([#351](https://github.com/ory/kratos/issues/351)) ([20ff289](https://github.com/ory/kratos/commit/20ff2890004745231073cd4fd6ef1b37521cde72))
* Update linux install guide ([3b8e549](https://github.com/ory/kratos/commit/3b8e5493a01357f8c442a8a2dc9437712498452c))
* Update linux install guide ([#354](https://github.com/ory/kratos/issues/354)) ([ec49cae](https://github.com/ory/kratos/commit/ec49caec6ddea2c800db0779005bac6da73903e1))
* Update self service reg docs ([#367](https://github.com/ory/kratos/issues/367)) ([4cf0323](https://github.com/ory/kratos/commit/4cf0323095990c5ec25283a01561cb9b8833f9ef)):

    The old links pointed at `/auth/browser/(login|registration)`
    which seems to be outdated now.
    
    From the ui node code: https://github.com/ory/kratos-selfservice-ui-node/blob/489c76d1b0474ee55ef56804b28f54d8718747ba/src/routes/auth.ts#L28
    and the api documentation for kratos https://www.ory.sh/kratos/docs/reference/api#get-the-request-context-of-browser-based-login-user-flows,
    these seem to be incorrect.
    
    The actual url hit is `/self-service/browser/flows/requests/(login|registration)`.
    This commit updates those links
    
    This blob was previously one large inline string, which personally made
    the docs a bit hard to read. This formats it into an (arguably) easier
    to parse code block

* Update user-settings-profile-management.md ([#322](https://github.com/ory/kratos/issues/322)) ([45dc3a5](https://github.com/ory/kratos/commit/45dc3a56c15ae442890313a7dbc784b75644248a))
* Updates issue and pull request templates ([#298](https://github.com/ory/kratos/issues/298)) ([1be738d](https://github.com/ory/kratos/commit/1be738d3f8e9bbc6dae31ffad5d990657a66761c))
* Updates issue and pull request templates ([#313](https://github.com/ory/kratos/issues/313)) ([299063c](https://github.com/ory/kratos/commit/299063caf2fdde40713bae4c36abb3b6fac7271d))
* Updates issue and pull request templates ([#314](https://github.com/ory/kratos/issues/314)) ([d5ae452](https://github.com/ory/kratos/commit/d5ae452a8ce5f641a40e510e82441d4eb8137218))
* Updates issue and pull request templates ([#315](https://github.com/ory/kratos/issues/315)) ([8b68db1](https://github.com/ory/kratos/commit/8b68db140a7fc1c0eaa9318c1759ea9d8d0c27df))
* Use git checkout <tag> in quickstart ([#339](https://github.com/ory/kratos/issues/339)) ([2d2562b](https://github.com/ory/kratos/commit/2d2562b587a69a2891ff29d927cb001e15d75b5d)), closes [#335](https://github.com/ory/kratos/issues/335)

### Features

* Add `dsn: memory` shorthand ([#284](https://github.com/ory/kratos/issues/284)) ([e66a030](https://github.com/ory/kratos/commit/e66a030f7d67dec639121fb23dfc7f1444474c6b)), closes [#228](https://github.com/ory/kratos/issues/228)
* Add and test id hint in reauth flow ([2298f01](https://github.com/ory/kratos/commit/2298f0140e77da870c842daa8eaca274e5d64254)), closes [#323](https://github.com/ory/kratos/issues/323)
* Add cypress e2e tests ([#334](https://github.com/ory/kratos/issues/334)) ([abc0e91](https://github.com/ory/kratos/commit/abc0e91e278f7938b264598ac0c60d18c5a9e8a0))
* Allow configuring same-site for session cookies ([#303](https://github.com/ory/kratos/issues/303)) ([2eb2054](https://github.com/ory/kratos/commit/2eb2054a94281aefa9a0818110d168cc9c052094)), closes [#257](https://github.com/ory/kratos/issues/257):

    It is now possible to set SameSite for the session cookie via the key `security.session.cookie.same_site`.

* **continuity:** Implement request continuity ([135e047](https://github.com/ory/kratos/commit/135e04750b1855ab0db812517c61e292a770ba94)), closes [#304](https://github.com/ory/kratos/issues/304) [#311](https://github.com/ory/kratos/issues/311):

    This patch adds a module which is capable of aborting a request, waiting for
    another option to complete, and then resuming the request again.
    
    This feature makes use of a temporary cookie which keeps track of the
    request state.
    
    This feature is required for several workflows that update privileged
    fields such as passwords, 2fa recovery codes, email addresses.
    
    refactor: rename profile to settings flow
    
    Renames selfservice/profile to settings. The settings flow includes a strategy for managing profile information

* Enable CockroachDB integration ([#260](https://github.com/ory/kratos/issues/260)) ([adc5153](https://github.com/ory/kratos/commit/adc5153410fb4d9f99702d7c73a78aeec8c1e9f1)), closes [#132](https://github.com/ory/kratos/issues/132) [#155](https://github.com/ory/kratos/issues/155)
* Enable continuity management for settings module ([009d755](https://github.com/ory/kratos/commit/009d7558f525168fecf86168de2906088662535e))
* Enable updating auth related traits ([#266](https://github.com/ory/kratos/issues/266)) ([65b88ba](https://github.com/ory/kratos/commit/65b88ba52fb9e6da3c1a65f734352519303327a6)), closes [#243](https://github.com/ory/kratos/issues/243)
* Implement password profile management flow ([a31839a](https://github.com/ory/kratos/commit/a31839a5c33c80500c900fb50d1dd499ab1161a1)), closes [#243](https://github.com/ory/kratos/issues/243)
* Introduce fallbacks for required configs ([#376](https://github.com/ory/kratos/issues/376)) ([b3bcb25](https://github.com/ory/kratos/commit/b3bcb25be6b417647ece2b3dda26d691f8e8d685)), closes [#369](https://github.com/ory/kratos/issues/369) [#352](https://github.com/ory/kratos/issues/352)
* **login:** Forced reauthentication ([#248](https://github.com/ory/kratos/issues/248)) ([344fc9c](https://github.com/ory/kratos/commit/344fc9cddccff958f13249b999a835d3e46a7771)), closes [#243](https://github.com/ory/kratos/issues/243)
* Return 410 when selfservice requests expire ([#289](https://github.com/ory/kratos/issues/289)) ([b414607](https://github.com/ory/kratos/commit/b4146076148d9ff079e9d433f0a90f5bc938650c)), closes [#235](https://github.com/ory/kratos/issues/235)
* Send verification emails on profile update ([#333](https://github.com/ory/kratos/issues/333)) ([1cacc80](https://github.com/ory/kratos/commit/1cacc80c54f92b380ef3752591970cc4dd97085e)), closes [#267](https://github.com/ory/kratos/issues/267)

### Unclassified

* u ([0b6fa48](https://github.com/ory/kratos/commit/0b6fa48e90fa0c50b9c26bae034eb1662c855d69))
* u ([03fa4f0](https://github.com/ory/kratos/commit/03fa4f05363aa1f38fe45730317375ce380cfa31))
* u ([a3dfd9d](https://github.com/ory/kratos/commit/a3dfd9d15e1f7287558b85c3a4f23d02444b0bf4))
* u ([616aa0f](https://github.com/ory/kratos/commit/616aa0f0cf3d662b48fcaa02715e02e854e05581))
* fix:add graceful shutdown to courier handler (#296) ([235d784](https://github.com/ory/kratos/commit/235d784b7f8bf38859d15d68c37b089fc9371195)), closes [#296](https://github.com/ory/kratos/issues/296) [#295](https://github.com/ory/kratos/issues/295):

    Courier would not stop with the provided Background handler.
    This changes the methods of Courier so that the graceful package can be
    used in the same way as the http endpoints can be used.

* fix(sql) change courier body to text field (#276) ([ed5268d](https://github.com/ory/kratos/commit/ed5268d539b2a28f5367e8ba2e2e6bd3a605ce5b)), closes [#276](https://github.com/ory/kratos/issues/276) [#269](https://github.com/ory/kratos/issues/269)
* Make format ([b85e5af](https://github.com/ory/kratos/commit/b85e5af2e29f9ca3bc3341ba4f2b1b338b441398))


# [0.1.1-alpha.1](https://github.com/ory/kratos/compare/v0.1.0-alpha.6...v0.1.1-alpha.1) (2020-02-18)

docs: Regenerate and update changelog





### Bug Fixes

* Add verify return to address ([#252](https://github.com/ory/kratos/issues/252)) ([64ab9e5](https://github.com/ory/kratos/commit/64ab9e510e6b65f9dd16fdfaadfd24785dab0c93))
* Clean up docker quickstart ([#255](https://github.com/ory/kratos/issues/255)) ([7f0996b](https://github.com/ory/kratos/commit/7f0996b99646e57136f20c04a77a6f682eecdd9c))
* Resolve several verification problems ([#253](https://github.com/ory/kratos/issues/253)) ([30d4632](https://github.com/ory/kratos/commit/30d46326373cf038b600ee07db3e95ce6d94ab12))
* Update verify URLs ([#258](https://github.com/ory/kratos/issues/258)) ([5d4f909](https://github.com/ory/kratos/commit/5d4f9099b5c61ff9572ad23a3eb9c0e0025d92da))

### Code Refactoring

* Support context-based SQL transactions ([#254](https://github.com/ory/kratos/issues/254)) ([6ace1ee](https://github.com/ory/kratos/commit/6ace1ee2070c35b0da3e36dcd5417ff70a4ff9cb))

### Documentation

* Regenerate and update changelog ([a125822](https://github.com/ory/kratos/commit/a1258221a1fef82cc525be7b1042e91e2d20b1eb))
* Regenerate and update changelog ([b3a8220](https://github.com/ory/kratos/commit/b3a822035509ec2c9fb04037b2088ce6df8191da))
* Regenerate and update changelog ([a141b30](https://github.com/ory/kratos/commit/a141b309a1fc22bc45d70a090869fdee198a065e))
* Regenerate and update changelog ([7e12e20](https://github.com/ory/kratos/commit/7e12e20be0fa61a2f41a416a3edcd2b522165196))
* Regenerate and update changelog ([3c1c67b](https://github.com/ory/kratos/commit/3c1c67b31a54dd8d5fceac9449d305db82ff8844))
* Regenerate and update changelog ([ee07937](https://github.com/ory/kratos/commit/ee07937d5e797f0217c86946da42d0070ca7c250))


# [0.1.0-alpha.6](https://github.com/ory/kratos/compare/v0.1.0-alpha.5...v0.1.0-alpha.6) (2020-02-16)

feat: Add verification to quickstart (#251)






### Bug Fixes

* Adapt quickstart to verify changes ([#247](https://github.com/ory/kratos/issues/247)) ([24eceb7](https://github.com/ory/kratos/commit/24eceb7147cef1081ac1ad969713ca1bc36229cb))
* Gracefully handle selfservice request expiry ([#242](https://github.com/ory/kratos/issues/242)) ([4421e6b](https://github.com/ory/kratos/commit/4421e6bde494fbe9672251cf813a39e3031bf3fd)), closes [#233](https://github.com/ory/kratos/issues/233)
* Set AuthenticatedAt in session issuer hook ([#246](https://github.com/ory/kratos/issues/246)) ([29c83fa](https://github.com/ory/kratos/commit/29c83fa986c612fb17e13fe9415f7836062159d2)), closes [#224](https://github.com/ory/kratos/issues/224)
* **swagger:** Sanitize before validate ([c72f140](https://github.com/ory/kratos/commit/c72f140083e94f3a47ee2398c56d188e6d4edcb4))
* **swagger:** Use correct annotations for request methods ([#237](https://github.com/ory/kratos/issues/237)) ([8473c85](https://github.com/ory/kratos/commit/8473c85d8282b27375b53babbbc79046d407b3fb)), closes [#234](https://github.com/ory/kratos/issues/234)

### Code Refactoring

* Move to ory/jsonschema/v3 everywhere ([#229](https://github.com/ory/kratos/issues/229)) ([61f5c1d](https://github.com/ory/kratos/commit/61f5c1d3d896841b08deb08c42ba896118e3fc71)), closes [#225](https://github.com/ory/kratos/issues/225)

### Documentation

* Regenerate and update changelog ([922cf0f](https://github.com/ory/kratos/commit/922cf0f3d7ec8860d13aff3b88849a71fb59e2c9))
* Regenerate and update changelog ([e097c23](https://github.com/ory/kratos/commit/e097c23d8b4902a9013f3a8fa9a397033a92fb88))
* Regenerate and update changelog ([2d1685f](https://github.com/ory/kratos/commit/2d1685f4f4235e9293b1ab79e67050042787c6e9))
* Regenerate and update changelog ([f8964e9](https://github.com/ory/kratos/commit/f8964e9e5c442f75ba501ce7cfcb18916b781dc1))
* Regenerate and update changelog ([92b8001](https://github.com/ory/kratos/commit/92b80013c98e9556138eff04aa24dc696b8d6128))
* Regenerate and update changelog ([d7083ab](https://github.com/ory/kratos/commit/d7083ab9fb8e8172707cae3ac4a8a183f0c25903))
* Regenerate and update changelog ([c4547dc](https://github.com/ory/kratos/commit/c4547dc53ecf167b63e5d7d3b6764535bd86fa5a))
* Regenerate and update changelog ([d8d8bba](https://github.com/ory/kratos/commit/d8d8bbae055e2220023a45b832d2435984191029))
* Regenerate and update changelog ([b012ed9](https://github.com/ory/kratos/commit/b012ed9ce1f4fd0ece2e3463e952711b4380f4a4))

### Features

* Add disabled flag to identifier form fields ([#238](https://github.com/ory/kratos/issues/238)) ([a2178bd](https://github.com/ory/kratos/commit/a2178bdbbe20798a3e1e3fb5ed7b44afc187c640)), closes [#227](https://github.com/ory/kratos/issues/227)
* Add verification to quickstart ([#251](https://github.com/ory/kratos/issues/251)) ([172dc87](https://github.com/ory/kratos/commit/172dc87d22f925668c21da1b3b581156e01d45a4))
* Implement email verification ([#245](https://github.com/ory/kratos/issues/245)) ([eed00f4](https://github.com/ory/kratos/commit/eed00f4b328c173057455980ce0e1aad909c278f)), closes [#27](https://github.com/ory/kratos/issues/27)
* Improve password validation strategy ([#231](https://github.com/ory/kratos/issues/231)) ([256fad3](https://github.com/ory/kratos/commit/256fad37164c81cc44c35e77b99911996722a86a))


# [0.1.0-alpha.5](https://github.com/ory/kratos/compare/v0.1.0-alpha.4...v0.1.0-alpha.5) (2020-02-06)

docs: Regenerate and update changelog





### Documentation

* Regenerate and update changelog ([e87e9c9](https://github.com/ory/kratos/commit/e87e9c9ec9cf55351439ab16a778f3ea303ec646))
* Regenerate and update changelog ([d6f0794](https://github.com/ory/kratos/commit/d6f0794d53b6e7d6d9e3bc63a77d402e43a29bed))
* Regenerate and update changelog ([eb7326c](https://github.com/ory/kratos/commit/eb7326c98c2d5e87a8ac3cd9f2efb43f2552164a))

### Features

* Redirect to new auth session on expired auth sessions ([#230](https://github.com/ory/kratos/issues/230)) ([b477ecd](https://github.com/ory/kratos/commit/b477ecd47de33a9a45159a298ac288c4ad5a0b55)), closes [#96](https://github.com/ory/kratos/issues/96)


# [0.1.0-alpha.4](https://github.com/ory/kratos/compare/v0.1.0-alpha.3...v0.1.0-alpha.4) (2020-02-06)

ci: Bump ory/sdk to 0.1.22




### Continuous Integration

* Bump ory/sdk to 0.1.22 ([c0d0edf](https://github.com/ory/kratos/commit/c0d0edf1f369ecaeb28d1337930b16222b97337f))

### Documentation

* Regenerate and update changelog ([f02afb3](https://github.com/ory/kratos/commit/f02afb3fed310f7fe9c5e6f7df34dfc9738018ad))


# [0.1.0-alpha.3](https://github.com/ory/kratos/compare/v0.1.0-alpha.2...v0.1.0-alpha.3) (2020-02-06)

ci: Bump ory/sdk orb




### Continuous Integration

* Bump ory/sdk orb ([65b2ca0](https://github.com/ory/kratos/commit/65b2ca0b8a1da8249aa4b4cb439b1d63aecaf8e0))


# [0.1.0-alpha.2](https://github.com/ory/kratos/compare/v0.1.0-alpha.1...v0.1.0-alpha.2) (2020-02-03)

docs: Regenerate and update changelog





### Bug Fixes

* Add paths to sqa middleware ([#216](https://github.com/ory/kratos/issues/216)) ([130c9c2](https://github.com/ory/kratos/commit/130c9c242e1434074d9fa4970b60ccb9b4f2ff47))
* **daemon:** Register error routes on admin port ([#226](https://github.com/ory/kratos/issues/226)) ([decd8d8](https://github.com/ory/kratos/commit/decd8d8ef8dac3674938b564962238195ffaf017))
* Set csrf token on public endpoints ([d0b15ae](https://github.com/ory/kratos/commit/d0b15aeca991a94771715a6eabd4a956be41ceda))

### Documentation

* Introduce upgrade guide ([736a3b1](https://github.com/ory/kratos/commit/736a3b19bfe35cc699dea508b4bdb56b3302ba7e))
* Prepare ecosystem automation ([7013b6c](https://github.com/ory/kratos/commit/7013b6c9a856e05f6ad385eb8ce36c5faf342f5a))
* Regenerate and update changelog ([f39b942](https://github.com/ory/kratos/commit/f39b9422d79d3e69304f013c85f3850337ca1730))
* Regenerate and update changelog ([c121601](https://github.com/ory/kratos/commit/c121601b5c741c846d9c478b01aabb9907d81b95))
* Regenerate and update changelog ([a947d55](https://github.com/ory/kratos/commit/a947d554ba2be94f334568a4e77a501742ca95af))
* Regenerate and update changelog ([8ba2044](https://github.com/ory/kratos/commit/8ba2044ebb369ea741f99c65163f650c607e6c07))
* Regenerate and update changelog ([9c023e1](https://github.com/ory/kratos/commit/9c023e1a9288f156c79ea78b3a979d0fefab8825))
* Regenerate and update changelog ([1e855a9](https://github.com/ory/kratos/commit/1e855a9e0ebd232ba2b07dc4a8bb79b84cd548e6))
* Regenerate and update changelog ([01ce3a8](https://github.com/ory/kratos/commit/01ce3a891edd84174694111637dd44fe65e48b37))
* Updates issue and pull request templates ([#222](https://github.com/ory/kratos/issues/222)) ([4daae88](https://github.com/ory/kratos/commit/4daae88af527018e9ee4e1e9717a07dffab427fe))

### Features

* Override semantic config ([#220](https://github.com/ory/kratos/issues/220)) ([9b4214b](https://github.com/ory/kratos/commit/9b4214bf5eac81a92513e04dc5f862b93df86935))

### Unclassified

* Update CHANGELOG [ci skip] ([ce9390c](https://github.com/ory/kratos/commit/ce9390c27f61966b7ed23244400215c2218bbc0b))
* refactor!: Improve user-facing error APIs (#219) ([7d4054f](https://github.com/ory/kratos/commit/7d4054f4363da7bc0e943e7abfbd0c804eb7f0c1)), closes [#219](https://github.com/ory/kratos/issues/219) [#204](https://github.com/ory/kratos/issues/204):

    This patch refactors user-facing error APIs:
    
    - The `/errors` endpoint moved to `/self-service/errors`
    - The endpoint is now available at both the Admin and Public API. The Public API requires CSRF Token match or a 403 error will be returned.
    - The Public API endpoint no longer returns 404 errors but 403 instead.
    - The response payload changed. What was `[{"code": ...}]` is now `{"id": "...", "errors": [{"code": ...}]}`
    
    This patch requires running `kratos migrate sql` as a new column (`csrf_token`) has been added to the user-facing error store.

* Update CHANGELOG [ci skip] ([c368a11](https://github.com/ory/kratos/commit/c368a11523a9bcb30a830d65c11e4f6d27417a78))


# [0.1.0-alpha.1](https://github.com/ory/kratos/compare/v0.0.3-alpha.15...v0.1.0-alpha.1) (2020-01-31)

docs: Updates issue and pull request templates (#215)

Signed-off-by: aeneasr <aeneas@ory.sh>




### Documentation

* Updates issue and pull request templates ([#215](https://github.com/ory/kratos/issues/215)) ([10c45f2](https://github.com/ory/kratos/commit/10c45f23e11abba1ca82095548769cd923a6a6a6))


# [0.0.3-alpha.15](https://github.com/ory/kratos/compare/v0.0.3-alpha.14...v0.0.3-alpha.15) (2020-01-31)

Update permissions in SQLite Dockerfile





### Unclassified

* Update permissions in SQLite Dockerfile ([1266e53](https://github.com/ory/kratos/commit/1266e533ac9a1f6ec375980cadce9755998f9fe6))


# [0.0.3-alpha.14](https://github.com/ory/kratos/compare/v0.0.3-alpha.13...v0.0.3-alpha.14) (2020-01-31)

Update README.md




### Unclassified

* Update README.md ([db8d65b](https://github.com/ory/kratos/commit/db8d65bf136223df546aa27f1ecff03d01159624))


# [0.0.3-alpha.13](https://github.com/ory/kratos/compare/v0.0.3-alpha.12...v0.0.3-alpha.13) (2020-01-31)

Allow mounting SQLite in /home/ory/sqlite (#212)






### Unclassified

* Allow mounting SQLite in /home/ory/sqlite (#212) ([2fe8c0f](https://github.com/ory/kratos/commit/2fe8c0f752e870028d68e8593a46c0902f673a65)), closes [#212](https://github.com/ory/kratos/issues/212)


# [0.0.3-alpha.11](https://github.com/ory/kratos/compare/v0.0.3-alpha.10...v0.0.3-alpha.11) (2020-01-31)

Clean up cmd and resolve packr2 issues (#211)

This patch addresses issues with the build pipeline caused by an invalid import. Profiling was also added.




### Unclassified

* Clean up cmd and resolve packr2 issues (#211) ([2e43ec0](https://github.com/ory/kratos/commit/2e43ec09e9d6aa572c4351bfef4c59dfc43f2343)), closes [#211](https://github.com/ory/kratos/issues/211):

    This patch addresses issues with the build pipeline caused by an invalid import. Profiling was also added.

* Improve field types (#209) ([aeefa93](https://github.com/ory/kratos/commit/aeefa93bf0427685f6ffadad5abfaa1fc26ce074)), closes [#209](https://github.com/ory/kratos/issues/209)
* Update CHANGELOG [ci skip] ([fc32207](https://github.com/ory/kratos/commit/fc32207482861b8f989cb1d6fe5d96bf34c54e4c))


# [0.0.3-alpha.10](https://github.com/ory/kratos/compare/v0.0.3-alpha.9...v0.0.3-alpha.10) (2020-01-31)

Update README




### Unclassified

* Update README ([35a310d](https://github.com/ory/kratos/commit/35a310d6de52fa74ad8728b1df67f88ce900aa61))
* Update CHANGELOG [ci skip] ([3c98745](https://github.com/ory/kratos/commit/3c987455a44b9e12e31619ba9f447e8a5feafc38))
* Update CHANGELOG [ci skip] ([c1c01df](https://github.com/ory/kratos/commit/c1c01df3a04fc7988bf847e3f31680112f5a642d))


# [0.0.3-alpha.7](https://github.com/ory/kratos/compare/v0.0.3-alpha.5...v0.0.3-alpha.7) (2020-01-30)

Use correct project root in Dockerfile





### Unclassified

* Use correct project root in Dockerfile ([3528758](https://github.com/ory/kratos/commit/352875878c74d15b522336b518df339c8ad48e49))
* Update CHANGELOG [ci skip] ([e78bbbe](https://github.com/ory/kratos/commit/e78bbbecbd9515c02e447efc3208599bf27ef85c))


# [0.0.3-alpha.5](https://github.com/ory/kratos/compare/v0.0.3-alpha.4...v0.0.3-alpha.5) (2020-01-30)

ci: Resolve final docker build issues (#210)






### Continuous Integration

* Resolve final docker build issues ([#210](https://github.com/ory/kratos/issues/210)) ([d703a1e](https://github.com/ory/kratos/commit/d703a1e328808df6761a9da5866a3f4df4c7923e))

### Unclassified

* Update CHANGELOG [ci skip] ([ebb1744](https://github.com/ory/kratos/commit/ebb1744d68b8a416774477182b1e2b2cd8bdfc43))
* Add libmusl to binary output ([e9b8445](https://github.com/ory/kratos/commit/e9b8445f2fc8e9e571ec0b8480cc70fe3251db9e))


# [0.0.3-alpha.4](https://github.com/ory/kratos/compare/v0.0.3-alpha.3...v0.0.3-alpha.4) (2020-01-30)

Update CHANGELOG [ci skip]





### Unclassified

* Update CHANGELOG [ci skip] ([018c229](https://github.com/ory/kratos/commit/018c229c4cff62e47c1154ca29ab9c70766a43e5))
* Add and use ory docker user ([cccbe09](https://github.com/ory/kratos/commit/cccbe09cc6e2ad72847206d46afe3e0bf7f79ab5))
* Update CHANGELOG [ci skip] ([0e436e5](https://github.com/ory/kratos/commit/0e436e57f79692c4c6e0a0c25f48a41654afcda1))
* Update goreleaser changelog filters ([7e5af97](https://github.com/ory/kratos/commit/7e5af97fded9f56a3cc9d1d92a7726e7b613b586))
* Update CHANGELOG [ci skip] ([4387503](https://github.com/ory/kratos/commit/438750326c5d6ad1569802c82806e831f43e785e))


# [0.0.3-alpha.2](https://github.com/ory/kratos/compare/v0.0.3-alpha.1...v0.0.3-alpha.2) (2020-01-30)

Resolve goreleaser build issues (#208)







### Unclassified

* Resolve goreleaser build issues (#208) ([d59a08a](https://github.com/ory/kratos/commit/d59a08a0ef680a984352d7f5068626cc1958185a)), closes [#208](https://github.com/ory/kratos/issues/208)


# [0.0.3-alpha.1](https://github.com/ory/kratos/compare/v0.0.1-alpha.9...v0.0.3-alpha.1) (2020-01-30)

Update CHANGELOG [ci skip]





### Unclassified

* Update CHANGELOG [ci skip] ([49e09ea](https://github.com/ory/kratos/commit/49e09eaaab1fc681f9330e12ce6e5483c62ee9e3))
* Take form field orders from JSON Schema (#205) ([a880f0d](https://github.com/ory/kratos/commit/a880f0ddb52fb4366acf8fbd80aabaa9843445a9)), closes [#205](https://github.com/ory/kratos/issues/205) [#176](https://github.com/ory/kratos/issues/176)
* Update CHANGELOG [ci skip] ([ff52bbb](https://github.com/ory/kratos/commit/ff52bbb264542b48658679bf5563b0f3b7ad73c7))
* Adapt quickstart docker compose config (#207) ([e532583](https://github.com/ory/kratos/commit/e532583b35a22cb39bbab0101bf86c0bf01b1088)), closes [#207](https://github.com/ory/kratos/issues/207)
* Update CHANGELOG [ci skip] ([7f4800b](https://github.com/ory/kratos/commit/7f4800b07556e688ba0cd551438876b3bf23ace5))
* Update CHANGELOG [ci skip] ([1b2c3f6](https://github.com/ory/kratos/commit/1b2c3f645e64848e7fba6656aa730c7e346ed75d))
* Rework public and admin fetch strategy (#203) ([99aa169](https://github.com/ory/kratos/commit/99aa1693e758f706f264c2439594e2be37ae9bc6)), closes [#203](https://github.com/ory/kratos/issues/203) [#122](https://github.com/ory/kratos/issues/122)
* Update CHANGELOG [ci skip] ([1cea427](https://github.com/ory/kratos/commit/1cea42780a95d4ebf5520e1c1803fb13ef596d52))
* ss/profile: Use request ID as query param everywhere (#202) ([ed32b14](https://github.com/ory/kratos/commit/ed32b14f8ea972cf549480f29cbf1b95d010789c)), closes [#202](https://github.com/ory/kratos/issues/202) [#190](https://github.com/ory/kratos/issues/190)
* Update CHANGELOG [ci skip] ([a392027](https://github.com/ory/kratos/commit/a3920278129399ce576c5336c2e50dd015b8f2f8))
* Update HTTP routes for a consistent API naming (#199) ([9ed4bda](https://github.com/ory/kratos/commit/9ed4bda9f0b0d45e8ac0de0c42b78f717f3d92f3)), closes [#199](https://github.com/ory/kratos/issues/199) [#195](https://github.com/ory/kratos/issues/195)


# [0.0.1-alpha.9](https://github.com/ory/kratos/compare/v0.0.1-alpha.11...v0.0.1-alpha.9) (2020-01-29)

ci: Bump goreleaser orb




### Continuous Integration

* Bump goreleaser orb ([29cd754](https://github.com/ory/kratos/commit/29cd754d33ec2f800730bd007f17fc0ce53a51eb))


# [0.0.2-alpha.1](https://github.com/ory/kratos/compare/v0.0.1-alpha.8...v0.0.2-alpha.1) (2020-01-29)

Use correct build archive for homebrew




### Unclassified

* Use correct build archive for homebrew ([74ac29f](https://github.com/ory/kratos/commit/74ac29f43f2937cad9065ad3c03cf3cf909cff42))


# [0.0.1-alpha.6](https://github.com/ory/kratos/compare/v0.0.1-alpha.5...v0.0.1-alpha.6) (2020-01-29)

ci: Bump goreleaser orb




### Continuous Integration

* Bump goreleaser orb ([018c94c](https://github.com/ory/kratos/commit/018c94ccc9e833f28f827fd10d607a7a1c954ac5))


# [0.0.1-alpha.5](https://github.com/ory/kratos/compare/v0.0.1-alpha.3...v0.0.1-alpha.5) (2020-01-29)

ci: Bump goreleaser dependency





### Continuous Integration

* Bump goreleaser dependency ([ec49bfb](https://github.com/ory/kratos/commit/ec49bfb4b636a72e51d3a68521ba047f97d4c5e6))

### Unclassified

* Resolve build issues with CGO (#196) ([298f4ea](https://github.com/ory/kratos/commit/298f4ea85b3e7405929f481b756efe8c5c133479)), closes [#196](https://github.com/ory/kratos/issues/196)
* ss/password: Make form fields an array (#197) ([6cb0058](https://github.com/ory/kratos/commit/6cb005860755ff897ad847f09af50bc911bbc7f0)), closes [#197](https://github.com/ory/kratos/issues/197) [#186](https://github.com/ory/kratos/issues/186)


# [0.0.1-alpha.3](https://github.com/ory/kratos/compare/ab6f24a85276bdd8687f2fc06390c1279892b005...v0.0.1-alpha.3) (2020-01-28)

ci: Only compile goarmv7





### Continuous Integration

* Only compile goarmv7 ([d8e7ec7](https://github.com/ory/kratos/commit/d8e7ec788d1b43bcbbe221becde3432fdbf28e9b))

### Documentation

* Present ORY Hive to the world ([#107](https://github.com/ory/kratos/issues/107)) ([7883589](https://github.com/ory/kratos/commit/78835897664a5ab5564751fc9f04172f7d20d572))
* Updates issue and pull request templates ([0441dff](https://github.com/ory/kratos/commit/0441dffe0c439cc54214bf9ee8f4a4bd25206999))
* Updates issue and pull request templates ([#174](https://github.com/ory/kratos/issues/174)) ([ad405e9](https://github.com/ory/kratos/commit/ad405e9037e2db2910a012f414556fea672e732a))
* Updates issue and pull request templates ([#39](https://github.com/ory/kratos/issues/39)) ([daf5aa8](https://github.com/ory/kratos/commit/daf5aa89c717de6176ee25119d2e751ae2ef6558))
* Updates issue and pull request templates ([#40](https://github.com/ory/kratos/issues/40)) ([f5907f3](https://github.com/ory/kratos/commit/f5907f3f248e05511b19ff6dc15bf6f60f8b62da))
* Updates issue and pull request templates ([#59](https://github.com/ory/kratos/issues/59)) ([8c5612c](https://github.com/ory/kratos/commit/8c5612c080e5b7531028b778b86cc4cde2abd516))
* Updates issue and pull request templates ([#7](https://github.com/ory/kratos/issues/7)) ([a1220ba](https://github.com/ory/kratos/commit/a1220ba1e950498a6e9594266dc730c9a8731b49))
* Updates issue and pull request templates ([#8](https://github.com/ory/kratos/issues/8)) ([c56798a](https://github.com/ory/kratos/commit/c56798ab29e72ed308fff840e3b1b98ead19aea6))

### Unclassified

* Remove redundant return statement ([7c2989f](https://github.com/ory/kratos/commit/7c2989f52c090bb9900380b4ec74e04d9c37a441))
* ss/oidc: Remove obsolete request field from form (#193) ([59671ba](https://github.com/ory/kratos/commit/59671badb63009e2440b14868b622adc75cf882f)), closes [#193](https://github.com/ory/kratos/issues/193) [#180](https://github.com/ory/kratos/issues/180)
* strategy/oidc: Allow multiple OIDC Connections (#191) ([8984831](https://github.com/ory/kratos/commit/898483137ff9dc47d65750cd94a973f2e5bee770)), closes [#191](https://github.com/ory/kratos/issues/191) [#114](https://github.com/ory/kratos/issues/114)
* Improve Docker Compose Quickstart (#187) ([9459072](https://github.com/ory/kratos/commit/945907297ded4b18e1bd0e7c9824a975ac7395c6)), closes [#187](https://github.com/ory/kratos/issues/187) [#188](https://github.com/ory/kratos/issues/188)
* selfservice/password: Remove request field and ensure method is set (#183) ([e035adc](https://github.com/ory/kratos/commit/e035adc233198e9b5c9a6e08d442fb5fb3290816)), closes [#183](https://github.com/ory/kratos/issues/183)
* Add tests and fixtures for the config JSON Schema (#171) ([ede9c0e](https://github.com/ory/kratos/commit/ede9c0e9c45ee91e60587311dc18a0a04ff62295)), closes [#171](https://github.com/ory/kratos/issues/171)
* Add example values for config JSON Schema ([12ba728](https://github.com/ory/kratos/commit/12ba7283bf879cd7682d3017c3b3f12e49029d6b))
* Replace `url` with `uri` format in config JSON Schema ([68eddef](https://github.com/ory/kratos/commit/68eddef0cf179bf61abb999d84d2af19c3703c80))
* Replace number with integer in config JSON Schema (#177) ([9eff6fd](https://github.com/ory/kratos/commit/9eff6fd09720b11acae089ebfcaf37288bc031b0)), closes [#177](https://github.com/ory/kratos/issues/177)
* Improve `--dev` flag (#167) ([9b61ee1](https://github.com/ory/kratos/commit/9b61ee10bbb4710d6694addfa60c04313855516f)), closes [#167](https://github.com/ory/kratos/issues/167) [#162](https://github.com/ory/kratos/issues/162)
* Add goreleaser orb task (#170) ([5df0def](https://github.com/ory/kratos/commit/5df0defefc95ced289a9c59a4f5deb3c67446e75)), closes [#170](https://github.com/ory/kratos/issues/170)
* Add changelog generation task (#169) ([edd937c](https://github.com/ory/kratos/commit/edd937c21b7e37b2f2e926f0fe62c2e7d4a7d608)), closes [#169](https://github.com/ory/kratos/issues/169)
* Adopt new SDK pipeline (#168) ([21d9b6d](https://github.com/ory/kratos/commit/21d9b6d27adbfe8504fb46ac95952e7cea239085)), closes [#168](https://github.com/ory/kratos/issues/168)
* Add docker-compose quickstart (#153) ([e096190](https://github.com/ory/kratos/commit/e096190e778f22573e30f35e85b7cf147caf851b)), closes [#153](https://github.com/ory/kratos/issues/153)
* Update README (#160) ([533775b](https://github.com/ory/kratos/commit/533775ba78a2c1758c47ed093da6acc18ab951c2)), closes [#160](https://github.com/ory/kratos/issues/160)
* Separate post register/login hooks (#150) ([f4b7812](https://github.com/ory/kratos/commit/f4b78122d9cbe4dcc05b4fd52d94a2d9f1b16eb2)), closes [#150](https://github.com/ory/kratos/issues/150) [#149](https://github.com/ory/kratos/issues/149)
* Update README badges ([4f7838e](https://github.com/ory/kratos/commit/4f7838e69181c5a10e27cde1e241779e4e724909))
* Bump go-acc and resolve test issues (#154) ([15b1b63](https://github.com/ory/kratos/commit/15b1b630c5363e0e1afbed53285b3f39098c0792)), closes [#154](https://github.com/ory/kratos/issues/154) [#152](https://github.com/ory/kratos/issues/152) [#151](https://github.com/ory/kratos/issues/151):

    Due to a bug in `go-acc`, tests would not run if `-tags sqlite` was supplied as a go tool argument to `go-acc`. This patch resolves that issue and also includes several test patches from previous community PRs and some internal test issues.

* Add ORY Kratos banner to README (#145) ([23b824f](https://github.com/ory/kratos/commit/23b824f7f99efbc23787508c03506e73a3240a2a)), closes [#145](https://github.com/ory/kratos/issues/145)
* Replace DBAL layer with gobuffalo/pop (#130) ([21d08b8](https://github.com/ory/kratos/commit/21d08b84560230d8a063a418a74efcf53c146872)), closes [#130](https://github.com/ory/kratos/issues/130):

    This is a major refactoring of the internal DBAL. After a successful proof of concept and evaluation of gobuffalo/pop, we believe this to be the best DBAL for Go at the moment. It abstracts a lot of boilerplate code away.
    
    As with all sophisticated DBALs, pop too has its quirks. There are several issues that have been discovered during testing and adoption: https://github.com/gobuffalo/pop/issues/136 https://github.com/gobuffalo/pop/issues/476 https://github.com/gobuffalo/pop/issues/473 https://github.com/gobuffalo/pop/issues/469 https://github.com/gobuffalo/pop/issues/466
    
    However, the upside of moving much of the hard database/sql plumbing into another library cleans up the code base significantly and reduces complexity.
    
    As part of this change, the "ephermal" DBAL ("in memory") will be removed and sqlite will be used instead. This further reduces complexity of the code base and code-duplication.
    
    To support sqlite, CGO is required, which means that we need to run tests with `go test -tags sqlite` on a machine that has g++ installed. This also means that we need a Docker Image with `alpine` as opposed to pure `scratch`. While this is certainly a downside, the upside of less maintenance and "free" support for SQLite, PostgreSQL, MySQL, and CockroachDB simply outweighs any downsides that come with CGO.

* Replace local deps with remote ones ([8605e45](https://github.com/ory/kratos/commit/8605e454cf538e047c5a9c3479372892d6b3f483))
* ss/profile: Improve success and error flows ([9e0015a](https://github.com/ory/kratos/commit/9e0015acec7f8d927498e48366b377e22ec768b7)), closes [#112](https://github.com/ory/kratos/issues/112):

    This patch completes the profile management flow by implementing proper error and success states and adding several data integrity tests.

* Rebrand ORY Hive to ORY Kratos (#111) ([ceda7fb](https://github.com/ory/kratos/commit/ceda7fb3472b081f0c6066aa1f282d4ec1787f7b)), closes [#111](https://github.com/ory/kratos/issues/111)
* Fix broken tests and ci linter issues  (#104) ([69760fe](https://github.com/ory/kratos/commit/69760fe9fecb2f302dd5c1821185ea990f4e411c)), closes [#104](https://github.com/ory/kratos/issues/104)
* Update to Go modules 1.13 ([1da4d75](https://github.com/ory/kratos/commit/1da4d757bc2434f97c588e395305066edce9ef0d))
* Resolve minor configuration issues and response errors (#85) ([a44913b](https://github.com/ory/kratos/commit/a44913b26b515333576def6b882861ff2c8d4aff)), closes [#85](https://github.com/ory/kratos/issues/85)
* Clean up dead files (#84) ([e0c96ef](https://github.com/ory/kratos/commit/e0c96effbee2521b12eeedc851b67fa3a1ae41c8)), closes [#84](https://github.com/ory/kratos/issues/84)
* Add health endpoints (#83) ([0e936f7](https://github.com/ory/kratos/commit/0e936f7047bb9eacae0c5107360ce752a23d8282)), closes [#83](https://github.com/ory/kratos/issues/83) [#82](https://github.com/ory/kratos/issues/82)
* Update Dockerfile and related build tools (#80) ([d20c701](https://github.com/ory/kratos/commit/d20c701433cea916d3df4863846cf09743150966)), closes [#80](https://github.com/ory/kratos/issues/80)
* Implement SQL Database adapter (#79) ([86d07c4](https://github.com/ory/kratos/commit/86d07c4a9e3b3e6607e73f4d54b4e7b9f0382e59)), closes [#79](https://github.com/ory/kratos/issues/79) [#69](https://github.com/ory/kratos/issues/69)
* Prevent duplicate signups (#76) ([4c88968](https://github.com/ory/kratos/commit/4c88968a6853396755f61db2673a0cb2201868f7)), closes [#76](https://github.com/ory/kratos/issues/76) [#46](https://github.com/ory/kratos/issues/46)
* Contributing 08 10 19 00 52 45 (#74) ([43b511f](https://github.com/ory/kratos/commit/43b511f1a43be114ac04b377434b22ec8afe465b)), closes [#74](https://github.com/ory/kratos/issues/74)
* Echo form values from oidc signup ([98b1da5](https://github.com/ory/kratos/commit/98b1da5f59d5dcde4416b74ea323af3e29fefa75)), closes [#71](https://github.com/ory/kratos/issues/71)
* Properly decode values in error handler ([5eb9088](https://github.com/ory/kratos/commit/5eb9088efb291256d65fadbd5a803369cc96bdd2)), closes [#71](https://github.com/ory/kratos/issues/71)
* Force path and domain on CSRF cookie (#70) ([a80d8b0](https://github.com/ory/kratos/commit/a80d8b0e0bb16fce530559826de29fd6b9836873)), closes [#70](https://github.com/ory/kratos/issues/70) [#68](https://github.com/ory/kratos/issues/68)
* Require no session when accessing login or sign up (#67) ([c0e0da1](https://github.com/ory/kratos/commit/c0e0da1b38ebadaa33eb5b59dc566731b3320b70)), closes [#67](https://github.com/ory/kratos/issues/67) [#63](https://github.com/ory/kratos/issues/63)
* Add tests for selfservice ErrorHandler (#62) ([4bb9e70](https://github.com/ory/kratos/commit/4bb9e7086ee57c4eb1a73fea436c7b2dec0257b7)), closes [#62](https://github.com/ory/kratos/issues/62)
* Enable Circle CI (#57) ([6fb0afd](https://github.com/ory/kratos/commit/6fb0afd30e3755026b6ffca0cc80f2fe00267681)), closes [#57](https://github.com/ory/kratos/issues/57) [#53](https://github.com/ory/kratos/issues/53)
* OIDC provider selfservice data enrichment (#56) ([936970a](https://github.com/ory/kratos/commit/936970a9abaadeab5c191ff52218bf4f65af2220)), closes [#56](https://github.com/ory/kratos/issues/56) [#23](https://github.com/ory/kratos/issues/23) [#55](https://github.com/ory/kratos/issues/55)
* Remove local jsonschema module override ([cd2a5d8](https://github.com/ory/kratos/commit/cd2a5d8c74b21b122f5d5437702d8c74fb1cb726))
* Implement identity management, login, and registration (#22) ([bf3395e](https://github.com/ory/kratos/commit/bf3395ea34ecf85303034f3e941a049c8cbd6229)), closes [#22](https://github.com/ory/kratos/issues/22)
* Revert incorrect license changes ([fb9740b](https://github.com/ory/kratos/commit/fb9740b37a94dbdde1a8f4433fb7e5a8b4dac295))
* Create FUNDING.yml ([3c67ac8](https://github.com/ory/kratos/commit/3c67ac83f58c5b03dc3935d279083268b8a85e0d))
* Initial commit ([ab6f24a](https://github.com/ory/kratos/commit/ab6f24a85276bdd8687f2fc06390c1279892b005))
* Add ability to define multiple schemas and serve them over HTTP ([#164](https://github.com/ory/kratos/issues/164)) ([c65119c](https://github.com/ory/kratos/commit/c65119c24378dabd306e5a49f89c28c0367f7c2e)), closes [#86](https://github.com/ory/kratos/issues/86):

    All identity traits schemas have to be configured using a human readable ID and the corresponding URL. This PR enables multiple schemas to be used next to the default schema.
    It also adds the kratos.public/schemas/:id endpoint that mirrors all schemas.

* Add helper for requiring authentication ([3888fbd](https://github.com/ory/kratos/commit/3888fbdc239b7a06c7fca34d08de7d55af69a48c))
* Add helpers for go-swagger ([165a660](https://github.com/ory/kratos/commit/165a660f277588ed572d7843354c207f72f1678d)):

    See https://github.com/go-swagger/go-swagger/issues/2119

* Add profile management and refactor internals ([3ec9263](https://github.com/ory/kratos/commit/3ec9263f597a5949d0de6d10073cc626cfcfcca4)), closes [#112](https://github.com/ory/kratos/issues/112)
* Add session destroyer hook  ([#148](https://github.com/ory/kratos/issues/148)) ([d17f002](https://github.com/ory/kratos/commit/d17f002cdfe1f11ebb6bcbb17f6976aa329eab4a)), closes [#139](https://github.com/ory/kratos/issues/139):

    This patch adds a hook that destroys all active session by the identity which is being logged in. This can be useful in scenarios where only one session should be active at any given time.

* Add SQL adapter ([#100](https://github.com/ory/kratos/issues/100)) ([9e7f998](https://github.com/ory/kratos/commit/9e7f99871e3f09e7ae9ec1c38c8b8cf94d076f45)), closes [#92](https://github.com/ory/kratos/issues/92)
* Explicitly whitelist form parser keys ([#105](https://github.com/ory/kratos/issues/105)) ([28b056e](https://github.com/ory/kratos/commit/28b056e5bbfec645262914c52f0386d70c787a32)), closes [#98](https://github.com/ory/kratos/issues/98):

    Previously the form parser would try to detect the field type by
    asserting types for the whole form. That caused passwords
    containing only numbers to fail to unmarshal into a string
    value.
    
    This patch resolves that issue by introducing a prefix
    option to the BodyParser

* Fix broken import ([308aa13](https://github.com/ory/kratos/commit/308aa1334dd43bc4bebade4e70e9c81c83fe8806))
* Handle securecookie errors appropriately ([#101](https://github.com/ory/kratos/issues/101)) ([75bf6fe](https://github.com/ory/kratos/commit/75bf6fe3f79d025f2aaa79d06db39c26430dc3fc)), closes [#97](https://github.com/ory/kratos/issues/97):

    Previously, IsNotAuthenticated would not handle securecookie errors appropriately.
    This has been resolved.

* Implement CRUD for identities ([#60](https://github.com/ory/kratos/issues/60)) ([58a3c24](https://github.com/ory/kratos/commit/58a3c240fca66e1195bf310024a2f8473826bce6)), closes [#58](https://github.com/ory/kratos/issues/58)
* Implement message templates and SMTP delivery ([#146](https://github.com/ory/kratos/issues/146)) ([dc674bf](https://github.com/ory/kratos/commit/dc674bfa7d1fa9ee94b014d09866bbdc0a97c321)), closes [#99](https://github.com/ory/kratos/issues/99):

    This patch adds a message templates (with override capabilities)
    and SMTP delivery.
    
    Integration tests using MailHog test fault resilience and e2e email
    delivery.
    
    This system is designed to be extended for SMS and other use cases.

* Improve migration command ([#94](https://github.com/ory/kratos/issues/94)) ([2b631de](https://github.com/ory/kratos/commit/2b631de6d621dcebac5318f6dd628646fec7712f))
* Inject Identity Traits JSON Schema ([3a4c5ad](https://github.com/ory/kratos/commit/3a4c5ad35f885c7d38ffcf1d5836fb485f122fe9)), closes [#189](https://github.com/ory/kratos/issues/189)
* Mark active field as nullable ([#89](https://github.com/ory/kratos/issues/89)) ([292702d](https://github.com/ory/kratos/commit/292702d9e031e43c63e0ecb59354557139499e87))
* Move package to selfservice ([063b767](https://github.com/ory/kratos/commit/063b7679af76333fc546e94e92b197079e5bdb30)):

    Because this module is primarily used
    in selfservice scenarios, it has been
    moved to the selfservice parent.

* Omit request header from login/registration request ([#106](https://github.com/ory/kratos/issues/106)) ([9b07587](https://github.com/ory/kratos/commit/9b07587f2de2b270c5c326e37b2b6b3dbbfa8595)), closes [#95](https://github.com/ory/kratos/issues/95):

    When fetching a login and registration request, the HTTP Request Headers
    must not be included in the response, as they contain irrelevant
    information for the API caller.

* Properly handle empty credentials config in sql ([#93](https://github.com/ory/kratos/issues/93)) ([b79c5d1](https://github.com/ory/kratos/commit/b79c5d1d5216e994f986ce739285cb1a89523df5))
* Re-introduce migration plans to CLI command ([#192](https://github.com/ory/kratos/issues/192)) ([bb32cd3](https://github.com/ory/kratos/commit/bb32cd3cad3cd0bd6f3166de0166701e1f676ac6)), closes [#131](https://github.com/ory/kratos/issues/131)
* Reset CSRF token on principal change ([#64](https://github.com/ory/kratos/issues/64)) ([9c889ab](https://github.com/ory/kratos/commit/9c889ab4f6c846812a4290545fef7d8106da35f0)), closes [#38](https://github.com/ory/kratos/issues/38):

    Add tests for logout.

* Resolve wrong column reference in sql ([#90](https://github.com/ory/kratos/issues/90)) ([0c0eb87](https://github.com/ory/kratos/commit/0c0eb87cd341bd3e73eb9adb303054b38c103ba9)):

    Reference ic.method instead of ici.method.
    
    Added regression tests against this particular issue.

* Update keyword from kratos to ory.sh/kratos ([f45cbe0](https://github.com/ory/kratos/commit/f45cbe0339db8d129522314f3099e6944e4a6ea3)), closes [#115](https://github.com/ory/kratos/issues/115)
* Update sdk generation method ([24aa3d7](https://github.com/ory/kratos/commit/24aa3d73354d5a28f05999a09e7bbbe51a44d44e))
* Update to ory/x 0.0.80 ([#110](https://github.com/ory/kratos/issues/110)) ([64de2f8](https://github.com/ory/kratos/commit/64de2f86540bf8715a1703d773fa95011603a854)):

    Removes the need for BindEnv()

* Use JSON Schema to type assert form body ([#116](https://github.com/ory/kratos/issues/116)) ([1944c7c](https://github.com/ory/kratos/commit/1944c7c6e82b5b6a3b9d47db94c8f8f45248feb7)), closes [#109](https://github.com/ory/kratos/issues/109)
