"use strict";

const $Ref = require("./ref");
const Pointer = require("./pointer");
const parse = require("./parse");
const url = require("./util/url");

module.exports = resolveExternal;

/**
 * Crawls the JSON schema, finds all external JSON references, and resolves their values.
 * This method does not mutate the JSON schema. The resolved values are added to {@link $RefParser#$refs}.
 *
 * NOTE: We only care about EXTERNAL references here. INTERNAL references are only relevant when dereferencing.
 *
 * @param {$RefParser} parser
 * @param {$RefParserOptions} options
 *
 * @returns {Promise}
 * The promise resolves once all JSON references in the schema have been resolved,
 * including nested references that are contained in externally-referenced files.
 */
function resolveExternal (parser, options) {
  if (!options.resolve.external) {
    // Nothing to resolve, so exit early
    return Promise.resolve();
  }

  try {
    // console.log('Resolving $ref pointers in %s', parser.$refs._root$Ref.path);
    let promises = crawl(parser.schema, parser.$refs._root$Ref.path + "#", parser.$refs, options);
    return Promise.all(promises);
  }
  catch (e) {
    return Promise.reject(e);
  }
}

/**
 * Recursively crawls the given value, and resolves any external JSON references.
 *
 * @param {*} obj - The value to crawl. If it's not an object or array, it will be ignored.
 * @param {string} path - The full path of `obj`, possibly with a JSON Pointer in the hash
 * @param {$Refs} $refs
 * @param {$RefParserOptions} options
 *
 * @returns {Promise[]}
 * Returns an array of promises. There will be one promise for each JSON reference in `obj`.
 * If `obj` does not contain any JSON references, then the array will be empty.
 * If any of the JSON references point to files that contain additional JSON references,
 * then the corresponding promise will internally reference an array of promises.
 */
function crawl (obj, path, $refs, options) {
  let promises = [];

  if (obj && typeof obj === "object") {
    if ($Ref.isExternal$Ref(obj)) {
      promises.push(resolve$Ref(obj, path, $refs, options));
    }
    else {
      for (let key of Object.keys(obj)) {
        let keyPath = Pointer.join(path, key);
        let value = obj[key];

        if ($Ref.isExternal$Ref(value)) {
          promises.push(resolve$Ref(value, keyPath, $refs, options));
        }
        else {
          promises = promises.concat(crawl(value, keyPath, $refs, options));
        }
      }
    }
  }

  return promises;
}

/**
 * Resolves the given JSON Reference, and then crawls the resulting value.
 *
 * @param {{$ref: string}} $ref - The JSON Reference to resolve
 * @param {string} path - The full path of `$ref`, possibly with a JSON Pointer in the hash
 * @param {$Refs} $refs
 * @param {$RefParserOptions} options
 *
 * @returns {Promise}
 * The promise resolves once all JSON references in the object have been resolved,
 * including nested references that are contained in externally-referenced files.
 */
async function resolve$Ref ($ref, path, $refs, options) {
  // console.log('Resolving $ref pointer "%s" at %s', $ref.$ref, path);

  let resolvedPath = url.resolve(path, $ref.$ref);
  let withoutHash = url.stripHash(resolvedPath);

  // Do we already have this $ref?
  $ref = $refs._$refs[withoutHash];
  if ($ref) {
    // We've already parsed this $ref, so use the existing value
    return Promise.resolve($ref.value);
  }

  // Parse the $referenced file/url
  const result = await parse(resolvedPath, $refs, options);

  // Crawl the parsed value
  // console.log('Resolving $ref pointers in %s', withoutHash);
  let promises = crawl(result, withoutHash + "#", $refs, options);

  return Promise.all(promises);
}
