Change Log
====================================================================================================
All notable changes will be documented in this file.
JSON Schema $Ref Parser adheres to [Semantic Versioning](http://semver.org/).



[v8.0.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v8.0.0) (2020-03-13)
----------------------------------------------------------------------------------------------------

- Moved JSON Schema $Ref Parser to the [@APIDevTools scope](https://www.npmjs.com/org/apidevtools) on NPM

- The "json-schema-ref-parser" NPM package is now just a wrapper around the scoped "@apidevtools/json-schema-ref-parser" package

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v7.1.4...v8.0.0)



[v7.1.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v7.1.0) (2019-06-21)
----------------------------------------------------------------------------------------------------

- Merged [PR #124](https://github.com/APIDevTools/json-schema-ref-parser/pull/124/), which provides more context to [custom resolvers](https://apitools.dev/json-schema-ref-parser/docs/plugins/resolvers.html).

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v7.0.1...v7.1.0)



[v7.0.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v7.0.0) (2019-06-11)
----------------------------------------------------------------------------------------------------

- Dropped support for Node 6

- Updated all code to ES6+ syntax (async/await, template literals, arrow functions, etc.)

- No longer including a pre-built bundle in the package. such as [Webpack](https://webpack.js.org/), [Rollup](https://rollupjs.org/), [Parcel](https://parceljs.org/), or [Browserify](http://browserify.org/) to include JSON Schema $Ref Parser in your app

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v6.1.0...v7.0.0)



[v6.0.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v6.0.0) (2018-10-04)
----------------------------------------------------------------------------------------------------

- Dropped support for [Bower](https://www.npmjs.com/package/bower), since it has been deprecated

- Removed the [`debug`](https://npmjs.com/package/debug) dependency

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v5.1.0...v6.0.0)



[v5.1.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v5.1.0) (2018-07-11)
----------------------------------------------------------------------------------------------------

- Improved the logic of the [`bundle()` method](https://github.com/APIDevTools/json-schema-ref-parser/blob/master/docs/ref-parser.md#bundleschema-options-callback) to produce shorter reference paths when possible.  This is not a breaking change, since both the old reference paths and the new reference paths are valid.  The new ones are just shorter.  Big thanks to [@hipstersmoothie](https://github.com/hipstersmoothie) for [PR #68](https://github.com/APIDevTools/json-schema-ref-parser/pull/68), which helped a lot with this.

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v5.0.0...v5.1.0)



[v5.0.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v5.0.0) (2018-03-18)
----------------------------------------------------------------------------------------------------

This release contains two bug fixes related to file paths.  They are _technically_ breaking changes &mdash; hence the major version bump &mdash; but they're both edge cases that probably won't affect most users.

- Fixed a bug in the [`$refs.paths()`](docs/refs.md#pathstypes) and [`$refs.values()`](docs/refs.md#valuestypes) methods that caused the path of the root schema file to always be represented as a URL, rather than a filesystem path (see [this commit](https://github.com/APIDevTools/json-schema-ref-parser/commit/a95cf50fdf16c864cc1c18d2021d9ce3ec35f5de))

- Merged [PR #75](https://github.com/APIDevTools/json-schema-ref-parser/pull/75), which resolves [issue #76](https://github.com/APIDevTools/swagger-parser/issues/76).  Error messages no longer include the current working directory path when there is no file path.

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v4.1.1...v5.0.0)



[v4.1.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v4.1.0) (2018-01-17)
----------------------------------------------------------------------------------------------------

- Updated dependencies

- Improved the `bundle()` algorithm to favor direct references rather than indirect references (see [PR #62](https://github.com/APIDevTools/json-schema-ref-parser/pull/62) for details).  This will produce different bundled output than previous versions for some schemas. Both the old output and the new output are valid, but the new output is arguably better.

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v4.0.0...v4.1.0)



[v4.0.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v4.0.0) (2017-10-13)
----------------------------------------------------------------------------------------------------

#### Breaking Changes

- To reduce the size of this library, it no longer includes polyfills for [Promises](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise) and [TypedArrays](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/TypedArray), which are natively supported in the latest versions of Node and web browsers.  If you need to support older browsers (such as IE9), then just use [this `Promise` polyfill](https://github.com/stefanpenner/es6-promise) and [this `TypedArray` polyfill](https://github.com/inexorabletash/polyfill/blob/master/typedarray.js).

#### Minor Changes

- Updated dependencies

- [PR #53](https://github.com/APIDevTools/json-schema-ref-parser/pull/53) - Fixes [an edge-case bug](https://github.com/APIDevTools/json-schema-ref-parser/issues/52) with the `bundle()` method

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v3.3.0...v4.0.0)



[v3.3.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v3.3.0) (2017-08-09)
----------------------------------------------------------------------------------------------------

- Updated dependencies

- [PR #30](https://github.com/APIDevTools/json-schema-ref-parser/pull/30) - Added a `browser` field to the `package.json` file to support bundlers such as Browserify, Rollup, and Webpack

- [PR #45](https://github.com/APIDevTools/json-schema-ref-parser/pull/45) - Implemented a temporary workaround for [issue #42](https://github.com/APIDevTools/json-schema-ref-parser/issues/42). JSON Schema $Ref Parser does _not_ currently support [named internal references](http://json-schema.org/latest/json-schema-core.html#id-keyword), but support will be added in the next major release.

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v3.0.0...v3.3.0)



[v3.0.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v3.0.0) (2016-04-03)
----------------------------------------------------------------------------------------------------

#### Plug-ins !!!
That's the major new feature in this version. Originally requested in [PR #8](https://github.com/APIDevTools/json-schema-ref-parser/pull/8), and refined a few times over the past few months, the plug-in API is now finalized and ready to use. You can now define your own [resolvers](https://github.com/APIDevTools/json-schema-ref-parser/blob/v3.0.0/docs/plugins/resolvers.md) and [parsers](https://github.com/APIDevTools/json-schema-ref-parser/blob/v3.0.0/docs/plugins/parsers.md).

#### Breaking Changes
The available [options have changed](https://github.com/APIDevTools/json-schema-ref-parser/blob/v3.0.0/docs/options.md), mostly due to the new plug-in API.  There's not a one-to-one mapping of old options to new options, so you'll have to read the docs and determine which options you need to set. If any. The out-of-the-box configuration works for most people.

All of the [caching options have been removed](https://github.com/APIDevTools/json-schema-ref-parser/commit/1f4260184bfd370e9cd385b523fb08c098fac6db). Instead, all files are now cached, and the entire cache is reset for each new parse operation. Caching options may come back in a future release, if there is enough demand for it. If you used the old caching options, please open an issue and explain your use-case and requirements.  I need a better understanding of what caching functionality is actually needed by users.

#### Bug Fixes
Lots of little bug fixes.  The only major bug fix is to [support root-level `$ref`s](https://github.com/APIDevTools/json-schema-ref-parser/issues/16)


[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v2.2.0...v3.0.0)



[v2.2.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v2.2.0) (2016-01-03)
----------------------------------------------------------------------------------------------------

This version includes a **complete rewrite** of the [`bundle` method](https://github.com/APIDevTools/json-schema-ref-parser/blob/master/docs/ref-parser.md#bundleschema-options-callback) method, mostly to fix [this bug](https://github.com/APIDevTools/swagger-parser/issues/16), but also to address a few [edge-cases](https://github.com/APIDevTools/json-schema-ref-parser/commit/ca9b322879519e4bcb2dcf6e63f08ac254b90868) that weren't handled before.  As a side-effect of this rewrite, there was also some pretty significant refactoring and code-cleanup done throughout the codebase.

Despite the significant code changes, there were no changes to any public-facing APIs, and [all tests are passing](https://apitools.dev/json-schema-ref-parser/test/) as expected.

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v2.1.0...v2.2.0)



[v2.1.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v2.1.0) (2015-12-31)
----------------------------------------------------------------------------------------------------

JSON Schema $Ref Parser now automatically follows HTTP redirects. This is especially great for servers that automatically "ugrade" your connection from HTTP to HTTPS via a 301 redirect. Now that won't break your code.

There are a few [new options](https://github.com/APIDevTools/json-schema-ref-parser/blob/master/docs/options.md) that allow you to set the number of redirects (default is 5) and a few other HTTP request properties.

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v2.0.0...v2.1.0)



[v2.0.0](https://github.com/APIDevTools/json-schema-ref-parser/tree/v2.0.0) (2015-12-31)
----------------------------------------------------------------------------------------------------

Bumping the major version number because [this change](https://github.com/APIDevTools/json-schema-ref-parser/pull/5) technically breaks backward-compatibility &mdash; although I doubt it will actually affect many people.  Basically, if you're using JSON Schema $Ref Parser to download files from a CORS-enabled server that requires authentication, then you'll need to set the `http.withCredentials` option to `true`.

```javascript
$RefParser.dereference('http://some.server.com/file.json', {
    http: { withCredentials: true }
});
```

[Full Changelog](https://github.com/APIDevTools/json-schema-ref-parser/compare/v1.4.1...v2.0.0)
