# drange

For adding/subtracting sets of range of numbers.

[![Build Status](https://secure.travis-ci.org/fent/node-drange.svg)](http://travis-ci.org/fent/node-drange)
[![Dependency Status](https://david-dm.org/fent/node-drange.svg)](https://david-dm.org/fent/node-drange)
[![codecov](https://codecov.io/gh/fent/node-drange/branch/master/graph/badge.svg)](https://codecov.io/gh/fent/node-drange)

# Usage

```
const DRange = require('drange');

var allNums = new DRange(1, 100); //[ 1-100 ]
var badNums = DRange(13).add(8).add(60,80); //[8, 13, 60-80]
var goodNums = allNums.clone().subtract(badNums);
console.log(goodNums.toString()); //[ 1-7, 9-12, 14-59, 81-100 ]
var randomGoodNum = goodNums.index(Math.floor(Math.random() * goodNums.length));
```

# API
### new DRange([low], [high])
Creates a new instance of DRange.

### DRange#length
The total length of all subranges

### DRange#add(low, high)
Adds a subrange

### DRange#add(drange)
Adds all of another DRange's subranges

### DRange#subtract(low, high)
Subtracts a subrange

### DRange#subtract(drange)
Subtracts all of another DRange's subranges

### DRange#intersect(low, range)
Keep only subranges that overlap the given subrange

### DRange#intersect(drange)
Intersect all of another DRange's subranges

### DRange#index(i)
Get the number at the specified index

```js
var drange = new DRange()
drange.add(1, 10);
drange.add(21, 30);
console.log(drange.index(15)); // 25
```

### DRange#numbers()
Get contained numbers

```js
var drange = new DRange(1, 4)
drange.add(6);
drange.subtract(2);
console.log(drange.numbers()); // [1, 3, 4, 6]
```

### DRange#subranges()
Get copy of subranges

```js
var drange = new DRange(1, 4)
drange.add(6, 8);
console.log(drange.subranges());
/*
[
  { low: 1, high: 4, length: 4 },
  { low: 6, high: 8, length: 3 }
]
*/
```

### DRange#clone()
Clones the drange, so that changes to it are not reflected on its clone


# Install

    npm install drange

# Tests

Tests are written with [mocha](https://mochajs.org)

    npm test

# Integration with TypeScript

DRange includes TypeScript definitions.

```typescript
import * as DRange from "drange";
const range: DRange = new Drange(2, 5);
```

Use dtslint to check the definition file.

    npm install -g dtslint
    npm run dtslint
