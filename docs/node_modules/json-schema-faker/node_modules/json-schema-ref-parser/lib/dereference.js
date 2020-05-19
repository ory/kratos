"use strict";

var $Ref = require("./ref"),
    Pointer = require("./pointer"),
    ono = require("ono"),
    url = require("./util/url");

module.exports = dereference;

/**
 * Crawls the JSON schema, finds all JSON references, and dereferences them.
 * This method mutates the JSON schema object, replacing JSON references with their resolved value.
 *
 * @param {$RefParser} parser
 * @param {$RefParserOptions} options
 */
function dereference (parser, options) {
  // console.log('Dereferencing $ref pointers in %s', parser.$refs._root$Ref.path);
  var dereferenced = crawl(parser.schema, parser.$refs._root$Ref.path, "#", [], parser.$refs, options);
  parser.$refs.circular = dereferenced.circular;
  parser.schema = dereferenced.value;
}

/**
 * Recursively crawls the given value, and dereferences any JSON references.
 *
 * @param {*} obj - The value to crawl. If it's not an object or array, it will be ignored.
 * @param {string} path - The full path of `obj`, possibly with a JSON Pointer in the hash
 * @param {string} pathFromRoot - The path of `obj` from the schema root
 * @param {object[]} parents - An array of the parent objects that have already been dereferenced
 * @param {$Refs} $refs
 * @param {$RefParserOptions} options
 * @returns {{value: object, circular: boolean}}
 */
function crawl (obj, path, pathFromRoot, parents, $refs, options) {
  var dereferenced;
  var result = {
    value: obj,
    circular: false
  };

  if (obj && typeof obj === "object") {
    parents.push(obj);

    if ($Ref.isAllowed$Ref(obj, options)) {
      dereferenced = dereference$Ref(obj, path, pathFromRoot, parents, $refs, options);
      result.circular = dereferenced.circular;
      result.value = dereferenced.value;
    }
    else {
      Object.keys(obj).forEach(function (key) {
        var keyPath = Pointer.join(path, key);
        var keyPathFromRoot = Pointer.join(pathFromRoot, key);
        var value = obj[key];
        var circular = false;

        if ($Ref.isAllowed$Ref(value, options)) {
          dereferenced = dereference$Ref(value, keyPath, keyPathFromRoot, parents, $refs, options);
          circular = dereferenced.circular;
          obj[key] = dereferenced.value;
        }
        else {
          if (parents.indexOf(value) === -1) {
            dereferenced = crawl(value, keyPath, keyPathFromRoot, parents, $refs, options);
            circular = dereferenced.circular;
            obj[key] = dereferenced.value;
          }
          else {
            circular = foundCircularReference(keyPath, $refs, options);
          }
        }

        // Set the "isCircular" flag if this or any other property is circular
        result.circular = result.circular || circular;
      });
    }

    parents.pop();
  }

  return result;
}

/**
 * Dereferences the given JSON Reference, and then crawls the resulting value.
 *
 * @param {{$ref: string}} $ref - The JSON Reference to resolve
 * @param {string} path - The full path of `$ref`, possibly with a JSON Pointer in the hash
 * @param {string} pathFromRoot - The path of `$ref` from the schema root
 * @param {object[]} parents - An array of the parent objects that have already been dereferenced
 * @param {$Refs} $refs
 * @param {$RefParserOptions} options
 * @returns {{value: object, circular: boolean}}
 */
function dereference$Ref ($ref, path, pathFromRoot, parents, $refs, options) {
  // console.log('Dereferencing $ref pointer "%s" at %s', $ref.$ref, path);

  var $refPath = url.resolve(path, $ref.$ref);
  var pointer = $refs._resolve($refPath, options);

  // Check for circular references
  var directCircular = pointer.circular;
  var circular = directCircular || parents.indexOf(pointer.value) !== -1;
  circular && foundCircularReference(path, $refs, options);

  // Dereference the JSON reference
  var dereferencedValue = $Ref.dereference($ref, pointer.value);

  // Crawl the dereferenced value (unless it's circular)
  if (!circular) {
    // Determine if the dereferenced value is circular
    var dereferenced = crawl(dereferencedValue, pointer.path, pathFromRoot, parents, $refs, options);
    circular = dereferenced.circular;
    dereferencedValue = dereferenced.value;
  }

  if (circular && !directCircular && options.dereference.circular === "ignore") {
    // The user has chosen to "ignore" circular references, so don't change the value
    dereferencedValue = $ref;
  }

  if (directCircular) {
    // The pointer is a DIRECT circular reference (i.e. it references itself).
    // So replace the $ref path with the absolute path from the JSON Schema root
    dereferencedValue.$ref = pathFromRoot;
  }

  return {
    circular: circular,
    value: dereferencedValue
  };
}

/**
 * Called when a circular reference is found.
 * It sets the {@link $Refs#circular} flag, and throws an error if options.dereference.circular is false.
 *
 * @param {string} keyPath - The JSON Reference path of the circular reference
 * @param {$Refs} $refs
 * @param {$RefParserOptions} options
 * @returns {boolean} - always returns true, to indicate that a circular reference was found
 */
function foundCircularReference (keyPath, $refs, options) {
  $refs.circular = true;
  if (!options.dereference.circular) {
    throw ono.reference("Circular $ref pointer found at %s", keyPath);
  }
  return true;
}
