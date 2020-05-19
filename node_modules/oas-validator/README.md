# oas-validator

Usage:

```javascript
const validator = require('oas-validator');
const options = {};
validator.validate(openapi, options, function(err, options){
  // options.valid contains the result of the validation
  // options.context now contains a stack (array) of JSON-Pointer strings
});
```

If the callback argument of `validate` is omitted, a Promise is returned instead.

Please note the `validateSync` function is now a misnomer, as it also returns
a Promise or takes a callback. It will likely be renamed `validateInner`
or similar in the next major release.

See here for complete [documentation](/docs/options.md) of the `options` object.
