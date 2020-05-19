ono (Oh No!)
============================
#### Throw better errors.

[![Build Status](https://api.travis-ci.com/JS-DevTools/ono.svg?branch=master)](https://travis-ci.com/JS-DevTools/ono)
[![Coverage Status](https://coveralls.io/repos/github/JS-DevTools/ono/badge.svg?branch=master)](https://coveralls.io/github/JS-DevTools/ono)

[![npm](https://img.shields.io/npm/v/ono.svg)](https://www.npmjs.com/package/ono)
[![Dependencies](https://david-dm.org/JS-DevTools/ono.svg)](https://david-dm.org/JS-DevTools/ono)
[![License](https://img.shields.io/npm/l/ono.svg)](LICENSE)

[![OS and Browser Compatibility](https://jsdevtools.org/img/ci-badges-with-ie.svg)](https://travis-ci.com/JS-DevTools/ono)

Features
--------------------------
* Formatted error messages, using Node's [`util.format()`](https://nodejs.org/api/util.html#util_util_format_format_args) or your own custom formatter
* Wrap and re-throw an error _without_ losing the original error's message, stack trace, and properties
* Add custom properties to your errors &mdash; great for error codes, support numbers, help URLs, etc.
* Errors can be serialized as JSON, including all native and custom properties
* [Tested](https://jsdevtools.org/ono/test/) on Node.js and all modern web browsers on Mac, Windows, Linux, iOS, and Android


Example
--------------------------

```javascript
// Throw an error with formatted message
throw ono("%s is invalid. Must be at least %d characters.", username, minLength);

// Wrap and re-throw an error without losing the original error's message and stack
throw ono(err, "An error occurred while saving your changes");

// Throw an error with custom properties (even a custom method!)
throw ono({code: 413, status: "Invalid data", retry: function() {...}});

// Add custom properties to an existing Error
throw ono(err, {code: 413, status: "Invalid data", retry: function() {...}})

// Any of the above can throw a specific Error subtype instead
throw ono.range(...);       // RangeError
throw ono.syntax(...);      // SyntaxError
throw ono.reference(...);   // ReferenceError
```


Installation
--------------------------
#### Node
Install using [npm](https://docs.npmjs.com/getting-started/what-is-npm):

```bash
npm install ono
```

Then require it in your code:

```javascript
var ono = require("ono");
```

#### Web Browsers
Reference [`ono.js`](dist/ono.js) or [`ono.min.js`](dist/ono.min.js) in your HTML:

```html
<script src="https://cdn.rawgit.com/JS-DevTools/ono/master/dist/ono.js"></script>
<script>
    var timestamp = new Date().toISOString();
    throw ono('This error was thrown at %s', timestamp);
</script>
```


API
--------------------------
### `ono([err], [props], [message, ...])`
Creates an [`Error`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Error) object with the given properties.

* `err` - _(optional)_ An existing error object. This error's message, stack trace, and properties will be appended to the new error.

* `props` - _(optional)_ An object whose properties will be added to the new error. Properties can be anything, including objects and functions.

* `message` - _(optional)_ The error message string. If it contains placeholders, then pass each placeholder's value as an additional parameter.  See [`ono.formatter`](#onoformatter) for more info.

##### Specific error types
The default `ono()` function always creates [`Error`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Error) objects, but you can use any of the following methods to explicitly create the corresponding Error subclass.  The method signatures are exactly the same as above.

Method            | Error type
:-----------------|:-------------------
`ono.error()`     |[`Error`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Error) (this is just an alternate syntax)
`ono.eval()`      |[`EvalError`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/EvalError)
`ono.range()`     |[`RangeError`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/RangeError)
`ono.reference()` |[`ReferenceError`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/ReferenceError)
`ono.syntax()`    |[`SyntaxError`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/SyntaxError)
`ono.type()`      |[`TypeError`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/TypeError)
`ono.uri()`       |[`URIError`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/URIError)

### `ono.formatter`
By default, Node's [`util.format()`](https://nodejs.org/api/util.html#util_util_format_format) function is used (even in browsers) to format error messages and substitute placeholders with their corresponding values. You can set `ono.formatter` to a [third-party formatter](https://www.npmjs.com/package/format) or even your own custom implementation, like this:

```javascript
ono.formatter = function(message) {
    var params = Array.prototype.slice.call(arguments, 1);
    return params.reduce(function(message, param, index) {
        return message.replace("$" + index, param);
    }, message);
}

throw ono("$0 must be greater than $1", 4, 10);
```


Contributing
--------------------------
I welcome any contributions, enhancements, and bug-fixes.  [File an issue](https://github.com/JS-DevTools/ono/issues) on GitHub and [submit a pull request](https://github.com/JS-DevTools/ono/pulls).

#### Building/Testing
To build/test the project locally on your computer:

1. __Clone this repo__<br>
`git clone https://github.com/JS-DevTools/ono.git`

2. __Install dependencies__<br>
`npm install`

3. __Run the build script__<br>
`npm run build`

4. __Run the tests__<br>
`npm test`


License
--------------------------
Ono is 100% free and open-source, under the [MIT license](LICENSE). Use it however you want.
