module.exports = {
  "extends": ["ash-nazg/sauron-node"],
  "settings": {
      "polyfills": [
          "Array.isArray",
          "console",
          "Date.now",
          "document.head",
          "document.querySelector",
          "JSON",
          "Object.keys",
          "XMLHttpRequest"
      ]
  },
  "overrides": [
      {
          "files": ["src/jsonpath.js", "test-helpers/node-env.js"],
          // Apparent bug with `overrides` necessitating this
          "globals": {
              "require": "readonly",
              "module": "readonly"
          }
      },
      {
          "files": ["*.md"],
          "rules": {
              "import/unambiguous": 0,
              "import/no-commonjs": 0,
              "import/no-unresolved": ["error", {"ignore": ["jsonpath-plus"]}],
              "no-undef": 0,
              "no-unused-vars": ["error", {
                  "varsIgnorePattern": "json|result"
              }],
              "node/no-missing-require": ["error", {
                  "allowModules": ["jsonpath-plus"]
              }],
              "node/no-missing-import": ["error", {
                  "allowModules": ["jsonpath-plus"]
              }]
          }
      },
      {
          "files": ["test/**"],
          "globals": {
              "assert": "readonly",
              "jsonpath": "readonly",
              "require": "readonly",
              "module": "readonly"
          },
          "parserOptions": {
              "sourceType": "script"
          },
          "env": {"mocha": true},
          "rules": {
              "strict": ["error", "global"],
              "import/no-commonjs": 0,
              "import/unambiguous": 0,
              "quotes": 0,
              // Todo: Reenable
              "max-len": 0
          }
      }
  ],
  "rules": {
    "indent": ["error", 4, {"outerIIFEBody": 0}],
    "promise/prefer-await-to-callbacks": 0,
    "quote-props": 0,
    "require-jsdoc": 0,
    // Reenable when no longer having problems
    "unicorn/no-unsafe-regex": 0
  }
};
