# CHANGES for jsonpath-plus

## 2.0.0 (November 23, 2019)

- Breaking change: Throw `TypeError` instead of `Error` for missing
    `otherTypeCallback` when using `@other`
- Breaking change: Throw `TypeError` instead of `Error` for missing `path`
- Enhancement: Throw `TypeError` for missing `json` (fixes #110)
- Enhancement: Use more efficient `new Function` over `eval`;
    also allows use of cyclic context objects
- Maintenance: Add `.editorconfig`
- Docs: Document options in jsdoc; add return values to callbacks;
    fix constructor doc sig.
- Testing: Add test for missing `path` or `json`
- Testing: Remove unneeded closures
- npm: Update devDeps

## 1.2.0 (October 13, 2019)

- Enhancement: Add `@root` filter selector
- Enhancement: Use more efficient `new Function` over `eval`;
    also allows use of cyclic context objects
- npm: Update devDeps and `package-lock.json`

## 1.1.0 (September 26, 2019)

- Enhancement: Add explicit 'any' to `evaluate()` declaration (for use
  with `noImplicitAny` TypeScript option)
- Build: Update minified build files
- Travis: Update to check Node 6, 10, 12
- npm: Ignore `.idea`/`.remarkrc` files
- npm: Update devDeps (Babel, linting, Rollup, TypeScript related)
- npm: Avoid eslint script within test script
- npm: Ignore typescript docs

## 1.0.0 (August 7, 2019)

- Add TypeScript declaration

## 0.20.1 (June 12, 2019)

- npm: Avoid adding `core-js-bundle` as peerDep. (fixes #95)

## 0.20.0 (June 4, 2019)

- Build: Add `browserslist` for Babel builds
- Linting: Conform to ESLint updates (jsdoc)
- Testing: Switch from end-of-lifed nodeunit to Mocha
- Testing: Add performance test to browser, but bump duration
- npm: Update devDeps; add core-js-bundle to peerDependencies
- npm: Ignore some unneeded files
- Bump Node version in Travis to avoid erring with object rest
    in eslint-plugin-node routine

## 0.19.0 (May 16, 2019)

- Docs (README): Indicate features, including performance (removing old note)
- Docs (README): Add headings for setup and fix headings levels
- Docs (README): Indicate parent selector was not present in original spec
    (not just not documented)
- Docs (README): Fix escaping
- Linting: Switch to Unix line breaks and other changes per ash-nazg, including linting Markdown JS
- Linting: Use recommended `.json` extension
- Linting: Switch to ash-nazg
- Linting: Add lgtm.yml file for lgtm.com
- npm: Update devDeps, and update per security audit

## 0.18.1 (May 14, 2019)

- Fix: Expose `pointer` on `resultType: "all"`

## 0.18.0 (October 20, 2018)

- Security enhancement: Use global eval instead of regular eval
- Fix: Handle React-Native environment's lack of support for
    Node vm (@simon-scherzinger); closes #87
- Refactoring: Use arrow functions, for-of, declare block scope vars
    closer to block
- Docs: Clarify current `wrap` behavior
- npm: Add Rollup to test scripts

## 0.17.0 (October 19, 2018)

- Breaking change: With Node use, must now use
    `require('jsonpath-plus').JSONPath`.
- Breaking change: Stop including polyfills for array and string `includes`
    (can get with `@babel/polyfill` or own)
- Breaking change: Remove deprecated `JSONPath.eval`
- License: Remove old and unneeded license portion from within source file
    (already have external file)
- Fix: Support object shorthand functions on sandbox objects
    (`toString()` had not been working properly with them)
- Enhancement: Add Rollup/Babel/Terser and `module` in `package.json`
- Refactoring: Use ES6 features such as object shorthand
- Linting: prefer const and no var
- Testing: Replace custom server code with `node-static` and add `opn-cli`;
    mostly switch to ESM
- npm: Update devDeps; add `package-lock.json`; remove non-functioning remark

## 0.16.0 (January 14, 2017)

- Breaking change: Give preference to treating special chars in a property
    as special (override with backtick operator)
- Breaking feature: Add custom \` operator to allow unambiguous literal
    sequences (if an initial backtick is needed, an additional one must
    now be added)
- Fix: `toPathArray` caching bug
- Improvements: Performance optimizations
- Dev testing: Rename test file

## 0.15.0 (Mar 15, 2016)

- Fix: Fixing support for sandbox in the case of functions
- Feature: Use `this` if present for global export
- Docs: Clarify function signature
- Docs: Update testing section
- Dev testing: Add in missing test for browser testing
- Dev testing: Add remark linting to testing process (#70)
- Dev testing: Lint JS test support files
- Dev testing: Split out tests into `eslint`, `remark`, `lint`, `nodeunit`
- Dev testing: Remove need for nodeunit build step
- Dev testing: Simplify nodeunit usage and make available
  as `npm run browser-test`

## 0.14.0 (Jan 10, 2016)

- Feature: Add `@scalar()` type operator (in JavaScript mode, will also
    include)

## 0.13.1 (Jan 5, 2016)

- Fix: Avoid double-encoding path in results

## 0.13.0 (Dec 13, 2015)

- Breaking change (from version 0.11): Silently strip `~` and `^` operators
  and type operators such as `@string()` in `JSONPath.toPathString()` calls.
- Breaking change: Remove `Array.isArray` polyfill as no longer
  supporting IE <= 8
- Feature: Allow omission of options first argument to `JSONPath`
- Feature: Add `JSONPath.toPointer()` and "pointer" `resultType` option.
- Fix: Correctly support `callback` and `otherTypeCallback` as numbered
  arguments to `JSONPath`.
- Fix: Enhance Node checking to avoid issue reported with angular-mock
- Fix: Allow for `@` or other special characters in at-sign-prefixed
  property names (by use of `[?(@['...'])]` or  `[(@['...'])]`).

## 0.12.0 (Dec 12, 2015 10:39pm)

- Breaking change: Problems with upper-case letters in npm is causing
  us to rename the package, so have renamed package to "jsonpath-plus"
  (there are already package with lower-case "jsonpath" or "json-path").
  The new name also reflects that there have been changes to the
  original spec.

## 0.11.2 (Dec 12, 2015 10:36pm)

- Docs: Actually add the warning in the README that problems in npm
  with upper-case letters is causing us to rename to "jsonpath-plus"
  (next version will actually apply the change).

## 0.11.1 (Dec 12, 2015 10:11pm)

- Docs: Give warning in README that problems in npm with upper-case letters
  is causing us to rename to "jsonpath-plus" (next version will actually
  apply the change).

## 0.11.0 (Dec 12, 2015)

- Breaking change: For unwrapped results, return `undefined` instead
  of `false` upon failure to find path (to allow distinguishing of
  `undefined`--a non-allowed JSON value--from the valid JSON values,
  `null` or `false`) and return the exact value upon falsy single
  results (in order to allow return of `null`)
- Deprecated: Use of `jsonPath.eval()`; use new class-based API instead
- Feature: AMD export
- Feature: By using `self` instead of `window` export, allow JSONPath
  to be trivially imported into web workers, without breaking
  compatibility in normal scenarios. See [MDN on self](https://developer.mozilla.org/en-US/docs/Web/API/Window/self)
- Feature: Offer new class-based API and object-based arguments (with
  option to run new queries via `evaluate()` method without resupplying config)
- Feature: Allow new `preventEval=true` and `autostart=false` option
- Feature: Allow new callback option to allow a callback function to execute as
  each final result node is obtained
- Feature: Allow type operators: JavaScript types (`@boolean()`, `@number()`,
  `@string()`), other fundamental JavaScript types (`@null()`, `@object()`,
  `@array()`), the JSONSchema-added type, `@integer()`, and the following
  non-JSON types that can nevertheless be used with JSONPath when querying
  non-JSON JavaScript objects (`@undefined()`, `@function()`, `@nonFinite()`).
  Finally, `@other()` is made available in conjunction with a new callback
  option, `otherTypeCallback`, can be used to allow user-defined type
  detection (at least until JSON Schema awareness may be provided).
- Feature: Support "parent" and "parentProperty" for resultType along with
  "all" (which also includes "path" and "value" together)
- Feature: Support custom `@parent`, `@parentProperty`, `@property` (in
  addition to custom property `@path`) inside evaluations
- Feature: Support a custom operator (`~`) to allow grabbing of property names
- Feature: Support `$` for retrieval of root, and document this as well as
  `$..` behavior
- Feature: Expose cache on `JSONPath.cache` for those who wish to preserve and
  reuse it
- Feature: Expose class methods `toPathString` for converting a path as array
  into a (normalized) path as string and `toPathArray` for the reverse (though
  accepting unnormalized strings as well as normalized)
- Fix: Allow `^` as property name
- Fix: Support `.` within properties
- Fix: `@path` in index/property evaluations

## 0.10.0 (Oct 23, 2013)

- Feature: Support for parent selection via `^`
- Feature: Access current path via `@path` in test statements
- Feature: Allowing for multi-statement evals
- Improvements: Performance

## 0.9.0 (Mar 28, 2012)

- Feature: Support a sandbox arg to eval
- Improvements: Use `vm.runInNewContext` in place of eval
