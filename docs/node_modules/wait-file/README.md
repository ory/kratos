# Wait-file

Wait for file resource(s) to become available in Node.js.

[![Build Status](https://travis-ci.com/endiliey/wait-file.svg?branch=master)](https://travis-ci.com/endiliey/wait-file)

## Installation

```bash
npm install wait-file
```

### Node.js API usage

waitFile(opts, [cb]) - function which triggers resource checks

- opts - see below example
- cb(err) - if err is provided then, resource checks did not succeed

Requires Node.js v8+

```javascript
const { waitFile } = require('wait-file');
const opts = {
  resources: ['file1', '/path/to/file2'],
  delay: 0, // initial delay in ms, default 0ms
  interval: 100, // poll interval in ms, default 250ms
  log: false, // outputs to stdout, remaining resources waited on and when complete or errored, default false
  reverse: false, // resources being NOT available, default false
  timeout: 30000, // timeout in ms, default Infinity
  verbose: false, // optional flag which outputs debug output, default false
  window: 1000, // stabilization time in ms, default 750ms
};

// Usage with callback function
waitFile(opts, function(err) {
  if (err) {
    return console.error(err);
  }
  // once here, all resources are available
});

// Usage with promises
waitFile(opts)
  .then(function() {
    // once here, all resources are available
  })
  .catch(function(err) {
    console.error(err);
  });

// Usage with async await
try {
  await waitFile(opts);
  // once here, all resources are available
} catch (err) {
  console.error(err);
}
```
