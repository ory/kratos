"use strict";

module.exports = Pointer;

const $Ref = require("./ref");
const url = require("./util/url");
const { ono } = require("@jsdevtools/ono");
const slashes = /\//g;
const tildes = /~/g;
const escapedSlash = /~1/g;
const escapedTilde = /~0/g;

/**
 * This class represents a single JSON pointer and its resolved value.
 *
 * @param {$Ref} $ref
 * @param {string} path
 * @param {string} [friendlyPath] - The original user-specified path (used for error messages)
 * @constructor
 */
function Pointer ($ref, path, friendlyPath) {
  /**
   * The {@link $Ref} object that contains this {@link Pointer} object.
   * @type {$Ref}
   */
  this.$ref = $ref;

  /**
   * The file path or URL, containing the JSON pointer in the hash.
   * This path is relative to the path of the main JSON schema file.
   * @type {string}
   */
  this.path = path;

  /**
   * The original path or URL, used for error messages.
   * @type {string}
   */
  this.originalPath = friendlyPath || path;

  /**
   * The value of the JSON pointer.
   * Can be any JSON type, not just objects. Unknown file types are represented as Buffers (byte arrays).
   * @type {?*}
   */
  this.value = undefined;

  /**
   * Indicates whether the pointer references itself.
   * @type {boolean}
   */
  this.circular = false;

  /**
   * The number of indirect references that were traversed to resolve the value.
   * Resolving a single pointer may require resolving multiple $Refs.
   * @type {number}
   */
  this.indirections = 0;
}

/**
 * Resolves the value of a nested property within the given object.
 *
 * @param {*} obj - The object that will be crawled
 * @param {$RefParserOptions} options
 *
 * @returns {Pointer}
 * Returns a JSON pointer whose {@link Pointer#value} is the resolved value.
 * If resolving this value required resolving other JSON references, then
 * the {@link Pointer#$ref} and {@link Pointer#path} will reflect the resolution path
 * of the resolved value.
 */
Pointer.prototype.resolve = function (obj, options) {
  let tokens = Pointer.parse(this.path);

  // Crawl the object, one token at a time
  this.value = obj;
  for (let i = 0; i < tokens.length; i++) {
    if (resolveIf$Ref(this, options)) {
      // The $ref path has changed, so append the remaining tokens to the path
      this.path = Pointer.join(this.path, tokens.slice(i));
    }

    let token = tokens[i];
    if (this.value[token] === undefined) {
      throw ono.syntax(`Error resolving $ref pointer "${this.originalPath}". \nToken "${token}" does not exist.`);
    }
    else {
      this.value = this.value[token];
    }
  }

  // Resolve the final value
  resolveIf$Ref(this, options);
  return this;
};

/**
 * Sets the value of a nested property within the given object.
 *
 * @param {*} obj - The object that will be crawled
 * @param {*} value - the value to assign
 * @param {$RefParserOptions} options
 *
 * @returns {*}
 * Returns the modified object, or an entirely new object if the entire object is overwritten.
 */
Pointer.prototype.set = function (obj, value, options) {
  let tokens = Pointer.parse(this.path);
  let token;

  if (tokens.length === 0) {
    // There are no tokens, replace the entire object with the new value
    this.value = value;
    return value;
  }

  // Crawl the object, one token at a time
  this.value = obj;
  for (let i = 0; i < tokens.length - 1; i++) {
    resolveIf$Ref(this, options);

    token = tokens[i];
    if (this.value && this.value[token] !== undefined) {
      // The token exists
      this.value = this.value[token];
    }
    else {
      // The token doesn't exist, so create it
      this.value = setValue(this, token, {});
    }
  }

  // Set the value of the final token
  resolveIf$Ref(this, options);
  token = tokens[tokens.length - 1];
  setValue(this, token, value);

  // Return the updated object
  return obj;
};

