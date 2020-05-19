JSON Schema $Ref Parser
============================
#### Parse, Resolve, and Dereference JSON Schema $ref pointers

[![Build Status](https://api.travis-ci.com/APIDevTools/json-schema-ref-parser.svg?branch=master)](https://travis-ci.com/APIDevTools/json-schema-ref-parser)
[![Coverage Status](https://coveralls.io/repos/github/APIDevTools/json-schema-ref-parser/badge.svg?branch=master)](https://coveralls.io/github/APIDevTools/json-schema-ref-parser)

[![npm](https://img.shields.io/npm/v/json-schema-ref-parser.svg)](https://www.npmjs.com/package/json-schema-ref-parser)
[![Dependencies](https://david-dm.org/APIDevTools/json-schema-ref-parser.svg)](https://david-dm.org/APIDevTools/json-schema-ref-parser)
[![License](https://img.shields.io/npm/l/json-schema-ref-parser.svg)](LICENSE)


[![OS and Browser Compatibility](https://apidevtools.org/img/badges/ci-badges-with-ie.svg)](https://travis-ci.com/APIDevTools/json-schema-ref-parser)


The Problem:
--------------------------
You've got a JSON Schema with `$ref` pointers to other files and/or URLs.  Maybe you know all the referenced files ahead of time.  Maybe you don't.  Maybe some are local files, and others are remote URLs.  Maybe they are a mix of JSON and YAML format.  Maybe some of the files contain cross-references to each other.

```javascript
{
  "definitions": {
    "person": {
      // references an external file
      "$ref": "schemas/people/Bruce-Wayne.json"
    },
    "place": {
      // references a sub-schema in an external file
      "$ref": "schemas/places.yaml#/definitions/Gotham-City"
    },
    "thing": {
      // references a URL
      "$ref": "http://wayne-enterprises.com/things/batmobile"
    },
    "color": {
      // references a value in an external file via an internal reference
      "$ref": "#/definitions/thing/properties/colors/black-as-the-night"
    }
  }
}
```


The Solution:
--------------------------
JSON Schema $Ref Parser is a full [JSON Reference](https://tools.ietf.org/html/draft-pbryan-zyp-json-ref-03) and [JSON Pointer](https://tools.ietf.org/html/rfc6901) implementation that crawls even the most complex [JSON Schemas](http://json-schema.org/latest/json-schema-core.html) and gives you simple, straightforward JavaScript objects.

- Use **JSON** or **YAML** schemas &mdash; or even a mix of both!
- Supports `$ref` pointers to external files and URLs, as well as [custom sources](https://apidevtools.org/json-schema-ref-parser/docs/plugins/resolvers.html) such as databases
- Can [bundle](https://apidevtools.org/json-schema-ref-parser/docs/ref-parser.html#bundlepath-options-callback) multiple files into a single schema that only has _internal_ `$ref` pointers
- Can [dereference](https://apidevtools.org/json-schema-ref-parser/docs/ref-parser.html#dereferencepath-options-callback) your schema, producing a plain-old JavaScript object that's easy to work with
- Supports [circular references](https://apidevtools.org/json-schema-ref-parser/docs/#circular-refs), nested references, back-references, and cross-references between files
- Maintains object reference equality &mdash; `$ref` pointers to the same value always resolve to the same object instance
- [Tested](https://apidevtools.org/json-schema-ref-parser/test/) in Node, io.js, and all major web browsers on Windows, Mac, and Linux


Example
--------------------------

```javascript
$RefParser.dereference(mySchema, function(err, schema) {
  if (err) {
    console.error(err);
  }
  else {
    // `schema` is just a normal JavaScript object that contains your entire JSON Schema,
    // including referenced files, combined into a single object
    console.log(schema.definitions.person.properties.firstName);
  }
});
```

Or use [Promises syntax](http://javascriptplayground.com/blog/2015/02/promises/) instead. The following example is the same as above:

```javascript
$RefParser.dereference(mySchema)
  .then(function(schema) {
    console.log(schema.definitions.person.properties.firstName);
  })
  .catch(function(err) {
    console.error(err);
  });
```

For more detailed examples, please see the [API Documentation](https://apidevtools.org/json-schema-ref-parser/docs/)


Installation
--------------------------
#### Node
Install using [npm](https://docs.npmjs.com/about-npm/):

```bash
npm install json-schema-ref-parser
```

Then require it in your code:

```javascript
var $RefParser = require('json-schema-ref-parser');
```

#### Web Browsers
Reference [`ref-parser.js`](dist/ref-parser.js) or [`ref-parser.min.js`](dist/ref-parser.min.js) in your HTML:

```html
<script src="https://unpkg.com/json-schema-ref-parser/dist/ref-parser.min.js"></script>
<script>
  $RefParser.dereference(mySchema)
    .then(function(schema) {
      console.log(schema.definitions.person.properties.firstName);
    })
    .catch(function(err) {
      console.error(err);
    });
</script>
```


API Documentation
--------------------------
Full API documentation is available [right here](https://apidevtools.org/json-schema-ref-parser/docs/)


Contributing
--------------------------
I welcome any contributions, enhancements, and bug-fixes.  [File an issue](https://github.com/APIDevTools/json-schema-ref-parser/issues) on GitHub and [submit a pull request](https://github.com/APIDevTools/json-schema-ref-parser/pulls).

#### Building/Testing
To build/test the project locally on your computer:

1. __Clone this repo__<br>
`git clone https://github.com/APIDevTools/json-schema-ref-parser.git`

2. __Install dependencies__<br>
`npm install`

3. __Run the build script__<br>
`npm run build`

4. __Run the tests__<br>
`npm test`

5. __Start the local web server__<br>
`npm start` (then browse to [http://localhost:8080/test/](http://localhost:8080/test/))


License
--------------------------
JSON Schema $Ref Parser is 100% free and open-source, under the [MIT license](LICENSE). Use it however you want.

Big Thanks To
--------------------------
Thanks to these awesome companies for their support of Open Source developers ❤

[![Travis CI](https://jsdevtools.org/img/badges/travis-ci.svg)](https://travis-ci.com)
[![SauceLabs](https://jsdevtools.org/img/badges/sauce-labs.svg)](https://saucelabs.com)
[![Coveralls](https://jsdevtools.org/img/badges/coveralls.svg)](https://coveralls.io)
