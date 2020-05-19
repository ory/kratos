"use strict";

/**
 * Returns the given plugins as an array, rather than an object map.
 * All other methods in this module expect an array of plugins rather than an object map.
 *
 * @param  {object} plugins - A map of plugin objects
 * @return {object[]}
 */
exports.all = function (plugins) {
  return Object.keys(plugins)
    .filter(function (key) {
      return typeof plugins[key] === "object";
    })
    .map(function (key) {
      plugins[key].name = key;
      return plugins[key];
    });
};

/**
 * Filters the given plugins, returning only the ones return `true` for the given method.
 *
 * @param  {object[]} plugins - An array of plugin objects
 * @param  {string}   method  - The name of the filter method to invoke for each plugin
 * @param  {object}   file    - A file info object, which will be passed to each method
 * @return {object[]}
 */
exports.filter = function (plugins, method, file) {
  return plugins
    .filter(function (plugin) {
      return !!getResult(plugin, method, file);
    });
};

/**
 * Sorts the given plugins, in place, by their `order` property.
 *
 * @param {object[]} plugins - An array of plugin objects
 * @returns {object[]}
 */
exports.sort = function (plugins) {
  plugins.forEach(function (plugin) {
    plugin.order = plugin.order || Number.MAX_SAFE_INTEGER;
  });

  return plugins.sort(function (a, b) { return a.order - b.order; });
};

/**
 * Runs the specified method of the given plugins, in order, until one of them returns a successful result.
 * Each method can return a synchronous value, a Promise, or call an error-first callback.
 * If the promise resolves successfully, or the callback is called without an error, then the result
 * is immediately returned and no further plugins are called.
 * If the promise rejects, or the callback is called with an error, then the next plugin is called.
 * If ALL plugins fail, then the last error is thrown.
 *
 * @param {object[]}  plugins - An array of plugin objects
 * @param {string}    method  - The name of the method to invoke for each plugin
 * @param {object}    file    - A file info object, which will be passed to each method
 * @returns {Promise}
 */
exports.run = function (plugins, method, file) {
  var plugin, lastError, index = 0;

  return new Promise(function (resolve, reject) {
    runNextPlugin();

    function runNextPlugin () {
      plugin = plugins[index++];
      if (!plugin) {
        // There are no more functions, so re-throw the last error
        return reject(lastError);
      }

      try {
        // console.log('  %s', plugin.name);
        var result = getResult(plugin, method, file, callback);
        if (result && typeof result.then === "function") {
          // A promise was returned
          result.then(onSuccess, onError);
        }
        else if (result !== undefined) {
          // A synchronous result was returned
          onSuccess(result);
        }
        // else { the callback will be called }
      }
      catch (e) {
        onError(e);
      }
    }

    function callback (err, result) {
      if (err) {
        onError(err);
      }
      else {
        onSuccess(result);
      }
    }

    function onSuccess (result) {
      // console.log('    success');
      resolve({
        plugin: plugin,
        result: result
      });
    }

    function onError (err) {
      // console.log('    %s', err.message || err);
      lastError = err;
      runNextPlugin();
    }
  });
};

/**
 * Returns the value of the given property.
 * If the property is a function, then the result of the function is returned.
 * If the value is a RegExp, then it will be tested against the file URL.
 * If the value is an aray, then it will be compared against the file extension.
 *
 * @param   {object}   obj        - The object whose property/method is called
 * @param   {string}   prop       - The name of the property/method to invoke
 * @param   {object}   file       - A file info object, which will be passed to the method
 * @param   {function} [callback] - A callback function, which will be passed to the method
 * @returns {*}
 */
function getResult (obj, prop, file, callback) {
  var value = obj[prop];

  if (typeof value === "function") {
    return value.apply(obj, [file, callback]);
  }

  if (!callback) {
    // The synchronous plugin functions (canParse and canRead)
    // allow a "shorthand" syntax, where the user can match
    // files by RegExp or by file extension.
    if (value instanceof RegExp) {
      return value.test(file.url);
    }
    else if (typeof value === "string") {
      return value === file.extension;
    }
    else if (Array.isArray(value)) {
      return value.indexOf(file.extension) !== -1;
    }
  }

  return value;
}
