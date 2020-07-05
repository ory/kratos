# Changelog

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Unreleased (2020-07-05)](#unreleased-2020-07-05)
    - [Bug Fixes](#bug-fixes)
    - [Code Refactoring](#code-refactoring)
    - [Documentation](#documentation)
    - [Features](#features)
    - [Unclassified](#unclassified)
    - [BREAKING CHANGES](#breaking-changes)
- [0.3.0-alpha.1 (2020-05-15)](#030-alpha1-2020-05-15)
    - [Bug Fixes](#bug-fixes-1)
    - [Code Refactoring](#code-refactoring-1)
    - [Documentation](#documentation-1)
    - [Features](#features-1)
    - [Unclassified](#unclassified-1)
    - [BREAKING CHANGES](#breaking-changes-1)
  - [0.2.1-alpha.1 (2020-05-05)](#021-alpha1-2020-05-05)
    - [Documentation](#documentation-2)
- [0.2.0-alpha.2 (2020-05-04)](#020-alpha2-2020-05-04)
    - [Bug Fixes](#bug-fixes-2)
    - [Code Refactoring](#code-refactoring-2)
    - [Documentation](#documentation-3)
    - [Features](#features-2)
    - [Unclassified](#unclassified-2)
    - [BREAKING CHANGES](#breaking-changes-2)
  - [0.1.1-alpha.1 (2020-02-18)](#011-alpha1-2020-02-18)
    - [Bug Fixes](#bug-fixes-3)
    - [Code Refactoring](#code-refactoring-3)
    - [Documentation](#documentation-4)
- [0.1.0-alpha.6 (2020-02-16)](#010-alpha6-2020-02-16)
    - [Bug Fixes](#bug-fixes-4)
    - [Code Refactoring](#code-refactoring-4)
    - [Documentation](#documentation-5)
    - [Features](#features-3)
- [0.1.0-alpha.5 (2020-02-06)](#010-alpha5-2020-02-06)
    - [Documentation](#documentation-6)
    - [Features](#features-4)
- [0.1.0-alpha.4 (2020-02-06)](#010-alpha4-2020-02-06)
    - [Documentation](#documentation-7)
- [0.1.0-alpha.3 (2020-02-06)](#010-alpha3-2020-02-06)
- [0.1.0-alpha.2 (2020-02-03)](#010-alpha2-2020-02-03)
    - [Bug Fixes](#bug-fixes-5)
    - [Documentation](#documentation-8)
    - [Features](#features-5)
    - [Unclassified](#unclassified-3)
- [0.1.0-alpha.1 (2020-01-31)](#010-alpha1-2020-01-31)
    - [Documentation](#documentation-9)
  - [0.0.3-alpha.15 (2020-01-31)](#003-alpha15-2020-01-31)
    - [Unclassified](#unclassified-4)
  - [0.0.3-alpha.14 (2020-01-31)](#003-alpha14-2020-01-31)
    - [Unclassified](#unclassified-5)
  - [0.0.3-alpha.13 (2020-01-31)](#003-alpha13-2020-01-31)
    - [Unclassified](#unclassified-6)
  - [0.0.3-alpha.11 (2020-01-31)](#003-alpha11-2020-01-31)
    - [Unclassified](#unclassified-7)
  - [0.0.3-alpha.10 (2020-01-31)](#003-alpha10-2020-01-31)
    - [Unclassified](#unclassified-8)
  - [0.0.3-alpha.7 (2020-01-30)](#003-alpha7-2020-01-30)
    - [Unclassified](#unclassified-9)
  - [0.0.3-alpha.5 (2020-01-30)](#003-alpha5-2020-01-30)
    - [Unclassified](#unclassified-10)
  - [0.0.3-alpha.4 (2020-01-30)](#003-alpha4-2020-01-30)
    - [Unclassified](#unclassified-11)
  - [0.0.3-alpha.2 (2020-01-30)](#003-alpha2-2020-01-30)
    - [Unclassified](#unclassified-12)
  - [0.0.3-alpha.1 (2020-01-30)](#003-alpha1-2020-01-30)
    - [Unclassified](#unclassified-13)
  - [0.0.1-alpha.9 (2020-01-29)](#001-alpha9-2020-01-29)
  - [0.0.2-alpha.1 (2020-01-29)](#002-alpha1-2020-01-29)
    - [Unclassified](#unclassified-14)
  - [0.0.1-alpha.6 (2020-01-29)](#001-alpha6-2020-01-29)
  - [0.0.1-alpha.5 (2020-01-29)](#001-alpha5-2020-01-29)
    - [Unclassified](#unclassified-15)
  - [0.0.1-alpha.3 (2020-01-28)](#001-alpha3-2020-01-28)
  - [0.0.1-alpha.2 (2020-01-28)](#001-alpha2-2020-01-28)
  - [0.0.1-alpha.1 (2020-01-28)](#001-alpha1-2020-01-28)
    - [Documentation](#documentation-10)
    - [Unclassified](#unclassified-16)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# [Unreleased](https://github.com/ory/kratos/compare/v0.3.0-alpha.1...39c1234f8ff3f6c7b0923053c8a317677d6cb667) (2020-07-05)


### Bug Fixes

* Account recovery can't use recovery token ([#526](https://github.com/ory/kratos/issues/526)) ([379f24e](https://github.com/ory/kratos/commit/379f24e96e50a3e5c71b53a11195bdd84a8dc957)), closes [#525](https://github.com/ory/kratos/issues/525)
* Add and document recovery to quickstart ([c229c54](https://github.com/ory/kratos/commit/c229c54603bdc3efb863fd76b64096ae599d1aac))
* Add pkger to docker builds ([d3ef5a0](https://github.com/ory/kratos/commit/d3ef5a0fe90f430999d0d94cb2f55acc8d628212))
* Allow linking oidc credentials without existing oidc connection ([#548](https://github.com/ory/kratos/issues/548)) ([39c1234](https://github.com/ory/kratos/commit/39c1234f8ff3f6c7b0923053c8a317677d6cb667)), closes [#532](https://github.com/ory/kratos/issues/532)
* Clear error messages after updating settings successfully ([#421](https://github.com/ory/kratos/issues/421)) ([7eec388](https://github.com/ory/kratos/commit/7eec38829449237cffe345d8bec67578764559be)), closes [#420](https://github.com/ory/kratos/issues/420)
* Do not send debug on session/whoami ([16d3670](https://github.com/ory/kratos/commit/16d3670070bf46170c4540203e8380ad81bfb4c3)), closes [#483](https://github.com/ory/kratos/issues/483)
* Document login refresh parameter in swagger ([#482](https://github.com/ory/kratos/issues/482)) ([6b94993](https://github.com/ory/kratos/commit/6b949936725a6100a31851a5d879c877c2c76cbf))
* Embedded video link properly ([#514](https://github.com/ory/kratos/issues/514)) ([962bbc6](https://github.com/ory/kratos/commit/962bbc6e4af0797c190418b812f6298372dabdde))
* Embedded video link properly ([#515](https://github.com/ory/kratos/issues/515)) ([821ca93](https://github.com/ory/kratos/commit/821ca93838a360551378e336e9ce10cfe13369ec))
* Enable recovery for quickstart ([0ccc651](https://github.com/ory/kratos/commit/0ccc651f809b1e39dd6c41b88f1a10c67451eae2))
* Improve grammar of similar password error ([#471](https://github.com/ory/kratos/issues/471)) ([39873bf](https://github.com/ory/kratos/commit/39873bfad89a654fe12e101b54e9b0c2f95714ec))
* Initialize verification request with correct state ([3264ecf](https://github.com/ory/kratos/commit/3264ecfbb8f7b34d9dbb22237df8d9f591ac09f3)), closes [#543](https://github.com/ory/kratos/issues/543)
* Re-add all databases to persister ([#527](https://github.com/ory/kratos/issues/527)) ([b04d178](https://github.com/ory/kratos/commit/b04d17815b5a28b5fe73a6a94ce1d907a63115e1))
* Re-add redirect targets for quickstart ([3c48ad2](https://github.com/ory/kratos/commit/3c48ad26961560d6e10a627a64052e316d9ffdc7))
* Reduce docker bloat by ignoring docs and others ([ecc555b](https://github.com/ory/kratos/commit/ecc555b5ad0fa888a8d5ba39cc09094fd251e655))
* Resolve broken redirect in verify flow ([a9ca8fd](https://github.com/ory/kratos/commit/a9ca8fd793347ed8e4404a4bd29e330a3f1ef684)), closes [#436](https://github.com/ory/kratos/issues/436)
* Respect multiple secrets and fix used flag ([#526](https://github.com/ory/kratos/issues/526)) ([b16c2b8](https://github.com/ory/kratos/commit/b16c2b80edfc78afca0c72fa8da7d73b51b3075a)), closes [#525](https://github.com/ory/kratos/issues/525)
* Respect self-service enabled flag ([#470](https://github.com/ory/kratos/issues/470)) ([b198faf](https://github.com/ory/kratos/commit/b198fafce9d96fbb644300243e6a757242fbbd06)), closes [#417](https://github.com/ory/kratos/issues/417):

    > Respects the `enabled` flag for self-service strategies.
    > 
    > Also a new testhelper function was needed, to defer route registration
    > (because whether strategies are enabled or not is determined only once:
    > at route registration)
* Typo accent -> account ([984d978](https://github.com/ory/kratos/commit/984d978cf44763d916a9329742d046e00f21577b))
* Use correct brew replacements ([fd269b1](https://github.com/ory/kratos/commit/fd269b1afa784becac7ee79cd7a6f9d2bbe39121)), closes [#423](https://github.com/ory/kratos/issues/423)
* Write migration tests ([#499](https://github.com/ory/kratos/issues/499)) ([d32413a](https://github.com/ory/kratos/commit/d32413a1fcd0ce1a82d2529f18b5d4334a490a2a)), closes [#481](https://github.com/ory/kratos/issues/481)


### Code Refactoring

* Improve and simplify configuration ([#536](https://github.com/ory/kratos/issues/536)) ([8e7f9f5](https://github.com/ory/kratos/commit/8e7f9f5ec3ac6f5675584974e8d189247b539634)), closes [#432](https://github.com/ory/kratos/issues/432)
* Move schema packing to pkger ([173f9d2](https://github.com/ory/kratos/commit/173f9d2b09d597376490b5d4588f7c0a4f525857))
* Move verify fallback to verification ([1ce6469](https://github.com/ory/kratos/commit/1ce64695ec61c3a31e00875069d2847be502744b))
* Rename prompt=login to refresh=true ([#478](https://github.com/ory/kratos/issues/478)) ([c04346e](https://github.com/ory/kratos/commit/c04346e0f01aa7ce5627c0b7135032b225e7faf9)), closes [#477](https://github.com/ory/kratos/issues/477)
* Replace settings update_successful with state ([#488](https://github.com/ory/kratos/issues/488)) ([ca3b3f4](https://github.com/ory/kratos/commit/ca3b3f4dbdcd75ceb13c9a1b2c8dc991aba7c7e4)), closes [#449](https://github.com/ory/kratos/issues/449)
* Text errors to text messages ([#476](https://github.com/ory/kratos/issues/476)) ([8106951](https://github.com/ory/kratos/commit/81069514e5ef1d851f76d44bb45d6a896d4985a6)), closes [#428](https://github.com/ory/kratos/issues/428):

    > This patch implements a better way to deal with text messages by giving them a unique ID, a context, and a default message.


### Documentation

* Add azure to next docs ([e1dd3fa](https://github.com/ory/kratos/commit/e1dd3fad30a07be6f105201a8478642e9792df46))
* Add fixme note for viper workaround ([7e3eef6](https://github.com/ory/kratos/commit/7e3eef6d36dcbb1a06ce0a20e2de0874a7dc5d38)):

    > See https://github.com/ory/x/issues/169
* Add guide for setting up account recovery ([bbf3762](https://github.com/ory/kratos/commit/bbf37620d5b47fd18cb754c8ed43856652ee33c0))
* Add guide for setting up email verification ([1435cbc](https://github.com/ory/kratos/commit/1435cbcea5d45c9cde1a0eb7e5ebb66ce65c4b82))
* Add guide for SSO via Google ([#424](https://github.com/ory/kratos/issues/424)) ([5c45b16](https://github.com/ory/kratos/commit/5c45b1653791cc3ab5d4e4694da98da7543e816d))
* Add new guides to sidebar ([24c5cbc](https://github.com/ory/kratos/commit/24c5cbc129ad185ec02883c3451d7e573409b865))
* Added video tutorials to guides ([#513](https://github.com/ory/kratos/issues/513)) ([956731d](https://github.com/ory/kratos/commit/956731d562f33f2849197b2e692a4f20b18279f9))
* Added youtube manual ([#490](https://github.com/ory/kratos/issues/490)) ([ec232f7](https://github.com/ory/kratos/commit/ec232f72d7204b2cdf946874d51f7473a10a76a4))
* Connecting Kratos to AzureAD ([#433](https://github.com/ory/kratos/issues/433)) ([7660bcd](https://github.com/ory/kratos/commit/7660bcd2ba90d83c4ab0683a2f011e6841b2c810))
* Correct claims.email in github guide ([#422](https://github.com/ory/kratos/issues/422)) ([052a622](https://github.com/ory/kratos/commit/052a622de79d34e32ccab9c7da12a1275c7be51b)):

    > There is no email_primary in claims, and the selfservice strategy is currently using claims.email.
* Correct claims.email in github guide ([#422](https://github.com/ory/kratos/issues/422)) ([58f7e15](https://github.com/ory/kratos/commit/58f7e15093d2461d4322fe68adb0723ae244bed9)):

    > There is no email_primary in claims, and the selfservice strategy is currently using claims.email.
* Correct link in user-settings ([d13317d](https://github.com/ory/kratos/commit/d13317d9bf71db775067a7c17f4c98cdbf1cc7e5))
* Correct SDK use in quickstart ([#480](https://github.com/ory/kratos/issues/480)) ([dfdf975](https://github.com/ory/kratos/commit/dfdf9751d9333994a49537d82a15b780ebd8bc76)), closes [#430](https://github.com/ory/kratos/issues/430)
* Correct stray dot ([e820f41](https://github.com/ory/kratos/commit/e820f41e63aff1a85094a9e14dfd968353ae6b1b))
* Correct user settings render form ([197e246](https://github.com/ory/kratos/commit/197e24603fc67707131e54e52e1bfb52011ca839))
* Delete old redirect homepage ([b6d9244](https://github.com/ory/kratos/commit/b6d9244b5d683f5baf27e9af5970596261a4fd20))
* Document new account recovery feature ([2252a86](https://github.com/ory/kratos/commit/2252a8676e573b9ade85814acc40b212dcfd48c1)), closes [#436](https://github.com/ory/kratos/issues/436)
* Document refresh=true for login ([#479](https://github.com/ory/kratos/issues/479)) ([2ab5ead](https://github.com/ory/kratos/commit/2ab5ead77517ab5b750835195ab6673e219da71a)), closes [#464](https://github.com/ory/kratos/issues/464)
* Embedded quickstart video ([#491](https://github.com/ory/kratos/issues/491)) ([ee80346](https://github.com/ory/kratos/commit/ee80346a30ebc2c7b06292e58bd3578e002e242a))
* Fix broken link ([aa9d3e6](https://github.com/ory/kratos/commit/aa9d3e6347375170a84ba53b2a9050c9544e7e2a))
* Fix broken link ([#506](https://github.com/ory/kratos/issues/506)) ([dac8dfd](https://github.com/ory/kratos/commit/dac8dfd970255f8e79e7fc7811f563e6903f6fc9)):

    > The rest api is no longer under sdk but under reference.
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

    > Fixed a few broken links, .md in the url was the problem.
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
* Properly annotate forms disabled field ([#486](https://github.com/ory/kratos/issues/486)) ([be1acb3](https://github.com/ory/kratos/commit/be1acb3d161412d18599c970364f0c91fa6ebffb)), closes [/github.com/ory/kratos/pull/467#discussion_r434764266](https://github.com//github.com/ory/kratos/pull/467/issues/discussion_r434764266)
* Remove rogue slash and fix closing tag ([#521](https://github.com/ory/kratos/issues/521)) ([3fd1076](https://github.com/ory/kratos/commit/3fd1076929eeecffb7e8aa8e906970774283daeb))
* Rename redirect page to browser-redirect-flow-completion ([ae77d48](https://github.com/ory/kratos/commit/ae77d48a3435069556382b9403cb1ad45a9d7c07))
* Replace mailhog references with mailslurper ([#509](https://github.com/ory/kratos/issues/509)) ([d0e5a0f](https://github.com/ory/kratos/commit/d0e5a0fa64e2d46437fb2abd17dc306bdec34a91))
* Run format ([2b3f299](https://github.com/ory/kratos/commit/2b3f29913be844498a02b9869789c2b2d4aaacf8))
* Typos and stale links ([29fb466](https://github.com/ory/kratos/commit/29fb466d9881b6574ee697d7e25e45785f07114b))
* Typos and stale links ([#510](https://github.com/ory/kratos/issues/510)) ([7557ab8](https://github.com/ory/kratos/commit/7557ab85ddf8501935d70e2558682dff2024897b))
* Update repository templates ([4c89834](https://github.com/ory/kratos/commit/4c89834ce59195c5b59da5bc5b41db7ed03bf1c4))
* Use central banner repo for README ([d1e8a82](https://github.com/ory/kratos/commit/d1e8a8272cd536b6e12326778258bfbe0b7e8af7))
* Use shorthand closing tag for Mermaid ([f9f2dbc](https://github.com/ory/kratos/commit/f9f2dbc063f82a852b540013ddff81501f7c1222))


### Features

* Add tests for defaults ([a16fc51](https://github.com/ory/kratos/commit/a16fc5121b36353cf2e684190eda976a1ea53a8f))
* Add User ID to a header when calling whoami ([#530](https://github.com/ory/kratos/issues/530)) ([183b4d0](https://github.com/ory/kratos/commit/183b4d075a9ff50c1f9f53d108a48789e49a5138))
* Implement account recovery ([#428](https://github.com/ory/kratos/issues/428)) ([e169a3e](https://github.com/ory/kratos/commit/e169a3e4079b1ef3a18564e0723baf81c44c38ec)), closes [#37](https://github.com/ory/kratos/issues/37):

    > This patch implements the account recovery with endpoints such as "Init Account Recovery", a new config value `urls.recovery_ui` and so on. A new identity field has been added `identity.recovery_addresses` containing all recovery addresses.
    > 
    > Additionally, some refactoring was made to DRY code and make naming consistent. As part of dependency upgrades, structured logging has also improved and an audit trail prototype has been added (currently streams to stderr only).


### Unclassified

* Allow kratos to talk to databases in docker-compose quickstart ([#522](https://github.com/ory/kratos/issues/522)) ([8bf9a1a](https://github.com/ory/kratos/commit/8bf9a1ac4162c677a455c2f02de658bd5d146905)):

    > All of the databases must exist on the same docker network to allow the
    > main kratos applications to communicate with them.
* Fixed typo ([#472](https://github.com/ory/kratos/issues/472)) ([31263b6](https://github.com/ory/kratos/commit/31263b68ab8d81d264e0fa375a915f8f82d70bb3))
* docs:fixed broken link (#454) ([22720c6](https://github.com/ory/kratos/commit/22720c6c5e3d31acc175980223183e2336b3751d)), closes [#454](https://github.com/ory/kratos/issues/454)


### BREAKING CHANGES

* To address these refactorings, the configuration had to be changed and with breaking changes
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
* Replaces the `update_successful` field of the settings request
with a field called `state` which can be either `show_form` or `success`.
* Flows, request methods, form fields have had a key errors to show e.g. validation errors such as ("not an email address", "incorrect username/password", and so on. The `errors` key is now called `messages`. Each message now has a `type` which can be `error` or `info`, an `id` which can be used to translate messages, a `text` (which was previously errors[*].message). This affects all login, request, settings, and recovery flows and methods.
* To refresh a login session it is now required to append `refresh=true` instead of `prompt=login` as the second has implications for revoking an existing issue and might be confusing when used in combination with OpenID Connect.
* * Applying this patch requires running SQL Migrations.
* The field `identity.addresses` has moved to `identity.verifiable_addresses`.
* Configuration key `selfservice.verification.link_lifespan`
has been merged with  `selfservice.verification.request_lifespan`.



# [0.3.0-alpha.1](https://github.com/ory/kratos/compare/v0.2.1-alpha.1...v0.3.0-alpha.1) (2020-05-15)


### Bug Fixes

* Access rules of oathkeeper for quick start ([#390](https://github.com/ory/kratos/issues/390)) ([5ed6d05](https://github.com/ory/kratos/commit/5ed6d05b3e13027e4e7ffef1ff10ab2fb948093d)), closes [#389](https://github.com/ory/kratos/issues/389):

    > To access `/` as dashboard
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


### Code Refactoring

* Adopt new request parser ([ad16cc9](https://github.com/ory/kratos/commit/ad16cc917c8067eb1c4b89ef8192287be1c912c8))
* Dry config and oidc tests ([3e98756](https://github.com/ory/kratos/commit/3e9875612ea895f9b565d34f4d5b0f80d136868f))
* Improve oidc flows and payloads and add e2e tests ([#381](https://github.com/ory/kratos/issues/381)) ([f9a5079](https://github.com/ory/kratos/commit/f9a50790637a848897ba275373bc538728e09f3d)), closes [#387](https://github.com/ory/kratos/issues/387):

    > This patch improves the OpenID Connect login and registration user experience by simplifying the network flows and introduces e2e tests using ORY Hydra.
* Move cypress files to test/e2e ([df8e627](https://github.com/ory/kratos/commit/df8e627d81d69682e01ec5670c7088ba564df578))
* Partition files and change creds structure ([4f1eb94](https://github.com/ory/kratos/commit/4f1eb946fe1e74e537fc2166fc000180a11c2048)):

    > This patch changes the data model of the OpenID Connect strategy. Instead of using an array of providers as the base config item (e.g. `{"type":"oidc","config":[{"provider":"google","subject":"..."}]}`) the credentials config is now an object with a `providers` key: `{"type":"oidc","config":{"providers":[{"provider":"google","subject":"..."}]}}`. This change allows introduction of future changes to the schema without breaking compatibility.
* **settings:** Use common request parser ([ad6c402](https://github.com/ory/kratos/commit/ad6c4026e5fd15924dc906cdc9cb6c9de2fc4daa))
* Moved scanner json to ory/x ([#412](https://github.com/ory/kratos/issues/412)) ([8a0967d](https://github.com/ory/kratos/commit/8a0967daef4329981b01e6c2b8bb55a8105b4829))
* Replace oidc jsonschema with jsonnet ([2b45e79](https://github.com/ory/kratos/commit/2b45e7953787ad46a6937fe44cb24b6c786eb223)), closes [#380](https://github.com/ory/kratos/issues/380):

    > This patch replaces the previous methodology of merging OIDC data which used JSON Schema with Extensions and JSON Path in favor of a much easier to use approach with JSONNet.


### Documentation

* Document account enumeration defenses for oidc ([266329c](https://github.com/ory/kratos/commit/266329cd2969627c823418c1267360193e6342df)), closes [#32](https://github.com/ory/kratos/issues/32)
* Document new oidc jsonnet mapper ([#392](https://github.com/ory/kratos/issues/392)) ([088b30f](https://github.com/ory/kratos/commit/088b30feb6845863e6651489e0c963cde7e10516))
* Document oidc strategy ([#415](https://github.com/ory/kratos/issues/415)) ([9f079f4](https://github.com/ory/kratos/commit/9f079f4f77e54f7be67ac59e13e8ec2696522637)), closes [#409](https://github.com/ory/kratos/issues/409) [#124](https://github.com/ory/kratos/issues/124) [#32](https://github.com/ory/kratos/issues/32)
* Explain that form data is merged with oidc data ([#394](https://github.com/ory/kratos/issues/394)) ([b0dbec4](https://github.com/ory/kratos/commit/b0dbec403c96af41346b6b14fc74b7010e7f8e8a)), closes [#127](https://github.com/ory/kratos/issues/127)
* Fix links in README ([efb6102](https://github.com/ory/kratos/commit/efb610239ac2ae828db26ee84c4c5a83c54c0a6a)), closes [#403](https://github.com/ory/kratos/issues/403)
* Improve social sign in guide ([#393](https://github.com/ory/kratos/issues/393)) ([647ced3](https://github.com/ory/kratos/commit/647ced3084d203e9954ca037afea34316f2080d8)), closes [#49](https://github.com/ory/kratos/issues/49):

    > This patch changes the social sign in guide to represent more use cases such as Google and Facebook. Additionally, the example has been updated to work with Jsonnet.
    > 
    > This patch also documents limitations around merging user data from GitHub.
* Improve the identity data model page ([#410](https://github.com/ory/kratos/issues/410)) ([2915b8f](https://github.com/ory/kratos/commit/2915b8faf3530fe7b9d252094c3aeb9fdbe9dd08))
* Include redirect doc in nav ([5aaebff](https://github.com/ory/kratos/commit/5aaebffd8c03e613ec60735536b6ef38d4da39e3)), closes [#406](https://github.com/ory/kratos/issues/406)
* Prepare v0.3.0-alpha.1 ([d6a6f43](https://github.com/ory/kratos/commit/d6a6f432f375018a2dc79d6b60de18455057c25a))
* Ui should show only active form sections ([#395](https://github.com/ory/kratos/issues/395)) ([4db674d](https://github.com/ory/kratos/commit/4db674de14bc50e782321c7bd88ac8077db2bf75))
* Update github templates ([#408](https://github.com/ory/kratos/issues/408)) ([6e646b0](https://github.com/ory/kratos/commit/6e646b033e0d43499bf37579a2f04b726af0e3f7))


### Features

* Add format and lint for JSONNet files ([0a1b244](https://github.com/ory/kratos/commit/0a1b244a6fd2f714a12d101071b3c0f82b4da584)):

    > This patch adds two commands `kratos jsonnet format` and `kratos jsonnet lint` that help with formatting and linting JSONNet code.
* Implement oidc settings e2e tests ([919925c](https://github.com/ory/kratos/commit/919925c87be561064300c3981b5a230c6cada4f7))
* Introduce leaklog for debugging oidc map payloads ([238d7a4](https://github.com/ory/kratos/commit/238d7a493566bcc28f08b1b2bf6463f95b100254))
* Write tests and fix bugs for oidc settings ([575a61f](https://github.com/ory/kratos/commit/575a61f58a887fefa6b2917761c06304c94c9892))


### Unclassified

* Format code ([bc7557a](https://github.com/ory/kratos/commit/bc7557a4247ede1fdb4141f2670532aec7cbd456))


### BREAKING CHANGES

* If you upgrade and have existing Social Sign In connections, it will no longer be possible to use them to sign in. Because the oidc strategy was undocumented and not officially released we do not provide an upgrade guide. If you run into this issue on a production system you may need to use SQL to change the config of those identities. If this is a real issue for you that you're unable to solve, please create an issue on GitHub.
* This is a breaking change as previous OIDC configurations will not work. Please consult the newly written documentation on OpenID Connect to learn how to use OIDC in your login and registration flows. Since the OIDC feature was not publicly broadcasted yet we have chosen not to provide an upgrade path. If you have issues, please reach out on the forums or slack.



## [0.2.1-alpha.1](https://github.com/ory/kratos/compare/v0.2.0-alpha.2...v0.2.1-alpha.1) (2020-05-05)


### Documentation

* Fix quickstart hero sections ([7c6c439](https://github.com/ory/kratos/commit/7c6c4397bccd2b505fc04cc8d3b0944ceca18982))
* Fix typo in upgrade guide ([a1b1d7c](https://github.com/ory/kratos/commit/a1b1d7c9cbe5fad3b1112a16eced4f3064cfdda0))



# [0.2.0-alpha.2](https://github.com/ory/kratos/compare/v0.1.1-alpha.1...v0.2.0-alpha.2) (2020-05-04)


### Bug Fixes

* Allow setting new password in profile flow ([3b5fd5c](https://github.com/ory/kratos/commit/3b5fd5ca8c09b2344c0262547f2b387bda362362))
* Automatically append multiStatements parameter to mySQL URI ([#374](https://github.com/ory/kratos/issues/374)) ([39f77bb](https://github.com/ory/kratos/commit/39f77bb29637db048b15c097d869d8828b0d292b))
* Create pop connection without parsed connection options ([#366](https://github.com/ory/kratos/issues/366)) ([10b6481](https://github.com/ory/kratos/commit/10b6481774aaff42b70b9c6af3ed776ac8f7734c))
* Declare proper vars for setting version ([#383](https://github.com/ory/kratos/issues/383)) ([2fc7556](https://github.com/ory/kratos/commit/2fc7556b70b11e519162326ded0ba2638b6d32df))
* Decouple quickstart scenarios ([#336](https://github.com/ory/kratos/issues/336)) ([17363b3](https://github.com/ory/kratos/commit/17363b312deff8b92fc1b0d158dc70670d5938e5)), closes [#262](https://github.com/ory/kratos/issues/262):

    > Creates several docker compose examples which include various
    > scenarios of the quickstart.
    > 
    > The regular quickstart guide now works without ORY Oathkeeper
    > and uses the standalone mode of the example app instead.
    > 
    > Additionally, the Makefile was improved and now automatically pulls
    > required dependencies in the appropriate version.
* Document Schema API and serve over admin endpoint ([#299](https://github.com/ory/kratos/issues/299)) ([4be417c](https://github.com/ory/kratos/commit/4be417c0ee18622247a15d2803f7f436cfe3c229)), closes [#287](https://github.com/ory/kratos/issues/287)
* Exempt whomai from csrf protection ([#329](https://github.com/ory/kratos/issues/329)) ([31d4065](https://github.com/ory/kratos/commit/31d4065c2b0cbd6c8d2b0031ce8f6f157ff967cf))
* Fix swagger annotation ([#331](https://github.com/ory/kratos/issues/331)) ([5c5c78f](https://github.com/ory/kratos/commit/5c5c78f404a11d5df25cb68584b826b685bf5385)):

    > Closes https://github.com/ory/sdk/issues/10
* Move to ory sqa service ([#309](https://github.com/ory/kratos/issues/309)) ([7c244e0](https://github.com/ory/kratos/commit/7c244e0a28a010e56e07d061132dad7a0309ea75))
* Properly annotate error API ([a6f1300](https://github.com/ory/kratos/commit/a6f1300951010e7c862c410e93653f7c02c2e79f))
* Resolve docker build permission issues ([f3612e8](https://github.com/ory/kratos/commit/f3612e8f82018bae17c9146d273fe7e82ceb033d))
* Resolve failing test issues ([2e968e5](https://github.com/ory/kratos/commit/2e968e52d3ae3396a3f2e212c0dab22677b4b5fd))
* Resolve NULL value for seen_at ([#259](https://github.com/ory/kratos/issues/259)) ([a7d1e86](https://github.com/ory/kratos/commit/a7d1e86844a9cdd0c58353e1f1e4340dac4260b3)), closes [#244](https://github.com/ory/kratos/issues/244):

    > Previously, errorx tests were not executed which caused several bugs.
* Revert use host volume mount for sqlite ([#272](https://github.com/ory/kratos/issues/272)) ([#285](https://github.com/ory/kratos/issues/285)) ([a7477ab](https://github.com/ory/kratos/commit/a7477ab1db0d986f96e754946607d05888de4c97)):

    > This reverts commit 230ab2d83f4d187f410e267c6d68554e82514948.
* Show log in ui only when unauthenticated or forced ([df77310](https://github.com/ory/kratos/commit/df77310ffbe7cfc90fa3bc5dad0450e79c34ebef)), closes [#323](https://github.com/ory/kratos/issues/323)
* Use semver-regex replacer func ([d5c9a47](https://github.com/ory/kratos/commit/d5c9a47800fc2a55b96c7b9330f68b0a2db328cb))
* Use sqlite tag on make install ([2c82784](https://github.com/ory/kratos/commit/2c82784cd69e0468a72354f6898945032d826306))
* **docker:** Throw away build artifacts ([481ec1b](https://github.com/ory/kratos/commit/481ec1ba14480ced39516f6e0c47a40b6a44a631))
* Remove unused returnTo ([e64e5b0](https://github.com/ory/kratos/commit/e64e5b0cecceedda29a525f683cbf6070a9ef1eb))
* Resolve linux install script archive naming ([#302](https://github.com/ory/kratos/issues/302)) ([c98b8aa](https://github.com/ory/kratos/commit/c98b8aa4cd3ab881b904e9dc4cdcb6383a8ad09b))
* Resolve password continuity issues ([56a44fa](https://github.com/ory/kratos/commit/56a44fa33d325eea9fddec4269e34e632310f77b))
* Self-service error query parameter name ([#308](https://github.com/ory/kratos/issues/308)) ([be257f5](https://github.com/ory/kratos/commit/be257f5448abaa48e25735a088757f3fd6dc6d22)):

    > The query parameter for the self-service errors endpoint was named `id`
    > in the API docs, whereas it is the `error` param that is used by the
    > handler.
* Use host volume mount for sqlite ([#272](https://github.com/ory/kratos/issues/272)) ([230ab2d](https://github.com/ory/kratos/commit/230ab2d83f4d187f410e267c6d68554e82514948))
* Use resilient client for HIBP lookup ([#288](https://github.com/ory/kratos/issues/288)) ([735b435](https://github.com/ory/kratos/commit/735b43508392c6966a57907c20caa7cf9df4fc4d)), closes [#261](https://github.com/ory/kratos/issues/261)
* Verified_at field should not be required ([#353](https://github.com/ory/kratos/issues/353)) ([15d5e26](https://github.com/ory/kratos/commit/15d5e268d2ec397f0647d2407d86404c4ee8bfa3)):

    > Closes https://github.com/ory/sdk/issues/11
    > 
    > 
* **config:** Rename config key stmp to smtp ([#278](https://github.com/ory/kratos/issues/278)) ([ef95811](https://github.com/ory/kratos/commit/ef95811bb891afe3a0ef3b19514f13a56a32ea3b))
* **session:** Regenerate CSRF Token on principal change ([#290](https://github.com/ory/kratos/issues/290)) ([1527ef4](https://github.com/ory/kratos/commit/1527ef4209b937e2175b60d56efd019f17b33b04)), closes [#217](https://github.com/ory/kratos/issues/217)
* **session:** Whoami endpoint now supports all HTTP methods ([#283](https://github.com/ory/kratos/issues/283)) ([4bf645b](https://github.com/ory/kratos/commit/4bf645b66c7a128182ff55e52fdad7f53d752ce7)), closes [#270](https://github.com/ory/kratos/issues/270)
* **sql:** Rename migrations with same version ([#280](https://github.com/ory/kratos/issues/280)) ([07e46b9](https://github.com/ory/kratos/commit/07e46b9c9e57940bec904d744ffdd272d610a77b)), closes [#279](https://github.com/ory/kratos/issues/279)
* **swagger:** Move nolint,deadcode instructions to own file ([#293](https://github.com/ory/kratos/issues/293)) ([1935510](https://github.com/ory/kratos/commit/1935510ad9b0f387eb3b2e690e31c5313a06883e)):

    > Closes https://github.com/ory/docs/pull/279


### Code Refactoring

* Move docs to this repository ([#317](https://github.com/ory/kratos/issues/317)) ([aa0d726](https://github.com/ory/kratos/commit/aa0d72639ecae3b0649761e6ee881a59b2f3e94e))
* Prepare profile management payloads for credentials ([44493f3](https://github.com/ory/kratos/commit/44493f3ddbb449981576ec317ac45530ca3be14d))
* Rename traits method to profile ([4f1e033](https://github.com/ory/kratos/commit/4f1e0339ecc1efbdfa3d3680ad64b7683e90e447))
* Rework hooks and self-service flow completion ([#349](https://github.com/ory/kratos/issues/349)) ([a7c7fef](https://github.com/ory/kratos/commit/a7c7fef758e843393b0dc1e60bee11b88b8c9b4a)), closes [#348](https://github.com/ory/kratos/issues/348) [#347](https://github.com/ory/kratos/issues/347) [#179](https://github.com/ory/kratos/issues/179) [#51](https://github.com/ory/kratos/issues/51) [#50](https://github.com/ory/kratos/issues/50) [#31](https://github.com/ory/kratos/issues/31):

    > This patch focuses on refactoring how self-service flows terminate and
    > changes how hooks behave and when they are executed.
    > 
    > Before this patch, it was not clear whether hooks run before or
    > after an identity is persisted. This caused problems with multiple
    > writes on the HTTP ResponseWriter and other bugs.
    > 
    > This patch removes certain hooks from after login, registration, and profile flows.
    > Per default, these flows now respond with an appropriate payload (
    > redirect for browsers, JSON for API clients) and deprecate
    > the `redirect` hook. This patch includes documentation which explains
    > how these hooks work now.
    > 
    > Additionally, the documentation was updated. Especially the sections
    > about hooks have been refactored. The login and user registration docs
    > have been updated to reflect the latest changes as well.
    > 
    > Also, some other minor, cosmetic, changes to the documentation have been made.


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
* Update self service reg docs ([#367](https://github.com/ory/kratos/issues/367)) ([4cf0323](https://github.com/ory/kratos/commit/4cf0323095990c5ec25283a01561cb9b8833f9ef)), closes [/github.com/ory/kratos-selfservice-ui-node/blob/489c76d1b0474ee55ef56804b28f54d8718747ba/src/routes/auth.ts#L28](https://github.com//github.com/ory/kratos-selfservice-ui-node/blob/489c76d1b0474ee55ef56804b28f54d8718747ba/src/routes/auth.ts/issues/L28):

    > The old links pointed at `/auth/browser/(login|registration)`
    > which seems to be outdated now.
* Update user-settings-profile-management.md ([#322](https://github.com/ory/kratos/issues/322)) ([45dc3a5](https://github.com/ory/kratos/commit/45dc3a56c15ae442890313a7dbc784b75644248a))
* Updates issue and pull request templates ([#298](https://github.com/ory/kratos/issues/298)) ([1be738d](https://github.com/ory/kratos/commit/1be738d3f8e9bbc6dae31ffad5d990657a66761c))
* Updates issue and pull request templates ([#313](https://github.com/ory/kratos/issues/313)) ([299063c](https://github.com/ory/kratos/commit/299063caf2fdde40713bae4c36abb3b6fac7271d))
* Updates issue and pull request templates ([#314](https://github.com/ory/kratos/issues/314)) ([d5ae452](https://github.com/ory/kratos/commit/d5ae452a8ce5f641a40e510e82441d4eb8137218))
* Use git checkout <tag> in quickstart ([#339](https://github.com/ory/kratos/issues/339)) ([2d2562b](https://github.com/ory/kratos/commit/2d2562b587a69a2891ff29d927cb001e15d75b5d)), closes [#335](https://github.com/ory/kratos/issues/335)
* **concepts:** Fix typo ([a49184c](https://github.com/ory/kratos/commit/a49184c30d9c2ccff5a2d41d3aff61b24e7d2ea9)):

    > Closes https://github.com/ory/docs/pull/296
* **concepts:** Properly close code tag ([1c841c2](https://github.com/ory/kratos/commit/1c841c213bdbc79a6aa41e8450444d8d6c1f0284))
* Updates issue and pull request templates ([#315](https://github.com/ory/kratos/issues/315)) ([8b68db1](https://github.com/ory/kratos/commit/8b68db140a7fc1c0eaa9318c1759ea9d8d0c27df))


### Features

* Add `dsn: memory` shorthand ([#284](https://github.com/ory/kratos/issues/284)) ([e66a030](https://github.com/ory/kratos/commit/e66a030f7d67dec639121fb23dfc7f1444474c6b)), closes [#228](https://github.com/ory/kratos/issues/228)
* Add and test id hint in reauth flow ([2298f01](https://github.com/ory/kratos/commit/2298f0140e77da870c842daa8eaca274e5d64254)), closes [#323](https://github.com/ory/kratos/issues/323)
* Add cypress e2e tests ([#334](https://github.com/ory/kratos/issues/334)) ([abc0e91](https://github.com/ory/kratos/commit/abc0e91e278f7938b264598ac0c60d18c5a9e8a0))
* Allow configuring same-site for session cookies ([#303](https://github.com/ory/kratos/issues/303)) ([2eb2054](https://github.com/ory/kratos/commit/2eb2054a94281aefa9a0818110d168cc9c052094)), closes [#257](https://github.com/ory/kratos/issues/257):

    > It is now possible to set SameSite for the session cookie via the key `security.session.cookie.same_site`.
* Enable CockroachDB integration ([#260](https://github.com/ory/kratos/issues/260)) ([adc5153](https://github.com/ory/kratos/commit/adc5153410fb4d9f99702d7c73a78aeec8c1e9f1)), closes [#132](https://github.com/ory/kratos/issues/132) [#155](https://github.com/ory/kratos/issues/155)
* Enable continuity management for settings module ([009d755](https://github.com/ory/kratos/commit/009d7558f525168fecf86168de2906088662535e))
* Enable updating auth related traits ([#266](https://github.com/ory/kratos/issues/266)) ([65b88ba](https://github.com/ory/kratos/commit/65b88ba52fb9e6da3c1a65f734352519303327a6)), closes [#243](https://github.com/ory/kratos/issues/243)
* Implement password profile management flow ([a31839a](https://github.com/ory/kratos/commit/a31839a5c33c80500c900fb50d1dd499ab1161a1)), closes [#243](https://github.com/ory/kratos/issues/243)
* Introduce fallbacks for required configs ([#376](https://github.com/ory/kratos/issues/376)) ([b3bcb25](https://github.com/ory/kratos/commit/b3bcb25be6b417647ece2b3dda26d691f8e8d685)), closes [#369](https://github.com/ory/kratos/issues/369) [#352](https://github.com/ory/kratos/issues/352)
* Return 410 when selfservice requests expire ([#289](https://github.com/ory/kratos/issues/289)) ([b414607](https://github.com/ory/kratos/commit/b4146076148d9ff079e9d433f0a90f5bc938650c)), closes [#235](https://github.com/ory/kratos/issues/235)
* Send verification emails on profile update ([#333](https://github.com/ory/kratos/issues/333)) ([1cacc80](https://github.com/ory/kratos/commit/1cacc80c54f92b380ef3752591970cc4dd97085e)), closes [#267](https://github.com/ory/kratos/issues/267)
* **continuity:** Implement request continuity ([135e047](https://github.com/ory/kratos/commit/135e04750b1855ab0db812517c61e292a770ba94)), closes [#304](https://github.com/ory/kratos/issues/304) [#311](https://github.com/ory/kratos/issues/311):

    > This patch adds a module which is capable of aborting a request, waiting for
    > another option to complete, and then resuming the request again.
    > 
    > This feature makes use of a temporary cookie which keeps track of the
    > request state.
    > 
    > This feature is required for several workflows that update privileged
    > fields such as passwords, 2fa recovery codes, email addresses.
    > 
    > refactor: rename profile to settings flow
    > 
    > Renames selfservice/profile to settings. The settings flow includes a strategy for managing profile information
* **login:** Forced reauthentication ([#248](https://github.com/ory/kratos/issues/248)) ([344fc9c](https://github.com/ory/kratos/commit/344fc9cddccff958f13249b999a835d3e46a7771)), closes [#243](https://github.com/ory/kratos/issues/243)


### Unclassified

* u ([0b6fa48](https://github.com/ory/kratos/commit/0b6fa48e90fa0c50b9c26bae034eb1662c855d69))
* Make format ([b85e5af](https://github.com/ory/kratos/commit/b85e5af2e29f9ca3bc3341ba4f2b1b338b441398))
* u ([03fa4f0](https://github.com/ory/kratos/commit/03fa4f05363aa1f38fe45730317375ce380cfa31))
* u ([a3dfd9d](https://github.com/ory/kratos/commit/a3dfd9d15e1f7287558b85c3a4f23d02444b0bf4))
* u ([616aa0f](https://github.com/ory/kratos/commit/616aa0f0cf3d662b48fcaa02715e02e854e05581))
* fix:add graceful shutdown to courier handler (#296) ([235d784](https://github.com/ory/kratos/commit/235d784b7f8bf38859d15d68c37b089fc9371195)), closes [#296](https://github.com/ory/kratos/issues/296) [#295](https://github.com/ory/kratos/issues/295):

    > Courier would not stop with the provided Background handler.
    > This changes the methods of Courier so that the graceful package can be
    > used in the same way as the http endpoints can be used.
* fix(sql) change courier body to text field (#276) ([ed5268d](https://github.com/ory/kratos/commit/ed5268d539b2a28f5367e8ba2e2e6bd3a605ce5b)), closes [#276](https://github.com/ory/kratos/issues/276) [#269](https://github.com/ory/kratos/issues/269)


### BREAKING CHANGES

* Please remove the `redirect` hook from both login,
registration, and settings after configuration. Please remove
the `session` hook from your login after configuration. Hooks
have moved down a level and are now configured at
`selfservice.<login|registration|settings>.<after|before>.hooks`
instead of
`selfservice.<login|registration|settings>.<after|before>.hooks`.
Hooks are now identified by `hook:` instead of `job:`. Please
rename those sections accordingly.
* **continuity:** Several profile-related URLs have and payloads been updated. Please consult the most recent documentation.
* The payloads of the Profile Management Request API
that previously were set in `{ "methods": { "traits": { ... } }}` have now moved to
`{ "methods": { "profile": { ... } }}`.
* This patch introduces a refactor that is needed
for the profile management API to be capable of handling (password,
oidc, ...) credential changes as well.

To implement this, the payloads of the Profile Management Request API
that previously were set in `{"form": {...} }` have now moved to
`{"methods": { "traits": { ... } }}`.

In the future, as more credential updates are handled, there will
be additional keys in the forms key
`{"methods": { "traits": { ... }, "password": { ... } }}`.



## [0.1.1-alpha.1](https://github.com/ory/kratos/compare/v0.1.0-alpha.6...v0.1.1-alpha.1) (2020-02-18)


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


### Documentation

* Regenerate and update changelog ([e87e9c9](https://github.com/ory/kratos/commit/e87e9c9ec9cf55351439ab16a778f3ea303ec646))
* Regenerate and update changelog ([d6f0794](https://github.com/ory/kratos/commit/d6f0794d53b6e7d6d9e3bc63a77d402e43a29bed))
* Regenerate and update changelog ([eb7326c](https://github.com/ory/kratos/commit/eb7326c98c2d5e87a8ac3cd9f2efb43f2552164a))


### Features

* Redirect to new auth session on expired auth sessions ([#230](https://github.com/ory/kratos/issues/230)) ([b477ecd](https://github.com/ory/kratos/commit/b477ecd47de33a9a45159a298ac288c4ad5a0b55)), closes [#96](https://github.com/ory/kratos/issues/96)



# [0.1.0-alpha.4](https://github.com/ory/kratos/compare/v0.1.0-alpha.3...v0.1.0-alpha.4) (2020-02-06)


### Documentation

* Regenerate and update changelog ([f02afb3](https://github.com/ory/kratos/commit/f02afb3fed310f7fe9c5e6f7df34dfc9738018ad))



# [0.1.0-alpha.3](https://github.com/ory/kratos/compare/v0.1.0-alpha.2...v0.1.0-alpha.3) (2020-02-06)

No significant changes have been made for this release.


# [0.1.0-alpha.2](https://github.com/ory/kratos/compare/v0.1.0-alpha.1...v0.1.0-alpha.2) (2020-02-03)


### Bug Fixes

* **daemon:** Register error routes on admin port ([#226](https://github.com/ory/kratos/issues/226)) ([decd8d8](https://github.com/ory/kratos/commit/decd8d8ef8dac3674938b564962238195ffaf017))
* Add paths to sqa middleware ([#216](https://github.com/ory/kratos/issues/216)) ([130c9c2](https://github.com/ory/kratos/commit/130c9c242e1434074d9fa4970b60ccb9b4f2ff47))
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

    > This patch refactors user-facing error APIs:
    > 
    > - The `/errors` endpoint moved to `/self-service/errors`
    > - The endpoint is now available at both the Admin and Public API. The Public API requires CSRF Token match or a 403 error will be returned.
    > - The Public API endpoint no longer returns 404 errors but 403 instead.
    > - The response payload changed. What was `[{"code": ...}]` is now `{"id": "...", "errors": [{"code": ...}]}`
    > 
    > This patch requires running `kratos migrate sql` as a new column (`csrf_token`) has been added to the user-facing error store.
* Update CHANGELOG [ci skip] ([c368a11](https://github.com/ory/kratos/commit/c368a11523a9bcb30a830d65c11e4f6d27417a78))



# [0.1.0-alpha.1](https://github.com/ory/kratos/compare/v0.0.3-alpha.15...v0.1.0-alpha.1) (2020-01-31)


### Documentation

* Updates issue and pull request templates ([#215](https://github.com/ory/kratos/issues/215)) ([10c45f2](https://github.com/ory/kratos/commit/10c45f23e11abba1ca82095548769cd923a6a6a6))



## [0.0.3-alpha.15](https://github.com/ory/kratos/compare/v0.0.3-alpha.14...v0.0.3-alpha.15) (2020-01-31)


### Unclassified

* Update permissions in SQLite Dockerfile ([1266e53](https://github.com/ory/kratos/commit/1266e533ac9a1f6ec375980cadce9755998f9fe6))



## [0.0.3-alpha.14](https://github.com/ory/kratos/compare/v0.0.3-alpha.13...v0.0.3-alpha.14) (2020-01-31)


### Unclassified

* Update README.md ([db8d65b](https://github.com/ory/kratos/commit/db8d65bf136223df546aa27f1ecff03d01159624))



## [0.0.3-alpha.13](https://github.com/ory/kratos/compare/v0.0.3-alpha.12...v0.0.3-alpha.13) (2020-01-31)


### Unclassified

* Allow mounting SQLite in /home/ory/sqlite (#212) ([2fe8c0f](https://github.com/ory/kratos/commit/2fe8c0f752e870028d68e8593a46c0902f673a65)), closes [#212](https://github.com/ory/kratos/issues/212)



## [0.0.3-alpha.11](https://github.com/ory/kratos/compare/v0.0.3-alpha.10...v0.0.3-alpha.11) (2020-01-31)


### Unclassified

* Clean up cmd and resolve packr2 issues (#211) ([2e43ec0](https://github.com/ory/kratos/commit/2e43ec09e9d6aa572c4351bfef4c59dfc43f2343)), closes [#211](https://github.com/ory/kratos/issues/211):

    > This patch addresses issues with the build pipeline caused by an invalid import. Profiling was also added.
* Improve field types (#209) ([aeefa93](https://github.com/ory/kratos/commit/aeefa93bf0427685f6ffadad5abfaa1fc26ce074)), closes [#209](https://github.com/ory/kratos/issues/209)
* Update CHANGELOG [ci skip] ([fc32207](https://github.com/ory/kratos/commit/fc32207482861b8f989cb1d6fe5d96bf34c54e4c))



## [0.0.3-alpha.10](https://github.com/ory/kratos/compare/v0.0.3-alpha.9...v0.0.3-alpha.10) (2020-01-31)


### Unclassified

* Update README ([35a310d](https://github.com/ory/kratos/commit/35a310d6de52fa74ad8728b1df67f88ce900aa61))
* Update CHANGELOG [ci skip] ([3c98745](https://github.com/ory/kratos/commit/3c987455a44b9e12e31619ba9f447e8a5feafc38))
* Update CHANGELOG [ci skip] ([c1c01df](https://github.com/ory/kratos/commit/c1c01df3a04fc7988bf847e3f31680112f5a642d))



## [0.0.3-alpha.7](https://github.com/ory/kratos/compare/v0.0.3-alpha.5...v0.0.3-alpha.7) (2020-01-30)


### Unclassified

* Use correct project root in Dockerfile ([3528758](https://github.com/ory/kratos/commit/352875878c74d15b522336b518df339c8ad48e49))
* Update CHANGELOG [ci skip] ([e78bbbe](https://github.com/ory/kratos/commit/e78bbbecbd9515c02e447efc3208599bf27ef85c))



## [0.0.3-alpha.5](https://github.com/ory/kratos/compare/v0.0.3-alpha.4...v0.0.3-alpha.5) (2020-01-30)


### Unclassified

* Update CHANGELOG [ci skip] ([ebb1744](https://github.com/ory/kratos/commit/ebb1744d68b8a416774477182b1e2b2cd8bdfc43))
* Add libmusl to binary output ([e9b8445](https://github.com/ory/kratos/commit/e9b8445f2fc8e9e571ec0b8480cc70fe3251db9e))



## [0.0.3-alpha.4](https://github.com/ory/kratos/compare/v0.0.3-alpha.3...v0.0.3-alpha.4) (2020-01-30)


### Unclassified

* Update CHANGELOG [ci skip] ([018c229](https://github.com/ory/kratos/commit/018c229c4cff62e47c1154ca29ab9c70766a43e5))
* Add and use ory docker user ([cccbe09](https://github.com/ory/kratos/commit/cccbe09cc6e2ad72847206d46afe3e0bf7f79ab5))
* Update CHANGELOG [ci skip] ([0e436e5](https://github.com/ory/kratos/commit/0e436e57f79692c4c6e0a0c25f48a41654afcda1))
* Update goreleaser changelog filters ([7e5af97](https://github.com/ory/kratos/commit/7e5af97fded9f56a3cc9d1d92a7726e7b613b586))
* Update CHANGELOG [ci skip] ([4387503](https://github.com/ory/kratos/commit/438750326c5d6ad1569802c82806e831f43e785e))



## [0.0.3-alpha.2](https://github.com/ory/kratos/compare/v0.0.3-alpha.1...v0.0.3-alpha.2) (2020-01-30)


### Unclassified

* Resolve goreleaser build issues (#208) ([d59a08a](https://github.com/ory/kratos/commit/d59a08a0ef680a984352d7f5068626cc1958185a)), closes [#208](https://github.com/ory/kratos/issues/208)



## [0.0.3-alpha.1](https://github.com/ory/kratos/compare/v0.0.1-alpha.9...v0.0.3-alpha.1) (2020-01-30)


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



## [0.0.1-alpha.9](https://github.com/ory/kratos/compare/v0.0.1-alpha.11...v0.0.1-alpha.9) (2020-01-29)

No significant changes have been made for this release.


## [0.0.2-alpha.1](https://github.com/ory/kratos/compare/v0.0.1-alpha.8...v0.0.2-alpha.1) (2020-01-29)


### Unclassified

* Use correct build archive for homebrew ([74ac29f](https://github.com/ory/kratos/commit/74ac29f43f2937cad9065ad3c03cf3cf909cff42))



## [0.0.1-alpha.6](https://github.com/ory/kratos/compare/v0.0.1-alpha.5...v0.0.1-alpha.6) (2020-01-29)

No significant changes have been made for this release.


## [0.0.1-alpha.5](https://github.com/ory/kratos/compare/v0.0.1-alpha.3...v0.0.1-alpha.5) (2020-01-29)


### Unclassified

* Resolve build issues with CGO (#196) ([298f4ea](https://github.com/ory/kratos/commit/298f4ea85b3e7405929f481b756efe8c5c133479)), closes [#196](https://github.com/ory/kratos/issues/196)
* ss/password: Make form fields an array (#197) ([6cb0058](https://github.com/ory/kratos/commit/6cb005860755ff897ad847f09af50bc911bbc7f0)), closes [#197](https://github.com/ory/kratos/issues/197) [#186](https://github.com/ory/kratos/issues/186)



## [0.0.1-alpha.3](https://github.com/ory/kratos/compare/v0.0.1-alpha.2...v0.0.1-alpha.3) (2020-01-28)

No significant changes have been made for this release.


## [0.0.1-alpha.2](https://github.com/ory/kratos/compare/v0.0.1-alpha.1...v0.0.1-alpha.2) (2020-01-28)

No significant changes have been made for this release.


## [0.0.1-alpha.1](https://github.com/ory/kratos/compare/ab6f24a85276bdd8687f2fc06390c1279892b005...v0.0.1-alpha.1) (2020-01-28)


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

* Inject Identity Traits JSON Schema ([3a4c5ad](https://github.com/ory/kratos/commit/3a4c5ad35f885c7d38ffcf1d5836fb485f122fe9)), closes [#189](https://github.com/ory/kratos/issues/189)
* Remove redundant return statement ([7c2989f](https://github.com/ory/kratos/commit/7c2989f52c090bb9900380b4ec74e04d9c37a441))
* ss/oidc: Remove obsolete request field from form (#193) ([59671ba](https://github.com/ory/kratos/commit/59671badb63009e2440b14868b622adc75cf882f)), closes [#193](https://github.com/ory/kratos/issues/193) [#180](https://github.com/ory/kratos/issues/180)
* Re-introduce migration plans to CLI command ([#192](https://github.com/ory/kratos/issues/192)) ([bb32cd3](https://github.com/ory/kratos/commit/bb32cd3cad3cd0bd6f3166de0166701e1f676ac6)), closes [#131](https://github.com/ory/kratos/issues/131)
* strategy/oidc: Allow multiple OIDC Connections (#191) ([8984831](https://github.com/ory/kratos/commit/898483137ff9dc47d65750cd94a973f2e5bee770)), closes [#191](https://github.com/ory/kratos/issues/191) [#114](https://github.com/ory/kratos/issues/114)
* Improve Docker Compose Quickstart (#187) ([9459072](https://github.com/ory/kratos/commit/945907297ded4b18e1bd0e7c9824a975ac7395c6)), closes [#187](https://github.com/ory/kratos/issues/187) [#188](https://github.com/ory/kratos/issues/188)
* Fix broken import ([308aa13](https://github.com/ory/kratos/commit/308aa1334dd43bc4bebade4e70e9c81c83fe8806))
* selfservice/password: Remove request field and ensure method is set (#183) ([e035adc](https://github.com/ory/kratos/commit/e035adc233198e9b5c9a6e08d442fb5fb3290816)), closes [#183](https://github.com/ory/kratos/issues/183)
* Add tests and fixtures for the config JSON Schema (#171) ([ede9c0e](https://github.com/ory/kratos/commit/ede9c0e9c45ee91e60587311dc18a0a04ff62295)), closes [#171](https://github.com/ory/kratos/issues/171)
* Add example values for config JSON Schema ([12ba728](https://github.com/ory/kratos/commit/12ba7283bf879cd7682d3017c3b3f12e49029d6b))
* Replace `url` with `uri` format in config JSON Schema ([68eddef](https://github.com/ory/kratos/commit/68eddef0cf179bf61abb999d84d2af19c3703c80))
* Replace number with integer in config JSON Schema (#177) ([9eff6fd](https://github.com/ory/kratos/commit/9eff6fd09720b11acae089ebfcaf37288bc031b0)), closes [#177](https://github.com/ory/kratos/issues/177)
* Improve `--dev` flag (#167) ([9b61ee1](https://github.com/ory/kratos/commit/9b61ee10bbb4710d6694addfa60c04313855516f)), closes [#167](https://github.com/ory/kratos/issues/167) [#162](https://github.com/ory/kratos/issues/162)
* Add goreleaser orb task (#170) ([5df0def](https://github.com/ory/kratos/commit/5df0defefc95ced289a9c59a4f5deb3c67446e75)), closes [#170](https://github.com/ory/kratos/issues/170)
* Add changelog generation task (#169) ([edd937c](https://github.com/ory/kratos/commit/edd937c21b7e37b2f2e926f0fe62c2e7d4a7d608)), closes [#169](https://github.com/ory/kratos/issues/169)
* Adopt new SDK pipeline (#168) ([21d9b6d](https://github.com/ory/kratos/commit/21d9b6d27adbfe8504fb46ac95952e7cea239085)), closes [#168](https://github.com/ory/kratos/issues/168)
* Add ability to define multiple schemas and serve them over HTTP ([#164](https://github.com/ory/kratos/issues/164)) ([c65119c](https://github.com/ory/kratos/commit/c65119c24378dabd306e5a49f89c28c0367f7c2e)), closes [#86](https://github.com/ory/kratos/issues/86):

    > All identity traits schemas have to be configured using a human readable ID and the corresponding URL. This PR enables multiple schemas to be used next to the default schema.
    > It also adds the kratos.public/schemas/:id endpoint that mirrors all schemas.
* Add docker-compose quickstart (#153) ([e096190](https://github.com/ory/kratos/commit/e096190e778f22573e30f35e85b7cf147caf851b)), closes [#153](https://github.com/ory/kratos/issues/153)
* Update README (#160) ([533775b](https://github.com/ory/kratos/commit/533775ba78a2c1758c47ed093da6acc18ab951c2)), closes [#160](https://github.com/ory/kratos/issues/160)
* Separate post register/login hooks (#150) ([f4b7812](https://github.com/ory/kratos/commit/f4b78122d9cbe4dcc05b4fd52d94a2d9f1b16eb2)), closes [#150](https://github.com/ory/kratos/issues/150) [#149](https://github.com/ory/kratos/issues/149)
* Update README badges ([4f7838e](https://github.com/ory/kratos/commit/4f7838e69181c5a10e27cde1e241779e4e724909))
* Bump go-acc and resolve test issues (#154) ([15b1b63](https://github.com/ory/kratos/commit/15b1b630c5363e0e1afbed53285b3f39098c0792)), closes [#154](https://github.com/ory/kratos/issues/154) [#152](https://github.com/ory/kratos/issues/152) [#151](https://github.com/ory/kratos/issues/151):

    > Due to a bug in `go-acc`, tests would not run if `-tags sqlite` was supplied as a go tool argument to `go-acc`. This patch resolves that issue and also includes several test patches from previous community PRs and some internal test issues.
* Add helper for requiring authentication ([3888fbd](https://github.com/ory/kratos/commit/3888fbdc239b7a06c7fca34d08de7d55af69a48c))
* Add session destroyer hook  ([#148](https://github.com/ory/kratos/issues/148)) ([d17f002](https://github.com/ory/kratos/commit/d17f002cdfe1f11ebb6bcbb17f6976aa329eab4a)), closes [#139](https://github.com/ory/kratos/issues/139):

    > This patch adds a hook that destroys all active session by the identity which is being logged in. This can be useful in scenarios where only one session should be active at any given time.
* Add ORY Kratos banner to README (#145) ([23b824f](https://github.com/ory/kratos/commit/23b824f7f99efbc23787508c03506e73a3240a2a)), closes [#145](https://github.com/ory/kratos/issues/145)
* Implement message templates and SMTP delivery ([#146](https://github.com/ory/kratos/issues/146)) ([dc674bf](https://github.com/ory/kratos/commit/dc674bfa7d1fa9ee94b014d09866bbdc0a97c321)), closes [#99](https://github.com/ory/kratos/issues/99):

    > This patch adds a message templates (with override capabilities)
    > and SMTP delivery.
    > 
    > Integration tests using MailHog test fault resilience and e2e email
    > delivery.
    > 
    > This system is designed to be extended for SMS and other use cases.
* Replace DBAL layer with gobuffalo/pop (#130) ([21d08b8](https://github.com/ory/kratos/commit/21d08b84560230d8a063a418a74efcf53c146872)), closes [#130](https://github.com/ory/kratos/issues/130):

    > This is a major refactoring of the internal DBAL. After a successful proof of concept and evaluation of gobuffalo/pop, we believe this to be the best DBAL for Go at the moment. It abstracts a lot of boilerplate code away.
    > 
    > As with all sophisticated DBALs, pop too has its quirks. There are several issues that have been discovered during testing and adoption: https://github.com/gobuffalo/pop/issues/136 https://github.com/gobuffalo/pop/issues/476 https://github.com/gobuffalo/pop/issues/473 https://github.com/gobuffalo/pop/issues/469 https://github.com/gobuffalo/pop/issues/466
    > 
    > However, the upside of moving much of the hard database/sql plumbing into another library cleans up the code base significantly and reduces complexity.
    > 
    > As part of this change, the "ephermal" DBAL ("in memory") will be removed and sqlite will be used instead. This further reduces complexity of the code base and code-duplication.
    > 
    > To support sqlite, CGO is required, which means that we need to run tests with `go test -tags sqlite` on a machine that has g++ installed. This also means that we need a Docker Image with `alpine` as opposed to pure `scratch`. While this is certainly a downside, the upside of less maintenance and "free" support for SQLite, PostgreSQL, MySQL, and CockroachDB simply outweighs any downsides that come with CGO.
* Replace local deps with remote ones ([8605e45](https://github.com/ory/kratos/commit/8605e454cf538e047c5a9c3479372892d6b3f483))
* ss/profile: Improve success and error flows ([9e0015a](https://github.com/ory/kratos/commit/9e0015acec7f8d927498e48366b377e22ec768b7)), closes [#112](https://github.com/ory/kratos/issues/112):

    > This patch completes the profile management flow by implementing proper error and success states and adding several data integrity tests.
* Add helpers for go-swagger ([165a660](https://github.com/ory/kratos/commit/165a660f277588ed572d7843354c207f72f1678d)):

    > See https://github.com/go-swagger/go-swagger/issues/2119
* Add profile management and refactor internals ([3ec9263](https://github.com/ory/kratos/commit/3ec9263f597a5949d0de6d10073cc626cfcfcca4)), closes [#112](https://github.com/ory/kratos/issues/112)
* Update keyword from kratos to ory.sh/kratos ([f45cbe0](https://github.com/ory/kratos/commit/f45cbe0339db8d129522314f3099e6944e4a6ea3)), closes [#115](https://github.com/ory/kratos/issues/115)
* Update sdk generation method ([24aa3d7](https://github.com/ory/kratos/commit/24aa3d73354d5a28f05999a09e7bbbe51a44d44e))
* Use JSON Schema to type assert form body ([#116](https://github.com/ory/kratos/issues/116)) ([1944c7c](https://github.com/ory/kratos/commit/1944c7c6e82b5b6a3b9d47db94c8f8f45248feb7)), closes [#109](https://github.com/ory/kratos/issues/109)
* Rebrand ORY Hive to ORY Kratos (#111) ([ceda7fb](https://github.com/ory/kratos/commit/ceda7fb3472b081f0c6066aa1f282d4ec1787f7b)), closes [#111](https://github.com/ory/kratos/issues/111)
* Add SQL adapter ([#100](https://github.com/ory/kratos/issues/100)) ([9e7f998](https://github.com/ory/kratos/commit/9e7f99871e3f09e7ae9ec1c38c8b8cf94d076f45)), closes [#92](https://github.com/ory/kratos/issues/92)
* Explicitly whitelist form parser keys ([#105](https://github.com/ory/kratos/issues/105)) ([28b056e](https://github.com/ory/kratos/commit/28b056e5bbfec645262914c52f0386d70c787a32)), closes [#98](https://github.com/ory/kratos/issues/98):

    > Previously the form parser would try to detect the field type by
    > asserting types for the whole form. That caused passwords
    > containing only numbers to fail to unmarshal into a string
    > value.
    > 
    > This patch resolves that issue by introducing a prefix
    > option to the BodyParser
* Handle securecookie errors appropriately ([#101](https://github.com/ory/kratos/issues/101)) ([75bf6fe](https://github.com/ory/kratos/commit/75bf6fe3f79d025f2aaa79d06db39c26430dc3fc)), closes [#97](https://github.com/ory/kratos/issues/97):

    > Previously, IsNotAuthenticated would not handle securecookie errors appropriately.
    > This has been resolved.
* Improve migration command ([#94](https://github.com/ory/kratos/issues/94)) ([2b631de](https://github.com/ory/kratos/commit/2b631de6d621dcebac5318f6dd628646fec7712f))
* Mark active field as nullable ([#89](https://github.com/ory/kratos/issues/89)) ([292702d](https://github.com/ory/kratos/commit/292702d9e031e43c63e0ecb59354557139499e87))
* Move package to selfservice ([063b767](https://github.com/ory/kratos/commit/063b7679af76333fc546e94e92b197079e5bdb30)):

    > Because this module is primarily used
    > in selfservice scenarios, it has been
    > moved to the selfservice parent.
* Omit request header from login/registration request ([#106](https://github.com/ory/kratos/issues/106)) ([9b07587](https://github.com/ory/kratos/commit/9b07587f2de2b270c5c326e37b2b6b3dbbfa8595)), closes [#95](https://github.com/ory/kratos/issues/95):

    > When fetching a login and registration request, the HTTP Request Headers
    > must not be included in the response, as they contain irrelevant
    > information for the API caller.
* Properly handle empty credentials config in sql ([#93](https://github.com/ory/kratos/issues/93)) ([b79c5d1](https://github.com/ory/kratos/commit/b79c5d1d5216e994f986ce739285cb1a89523df5))
* Resolve wrong column reference in sql ([#90](https://github.com/ory/kratos/issues/90)) ([0c0eb87](https://github.com/ory/kratos/commit/0c0eb87cd341bd3e73eb9adb303054b38c103ba9)):

    > Reference ic.method instead of ici.method.
    > 
    > Added regression tests against this particular issue.
* Update to ory/x 0.0.80 ([#110](https://github.com/ory/kratos/issues/110)) ([64de2f8](https://github.com/ory/kratos/commit/64de2f86540bf8715a1703d773fa95011603a854)):

    > Removes the need for BindEnv()
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
* Reset CSRF token on principal change ([#64](https://github.com/ory/kratos/issues/64)) ([9c889ab](https://github.com/ory/kratos/commit/9c889ab4f6c846812a4290545fef7d8106da35f0)), closes [#38](https://github.com/ory/kratos/issues/38):

    > Add tests for logout.
* Add tests for selfservice ErrorHandler (#62) ([4bb9e70](https://github.com/ory/kratos/commit/4bb9e7086ee57c4eb1a73fea436c7b2dec0257b7)), closes [#62](https://github.com/ory/kratos/issues/62)
* Implement CRUD for identities ([#60](https://github.com/ory/kratos/issues/60)) ([58a3c24](https://github.com/ory/kratos/commit/58a3c240fca66e1195bf310024a2f8473826bce6)), closes [#58](https://github.com/ory/kratos/issues/58)
* Enable Circle CI (#57) ([6fb0afd](https://github.com/ory/kratos/commit/6fb0afd30e3755026b6ffca0cc80f2fe00267681)), closes [#57](https://github.com/ory/kratos/issues/57) [#53](https://github.com/ory/kratos/issues/53)
* OIDC provider selfservice data enrichment (#56) ([936970a](https://github.com/ory/kratos/commit/936970a9abaadeab5c191ff52218bf4f65af2220)), closes [#56](https://github.com/ory/kratos/issues/56) [#23](https://github.com/ory/kratos/issues/23) [#55](https://github.com/ory/kratos/issues/55)
* Remove local jsonschema module override ([cd2a5d8](https://github.com/ory/kratos/commit/cd2a5d8c74b21b122f5d5437702d8c74fb1cb726))
* Implement identity management, login, and registration (#22) ([bf3395e](https://github.com/ory/kratos/commit/bf3395ea34ecf85303034f3e941a049c8cbd6229)), closes [#22](https://github.com/ory/kratos/issues/22)
* Revert incorrect license changes ([fb9740b](https://github.com/ory/kratos/commit/fb9740b37a94dbdde1a8f4433fb7e5a8b4dac295))
* Create FUNDING.yml ([3c67ac8](https://github.com/ory/kratos/commit/3c67ac83f58c5b03dc3935d279083268b8a85e0d))
* Initial commit ([ab6f24a](https://github.com/ory/kratos/commit/ab6f24a85276bdd8687f2fc06390c1279892b005))