/**
 * Parses a JSON pointer (or a path containing a JSON pointer in the hash)
 * and returns an array of the pointer's tokens.
 * (e.g. "schema.json#/definitions/person/name" => ["definitions", "person", "name"])
 *
 * The pointer is parsed according to RFC 6901
 * {@link https://tools.ietf.org/html/rfc6901#section-3}
 *
 * @param {string} path
 * @returns {string[]}
 */
Pointer.parse = function (path) {
  // Get the JSON pointer from the path's hash
  let pointer = url.getHash(path).substr(1);

  // If there's no pointer, then there are no tokens,
  // so return an empty array
  if (!pointer) {
    return [];
  }

  // Split into an array
  pointer = pointer.split("/");

  // Decode each part, according to RFC 6901
  for (let i = 0; i < pointer.length; i++) {
    pointer[i] = decodeURIComponent(pointer[i].replace(escapedSlash, "/").replace(escapedTilde, "~"));
  }

  if (pointer[0] !== "") {
    throw ono.syntax(`Invalid $ref pointer "${pointer}". Pointers must begin with "#/"`);
  }

  return pointer.slice(1);
};

/**
 * Creates a JSON pointer path, by joining one or more tokens to a base path.
 *
 * @param {string} base - The base path (e.g. "schema.json#/definitions/person")
 * @param {string|string[]} tokens - The token(s) to append (e.g. ["name", "first"])
 * @returns {string}
 */
Pointer.join = function (base, tokens) {
  // Ensure that the base path contains a hash
  if (base.indexOf("#") === -1) {
    base += "#";
  }

  // Append each token to the base path
  tokens = Array.isArray(tokens) ? tokens : [tokens];
  for (let i = 0; i < tokens.length; i++) {
    let token = tokens[i];
    // Encode the token, according to RFC 6901
    base += "/" + encodeURIComponent(token.replace(tildes, "~0").replace(slashes, "~1"));
  }

  return base;
};

/**
 * If the given pointer's {@link Pointer#value} is a JSON reference,
 * then the reference is resolved and {@link Pointer#value} is replaced with the resolved value.
 * In addition, {@link Pointer#path} and {@link Pointer#$ref} are updated to reflect the
 * resolution path of the new value.
 *
 * @param {Pointer} pointer
 * @param {$RefParserOptions} options
 * @returns {boolean} - Returns `true` if the resolution path changed
 */
function resolveIf$Ref (pointer, options) {
  // Is the value a JSON reference? (and allowed?)

  if ($Ref.isAllowed$Ref(pointer.value, options)) {
    let $refPath = url.resolve(pointer.path, pointer.value.$ref);

    if ($refPath === pointer.path) {
      // The value is a reference to itself, so there's nothing to do.
      pointer.circular = true;
    }
    else {
      let resolved = pointer.$ref.$refs._resolve($refPath, options);
      pointer.indirections += resolved.indirections + 1;

      if ($Ref.isExtended$Ref(pointer.value)) {
        // This JSON reference "extends" the resolved value, rather than simply pointing to it.
        // So the resolved path does NOT change.  Just the value does.
        pointer.value = $Ref.dereference(pointer.value, resolved.value);
        return false;
      }
      else {
        // Resolve the reference
        pointer.$ref = resolved.$ref;
        pointer.path = resolved.path;
        pointer.value = resolved.value;
      }

      return true;
    }
  }
}

/**
 * Sets the specified token value of the {@link Pointer#value}.
 *
 * The token is evaluated according to RFC 6901.
 * {@link https://tools.ietf.org/html/rfc6901#section-4}
 *
 * @param {Pointer} pointer - The JSON Pointer whose value will be modified
 * @param {string} token - A JSON Pointer token that indicates how to modify `obj`
 * @param {*} value - The value to assign
 * @returns {*} - Returns the assigned value
 */
function setValue (pointer, token, value) {
  if (pointer.value && typeof pointer.value === "object") {
    if (token === "-" && Array.isArray(pointer.value)) {
      pointer.value.push(value);
    }
    else {
      pointer.value[token] = value;
    }
  }
  else {
    throw ono.syntax(`Error assigning $ref pointer "${pointer.path}". \nCannot set "${token}" of a non-object.`);
  }
  return value;
}
